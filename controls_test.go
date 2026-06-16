package aac

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestNewOptionsControlPlane(t *testing.T) {
	adts, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer adts.Close()
	if got := adts.Transport(); got != TransportADTS {
		t.Fatalf("auto transport without config = %s, want %s", got, TransportADTS)
	}

	raw, err := New(Options{
		Config: Config{
			ObjectType:    AOTAACLC,
			SampleRate:    44100,
			ChannelConfig: 2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if got := raw.Transport(); got != TransportRaw {
		t.Fatalf("auto transport with raw config = %s, want %s", got, TransportRaw)
	}

	if _, err := New(Options{Transport: Transport(99)}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("unknown transport err = %v, want ErrInvalidConfig", err)
	}
	if _, err := New(Options{Transport: TransportRaw}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("raw without config err = %v, want ErrInvalidConfig", err)
	}
}

func TestDecoderCloseTransitions(t *testing.T) {
	adts, err := New(Options{Transport: TransportADTS})
	if err != nil {
		t.Fatal(err)
	}
	if err := adts.Close(); err != nil {
		t.Fatal(err)
	}
	if err := adts.Close(); err != nil {
		t.Fatalf("second close = %v, want nil", err)
	}
	dst := []int16{7, 8}
	got, _, err := adts.Decode(dst, nil)
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("Decode after close err = %v, want ErrClosed", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("Decode after close changed dst: got %v, want %v", got, dst)
	}

	raw := newRawTestDecoder(t)
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}
	got, _, err = raw.DecodeRawInto(dst, []byte{0})
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("DecodeRawInto after close err = %v, want ErrClosed", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("DecodeRawInto after close changed dst: got %v, want %v", got, dst)
	}
}

func TestNilDecoderControls(t *testing.T) {
	var dec *Decoder
	if got := dec.Transport(); got != TransportAuto {
		t.Fatalf("nil decoder transport = %s, want %s", got, TransportAuto)
	}
	if cfg := dec.Config(); cfg.SampleRate != 0 || cfg.Channels != 0 {
		t.Fatalf("nil decoder config = %+v, want zero value", cfg)
	}
	if err := dec.Close(); err != nil {
		t.Fatalf("nil decoder close = %v, want nil", err)
	}

	dst := []int16{7, 8}
	if got, _, err := dec.Decode(dst, nil); !errors.Is(err, ErrClosed) || !equalInt16s(got, dst) {
		t.Fatalf("nil Decode got=%v err=%v, want dst/ErrClosed", got, err)
	}
	if got, err := dec.DecodeRaw(nil); !errors.Is(err, ErrClosed) || got != nil {
		t.Fatalf("nil DecodeRaw got=%v err=%v, want nil/ErrClosed", got, err)
	}
	if got, _, err := dec.DecodeRawInto(dst, nil); !errors.Is(err, ErrClosed) || !equalInt16s(got, dst) {
		t.Fatalf("nil DecodeRawInto got=%v err=%v, want dst/ErrClosed", got, err)
	}
	if got, _, err := dec.DecodeADTSFrameInto(dst, nil); !errors.Is(err, ErrClosed) || !equalInt16s(got, dst) {
		t.Fatalf("nil DecodeADTSFrameInto got=%v err=%v, want dst/ErrClosed", got, err)
	}
}

func TestDecoderTransportMisusePreservesDst(t *testing.T) {
	adts, err := New(Options{Transport: TransportADTS})
	if err != nil {
		t.Fatal(err)
	}
	defer adts.Close()

	dst := []int16{1, 2, 3}
	got, _, err := adts.DecodeRawInto(dst, []byte{0})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("ADTS decoder raw err = %v, want ErrInvalidConfig", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("ADTS decoder raw misuse changed dst: got %v, want %v", got, dst)
	}

	raw := newRawTestDecoder(t)
	defer raw.Close()
	frame, err := appendTestADTSHeader(nil, Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}, 0)
	if err != nil {
		t.Fatal(err)
	}
	got, _, err = raw.DecodeADTSFrameInto(dst, frame)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("raw decoder ADTS err = %v, want ErrInvalidConfig", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("raw decoder ADTS misuse changed dst: got %v, want %v", got, dst)
	}
}

func TestDecoderConfigReturnsCopy(t *testing.T) {
	dec := newRawTestDecoder(t)
	defer dec.Close()

	cfg := dec.Config()
	if len(cfg.ExtraData) == 0 {
		t.Fatal("raw config ExtraData is empty")
	}
	cfg.ExtraData[0] ^= 0xff

	again := dec.Config()
	if bytes.Equal(cfg.ExtraData, again.ExtraData) {
		t.Fatal("Config returned alias to decoder ExtraData")
	}
}

func TestADTSReaderTruncatedFrameDoesNotAdvance(t *testing.T) {
	header, err := appendTestADTSHeader(nil, Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}, 4)
	if err != nil {
		t.Fatal(err)
	}
	reader := NewADTSReader(bytes.NewReader(append(header, 1, 2)))
	_, err = reader.ReadFrame(nil)
	if !errors.Is(err, ErrNeedMoreData) {
		t.Fatalf("truncated frame err = %v, want ErrNeedMoreData", err)
	}
	if reader.FrameIndex() != 0 || reader.Offset() != 0 {
		t.Fatalf("truncated frame advanced reader: index=%d offset=%d", reader.FrameIndex(), reader.Offset())
	}
}

func TestADTSStreamDecodeErrorsIncludeFrameIndex(t *testing.T) {
	frame, err := appendTestADTSHeader(nil, Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}, 0)
	if err != nil {
		t.Fatal(err)
	}
	frame[2] &^= 0xc0

	_, _, err = DecodeADTSInto(nil, frame)
	if !errors.Is(err, ErrUnsupportedProfile) || !strings.Contains(err.Error(), "frame 0:") {
		t.Fatalf("DecodeADTSInto err = %v, want frame-indexed ErrUnsupportedProfile", err)
	}

	_, _, err = DecodeADTSReaderInto(nil, bytes.NewReader(frame))
	if !errors.Is(err, ErrUnsupportedProfile) || !strings.Contains(err.Error(), "frame 0:") {
		t.Fatalf("DecodeADTSReaderInto err = %v, want frame-indexed ErrUnsupportedProfile", err)
	}
}

func TestADTSFrameMetadataTransitions(t *testing.T) {
	m := loadTestVectorManifest(t)
	v := m.Vectors[0]
	frames, err := SplitADTSFrames(readTestVector(t, v.File))
	if err != nil {
		t.Fatal(err)
	}
	dec, err := New(Options{Transport: TransportADTS})
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()
	if cfg := dec.Config(); cfg.SampleRate != 0 || cfg.Channels != 0 {
		t.Fatalf("initial ADTS config = %+v, want zero value", cfg)
	}

	var pcm []int16
	for i, frame := range frames {
		before := len(pcm)
		var info FrameInfo
		pcm, info, err = dec.Decode(pcm, frame.Data)
		if err != nil {
			t.Fatalf("frame %d: %v", i, err)
		}
		if info.Transport != TransportADTS {
			t.Fatalf("frame %d transport = %s, want %s", i, info.Transport, TransportADTS)
		}
		if info.InputBytes != frame.Header.FrameLength {
			t.Fatalf("frame %d input bytes = %d, want %d", i, info.InputBytes, frame.Header.FrameLength)
		}
		if info.OutputSamples != len(pcm)-before {
			t.Fatalf("frame %d output samples = %d, want %d", i, info.OutputSamples, len(pcm)-before)
		}
		if info.Config.SampleRate != v.SampleRate || info.Config.Channels != v.Channels {
			t.Fatalf("frame %d config = %+v, want %d Hz/%d ch", i, info.Config, v.SampleRate, v.Channels)
		}
		if i == 0 && info.OutputSamples != 0 {
			t.Fatalf("first frame output samples = %d, want decoder-delay transition with 0", info.OutputSamples)
		}
		if i > 0 && info.OutputSamples == 0 {
			t.Fatalf("frame %d output samples = 0, want PCM after decoder-delay transition", i)
		}
	}
	assertPCMHash(t, v, pcm)
}

func newRawTestDecoder(t *testing.T) *Decoder {
	t.Helper()
	dec, err := New(Options{
		Transport: TransportRaw,
		Config: Config{
			ObjectType:    AOTAACLC,
			SampleRate:    44100,
			ChannelConfig: 2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return dec
}
