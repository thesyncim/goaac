package aac

import (
	"bytes"
	"errors"
	"strings"
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
	headerCfg := h.Config()
	if headerCfg.ObjectType != AOTAACLC || headerCfg.SampleRate != 44100 || headerCfg.SampleRateIndex != 4 || headerCfg.ChannelConfig != 2 || headerCfg.Channels != 2 {
		t.Fatalf("header config = %+v", headerCfg)
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
	if frames[0].PayloadOffset() != ADTSHeaderSize || frames[0].EndOffset() != int64(len(frame)) {
		t.Fatalf("frame byte range = payload %d end %d, want %d/%d", frames[0].PayloadOffset(), frames[0].EndOffset(), ADTSHeaderSize, len(frame))
	}
	payload := frames[0].Payload()
	if !bytes.Equal(payload, []byte{1, 2, 3, 4}) {
		t.Fatalf("payload = %v, want [1 2 3 4]", payload)
	}
	payload[0] = 9
	if frames[0].Data[ADTSHeaderSize] != 9 {
		t.Fatal("payload does not alias frame data")
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
	if frames[0].PayloadOffset() != ADTSHeaderSizeCRC || frames[0].EndOffset() != int64(len(frame)) {
		t.Fatalf("crc frame byte range = payload %d end %d, want %d/%d", frames[0].PayloadOffset(), frames[0].EndOffset(), ADTSHeaderSizeCRC, len(frame))
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
	if !bytes.Equal(got.Payload(), []byte{1, 2, 3}) {
		t.Fatalf("reader crc payload = %v, want [1 2 3]", got.Payload())
	}
}

func TestADTSLeadingID3v2Tag(t *testing.T) {
	cfg := Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}
	frame, err := appendTestADTSHeader(nil, cfg, 2)
	if err != nil {
		t.Fatal(err)
	}
	frame = append(frame, 1, 2)
	id3 := appendTestID3v2(nil, []byte("hello"))
	stream := append(append([]byte(nil), id3...), frame...)

	frames, err := SplitADTSFrames(stream)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames = %d, want 1", len(frames))
	}
	if frames[0].Index != 0 || frames[0].Offset != int64(len(id3)) {
		t.Fatalf("frame position = index %d offset %d, want 0/%d", frames[0].Index, frames[0].Offset, len(id3))
	}
	if frames[0].PayloadOffset() != int64(len(id3)+ADTSHeaderSize) || frames[0].EndOffset() != int64(len(stream)) {
		t.Fatalf("ID3 frame byte range = payload %d end %d, want %d/%d", frames[0].PayloadOffset(), frames[0].EndOffset(), len(id3)+ADTSHeaderSize, len(stream))
	}
	if !bytes.Equal(frames[0].Data, frame) {
		t.Fatal("split frame data does not match input frame")
	}

	reader := NewADTSReader(bytes.NewReader(stream))
	got, err := reader.ReadFrame(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Index != 0 || got.Offset != int64(len(id3)) {
		t.Fatalf("reader frame position = index %d offset %d, want 0/%d", got.Index, got.Offset, len(id3))
	}
	if reader.FrameIndex() != 1 || reader.Offset() != int64(len(stream)) {
		t.Fatalf("reader state = index %d offset %d, want 1/%d", reader.FrameIndex(), reader.Offset(), len(stream))
	}

	frame[2] &^= 0xc0
	badStream := append(append([]byte(nil), id3...), frame...)
	_, _, err = DecodeADTSInto(nil, badStream)
	if !errors.Is(err, ErrUnsupportedProfile) || !strings.Contains(err.Error(), "frame 0 at byte 15:") {
		t.Fatalf("DecodeADTSInto err = %v, want frame-positioned ErrUnsupportedProfile after ID3", err)
	}
}

func TestADTSRejectsInvalidLeadingID3v2Tag(t *testing.T) {
	invalid := []byte{'I', 'D', '3', 4, 0, 0, 0x80, 0, 0, 0}
	_, err := SplitADTSFrames(invalid)
	if !errors.Is(err, ErrInvalidADTS) || !strings.Contains(err.Error(), "frame 0 at byte 0:") {
		t.Fatalf("invalid ID3 err = %v, want frame-positioned ErrInvalidADTS", err)
	}

	_, err = NewADTSReader(bytes.NewReader(invalid)).ReadFrame(nil)
	if !errors.Is(err, ErrInvalidADTS) || !strings.Contains(err.Error(), "frame 0 at byte 0:") {
		t.Fatalf("reader invalid ID3 err = %v, want frame-positioned ErrInvalidADTS", err)
	}
}

func TestADTSFramePayloadRejectsInvalidBounds(t *testing.T) {
	frame := ADTSFrame{
		Header: ADTSHeader{HeaderLength: ADTSHeaderSize, FrameLength: ADTSHeaderSize + 1},
		Data:   make([]byte, ADTSHeaderSize),
	}
	if frame.Payload() != nil {
		t.Fatal("invalid frame bounds returned payload")
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

func appendTestID3v2(dst, payload []byte) []byte {
	size := len(payload)
	dst = append(dst,
		'I', 'D', '3',
		4, 0, 0,
		byte((size>>21)&0x7f),
		byte((size>>14)&0x7f),
		byte((size>>7)&0x7f),
		byte(size&0x7f),
	)
	return append(dst, payload...)
}

func TestADTSNeedMoreData(t *testing.T) {
	_, err := ParseADTSHeader([]byte{0xff, 0xf1})
	if !errors.Is(err, ErrNeedMoreData) {
		t.Fatalf("err = %v, want need more data", err)
	}
}

func TestSplitADTSFramesErrorsIncludeFramePosition(t *testing.T) {
	frame, err := appendTestADTSHeader(nil, Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = SplitADTSFrames(append(frame, 0xff, 0xf1))
	if !errors.Is(err, ErrNeedMoreData) || !strings.Contains(err.Error(), "frame 1 at byte 7:") {
		t.Fatalf("split err = %v, want frame-positioned ErrNeedMoreData", err)
	}
}

func TestADTSRejectEscapeSampleRate(t *testing.T) {
	h := []byte{0xff, 0xf1, 0x3c, 0x80, 0x01, 0x7f, 0xfc}
	_, err := ParseADTSHeader(h)
	if !errors.Is(err, ErrInvalidADTS) {
		t.Fatalf("err = %v, want invalid ADTS", err)
	}
}
