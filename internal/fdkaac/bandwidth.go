package fdkaac

type ElementMode int

const (
	ElementModeInvalid ElementMode = 0
	ElementModeMono    ElementMode = 1
	ElementModeStereo  ElementMode = 2
)

type ChannelModeConfig struct {
	EncMode      ChannelMode
	NChannels    int
	NChannelsEff int
	NElements    int
}

type bandwidthTab struct {
	ChanBitRate          int
	BandWidthMono        int
	BandWidth2AndMoreCha int
}

type bandwidthTabVBR struct {
	BitrateMode          QCBitrateMode
	BandWidthMono        int
	BandWidth2AndMoreCha int
}

type configTabEntryVBR struct {
	BitrateMode QCBitrateMode
	ChanBitrate [2]int
}

var channelModeConfig = [...]ChannelModeConfig{
	{Mode1, 1, 1, 1},
	{Mode2, 2, 2, 1},
	{Mode1_2, 3, 3, 2},
	{Mode1_2_1, 4, 4, 3},
	{Mode1_2_2, 5, 5, 3},
	{Mode1_2_2_1, 6, 5, 4},
	{Mode1_2_2_2_1, 8, 7, 5},
	{Mode6_1, 7, 6, 5},
	{Mode7_1Back, 8, 7, 5},
	{Mode7_1TopFront, 8, 7, 5},
	{Mode7_1RearSurround, 8, 7, 5},
	{Mode7_1FrontCenter, 8, 7, 5},
}

var bandWidthTable = [...]bandwidthTab{
	{0, 3700, 5000},
	{12000, 5000, 6400},
	{20000, 6900, 9640},
	{28000, 9600, 13050},
	{40000, 12060, 14260},
	{56000, 13950, 15500},
	{72000, 14200, 16120},
	{96000, 17000, 17000},
	{576001, 17000, 17000},
}

var bandWidthTableLD22050 = [...]bandwidthTab{
	{8000, 2000, 2400},
	{12000, 2500, 2700},
	{16000, 3300, 3100},
	{24000, 6250, 7200},
	{32000, 9200, 10500},
	{40000, 16000, 16000},
	{48000, 16000, 16000},
	{282241, 16000, 16000},
}

var bandWidthTableLD24000 = [...]bandwidthTab{
	{8000, 2000, 2000},
	{12000, 2000, 2300},
	{16000, 2200, 2500},
	{24000, 5650, 7200},
	{32000, 11600, 12000},
	{40000, 12000, 16000},
	{48000, 16000, 16000},
	{64000, 16000, 16000},
	{307201, 16000, 16000},
}

var bandWidthTableLD32000 = [...]bandwidthTab{
	{8000, 2000, 2000},
	{12000, 2000, 2000},
	{24000, 4250, 7200},
	{32000, 8400, 9000},
	{40000, 9400, 11300},
	{48000, 11900, 14700},
	{64000, 14800, 16000},
	{76000, 16000, 16000},
	{409601, 16000, 16000},
}

var bandWidthTableLD44100 = [...]bandwidthTab{
	{8000, 2000, 2000},
	{24000, 2000, 2000},
	{32000, 4400, 5700},
	{40000, 7400, 8800},
	{48000, 9000, 10700},
	{56000, 11000, 12900},
	{64000, 14400, 15500},
	{80000, 16000, 16200},
	{96000, 16500, 16000},
	{128000, 16000, 16000},
	{564481, 16000, 16000},
}

var bandWidthTableLD48000 = [...]bandwidthTab{
	{8000, 2000, 2000},
	{24000, 2000, 2000},
	{32000, 4400, 5700},
	{40000, 7400, 8800},
	{48000, 9000, 10700},
	{56000, 11000, 12800},
	{64000, 14300, 15400},
	{80000, 16000, 16200},
	{96000, 16500, 16000},
	{128000, 16000, 16000},
	{614401, 16000, 16000},
}

var bandWidthTableVBR = [...]bandwidthTabVBR{
	{QCBitrateModeCBR, 0, 0},
	{QCBitrateModeVBR1, 13000, 13000},
	{QCBitrateModeVBR2, 13000, 13000},
	{QCBitrateModeVBR3, 15750, 15750},
	{QCBitrateModeVBR4, 16500, 16500},
	{QCBitrateModeVBR5, 19293, 19293},
	{QCBitrateModeSFR, 0, 0},
	{QCBitrateModeFF, 0, 0},
}

var configTabVBR = [...]configTabEntryVBR{
	{QCBitrateModeCBR, [2]int{0, 0}},
	{QCBitrateModeVBR1, [2]int{32000, 20000}},
	{QCBitrateModeVBR2, [2]int{40000, 32000}},
	{QCBitrateModeVBR3, [2]int{56000, 48000}},
	{QCBitrateModeVBR4, [2]int{72000, 64000}},
	{QCBitrateModeVBR5, [2]int{112000, 96000}},
}

func FDKaacEncDetermineBandWidth(
	proposedBandWidth int,
	bitrate int,
	bitrateMode QCBitrateMode,
	sampleRate int,
	frameLength int,
	cm *ChannelMapping,
	encoderMode ChannelMode,
) (int, int) {
	if cm == nil {
		panic("fdkaac: nil bandwidth channel mapping")
	}
	nChannelsEff := fdkaacEncEffectiveChannels(cm, encoderMode)
	if nChannelsEff <= 0 {
		panic("fdkaac: invalid bandwidth channel count")
	}
	chanBitRate := bitrate / nChannelsEff
	errorStatus := AACEncOK
	bandWidth := 0

	switch bitrateMode {
	case QCBitrateModeVBR1, QCBitrateModeVBR2, QCBitrateModeVBR3, QCBitrateModeVBR4, QCBitrateModeVBR5:
		if proposedBandWidth != 0 {
			bandWidth = proposedBandWidth
		} else {
			switch FDKaacEncGetMonoStereoMode(encoderMode) {
			case ElementModeMono:
				bandWidth = bandWidthTableVBR[bitrateMode].BandWidthMono
			case ElementModeStereo:
				bandWidth = bandWidthTableVBR[bitrateMode].BandWidth2AndMoreCha
			default:
				return 0, AACEncUnsupportedChannelConf
			}
		}
	case QCBitrateModeCBR, QCBitrateModeSFR, QCBitrateModeFF:
		if proposedBandWidth != 0 {
			bandWidth = minInt(proposedBandWidth, minInt(20000, sampleRate>>1))
		} else {
			entryNo := 0
			switch FDKaacEncGetMonoStereoMode(encoderMode) {
			case ElementModeMono:
				entryNo = 0
			case ElementModeStereo:
				entryNo = 1
			default:
				return 0, AACEncUnsupportedChannelConf
			}

			bandWidth = fdkaacEncGetBandwidthEntry(frameLength, sampleRate, chanBitRate, entryNo)
			if bandWidth == -1 {
				switch frameLength {
				case 120, 128, 240, 256:
					bandWidth = 16000
				default:
					errorStatus = AACEncInvalidChannelBitrate
				}
			}
		}
	default:
		return 0, AACEncUnsupportedBitrateMode
	}

	bandWidth = minInt(bandWidth, sampleRate/2)
	return bandWidth, errorStatus
}

func FDKaacEncGetVBRBitrate(bitrateMode QCBitrateMode, channelMode ChannelMode) int {
	monoStereoMode := 0
	if FDKaacEncGetMonoStereoMode(channelMode) == ElementModeStereo {
		monoStereoMode = 1
	}
	bitrate := 0
	switch bitrateMode {
	case QCBitrateModeVBR1, QCBitrateModeVBR2, QCBitrateModeVBR3, QCBitrateModeVBR4, QCBitrateModeVBR5:
		bitrate = configTabVBR[bitrateMode].ChanBitrate[monoStereoMode]
	case QCBitrateModeInvalid, QCBitrateModeCBR, QCBitrateModeSFR, QCBitrateModeFF:
		bitrate = 0
	default:
		bitrate = 0
	}

	cfg, ok := FDKaacEncGetChannelModeConfiguration(channelMode)
	if !ok {
		panic("fdkaac: invalid VBR channel mode")
	}
	return bitrate * cfg.NChannelsEff
}

func FDKaacEncAdjustVBRBitrateMode(bitrateMode QCBitrateMode, bitrate int, channelMode ChannelMode) QCBitrateMode {
	newBitrateMode := bitrateMode
	if bitrate != -1 {
		monoStereoMode := 0
		if FDKaacEncGetMonoStereoMode(channelMode) == ElementModeStereo {
			monoStereoMode = 1
		}
		cfg, ok := FDKaacEncGetChannelModeConfiguration(channelMode)
		if !ok {
			panic("fdkaac: invalid VBR channel mode")
		}
		nChannelsEff := cfg.NChannelsEff
		newBitrateMode = QCBitrateModeInvalid

		for idx := len(configTabVBR) - 1; idx >= 0; idx-- {
			threshold := configTabVBR[idx].ChanBitrate[monoStereoMode] * nChannelsEff
			if bitrate >= threshold {
				if threshold < FDKaacEncGetVBRBitrate(bitrateMode, channelMode) {
					newBitrateMode = configTabVBR[idx].BitrateMode
				} else {
					newBitrateMode = bitrateMode
				}
				break
			}
		}
	}

	if fdkaacEncBitrateModeIsVBR(newBitrateMode) {
		return newBitrateMode
	}
	return QCBitrateModeInvalid
}

func FDKaacEncGetMonoStereoMode(mode ChannelMode) ElementMode {
	switch mode {
	case Mode1:
		return ElementModeMono
	case Mode2, Mode1_2, Mode1_2_1, Mode1_2_2, Mode1_2_2_1, Mode6_1, Mode1_2_2_2_1, Mode7_1RearSurround, Mode7_1FrontCenter, Mode7_1Back, Mode7_1TopFront:
		return ElementModeStereo
	default:
		return ElementModeInvalid
	}
}

func FDKaacEncGetChannelModeConfiguration(mode ChannelMode) (ChannelModeConfig, bool) {
	for _, cfg := range channelModeConfig {
		if cfg.EncMode == mode {
			return cfg, true
		}
	}
	return ChannelModeConfig{}, false
}

func fdkaacEncGetBandwidthEntry(frameLength int, sampleRate int, chanBitRate int, entryNo int) int {
	table := fdkaacEncBandwidthTable(frameLength, sampleRate)
	if len(table) == 0 {
		return -1
	}

	for i := 0; i < len(table)-1; i++ {
		if chanBitRate < table[i].ChanBitRate || chanBitRate >= table[i+1].ChanBitRate {
			continue
		}

		startBw, endBw := table[i].BandWidthMono, table[i+1].BandWidthMono
		if entryNo != 0 {
			startBw, endBw = table[i].BandWidth2AndMoreCha, table[i+1].BandWidth2AndMoreCha
		}

		switch frameLength {
		case 960, 1024:
			return startBw
		case 120, 128, 240, 256, 480, 512:
			startBr := table[i].ChanBitRate
			endBr := table[i+1].ChanBitRate
			bwFac, qRes := fDivNormExp(FixpDBL(chanBitRate-startBr), FixpDBL(endBr-startBr))
			return int(ScaleValueDBL(FMultDD(bwFac, FixpDBL(endBw-startBw)), qRes)) + startBw
		default:
			return -1
		}
	}
	return -1
}

func fdkaacEncBandwidthTable(frameLength int, sampleRate int) []bandwidthTab {
	switch frameLength {
	case 960, 1024:
		return bandWidthTable[:]
	case 120, 128, 240, 256, 480, 512:
		switch sampleRate {
		case 8000, 11025, 12000, 16000, 22050:
			return bandWidthTableLD22050[:]
		case 24000:
			return bandWidthTableLD24000[:]
		case 32000:
			return bandWidthTableLD32000[:]
		case 44100:
			return bandWidthTableLD44100[:]
		case 48000, 64000, 88200, 96000:
			return bandWidthTableLD48000[:]
		}
	}
	return nil
}

func fdkaacEncEffectiveChannels(cm *ChannelMapping, encoderMode ChannelMode) int {
	nChannelsEff := 0
	for i := 0; i < cm.NElements; i++ {
		elInfo := cm.ElInfo[i]
		switch elInfo.ElType {
		case idSCE, idCPE:
			nChannelsEff += elInfo.NChannelsInEl
		case idLFE, idDSE:
		default:
			nChannelsEff += elInfo.NChannelsInEl
		}
	}
	if nChannelsEff != 0 {
		return nChannelsEff
	}
	cfg, ok := FDKaacEncGetChannelModeConfiguration(encoderMode)
	if !ok {
		return 0
	}
	return cfg.NChannelsEff
}

func fdkaacEncBitrateModeIsVBR(mode QCBitrateMode) bool {
	return mode >= QCBitrateModeVBR1 && mode <= QCBitrateModeVBR5
}
