package aac

import (
	"errors"
	"testing"
)

func TestParseAudioSpecificConfigAACLC(t *testing.T) {
	cfg, err := ParseAudioSpecificConfig([]byte{0x12, 0x10})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ObjectType != AOTAACLC {
		t.Fatalf("object type = %v, want AAC LC", cfg.ObjectType)
	}
	if cfg.SampleRate != 44100 || cfg.SampleRateIndex != 4 {
		t.Fatalf("sample rate = %d index %d, want 44100 index 4", cfg.SampleRate, cfg.SampleRateIndex)
	}
	if cfg.ChannelConfig != 2 || cfg.Channels != 2 {
		t.Fatalf("channels = config %d count %d, want stereo", cfg.ChannelConfig, cfg.Channels)
	}
}

func TestBuildAudioSpecificConfigAACLC(t *testing.T) {
	cfg := Config{
		ObjectType:    AOTAACLC,
		SampleRate:    44100,
		ChannelConfig: 2,
	}
	extra, err := cfg.AudioSpecificConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := extra, []byte{0x12, 0x10}; string(got) != string(want) {
		t.Fatalf("ASC = % x, want % x", got, want)
	}
	parsed, err := ParseAudioSpecificConfig(extra)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ObjectType != AOTAACLC || parsed.SampleRate != 44100 || parsed.Channels != 2 {
		t.Fatalf("round trip = %+v", parsed)
	}
}

func TestRejectSBRConfig(t *testing.T) {
	_, err := NewDecoder(Config{ExtraData: []byte{0x2b, 0x92, 0x08}})
	if !errors.Is(err, ErrUnsupportedProfile) {
		t.Fatalf("err = %v, want unsupported profile", err)
	}
}
