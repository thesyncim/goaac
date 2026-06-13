package fdkaac

type QCBitrateMode int

const (
	QCBitrateModeInvalid QCBitrateMode = -1
	QCBitrateModeCBR     QCBitrateMode = 0
	QCBitrateModeVBR1    QCBitrateMode = 1
	QCBitrateModeVBR2    QCBitrateMode = 2
	QCBitrateModeVBR3    QCBitrateMode = 3
	QCBitrateModeVBR4    QCBitrateMode = 4
	QCBitrateModeVBR5    QCBitrateMode = 5
	QCBitrateModeFF      QCBitrateMode = 6
	QCBitrateModeSFR     QCBitrateMode = 7
)

type QCKernel struct {
	GlobHdrBits     int
	MaxBitsPerFrame int
	MinBitsPerFrame int
	BitrateMode     QCBitrateMode
	BitResMode      BitresMode
	BitResTot       int
	BitResTotMax    int
	MaxBitFac       FixpDBL
}

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

type QCPrepareBitDistributionResult struct {
	TotalAvailableBits  int
	AvgTotalDynBits     int
	DistributedBits     int
	DistributedElements int
	TotalGrantedPeCorr  int
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

func FDKaacEncDistributeElementDynBits(qcElement []*QCOutElement, cm *ChannelMapping, elementBits []*ElementBits, codeBits int) int {
	checkDistributeElementDynBitsInputs(qcElement, cm, elementBits)

	totalBits := 0
	firstAudio := -1
	for i := cm.NElements - 1; i >= 0; i-- {
		if fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
			if firstAudio < 0 {
				firstAudio = i
			}
			qcElement[i].GrantedDynBits = maxInt(0, FMultI(elementBits[i].RelativeBitsEl, codeBits))
			totalBits += qcElement[i].GrantedDynBits
		}
	}
	if firstAudio < 0 {
		return AACEncOK
	}

	if codeBits != totalBits {
		elMaxBits := firstAudio
		elMinBits := firstAudio
		for i := cm.NElements - 1; i >= 0; i-- {
			if fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
				if qcElement[i].GrantedDynBits > qcElement[elMaxBits].GrantedDynBits {
					elMaxBits = i
				}
				if qcElement[i].GrantedDynBits < qcElement[elMinBits].GrantedDynBits {
					elMinBits = i
				}
			}
		}
		if codeBits-totalBits > 0 {
			qcElement[elMinBits].GrantedDynBits += codeBits - totalBits
		} else {
			qcElement[elMaxBits].GrantedDynBits += codeBits - totalBits
		}
	}
	return AACEncOK
}

func FDKaacEncBitResRedistribution(qcKernel *QCKernel, cm *ChannelMapping, elementBits []*ElementBits, avgTotalBits int) int {
	checkBitResRedistributionInputs(qcKernel, cm, elementBits)

	if qcKernel.BitResTot < 0 {
		return AACEncBitresTooLow
	}
	if qcKernel.BitResTot > qcKernel.BitResTotMax {
		return AACEncBitresTooHigh
	}

	totalBits := 0
	totalBitsMax := 0
	totalBitreservoir := minInt(qcKernel.BitResTot, qcKernel.MaxBitsPerFrame-avgTotalBits)
	totalBitreservoirMax := minInt(qcKernel.BitResTotMax, qcKernel.MaxBitsPerFrame-avgTotalBits)

	for i := cm.NElements - 1; i >= 0; i-- {
		if fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
			elementBits[i].BitResLevelEl = FMultI(elementBits[i].RelativeBitsEl, totalBitreservoir)
			totalBits += elementBits[i].BitResLevelEl

			elementBits[i].MaxBitResBitsEl = FMultI(elementBits[i].RelativeBitsEl, totalBitreservoirMax)
			totalBitsMax += elementBits[i].MaxBitResBitsEl
		}
	}

	for i := 0; i < cm.NElements; i++ {
		if fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
			deltaBits := maxInt(totalBitreservoir-totalBits, -elementBits[i].BitResLevelEl)
			elementBits[i].BitResLevelEl += deltaBits
			totalBits += deltaBits

			deltaBits = maxInt(totalBitreservoirMax-totalBitsMax, -elementBits[i].MaxBitResBitsEl)
			elementBits[i].MaxBitResBitsEl += deltaBits
			totalBitsMax += deltaBits
		}
	}
	return AACEncOK
}

func FDKaacEncPrepareBitDistribution(
	qcKernel *QCKernel,
	adjThrState *AdjThrState,
	psyOutElement []*PsyOutElement,
	qcOut []*QCOut,
	qcElement [][maxChannelElements]*QCOutElement,
	cm *ChannelMapping,
	elementBits []*ElementBits,
	avgTotalBits int,
) (QCPrepareBitDistributionResult, int) {
	checkPrepareBitDistributionInputs(qcKernel, adjThrState, psyOutElement, qcOut, qcElement, cm, elementBits, avgTotalBits)

	result := QCPrepareBitDistributionResult{}
	qcOut[0].GrantedDynBits = (minInt(qcKernel.MaxBitsPerFrame, avgTotalBits) - qcKernel.GlobHdrBits) &^ 7
	qcOut[0].GrantedDynBits -= qcOut[0].GlobalExtBits + qcOut[0].StaticBits + qcOut[0].ElementExtBits
	qcOut[0].MaxDynBits = (qcKernel.MaxBitsPerFrame &^ 7) - (qcOut[0].GlobalExtBits + qcOut[0].StaticBits + qcOut[0].ElementExtBits)

	if qcOut[0].GrantedDynBits+qcKernel.BitResTot < 0 {
		return result, AACEncBitresTooLow
	}
	if qcOut[0].GrantedDynBits < 0 {
		return result, AACEncBitresTooLow
	}

	FDKaacEncDistributeElementDynBits(qcElement[0][:], cm, elementBits, qcOut[0].GrantedDynBits)

	result.AvgTotalDynBits = 0
	result.TotalAvailableBits = avgTotalBits
	qcOut[0].TotalGrantedPeCorr = 0

	for i := 0; i < cm.NElements; i++ {
		elInfo := cm.ElInfo[i]
		if !fdkaacEncIsAdjustableElement(elInfo.ElType) {
			continue
		}
		nChannels := elInfo.NChannelsInEl
		psyChannels := psyOutElement[i].PsyOutChannel[:nChannels]
		grantedPe, grantedPeCorr := FDKaacEncDistributeBits(
			adjThrState,
			adjThrState.AdjThrStateElem[i],
			psyChannels,
			&qcElement[0][i].PEData,
			nChannels,
			psyOutElement[i].CommonWindow,
			qcElement[0][i].GrantedDynBits,
			elementBits[i].BitResLevelEl,
			elementBits[i].MaxBitResBitsEl,
			qcKernel.MaxBitFac,
			qcKernel.BitResMode,
		)
		qcElement[0][i].GrantedPe = grantedPe
		qcElement[0][i].GrantedPeCorr = grantedPeCorr
		result.TotalAvailableBits += elementBits[i].BitResLevelEl
		qcOut[0].TotalGrantedPeCorr += grantedPeCorr
		result.DistributedElements++
	}

	result.TotalAvailableBits = minInt(qcKernel.MaxBitsPerFrame, result.TotalAvailableBits)
	result.DistributedBits = qcOut[0].GrantedDynBits
	result.TotalGrantedPeCorr = qcOut[0].TotalGrantedPeCorr
	return result, AACEncOK
}

func FDKaacEncGetTotalConsumedBits(qcOut []*QCOut, qcElement [][maxChannelElements]*QCOutElement, cm *ChannelMapping, globHdrBits int, nSubFrames int) int {
	checkGetTotalConsumedBitsInputs(qcOut, qcElement, cm, globHdrBits, nSubFrames)

	totalUsedBits := 0
	for c := 0; c < nSubFrames; c++ {
		dataBits := 0
		for i := 0; i < cm.NElements; i++ {
			if fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
				dataBits += qcElement[c][i].DynBitsUsed + qcElement[c][i].StaticBitsUsed + qcElement[c][i].ExtBitsUsed
			}
		}
		dataBits += qcOut[c].GlobalExtBits

		totalUsedBits += (8 - dataBits%8) % 8
		totalUsedBits += dataBits + globHdrBits
	}
	return totalUsedBits
}

func FDKaacEncUpdateFillBits(qcKernel *QCKernel, qcOut []*QCOut) int {
	checkUpdateFillBitsInputs(qcKernel, qcOut)

	switch qcKernel.BitrateMode {
	case QCBitrateModeSFR:
	case QCBitrateModeFF:
	case QCBitrateModeVBR1, QCBitrateModeVBR2, QCBitrateModeVBR3, QCBitrateModeVBR4, QCBitrateModeVBR5:
		qcOut[0].TotFillBits = (qcOut[0].GrantedDynBits - qcOut[0].UsedDynBits) & 7
		qcOut[0].TotalBits = qcOut[0].StaticBits + qcOut[0].UsedDynBits + qcOut[0].TotFillBits + qcOut[0].ElementExtBits + qcOut[0].GlobalExtBits
		qcOut[0].TotFillBits += (maxInt(0, qcKernel.MinBitsPerFrame-qcOut[0].TotalBits) + 7) &^ 7
	case QCBitrateModeCBR, QCBitrateModeInvalid:
		fallthrough
	default:
		bitResSpace := qcKernel.BitResTotMax - qcKernel.BitResTot
		deltaBitRes := qcOut[0].GrantedDynBits - qcOut[0].UsedDynBits
		qcOut[0].TotFillBits = maxInt(
			deltaBitRes&7,
			deltaBitRes-(maxInt(0, bitResSpace-7)&^7),
		)
		qcOut[0].TotalBits = qcOut[0].StaticBits + qcOut[0].UsedDynBits + qcOut[0].TotFillBits + qcOut[0].ElementExtBits + qcOut[0].GlobalExtBits
		qcOut[0].TotFillBits += (maxInt(0, qcKernel.MinBitsPerFrame-qcOut[0].TotalBits) + 7) &^ 7
	}
	return AACEncOK
}

func FDKaacEncFinalizeBitConsumption(
	qcKernel *QCKernel,
	qcOut *QCOut,
	transportStaticBits int,
	aot int,
	syntaxFlags uint32,
	epConfig int8,
) int {
	checkFinalizeBitConsumptionInputs(qcKernel, qcOut, transportStaticBits)

	qcOut.TotalBits = qcOut.StaticBits + qcOut.UsedDynBits + qcOut.TotFillBits + qcOut.ElementExtBits + qcOut.GlobalExtBits

	if qcKernel.BitrateMode == QCBitrateModeCBR {
		exactTpBits := transportStaticBits
		if exactTpBits != qcKernel.GlobHdrBits {
			bitresSpace := qcKernel.BitResTotMax -
				(qcKernel.BitResTot + (qcOut.GrantedDynBits - (qcOut.UsedDynBits + qcOut.TotFillBits)))
			bitsToBitres := qcKernel.GlobHdrBits - exactTpBits
			if bitsToBitres < 0 {
				panic("fdkaac: transport static bits exceed estimated header bits")
			}
			diffFillBits := maxInt(0, bitsToBitres-bitresSpace)
			diffFillBits = (diffFillBits + 7) &^ 7

			qcKernel.BitResTot += bitsToBitres - diffFillBits
			qcOut.TotFillBits += diffFillBits
			qcOut.TotalBits += diffFillBits
			qcOut.GrantedDynBits += diffFillBits
			qcKernel.GlobHdrBits = transportStaticBits
			if qcKernel.GlobHdrBits != exactTpBits {
				qcKernel.BitResTot -= qcKernel.GlobHdrBits - exactTpBits
			}
		}
	}

	qcKernel.GlobHdrBits = transportStaticBits
	totFillBits := qcOut.TotFillBits
	fillExtPayload := QCOutExtension{Type: ExtFillData, PayloadBits: totFillBits}
	qcOut.TotFillBits = FDKaacEncWriteExtensionData(nil, &fillExtPayload, 0, 0, syntaxFlags, aot, epConfig)

	alignBits := 7 - (qcOut.StaticBits+qcOut.UsedDynBits+qcOut.ElementExtBits+qcOut.TotFillBits+qcOut.GlobalExtBits-1)%8
	if (alignBits+qcOut.TotFillBits-totFillBits) == 8 && qcOut.TotFillBits > 8 {
		qcOut.TotFillBits -= 8
	}

	qcOut.TotalBits = qcOut.StaticBits + qcOut.UsedDynBits + qcOut.TotFillBits + alignBits + qcOut.ElementExtBits + qcOut.GlobalExtBits
	if qcOut.TotalBits > qcKernel.MaxBitsPerFrame || qcOut.TotalBits < qcKernel.MinBitsPerFrame {
		return AACEncQuantError
	}
	qcOut.AlignBits = alignBits
	return AACEncOK
}

func FDKaacEncUpdateBitres(qcKernel *QCKernel, qcOut []*QCOut) {
	checkUpdateBitresInputs(qcKernel, qcOut)

	switch qcKernel.BitrateMode {
	case QCBitrateModeVBR1, QCBitrateModeVBR2, QCBitrateModeVBR3, QCBitrateModeVBR4, QCBitrateModeVBR5:
		qcKernel.BitResTot = minInt(qcKernel.MaxBitsPerFrame, qcKernel.BitResTotMax)
	case QCBitrateModeCBR, QCBitrateModeSFR, QCBitrateModeInvalid:
		fallthrough
	default:
		qcKernel.BitResTot += qcOut[0].GrantedDynBits - (qcOut[0].UsedDynBits + qcOut[0].TotFillBits + qcOut[0].AlignBits)
	}
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

func checkDistributeElementDynBitsInputs(qcElement []*QCOutElement, cm *ChannelMapping, elementBits []*ElementBits) {
	if cm == nil {
		panic("fdkaac: nil element-bit channel mapping")
	}
	if cm.NElements < 0 || cm.NElements > maxChannelElements {
		panic("fdkaac: invalid element-bit element count")
	}
	if len(qcElement) < cm.NElements || len(elementBits) < cm.NElements {
		panic("fdkaac: short element-bit inputs")
	}
	for i := 0; i < cm.NElements; i++ {
		if !fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
			continue
		}
		if qcElement[i] == nil {
			panic("fdkaac: nil element-bit output element")
		}
		if elementBits[i] == nil {
			panic("fdkaac: nil element-bit weights")
		}
		if elementBits[i].RelativeBitsEl < 0 {
			panic("fdkaac: invalid element-bit relative weight")
		}
	}
}

func checkBitResRedistributionInputs(qcKernel *QCKernel, cm *ChannelMapping, elementBits []*ElementBits) {
	if qcKernel == nil {
		panic("fdkaac: nil bit-reservoir redistribution kernel")
	}
	if qcKernel.MaxBitsPerFrame < 0 || qcKernel.BitResTotMax < 0 {
		panic("fdkaac: invalid bit-reservoir redistribution kernel")
	}
	if cm == nil {
		panic("fdkaac: nil bit-reservoir redistribution mapping")
	}
	if cm.NElements < 0 || cm.NElements > maxChannelElements || len(elementBits) < cm.NElements {
		panic("fdkaac: invalid bit-reservoir redistribution element count")
	}
	for i := 0; i < cm.NElements; i++ {
		if !fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
			continue
		}
		if elementBits[i] == nil {
			panic("fdkaac: nil bit-reservoir redistribution element")
		}
		if elementBits[i].RelativeBitsEl < 0 {
			panic("fdkaac: invalid bit-reservoir redistribution relative weight")
		}
	}
}

func checkPrepareBitDistributionInputs(
	qcKernel *QCKernel,
	adjThrState *AdjThrState,
	psyOutElement []*PsyOutElement,
	qcOut []*QCOut,
	qcElement [][maxChannelElements]*QCOutElement,
	cm *ChannelMapping,
	elementBits []*ElementBits,
	avgTotalBits int,
) {
	if qcKernel == nil {
		panic("fdkaac: nil bit-distribution kernel")
	}
	if adjThrState == nil {
		panic("fdkaac: nil bit-distribution threshold state")
	}
	if cm == nil {
		panic("fdkaac: nil bit-distribution channel mapping")
	}
	if avgTotalBits < 0 || qcKernel.MaxBitsPerFrame < 0 || qcKernel.GlobHdrBits < 0 || qcKernel.MaxBitFac < 0 {
		panic("fdkaac: invalid bit-distribution kernel levels")
	}
	if qcKernel.BitResMode != BitresModeFull && qcKernel.BitResMode != BitresModeReduced && qcKernel.BitResMode != BitresModeDisabled {
		panic("fdkaac: invalid bit-distribution reservoir mode")
	}
	if cm.NElements < 0 || cm.NElements > maxChannelElements ||
		len(psyOutElement) < cm.NElements || len(qcOut) < 1 || len(qcElement) < 1 ||
		len(elementBits) < cm.NElements {
		panic("fdkaac: invalid bit-distribution element count")
	}
	if qcOut[0] == nil {
		panic("fdkaac: nil bit-distribution frame")
	}
	if qcOut[0].GlobalExtBits < 0 || qcOut[0].StaticBits < 0 || qcOut[0].ElementExtBits < 0 {
		panic("fdkaac: invalid bit-distribution frame sizes")
	}
	for i := 0; i < cm.NElements; i++ {
		if !fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
			continue
		}
		if cm.ElInfo[i].NChannelsInEl != channelElementCount(cm.ElInfo[i].ElType) {
			panic("fdkaac: invalid bit-distribution channel count")
		}
		if psyOutElement[i] == nil {
			panic("fdkaac: nil bit-distribution psy element")
		}
		if qcElement[0][i] == nil {
			panic("fdkaac: nil bit-distribution QC element")
		}
		if elementBits[i] == nil {
			panic("fdkaac: nil bit-distribution element bits")
		}
		if adjThrState.AdjThrStateElem[i] == nil {
			panic("fdkaac: nil bit-distribution threshold element")
		}
		if elementBits[i].RelativeBitsEl < 0 || elementBits[i].BitResLevelEl < 0 || elementBits[i].MaxBitResBitsEl < 0 {
			panic("fdkaac: invalid bit-distribution element bits")
		}
		if qcKernel.BitResMode == BitresModeFull && elementBits[i].MaxBitResBitsEl <= 0 {
			panic("fdkaac: invalid bit-distribution reservoir maximum")
		}
		for ch := 0; ch < cm.ElInfo[i].NChannelsInEl; ch++ {
			if psyOutElement[i].PsyOutChannel[ch] == nil {
				panic("fdkaac: nil bit-distribution psy channel")
			}
		}
		if qcElement[0][i].PEData.Pe < 0 {
			panic("fdkaac: invalid bit-distribution PE data")
		}
	}
}

func checkGetTotalConsumedBitsInputs(qcOut []*QCOut, qcElement [][maxChannelElements]*QCOutElement, cm *ChannelMapping, globHdrBits int, nSubFrames int) {
	if cm == nil {
		panic("fdkaac: nil consumed-bit channel mapping")
	}
	if globHdrBits < 0 {
		panic("fdkaac: invalid consumed-bit header size")
	}
	if nSubFrames < 0 || len(qcOut) < nSubFrames || len(qcElement) < nSubFrames {
		panic("fdkaac: invalid consumed-bit frame count")
	}
	if cm.NElements < 0 || cm.NElements > maxChannelElements {
		panic("fdkaac: invalid consumed-bit element count")
	}
	for c := 0; c < nSubFrames; c++ {
		if qcOut[c] == nil {
			panic("fdkaac: nil consumed-bit frame")
		}
		if qcOut[c].GlobalExtBits < 0 {
			panic("fdkaac: invalid consumed-bit global extension size")
		}
		for i := 0; i < cm.NElements; i++ {
			if !fdkaacEncIsAdjustableElement(cm.ElInfo[i].ElType) {
				continue
			}
			if qcElement[c][i] == nil {
				panic("fdkaac: nil consumed-bit element")
			}
			if qcElement[c][i].DynBitsUsed < 0 || qcElement[c][i].StaticBitsUsed < 0 || qcElement[c][i].ExtBitsUsed < 0 {
				panic("fdkaac: invalid consumed-bit element size")
			}
		}
	}
}

func checkUpdateFillBitsInputs(qcKernel *QCKernel, qcOut []*QCOut) {
	if qcKernel == nil {
		panic("fdkaac: nil fill-bit QC kernel")
	}
	if len(qcOut) < 1 || qcOut[0] == nil {
		panic("fdkaac: nil fill-bit frame")
	}
	if qcKernel.MinBitsPerFrame < 0 || qcKernel.BitResTotMax < 0 {
		panic("fdkaac: invalid fill-bit kernel")
	}
	if qcOut[0].StaticBits < 0 || qcOut[0].UsedDynBits < 0 ||
		qcOut[0].ElementExtBits < 0 || qcOut[0].GlobalExtBits < 0 {
		panic("fdkaac: invalid fill-bit frame sizes")
	}
}

func checkFinalizeBitConsumptionInputs(qcKernel *QCKernel, qcOut *QCOut, transportStaticBits int) {
	if qcKernel == nil {
		panic("fdkaac: nil finalize QC kernel")
	}
	if qcOut == nil {
		panic("fdkaac: nil finalize frame")
	}
	if transportStaticBits < 0 {
		panic("fdkaac: invalid finalize transport header size")
	}
	if qcKernel.MaxBitsPerFrame < qcKernel.MinBitsPerFrame || qcKernel.MinBitsPerFrame < 0 {
		panic("fdkaac: invalid finalize frame bounds")
	}
	if qcOut.StaticBits < 0 || qcOut.UsedDynBits < 0 || qcOut.TotFillBits < 0 ||
		qcOut.ElementExtBits < 0 || qcOut.GlobalExtBits < 0 {
		panic("fdkaac: invalid finalize frame sizes")
	}
}

func checkUpdateBitresInputs(qcKernel *QCKernel, qcOut []*QCOut) {
	if qcKernel == nil {
		panic("fdkaac: nil bit-reservoir QC kernel")
	}
	if len(qcOut) < 1 || qcOut[0] == nil {
		panic("fdkaac: nil bit-reservoir frame")
	}
	if qcKernel.MaxBitsPerFrame < 0 || qcKernel.BitResTotMax < 0 {
		panic("fdkaac: invalid bit-reservoir kernel")
	}
	if qcOut[0].UsedDynBits < 0 || qcOut[0].TotFillBits < 0 || qcOut[0].AlignBits < 0 {
		panic("fdkaac: invalid bit-reservoir frame sizes")
	}
}
