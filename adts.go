package aac

import "fmt"

const (
	ADTSHeaderSize    = 7
	ADTSHeaderSizeCRC = 9
	adtsMaxFrameBytes = (1 << 13) - 1
)

type ADTSHeader struct {
	MPEGVersion       int
	ProtectionAbsent  bool
	ObjectType        AudioObjectType
	SampleRateIndex   int
	SampleRate        int
	ChannelConfig     int
	Channels          int
	FrameLength       int
	HeaderLength      int
	PayloadLength     int
	BufferFullness    int
	RawDataBlockCount int
	OriginalCopy      bool
	Home              bool
	CopyrightID       bool
	CopyrightIDStart  bool
}

type ADTSFrame struct {
	Header ADTSHeader
	Data   []byte
}

func ParseADTSHeader(data []byte) (ADTSHeader, error) {
	if len(data) < ADTSHeaderSize {
		return ADTSHeader{}, ErrNeedMoreData
	}
	if data[0] != 0xff || data[1]&0xf0 != 0xf0 {
		return ADTSHeader{}, fmt.Errorf("%w: syncword", ErrInvalidADTS)
	}
	if (data[1]>>1)&0x03 != 0 {
		return ADTSHeader{}, fmt.Errorf("%w: layer", ErrInvalidADTS)
	}
	protectionAbsent := data[1]&0x01 != 0
	profile := (data[2] >> 6) & 0x03
	srIndex := int((data[2] >> 2) & 0x0f)
	if srIndex == 15 {
		return ADTSHeader{}, fmt.Errorf("%w: escape sample rate is illegal in ADTS", ErrInvalidADTS)
	}
	sampleRate, ok := SampleRateFromIndex(srIndex)
	if !ok {
		return ADTSHeader{}, fmt.Errorf("%w: sample rate index %d", ErrInvalidADTS, srIndex)
	}
	channelConfig := int(((data[2] & 0x01) << 2) | ((data[3] >> 6) & 0x03))
	channels, ok := ChannelsFromConfig(channelConfig)
	if !ok {
		return ADTSHeader{}, fmt.Errorf("%w: channel config %d", ErrInvalidADTS, channelConfig)
	}
	frameLength := int(data[3]&0x03)<<11 | int(data[4])<<3 | int(data[5]>>5)
	headerLength := ADTSHeaderSize
	if !protectionAbsent {
		headerLength = ADTSHeaderSizeCRC
	}
	if frameLength < headerLength {
		return ADTSHeader{}, fmt.Errorf("%w: frame length %d < header length %d", ErrInvalidADTS, frameLength, headerLength)
	}
	if len(data) < frameLength {
		return ADTSHeader{}, ErrNeedMoreData
	}
	rawBlocks := int(data[6] & 0x03)
	return ADTSHeader{
		MPEGVersion:       int((data[1] >> 3) & 0x01),
		ProtectionAbsent:  protectionAbsent,
		ObjectType:        AudioObjectType(profile + 1),
		SampleRateIndex:   srIndex,
		SampleRate:        sampleRate,
		ChannelConfig:     channelConfig,
		Channels:          channels,
		FrameLength:       frameLength,
		HeaderLength:      headerLength,
		PayloadLength:     frameLength - headerLength,
		BufferFullness:    int(data[5]&0x1f)<<6 | int(data[6]>>2),
		RawDataBlockCount: rawBlocks,
		OriginalCopy:      data[3]&0x20 != 0,
		Home:              data[3]&0x10 != 0,
		CopyrightID:       data[3]&0x08 != 0,
		CopyrightIDStart:  data[3]&0x04 != 0,
	}, nil
}

func SplitADTSFrames(data []byte) ([]ADTSFrame, error) {
	var frames []ADTSFrame
	for off := 0; off < len(data); {
		h, err := ParseADTSHeader(data[off:])
		if err != nil {
			return nil, fmt.Errorf("frame %d: %w", len(frames), err)
		}
		end := off + h.FrameLength
		frameData := data[off:end]
		frames = append(frames, ADTSFrame{Header: h, Data: frameData})
		off = end
	}
	return frames, nil
}

func AppendADTSHeader(dst []byte, cfg Config, payloadLen int) ([]byte, error) {
	cfg, err := normalizeRawConfig(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.SampleRateIndex == 15 {
		return nil, fmt.Errorf("%w: explicit sample rate is illegal in ADTS", ErrInvalidConfig)
	}
	if cfg.ChannelConfig < 0 || cfg.ChannelConfig > 7 {
		return nil, fmt.Errorf("%w: channel config %d", ErrInvalidConfig, cfg.ChannelConfig)
	}
	fullLen := ADTSHeaderSize + payloadLen
	if payloadLen < 0 || fullLen > adtsMaxFrameBytes {
		return nil, fmt.Errorf("%w: payload length %d", ErrInvalidADTS, payloadLen)
	}
	profile := int(cfg.ObjectType) - 1
	if profile < 0 || profile > 3 {
		return nil, fmt.Errorf("%w: %s cannot be signaled in ADTS", ErrUnsupportedProfile, cfg.ObjectType)
	}
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
