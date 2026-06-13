package fdkaac

const (
	bitresMin     = 300
	bitresMaxLD   = 4000
	bitresMinLD   = 500
	bitrateMaxLD  = 70000
	bitrateMinLD  = 12000
	maxAncRate    = 19200
	tnsEnableMask = 0xf
)

type FDKaacEncStaticBitsFunc func(bits int) int

type AACEncConfig struct {
	SampleRate     int
	BitRate        int
	AncDataBitRate int
	NSubFrames     int
	AOT            AudioObjectType

	AverageBits  int
	BitrateMode  QCBitrateMode
	NChannels    int
	ChannelOrder ChannelOrder
	BandWidth    int
	ChannelMode  ChannelMode
	FrameLength  int

	SyntaxFlags uint32
	EpConfig    int8

	AncRate          int
	MaxAncBytesPerAU uint
	MinBitsPerFrame  int
	MaxBitsPerFrame  int

	AudioMuxVersion int

	UseTns     int
	UsePns     int
	UseIS      int
	UseMS      int
	UseRequant int

	DownscaleFactor uint
}

type AACEncInitState struct {
	ChannelMapping        ChannelMapping
	QCInit                QCInit
	AverageBitsPerFrame   int
	AncillaryBitsPerFrame int
	PsyBitrate            int
	Bandwidth90dB         int
	TNSMask               int
	InternalAOT           AudioObjectType
}

func FDKaacEncAacInitDefaultConfig(config *AACEncConfig) {
	if config == nil {
		panic("fdkaac: nil encoder config")
	}
	*config = AACEncConfig{
		BitRate:         -1,
		AverageBits:     -1,
		BitrateMode:     QCBitrateModeCBR,
		UseTns:          tnsEnableMask,
		UsePns:          1,
		UseIS:           1,
		UseMS:           1,
		FrameLength:     -1,
		EpConfig:        -1,
		NSubFrames:      1,
		ChannelOrder:    ChannelOrderMPEG,
		ChannelMode:     ModeUnknown,
		MinBitsPerFrame: -1,
		MaxBitsPerFrame: -1,
		AudioMuxVersion: -1,
		DownscaleFactor: 1,
	}
}

func FDKaacEncCalcBitsPerFrame(bitRate int, frameLength int, samplingRate int) int {
	shift := 0
	for (frameLength&^((1<<(shift+1))-1)) == frameLength &&
		(samplingRate&^((1<<(shift+1))-1)) == samplingRate {
		shift++
	}
	return (bitRate * (frameLength >> shift)) / (samplingRate >> shift)
}

func FDKaacEncCalcBitrate(bitsPerFrame int, frameLength int, samplingRate int) int {
	shift := 0
	for (frameLength&^((1<<(shift+1))-1)) == frameLength &&
		(samplingRate&^((1<<(shift+1))-1)) == samplingRate {
		shift++
	}
	return (bitsPerFrame * (samplingRate >> shift)) / (frameLength >> shift)
}

func FDKaacEncLimitBitrate(
	staticBits FDKaacEncStaticBitsFunc,
	aot AudioObjectType,
	coreSamplingRate int,
	frameLength int,
	nChannels int,
	nChannelsEff int,
	bitRate int,
	averageBits int,
	bitrateMode QCBitrateMode,
	nSubFrames int,
) (int, int) {
	_ = averageBits
	_ = bitrateMode
	transportBits := 0
	prevBitrate := 0
	averageBitsPerFrame := 0
	minBitrate := 0
	iter := 0
	minBitsPerFrame := 40 * nChannels
	if isLowDelayAOT(aot) {
		minBitrate = 8000 * nChannelsEff
	}

	for {
		prevBitrate = bitRate
		averageBitsPerFrame = FDKaacEncCalcBitsPerFrame(bitRate, frameLength, coreSamplingRate) / nSubFrames
		transportBits = fdkaacEncStaticBits(staticBits, averageBitsPerFrame)
		bitRate = maxInt(bitRate, maxInt(minBitrate, FDKaacEncCalcBitrate(minBitsPerFrame+transportBits, frameLength, coreSamplingRate)))
		bitRate = minInt(bitRate, FDKaacEncCalcBitrate(nChannelsEff*minBufSizePerEffChan, frameLength, coreSamplingRate))
		if prevBitrate == bitRate || iter >= 3 {
			break
		}
		iter++
	}

	return bitRate, averageBitsPerFrame
}

func FDKaacEncInitCheckAncillary(bitRate int, frameLength int, ancillaryRate int, ancillaryBitsPerFrame *int, sampleRate int) int {
	if ancillaryBitsPerFrame == nil {
		panic("fdkaac: nil ancillary bits output")
	}
	if ancillaryRate < -1 {
		return AACEncUnsupportedAncBitrate
	}
	if ancillaryRate != -1 && ancillaryRate != 0 {
		if ancillaryRate >= maxAncRate || ancillaryRate*20 > bitRate*3 {
			return AACEncUnsupportedAncBitrate
		}
	} else if ancillaryRate == -1 {
		if bitRate >= maxAncRate*10 {
			ancillaryRate = maxAncRate - 1
		} else {
			ancillaryRate = bitRate / 10
		}
	}
	*ancillaryBitsPerFrame = FDKaacEncCalcBitsPerFrame(ancillaryRate, frameLength, sampleRate) &^ 0x7
	return AACEncOK
}

func FDKaacEncPrepareQCInitFromConfig(state *AACEncInitState, config *AACEncConfig, staticBits FDKaacEncStaticBitsFunc) int {
	if state == nil || config == nil {
		return AACEncInvalidHandle
	}
	if config.NChannels < 1 || config.NChannels > 8 {
		return AACEncUnsupportedChannelConf
	}
	if !fdkaacEncValidSamplingRate(config.SampleRate) {
		return AACEncUnsupportedSamplingRate
	}
	if config.BitRate == -1 {
		return AACEncUnsupportedBitrate
	}
	modeConfig, ok := FDKaacEncGetChannelModeConfiguration(config.ChannelMode)
	if !ok || modeConfig.NChannels != config.NChannels {
		return AACEncUnsupportedChannelConf
	}

	limitedBitrate, averageBitsPerFrame := FDKaacEncLimitBitrate(
		staticBits,
		config.AOT,
		config.SampleRate,
		config.FrameLength,
		config.NChannels,
		modeConfig.NChannelsEff,
		config.BitRate,
		config.AverageBits,
		config.BitrateMode,
		config.NSubFrames,
	)
	if limitedBitrate != config.BitRate && !fdkaacEncBitrateModeIsVBR(config.BitrateMode) {
		return AACEncUnsupportedBitrate
	}
	if config.SyntaxFlags&acERVCB11 != 0 || config.SyntaxFlags&acERHCR != 0 {
		return AACEncUnsupportedERFormat
	}
	if !fdkaacEncValidFrameLength(config.FrameLength, config.AOT) {
		return AACEncInvalidFrameLength
	}

	*state = AACEncInitState{}
	if config.AncRate != 0 {
		errCode := FDKaacEncInitCheckAncillary(
			config.BitRate,
			config.FrameLength,
			config.AncRate,
			&state.AncillaryBitsPerFrame,
			config.SampleRate,
		)
		if errCode != AACEncOK {
			return errCode
		}
		config.AncDataBitRate += FDKaacEncCalcBitrate(state.AncillaryBitsPerFrame, config.FrameLength, config.SampleRate)
	}
	config.MaxAncBytesPerAU = uint(minInt(256, maxInt(0, FDKaacEncCalcBitsPerFrame(config.BitRate-config.NChannels*8000, config.FrameLength, config.SampleRate)>>3)))

	state.PsyBitrate = config.BitRate - config.AncDataBitRate
	if config.UseTns != 0 {
		state.TNSMask = tnsEnableMask
	}
	if errCode := FDKaacEncInitChannelMapping(config.ChannelMode, config.ChannelOrder, &state.ChannelMapping); errCode != AACEncOK {
		return errCode
	}
	bandwidth, errCode := FDKaacEncDetermineBandWidth(
		config.BandWidth,
		state.PsyBitrate,
		config.BitrateMode,
		config.SampleRate,
		config.FrameLength,
		&state.ChannelMapping,
		config.ChannelMode,
	)
	if errCode != AACEncOK {
		return errCode
	}
	config.BandWidth = bandwidth
	state.Bandwidth90dB = bandwidth
	state.AverageBitsPerFrame = averageBitsPerFrame
	state.InternalAOT = fdkaacEncMapInternalAOT(config.AOT)

	qcInit := &state.QCInit
	qcInit.ChannelMapping = &state.ChannelMapping
	qcInit.SceCpe = 0
	if fdkaacEncBitrateModeIsVBR(config.BitrateMode) {
		qcInit.AverageBits = (averageBitsPerFrame + 7) &^ 7
		qcInit.BitRes = minBufSizePerEffChan * state.ChannelMapping.NChannelsEff
		qcInit.MaxBits = minBufSizePerEffChan * state.ChannelMapping.NChannelsEff
		if config.MaxBitsPerFrame != -1 {
			qcInit.MaxBits = minInt(qcInit.MaxBits, config.MaxBitsPerFrame)
		}
		qcInit.MaxBits = maxInt(qcInit.MaxBits, (averageBitsPerFrame+7)&^7)
		if config.MinBitsPerFrame != -1 {
			qcInit.MinBits = config.MinBitsPerFrame
		}
		qcInit.MinBits = minInt(qcInit.MinBits, averageBitsPerFrame&^7)
	} else {
		bitreservoir := -1
		if isLowDelayAOT(config.AOT) {
			brPerChannel := config.BitRate / config.NChannels
			brPerChannel = minInt(bitrateMaxLD, maxInt(bitrateMinLD, brPerChannel))
			slope := fDivNorm(FixpDBL(brPerChannel-bitrateMinLD), FixpDBL(bitrateMaxLD-bitrateMinLD))
			bitreservoir = (FMultI(slope, bitresMaxLD-bitresMinLD) + bitresMinLD) &^ 7
		}
		qcInit.AverageBits = (averageBitsPerFrame + 7) &^ 7
		maxBitres := minBufSizePerEffChan*state.ChannelMapping.NChannelsEff - qcInit.AverageBits
		if bitreservoir != -1 {
			qcInit.BitRes = minInt(bitreservoir, maxBitres)
		} else {
			qcInit.BitRes = maxBitres
		}
		qcInit.MaxBits = minInt(minBufSizePerEffChan*state.ChannelMapping.NChannelsEff, qcInit.AverageBits+qcInit.BitRes)
		if config.MaxBitsPerFrame != -1 {
			qcInit.MaxBits = minInt(qcInit.MaxBits, config.MaxBitsPerFrame)
		}
		qcInit.MaxBits = minInt(
			minBufSizePerEffChan*state.ChannelMapping.NChannelsEff,
			maxInt(qcInit.MaxBits, (averageBitsPerFrame+7+8)&^7),
		)
		qcInit.MinBits = maxInt(0, ((averageBitsPerFrame-1)&^7)-qcInit.BitRes-fdkaacEncStaticBits(staticBits, ((averageBitsPerFrame+7)&^7)+qcInit.BitRes))
		if config.MinBitsPerFrame != -1 {
			qcInit.MinBits = maxInt(qcInit.MinBits, config.MinBitsPerFrame)
		}
		qcInit.MinBits = minInt(qcInit.MinBits, (averageBitsPerFrame-fdkaacEncStaticBits(staticBits, qcInit.MaxBits))&^7)
	}

	qcInit.SampleRate = config.SampleRate
	if isLowDelayAOT(config.AOT) {
		qcInit.IsLowDelay = 1
	}
	qcInit.NSubFrames = config.NSubFrames
	qcInit.PaddingRest = config.SampleRate
	if qcInit.MaxBits-qcInit.AverageBits >= fdkaacEncBitresMin(qcInit.IsLowDelay)*config.NChannels {
		qcInit.BitResMode = BitresModeFull
	} else if qcInit.MaxBits > qcInit.AverageBits {
		qcInit.BitResMode = BitresModeReduced
	} else {
		qcInit.BitResMode = BitresModeDisabled
	}
	qcInit.BitDistributionMode = fdkaacEncBitDistributionMode(config.ChannelMode)
	qcInit.MeanPe = fdkaacEncCalcMeanPe(config.FrameLength, state.Bandwidth90dB, config.SampleRate)
	qcInit.MaxBitFac = fdkaacEncCalcMaxBitFac(qcInit.MaxBits, qcInit.AverageBits/qcInit.NSubFrames)
	if !fdkaacEncMapQCBitrateMode(config.BitrateMode, &qcInit.BitrateMode) {
		return AACEncUnsupportedBitrateMode
	}
	if config.UseRequant != 0 {
		qcInit.InvQuant = 2
	}
	if isLowDelayAOT(config.AOT) {
		qcInit.MaxIterations = 2
	} else {
		qcInit.MaxIterations = 5
	}
	qcInit.Bitrate = state.PsyBitrate
	qcInit.StaticBits = fdkaacEncStaticBits(staticBits, qcInit.AverageBits/qcInit.NSubFrames)
	return AACEncOK
}

func FDKaacEncEncBitresToTpBitres(qcKernel *QCKernel, bitrateMode QCBitrateMode, audioMuxVersion int, nChannelsEff int) int {
	if qcKernel == nil {
		panic("fdkaac: nil transport bitreservoir kernel")
	}
	transportBitreservoir := 0
	switch bitrateMode {
	case QCBitrateModeCBR:
		transportBitreservoir = qcKernel.BitResTot
	case QCBitrateModeVBR1, QCBitrateModeVBR2, QCBitrateModeVBR3, QCBitrateModeVBR4, QCBitrateModeVBR5:
		transportBitreservoir = fdkIntMax
	case QCBitrateModeSFR:
		transportBitreservoir = 0
	case QCBitrateModeInvalid:
		transportBitreservoir = 0
	default:
		transportBitreservoir = 0
	}
	if audioMuxVersion == 2 {
		transportBitreservoir = minBufSizePerEffChan * nChannelsEff
	}
	return transportBitreservoir
}

func fdkaacEncStaticBits(staticBits FDKaacEncStaticBitsFunc, bits int) int {
	if staticBits == nil {
		return 208
	}
	return staticBits(bits)
}

func fdkaacEncValidSamplingRate(sampleRate int) bool {
	switch sampleRate {
	case 8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000, 64000, 88200, 96000:
		return true
	default:
		return false
	}
}

func fdkaacEncValidFrameLength(frameLength int, aot AudioObjectType) bool {
	switch frameLength {
	case 1024:
		return !isLowDelayAOT(aot)
	case 128, 256, 512, 120, 240, 480:
		return isLowDelayAOT(aot)
	default:
		return false
	}
}

func isLowDelayAOT(aot AudioObjectType) bool {
	return aot == AOTERAACLD || aot == AOTERAACELD
}

func fdkaacEncBitresMin(isLowDelay int) int {
	if isLowDelay != 0 {
		return bitresMinLD
	}
	return bitresMin
}

func fdkaacEncBitDistributionMode(mode ChannelMode) int {
	switch mode {
	case Mode1_2, Mode1_2_1, Mode1_2_2, Mode1_2_2_1, Mode6_1, Mode1_2_2_2_1,
		Mode7_1Back, Mode7_1TopFront, Mode7_1RearSurround, Mode7_1FrontCenter:
		return BitDistributionModeInterElement
	default:
		return BitDistributionModeIntraElement
	}
}

func fdkaacEncCalcMeanPe(frameLength int, bandwidth int, sampleRate int) int {
	bwRatio, qbw := fDivNormExp(FixpDBL(10*frameLength*bandwidth), FixpDBL(sampleRate))
	return maxInt(int(ScaleValueDBL(bwRatio, qbw+1-(DfractBits-1))), 1)
}

func fdkaacEncCalcMaxBitFac(maxBits int, averageBitsPerSubFrame int) FixpDBL {
	mbfac, mbfacE := fDivNormExp(FixpDBL(maxBits), FixpDBL(averageBitsPerSubFrame))
	return ScaleValueDBL(mbfac, mbfacE-(DfractBits-1-qBitFac))
}

func fdkaacEncMapQCBitrateMode(mode QCBitrateMode, out *QCBitrateMode) bool {
	switch mode {
	case QCBitrateModeCBR, QCBitrateModeVBR1, QCBitrateModeVBR2, QCBitrateModeVBR3,
		QCBitrateModeVBR4, QCBitrateModeVBR5, QCBitrateModeSFR, QCBitrateModeFF:
		*out = mode
		return true
	default:
		return false
	}
}

func fdkaacEncMapInternalAOT(aot AudioObjectType) AudioObjectType {
	switch aot {
	case AOTMP2AACLC:
		return AOTAACLC
	case AOTMP2SBR:
		return AudioObjectType(aotSBR)
	default:
		return aot
	}
}
