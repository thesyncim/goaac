package fdkaac

import "fmt"

type AudioObjectType int

const (
	AOTNullObject AudioObjectType = 0
	AOTAACMain    AudioObjectType = 1
	AOTAACLC      AudioObjectType = 2
	AOTAACSSR     AudioObjectType = 3
	AOTAACLTP     AudioObjectType = 4
	AOTEscape     AudioObjectType = 31
)

type ChannelMode int

const (
	ModeInvalid     ChannelMode = -1
	ModeUnknown     ChannelMode = 0
	Mode1           ChannelMode = 1
	Mode2           ChannelMode = 2
	Mode1_2         ChannelMode = 3
	Mode1_2_1       ChannelMode = 4
	Mode1_2_2       ChannelMode = 5
	Mode1_2_2_1     ChannelMode = 6
	Mode1_2_2_2_1   ChannelMode = 7
	Mode6_1         ChannelMode = 11
	Mode7_1Back     ChannelMode = 12
	Mode7_1TopFront ChannelMode = 14
)

const (
	ConfigFlagMPEGID     uint32 = 0x00100000
	ConfigFlagProtection uint32 = 0x00400000
)

var samplingRateTable = [...]uint32{
	96000, 88200, 64000, 48000, 44100, 32000, 24000, 22050,
	16000, 12000, 11025, 8000, 7350, 0, 0, 57600,
	51200, 40000, 38400, 34150, 28800, 25600, 20000, 19200,
	17075, 14400, 12800, 9600, 0, 0, 0, 0,
}

type CoderConfig struct {
	AOT               AudioObjectType
	ChannelMode       ChannelMode
	ChannelConfigZero bool
	SamplingRate      int
	SamplesPerFrame   int
	NSubFrames        int
	Flags             uint32
}

type TransportScratch struct {
	bitBuffer [64]byte
	fetch     [64]byte
	bitStream BitStream
}

func AACLCConfig(sampleRate, channels int) (CoderConfig, error) {
	mode := ChannelModeFromChannels(channels)
	if mode == ModeInvalid || channels > 2 {
		return CoderConfig{}, fmt.Errorf("fdkaac: unsupported AAC-LC channel count %d", channels)
	}
	return CoderConfig{
		AOT:             AOTAACLC,
		ChannelMode:     mode,
		SamplingRate:    sampleRate,
		SamplesPerFrame: 1024,
		NSubFrames:      1,
		Flags:           ConfigFlagMPEGID,
	}, nil
}

func ChannelModeFromChannels(channels int) ChannelMode {
	if channels <= 8 && channels > 0 {
		if channels == 8 {
			return Mode1_2_2_2_1
		}
		return ChannelMode(channels)
	}
	return ModeInvalid
}

func GetChannelConfig(channelMode ChannelMode, channelConfigZero bool) int {
	if channelConfigZero {
		return 0
	}
	switch channelMode {
	case Mode1:
		return 1
	case Mode2:
		return 2
	case Mode1_2:
		return 3
	case Mode1_2_1:
		return 4
	case Mode1_2_2:
		return 5
	case Mode1_2_2_1:
		return 6
	case Mode1_2_2_2_1:
		return 7
	case Mode6_1:
		return 11
	case Mode7_1Back:
		return 12
	case Mode7_1TopFront:
		return 14
	default:
		return 0
	}
}

func GetSamplingRateIndex(sampleRate int, nBits uint32) int {
	tableSize := uint32(1<<nBits) - 1
	for sfIndex := uint32(0); sfIndex < tableSize; sfIndex++ {
		if samplingRateTable[sfIndex] == uint32(sampleRate) {
			return int(sfIndex)
		}
	}
	return int(tableSize)
}

func AppendAudioSpecificConfig(dst []byte, cfg CoderConfig) ([]byte, error) {
	var scratch TransportScratch
	return AppendAudioSpecificConfigWithScratch(dst, cfg, &scratch)
}

func AppendAudioSpecificConfigWithScratch(dst []byte, cfg CoderConfig, scratch *TransportScratch) ([]byte, error) {
	if err := validateAACLCTransportConfig(cfg, true); err != nil {
		return dst, err
	}
	if scratch == nil {
		return dst, fmt.Errorf("fdkaac: nil transport scratch")
	}
	clear(scratch.bitBuffer[:])
	if err := InitBitStream(&scratch.bitStream, scratch.bitBuffer[:], 0, BSWriter); err != nil {
		return dst, err
	}
	writeAOT(&scratch.bitStream, cfg.AOT)
	writeSampleRate(&scratch.bitStream, cfg.SamplingRate, 4)
	WriteBits(&scratch.bitStream, uint32(GetChannelConfig(cfg.ChannelMode, cfg.ChannelConfigZero)), 4)
	writeGASpecificConfig(&scratch.bitStream, cfg, 0)
	n := FetchBuffer(&scratch.bitStream, scratch.fetch[:])
	return append(dst, scratch.fetch[:n]...), nil
}

func AppendADTSHeader(dst []byte, cfg CoderConfig, payloadBytes int, bufferFullness int) ([]byte, error) {
	var scratch TransportScratch
	return AppendADTSHeaderWithScratch(dst, cfg, payloadBytes, bufferFullness, &scratch)
}

func AppendADTSHeaderWithScratch(dst []byte, cfg CoderConfig, payloadBytes int, bufferFullness int, scratch *TransportScratch) ([]byte, error) {
	if err := validateAACLCTransportConfig(cfg, false); err != nil {
		return dst, err
	}
	if scratch == nil {
		return dst, fmt.Errorf("fdkaac: nil transport scratch")
	}
	if cfg.Flags&ConfigFlagProtection != 0 {
		return dst, fmt.Errorf("fdkaac: protected ADTS headers are not ported")
	}
	if payloadBytes < 0 {
		return dst, fmt.Errorf("fdkaac: invalid payload bytes %d", payloadBytes)
	}
	if bufferFullness < 0 {
		bufferFullness = 0x7ff
	}
	if bufferFullness >= 0x800 {
		return dst, fmt.Errorf("fdkaac: ADTS buffer fullness %d exceeds 11 bits", bufferFullness)
	}
	srIndex := GetSamplingRateIndex(cfg.SamplingRate, 4)
	if srIndex == 15 {
		return dst, fmt.Errorf("fdkaac: explicit sample rate is illegal in ADTS")
	}
	headerBits := 56
	frameLengthBits := payloadBytes << 3
	fullFrameBytes := (frameLengthBits + headerBits) >> 3
	if fullFrameBytes >= 0x2000 {
		return dst, fmt.Errorf("fdkaac: ADTS frame length %d exceeds 13 bits", fullFrameBytes)
	}

	clear(scratch.bitBuffer[:])
	if err := InitBitStream(&scratch.bitStream, scratch.bitBuffer[:], 0, BSWriter); err != nil {
		return dst, err
	}
	bs := &scratch.bitStream
	mpegID := uint32(1)
	if cfg.Flags&ConfigFlagMPEGID != 0 {
		mpegID = 0
	}
	WriteBits(bs, 0xfff, 12)
	WriteBits(bs, mpegID, 1)
	WriteBits(bs, 0, 2)
	WriteBits(bs, 1, 1)
	WriteBits(bs, uint32(cfg.AOT-1), 2)
	WriteBits(bs, uint32(srIndex), 4)
	WriteBits(bs, 0, 1)
	WriteBits(bs, uint32(GetChannelConfig(cfg.ChannelMode, cfg.ChannelConfigZero)), 3)
	WriteBits(bs, 0, 1)
	WriteBits(bs, 0, 1)
	WriteBits(bs, 0, 1)
	WriteBits(bs, 0, 1)
	WriteBits(bs, uint32(fullFrameBytes), 13)
	WriteBits(bs, uint32(bufferFullness), 11)
	WriteBits(bs, uint32(cfg.NSubFrames-1), 2)
	n := FetchBuffer(bs, scratch.fetch[:])
	return append(dst, scratch.fetch[:n]...), nil
}

func validateAACLCTransportConfig(cfg CoderConfig, allowExplicitSampleRate bool) error {
	if cfg.AOT != AOTAACLC {
		return fmt.Errorf("fdkaac: unsupported AOT %d", cfg.AOT)
	}
	if cfg.SamplingRate <= 0 {
		return fmt.Errorf("fdkaac: missing sampling rate")
	}
	if cfg.SamplesPerFrame == 0 {
		cfg.SamplesPerFrame = 1024
	}
	if cfg.SamplesPerFrame != 1024 && cfg.SamplesPerFrame != 960 {
		return fmt.Errorf("fdkaac: unsupported AAC-LC frame length %d", cfg.SamplesPerFrame)
	}
	if cfg.NSubFrames == 0 {
		cfg.NSubFrames = 1
	}
	if cfg.NSubFrames != 1 {
		return fmt.Errorf("fdkaac: ADTS subframing is not ported")
	}
	chConfig := GetChannelConfig(cfg.ChannelMode, cfg.ChannelConfigZero)
	if chConfig == 0 {
		return fmt.Errorf("fdkaac: program config element is not ported")
	}
	if !allowExplicitSampleRate && GetSamplingRateIndex(cfg.SamplingRate, 4) == 15 {
		return fmt.Errorf("fdkaac: explicit sample rate is illegal in ADTS")
	}
	return nil
}

func writeAOT(bs *BitStream, aot AudioObjectType) {
	tmp := int(aot)
	if tmp > 31 {
		WriteBits(bs, uint32(AOTEscape), 5)
		WriteBits(bs, uint32(tmp-32), 6)
		return
	}
	WriteBits(bs, uint32(tmp), 5)
}

func writeSampleRate(bs *BitStream, sampleRate int, nBits uint32) {
	srIndex := GetSamplingRateIndex(sampleRate, nBits)
	WriteBits(bs, uint32(srIndex), nBits)
	if srIndex == (1<<nBits)-1 {
		WriteBits(bs, uint32(sampleRate), 24)
	}
}

func writeGASpecificConfig(bs *BitStream, cfg CoderConfig, extFlag uint32) {
	frameLengthFlag := uint32(0)
	if cfg.SamplesPerFrame == 960 || cfg.SamplesPerFrame == 480 {
		frameLengthFlag = 1
	}
	WriteBits(bs, frameLengthFlag, 1)
	WriteBits(bs, 0, 1)
	WriteBits(bs, extFlag, 1)
}
