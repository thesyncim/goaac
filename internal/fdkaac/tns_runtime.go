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

const tnsHlmMinNrg = FixpDBL(0x00000008)

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

func FDKaacEncTnsDetect(
	tnsData *TNSData,
	tC *TNSConfig,
	tnsInfo *TNSInfo,
	sfbCnt int,
	spectrum []FixpDBL,
	subBlockNumber int,
	blockType int,
) int {
	checkTnsDetectInputs(tnsData, tC, tnsInfo, sfbCnt, spectrum, subBlockNumber, blockType)

	var rxx1 [tnsMaxOrder + 1]FixpDBL
	var rxx2 [tnsMaxOrder + 1]FixpDBL
	var parcor [tnsMaxOrder]FixpSGL

	tsbi := tnsSubblockInfo(tnsData, blockType, subBlockNumber)
	tnsData.FiltersMerged = 0
	tsbi.TnsActive[tnsHiFilt] = 0
	tsbi.PredictionGain[tnsHiFilt] = 1000
	tsbi.TnsActive[tnsLoFilt] = 0
	tsbi.PredictionGain[tnsLoFilt] = 1000

	tnsInfo.NumOfFilters[subBlockNumber] = 0
	tnsInfo.CoefRes[subBlockNumber] = tC.CoefRes
	for i := 0; i < tC.MaxOrder; i++ {
		tnsInfo.Coef[subBlockNumber][tnsHiFilt][i] = 0
		tnsInfo.Coef[subBlockNumber][tnsLoFilt][i] = 0
	}
	tnsInfo.Length[subBlockNumber][tnsHiFilt] = 0
	tnsInfo.Length[subBlockNumber][tnsLoFilt] = 0
	tnsInfo.Order[subBlockNumber][tnsHiFilt] = 0
	tnsInfo.Order[subBlockNumber][tnsLoFilt] = 0

	if tC.TnsActive == 0 || tC.MaxOrder <= 0 {
		return 0
	}

	fdkaacEncMergedAutoCorrelation(spectrum, tC, rxx1[:], rxx2[:])

	limitHi := tC.ConfTab.TnsLimitOrder[tnsHiFilt]
	var predictionGainM FixpDBL
	var predictionGainE int
	clpcAutoToParcor(rxx2[:], 0, parcor[:], limitHi, &predictionGainM, &predictionGainE)
	tsbi.PredictionGain[tnsHiFilt] = fdkaacEncPredictionGain1000(predictionGainM, predictionGainE)

	fdkaacEncParcor2Index(parcor[:], tnsInfo.Coef[subBlockNumber][tnsHiFilt][:], limitHi, tC.CoefRes)

	i := limitHi - 1
	for ; i >= 0; i-- {
		if tnsInfo.Coef[subBlockNumber][tnsHiFilt][i] != 0 {
			break
		}
	}
	tnsInfo.Order[subBlockNumber][tnsHiFilt] = i + 1

	sumSqrCoef := 0
	for ; i >= 0; i-- {
		coef := tnsInfo.Coef[subBlockNumber][tnsHiFilt][i]
		sumSqrCoef += coef * coef
	}

	tnsInfo.Direction[subBlockNumber][tnsHiFilt] = tC.ConfTab.TnsFilterDirection[tnsHiFilt]
	tnsInfo.Length[subBlockNumber][tnsHiFilt] = sfbCnt - tC.LpcStartBand[tnsHiFilt]

	if tsbi.PredictionGain[tnsHiFilt] <= tC.ConfTab.ThreshOn[tnsHiFilt] &&
		sumSqrCoef <= (limitHi/2+2) {
		return 0
	}

	tsbi.TnsActive[tnsHiFilt] = 1
	tnsInfo.NumOfFilters[subBlockNumber]++

	if blockType == ShortWindow ||
		tC.ConfTab.FilterEnabled[tnsLoFilt] == 0 ||
		tC.ConfTab.SeperateFiltersAllowed == 0 {
		return 0
	}

	limitLo := tC.ConfTab.TnsLimitOrder[tnsLoFilt]
	clpcAutoToParcor(rxx1[:], 0, parcor[:], limitLo, &predictionGainM, &predictionGainE)
	predGain := fdkaacEncPredictionGain1000(predictionGainM, predictionGainE)

	fdkaacEncParcor2Index(parcor[:], tnsInfo.Coef[subBlockNumber][tnsLoFilt][:], limitLo, tC.CoefRes)

	i = limitLo - 1
	for ; i >= 0; i-- {
		if tnsInfo.Coef[subBlockNumber][tnsLoFilt][i] != 0 {
			break
		}
	}
	tnsInfo.Order[subBlockNumber][tnsLoFilt] = i + 1

	sumSqrCoef = 0
	for ; i >= 0; i-- {
		coef := tnsInfo.Coef[subBlockNumber][tnsLoFilt][i]
		sumSqrCoef += coef * coef
	}

	tnsInfo.Direction[subBlockNumber][tnsLoFilt] = tC.ConfTab.TnsFilterDirection[tnsLoFilt]
	tnsInfo.Length[subBlockNumber][tnsLoFilt] = tC.LpcStartBand[tnsHiFilt] - tC.LpcStartBand[tnsLoFilt]

	if ((predGain > tC.ConfTab.ThreshOn[tnsLoFilt]) &&
		(predGain < 16000*limitLo)) ||
		((sumSqrCoef > 9) && (sumSqrCoef < 22*limitLo)) {
		tsbi.TnsActive[tnsLoFilt] = 1
		sumSqrCoef = 0
		for i = 0; i < limitLo; i++ {
			sumSqrCoef += absInt(tnsInfo.Coef[subBlockNumber][tnsHiFilt][i] -
				tnsInfo.Coef[subBlockNumber][tnsLoFilt][i])
		}
		if sumSqrCoef < 2 &&
			tnsInfo.Direction[subBlockNumber][tnsLoFilt] == tnsInfo.Direction[subBlockNumber][tnsHiFilt] {
			tnsData.FiltersMerged = 1
			tnsInfo.Length[subBlockNumber][tnsHiFilt] = sfbCnt - tC.LpcStartBand[tnsLoFilt]
			for ; i < tnsInfo.Order[subBlockNumber][tnsHiFilt]; i++ {
				if absInt(tnsInfo.Coef[subBlockNumber][tnsHiFilt][i]) > 1 {
					break
				}
			}
			for i--; i >= 0; i-- {
				if tnsInfo.Coef[subBlockNumber][tnsHiFilt][i] != 0 {
					break
				}
			}
			if i < tnsInfo.Order[subBlockNumber][tnsHiFilt] {
				tnsInfo.Order[subBlockNumber][tnsHiFilt] = i + 1
			}
		} else {
			tnsInfo.NumOfFilters[subBlockNumber]++
		}
	}
	tsbi.PredictionGain[tnsLoFilt] = predGain

	return 0
}

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

func fdkaacEncMergedAutoCorrelation(spectrum []FixpDBL, tC *TNSConfig, rxx1 []FixpDBL, rxx2 []FixpDBL) {
	checkTnsMergedAutoCorrelationInputs(spectrum, tC, rxx1, rxx2)

	var pSpectrum [maxSpectralLines]FixpDBL

	var idx0, idx1, idx2, idx3, idx4 int
	if tC.ConfTab.ACFSplit[tnsLoFilt] == -1 || tC.ConfTab.ACFSplit[tnsHiFilt] == -1 {
		idx0 = tC.LpcStartLine[tnsLoFilt]
		width := tC.LpcStopLine - tC.LpcStartLine[tnsLoFilt]
		idx1 = idx0 + width/4
		idx2 = idx0 + width/2
		idx3 = idx0 + width*3/4
		idx4 = tC.LpcStopLine
	} else {
		width := (tC.LpcStopLine - tC.LpcStartLine[tnsHiFilt]) / 3
		idx0 = tC.LpcStartLine[tnsLoFilt]
		idx1 = tC.LpcStartLine[tnsHiFilt]
		idx2 = idx1 + width
		idx3 = idx2 + width
		idx4 = tC.LpcStopLine
	}

	sc1 := fdkaacEncScaleUpSpectrum(pSpectrum[:], spectrum, idx0, idx1)
	sc2 := fdkaacEncScaleUpSpectrum(pSpectrum[:], spectrum, idx1, idx2)
	sc3 := fdkaacEncScaleUpSpectrum(pSpectrum[:], spectrum, idx2, idx3)
	sc4 := fdkaacEncScaleUpSpectrum(pSpectrum[:], spectrum, idx3, idx4)

	nsc1 := fdkaacEncAutoCorrSumScale(idx1 - idx0)
	nsc2 := fdkaacEncAutoCorrSumScale(idx2 - idx1)
	nsc3 := fdkaacEncAutoCorrSumScale(idx3 - idx2)
	nsc4 := fdkaacEncAutoCorrSumScale(idx4 - idx3)

	rxx1_0 := fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx0, idx1, 0, nsc1)
	rxx2_0 := fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx1, idx2, 0, nsc2)
	rxx3_0 := fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx2, idx3, 0, nsc3)
	rxx4_0 := fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx3, idx4, 0, nsc4)

	if rxx1_0 != 0 {
		scFac1 := -1
		fac1 := fdkaacEncAutoCorrNormFac(rxx1_0, -2*sc1+nsc1, &scFac1)
		rxx1[0] = ScaleValueDBL(FMultDD(rxx1_0, fac1), scFac1)

		if tC.IsLowDelay != 0 {
			for lag := 1; lag <= tC.MaxOrder; lag++ {
				x1 := fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx0, idx1, lag, nsc1)
				rxx1[lag] = FMultDD(ScaleValueDBL(FMultDD(x1, fac1), scFac1), tC.ACFWindow[tnsLoFilt][lag])
			}
		} else {
			for lag := 1; lag <= tC.MaxOrder; lag++ {
				if 3*lag <= tC.MaxOrder+3 {
					x1 := fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx0, idx1, lag, nsc1)
					rxx1[lag] = FMultDD(ScaleValueDBL(FMultDD(x1, fac1), scFac1), tC.ACFWindow[tnsLoFilt][3*lag])
				}
			}
		}
	}

	if rxx2_0 == 0 && rxx3_0 == 0 && rxx4_0 == 0 {
		return
	}

	var fac2, fac3, fac4 FixpDBL
	scFac2, scFac3, scFac4 := 0, 0, 0
	if rxx2_0 != 0 {
		fac2 = fdkaacEncAutoCorrNormFac(rxx2_0, -2*sc2+nsc2, &scFac2)
		scFac2 -= 2
	}
	if rxx3_0 != 0 {
		fac3 = fdkaacEncAutoCorrNormFac(rxx3_0, -2*sc3+nsc3, &scFac3)
		scFac3 -= 2
	}
	if rxx4_0 != 0 {
		fac4 = fdkaacEncAutoCorrNormFac(rxx4_0, -2*sc4+nsc4, &scFac4)
		scFac4 -= 2
	}

	rxx2[0] = ScaleValueDBL(FMultDD(rxx2_0, fac2), scFac2) +
		ScaleValueDBL(FMultDD(rxx3_0, fac3), scFac3) +
		ScaleValueDBL(FMultDD(rxx4_0, fac4), scFac4)

	for lag := 1; lag <= tC.MaxOrder; lag++ {
		x2 := ScaleValueDBL(FMultDD(fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx1, idx2, lag, nsc2), fac2), scFac2) +
			ScaleValueDBL(FMultDD(fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx2, idx3, lag, nsc3), fac3), scFac3) +
			ScaleValueDBL(FMultDD(fdkaacEncCalcAutoCorrValue(pSpectrum[:], idx3, idx4, lag, nsc4), fac4), scFac4)

		rxx2[lag] = FMultDD(x2, tC.ACFWindow[tnsHiFilt][lag])
	}
}

func fdkaacEncScaleUpSpectrum(dest []FixpDBL, src []FixpDBL, startLine int, stopLine int) int {
	maxVal := FixpDBL(0)
	for i := startLine; i < stopLine; i++ {
		maxVal = maxFixpDBL(maxVal, fixpAbsDBL(src[i]))
	}
	scale := CountLeadingBits(maxVal)
	for i := startLine; i < stopLine; i++ {
		dest[i] = src[i] << uint(scale)
	}
	return scale
}

func fdkaacEncAutoCorrSumScale(length int) int {
	scale := 1
	for (1 << uint(scale)) < length {
		scale++
	}
	return scale
}

func fdkaacEncCalcAutoCorrValue(spectrum []FixpDBL, startLine int, stopLine int, lag int, scale int) FixpDBL {
	result := FixpDBL(0)
	if lag == 0 {
		for i := startLine; i < stopLine; i++ {
			result += FixPow2D(spectrum[i]) >> uint(scale)
		}
		return result
	}

	for i := startLine; i < stopLine-lag; i++ {
		result += FMultDD(spectrum[i], spectrum[i+lag]) >> uint(scale)
	}
	return result
}

func fdkaacEncAutoCorrNormFac(value FixpDBL, scale int, sc *int) FixpDBL {
	if sc == nil {
		panic("fdkaac: nil tns autocorrelation scale")
	}

	var a, b FixpDBL
	if scale >= 0 {
		a = value
		b = tnsHlmMinNrg >> uint(minInt(DfractBits-1, scale))
	} else {
		a = value >> uint(minInt(DfractBits-1, -scale))
		b = tnsHlmMinNrg
	}

	if a > b {
		shift := 0
		tmp := invSqrtNorm2(value, &shift)
		*sc += 2 * shift
		return FMultDD(tmp, tmp)
	}

	*sc += scale + 28
	return MaxValDBL
}

func clpcAutoToParcor(
	acorr []FixpDBL,
	acorrE int,
	reflCoeff []FixpSGL,
	numOfCoeff int,
	predictionGainM *FixpDBL,
	predictionGainE *int,
) {
	_ = acorrE
	checkClpcAutoToParcorInputs(acorr, reflCoeff, numOfCoeff, predictionGainM, predictionGainE)
	for i := 0; i < numOfCoeff; i++ {
		reflCoeff[i] = 0
	}

	autoCorr0 := acorr[0]
	if autoCorr0 == 0 {
		if predictionGainM != nil {
			*predictionGainM = halfDBL
			*predictionGainE = 1
		}
		return
	}

	var parcorWorkBuffer [tnsMaxOrder]FixpDBL
	copy(parcorWorkBuffer[:], acorr[1:1+numOfCoeff])

	workOffset := 0
	for i := 0; i < numOfCoeff; i++ {
		sign := int32(parcorWorkBuffer[workOffset]) >> (DfractBits - 1)
		tmp := FixpDBL(int32(parcorWorkBuffer[workOffset]) ^ sign)
		if acorr[0] < tmp {
			break
		}

		tmp = FixpDBL(int32(schurDiv(tmp, acorr[0], FractBits)) ^ (^sign))
		reflCoeff[i] = FXDBL2FXSGL(tmp)

		for j := numOfCoeff - i - 1; j >= 0; j-- {
			accu1 := FMultDD(tmp, acorr[j])
			accu2 := FMultDD(tmp, parcorWorkBuffer[workOffset+j])
			parcorWorkBuffer[workOffset+j] += accu1
			acorr[j] += accu2
		}
		if acorr[0] == 0 {
			break
		}
		workOffset++
	}

	if predictionGainM == nil {
		return
	}
	if acorr[0] > 0 {
		*predictionGainM, *predictionGainE = fDivNormSignedExp(autoCorr0, acorr[0])
	} else {
		*predictionGainM = 0
		*predictionGainE = 0
	}
}

func fdkaacEncPredictionGain1000(predictionGainM FixpDBL, predictionGainE int) int {
	return int(fdkaacEncFMultNormExp(predictionGainM, predictionGainE, FixpDBL(1000), 31, 31))
}

func fdkaacEncFMultNormExp(f1M FixpDBL, f1E int, f2M FixpDBL, f2E int, resultE int) FixpDBL {
	m, e := fMultNorm(f1M, f2M)
	return ScaleValueSaturateDBL(m, e+f1E+f2E-resultE)
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

func checkTnsDetectInputs(
	tnsData *TNSData,
	tC *TNSConfig,
	tnsInfo *TNSInfo,
	sfbCnt int,
	spectrum []FixpDBL,
	subBlockNumber int,
	blockType int,
) {
	if tnsData == nil {
		panic("fdkaac: nil tns data")
	}
	if tC == nil {
		panic("fdkaac: nil tns configuration")
	}
	if tnsInfo == nil {
		panic("fdkaac: nil tns info")
	}
	if !validSFBWindowSequence(blockType) {
		panic("fdkaac: invalid tns detect block type")
	}
	if subBlockNumber < 0 || subBlockNumber >= transFac {
		panic("fdkaac: invalid tns detect subblock")
	}
	if sfbCnt < 0 || sfbCnt > maxGroupedSFB {
		panic("fdkaac: invalid tns detect sfb count")
	}
	if tC.MaxOrder < 0 || tC.MaxOrder > tnsMaxOrder {
		panic("fdkaac: invalid tns detect order")
	}
	if tC.CoefRes != 3 && tC.CoefRes != 4 {
		panic("fdkaac: invalid tns detect coefficient resolution")
	}
	for filt := 0; filt < maxTnsFilters; filt++ {
		if tC.ConfTab.TnsLimitOrder[filt] < 0 || tC.ConfTab.TnsLimitOrder[filt] > tC.MaxOrder {
			panic("fdkaac: invalid tns detect limit order")
		}
	}
	if tC.TnsActive != 0 && tC.MaxOrder > 0 {
		checkTnsDetectActiveInputs(tC, spectrum, sfbCnt)
	}
}

func checkTnsDetectActiveInputs(tC *TNSConfig, spectrum []FixpDBL, sfbCnt int) {
	if tC.LpcStartLine[tnsLoFilt] < 0 ||
		tC.LpcStartLine[tnsHiFilt] < 0 ||
		tC.LpcStopLine < 0 ||
		tC.LpcStartLine[tnsLoFilt] > tC.LpcStartLine[tnsHiFilt] ||
		tC.LpcStartLine[tnsHiFilt] > tC.LpcStopLine ||
		tC.LpcStopLine > len(spectrum) ||
		tC.LpcStopLine > maxSpectralLines {
		panic("fdkaac: invalid tns detect lpc lines")
	}
	if tC.LpcStartBand[tnsLoFilt] < 0 ||
		tC.LpcStartBand[tnsHiFilt] < 0 ||
		tC.LpcStopBand < 0 ||
		tC.LpcStartBand[tnsLoFilt] > tC.LpcStartBand[tnsHiFilt] ||
		tC.LpcStartBand[tnsHiFilt] > sfbCnt ||
		tC.LpcStopBand > sfbCnt {
		panic("fdkaac: invalid tns detect lpc bands")
	}
	if tC.ConfTab.ACFSplit[tnsLoFilt] != -1 &&
		tC.ConfTab.ACFSplit[tnsHiFilt] != -1 &&
		(tC.ConfTab.ACFSplit[tnsLoFilt] != 1 || tC.ConfTab.ACFSplit[tnsHiFilt] != 3) {
		panic("fdkaac: invalid tns autocorrelation split")
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

func checkTnsMergedAutoCorrelationInputs(spectrum []FixpDBL, tC *TNSConfig, rxx1 []FixpDBL, rxx2 []FixpDBL) {
	if tC == nil {
		panic("fdkaac: nil tns autocorrelation configuration")
	}
	if tC.MaxOrder < 0 || tC.MaxOrder > tnsMaxOrder {
		panic("fdkaac: invalid tns autocorrelation order")
	}
	if len(rxx1) < tC.MaxOrder+1 || len(rxx2) < tC.MaxOrder+1 {
		panic("fdkaac: short tns autocorrelation output")
	}
	checkTnsDetectActiveInputs(tC, spectrum, maxGroupedSFB)
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

func checkClpcAutoToParcorInputs(acorr []FixpDBL, reflCoeff []FixpSGL, numOfCoeff int, predictionGainM *FixpDBL, predictionGainE *int) {
	if numOfCoeff < 0 || numOfCoeff > tnsMaxOrder {
		panic("fdkaac: invalid lpc autocorrelation order")
	}
	if len(acorr) < numOfCoeff+1 || len(reflCoeff) < numOfCoeff {
		panic("fdkaac: short lpc autocorrelation buffer")
	}
	if (predictionGainM == nil) != (predictionGainE == nil) {
		panic("fdkaac: incomplete lpc prediction gain output")
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
