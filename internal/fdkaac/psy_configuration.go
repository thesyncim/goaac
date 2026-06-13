package fdkaac

const (
	maxSFBLong     = 51
	maxSFB         = maxSFBLong
	lfeLowpassLine = 12

	bitsPerLineShift  = 3
	cRatio            = FixpDBL(0x02940a10)
	pcmQuantNoise     = FixpDBL(0x00547062)
	psyClipEnergyLong = FixpDBL(0x773593ff)
)

const (
	FilterbankLC = iota
	FilterbankLD
	FilterbankELD
)

type PsyConfiguration struct {
	SfbCnt                      int
	SfbActive                   int
	SfbActiveLFE                int
	SfbOffset                   [maxSFB + 1]int
	Filterbank                  int
	SfbPcmQuantThreshold        [maxSFB]FixpDBL
	MaxAllowedIncreaseFactor    int
	MinRemainingThresholdFactor FixpSGL
	LowpassLine                 int
	LowpassLineLFE              int
	ClipEnergy                  FixpDBL
	SfbMaskLowFactor            [maxSFB]FixpDBL
	SfbMaskHighFactor           [maxSFB]FixpDBL
	SfbMaskLowFactorSprEn       [maxSFB]FixpDBL
	SfbMaskHighFactorSprEn      [maxSFB]FixpDBL
	SfbMinSnrLdData             [maxSFB]FixpDBL
	GranuleLength               int
	AllowIS                     bool
	AllowMS                     bool
}

type sfbParamLong struct {
	sfbCnt   int
	sfbWidth [maxSFBLong]uint8
}

type sfbParamShort struct {
	sfbCnt   int
	sfbWidth [maxSFBShort]uint8
}

type sfbInfoTabEntry struct {
	sampleRate int
	paramLong  *sfbParamLong
	paramShort *sfbParamShort
}

var (
	sfb8000Long1024 = sfbParamLong{40, [maxSFBLong]uint8{
		12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 16, 16, 16, 16,
		16, 16, 16, 20, 20, 20, 20, 24, 24, 24, 28, 28, 32, 36, 36, 40, 44,
		48, 52, 56, 60, 64, 80,
	}}
	sfb8000Short128 = sfbParamShort{15, [maxSFBShort]uint8{
		4, 4, 4, 4, 4, 4, 4, 8, 8, 8, 8, 12, 16, 20, 20,
	}}
	sfb11025Long1024 = sfbParamLong{43, [maxSFBLong]uint8{
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 12, 12, 12, 12, 12, 12, 12, 12,
		12, 16, 16, 16, 16, 20, 20, 20, 24, 24, 28, 28, 32, 36, 40, 40, 44,
		48, 52, 56, 60, 64, 64, 64,
	}}
	sfb11025Short128 = sfbParamShort{15, [maxSFBShort]uint8{
		4, 4, 4, 4, 4, 4, 4, 4, 8, 8, 12, 12, 16, 20, 20,
	}}
	sfb12000Long1024 = sfb11025Long1024
	sfb12000Short128 = sfb11025Short128
	sfb16000Long1024 = sfb11025Long1024
	sfb16000Short128 = sfb11025Short128
	sfb22050Long1024 = sfbParamLong{47, [maxSFBLong]uint8{
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 12,
		12, 12, 12, 16, 16, 16, 20, 20, 24, 24, 28, 28, 32, 36, 36, 40, 44,
		48, 52, 52, 64, 64, 64, 64, 64,
	}}
	sfb22050Short128 = sfbParamShort{15, [maxSFBShort]uint8{
		4, 4, 4, 4, 4, 4, 4, 8, 8, 8, 12, 12, 16, 16, 20,
	}}
	sfb24000Long1024 = sfb22050Long1024
	sfb24000Short128 = sfb22050Short128
	sfb32000Long1024 = sfbParamLong{51, [maxSFBLong]uint8{
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 8, 8, 8, 8, 8, 8, 8, 12, 12, 12,
		12, 16, 16, 20, 20, 24, 24, 28, 28, 32, 32, 32, 32, 32, 32, 32,
		32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
	}}
	sfb32000Short128 = sfbParamShort{14, [maxSFBShort]uint8{
		4, 4, 4, 4, 4, 8, 8, 8, 12, 12, 12, 16, 16, 16,
	}}
	sfb44100Long1024 = sfbParamLong{49, [maxSFBLong]uint8{
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 8, 8, 8, 8, 8, 8, 8, 12, 12, 12,
		12, 16, 16, 20, 20, 24, 24, 28, 28, 32, 32, 32, 32, 32, 32, 32,
		32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 96,
	}}
	sfb44100Short128 = sfb32000Short128
	sfb48000Long1024 = sfb44100Long1024
	sfb48000Short128 = sfb44100Short128
	sfb64000Long1024 = sfbParamLong{47, [maxSFBLong]uint8{
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 8, 8, 8, 8, 12, 12, 12,
		16, 16, 16, 20, 24, 24, 28, 36, 40, 40, 40, 40, 40, 40, 40, 40, 40,
		40, 40, 40, 40, 40, 40, 40, 40, 40,
	}}
	sfb64000Short128 = sfbParamShort{12, [maxSFBShort]uint8{
		4, 4, 4, 4, 4, 4, 8, 8, 8, 16, 28, 36,
	}}
	sfb88200Long1024 = sfbParamLong{41, [maxSFBLong]uint8{
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 8, 8, 8, 8, 8, 12, 12,
		12, 12, 12, 16, 16, 24, 28, 36, 44, 64, 64, 64, 64, 64, 64, 64,
		64, 64, 64, 64,
	}}
	sfb88200Short128 = sfb64000Short128
	sfb96000Long1024 = sfb88200Long1024
	sfb96000Short128 = sfb88200Short128

	sfbInfoTab = [...]sfbInfoTabEntry{
		{8000, &sfb8000Long1024, &sfb8000Short128},
		{11025, &sfb11025Long1024, &sfb11025Short128},
		{12000, &sfb12000Long1024, &sfb12000Short128},
		{16000, &sfb16000Long1024, &sfb16000Short128},
		{22050, &sfb22050Long1024, &sfb22050Short128},
		{24000, &sfb24000Long1024, &sfb24000Short128},
		{32000, &sfb32000Long1024, &sfb32000Short128},
		{44100, &sfb44100Long1024, &sfb44100Short128},
		{48000, &sfb48000Long1024, &sfb48000Short128},
		{64000, &sfb64000Long1024, &sfb64000Short128},
		{88200, &sfb88200Long1024, &sfb88200Short128},
		{96000, &sfb96000Long1024, &sfb96000Short128},
	}
)

func FDKaacEncInitSfbTable(sampleRate int, blockType int, granuleLength int, sfbOffset []int, sfbCnt *int) int {
	checkInitSfbTableInputs(blockType, sfbOffset, sfbCnt)

	var sfbInfo []sfbInfoTabEntry
	granuleLengthWindow := granuleLength
	switch granuleLength {
	case 1024, 960:
		sfbInfo = sfbInfoTab[:]
	default:
		return AACEncInvalidFrameLength
	}

	var width []uint8
	found := false
	for i := range sfbInfo {
		if sfbInfo[i].sampleRate != sampleRate {
			continue
		}
		found = true
		switch blockType {
		case LongWindow, StartWindow, StopWindow:
			width = sfbInfo[i].paramLong.sfbWidth[:]
			*sfbCnt = sfbInfo[i].paramLong.sfbCnt
		case ShortWindow:
			width = sfbInfo[i].paramShort.sfbWidth[:]
			*sfbCnt = sfbInfo[i].paramShort.sfbCnt
			granuleLengthWindow /= transFac
		}
		break
	}
	if !found {
		return AACEncUnsupportedSamplingRate
	}

	specStartOffset := 0
	i := 0
	for ; i < *sfbCnt; i++ {
		sfbOffset[i] = specStartOffset
		specStartOffset += int(width[i])
		if specStartOffset >= granuleLengthWindow {
			i++
			break
		}
	}
	*sfbCnt = minInt(i, *sfbCnt)
	sfbOffset[*sfbCnt] = minInt(specStartOffset, granuleLengthWindow)
	return AACEncOK
}

func FDKaacEncInitPsyConfiguration(bitrate int, samplerate int, bandwidth int, blocktype int, granuleLength int, useIS int, useMS int, psyConf *PsyConfiguration, filterbank int) int {
	checkInitPsyConfigurationInputs(bitrate, samplerate, bandwidth, blocktype, granuleLength, psyConf, filterbank)

	*psyConf = PsyConfiguration{}
	psyConf.GranuleLength = granuleLength
	psyConf.Filterbank = filterbank
	psyConf.AllowIS = useIS != 0 && bitrate/bandwidth < 5
	psyConf.AllowMS = useMS != 0

	err := FDKaacEncInitSfbTable(samplerate, blocktype, granuleLength, psyConf.SfbOffset[:], &psyConf.SfbCnt)
	if err != AACEncOK {
		return err
	}

	var sfbBarcVal [maxSFB]FixpDBL
	initBarcValues(psyConf.SfbCnt, psyConf.SfbOffset[:], psyConf.SfbOffset[psyConf.SfbCnt], samplerate, sfbBarcVal[:])
	initMinPCMResolution(psyConf.SfbCnt, psyConf.SfbOffset[:], psyConf.SfbPcmQuantThreshold[:])
	initSpreading(psyConf.SfbCnt, sfbBarcVal[:], psyConf.SfbMaskLowFactor[:], psyConf.SfbMaskHighFactor[:], psyConf.SfbMaskLowFactorSprEn[:], psyConf.SfbMaskHighFactorSprEn[:], bitrate, blocktype)

	psyConf.MaxAllowedIncreaseFactor = 2
	psyConf.MinRemainingThresholdFactor = FixpSGL(0x0148)
	psyConf.ClipEnergy = psyClipEnergyLong

	if blocktype != ShortWindow {
		psyConf.LowpassLine = (2 * bandwidth * granuleLength) / samplerate
		psyConf.LowpassLineLFE = lfeLowpassLine
	} else {
		frameLengthShort := granuleLength / transFac
		psyConf.LowpassLine = (2 * bandwidth * frameLengthShort) / samplerate
		psyConf.LowpassLineLFE = 0
		psyConf.ClipEnergy >>= 6
	}

	sfb := 0
	for ; sfb < psyConf.SfbCnt; sfb++ {
		if psyConf.SfbOffset[sfb] >= psyConf.LowpassLine {
			break
		}
	}
	psyConf.SfbActive = maxInt(sfb, 1)

	for sfb = 0; sfb < psyConf.SfbCnt; sfb++ {
		if psyConf.SfbOffset[sfb] >= psyConf.LowpassLineLFE {
			break
		}
	}
	psyConf.SfbActiveLFE = sfb
	psyConf.SfbActive = maxInt(psyConf.SfbActive, psyConf.SfbActiveLFE)

	initMinSnr(bitrate, samplerate, psyConf.SfbOffset[psyConf.SfbCnt], psyConf.SfbOffset[:], psyConf.SfbActive, blocktype, psyConf.SfbMinSnrLdData[:])
	return AACEncOK
}

func FDKaacEncBarcLineValue(noOfLines int, fftLine int, samplingFreq int) FixpDBL {
	const (
		fourBy3Em4 FixpDBL = 0x45e7b273
		pzzz76     FixpDBL = 0x639d5e4a
		one3P3     FixpDBL = 0x35333333
		threeP5    FixpDBL = 0x1c000000
		inv480     FixpDBL = 0x44444444
	)

	centerFreq := FixpDBL(fftLine * samplingFreq)
	switch noOfLines {
	case 1024:
		centerFreq <<= 2
	case 128:
		centerFreq <<= 5
	case 512:
		centerFreq <<= 3
	case 480:
		centerFreq = FMultDD(centerFreq, inv480) << 4
	default:
		centerFreq = 0
	}

	x1 := FMultDD(centerFreq, fourBy3Em4)
	x2 := FMultDD(centerFreq, pzzz76) << 2
	atan1 := fixpAtan(x1)
	atan2 := fixpAtan(x2)
	return FMultDD(one3P3, atan2) + FMultDD(threeP5, FMultDD(atan1, atan1))
}

func initMinPCMResolution(numPb int, pbOffset []int, sfbPCMQuantThreshold []FixpDBL) {
	for i := 0; i < numPb; i++ {
		sfbPCMQuantThreshold[i] = FixpDBL(pbOffset[i+1]-pbOffset[i]) * pcmQuantNoise
	}
}

func getMaskFactor(dbValFix FixpDBL, dbValE int, tenFix FixpDBL, tenE int) FixpDBL {
	maskFactor, qMsk := fPow(tenFix, DfractBits-1-tenE, -dbValFix, DfractBits-1-dbValE)
	qMsk = minInt(DfractBits-1, maxInt(-(DfractBits-1), qMsk))
	if qMsk > 0 && maskFactor > MaxValDBL>>uint(qMsk) {
		return MaxValDBL
	}
	return ScaleValueDBL(maskFactor, qMsk)
}

func initSpreading(numPb int, pbBarcValue []FixpDBL, pbMaskLoFactor []FixpDBL, pbMaskHiFactor []FixpDBL, pbMaskLoFactorSprEn []FixpDBL, pbMaskHiFactorSprEn []FixpDBL, bitrate int, blockType int) {
	const (
		maskHigh             FixpDBL = 0x30000000
		maskLow              FixpDBL = 0x60000000
		maskLowSprEnLong     FixpDBL = 0x60000000
		maskHighSprEnLong    FixpDBL = 0x40000000
		maskHighSprEnLongLow FixpDBL = 0x30000000
		maskLowSprEnShort    FixpDBL = 0x40000000
		maskHighSprEnShort   FixpDBL = 0x30000000
		ten                  FixpDBL = 0x50000000
	)

	maskLowSprEn := maskLowSprEnShort
	maskHighSprEn := maskHighSprEnShort
	if blockType != ShortWindow {
		maskLowSprEn = maskLowSprEnLong
		if bitrate > 20000 {
			maskHighSprEn = maskHighSprEnLong
		} else {
			maskHighSprEn = maskHighSprEnLongLow
		}
	}

	for i := 0; i < numPb; i++ {
		if i > 0 {
			diff := pbBarcValue[i] - pbBarcValue[i-1]
			pbMaskHiFactor[i] = getMaskFactor(FMultDD(maskHigh, diff), 23, ten, 27)
			pbMaskLoFactor[i-1] = getMaskFactor(FMultDD(maskLow, diff), 23, ten, 27)
			pbMaskHiFactorSprEn[i] = getMaskFactor(FMultDD(maskHighSprEn, diff), 23, ten, 27)
			pbMaskLoFactorSprEn[i-1] = getMaskFactor(FMultDD(maskLowSprEn, diff), 23, ten, 27)
			continue
		}
		pbMaskHiFactor[i] = 0
		pbMaskLoFactor[numPb-1] = 0
		pbMaskHiFactorSprEn[i] = 0
		pbMaskLoFactorSprEn[numPb-1] = 0
	}
}

func initBarcValues(numPb int, pbOffset []int, numLines int, samplingFrequency int, pbBval []FixpDBL) {
	const maxBarc FixpDBL = 0x30000000
	for i := 0; i < numPb; i++ {
		v1 := FDKaacEncBarcLineValue(numLines, pbOffset[i], samplingFrequency)
		v2 := FDKaacEncBarcLineValue(numLines, pbOffset[i+1], samplingFrequency)
		curBark := (v1 >> 1) + (v2 >> 1)
		pbBval[i] = minFixpDBL(curBark, maxBarc)
	}
}

func initMinSnr(bitrate int, samplerate int, numLines int, sfbOffset []int, sfbActive int, blockType int, sfbMinSnrLdData []FixpDBL) {
	const (
		maxBarc    FixpDBL = 0x30000000
		maxBarcP1  FixpDBL = 0x32000000
		bits2PEFac FixpDBL = 0x4b851eb8
		pers2P4    FixpDBL = 0x624dd2f2
		oneP5      FixpDBL = 0x60000000
		maxSnr     FixpDBL = 0x33333333
		minSnr     FixpDBL = 0x003126e9
	)

	barcFactor, qBFac := fDivNormExp(minFixpDBL(FDKaacEncBarcLineValue(numLines, sfbOffset[sfbActive], samplerate), maxBarc), maxBarcP1)
	qBFac = DfractBits - 1 - qBFac

	pePerWindow, qPerWin := fDivNormExp(FixpDBL(bitrate), FixpDBL(samplerate))
	qPerWin = DfractBits - 1 - qPerWin
	pePerWindow = FMultDD(pePerWindow, bits2PEFac)
	qPerWin = qPerWin + 30 - (DfractBits - 1)
	pePerWindow = FMultDD(pePerWindow, pers2P4)
	qPerWin = qPerWin + 36 - (DfractBits - 1)

	switch numLines {
	case 1024:
		qPerWin -= 10
	case 128:
		qPerWin -= 7
	case 512, 480:
		qPerWin -= 9
		if numLines == 480 {
			pePerWindow = FMultDD(pePerWindow, 0x78000000)
		}
	}

	if blockType == ShortWindow {
		pePerWindow = FMultDD(pePerWindow, oneP5)
		qPerWin = qPerWin + 30 - (DfractBits - 1)
	}

	pePartConst, qDiv := fDivNormExp(pePerWindow, barcFactor)
	qPePartConst := qPerWin - qBFac + DfractBits - 1 - qDiv

	for sfb := 0; sfb < sfbActive; sfb++ {
		barcWidth := FDKaacEncBarcLineValue(numLines, sfbOffset[sfb+1], samplerate) -
			FDKaacEncBarcLineValue(numLines, sfbOffset[sfb], samplerate)
		pePart := FMultDD(pePartConst, barcWidth)
		qPePart := qPePartConst + 25 - (DfractBits - 1)

		sfbWidth := sfbOffset[sfb+1] - sfbOffset[sfb]
		pePart, qDiv = fDivNormExp(pePart, FixpDBL(sfbWidth))
		qPePart += DfractBits - 1 - qDiv

		tmp, qTmp := f2Pow(pePart, DfractBits-1-qPePart)
		qTmp = DfractBits - 1 - qTmp

		qSnr := minInt(qTmp, 30)
		tmp >>= uint(qTmp - qSnr)

		onePoint5 := FixpDBL(0)
		if 30+1-qSnr <= DfractBits-1 {
			onePoint5 = oneP5 >> uint(30+1-qSnr)
		}

		snr := (tmp >> 1) - onePoint5
		qSnr--

		oneQSnr := FixpDBL(0)
		if qSnr > 0 {
			oneQSnr = FixpDBL(1 << uint(qSnr))
		}

		snr = maxFixpDBL(oneQSnr, snr)
		snr, qSnr = fDivNormExp(oneQSnr, snr)
		qSnr = DfractBits - 1 - qSnr
		if qSnr > 30 {
			snr >>= uint(qSnr - 30)
		}

		snr = minFixpDBL(snr, maxSnr)
		snr = maxFixpDBL(snr, minSnr)
		snr <<= 1
		sfbMinSnrLdData[sfb] = CalcLdData(snr)
	}
}

func checkInitSfbTableInputs(blockType int, sfbOffset []int, sfbCnt *int) {
	if !validSFBWindowSequence(blockType) {
		panic("fdkaac: invalid sfb block type")
	}
	if len(sfbOffset) < maxSFB+1 {
		panic("fdkaac: short sfb offset table")
	}
	if sfbCnt == nil {
		panic("fdkaac: nil sfb count")
	}
}

func checkInitPsyConfigurationInputs(bitrate int, samplerate int, bandwidth int, blocktype int, granuleLength int, psyConf *PsyConfiguration, filterbank int) {
	if bitrate <= 0 || samplerate <= 0 || bandwidth <= 0 {
		panic("fdkaac: invalid psy configuration rate")
	}
	if !validSFBWindowSequence(blocktype) {
		panic("fdkaac: invalid psy configuration block type")
	}
	if granuleLength != 1024 {
		panic("fdkaac: unsupported psy configuration frame length")
	}
	if psyConf == nil {
		panic("fdkaac: nil psy configuration")
	}
	if filterbank != FilterbankLC && filterbank != FilterbankLD && filterbank != FilterbankELD {
		panic("fdkaac: invalid psy filterbank")
	}
}

func validSFBWindowSequence(blockType int) bool {
	return blockType == LongWindow || blockType == StartWindow || blockType == ShortWindow || blockType == StopWindow
}
