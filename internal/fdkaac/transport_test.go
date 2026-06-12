package fdkaac

import (
	"bytes"
	"testing"
)

func TestSamplingRateAndChannelConfig(t *testing.T) {
	if got := GetSamplingRateIndex(44100, 4); got != 4 {
		t.Fatalf("44100 index = %d, want 4", got)
	}
	if got := GetSamplingRateIndex(12345, 4); got != 15 {
		t.Fatalf("explicit index = %d, want 15", got)
	}
	if got := GetChannelConfig(Mode2, false); got != 2 {
		t.Fatalf("stereo channel config = %d, want 2", got)
	}
	if got := GetChannelConfig(Mode2, true); got != 0 {
		t.Fatalf("forced PCE channel config = %d, want 0", got)
	}
}

func TestAppendAudioSpecificConfigFDKVectors(t *testing.T) {
	// Expected bytes were checked against FDK-AAC v2.0.3
	// transportEnc_writeASC.
	tests := []struct {
		name       string
		sampleRate int
		channels   int
		want       []byte
	}{
		{name: "44100 stereo", sampleRate: 44100, channels: 2, want: []byte{0x12, 0x10}},
		{name: "48000 mono", sampleRate: 48000, channels: 1, want: []byte{0x11, 0x88}},
		{name: "explicit sample rate", sampleRate: 12345, channels: 2, want: []byte{0x17, 0x80, 0x18, 0x1c, 0x90}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := AACLCConfig(tt.sampleRate, tt.channels)
			if err != nil {
				t.Fatal(err)
			}
			got, err := AppendAudioSpecificConfig(nil, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("ASC = % x, want % x", got, tt.want)
			}
		})
	}
}

func TestAppendADTSHeaderFDKVectors(t *testing.T) {
	// Expected bytes were checked against FDK-AAC v2.0.3 adtsWrite_Init and
	// adtsWrite_EncodeHeader.
	tests := []struct {
		name         string
		sampleRate   int
		channels     int
		payloadBytes int
		fullness     int
		want         []byte
	}{
		{name: "44100 stereo", sampleRate: 44100, channels: 2, payloadBytes: 100, fullness: 0x7ff, want: []byte{0xff, 0xf1, 0x50, 0x80, 0x0d, 0x7f, 0xfc}},
		{name: "48000 mono", sampleRate: 48000, channels: 1, payloadBytes: 0, fullness: 0, want: []byte{0xff, 0xf1, 0x4c, 0x40, 0x00, 0xe0, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := AACLCConfig(tt.sampleRate, tt.channels)
			if err != nil {
				t.Fatal(err)
			}
			got, err := AppendADTSHeader(nil, cfg, tt.payloadBytes, tt.fullness)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("ADTS = % x, want % x", got, tt.want)
			}
		})
	}
}

func TestTransportRejectsUnsupportedBranches(t *testing.T) {
	cfg, err := AACLCConfig(44100, 2)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ChannelConfigZero = true
	if _, err := AppendAudioSpecificConfig(nil, cfg); err == nil {
		t.Fatal("ASC with PCE requirement succeeded")
	}

	cfg, err = AACLCConfig(12345, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendADTSHeader(nil, cfg, 0, 0x7ff); err == nil {
		t.Fatal("ADTS with explicit sample rate succeeded")
	}

	cfg, err = AACLCConfig(44100, 2)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Flags |= ConfigFlagProtection
	if _, err := AppendADTSHeader(nil, cfg, 0, 0x7ff); err == nil {
		t.Fatal("protected ADTS header succeeded")
	}
}

func TestTransportWriterAllocs(t *testing.T) {
	cfg, err := AACLCConfig(44100, 2)
	if err != nil {
		t.Fatal(err)
	}
	ascDst := make([]byte, 0, 8)
	adtsDst := make([]byte, 0, 8)
	var scratch TransportScratch
	allocs := testing.AllocsPerRun(1000, func() {
		var err error
		ascDst = ascDst[:0]
		adtsDst = adtsDst[:0]
		ascDst, err = AppendAudioSpecificConfigWithScratch(ascDst, cfg, &scratch)
		if err != nil {
			t.Fatal(err)
		}
		adtsDst, err = AppendADTSHeaderWithScratch(adtsDst, cfg, 100, 0x7ff, &scratch)
		if err != nil {
			t.Fatal(err)
		}
		if len(ascDst) != 2 || len(adtsDst) != 7 {
			t.Fatal("unexpected output length")
		}
	})
	if allocs != 0 {
		t.Fatalf("transport writer allocations = %v, want 0", allocs)
	}
}
