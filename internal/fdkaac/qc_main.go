package fdkaac

type QCMainQuantizeScratch struct {
	BitCounter   BitCounterState
	QuantSpecTmp [maxSpectralLines]int16
}

type QCMainQuantizeResult struct {
	QuantizedElements      int
	DynBitsConsumed        int
	SumDynBitsConsumed     int
	MaxValueAll            int
	DecreaseBitConsumption int
	ConstraintsFulfilled   int
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
	scratch *QCMainQuantizeScratch,
	invQuant int,
	dZoneQuantEnable int,
	syntaxFlags uint32,
) (QCMainQuantizeResult, int) {
	checkQCMainQuantizeFrameInputs(cm, psyOutElement, qcOut, qcElement, adjThrState, scratch)

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

		for ch := 0; ch < nChannels; ch++ {
			qcOutCh := qcChannels[ch]
			psyOutCh := psyChannels[ch]

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
				result.ConstraintsFulfilled = 0
				result.DecreaseBitConsumption = 1
				return result, AACEncQuantError
			}
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

		result.DynBitsConsumed += qcElement[i].DynBitsUsed
		result.QuantizedElements++
		maxElementDynBits := nChannels*minBufSizePerEffChan - qcElement[i].StaticBitsUsed - qcElement[i].ExtBitsUsed
		if qcElement[i].DynBitsUsed > maxElementDynBits {
			result.ConstraintsFulfilled = 0
			result.DecreaseBitConsumption = 1
			return result, AACEncQuantError
		}
	}

	FDKaacEncUpdateUsedDynBits(&qcOut.UsedDynBits, qcElement, cm)
	result.SumDynBitsConsumed = FDKaacEncGetTotalConsumedDynBits([]*QCOut{qcOut}, 1)
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
	scratch *QCMainQuantizeScratch,
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
	if cm.NElements <= 0 || cm.NElements > maxChannelElements {
		panic("fdkaac: invalid QC quantize element count")
	}
	if len(psyOutElement) < cm.NElements || len(qcElement) < cm.NElements {
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
