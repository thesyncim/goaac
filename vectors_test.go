package aac

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type testVectorManifest struct {
	Version int          `json:"version"`
	Source  string       `json:"source"`
	Oracle  string       `json:"oracle"`
	Vectors []testVector `json:"vectors"`
}

type testVector struct {
	Name        string `json:"name"`
	File        string `json:"file"`
	Description string `json:"description"`
	SampleRate  int    `json:"sample_rate"`
	Channels    int    `json:"channels"`
	Frames      int    `json:"frames"`
	InputBytes  int    `json:"input_bytes"`
	InputSHA256 string `json:"input_sha256"`
	PCMBytes    int    `json:"pcm_bytes"`
	PCMSHA256   string `json:"pcm_sha256"`
}

func TestCommittedAACLCVectors(t *testing.T) {
	m := loadTestVectorManifest(t)
	if m.Version != 1 {
		t.Fatalf("manifest version = %d, want 1", m.Version)
	}
	if len(m.Vectors) < 4 {
		t.Fatalf("manifest vectors = %d, want at least 4", len(m.Vectors))
	}

	for _, v := range m.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			data := readTestVector(t, v.File)
			if got := len(data); got != v.InputBytes {
				t.Fatalf("input bytes = %d, want %d", got, v.InputBytes)
			}
			if got := sha256Hex(data); got != v.InputSHA256 {
				t.Fatalf("input sha256 = %s, want %s", got, v.InputSHA256)
			}

			frames, err := SplitADTSFrames(data)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(frames); got != v.Frames {
				t.Fatalf("frames = %d, want %d", got, v.Frames)
			}
			assertVectorHeaders(t, v, frames)

			pcm, cfg, err := DecodeADTS(data)
			if err != nil {
				t.Fatal(err)
			}
			assertVectorConfig(t, v, cfg)
			assertPCMHash(t, v, pcm)

			streamPCM, streamCfg, err := DecodeADTSReader(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			assertVectorConfig(t, v, streamCfg)
			assertPCMHash(t, v, streamPCM)

			var intoPrefix = []int16{-123, 456}
			intoPCM, intoCfg, err := DecodeADTSInto(append([]int16(nil), intoPrefix...), data)
			if err != nil {
				t.Fatal(err)
			}
			assertVectorConfig(t, v, intoCfg)
			if !equalInt16s(intoPCM[:len(intoPrefix)], intoPrefix) {
				t.Fatalf("DecodeADTSInto modified prefix: got %v, want %v", intoPCM[:len(intoPrefix)], intoPrefix)
			}
			assertPCMHash(t, v, intoPCM[len(intoPrefix):])

			byFramePCM := decodeVectorByADTSFrame(t, frames)
			assertPCMHash(t, v, byFramePCM)

			rawPCM := decodeVectorAsRawPayloads(t, frames)
			assertPCMHash(t, v, rawPCM)
		})
	}
}

func loadTestVectorManifest(t *testing.T) testVectorManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "aaclc", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m testVectorManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func readTestVector(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "aaclc", file))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertVectorHeaders(t *testing.T, v testVector, frames []ADTSFrame) {
	t.Helper()
	total := 0
	for i, frame := range frames {
		h := frame.Header
		if h.ObjectType != AOTAACLC {
			t.Fatalf("frame %d object type = %s, want %s", i, h.ObjectType, AOTAACLC)
		}
		if h.SampleRate != v.SampleRate || h.Channels != v.Channels {
			t.Fatalf("frame %d config = %d Hz/%d ch, want %d Hz/%d ch", i, h.SampleRate, h.Channels, v.SampleRate, v.Channels)
		}
		if h.RawDataBlockCount != 0 {
			t.Fatalf("frame %d raw data blocks = %d, want 0", i, h.RawDataBlockCount+1)
		}
		if len(frame.Data) != h.FrameLength {
			t.Fatalf("frame %d data len = %d, want %d", i, len(frame.Data), h.FrameLength)
		}
		total += h.FrameLength
	}
	if total != v.InputBytes {
		t.Fatalf("ADTS frame bytes = %d, want %d", total, v.InputBytes)
	}
}

func assertVectorConfig(t *testing.T, v testVector, cfg Config) {
	t.Helper()
	if cfg.ObjectType != AOTAACLC || cfg.SampleRate != v.SampleRate || cfg.Channels != v.Channels {
		t.Fatalf("config = %+v, want AAC-LC %d Hz/%d ch", cfg, v.SampleRate, v.Channels)
	}
}

func assertPCMHash(t *testing.T, v testVector, samples []int16) {
	t.Helper()
	pcm := int16sToLE(samples)
	if got := len(pcm); got != v.PCMBytes {
		t.Fatalf("PCM bytes = %d, want %d", got, v.PCMBytes)
	}
	if got := sha256Hex(pcm); got != v.PCMSHA256 {
		t.Fatalf("PCM sha256 = %s, want %s", got, v.PCMSHA256)
	}
}

func decodeVectorByADTSFrame(t *testing.T, frames []ADTSFrame) []int16 {
	t.Helper()
	dec, err := New(Options{Transport: TransportADTS})
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()

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
		if info.InputBytes != len(frame.Data) {
			t.Fatalf("frame %d input bytes = %d, want %d", i, info.InputBytes, len(frame.Data))
		}
		if info.OutputSamples != len(pcm)-before {
			t.Fatalf("frame %d output samples = %d, want %d", i, info.OutputSamples, len(pcm)-before)
		}
	}
	return pcm
}

func decodeVectorAsRawPayloads(t *testing.T, frames []ADTSFrame) []int16 {
	t.Helper()
	if len(frames) == 0 {
		t.Fatal("no frames")
	}
	h := frames[0].Header
	dec, err := New(Options{
		Transport: TransportRaw,
		Config: Config{
			ObjectType:      AOTAACLC,
			SampleRate:      h.SampleRate,
			SampleRateIndex: h.SampleRateIndex,
			ChannelConfig:   h.ChannelConfig,
			Channels:        h.Channels,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()

	var pcm []int16
	for i, frame := range frames {
		h := frame.Header
		if h.HeaderLength > len(frame.Data) || h.FrameLength > len(frame.Data) {
			t.Fatalf("frame %d invalid payload bounds", i)
		}
		before := len(pcm)
		var info FrameInfo
		pcm, info, err = dec.DecodeRawInto(pcm, frame.Data[h.HeaderLength:h.FrameLength])
		if err != nil {
			t.Fatalf("raw frame %d: %v", i, err)
		}
		if info.Transport != TransportRaw {
			t.Fatalf("raw frame %d transport = %s, want %s", i, info.Transport, TransportRaw)
		}
		if info.OutputSamples != len(pcm)-before {
			t.Fatalf("raw frame %d output samples = %d, want %d", i, info.OutputSamples, len(pcm)-before)
		}
	}
	return pcm
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func equalInt16s(a, b []int16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
