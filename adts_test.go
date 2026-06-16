package aac

import (
	"bytes"
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
	if frames[0].Index != 0 || frames[0].Offset != 0 {
		t.Fatalf("first frame position = index %d offset %d, want 0/0", frames[0].Index, frames[0].Offset)
	}
	if frames[1].Index != 1 || frames[1].Offset != int64(len(frame)) {
		t.Fatalf("second frame position = index %d offset %d, want 1/%d", frames[1].Index, frames[1].Offset, len(frame))
	}
}

func TestADTSCRCHeaderRoundTrip(t *testing.T) {
	cfg := Config{
		ObjectType:    AOTAACLC,
		SampleRate:    48000,
		ChannelConfig: 1,
	}
	frame, err := appendTestADTSCRCHeader(nil, cfg, 3)
	if err != nil {
		t.Fatal(err)
	}
	frame = append(frame, 0xaa, 0xbb, 1, 2, 3)

	h, err := ParseADTSHeader(frame)
	if err != nil {
		t.Fatal(err)
	}
	if h.ProtectionAbsent || h.HeaderLength != ADTSHeaderSizeCRC {
		t.Fatalf("crc header = protection absent %v header length %d", h.ProtectionAbsent, h.HeaderLength)
	}
	if h.FrameLength != len(frame) || h.PayloadLength != 3 || h.SampleRate != 48000 || h.Channels != 1 {
		t.Fatalf("crc header = %+v", h)
	}

	frames, err := SplitADTSFrames(append(frame, frame...))
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 2 || frames[0].Header.HeaderLength != ADTSHeaderSizeCRC {
		t.Fatalf("crc split frames = %d first=%+v", len(frames), frames[0].Header)
	}
	if frames[1].Index != 1 || frames[1].Offset != int64(len(frame)) {
		t.Fatalf("crc second frame position = index %d offset %d, want 1/%d", frames[1].Index, frames[1].Offset, len(frame))
	}

	reader := NewADTSReader(bytes.NewReader(frame))
	got, err := reader.ReadFrame(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Header.HeaderLength != ADTSHeaderSizeCRC || got.Header.PayloadLength != 3 {
		t.Fatalf("reader crc frame = %+v", got.Header)
	}
	if got.Index != 0 || got.Offset != 0 {
		t.Fatalf("reader crc position = index %d offset %d, want 0/0", got.Index, got.Offset)
	}
	if !bytes.Equal(got.Data, frame) {
		t.Fatal("reader returned frame data different from input")
	}
	if reader.FrameIndex() != 1 || reader.Offset() != int64(len(frame)) {
		t.Fatalf("reader state = index %d offset %d", reader.FrameIndex(), reader.Offset())
	}
}

func appendTestADTSHeader(dst []byte, cfg Config, payloadLen int) ([]byte, error) {
	return appendTestADTSHeaderWithProtection(dst, cfg, payloadLen, true)
}

func appendTestADTSCRCHeader(dst []byte, cfg Config, payloadLen int) ([]byte, error) {
	return appendTestADTSHeaderWithProtection(dst, cfg, payloadLen, false)
}

func appendTestADTSHeaderWithProtection(dst []byte, cfg Config, payloadLen int, protectionAbsent bool) ([]byte, error) {
	const maxFrameBytes = (1 << 13) - 1
	cfg, err := normalizeRawConfig(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.SampleRateIndex == 15 {
		return nil, ErrInvalidConfig
	}
	headerLen := ADTSHeaderSize
	if !protectionAbsent {
		headerLen = ADTSHeaderSizeCRC
	}
	fullLen := headerLen + payloadLen
	if payloadLen < 0 || fullLen > maxFrameBytes {
		return nil, ErrInvalidADTS
	}
	profile := int(cfg.ObjectType) - 1
	var h [ADTSHeaderSize]byte
	h[0] = 0xff
	h[1] = 0xf0
	if protectionAbsent {
		h[1] |= 0x01
	}
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
