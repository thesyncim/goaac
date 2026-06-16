package aac

import "fmt"

const (
	ADTSHeaderSize    = 7
	ADTSHeaderSizeCRC = 9
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

// ADTSFrame is one ADTS frame and its parsed header. Data aliases the input
// stream bytes passed to the parser or reader.
type ADTSFrame struct {
	Header ADTSHeader
	Data   []byte
}

// ParseADTSHeader parses one complete ADTS frame header from data.
func ParseADTSHeader(data []byte) (ADTSHeader, error) {
	return parseADTSHeader(data, true)
}

func parseADTSHeaderPrefix(data []byte) (ADTSHeader, error) {
	return parseADTSHeader(data, false)
}

func parseADTSHeader(data []byte, requireFullFrame bool) (ADTSHeader, error) {
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
	if requireFullFrame && len(data) < frameLength {
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

// SplitADTSFrames splits a complete ADTS stream into frame slices that alias
// data.
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
