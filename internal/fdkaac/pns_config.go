package fdkaac

const (
	pnsTableError = -1

	pnsUsePowerDistribution = 1 << 0
	pnsUsePsychTonality     = 1 << 1
	pnsUseTnsGainThreshold  = 1 << 2
	pnsUseTnsPNS            = 1 << 3
	pnsJustLongWindow       = 1 << 4
	pnsIsLowComplexity      = 1 << 5

	pnsMinCorrelationEnergy FixpDBL = 0
	pnsNoiseCorrelationThr  FixpDBL = 0x2e147ae1
)

type NoiseParams struct {
	StartSfb                int
	DetectionAlgorithmFlags uint16
	RefPower                FixpDBL
	RefTonality             FixpDBL
	TnsGainThreshold        int
	TnsPNSGainThreshold     int
	MinSfbWidth             int
	PowDistPSDcurve         [maxGroupedSFB]FixpSGL
	GapFillThr              FixpSGL
}

type PNSConfig struct {
	NP                     NoiseParams
	MinCorrelationEnergy   FixpDBL
	NoiseCorrelationThresh FixpDBL
	UsePns                 int
}

type PNSData struct {
	NoiseFuzzyMeasure      [maxGroupedSFB]FixpSGL
	NoiseEnergyCorrelation [maxGroupedSFB]FixpDBL
	PNSFlag                [maxGroupedSFB]int
}

type pnsInfoTabEntry struct {
	startFreq               int
	refPower                FixpSGL
	refTonality             FixpSGL
	tnsGainThreshold        int
	tnsPNSGainThreshold     int
	gapFillThr              FixpSGL
	minSfbWidth             int
	detectionAlgorithmFlags uint16
}

type autoPNSTabEntry struct {
	brFrom int
	brTo   int
	s16000 uint8
	s22050 uint8
	s24000 uint8
	s32000 uint8
	s44100 uint8
	s48000 uint8
}

var (
	levelTableMono = [...]autoPNSTabEntry{
		{0, 11999, 0, 1, 1, 1, 1, 1},
		{12000, 19999, 0, 1, 1, 1, 1, 1},
		{20000, 28999, 0, 2, 1, 1, 1, 1},
		{29000, 40999, 0, 4, 4, 4, 2, 2},
		{41000, 55999, 0, 9, 9, 7, 7, 7},
		{56000, 61999, 0, 0, 0, 0, 9, 9},
		{62000, 75999, 0, 0, 0, 0, 0, 0},
		{76000, 92999, 0, 0, 0, 0, 0, 0},
		{93000, 999999, 0, 0, 0, 0, 0, 0},
	}
	levelTableStereo = [...]autoPNSTabEntry{
		{0, 11999, 0, 1, 1, 1, 1, 1},
		{12000, 19999, 0, 3, 1, 1, 1, 1},
		{20000, 28999, 0, 3, 3, 3, 2, 2},
		{29000, 40999, 0, 7, 6, 6, 5, 5},
		{41000, 55999, 0, 9, 9, 7, 7, 7},
		{56000, 79999, 0, 0, 0, 0, 0, 0},
		{80000, 99999, 0, 0, 0, 0, 0, 0},
		{100000, 999999, 0, 0, 0, 0, 0, 0},
	}
	pnsInfoTab = [...]pnsInfoTabEntry{
		{4000, 0x051f, 0x07ae, 1150, 1200, 0x028f, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS},
		{4000, 0x051f, 0x08f6, 1130, 1300, 0x0666, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS},
		{4100, 0x051f, 0x08f6, 1100, 1400, 0x0ccd, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS},
		{4100, 0x03d7, 0x0ccd, 1100, 1400, 0x1333, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS},
		{4300, 0x03d7, 0x0ccd, 1100, 1400, 0x1333, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
		{5000, 0x03d7, 0x0ccd, 1100, 1400, 0x2000, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
		{5500, 0x03d7, 0x0f5c, 1100, 1400, 0x2ccd, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
		{6000, 0x03d7, 0x0f5c, 1080, 1400, 0x3333, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
		{6000, 0x03d7, 0x11ec, 1070, 1400, 0x399a, 8, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
	}
	levelTableLowComplexity = [...]autoPNSTabEntry{
		{0, 27999, 0, 0, 0, 0, 0, 0},
		{28000, 31999, 0, 2, 2, 2, 2, 2},
		{32000, 47999, 0, 3, 3, 3, 3, 3},
		{48000, 48000, 0, 4, 4, 4, 4, 4},
		{48001, 999999, 0, 0, 0, 0, 0, 0},
	}
	pnsInfoTabLowComplexity = [...]pnsInfoTabEntry{
		{4100, 0x03d7, 0x147b, 1100, 1400, 0x4000, 16, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
		{4100, 0x0666, 0x0ccd, 1410, 1400, 0x4000, 16, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
		{4100, 0x0666, 0x0ccd, 1100, 1400, 0x4000, 16, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
		{4100, 0x199a, 0x0ccd, 1410, 1400, 0x4000, 16, pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow},
	}
)

func FDKaacEncLookUpPnsUse(bitRate int, sampleRate int, numChan int, isLC int) int {
	table := levelTableMono[:]
	if isLC != 0 {
		table = levelTableLowComplexity[:]
	} else if numChan > 1 {
		table = levelTableStereo[:]
	}

	i := 0
	for ; i < len(table); i++ {
		if bitRate >= table[i].brFrom && bitRate <= table[i].brTo {
			break
		}
	}
	if i >= len(table) {
		return pnsTableError
	}
	if len(pnsInfoTab) < i {
		return pnsTableError
	}

	switch sampleRate {
	case 16000:
		return int(table[i].s16000)
	case 22050:
		return int(table[i].s22050)
	case 24000:
		return int(table[i].s24000)
	case 32000:
		return int(table[i].s32000)
	case 44100:
		return int(table[i].s44100)
	case 48000:
		return int(table[i].s48000)
	default:
		if isLC != 0 {
			return int(table[i].s48000)
		}
		return 0
	}
}

func FDKaacEncGetPnsParam(np *NoiseParams, bitRate int, sampleRate int, sfbCnt int, sfbOffset []int, usePns *int, numChan int, isLC int) int {
	checkGetPnsParamInputs(np, bitRate, sampleRate, sfbCnt, sfbOffset, usePns)

	if *usePns <= 0 {
		return AACEncOK
	}

	var pnsInfo []pnsInfoTabEntry
	if isLC != 0 {
		np.DetectionAlgorithmFlags = pnsIsLowComplexity
		pnsInfo = pnsInfoTabLowComplexity[:]
	} else {
		np.DetectionAlgorithmFlags = 0
		pnsInfo = pnsInfoTab[:]
	}

	hUsePns := FDKaacEncLookUpPnsUse(bitRate, sampleRate, numChan, isLC)
	if hUsePns == 0 {
		*usePns = 0
		return AACEncOK
	}
	if hUsePns == pnsTableError || hUsePns > len(pnsInfo) {
		return AACEncPNSTableError
	}
	info := pnsInfo[hUsePns-1]

	np.StartSfb = FDKaacEncFreqToBandWidthRounding(info.startFreq, sampleRate, sfbCnt, sfbOffset)
	np.DetectionAlgorithmFlags |= info.detectionAlgorithmFlags
	np.RefPower = FXSGL2FXDBL(info.refPower)
	np.RefTonality = FXSGL2FXDBL(info.refTonality)
	np.TnsGainThreshold = info.tnsGainThreshold
	np.TnsPNSGainThreshold = info.tnsPNSGainThreshold
	np.MinSfbWidth = info.minSfbWidth
	np.GapFillThr = info.gapFillThr

	for i := 0; i < sfbCnt-1; i++ {
		sfbWidth := sfbOffset[i+1] - sfbOffset[i]
		tmp, qTmp := fPow(np.RefPower, 0, FixpDBL(sfbWidth), DfractBits-1-5)
		np.PowDistPSDcurve[i] = FixpSGL(ScaleValueDBL(tmp, qTmp) >> 16)
	}
	np.PowDistPSDcurve[sfbCnt] = np.PowDistPSDcurve[sfbCnt-1]
	return AACEncOK
}

func FDKaacEncInitPnsConfiguration(pnsConf *PNSConfig, bitRate int, sampleRate int, usePns int, sfbCnt int, sfbOffset []int, numChan int, isLC int) int {
	if pnsConf == nil {
		panic("fdkaac: nil pns configuration")
	}
	*pnsConf = PNSConfig{}
	err := FDKaacEncGetPnsParam(&pnsConf.NP, bitRate, sampleRate, sfbCnt, sfbOffset, &usePns, numChan, isLC)
	if err != AACEncOK {
		return err
	}
	pnsConf.MinCorrelationEnergy = pnsMinCorrelationEnergy
	pnsConf.NoiseCorrelationThresh = pnsNoiseCorrelationThr
	pnsConf.UsePns = usePns
	return AACEncOK
}

func checkGetPnsParamInputs(np *NoiseParams, bitRate int, sampleRate int, sfbCnt int, sfbOffset []int, usePns *int) {
	if np == nil {
		panic("fdkaac: nil pns noise params")
	}
	if bitRate < 0 || sampleRate <= 0 {
		panic("fdkaac: invalid pns rate")
	}
	if sfbCnt <= 0 || sfbCnt >= maxGroupedSFB {
		panic("fdkaac: invalid pns sfb count")
	}
	if len(sfbOffset) < sfbCnt+1 {
		panic("fdkaac: short pns sfb offsets")
	}
	if usePns == nil {
		panic("fdkaac: nil pns use flag")
	}
}
