package aac

import (
	"fmt"
	"sync"
)

// FLVSoundFormat identifies the high four bits of an FLV audio tag header.
//
// RTMP audio messages carry this same FLV audio tag body after the RTMP chunk
// and message headers.
type FLVSoundFormat uint8

const (
	FLVSoundFormatLinearPCMPlatformEndian FLVSoundFormat = 0
	FLVSoundFormatADPCM                   FLVSoundFormat = 1
	FLVSoundFormatMP3                     FLVSoundFormat = 2
	FLVSoundFormatLinearPCMLittleEndian   FLVSoundFormat = 3
	FLVSoundFormatNellymoser16kHzMono     FLVSoundFormat = 4
	FLVSoundFormatNellymoser8kHzMono      FLVSoundFormat = 5
	FLVSoundFormatNellymoser              FLVSoundFormat = 6
	FLVSoundFormatG711ALaw                FLVSoundFormat = 7
	FLVSoundFormatG711MuLaw               FLVSoundFormat = 8
	FLVSoundFormatAAC                     FLVSoundFormat = 10
	FLVSoundFormatSpeex                   FLVSoundFormat = 11
	FLVSoundFormatMP3_8kHz                FLVSoundFormat = 14
	FLVSoundFormatDeviceSpecific          FLVSoundFormat = 15
)

const flvAACHeader = byte(FLVSoundFormatAAC<<4) |
	byte(FLVSoundRate44100Hz<<2) |
	byte(FLVSoundSize16Bit<<1) |
	byte(FLVSoundTypeStereo)

func (f FLVSoundFormat) String() string {
	switch f {
	case FLVSoundFormatLinearPCMPlatformEndian:
		return "Linear PCM"
	case FLVSoundFormatADPCM:
		return "ADPCM"
	case FLVSoundFormatMP3:
		return "MP3"
	case FLVSoundFormatLinearPCMLittleEndian:
		return "Linear PCM little-endian"
	case FLVSoundFormatNellymoser16kHzMono:
		return "Nellymoser 16 kHz mono"
	case FLVSoundFormatNellymoser8kHzMono:
		return "Nellymoser 8 kHz mono"
	case FLVSoundFormatNellymoser:
		return "Nellymoser"
	case FLVSoundFormatG711ALaw:
		return "G.711 A-law"
	case FLVSoundFormatG711MuLaw:
		return "G.711 mu-law"
	case FLVSoundFormatAAC:
		return "AAC"
	case FLVSoundFormatSpeex:
		return "Speex"
	case FLVSoundFormatMP3_8kHz:
		return "MP3 8 kHz"
	case FLVSoundFormatDeviceSpecific:
		return "device-specific"
	default:
		return fmt.Sprintf("FLV sound format %d", uint8(f))
	}
}

// FLVSoundRate is the low-level FLV sound-rate code. For AAC, decoders must
// use the MPEG-4 AudioSpecificConfig sample rate instead.
type FLVSoundRate uint8

const (
	FLVSoundRate5512Hz  FLVSoundRate = 0
	FLVSoundRate11025Hz FLVSoundRate = 1
	FLVSoundRate22050Hz FLVSoundRate = 2
	FLVSoundRate44100Hz FLVSoundRate = 3
)

// Hertz returns the nominal FLV rate for non-AAC audio, or zero for an unknown
// code.
func (r FLVSoundRate) Hertz() int {
	switch r {
	case FLVSoundRate5512Hz:
		return 5512
	case FLVSoundRate11025Hz:
		return 11025
	case FLVSoundRate22050Hz:
		return 22050
	case FLVSoundRate44100Hz:
		return 44100
	default:
		return 0
	}
}

// FLVSoundSize identifies the one-bit FLV sample-size code.
type FLVSoundSize uint8

const (
	FLVSoundSize8Bit  FLVSoundSize = 0
	FLVSoundSize16Bit FLVSoundSize = 1
)

// FLVSoundType identifies the one-bit FLV channel-shape code. For AAC, decoders
// must use the MPEG-4 AudioSpecificConfig channel configuration instead.
type FLVSoundType uint8

const (
	FLVSoundTypeMono   FLVSoundType = 0
	FLVSoundTypeStereo FLVSoundType = 1
)

// FLVAACPacketType identifies the byte following the FLV audio tag header for
// AAC sound tags.
type FLVAACPacketType uint8

const (
	FLVAACPacketTypeSequenceHeader FLVAACPacketType = 0
	FLVAACPacketTypeRaw            FLVAACPacketType = 1
)

func (p FLVAACPacketType) String() string {
	switch p {
	case FLVAACPacketTypeSequenceHeader:
		return "AAC sequence header"
	case FLVAACPacketTypeRaw:
		return "AAC raw"
	default:
		return fmt.Sprintf("AAC packet type %d", uint8(p))
	}
}

// FLVAudioTag is the parsed body of an FLV audio tag or RTMP audio message.
//
// Payload aliases the input passed to ParseFLVAudioTag. For AAC tags, Payload
// starts after AACPacketType and is either an MPEG-4 AudioSpecificConfig
// sequence header or one raw AAC access unit.
type FLVAudioTag struct {
	SoundFormat   FLVSoundFormat
	SoundRate     FLVSoundRate
	SoundSize     FLVSoundSize
	SoundType     FLVSoundType
	AACPacketType FLVAACPacketType
	Payload       []byte
}

// ParseFLVAudioTag parses the body of one FLV audio tag. RTMP audio message
// payloads use the same layout and can be passed here directly.
func ParseFLVAudioTag(data []byte) (FLVAudioTag, error) {
	if len(data) < 1 {
		return FLVAudioTag{}, ErrNeedMoreData
	}
	header := data[0]
	tag := FLVAudioTag{
		SoundFormat: FLVSoundFormat(header >> 4),
		SoundRate:   FLVSoundRate((header >> 2) & 0x03),
		SoundSize:   FLVSoundSize((header >> 1) & 0x01),
		SoundType:   FLVSoundType(header & 0x01),
	}
	if tag.SoundFormat != FLVSoundFormatAAC {
		tag.Payload = data[1:]
		return tag, nil
	}
	if len(data) < 2 {
		return FLVAudioTag{}, ErrNeedMoreData
	}
	tag.AACPacketType = FLVAACPacketType(data[1])
	tag.Payload = data[2:]
	return tag, nil
}

// ParseRTMPAudioMessage parses one RTMP audio message payload. It is an alias
// for ParseFLVAudioTag because RTMP carries FLV audio tag bodies.
func ParseRTMPAudioMessage(data []byte) (FLVAudioTag, error) {
	return ParseFLVAudioTag(data)
}

// AppendFLVAACSequenceHeader appends an FLV/RTMP AAC sequence-header audio tag
// body carrying cfg's MPEG-4 AudioSpecificConfig.
func AppendFLVAACSequenceHeader(dst []byte, cfg Config) ([]byte, error) {
	extra, err := cfg.AudioSpecificConfig()
	if err != nil {
		return dst, err
	}
	dst = append(dst, flvAACHeader, byte(FLVAACPacketTypeSequenceHeader))
	return append(dst, extra...), nil
}

// AppendRTMPAACSequenceHeader appends an RTMP AAC sequence-header message
// payload. RTMP audio messages carry FLV audio tag bodies, so this is an alias
// for AppendFLVAACSequenceHeader.
func AppendRTMPAACSequenceHeader(dst []byte, cfg Config) ([]byte, error) {
	return AppendFLVAACSequenceHeader(dst, cfg)
}

// AppendFLVAACRawTag appends an FLV/RTMP AAC raw audio tag body for one raw AAC
// access unit. The access unit is appended without an ADTS header.
func AppendFLVAACRawTag(dst, accessUnit []byte) []byte {
	dst = append(dst, flvAACHeader, byte(FLVAACPacketTypeRaw))
	return append(dst, accessUnit...)
}

// AppendRTMPAACRawMessage appends an RTMP AAC raw message payload. RTMP audio
// messages carry FLV audio tag bodies, so this is an alias for AppendFLVAACRawTag.
func AppendRTMPAACRawMessage(dst, accessUnit []byte) []byte {
	return AppendFLVAACRawTag(dst, accessUnit)
}

// FLVAACFrameInfo describes one decoded FLV/RTMP AAC audio tag.
type FLVAACFrameInfo struct {
	// Tag is the parsed FLV audio tag. Tag.Payload aliases the DecodeTag input.
	Tag FLVAudioTag
	// Config is the decoder configuration after this tag.
	Config Config
	// Frame describes the underlying raw AAC access unit for raw AAC tags.
	Frame FrameInfo
	// SequenceHeader is true when this tag carried MPEG-4 AudioSpecificConfig.
	SequenceHeader bool
	// InputBytes is the number of FLV/RTMP audio payload bytes consumed.
	InputBytes int
	// OutputSamples is the number of interleaved int16 samples appended.
	OutputSamples int
}

// FLVAACDecoder decodes AAC-LC carried in FLV audio tags or RTMP audio message
// payloads. A decoder must receive an AAC sequence-header tag before raw AAC
// tags. It is safe for concurrent use, although decode calls are serialized.
type FLVAACDecoder struct {
	mu     sync.Mutex
	dec    *Decoder
	cfg    Config
	closed bool
}

// NewFLVAACDecoder creates a stateful decoder for AAC-LC in FLV/RTMP audio
// messages.
func NewFLVAACDecoder() *FLVAACDecoder {
	return &FLVAACDecoder{}
}

// NewRTMPAACDecoder creates a stateful decoder for AAC-LC in RTMP audio message
// payloads.
func NewRTMPAACDecoder() *FLVAACDecoder {
	return NewFLVAACDecoder()
}

// Config returns a copy of the current stream configuration. Before the first
// AAC sequence header, the zero Config is returned.
func (d *FLVAACDecoder) Config() Config {
	if d == nil {
		return Config{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return copyConfig(d.cfg)
}

// DecodeTag parses one FLV audio tag body or RTMP audio message payload and
// appends decoded PCM to dst for raw AAC tags. AAC sequence headers update the
// decoder configuration and return dst unchanged.
func (d *FLVAACDecoder) DecodeTag(dst []int16, data []byte) ([]int16, FLVAACFrameInfo, error) {
	if d == nil {
		return dst, FLVAACFrameInfo{}, ErrClosed
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return dst, FLVAACFrameInfo{}, ErrClosed
	}

	tag, err := ParseFLVAudioTag(data)
	if err != nil {
		return dst, FLVAACFrameInfo{}, err
	}
	info := FLVAACFrameInfo{
		Tag:        tag,
		Config:     copyConfig(d.cfg),
		InputBytes: len(data),
	}
	if tag.SoundFormat != FLVSoundFormatAAC {
		return dst, info, fmt.Errorf("%w: %s", ErrUnsupportedFormat, tag.SoundFormat)
	}

	switch tag.AACPacketType {
	case FLVAACPacketTypeSequenceHeader:
		cfg, err := ParseAudioSpecificConfig(tag.Payload)
		if err != nil {
			return dst, info, err
		}
		dec, err := New(Options{Transport: TransportRaw, Config: cfg})
		if err != nil {
			return dst, info, err
		}
		nextCfg := dec.Config()
		if d.dec != nil {
			_ = d.dec.Close()
		}
		d.dec = dec
		d.cfg = nextCfg
		info.Config = copyConfig(d.cfg)
		info.SequenceHeader = true
		return dst, info, nil
	case FLVAACPacketTypeRaw:
		if d.dec == nil {
			return dst, info, fmt.Errorf("%w: missing AAC sequence header", ErrInvalidConfig)
		}
		out, frame, err := d.dec.DecodeRawInto(dst, tag.Payload)
		if err != nil {
			return dst, info, err
		}
		d.cfg = frame.Config
		info.Config = copyConfig(d.cfg)
		info.Frame = frame
		info.OutputSamples = frame.OutputSamples
		return out, info, nil
	default:
		return dst, info, fmt.Errorf("%w: %s", ErrInvalidFLV, tag.AACPacketType)
	}
}

// DecodeRTMPMessage decodes one RTMP audio message payload.
func (d *FLVAACDecoder) DecodeRTMPMessage(dst []int16, data []byte) ([]int16, FLVAACFrameInfo, error) {
	return d.DecodeTag(dst, data)
}

// Close releases decoder state. It is valid to call Close more than once.
func (d *FLVAACDecoder) Close() error {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	if d.dec != nil {
		if err := d.dec.Close(); err != nil {
			return err
		}
		d.dec = nil
	}
	return nil
}
