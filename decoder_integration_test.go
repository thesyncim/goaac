package aac

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDecodeADTSMatchesFFmpegOracle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg CLI not available")
	}
	tmp := t.TempDir()
	input := filepath.Join(tmp, "sine.aac")
	oracle := filepath.Join(tmp, "oracle.pcm")

	run(t, ffmpeg,
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi",
		"-i", "sine=frequency=997:duration=0.10:sample_rate=44100",
		"-ac", "2",
		"-c:a", "aac",
		"-profile:a", "aac_low",
		"-b:a", "96k",
		"-f", "adts",
		input,
	)
	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	pcm, cfg, err := DecodeADTS(data)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ObjectType != AOTAACLC || cfg.SampleRate != 44100 || cfg.Channels != 2 {
		t.Fatalf("decoded config = %+v", cfg)
	}
	if len(pcm) == 0 {
		t.Fatal("decoded no PCM")
	}

	run(t, ffmpeg,
		"-hide_banner", "-loglevel", "error",
		"-i", input,
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		oracle,
	)
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	got := int16sToLE(pcm)
	if !bytes.Equal(got, want) {
		t.Fatalf("PCM mismatch: got %d bytes, want %d bytes", len(got), len(want))
	}
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func int16sToLE(samples []int16) []byte {
	out := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(s))
	}
	return out
}
