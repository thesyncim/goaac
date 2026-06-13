package fdkaac

const (
	tnsHiFilt = 0
	tnsLoFilt = 1

	tnsFilterDirection = 0
	tnsInternalError   = 1
)

type TNSParameterTabulated struct {
	FilterEnabled          [maxTnsFilters]int
	ThreshOn               [maxTnsFilters]int
	FilterStartFreq        [maxTnsFilters]int
	TnsLimitOrder          [maxTnsFilters]int
	TnsFilterDirection     [maxTnsFilters]int
	ACFSplit               [maxTnsFilters]int
	TnsTimeResolution      [maxTnsFilters]FixpDBL
	SeperateFiltersAllowed int
}

type TNSConfig struct {
	ConfTab      TNSParameterTabulated
	IsLowDelay   int
	TnsActive    int
	MaxOrder     int
	CoefRes      int
	ACFWindow    [maxTnsFilters][tnsMaxOrder + 3 + 1]FixpDBL
	LpcStartBand [maxTnsFilters]int
	LpcStartLine [maxTnsFilters]int
	LpcStopBand  int
	LpcStopLine  int
}

type tnsMaxTabEntry struct {
	samplingRate int
	maxBands     [2]int
}

var (
	tnsACFWindowLong = [...]FixpDBL{
		0x7fffffff, 0x7fb80000, 0x7ee00000, 0x7d780000,
		0x7b800000, 0x78f80000, 0x75e00000, 0x72380000,
		0x6e000000, 0x69380000, 0x63e00000, 0x5df80000,
		0x57800000, 0x50780000, 0x48e00000, 0x40b80000,
	}
	tnsACFWindowShort = [...]FixpDBL{
		0x7fffffff, 0x7e000000, 0x78000000, 0x6e000000,
		0x60000000, 0x4e000000, 0x38000000, 0x1e000000,
	}
	tnsMaxBandsTab1024 = [...]tnsMaxTabEntry{
		{96000, [2]int{31, 9}},
		{88200, [2]int{31, 9}},
		{64000, [2]int{34, 10}},
		{48000, [2]int{40, 14}},
		{44100, [2]int{42, 14}},
		{32000, [2]int{51, 14}},
		{24000, [2]int{46, 14}},
		{22050, [2]int{46, 14}},
		{16000, [2]int{42, 14}},
		{12000, [2]int{42, 14}},
		{11025, [2]int{42, 14}},
		{8000, [2]int{39, 14}},
	}
	tnsMaxBandsTab120 = [...]tnsMaxTabEntry{
		{48000, [2]int{12, -1}},
		{44100, [2]int{12, -1}},
		{32000, [2]int{15, -1}},
		{24000, [2]int{15, -1}},
		{22050, [2]int{15, -1}},
	}
	tnsMaxBandsTab128 = tnsMaxBandsTab120
	tnsMaxBandsTab240 = [...]tnsMaxTabEntry{
		{96000, [2]int{22, -1}},
		{48000, [2]int{22, -1}},
		{44100, [2]int{22, -1}},
		{32000, [2]int{21, -1}},
		{24000, [2]int{21, -1}},
		{22050, [2]int{21, -1}},
	}
	tnsMaxBandsTab256 = [...]tnsMaxTabEntry{
		{96000, [2]int{25, -1}},
		{48000, [2]int{25, -1}},
		{44100, [2]int{25, -1}},
		{32000, [2]int{24, -1}},
		{24000, [2]int{24, -1}},
		{22050, [2]int{24, -1}},
	}
	tnsMaxBandsTab480 = [...]tnsMaxTabEntry{
		{48000, [2]int{31, -1}},
		{44100, [2]int{32, -1}},
		{32000, [2]int{37, -1}},
		{24000, [2]int{30, -1}},
		{22050, [2]int{30, -1}},
	}
	tnsMaxBandsTab512 = [...]tnsMaxTabEntry{
		{48000, [2]int{31, -1}},
		{44100, [2]int{32, -1}},
		{32000, [2]int{37, -1}},
		{24000, [2]int{31, -1}},
		{22050, [2]int{31, -1}},
	}
)

func FDKaacEncInitTnsConfiguration(bitRate int, sampleRate int, channels int, blockType int, granuleLength int, isLowDelay int, ldSbrPresent int, tC *TNSConfig, pC *PsyConfiguration, active int, useTnsPeak int) int {
	_ = ldSbrPresent
	_ = useTnsPeak
	checkInitTnsConfigurationInputs(blockType, tC, pC)

	if channels <= 0 {
		return tnsInternalError
	}
	*tC = TNSConfig{}
	tC.IsLowDelay = isLowDelay
	if active != 0 {
		tC.TnsActive = 1
	}
	if blockType == ShortWindow {
		tC.MaxOrder = 5
	} else {
		tC.MaxOrder = 12
	}
	if bitRate < 16000 {
		tC.MaxOrder -= 2
	}
	if blockType == ShortWindow {
		tC.CoefRes = 3
	} else {
		tC.CoefRes = 4
	}

	stopBand := getTnsMaxBands(sampleRate, granuleLength, blockType == ShortWindow)
	if stopBand < 0 {
		return tnsInternalError
	}
	if granuleLength != 960 && granuleLength != 1024 {
		tC.TnsActive = 0
		return AACEncInvalidFrameLength
	}

	tC.LpcStopBand = minInt(stopBand, pC.SfbActive)
	tC.LpcStopLine = pC.SfbOffset[tC.LpcStopBand]

	if blockType == ShortWindow {
		tC.LpcStartBand[tnsLoFilt] = 0
	} else if sampleRate < 9391 {
		tC.LpcStartBand[tnsLoFilt] = 2
	} else if sampleRate < 18783 {
		tC.LpcStartBand[tnsLoFilt] = 4
	} else {
		tC.LpcStartBand[tnsLoFilt] = 8
	}
	tC.LpcStartLine[tnsLoFilt] = pC.SfbOffset[tC.LpcStartBand[tnsLoFilt]]

	i := tC.LpcStopBand
	quarterLine := tC.LpcStartLine[tnsLoFilt] + (tC.LpcStopLine-tC.LpcStartLine[tnsLoFilt])/4
	for i > 0 && pC.SfbOffset[i] > quarterLine {
		i--
	}
	tC.LpcStartBand[tnsHiFilt] = i
	tC.LpcStartLine[tnsHiFilt] = pC.SfbOffset[i]

	tC.ConfTab.ThreshOn[tnsHiFilt] = 1437
	tC.ConfTab.ThreshOn[tnsLoFilt] = 1500
	tC.ConfTab.TnsLimitOrder[tnsHiFilt] = tC.MaxOrder
	tC.ConfTab.TnsLimitOrder[tnsLoFilt] = maxInt(0, tC.MaxOrder-7)
	tC.ConfTab.TnsFilterDirection[tnsHiFilt] = tnsFilterDirection
	tC.ConfTab.TnsFilterDirection[tnsLoFilt] = tnsFilterDirection
	tC.ConfTab.ACFSplit[tnsHiFilt] = -1
	tC.ConfTab.ACFSplit[tnsLoFilt] = -1
	tC.ConfTab.FilterEnabled[tnsHiFilt] = 1
	tC.ConfTab.FilterEnabled[tnsLoFilt] = 1
	tC.ConfTab.SeperateFiltersAllowed = 1

	if blockType == ShortWindow {
		copy(tC.ACFWindow[tnsHiFilt][:], tnsACFWindowShort[:])
		copy(tC.ACFWindow[tnsLoFilt][:], tnsACFWindowShort[:])
	} else {
		copy(tC.ACFWindow[tnsHiFilt][:], tnsACFWindowLong[:])
		copy(tC.ACFWindow[tnsLoFilt][:], tnsACFWindowLong[:])
	}
	return AACEncOK
}

func FDKaacEncFreqToBandWidthRounding(freq int, fs int, numOfBands int, bandStartOffset []int) int {
	checkFreqToBandWidthInputs(freq, fs, numOfBands, bandStartOffset)

	lineNumber := (freq*bandStartOffset[numOfBands]*4/fs + 1) / 2
	if lineNumber >= bandStartOffset[numOfBands] {
		return numOfBands
	}

	band := 0
	for ; band < numOfBands; band++ {
		if bandStartOffset[band+1] > lineNumber {
			break
		}
	}
	if lineNumber-bandStartOffset[band] > bandStartOffset[band+1]-lineNumber {
		band++
	}
	return band
}

func getTnsMaxBands(sampleRate int, granuleLength int, isShortBlock bool) int {
	var table []tnsMaxTabEntry
	switch granuleLength {
	case 960, 1024:
		table = tnsMaxBandsTab1024[:]
	case 120:
		table = tnsMaxBandsTab120[:]
	case 128:
		table = tnsMaxBandsTab128[:]
	case 240:
		table = tnsMaxBandsTab240[:]
	case 256:
		table = tnsMaxBandsTab256[:]
	case 480:
		table = tnsMaxBandsTab480[:]
	case 512:
		table = tnsMaxBandsTab512[:]
	default:
		return -1
	}

	shortIndex := 0
	if isShortBlock {
		shortIndex = 1
	}
	numBands := -1
	for i := range table {
		numBands = table[i].maxBands[shortIndex]
		if sampleRate >= table[i].samplingRate {
			break
		}
	}
	return numBands
}

func checkInitTnsConfigurationInputs(blockType int, tC *TNSConfig, pC *PsyConfiguration) {
	if !validSFBWindowSequence(blockType) {
		panic("fdkaac: invalid tns block type")
	}
	if tC == nil {
		panic("fdkaac: nil tns configuration")
	}
	if pC == nil {
		panic("fdkaac: nil tns psy configuration")
	}
	if pC.SfbActive < 0 || pC.SfbActive > maxSFB {
		panic("fdkaac: invalid tns psy active band count")
	}
}

func checkFreqToBandWidthInputs(freq int, fs int, numOfBands int, bandStartOffset []int) {
	if freq < 0 || fs <= 0 {
		panic("fdkaac: invalid tns frequency")
	}
	if numOfBands < 0 || numOfBands > maxSFB {
		panic("fdkaac: invalid tns band count")
	}
	if len(bandStartOffset) < numOfBands+1 {
		panic("fdkaac: short tns band offsets")
	}
}
