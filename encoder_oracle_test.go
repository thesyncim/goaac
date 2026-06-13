package aac

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/thesyncim/goaac/internal/wav"
)

func TestEncoderFDKAACOracleParity(t *testing.T) {
	oracle := os.Getenv("GOAAC_FDK_ENCODER_ORACLE")
	if oracle == "" {
		t.Skip("set GOAAC_FDK_ENCODER_ORACLE to a pinned fdk-aac aac-enc binary")
	}

	tests := []struct {
		name          string
		channels      int
		channelConfig int
		bitRate       int
		frames        int
		fill          func([]int16, int, int)
	}{
		{
			name:          "mono_smooth_48k_64k",
			channels:      1,
			channelConfig: 1,
			bitRate:       64000,
			frames:        4,
			fill: func(dst []int16, channels int, frame int) {
				fillEncoderSmoothPCM(dst, channels)
				for i := range dst {
					dst[i] = int16((int(dst[i]) * (frame + 1)) / 4)
				}
			},
		},
		{
			name:          "stereo_transition_48k_128k",
			channels:      2,
			channelConfig: 2,
			bitRate:       128000,
			frames:        6,
			fill:          fillEncoderTransitionPCM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pcm := makeEncoderOraclePCM(tt.frames, tt.channels, tt.fill)
			goStream := encodeEncoderOracleADTS(t, pcm, tt.channels, tt.channelConfig, tt.bitRate)
			fdkStream := runFDKAACEncoderOracle(t, oracle, pcm, tt.channels, tt.bitRate)
			assertEncoderOracleParity(t, tt.name, goStream, fdkStream)
		})
	}
}

func makeEncoderOraclePCM(frames int, channels int, fill func([]int16, int, int)) []int16 {
	pcm := make([]int16, frames*encoderSamplesPerFrame*channels)
	for frame := 0; frame < frames; frame++ {
		start := frame * encoderSamplesPerFrame * channels
		fill(pcm[start:start+encoderSamplesPerFrame*channels], channels, frame)
	}
	return pcm
}

func encodeEncoderOracleADTS(t *testing.T, pcm []int16, channels int, channelConfig int, bitRate int) []byte {
	t.Helper()
	enc, err := NewEncoder(EncoderOptions{
		Config: Config{
			ObjectType:    AOTAACLC,
			SampleRate:    48000,
			ChannelConfig: channelConfig,
		},
		BitRate:   bitRate,
		Transport: TransportADTS,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	var out []byte
	frameSamples := encoderSamplesPerFrame * channels
	for frame, off := 0, 0; off < len(pcm); frame, off = frame+1, off+frameSamples {
		var info EncodedFrameInfo
		out, info, err = enc.EncodeADTSFrameInto(out, pcm[off:off+frameSamples])
		if err != nil {
			t.Fatalf("go encode frame %d: %v", frame, err)
		}
		if info.PayloadBytes <= 0 || info.ADTSHeaderBytes != ADTSHeaderSize {
			t.Fatalf("go encode frame %d info = %+v", frame, info)
		}
	}
	for flushFrame := 0; ; flushFrame++ {
		var info EncodedFrameInfo
		var more bool
		out, info, more, err = enc.FlushFrameInto(out)
		if err != nil {
			t.Fatalf("go flush frame %d: %v", flushFrame, err)
		}
		if !more {
			break
		}
		if info.InputSamples != 0 || info.PayloadBytes <= 0 || info.ADTSHeaderBytes != ADTSHeaderSize {
			t.Fatalf("go flush frame %d info = %+v", flushFrame, info)
		}
		if flushFrame > 3 {
			t.Fatalf("go flush emitted too many frames")
		}
	}
	return out
}

func runFDKAACEncoderOracle(t *testing.T, oracle string, pcm []int16, channels int, bitRate int) []byte {
	t.Helper()
	tmp := t.TempDir()
	in := filepath.Join(tmp, "input.wav")
	out := filepath.Join(tmp, "oracle.aac")

	var buf bytes.Buffer
	if err := wav.WriteS16(&buf, pcm, 48000, channels); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(in, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(oracle, "-t", "2", "-a", "1", "-r", strconv.Itoa(bitRate), in, out)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fdk-aac oracle failed: %v\n%s", err, output)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertEncoderOracleParity(t *testing.T, name string, goStream []byte, fdkStream []byte) {
	t.Helper()
	if bytes.Equal(goStream, fdkStream) {
		return
	}
	goFrames, goErr := SplitADTSFrames(goStream)
	fdkFrames, fdkErr := SplitADTSFrames(fdkStream)
	if goErr != nil || fdkErr != nil {
		t.Fatalf("%s ADTS parse: go=%v fdk=%v", name, goErr, fdkErr)
	}
	frameMsg := firstEncoderOracleFrameDiff(goFrames, fdkFrames)
	t.Fatalf(
		"%s fdk-aac parity mismatch: go len=%d sha256=%s frames=%d; fdk len=%d sha256=%s frames=%d; first byte diff=%d; %s",
		name,
		len(goStream), sha256Hex(goStream), len(goFrames),
		len(fdkStream), sha256Hex(fdkStream), len(fdkFrames),
		firstByteDiff(goStream, fdkStream),
		frameMsg,
	)
}

func firstEncoderOracleFrameDiff(goFrames []ADTSFrame, fdkFrames []ADTSFrame) string {
	n := len(goFrames)
	if len(fdkFrames) < n {
		n = len(fdkFrames)
	}
	for i := 0; i < n; i++ {
		g := goFrames[i].Data
		f := fdkFrames[i].Data
		if !bytes.Equal(g, f) {
			headerDiff := firstByteDiff(g[:goFrames[i].Header.HeaderLength], f[:fdkFrames[i].Header.HeaderLength])
			payloadDiff := firstByteDiff(g[goFrames[i].Header.HeaderLength:], f[fdkFrames[i].Header.HeaderLength:])
			return fmt.Sprintf(
				"frame %d differs: go len=%d payload=%d sha256=%s; fdk len=%d payload=%d sha256=%s; header diff=%d payload diff=%d",
				i,
				len(g), goFrames[i].Header.PayloadLength, sha256Hex(g),
				len(f), fdkFrames[i].Header.PayloadLength, sha256Hex(f),
				headerDiff, payloadDiff,
			)
		}
	}
	return fmt.Sprintf("matching prefix frames=%d, frame-count differs go=%d fdk=%d", n, len(goFrames), len(fdkFrames))
}

func firstByteDiff(a []byte, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	if len(a) != len(b) {
		return n
	}
	return -1
}
