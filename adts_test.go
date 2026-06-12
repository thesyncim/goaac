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
	frame, err := AppendADTSHeader(nil, cfg, 4)
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
