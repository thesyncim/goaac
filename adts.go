package aac

import "fmt"

const (
	ADTSHeaderSize    = 7
	ADTSHeaderSizeCRC = 9
	id3v2HeaderSize   = 10
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

// Config returns the MPEG-4 decoder configuration signaled by the ADTS header.
func (h ADTSHeader) Config() Config {
	return Config{
		ObjectType:      h.ObjectType,
		SampleRate:      h.SampleRate,
		SampleRateIndex: h.SampleRateIndex,
		ChannelConfig:   h.ChannelConfig,
		Channels:        h.Channels,
	}
}

// ADTSFrame is one ADTS frame and its parsed header. Data aliases the input
// stream bytes passed to the parser or reader.
type ADTSFrame struct {
	Header ADTSHeader
	Data   []byte
	Index  int
	Offset int64
}

// EndOffset returns the byte offset immediately after this frame in the source
// stream.
func (f ADTSFrame) EndOffset() int64 {
	if f.Header.FrameLength < 0 {
		return f.Offset
	}
	return f.Offset + int64(f.Header.FrameLength)
}

// PayloadOffset returns the byte offset of the raw AAC access unit in the
// source stream.
func (f ADTSFrame) PayloadOffset() int64 {
	if f.Header.HeaderLength < 0 {
		return f.Offset
	}
	return f.Offset + int64(f.Header.HeaderLength)
}

// Payload returns the raw AAC access unit bytes after the ADTS header and
// optional CRC field. The returned slice aliases Data.
func (f ADTSFrame) Payload() []byte {
	h := f.Header
	if h.HeaderLength < 0 || h.FrameLength < h.HeaderLength || h.FrameLength > len(f.Data) {
		return nil
	}
	return f.Data[h.HeaderLength:h.FrameLength]
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
	off, err := skipLeadingID3v2(data)
	if err != nil {
		return nil, adtsFrameError(0, int64(off), err)
	}
	for off < len(data) {
		h, err := ParseADTSHeader(data[off:])
		if err != nil {
			return nil, adtsFrameError(len(frames), int64(off), err)
		}
		end := off + h.FrameLength
		frameData := data[off:end]
		frames = append(frames, ADTSFrame{
			Header: h,
			Data:   frameData,
			Index:  len(frames),
			Offset: int64(off),
		})
		off = end
	}
	return frames, nil
}

func adtsFrameError(index int, offset int64, err error) error {
	return fmt.Errorf("frame %d at byte %d: %w", index, offset, err)
}

func skipLeadingID3v2(data []byte) (int, error) {
	off := 0
	for hasID3v2Prefix(data[off:]) {
		if len(data)-off < id3v2HeaderSize {
			return off, ErrNeedMoreData
		}
		tagLen, ok := id3v2TagLength(data[off : off+id3v2HeaderSize])
		if !ok {
			return off, fmt.Errorf("%w: invalid ID3v2 tag", ErrInvalidADTS)
		}
		if len(data)-off < tagLen {
			return off, ErrNeedMoreData
		}
		off += tagLen
	}
	return off, nil
}

func hasID3v2Prefix(data []byte) bool {
	return len(data) >= 3 && data[0] == 'I' && data[1] == 'D' && data[2] == '3'
}

func id3v2TagLength(header []byte) (int, bool) {
	if len(header) < id3v2HeaderSize || !hasID3v2Prefix(header) {
		return 0, false
	}
	if header[3] < 2 || header[3] > 4 || header[4] == 0xff {
		return 0, false
	}
	for _, b := range header[6:10] {
		if b&0x80 != 0 {
			return 0, false
		}
	}
	size := int(header[6])<<21 | int(header[7])<<14 | int(header[8])<<7 | int(header[9])
	tagLen := id3v2HeaderSize + size
	if header[5]&0x10 != 0 {
		tagLen += id3v2HeaderSize
	}
	return tagLen, true
}
