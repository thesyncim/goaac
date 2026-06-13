package fdkaac

type TNSSubblockInfo struct {
	TnsActive      [maxTnsFilters]int
	PredictionGain [maxTnsFilters]int
}

type TNSDataShort struct {
	SubBlockInfo   [transFac]TNSSubblockInfo
	RatioMultTable [transFac][maxSFBShort]FixpDBL
}

type TNSDataLong struct {
	SubBlockInfo   TNSSubblockInfo
	RatioMultTable [maxSFBLong]FixpDBL
}

type TNSDataRaw struct {
	Long  TNSDataLong
	Short TNSDataShort
}

type TNSData struct {
	NumOfSubblocks  int
	DataRaw         TNSDataRaw
	TnsMaxScaleSpec int
	FiltersMerged   int
}

var (
	fdkaacEncTnsEncCoeff3 = [...]FixpSGL{
		STC(0x81f1d201), STC(0x91261481),
		STC(0xadb92301), STC(0xd438af00),
		STC(0x00000000), STC(0x37898080),
		STC(0x64130dff), STC(0x7cca6fff),
	}
	fdkaacEncTnsCoeff3Borders = [...]FixpSGL{
		STC(0x80000001), STC(0x87b826df),
		STC(0x9df24154), STC(0xbfffffe5),
		STC(0xe9c5e578), STC(0x1c7b90f0),
		STC(0x4fce83a9), STC(0x7352f2c3),
	}
	fdkaacEncTnsEncCoeff4 = [...]FixpSGL{
		STC(0x808bc881), STC(0x84e2e581),
		STC(0x8d6b4a01), STC(0x99da9201),
		STC(0xa9c45701), STC(0xbc9dde81),
		STC(0xd1c2d500), STC(0xe87ae540),
		STC(0x00000000), STC(0x1a9cd9c0),
		STC(0x340ff240), STC(0x4b3c8bff),
		STC(0x5f1f5e7f), STC(0x6ed9eb7f),
		STC(0x79bc387f), STC(0x7f4c7e7f),
	}
	fdkaacEncTnsCoeff4Borders = [...]FixpSGL{
		STC(0x80000001), STC(0x822deff0),
		STC(0x88a4bfe6), STC(0x932c159d),
		STC(0xa16827c2), STC(0xb2dcde27),
		STC(0xc6f20b91), STC(0xdcf89c64),
		STC(0xf4308ce1), STC(0x0d613054),
		STC(0x278dde80), STC(0x4000001b),
		STC(0x55a6127b), STC(0x678dde8f),
		STC(0x74ef0ed7), STC(0x7d33f0da),
	}
)

func FDKaacEncTnsSync(
	tnsDataDest *TNSData,
	tnsDataSrc *TNSData,
	tnsInfoDest *TNSInfo,
	tnsInfoSrc *TNSInfo,
	blockTypeDest int,
	blockTypeSrc int,
	tC *TNSConfig,
) {
	checkTnsSyncInputs(tnsDataDest, tnsDataSrc, tnsInfoDest, tnsInfoSrc, blockTypeDest, blockTypeSrc, tC)

	if (blockTypeSrc == ShortWindow && blockTypeDest != ShortWindow) ||
		(blockTypeDest == ShortWindow && blockTypeSrc != ShortWindow) {
		return
	}

	nWindows := 1
	if blockTypeDest == ShortWindow {
		nWindows = transFac
	}

	for w := 0; w < nWindows; w++ {
		sbInfoSrc := tnsSubblockInfo(tnsDataSrc, blockTypeSrc, w)
		sbInfoDest := tnsSubblockInfo(tnsDataDest, blockTypeDest, w)
		doSync := 1
		absDiffSum := 0

		if sbInfoDest.TnsActive[tnsHiFilt] == 0 && sbInfoSrc.TnsActive[tnsHiFilt] == 0 {
			continue
		}

		for i := 0; i < tC.MaxOrder; i++ {
			absDiff := absInt(tnsInfoDest.Coef[w][tnsHiFilt][i] - tnsInfoSrc.Coef[w][tnsHiFilt][i])
			absDiffSum += absDiff
			if absDiff > 1 || absDiffSum > 2 {
				doSync = 0
				break
			}
		}

		if doSync == 0 {
			continue
		}

		if sbInfoSrc.TnsActive[tnsHiFilt] != 0 {
			if sbInfoDest.TnsActive[tnsHiFilt] == 0 ||
				tnsInfoDest.NumOfFilters[w] > tnsInfoSrc.NumOfFilters[w] {
				sbInfoDest.TnsActive[tnsHiFilt] = 1
				tnsInfoDest.NumOfFilters[w] = 1
			}
			tnsDataDest.FiltersMerged = tnsDataSrc.FiltersMerged
			tnsInfoDest.Order[w][tnsHiFilt] = tnsInfoSrc.Order[w][tnsHiFilt]
			tnsInfoDest.Length[w][tnsHiFilt] = tnsInfoSrc.Length[w][tnsHiFilt]
			tnsInfoDest.Direction[w][tnsHiFilt] = tnsInfoSrc.Direction[w][tnsHiFilt]
			tnsInfoDest.CoefCompress[w][tnsHiFilt] = tnsInfoSrc.CoefCompress[w][tnsHiFilt]
			for i := 0; i < tC.MaxOrder; i++ {
				tnsInfoDest.Coef[w][tnsHiFilt][i] = tnsInfoSrc.Coef[w][tnsHiFilt][i]
			}
		} else {
			sbInfoDest.TnsActive[tnsHiFilt] = 0
			tnsInfoDest.NumOfFilters[w] = 0
		}
	}
}

func FDKaacEncTnsEncode(
	tnsInfo *TNSInfo,
	tnsData *TNSData,
	numOfSfb int,
	tC *TNSConfig,
	lowPassLine int,
	spectrum []FixpDBL,
	subBlockNumber int,
	blockType int,
) int {
	checkTnsEncodeInputs(tnsInfo, tnsData, numOfSfb, tC, lowPassLine, spectrum, subBlockNumber, blockType)

	if !tnsHighFilterActive(tnsData, subBlockNumber, blockType) {
		return 1
	}

	checkTnsEncodeActiveInputs(tnsInfo, tC, spectrum, subBlockNumber)

	startLine := tC.LpcStartLine[tnsHiFilt]
	if tnsData.FiltersMerged != 0 {
		startLine = tC.LpcStartLine[tnsLoFilt]
	}
	stopLine := tC.LpcStopLine

	for i := 0; i < tnsInfo.NumOfFilters[subBlockNumber]; i++ {
		var lpcCoeff [tnsMaxOrder]FixpSGL
		var workBuffer [tnsMaxOrder]FixpDBL
		var parcor [tnsMaxOrder]FixpSGL

		order := tnsInfo.Order[subBlockNumber][i]
		fdkaacEncIndex2Parcor(tnsInfo.Coef[subBlockNumber][i][:], parcor[:], order, tC.CoefRes)
		lpcGainFactor := clpcParcorToLpc(parcor[:], lpcCoeff[:], order, workBuffer[:])

		for j := range workBuffer {
			workBuffer[j] = 0
		}
		clpcAnalysis(spectrum[startLine:stopLine], lpcCoeff[:], lpcGainFactor, order, workBuffer[:], nil)

		startLine = tC.LpcStartLine[tnsLoFilt]
		stopLine = tC.LpcStartLine[tnsHiFilt]
	}

	return 0
}

func fdkaacEncParcor2Index(parcor []FixpSGL, index []int, order int, bitsPerCoeff int) {
	checkTnsParcorIndexInputs(parcor, index, order, bitsPerCoeff)
	for i := 0; i < order; i++ {
		if bitsPerCoeff == 3 {
			index[i] = fdkaacEncSearch3(parcor[i])
		} else {
			index[i] = fdkaacEncSearch4(parcor[i])
		}
	}
}

func fdkaacEncIndex2Parcor(index []int, parcor []FixpSGL, order int, bitsPerCoeff int) {
	checkTnsParcorIndexInputs(parcor, index, order, bitsPerCoeff)
	for i := 0; i < order; i++ {
		if bitsPerCoeff == 4 {
			checkTnsCoefficientIndex(index[i], -8, 7)
			parcor[i] = fdkaacEncTnsEncCoeff4[index[i]+8]
		} else {
			checkTnsCoefficientIndex(index[i], -4, 3)
			parcor[i] = fdkaacEncTnsEncCoeff3[index[i]+4]
		}
	}
}

func fdkaacEncSearch3(parcor FixpSGL) int {
	index := 0
	for i := 0; i < len(fdkaacEncTnsCoeff3Borders); i++ {
		if parcor > fdkaacEncTnsCoeff3Borders[i] {
			index = i
		}
	}
	return index - 4
}

func fdkaacEncSearch4(parcor FixpSGL) int {
	index := 0
	for i := 0; i < len(fdkaacEncTnsCoeff4Borders); i++ {
		if parcor > fdkaacEncTnsCoeff4Borders[i] {
			index = i
		}
	}
	return index - 8
}

func clpcParcorToLpc(reflCoeff []FixpSGL, lpcCoeff []FixpSGL, numOfCoeff int, workBuffer []FixpDBL) int {
	checkClpcParcorToLpcInputs(reflCoeff, lpcCoeff, numOfCoeff, workBuffer)
	if numOfCoeff <= 0 {
		return 0
	}

	const par2LpcShiftVal = 6
	maxVal := FixpDBL(0)

	workBuffer[0] = FXSGL2FXDBL(reflCoeff[0]) >> par2LpcShiftVal
	for i := 1; i < numOfCoeff; i++ {
		j := 0
		for ; j < i/2; j++ {
			tmp1 := workBuffer[j]
			tmp2 := workBuffer[i-1-j]
			workBuffer[j] += FMultSD(reflCoeff[i], tmp2)
			workBuffer[i-1-j] += FMultSD(reflCoeff[i], tmp1)
		}
		if i&1 != 0 {
			workBuffer[j] += FMultSD(reflCoeff[i], workBuffer[j])
		}

		workBuffer[i] = FXSGL2FXDBL(reflCoeff[i]) >> par2LpcShiftVal
	}

	for i := 0; i < numOfCoeff; i++ {
		maxVal = maxFixpDBL(maxVal, fixpAbsDBL(workBuffer[i]))
	}

	shiftVal := minInt(CountLeadingBits(maxVal), par2LpcShiftVal)
	for i := 0; i < numOfCoeff; i++ {
		lpcCoeff[i] = FXDBL2FXSGL(workBuffer[i] << uint(shiftVal))
	}

	return par2LpcShiftVal - shiftVal
}

func clpcAnalysis(signal []FixpDBL, lpcCoeff []FixpSGL, lpcCoeffE int, order int, filtState []FixpDBL, filtStateIndex *int) {
	if order <= 0 {
		return
	}
	checkClpcAnalysisInputs(signal, lpcCoeff, order, filtState)

	stateIndex := 0
	if filtStateIndex != nil {
		stateIndex = *filtStateIndex
	}
	if stateIndex < 0 || stateIndex >= order {
		panic("fdkaac: invalid lpc filter state index")
	}

	shift := lpcCoeffE + 1
	if shift < 0 {
		panic("fdkaac: invalid lpc analysis shift")
	}

	var coeff [2 * tnsMaxOrder]FixpSGL
	copy(coeff[:], lpcCoeff[:order])
	copy(coeff[order:], lpcCoeff[:order])

	for j := 0; j < len(signal); j++ {
		pCoeff := coeff[order-stateIndex:]
		tmp := signal[j] >> uint(shift)
		for i := 0; i < order; i++ {
			tmp = FMultAddDiv2SD(tmp, pCoeff[i], filtState[i])
		}

		stateIndex--
		if stateIndex < 0 {
			stateIndex += order
		}
		filtState[stateIndex] = signal[j]
		signal[j] = tmp << uint(shift)
	}

	if filtStateIndex != nil {
		*filtStateIndex = stateIndex
	}
}

func tnsSubblockInfo(data *TNSData, blockType int, window int) *TNSSubblockInfo {
	if blockType == ShortWindow {
		return &data.DataRaw.Short.SubBlockInfo[window]
	}
	return &data.DataRaw.Long.SubBlockInfo
}

func checkTnsSyncInputs(
	tnsDataDest *TNSData,
	tnsDataSrc *TNSData,
	tnsInfoDest *TNSInfo,
	tnsInfoSrc *TNSInfo,
	blockTypeDest int,
	blockTypeSrc int,
	tC *TNSConfig,
) {
	if tnsDataDest == nil {
		panic("fdkaac: nil destination tns data")
	}
	if tnsDataSrc == nil {
		panic("fdkaac: nil source tns data")
	}
	if tnsInfoDest == nil {
		panic("fdkaac: nil destination tns info")
	}
	if tnsInfoSrc == nil {
		panic("fdkaac: nil source tns info")
	}
	if !validSFBWindowSequence(blockTypeDest) || !validSFBWindowSequence(blockTypeSrc) {
		panic("fdkaac: invalid tns sync block type")
	}
	if tC == nil {
		panic("fdkaac: nil tns sync configuration")
	}
	if tC.MaxOrder < 0 || tC.MaxOrder > tnsMaxOrder {
		panic("fdkaac: invalid tns sync order")
	}
}

func checkTnsEncodeInputs(
	tnsInfo *TNSInfo,
	tnsData *TNSData,
	numOfSfb int,
	tC *TNSConfig,
	lowPassLine int,
	spectrum []FixpDBL,
	subBlockNumber int,
	blockType int,
) {
	if tnsInfo == nil {
		panic("fdkaac: nil tns info")
	}
	if tnsData == nil {
		panic("fdkaac: nil tns data")
	}
	if tC == nil {
		panic("fdkaac: nil tns configuration")
	}
	if !validSFBWindowSequence(blockType) {
		panic("fdkaac: invalid tns encode block type")
	}
	if subBlockNumber < 0 || subBlockNumber >= transFac {
		panic("fdkaac: invalid tns subblock")
	}
	if numOfSfb < 0 || numOfSfb > maxGroupedSFB {
		panic("fdkaac: invalid tns sfb count")
	}
	if lowPassLine < 0 || lowPassLine > len(spectrum) {
		panic("fdkaac: invalid tns lowpass line")
	}
	if tC.MaxOrder < 0 || tC.MaxOrder > tnsMaxOrder {
		panic("fdkaac: invalid tns order")
	}
	if tC.CoefRes != 3 && tC.CoefRes != 4 {
		panic("fdkaac: invalid tns coefficient resolution")
	}
}

func checkTnsEncodeActiveInputs(tnsInfo *TNSInfo, tC *TNSConfig, spectrum []FixpDBL, subBlockNumber int) {
	if tC.LpcStartLine[tnsLoFilt] < 0 ||
		tC.LpcStartLine[tnsHiFilt] < 0 ||
		tC.LpcStopLine < 0 ||
		tC.LpcStartLine[tnsLoFilt] > len(spectrum) ||
		tC.LpcStartLine[tnsHiFilt] > len(spectrum) ||
		tC.LpcStopLine > len(spectrum) ||
		tC.LpcStartLine[tnsLoFilt] > tC.LpcStartLine[tnsHiFilt] ||
		tC.LpcStartLine[tnsHiFilt] > tC.LpcStopLine {
		panic("fdkaac: invalid tns lpc lines")
	}

	numFilters := tnsInfo.NumOfFilters[subBlockNumber]
	if numFilters < 0 || numFilters > maxTnsFilters {
		panic("fdkaac: invalid tns filter count")
	}
	for filter := 0; filter < numFilters; filter++ {
		order := tnsInfo.Order[subBlockNumber][filter]
		if order < 0 || order > tC.MaxOrder || order > tnsMaxOrder {
			panic("fdkaac: invalid tns filter order")
		}
		for i := 0; i < order; i++ {
			if tC.CoefRes == 4 {
				checkTnsCoefficientIndex(tnsInfo.Coef[subBlockNumber][filter][i], -8, 7)
			} else {
				checkTnsCoefficientIndex(tnsInfo.Coef[subBlockNumber][filter][i], -4, 3)
			}
		}
	}
}

func checkTnsParcorIndexInputs(parcor []FixpSGL, index []int, order int, bitsPerCoeff int) {
	if order < 0 || order > tnsMaxOrder {
		panic("fdkaac: invalid tns parcor order")
	}
	if bitsPerCoeff != 3 && bitsPerCoeff != 4 {
		panic("fdkaac: invalid tns parcor resolution")
	}
	if len(parcor) < order {
		panic("fdkaac: short tns parcor buffer")
	}
	if len(index) < order {
		panic("fdkaac: short tns coefficient index buffer")
	}
}

func checkTnsCoefficientIndex(index int, minValue int, maxValue int) {
	if index < minValue || index > maxValue {
		panic("fdkaac: invalid tns coefficient index")
	}
}

func checkClpcParcorToLpcInputs(reflCoeff []FixpSGL, lpcCoeff []FixpSGL, numOfCoeff int, workBuffer []FixpDBL) {
	if numOfCoeff < 0 || numOfCoeff > tnsMaxOrder {
		panic("fdkaac: invalid lpc order")
	}
	if len(reflCoeff) < numOfCoeff || len(lpcCoeff) < numOfCoeff || len(workBuffer) < numOfCoeff {
		panic("fdkaac: short lpc buffer")
	}
}

func checkClpcAnalysisInputs(signal []FixpDBL, lpcCoeff []FixpSGL, order int, filtState []FixpDBL) {
	if order < 0 || order > tnsMaxOrder {
		panic("fdkaac: invalid lpc analysis order")
	}
	if len(lpcCoeff) < order || len(filtState) < order {
		panic("fdkaac: short lpc analysis buffer")
	}
}

func tnsHighFilterActive(tnsData *TNSData, subBlockNumber int, blockType int) bool {
	if blockType == ShortWindow {
		return tnsData.DataRaw.Short.SubBlockInfo[subBlockNumber].TnsActive[tnsHiFilt] != 0
	}
	return tnsData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] != 0
}
