package fdkaac

type QCMainQuantizeScratch struct {
	BitCounter             BitCounterState
	QuantSpecTmp           [maxSpectralLines]int16
	Iterations             [maxChannelElements]int
	ConstraintsFulfilled   [maxChannelElements]int
	ChConstraintsFulfilled [maxChannelElements][2]int
	CalculateQuant         [maxChannelElements][2]int
}

type ElementBits struct {
	ChBitrateEl     int
	MaxBitsEl       int
	BitResLevelEl   int
	MaxBitResBitsEl int
	RelativeBitsEl  FixpDBL
}

type QCMainQuantizeResult struct {
	QuantizedElements      int
	DynBitsConsumed        int
	SumDynBitsConsumed     int
	MaxValueAll            int
	DecreaseBitConsumption int
	ConstraintsFulfilled   int
	ReductionIterations    int
	GainAdjustments        int
	CrashRecoveryNeeded    int
	BitsToSave             int
}

type QCMainReduceBitConsumptionResult struct {
	IterationsReached   int
	AdjustedChannels    int
	MaxIterationsHit    int
	CrashRecoveryNeeded int
	BitsToSave          int
}

func FDKaacEncQCMainPrepare(
	elInfo *ElementInfo,
	adjThrStateElement *ATSElement,
	psyOutElement *PsyOutElement,
	qcOutElement *QCOutElement,
	aot int,
	syntaxFlags uint32,
	epConfig int8,
) int {
	checkQCMainPrepareInputs(elInfo, adjThrStateElement, psyOutElement, qcOutElement)

	nChannels := elInfo.NChannelsInEl
	psyOutChannel := psyOutElement.PsyOutChannel[:nChannels]
	qcOutChannel := qcOutElement.QCOutChannel[:nChannels]

	FDKaacEncCalcFormFactor(qcOutChannel, psyOutChannel, nChannels)
	FDKaacEncPECalculation(&qcOutElement.PEData, psyOutChannel, qcOutChannel, &psyOutElement.ToolsInfo, adjThrStateElement, nChannels)

	bitDemand, errCode := FDKaacEncChannelElementWrite(nil, elInfo, nil, psyOutElement, psyOutChannel, syntaxFlags, aot, epConfig, 0)
	qcOutElement.StaticBitsUsed = bitDemand
	return errCode
}

func FDKaacEncQCMainQuantizeFrame(
	cm *ChannelMapping,
	psyOutElement []*PsyOutElement,
	qcOut *QCOut,
	qcElement []*QCOutElement,
	adjThrState *AdjThrState,
	elementBits []*ElementBits,
	scratch *QCMainQuantizeScratch,
	invQuant int,
	dZoneQuantEnable int,
	maxIterations int,
	syntaxFlags uint32,
) (QCMainQuantizeResult, int) {
	checkQCMainQuantizeFrameInputs(cm, psyOutElement, qcOut, qcElement, adjThrState, elementBits, scratch, maxIterations)

	result := QCMainQuantizeResult{ConstraintsFulfilled: 1, DecreaseBitConsumption: -1}
	for i := 0; i < cm.NElements; i++ {
		elInfo := cm.ElInfo[i]
		if !fdkaacEncIsAdjustableElement(elInfo.ElType) {
			continue
		}

		nChannels := elInfo.NChannelsInEl
		psyChannels := psyOutElement[i].PsyOutChannel[:nChannels]
		qcChannels := qcElement[i].QCOutChannel[:nChannels]

		FDKaacEncEstimateScaleFactors(psyChannels, qcChannels, invQuant, dZoneQuantEnable, nChannels, scratch.QuantSpecTmp[:])

		scratch.ConstraintsFulfilled[i] = 1
		scratch.Iterations[i] = 0
		for ch := 0; ch < nChannels; ch++ {
			scratch.ChConstraintsFulfilled[i][ch] = 1
			scratch.CalculateQuant[i][ch] = 1
		}

		for {
			if scratch.ConstraintsFulfilled[i] == 0 {
				gainAdjustment := 1
				if result.DecreaseBitConsumption == 0 {
					gainAdjustment = -1
				}
				reduceResult, errCode := FDKaacEncReduceBitConsumption(
					&scratch.Iterations[i],
					maxIterations,
					gainAdjustment,
					scratch.ChConstraintsFulfilled[i][:nChannels],
					scratch.CalculateQuant[i][:nChannels],
					nChannels,
					qcElement[i],
					elementBits[i],
				)
				result.ReductionIterations++
				result.GainAdjustments += reduceResult.AdjustedChannels
				if reduceResult.CrashRecoveryNeeded != 0 {
					result.CrashRecoveryNeeded = 1
					result.BitsToSave = reduceResult.BitsToSave
				}
				if errCode != AACEncOK {
					result.ConstraintsFulfilled = 0
					return result, errCode
				}
			}

			scratch.ConstraintsFulfilled[i] = 1
			for ch := 0; ch < nChannels; ch++ {
				scratch.ChConstraintsFulfilled[i][ch] = 1

				if scratch.CalculateQuant[i][ch] == 0 {
					continue
				}
				qcOutCh := qcChannels[ch]
				psyOutCh := psyChannels[ch]
				scratch.CalculateQuant[i][ch] = 0

				FDKaacEncQuantizeSpectrum(
					psyOutCh.SfbCnt,
					psyOutCh.MaxSfbPerGroup,
					psyOutCh.SfbPerGroup,
					psyOutCh.SfbOffsets[:],
					qcOutCh.MdctSpectrum[:],
					qcOutCh.GlobalGain,
					qcOutCh.Scf[:],
					qcOutCh.QuantSpec[:],
					dZoneQuantEnable,
				)

				maxValue := FDKaacEncCalcMaxValueInSfb(
					psyOutCh.SfbCnt,
					psyOutCh.MaxSfbPerGroup,
					psyOutCh.SfbPerGroup,
					psyOutCh.SfbOffsets[:],
					qcOutCh.QuantSpec[:],
					qcOutCh.MaxValueInSfb[:],
				)
				if maxValue > result.MaxValueAll {
					result.MaxValueAll = maxValue
				}
				if maxValue > maxQuant {
					scratch.ChConstraintsFulfilled[i][ch] = 0
					scratch.ConstraintsFulfilled[i] = 0
					result.ConstraintsFulfilled = 0
					result.DecreaseBitConsumption = 1
				}
			}
			if scratch.ConstraintsFulfilled[i] == 0 {
				continue
			}

			qcElement[i].DynBitsUsed = 0
			for ch := 0; ch < nChannels; ch++ {
				qcOutCh := qcChannels[ch]
				psyOutCh := psyChannels[ch]
				chDynBits := FDKaacEncDynBitCount(
					&scratch.BitCounter,
					qcOutCh.QuantSpec[:],
					qcOutCh.MaxValueInSfb[:],
					qcOutCh.Scf[:],
					psyOutCh.LastWindowSequence,
					psyOutCh.SfbCnt,
					psyOutCh.MaxSfbPerGroup,
					psyOutCh.SfbPerGroup,
					psyOutCh.SfbOffsets[:],
					&qcOutCh.SectionData,
					psyOutCh.NoiseNrg[:],
					psyOutCh.IsBook[:],
					psyOutCh.IsScale[:],
					syntaxFlags,
				)
				qcElement[i].DynBitsUsed += chDynBits
			}

			if adjThrState.AdjThrStateElem[i].DynBitsLast == -1 {
				adjThrState.AdjThrStateElem[i].DynBitsLast = qcElement[i].DynBitsUsed
			}

			maxElementDynBits := nChannels*minBufSizePerEffChan - qcElement[i].StaticBitsUsed - qcElement[i].ExtBitsUsed
			if qcElement[i].DynBitsUsed > maxElementDynBits {
				scratch.ConstraintsFulfilled[i] = 0
				result.ConstraintsFulfilled = 0
				result.DecreaseBitConsumption = 1
				continue
			}
			break
		}

		result.ConstraintsFulfilled = 1
		result.DynBitsConsumed += qcElement[i].DynBitsUsed
		result.QuantizedElements++
	}

	FDKaacEncUpdateUsedDynBits(&qcOut.UsedDynBits, qcElement, cm)
	result.SumDynBitsConsumed = FDKaacEncGetTotalConsumedDynBits([]*QCOut{qcOut}, 1)
	return result, AACEncOK
}

func FDKaacEncReduceBitConsumption(
	iterations *int,
	maxIterations int,
	gainAdjustment int,
	chConstraintsFulfilled []int,
	calculateQuant []int,
	nChannels int,
	qcOutElement *QCOutElement,
	elBits *ElementBits,
) (QCMainReduceBitConsumptionResult, int) {
	checkReduceBitConsumptionInputs(
		iterations, maxIterations, gainAdjustment, chConstraintsFulfilled,
		calculateQuant, nChannels, qcOutElement, elBits,
	)

	result := QCMainReduceBitConsumptionResult{IterationsReached: *iterations}
	if *iterations < maxIterations {
		for ch := 0; ch < nChannels; ch++ {
			if chConstraintsFulfilled[ch] == 0 {
				qcOutElement.QCOutChannel[ch].GlobalGain += gainAdjustment
				calculateQuant[ch] = 1
				result.AdjustedChannels++
			}
		}
	} else if *iterations == maxIterations {
		result.MaxIterationsHit = 1
		if qcOutElement.DynBitsUsed == 0 {
			return result, AACEncQuantError
		}

		bitsToSave := maxInt(
			(qcOutElement.DynBitsUsed+8)-(elBits.BitResLevelEl+qcOutElement.GrantedDynBits),
			(qcOutElement.DynBitsUsed+qcOutElement.StaticBitsUsed+8)-elBits.MaxBitsEl,
		)
		if bitsToSave > 0 {
			result.CrashRecoveryNeeded = 1
			result.BitsToSave = bitsToSave
			return result, AACEncQuantError
		}
		for ch := 0; ch < nChannels; ch++ {
			qcOutElement.QCOutChannel[ch].GlobalGain++
			calculateQuant[ch] = 1
			result.AdjustedChannels++
		}
	} else {
		return result, AACEncQuantError
	}

	(*iterations)++
	result.IterationsReached = *iterations
	return result, AACEncOK
}

func FDKaacEncCalcMaxValueInSfb(
	sfbCnt int,
	maxSfbPerGroup int,
	sfbPerGroup int,
	sfbOffset []int,
	quantSpectrum []int16,
	maxValue []uint32,
) int {
	checkCalcMaxValueInSfbInputs(sfbCnt, maxSfbPerGroup, sfbPerGroup, sfbOffset, quantSpectrum, maxValue)

	maxValueAll := 0
	for sfbOffs := 0; sfbOffs < sfbCnt; sfbOffs += sfbPerGroup {
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			maxThisSfb := 0
			for line := sfbOffset[sfbOffs+sfb]; line < sfbOffset[sfbOffs+sfb+1]; line++ {
				tmp := absInt16(quantSpectrum[line])
				if tmp > maxThisSfb {
					maxThisSfb = tmp
				}
			}
			maxValue[sfbOffs+sfb] = uint32(maxThisSfb)
			if maxThisSfb > maxValueAll {
				maxValueAll = maxThisSfb
			}
		}
	}
	return maxValueAll
}

func FDKaacEncUpdateUsedDynBits(sumDynBitsConsumed *int, qcElement []*QCOutElement, cm *ChannelMapping) int {
	checkUpdateUsedDynBitsInputs(sumDynBitsConsumed, qcElement, cm)

	*sumDynBitsConsumed = 0
	for i := 0; i < cm.NElements; i++ {
		elInfo := cm.ElInfo[i]
		if fdkaacEncIsAdjustableElement(elInfo.ElType) {
			*sumDynBitsConsumed += qcElement[i].DynBitsUsed
		}
	}
	return AACEncOK
}

func FDKaacEncGetTotalConsumedDynBits(qcOut []*QCOut, nSubFrames int) int {
	checkGetTotalConsumedDynBitsInputs(qcOut, nSubFrames)

	totalBits := 0
	for c := 0; c < nSubFrames; c++ {
		if qcOut[c].UsedDynBits == -1 {
			return -1
		}
		totalBits += qcOut[c].UsedDynBits
	}
	return totalBits
}

func checkQCMainPrepareInputs(
	elInfo *ElementInfo,
	adjThrStateElement *ATSElement,
	psyOutElement *PsyOutElement,
	qcOutElement *QCOutElement,
) {
	if elInfo == nil {
		panic("fdkaac: nil QC prepare element info")
	}
	if adjThrStateElement == nil {
		panic("fdkaac: nil QC prepare threshold state")
	}
	if psyOutElement == nil {
		panic("fdkaac: nil QC prepare psy element")
	}
	if qcOutElement == nil {
		panic("fdkaac: nil QC prepare output element")
	}
	if elInfo.ElType != idSCE && elInfo.ElType != idCPE && elInfo.ElType != idLFE {
		panic("fdkaac: invalid QC prepare element type")
	}
	if elInfo.NChannelsInEl != channelElementCount(elInfo.ElType) {
		panic("fdkaac: invalid QC prepare channel count")
	}
	for ch := 0; ch < elInfo.NChannelsInEl; ch++ {
		if psyOutElement.PsyOutChannel[ch] == nil {
			panic("fdkaac: nil QC prepare psy channel")
		}
		if qcOutElement.QCOutChannel[ch] == nil {
			panic("fdkaac: nil QC prepare output channel")
		}
	}
}

func checkQCMainQuantizeFrameInputs(
	cm *ChannelMapping,
	psyOutElement []*PsyOutElement,
	qcOut *QCOut,
	qcElement []*QCOutElement,
	adjThrState *AdjThrState,
	elementBits []*ElementBits,
	scratch *QCMainQuantizeScratch,
	maxIterations int,
) {
	if cm == nil {
		panic("fdkaac: nil QC quantize channel mapping")
	}
	if qcOut == nil {
		panic("fdkaac: nil QC quantize frame output")
	}
	if adjThrState == nil {
		panic("fdkaac: nil QC quantize threshold state")
	}
	if scratch == nil {
		panic("fdkaac: nil QC quantize scratch")
	}
	if maxIterations < 0 {
		panic("fdkaac: invalid QC quantize iteration limit")
	}
	if cm.NElements <= 0 || cm.NElements > maxChannelElements {
		panic("fdkaac: invalid QC quantize element count")
	}
	if len(psyOutElement) < cm.NElements || len(qcElement) < cm.NElements || len(elementBits) < cm.NElements {
		panic("fdkaac: short QC quantize elements")
	}
	for i := 0; i < cm.NElements; i++ {
		elInfo := cm.ElInfo[i]
		if !fdkaacEncIsAdjustableElement(elInfo.ElType) {
			if elInfo.ElType != idDSE {
				panic("fdkaac: unsupported QC quantize element")
			}
			continue
		}
		if elInfo.NChannelsInEl != channelElementCount(elInfo.ElType) {
			panic("fdkaac: invalid QC quantize channel count")
		}
		if psyOutElement[i] == nil {
			panic("fdkaac: nil QC quantize psy element")
		}
		if qcElement[i] == nil {
			panic("fdkaac: nil QC quantize output element")
		}
		if adjThrState.AdjThrStateElem[i] == nil {
			panic("fdkaac: nil QC quantize threshold element")
		}
		if elementBits[i] == nil {
			panic("fdkaac: nil QC quantize element bits")
		}
		for ch := 0; ch < elInfo.NChannelsInEl; ch++ {
			if psyOutElement[i].PsyOutChannel[ch] == nil {
				panic("fdkaac: nil QC quantize psy channel")
			}
			if qcElement[i].QCOutChannel[ch] == nil {
				panic("fdkaac: nil QC quantize output channel")
			}
		}
		if qcElement[i].StaticBitsUsed < 0 || qcElement[i].ExtBitsUsed < 0 {
			panic("fdkaac: invalid QC quantize static bits")
		}
	}
}

func checkReduceBitConsumptionInputs(
	iterations *int,
	maxIterations int,
	gainAdjustment int,
	chConstraintsFulfilled []int,
	calculateQuant []int,
	nChannels int,
	qcOutElement *QCOutElement,
	elBits *ElementBits,
) {
	if iterations == nil {
		panic("fdkaac: nil reduce-bit iteration counter")
	}
	if *iterations < 0 || maxIterations < 0 {
		panic("fdkaac: invalid reduce-bit iteration count")
	}
	if gainAdjustment != -1 && gainAdjustment != 1 {
		panic("fdkaac: invalid reduce-bit gain adjustment")
	}
	if nChannels <= 0 || nChannels > 2 {
		panic("fdkaac: invalid reduce-bit channel count")
	}
	if len(chConstraintsFulfilled) < nChannels || len(calculateQuant) < nChannels {
		panic("fdkaac: short reduce-bit channel controls")
	}
	if qcOutElement == nil {
		panic("fdkaac: nil reduce-bit output element")
	}
	if elBits == nil {
		panic("fdkaac: nil reduce-bit element bits")
	}
	for ch := 0; ch < nChannels; ch++ {
		if chConstraintsFulfilled[ch] != 0 && chConstraintsFulfilled[ch] != 1 {
			panic("fdkaac: invalid reduce-bit channel constraint")
		}
		if calculateQuant[ch] != 0 && calculateQuant[ch] != 1 {
			panic("fdkaac: invalid reduce-bit quantize flag")
		}
		if qcOutElement.QCOutChannel[ch] == nil {
			panic("fdkaac: nil reduce-bit output channel")
		}
	}
	if qcOutElement.DynBitsUsed < 0 || qcOutElement.StaticBitsUsed < 0 {
		panic("fdkaac: invalid reduce-bit bit counts")
	}
}

func checkCalcMaxValueInSfbInputs(
	sfbCnt int,
	maxSfbPerGroup int,
	sfbPerGroup int,
	sfbOffset []int,
	quantSpectrum []int16,
	maxValue []uint32,
) {
	if sfbCnt <= 0 || sfbCnt > maxGroupedSFB || sfbPerGroup <= 0 || sfbCnt%sfbPerGroup != 0 {
		panic("fdkaac: invalid max-value band count")
	}
	if maxSfbPerGroup < 0 || maxSfbPerGroup > sfbPerGroup {
		panic("fdkaac: invalid max-value group width")
	}
	if len(sfbOffset) < sfbCnt+1 || len(maxValue) < sfbCnt {
		panic("fdkaac: short max-value band data")
	}
	maxLine := 0
	for sfbOffs := 0; sfbOffs < sfbCnt; sfbOffs += sfbPerGroup {
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			start := sfbOffset[sfbOffs+sfb]
			stop := sfbOffset[sfbOffs+sfb+1]
			if start < 0 || stop < start {
				panic("fdkaac: invalid max-value offsets")
			}
			if stop > maxLine {
				maxLine = stop
			}
		}
	}
	if len(quantSpectrum) < maxLine {
		panic("fdkaac: short max-value spectrum")
	}
}

func checkUpdateUsedDynBitsInputs(sumDynBitsConsumed *int, qcElement []*QCOutElement, cm *ChannelMapping) {
	if sumDynBitsConsumed == nil {
		panic("fdkaac: nil dynamic-bit sum")
	}
	if cm == nil {
		panic("fdkaac: nil dynamic-bit channel mapping")
	}
	if cm.NElements < 0 || cm.NElements > maxChannelElements || len(qcElement) < cm.NElements {
		panic("fdkaac: invalid dynamic-bit element count")
	}
	for i := 0; i < cm.NElements; i++ {
		if fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) && qcElement[i] == nil {
			panic("fdkaac: nil dynamic-bit element")
		}
	}
}

func checkGetTotalConsumedDynBitsInputs(qcOut []*QCOut, nSubFrames int) {
	if nSubFrames < 0 || len(qcOut) < nSubFrames {
		panic("fdkaac: invalid consumed dynamic-bit frame count")
	}
	for i := 0; i < nSubFrames; i++ {
		if qcOut[i] == nil {
			panic("fdkaac: nil consumed dynamic-bit frame")
		}
	}
}
