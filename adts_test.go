package aac

import (
	"errors"
	"testing"
)

func TestADTSHeaderRoundTrip(t *testing.T) {
	cfg := Config{
		ObjectType:    AOTAACLC,
		SampleRate:    44100,
		ChannelConfig: 2,
	}
	frame, err := appendTestADTSHeader(nil, cfg, 4)
	if err != nil {
		t.Fatal(err)
	}
	frame = append(frame, 1, 2, 3, 4)
	h, err := ParseADTSHeader(frame)
	if err != nil {
		t.Fatal(err)
	}
	if h.ObjectType != AOTAACLC || h.SampleRate != 44100 || h.Channels != 2 {
		t.Fatalf("header = %+v", h)
	}
	if h.FrameLength != 11 || h.PayloadLength != 4 || h.HeaderLength != 7 {
		t.Fatalf("lengths = frame %d payload %d header %d", h.FrameLength, h.PayloadLength, h.HeaderLength)
	}
	frames, err := SplitADTSFrames(append(frame, frame...))
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 2 {
		t.Fatalf("frames = %d, want 2", len(frames))
	}
}

func appendTestADTSHeader(dst []byte, cfg Config, payloadLen int) ([]byte, error) {
	const maxFrameBytes = (1 << 13) - 1
	cfg, err := normalizeRawConfig(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.SampleRateIndex == 15 {
		return nil, ErrInvalidConfig
	}
	fullLen := ADTSHeaderSize + payloadLen
	if payloadLen < 0 || fullLen > maxFrameBytes {
		return nil, ErrInvalidADTS
	}
	profile := int(cfg.ObjectType) - 1
	var h [ADTSHeaderSize]byte
	h[0] = 0xff
	h[1] = 0xf1
	h[2] = byte(profile<<6) | byte(cfg.SampleRateIndex<<2) | byte((cfg.ChannelConfig>>2)&1)
	h[3] = byte((cfg.ChannelConfig&3)<<6) | byte((fullLen>>11)&0x03)
	h[4] = byte(fullLen >> 3)
	h[5] = byte((fullLen&7)<<5) | 0x1f
	h[6] = 0xfc
	return append(dst, h[:]...), nil
}

func TestADTSNeedMoreData(t *testing.T) {
	_, err := ParseADTSHeader([]byte{0xff, 0xf1})
	if !errors.Is(err, ErrNeedMoreData) {
		t.Fatalf("err = %v, want need more data", err)
	}
}

func TestADTSRejectEscapeSampleRate(t *testing.T) {
	h := []byte{0xff, 0xf1, 0x3c, 0x80, 0x01, 0x7f, 0xfc}
	_, err := ParseADTSHeader(h)
	if !errors.Is(err, ErrInvalidADTS) {
		t.Fatalf("err = %v, want invalid ADTS", err)
	}
}
