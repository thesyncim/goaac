package aac

import (
	"fmt"
)

// AudioObjectType is an MPEG-4 audio object type.
type AudioObjectType int

const (
	AOTNull     AudioObjectType = 0
	AOTMain     AudioObjectType = 1
	AOTAACLC    AudioObjectType = 2
	AOTSSR      AudioObjectType = 3
	AOTLTP      AudioObjectType = 4
	AOTSBR      AudioObjectType = 5
	AOTScalable AudioObjectType = 6
	AOTTwinVQ   AudioObjectType = 7
	AOTCELP     AudioObjectType = 8
	AOTHVXC     AudioObjectType = 9
	AOTES       AudioObjectType = 12
	AOTPS       AudioObjectType = 29
	AOTEscape   AudioObjectType = 31
	AOTERAACLC  AudioObjectType = 17
	AOTERAACLD  AudioObjectType = 23
	AOTERAACELD AudioObjectType = 39
)

func (a AudioObjectType) String() string {
	switch a {
	case AOTMain:
		return "AAC Main"
	case AOTAACLC:
		return "AAC LC"
	case AOTSSR:
		return "AAC SSR"
	case AOTLTP:
		return "AAC LTP"
	case AOTSBR:
		return "SBR"
	case AOTPS:
		return "PS"
	case AOTERAACLC:
		return "ER AAC LC"
	case AOTERAACLD:
		return "ER AAC LD"
	case AOTERAACELD:
		return "ER AAC ELD"
	default:
		return fmt.Sprintf("AOT %d", int(a))
	}
}

var mpeg4AudioSampleRates = [...]int{
	96000,
	88200,
	64000,
	48000,
	44100,
	32000,
	24000,
	22050,
	16000,
	12000,
	11025,
	8000,
	7350,
}

var mpeg4AudioChannels = [...]int{
	0,
	1,
	2,
	3,
	4,
	5,
	6,
	8,
	0,
	0,
	0,
	7,
	8,
	24,
	8,
}

// Config describes an MPEG-4 AudioSpecificConfig and decoder output shape.
//
// For raw AAC-LC decoding, pass either ExtraData containing an
// AudioSpecificConfig, or ObjectType/SampleRate/ChannelConfig (or Channels)
// so the decoder can build one.
type Config struct {
	ObjectType           AudioObjectType
	SampleRate           int
	SampleRateIndex      int
	ChannelConfig        int
	Channels             int
	ExtensionObjectType  AudioObjectType
	ExtensionSampleRate  int
	ExtensionSampleIndex int
	SBR                  bool
	PS                   bool
	ExtraData            []byte
}

// ParseAudioSpecificConfig parses MPEG-4 AudioSpecificConfig bytes.
func ParseAudioSpecificConfig(data []byte) (Config, error) {
	if len(data) == 0 {
		return Config{}, ErrInvalidConfig
	}
	r := newBitReader(data)
	obj, err := readObjectType(&r)
	if err != nil {
		return Config{}, fmt.Errorf("%w: object type: %v", ErrInvalidConfig, err)
	}
	sr, srIndex, err := readSampleRate(&r)
	if err != nil {
		return Config{}, fmt.Errorf("%w: sample rate: %v", ErrInvalidConfig, err)
	}
	chConfigBits, err := r.readBits(4)
	if err != nil {
		return Config{}, fmt.Errorf("%w: channel config: %v", ErrInvalidConfig, err)
	}
	chConfig := int(chConfigBits)
	if chConfig >= len(mpeg4AudioChannels) {
		return Config{}, fmt.Errorf("%w: channel config %d", ErrInvalidConfig, chConfig)
	}
	cfg := Config{
		ObjectType:      obj,
		SampleRate:      sr,
		SampleRateIndex: srIndex,
		ChannelConfig:   chConfig,
		Channels:        mpeg4AudioChannels[chConfig],
		ExtraData:       append([]byte(nil), data...),
	}
	if obj == AOTSBR || obj == AOTPS {
		cfg.ExtensionObjectType = AOTSBR
		cfg.SBR = true
		cfg.PS = obj == AOTPS
		extSR, extIndex, err := readSampleRate(&r)
		if err != nil {
			return Config{}, fmt.Errorf("%w: extension sample rate: %v", ErrInvalidConfig, err)
		}
		cfg.ExtensionSampleRate = extSR
		cfg.ExtensionSampleIndex = extIndex
		obj, err = readObjectType(&r)
		if err != nil {
			return Config{}, fmt.Errorf("%w: extension object type: %v", ErrInvalidConfig, err)
		}
		cfg.ObjectType = obj
	}
	return cfg, nil
}

// AudioSpecificConfig serializes c as MPEG-4 AudioSpecificConfig bytes.
func (c Config) AudioSpecificConfig() ([]byte, error) {
	cfg, err := normalizeRawConfig(c)
	if err != nil {
		return nil, err
	}
	return buildAudioSpecificConfig(cfg), nil
}

func buildAudioSpecificConfig(cfg Config) []byte {
	w := bitWriter{}
	writeObjectType(&w, cfg.ObjectType)
	if cfg.SampleRateIndex == 15 {
		w.writeBits(4, 15)
		w.writeBits(24, uint32(cfg.SampleRate))
	} else {
		w.writeBits(4, uint32(cfg.SampleRateIndex))
	}
	w.writeBits(4, uint32(cfg.ChannelConfig))
	w.writeBits(1, 0) // frameLengthFlag: 1024 samples.
	w.writeBits(1, 0) // dependsOnCoreCoder.
	w.writeBits(1, 0) // extensionFlag.
	return w.bytes()
}

// SampleRateIndex returns the MPEG-4 sampling-frequency index for rate.
func SampleRateIndex(rate int) (int, bool) {
	for i, v := range mpeg4AudioSampleRates {
		if v == rate {
			return i, true
		}
	}
	return 15, false
}

// SampleRateFromIndex returns the sampling frequency for a MPEG-4 index.
func SampleRateFromIndex(index int) (int, bool) {
	if index < 0 || index >= len(mpeg4AudioSampleRates) {
		return 0, false
	}
	return mpeg4AudioSampleRates[index], true
}

// ChannelsFromConfig returns the channel count for a channel configuration.
func ChannelsFromConfig(config int) (int, bool) {
	if config < 0 || config >= len(mpeg4AudioChannels) {
		return 0, false
	}
	return mpeg4AudioChannels[config], true
}

// ChannelConfigForChannels returns a channel configuration for common layouts.
func ChannelConfigForChannels(channels int) (int, bool) {
	for i, n := range mpeg4AudioChannels {
		if i != 0 && n == channels {
			return i, true
		}
	}
	return 0, false
}

func normalizeRawConfig(c Config) (Config, error) {
	if len(c.ExtraData) > 0 {
		parsed, err := ParseAudioSpecificConfig(c.ExtraData)
		if err != nil {
			return Config{}, err
		}
		if parsed.ObjectType != AOTAACLC || parsed.SBR || parsed.PS {
			return Config{}, fmt.Errorf("%w: %s", ErrUnsupportedProfile, parsed.ObjectType)
		}
		return parsed, nil
	}
	if c.ObjectType == 0 {
		c.ObjectType = AOTAACLC
	}
	if c.ObjectType != AOTAACLC {
		return Config{}, fmt.Errorf("%w: %s", ErrUnsupportedProfile, c.ObjectType)
	}
	if c.SampleRate <= 0 {
		return Config{}, fmt.Errorf("%w: missing sample rate", ErrInvalidConfig)
	}
	if c.SampleRateIndex == 0 {
		if idx, ok := SampleRateIndex(c.SampleRate); ok {
			c.SampleRateIndex = idx
		}
	}
	if c.SampleRateIndex < 0 || c.SampleRateIndex > 15 {
		return Config{}, fmt.Errorf("%w: sample rate index %d", ErrInvalidConfig, c.SampleRateIndex)
	}
	if c.SampleRateIndex != 15 {
		sr, ok := SampleRateFromIndex(c.SampleRateIndex)
		if !ok {
			return Config{}, fmt.Errorf("%w: sample rate index %d", ErrInvalidConfig, c.SampleRateIndex)
		}
		if c.SampleRate != sr {
			return Config{}, fmt.Errorf("%w: sample rate/index mismatch", ErrInvalidConfig)
		}
	}
	if c.ChannelConfig == 0 && c.Channels > 0 {
		chConfig, ok := ChannelConfigForChannels(c.Channels)
		if !ok {
			return Config{}, fmt.Errorf("%w: channels %d require program config element", ErrInvalidConfig, c.Channels)
		}
		c.ChannelConfig = chConfig
	}
	channels, ok := ChannelsFromConfig(c.ChannelConfig)
	if !ok {
		return Config{}, fmt.Errorf("%w: channel config %d", ErrInvalidConfig, c.ChannelConfig)
	}
	if c.ChannelConfig != 0 && c.Channels != 0 && c.Channels != channels {
		return Config{}, fmt.Errorf("%w: channel config/channels mismatch", ErrInvalidConfig)
	}
	c.Channels = channels
	c.ExtraData = buildAudioSpecificConfig(c)
	return c, nil
}

func readObjectType(r *bitReader) (AudioObjectType, error) {
	v, err := r.readBits(5)
	if err != nil {
		return 0, err
	}
	if AudioObjectType(v) == AOTEscape {
		ext, err := r.readBits(6)
		if err != nil {
			return 0, err
		}
		return AudioObjectType(32 + ext), nil
	}
	return AudioObjectType(v), nil
}

func writeObjectType(w *bitWriter, obj AudioObjectType) {
	if obj >= 32 {
		w.writeBits(5, uint32(AOTEscape))
		w.writeBits(6, uint32(obj-32))
		return
	}
	w.writeBits(5, uint32(obj))
}

func readSampleRate(r *bitReader) (rate int, index int, err error) {
	v, err := r.readBits(4)
	if err != nil {
		return 0, 0, err
	}
	index = int(v)
	if index == 15 {
		explicit, err := r.readBits(24)
		if err != nil {
			return 0, 0, err
		}
		return int(explicit), index, nil
	}
	if index >= len(mpeg4AudioSampleRates) {
		return 0, 0, fmt.Errorf("index %d", index)
	}
	return mpeg4AudioSampleRates[index], index, nil
}
