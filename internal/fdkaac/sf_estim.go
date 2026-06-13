package fdkaac

const formFacShift = 6
const asPeFacShift = 7
const asPeFacLdData = FixpDBL(0x0e000000)
const peC1 = FixpDBL(0x03000000)
const peC2 = FixpDBL(0x015269e2)
const peC3 = FixpDBL(0x47990500)
const peFac07 = FixpDBL(0x59999980)
const peFac0375 = FixpDBL(0x30000000)
const peSfbConstPart6p75 = FixpDBL(0x02c14050)
const formFactorLdScale = FixpDBL(0x0c000000)
const fdkIntMin = -1 << 31
const fdkIntMax = 1<<31 - 1
const upcountLimit = 1
const maxScfDelta = 60
const distFacShift = 3
const scfDeltaPeLimit = FixpDBL(0x00140000) // FDK FL2FXCONST_DBL(10.0f / (1 << (2 * AS_PE_FAC_SHIFT))).
const distFactorLdData = FixpDBL(-10802114) // FDK FL2FXCONST_DBL(-0.0050301265), ld64(1/1.25).
const scfDistMaxFac = FixpDBL(0x2aaaaaab)
const scfLd1e3 = FixpDBL(0x13ee7b47)
const scfFinerDistRatio = FixpDBL(0x66666666)
const scfDistIncreaseLimit = FixpDBL(0x0055077a)
const scfEnergyDecreaseLimit = FixpDBL(0x00ff2bfe)
const scfThresholdConstLdData = FixpDBL(0x0582809d)
const scfConvConst = FixpDBL(0x268826a1)
const scf08585 = FixpDBL(0x6de353f8)
const scfC1Const = FixpDBL(-581606939)
const maxSpectralLines = 1024

type QCOutChannel struct {
	MdctSpectrum        [maxSpectralLines]FixpDBL
	QuantSpec           [maxSpectralLines]int16
	MaxValueInSfb       [maxGroupedSFB]uint32
	Scf                 [maxGroupedSFB]int
	GlobalGain          int
	SectionData         SectionData
	SfbFormFactorLdData [maxGroupedSFB]FixpDBL
	SfbEnergyLdData     [maxGroupedSFB]FixpDBL
	SfbThresholdLdData  [maxGroupedSFB]FixpDBL
}

func FDKaacEncCalcFormFactor(qcOutChannel []*QCOutChannel, psyOutChannel []*PsyOutChannel, nChannels int) {
	if nChannels < 0 || len(qcOutChannel) < nChannels || len(psyOutChannel) < nChannels {
		panic("fdkaac: invalid form-factor channel count")
	}
	for j := 0; j < nChannels; j++ {
		if qcOutChannel[j] == nil {
			panic("fdkaac: nil form-factor qc output")
		}
		FDKaacEncCalcFormFactorChannel(qcOutChannel[j].SfbFormFactorLdData[:], psyOutChannel[j])
	}
}

func FDKaacEncCalcFormFactorChannel(sfbFormFactorLdData []FixpDBL, psyOutChan *PsyOutChannel) {
	checkFormFactorInputs(sfbFormFactorLdData, psyOutChan)

	tmp0 := psyOutChan.SfbCnt
	tmp1 := psyOutChan.MaxSfbPerGroup
	step := psyOutChan.SfbPerGroup
	for sfbGrp := 0; sfbGrp < tmp0; sfbGrp += step {
		sfb := 0
		for ; sfb < tmp1; sfb++ {
			formFactor := FixpDBL(0)
			for j := psyOutChan.SfbOffsets[sfbGrp+sfb]; j < psyOutChan.SfbOffsets[sfbGrp+sfb+1]; j++ {
				formFactor += sqrtFixp(fixpAbsDBL(psyOutChan.MdctSpectrum[j])) >> formFacShift
			}
			sfbFormFactorLdData[sfbGrp+sfb] = CalcLdData(formFactor)
		}
		for ; sfb < psyOutChan.SfbPerGroup; sfb++ {
			sfbFormFactorLdData[sfbGrp+sfb] = MinValDBL
		}
	}
}

func FDKaacEncCalcSfbRelevantLines(
	sfbFormFactorLdData []FixpDBL,
	sfbEnergyLdData []FixpDBL,
	sfbThresholdLdData []FixpDBL,
	sfbOffsets []int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	sfbNRelevantLines []FixpDBL,
) {
	checkSfbRelevantLinesInputs(
		sfbFormFactorLdData, sfbEnergyLdData, sfbThresholdLdData, sfbOffsets,
		sfbCnt, sfbPerGroup, maxSfbPerGroup, sfbNRelevantLines,
	)

	for i := 0; i < sfbCnt; i++ {
		sfbNRelevantLines[i] = 0
	}

	for sfbOffs := 0; sfbOffs < sfbCnt; sfbOffs += sfbPerGroup {
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			idx := sfbOffs + sfb
			if sfbEnergyLdData[idx] > sfbThresholdLdData[idx] {
				sfbWidth := sfbOffsets[idx+1] - sfbOffsets[idx]
				sfbWidthLdData := FixpDBL(sfbWidth << (DfractBits - 1 - asPeFacShift))
				sfbWidthLdData = CalcLdData(sfbWidthLdData)

				accu := sfbEnergyLdData[idx] - sfbWidthLdData - asPeFacLdData
				accu = sfbFormFactorLdData[idx] - (accu >> 2)

				sfbNRelevantLines[idx] = CalcInvLdData(accu) >> 1
			}
		}
	}
}

func FDKaacEncCountSingleScfBits(scf int, scfLeft int, scfRight int) FixpDBL {
	scfBitsFract := FixpDBL(FDKaacEncBitCountScalefactorDelta(scfLeft-scf) + FDKaacEncBitCountScalefactorDelta(scf-scfRight))
	return scfBitsFract << (DfractBits - 1 - (2 * asPeFacShift))
}

func FDKaacEncCalcSingleSpecPe(scf int, sfbConstPePart FixpDBL, nLines FixpDBL) FixpDBL {
	scfFract := FixpDBL(scf << (DfractBits - 1 - asPeFacShift))
	ldRatio := sfbConstPePart - FMultDD(peFac0375, scfFract)

	if ldRatio >= peC1 {
		return FMultDD(peFac07, FMultDD(nLines, ldRatio))
	}
	return FMultDD(peFac07, FMultDD(nLines, peC2+FMultDD(peC3, ldRatio)))
}

func FDKaacEncCountScfBitsDiff(scfOld []int, scfNew []int, sfbCnt int, startSfb int, stopSfb int) FixpDBL {
	checkScfDiffInputs(scfOld, scfNew, sfbCnt, startSfb, stopSfb)

	scfBitsDiff := 0
	sfbLast := startSfb
	for sfbLast < stopSfb && scfOld[sfbLast] == fdkIntMin {
		sfbLast++
	}
	if sfbLast == stopSfb {
		panic("fdkaac: empty scalefactor diff range")
	}

	sfbPrev := startSfb - 1
	for sfbPrev >= 0 && scfOld[sfbPrev] == fdkIntMin {
		sfbPrev--
	}
	if sfbPrev >= 0 {
		scfBitsDiff += FDKaacEncBitCountScalefactorDelta(scfNew[sfbPrev]-scfNew[sfbLast]) -
			FDKaacEncBitCountScalefactorDelta(scfOld[sfbPrev]-scfOld[sfbLast])
	}

	for sfb := sfbLast + 1; sfb < stopSfb; sfb++ {
		if scfOld[sfb] != fdkIntMin {
			scfBitsDiff += FDKaacEncBitCountScalefactorDelta(scfNew[sfbLast]-scfNew[sfb]) -
				FDKaacEncBitCountScalefactorDelta(scfOld[sfbLast]-scfOld[sfb])
			sfbLast = sfb
		}
	}

	sfbNext := stopSfb
	for sfbNext < sfbCnt && scfOld[sfbNext] == fdkIntMin {
		sfbNext++
	}
	if sfbNext < sfbCnt {
		scfBitsDiff += FDKaacEncBitCountScalefactorDelta(scfNew[sfbLast]-scfNew[sfbNext]) -
			FDKaacEncBitCountScalefactorDelta(scfOld[sfbLast]-scfOld[sfbNext])
	}

	return FixpDBL(scfBitsDiff << (DfractBits - 1 - (2 * asPeFacShift)))
}

func FDKaacEncCalcSpecPeDiff(
	sfbEnergyLdData []FixpDBL,
	scfOld []int,
	scfNew []int,
	sfbConstPePart []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbNRelevantLines []FixpDBL,
	startSfb int,
	stopSfb int,
) FixpDBL {
	checkSpecPeDiffInputs(
		sfbEnergyLdData, scfOld, scfNew, sfbConstPePart, sfbFormFactorLdData,
		sfbNRelevantLines, startSfb, stopSfb,
	)

	specPeDiff := FixpDBL(0)
	for sfb := startSfb; sfb < stopSfb; sfb++ {
		if scfOld[sfb] != fdkIntMin {
			if sfbConstPePart[sfb] == FixpDBL(fdkIntMin) {
				sfbConstPePart[sfb] = ((sfbEnergyLdData[sfb] - sfbFormFactorLdData[sfb] - formFactorLdScale) >> 1) + peSfbConstPart6p75
			}

			scfFract := FixpDBL(scfOld[sfb] << (DfractBits - 1 - asPeFacShift))
			ldRatioOld := sfbConstPePart[sfb] - FMultDD(peFac0375, scfFract)

			scfFract = FixpDBL(scfNew[sfb] << (DfractBits - 1 - asPeFacShift))
			ldRatioNew := sfbConstPePart[sfb] - FMultDD(peFac0375, scfFract)

			var pOld FixpDBL
			if ldRatioOld >= peC1 {
				pOld = ldRatioOld
			} else {
				pOld = peC2 + FMultDD(peC3, ldRatioOld)
			}

			var pNew FixpDBL
			if ldRatioNew >= peC1 {
				pNew = ldRatioNew
			} else {
				pNew = peC2 + FMultDD(peC3, ldRatioNew)
			}

			specPeDiff += FMultDD(peFac07, FMultDD(sfbNRelevantLines[sfb], pNew-pOld))
		}
	}
	return specPeDiff
}

func FDKaacEncImproveScf(
	spec []FixpDBL,
	quantSpec []int16,
	quantSpecTmp []int16,
	sfbWidth int,
	threshLdData FixpDBL,
	scf int,
	minScf int,
	distLdData *FixpDBL,
	minScfCalculated *int,
	dZoneQuantEnable int,
) int {
	checkImproveScfInputs(spec, quantSpec, quantSpecTmp, sfbWidth, distLdData, minScfCalculated)

	scfBest := scf
	sfbDistLdData := FDKaacEncCalcSfbDist(spec, quantSpec, sfbWidth, scf, dZoneQuantEnable)
	*minScfCalculated = scf

	if sfbDistLdData > threshLdData-distFactorLdData {
		scfEstimated := scf
		sfbDistBestLdData := sfbDistLdData
		cnt := 0
		for sfbDistLdData > threshLdData-distFactorLdData && cnt < upcountLimit {
			cnt++
			scf++
			sfbDistLdData = FDKaacEncCalcSfbDist(spec, quantSpecTmp, sfbWidth, scf, dZoneQuantEnable)

			if sfbDistLdData < sfbDistBestLdData {
				scfBest = scf
				sfbDistBestLdData = sfbDistLdData
				copy(quantSpec[:sfbWidth], quantSpecTmp[:sfbWidth])
			}
		}

		cnt = 0
		scf = scfEstimated
		sfbDistLdData = sfbDistBestLdData
		for sfbDistLdData > threshLdData-distFactorLdData && cnt < 1 && scf > minScf {
			cnt++
			scf--
			sfbDistLdData = FDKaacEncCalcSfbDist(spec, quantSpecTmp, sfbWidth, scf, dZoneQuantEnable)

			if sfbDistLdData < sfbDistBestLdData {
				scfBest = scf
				sfbDistBestLdData = sfbDistLdData
				copy(quantSpec[:sfbWidth], quantSpecTmp[:sfbWidth])
			}
			*minScfCalculated = scf
		}
		*distLdData = sfbDistBestLdData
	} else {
		sfbDistBestLdData := sfbDistLdData
		sfbDistAllowedLdData := minFixpDBL(sfbDistLdData-distFactorLdData, threshLdData)
		for cnt := 0; cnt < upcountLimit; cnt++ {
			scf++
			sfbDistLdData = FDKaacEncCalcSfbDist(spec, quantSpecTmp, sfbWidth, scf, dZoneQuantEnable)

			if sfbDistLdData < sfbDistAllowedLdData {
				*minScfCalculated = scfBest + 1
				scfBest = scf
				sfbDistBestLdData = sfbDistLdData
				copy(quantSpec[:sfbWidth], quantSpecTmp[:sfbWidth])
			}
		}
		*distLdData = sfbDistBestLdData
	}

	return scfBest
}

func FDKaacEncEstimateScaleFactorsChannel(
	qcOutChannel *QCOutChannel,
	psyOutChannel *PsyOutChannel,
	scf []int,
	globalGain *int,
	sfbFormFactorLdData []FixpDBL,
	invQuant int,
	quantSpec []int16,
	dZoneQuantEnable int,
	quantSpecTmp []int16,
) {
	finalOffset := checkEstimateScaleFactorsChannelInputs(
		qcOutChannel, psyOutChannel, scf, globalGain, sfbFormFactorLdData,
		invQuant, quantSpec, quantSpecTmp,
	)

	var minScfCalculated [maxGroupedSFB]int
	var sfbDistLdData [maxGroupedSFB]FixpDBL
	var minSfMaxQuant [maxGroupedSFB]int

	if invQuant > 0 {
		for i := 0; i < finalOffset; i++ {
			quantSpec[i] = 0
		}
	}

	for i := 0; i < psyOutChannel.SfbCnt; i++ {
		scf[i] = fdkIntMin
	}
	for i := 0; i < maxGroupedSFB; i++ {
		minSfMaxQuant[i] = fdkIntMin
	}

	for sfbOffs := 0; sfbOffs < psyOutChannel.SfbCnt; sfbOffs += psyOutChannel.SfbPerGroup {
		for sfb := 0; sfb < psyOutChannel.MaxSfbPerGroup; sfb++ {
			idx := sfbOffs + sfb
			threshLdData := qcOutChannel.SfbThresholdLdData[idx]
			energyLdData := qcOutChannel.SfbEnergyLdData[idx]

			sfbDistLdData[idx] = energyLdData
			if energyLdData <= threshLdData {
				continue
			}

			energyPartLdData := sfbFormFactorLdData[idx] + formFactorLdScale
			thresholdPartLdData := scfThresholdConstLdData + threshLdData
			scfFract := thresholdPartLdData - energyPartLdData
			scfFract = FMultDD(scfConvConst, scfFract)
			scfFract = scfFract + FMultDD(scf08585, scfFract>>3)
			scfInt := int(scfFract >> (DfractBits - 1 - 3 - ldDataShift))

			maxSpec := FixpDBL(0)
			start := psyOutChannel.SfbOffsets[idx]
			stop := psyOutChannel.SfbOffsets[idx+1]
			for j := start; j < stop; j++ {
				maxSpec = maxFixpDBL(maxSpec, fixpAbsDBL(qcOutChannel.MdctSpectrum[j]))
			}

			tmp := CalcLdData(maxSpec)
			if scfC1Const > MinValDBL-tmp {
				minSfMaxQuant[idx] = int((scfC1Const+tmp)>>(DfractBits-1-8)) + 1
			} else {
				minSfMaxQuant[idx] = int(MinValDBL>>(DfractBits-1-8)) + 1
			}

			scfInt = maxInt(scfInt, minSfMaxQuant[idx])
			if invQuant > 0 {
				scfInt = FDKaacEncImproveScf(
					qcOutChannel.MdctSpectrum[start:stop],
					quantSpec[start:stop],
					quantSpecTmp[start:stop],
					stop-start,
					threshLdData,
					scfInt,
					minSfMaxQuant[idx],
					&sfbDistLdData[idx],
					&minScfCalculated[idx],
					dZoneQuantEnable,
				)
			}
			scf[idx] = scfInt
		}
	}

	if invQuant > 0 {
		var sfbConstPePart [maxGroupedSFB]FixpDBL
		var sfbNRelevantLines [maxGroupedSFB]FixpDBL

		for i := 0; i < psyOutChannel.SfbCnt; i++ {
			sfbConstPePart[i] = FixpDBL(fdkIntMin)
		}

		FDKaacEncCalcSfbRelevantLines(
			sfbFormFactorLdData,
			qcOutChannel.SfbEnergyLdData[:],
			qcOutChannel.SfbThresholdLdData[:],
			psyOutChannel.SfbOffsets[:],
			psyOutChannel.SfbCnt,
			psyOutChannel.SfbPerGroup,
			psyOutChannel.MaxSfbPerGroup,
			sfbNRelevantLines[:],
		)

		FDKaacEncAssimilateSingleScf(
			psyOutChannel,
			qcOutChannel,
			quantSpec,
			quantSpecTmp,
			dZoneQuantEnable,
			scf,
			minSfMaxQuant[:],
			sfbDistLdData[:],
			sfbConstPePart[:],
			sfbFormFactorLdData,
			sfbNRelevantLines[:],
			minScfCalculated[:],
			1,
		)

		if invQuant > 1 {
			FDKaacEncAssimilateMultipleScf(
				psyOutChannel,
				qcOutChannel,
				quantSpec,
				quantSpecTmp,
				dZoneQuantEnable,
				scf,
				minSfMaxQuant[:],
				sfbDistLdData[:],
				sfbConstPePart[:],
				sfbFormFactorLdData,
				sfbNRelevantLines[:],
			)

			FDKaacEncAssimilateMultipleScf2(
				psyOutChannel,
				qcOutChannel,
				quantSpec,
				quantSpecTmp,
				dZoneQuantEnable,
				scf,
				minSfMaxQuant[:],
				sfbDistLdData[:],
				sfbConstPePart[:],
				sfbFormFactorLdData,
				sfbNRelevantLines[:],
			)
		}
	}

	minSf := fdkIntMax
	for sfbOffs := 0; sfbOffs < psyOutChannel.SfbCnt; sfbOffs += psyOutChannel.SfbPerGroup {
		for sfb := 0; sfb < psyOutChannel.MaxSfbPerGroup; sfb++ {
			idx := sfbOffs + sfb
			if scf[idx] != fdkIntMin {
				minSf = minInt(minSf, scf[idx])
			}
		}
	}

	for sfbOffs := 0; sfbOffs < psyOutChannel.SfbCnt; sfbOffs += psyOutChannel.SfbPerGroup {
		for sfb := 0; sfb < psyOutChannel.MaxSfbPerGroup; sfb++ {
			idx := sfbOffs + sfb
			if scf[idx] != fdkIntMin && minSf+maxScfDelta < scf[idx] {
				scf[idx] = minSf + maxScfDelta
				if invQuant > 0 {
					start := psyOutChannel.SfbOffsets[idx]
					stop := psyOutChannel.SfbOffsets[idx+1]
					sfbDistLdData[idx] = FDKaacEncCalcSfbDist(
						qcOutChannel.MdctSpectrum[start:stop],
						quantSpec[start:stop],
						stop-start,
						scf[idx],
						dZoneQuantEnable,
					)
				}
			}
		}
	}

	maxSf := fdkIntMin
	for sfbOffs := 0; sfbOffs < psyOutChannel.SfbCnt; sfbOffs += psyOutChannel.SfbPerGroup {
		for sfb := 0; sfb < psyOutChannel.MaxSfbPerGroup; sfb++ {
			maxSf = maxInt(maxSf, scf[sfbOffs+sfb])
		}
	}

	if maxSf > fdkIntMin {
		*globalGain = maxSf
		for sfbOffs := 0; sfbOffs < psyOutChannel.SfbCnt; sfbOffs += psyOutChannel.SfbPerGroup {
			for sfb := 0; sfb < psyOutChannel.MaxSfbPerGroup; sfb++ {
				idx := sfbOffs + sfb
				if scf[idx] == fdkIntMin {
					scf[idx] = 0
					for j := psyOutChannel.SfbOffsets[idx]; j < psyOutChannel.SfbOffsets[idx+1]; j++ {
						qcOutChannel.MdctSpectrum[j] = 0
					}
				} else {
					scf[idx] = maxSf - scf[idx]
				}
			}
		}
		return
	}

	*globalGain = 0
	for sfbOffs := 0; sfbOffs < psyOutChannel.SfbCnt; sfbOffs += psyOutChannel.SfbPerGroup {
		for sfb := 0; sfb < psyOutChannel.MaxSfbPerGroup; sfb++ {
			idx := sfbOffs + sfb
			scf[idx] = 0
			for j := psyOutChannel.SfbOffsets[idx]; j < psyOutChannel.SfbOffsets[idx+1]; j++ {
				qcOutChannel.MdctSpectrum[j] = 0
			}
		}
	}
}

func FDKaacEncEstimateScaleFactors(
	psyOutChannel []*PsyOutChannel,
	qcOutChannel []*QCOutChannel,
	invQuant int,
	dZoneQuantEnable int,
	nChannels int,
	quantSpecTmp []int16,
) {
	checkEstimateScaleFactorsInputs(psyOutChannel, qcOutChannel, invQuant, nChannels, quantSpecTmp)

	for ch := 0; ch < nChannels; ch++ {
		if qcOutChannel[ch] == nil {
			panic("fdkaac: nil estimate-scalefactor qc output")
		}
		if psyOutChannel[ch] == nil {
			panic("fdkaac: nil estimate-scalefactor psy output")
		}
		FDKaacEncEstimateScaleFactorsChannel(
			qcOutChannel[ch],
			psyOutChannel[ch],
			qcOutChannel[ch].Scf[:],
			&qcOutChannel[ch].GlobalGain,
			qcOutChannel[ch].SfbFormFactorLdData[:],
			invQuant,
			qcOutChannel[ch].QuantSpec[:],
			dZoneQuantEnable,
			quantSpecTmp,
		)
	}
}

func FDKaacEncAssimilateSingleScf(
	psyOutChan *PsyOutChannel,
	qcOutChannel *QCOutChannel,
	quantSpec []int16,
	quantSpecTmp []int16,
	dZoneQuantEnable int,
	scf []int,
	minScf []int,
	sfbDist []FixpDBL,
	sfbConstPePart []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbNRelevantLines []FixpDBL,
	minScfCalculated []int,
	restartOnSuccess int,
) {
	checkAssimilateSingleScfInputs(
		psyOutChan, qcOutChannel, quantSpec, quantSpecTmp, scf, minScf, sfbDist,
		sfbConstPePart, sfbFormFactorLdData, sfbNRelevantLines, minScfCalculated,
	)

	var prevScfLast [maxGroupedSFB]int
	var prevScfNext [maxGroupedSFB]int
	var deltaPeLast [maxGroupedSFB]FixpDBL
	for i := 0; i < psyOutChan.SfbCnt; i++ {
		prevScfLast[i] = fdkIntMax
		prevScfNext[i] = fdkIntMax
		deltaPeLast[i] = FixpDBL(fdkIntMax)
	}

	sfbLast := -1
	sfbAct := -1
	sfbNext := -1
	scfMin := fdkIntMax
	scfMax := fdkIntMax
	success := false
	deltaPe := FixpDBL(0)

	for {
		sfbNext++
		for sfbNext < psyOutChan.SfbCnt && scf[sfbNext] == fdkIntMin {
			sfbNext++
		}

		scfAct := 0
		scfLast := 0
		scfNextVal := 0
		if sfbLast >= 0 && sfbAct >= 0 && sfbNext < psyOutChan.SfbCnt {
			scfAct = scf[sfbAct]
			scfLast = scf[sfbLast]
			scfNextVal = scf[sfbNext]
			scfMin = minInt(scfLast, scfNextVal)
			scfMax = maxInt(scfLast, scfNextVal)
		} else if sfbLast == -1 && sfbAct >= 0 && sfbNext < psyOutChan.SfbCnt {
			scfAct = scf[sfbAct]
			scfLast = scfAct
			scfNextVal = scf[sfbNext]
			scfMin = scfNextVal
			scfMax = scfNextVal
		} else if sfbLast >= 0 && sfbAct >= 0 && sfbNext == psyOutChan.SfbCnt {
			scfAct = scf[sfbAct]
			scfLast = scf[sfbLast]
			scfNextVal = scfAct
			scfMin = scfLast
			scfMax = scfLast
		}
		if sfbAct >= 0 {
			scfMin = maxInt(scfMin, minScf[sfbAct])
		}

		if sfbAct >= 0 &&
			(sfbLast >= 0 || sfbNext < psyOutChan.SfbCnt) &&
			scfAct > scfMin &&
			scfAct <= scfMin+maxScfDelta &&
			scfAct >= scfMax-maxScfDelta &&
			scfAct <= minInt(scfMin, minInt(scfLast, scfNextVal))+maxScfDelta &&
			(scfLast != prevScfLast[sfbAct] || scfNextVal != prevScfNext[sfbAct] || deltaPe < deltaPeLast[sfbAct]) {
			success = false

			sfbWidth := psyOutChan.SfbOffsets[sfbAct+1] - psyOutChan.SfbOffsets[sfbAct]
			sfbOffs := psyOutChan.SfbOffsets[sfbAct]
			enLdData := qcOutChannel.SfbEnergyLdData[sfbAct]

			if sfbConstPePart[sfbAct] == FixpDBL(fdkIntMin) {
				sfbConstPePart[sfbAct] = ((enLdData - sfbFormFactorLdData[sfbAct] - formFactorLdScale) >> 1) + peSfbConstPart6p75
			}

			sfbPeOld := FDKaacEncCalcSingleSpecPe(scfAct, sfbConstPePart[sfbAct], sfbNRelevantLines[sfbAct]) +
				FDKaacEncCountSingleScfBits(scfAct, scfLast, scfNextVal)

			deltaPeNew := deltaPe
			updateMinScfCalculated := true
			for {
				scfAct--
				if scfAct < minScfCalculated[sfbAct] && scfAct >= scfMax-maxScfDelta {
					sfbPeNew := FDKaacEncCalcSingleSpecPe(scfAct, sfbConstPePart[sfbAct], sfbNRelevantLines[sfbAct]) +
						FDKaacEncCountSingleScfBits(scfAct, scfLast, scfNextVal)

					deltaPeTmp := deltaPe + sfbPeNew - sfbPeOld
					if deltaPeTmp < scfDeltaPeLimit {
						sfbDistNew := FDKaacEncCalcSfbDist(
							qcOutChannel.MdctSpectrum[sfbOffs:],
							quantSpecTmp[sfbOffs:],
							sfbWidth,
							scfAct,
							dZoneQuantEnable,
						)

						if sfbDistNew < sfbDist[sfbAct] {
							scf[sfbAct] = scfAct
							sfbDist[sfbAct] = sfbDistNew
							copy(quantSpec[sfbOffs:sfbOffs+sfbWidth], quantSpecTmp[sfbOffs:sfbOffs+sfbWidth])

							deltaPeNew = deltaPeTmp
							success = true
						}
						if updateMinScfCalculated {
							minScfCalculated[sfbAct] = scfAct
						}
					} else {
						updateMinScfCalculated = false
					}
				}
				if scfAct <= scfMin {
					break
				}
			}

			deltaPe = deltaPeNew
			prevScfLast[sfbAct] = scfLast
			prevScfNext[sfbAct] = scfNextVal
			deltaPeLast[sfbAct] = deltaPe
		}

		if success && restartOnSuccess != 0 {
			sfbLast = -1
			sfbAct = -1
			sfbNext = -1
			scfMin = fdkIntMax
			scfMax = fdkIntMax
			success = false
		} else {
			sfbLast = sfbAct
			sfbAct = sfbNext
		}

		if sfbNext >= psyOutChan.SfbCnt {
			break
		}
	}
}

func FDKaacEncAssimilateMultipleScf(
	psyOutChan *PsyOutChannel,
	qcOutChannel *QCOutChannel,
	quantSpec []int16,
	quantSpecTmp []int16,
	dZoneQuantEnable int,
	scf []int,
	minScf []int,
	sfbDist []FixpDBL,
	sfbConstPePart []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbNRelevantLines []FixpDBL,
) {
	checkAssimilateMultipleScfInputs(
		psyOutChan, qcOutChannel, quantSpec, quantSpecTmp, scf, minScf, sfbDist,
		sfbConstPePart, sfbFormFactorLdData, sfbNRelevantLines,
	)

	sfbCnt := psyOutChan.SfbCnt
	scfMin := fdkIntMax
	scfMax := fdkIntMin
	for sfb := 0; sfb < sfbCnt; sfb++ {
		if scf[sfb] != fdkIntMin {
			scfMin = minInt(scfMin, scf[sfb])
			scfMax = maxInt(scfMax, scf[sfb])
		}
	}

	if scfMax == fdkIntMin || scfMax > scfMin+maxScfDelta {
		return
	}

	var scfTmp [maxGroupedSFB]int
	var sfbDistNew [maxGroupedSFB]FixpDBL
	deltaPe := FixpDBL(0)
	scfAct := scfMax
	for {
		scfAct--
		copy(scfTmp[:], scf[:])
		stopSfb := 0
		for {
			sfb := stopSfb
			for sfb < sfbCnt && (scf[sfb] == fdkIntMin || scf[sfb] <= scfAct) {
				sfb++
			}
			startSfb := sfb
			sfb++
			for sfb < sfbCnt && (scf[sfb] == fdkIntMin || scf[sfb] > scfAct) {
				sfb++
			}
			stopSfb = sfb

			possibleRegionFound := false
			if startSfb < sfbCnt {
				possibleRegionFound = true
				for sfb = startSfb; sfb < stopSfb; sfb++ {
					if scf[sfb] != fdkIntMin && scfAct < minScf[sfb] {
						possibleRegionFound = false
						break
					}
				}
			}

			if possibleRegionFound {
				for sfb = startSfb; sfb < stopSfb; sfb++ {
					if scfTmp[sfb] != fdkIntMin {
						scfTmp[sfb] = scfAct
					}
				}

				deltaScfBits := FDKaacEncCountScfBitsDiff(scf, scfTmp[:], sfbCnt, startSfb, stopSfb)
				deltaSpecPe := FDKaacEncCalcSpecPeDiff(
					qcOutChannel.SfbEnergyLdData[:], scf, scfTmp[:],
					sfbConstPePart, sfbFormFactorLdData, sfbNRelevantLines, startSfb, stopSfb,
				)
				deltaPeNew := deltaPe + deltaScfBits + deltaSpecPe

				if deltaPeNew < scfDeltaPeLimit {
					distOldSum := FixpDBL(0)
					distNewSum := FixpDBL(0)
					for sfb = startSfb; sfb < stopSfb; sfb++ {
						if scfTmp[sfb] != fdkIntMin {
							distOldSum += CalcInvLdData(sfbDist[sfb]) >> distFacShift

							sfbWidth := psyOutChan.SfbOffsets[sfb+1] - psyOutChan.SfbOffsets[sfb]
							sfbOffs := psyOutChan.SfbOffsets[sfb]
							sfbDistNew[sfb] = FDKaacEncCalcSfbDist(
								qcOutChannel.MdctSpectrum[sfbOffs:],
								quantSpecTmp[sfbOffs:],
								sfbWidth,
								scfAct,
								dZoneQuantEnable,
							)

							if sfbDistNew[sfb] > qcOutChannel.SfbThresholdLdData[sfb] {
								distNewSum = distOldSum << 1
								break
							}
							distNewSum += CalcInvLdData(sfbDistNew[sfb]) >> distFacShift
						}
					}

					if distNewSum < distOldSum {
						deltaPe = deltaPeNew
						for sfb = startSfb; sfb < stopSfb; sfb++ {
							if scf[sfb] != fdkIntMin {
								sfbWidth := psyOutChan.SfbOffsets[sfb+1] - psyOutChan.SfbOffsets[sfb]
								sfbOffs := psyOutChan.SfbOffsets[sfb]
								scf[sfb] = scfAct
								sfbDist[sfb] = sfbDistNew[sfb]
								copy(quantSpec[sfbOffs:sfbOffs+sfbWidth], quantSpecTmp[sfbOffs:sfbOffs+sfbWidth])
							}
						}
					}
				}
			}

			if stopSfb > sfbCnt {
				break
			}
		}
		if scfAct <= scfMin {
			break
		}
	}
}

func FDKaacEncAssimilateMultipleScf2(
	psyOutChan *PsyOutChannel,
	qcOutChannel *QCOutChannel,
	quantSpec []int16,
	quantSpecTmp []int16,
	dZoneQuantEnable int,
	scf []int,
	minScf []int,
	sfbDist []FixpDBL,
	sfbConstPePart []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbNRelevantLines []FixpDBL,
) {
	checkAssimilateMultipleScfInputs(
		psyOutChan, qcOutChannel, quantSpec, quantSpecTmp, scf, minScf, sfbDist,
		sfbConstPePart, sfbFormFactorLdData, sfbNRelevantLines,
	)

	sfbCnt := psyOutChan.SfbCnt
	scfMin := fdkIntMax
	scfMax := fdkIntMin
	for sfb := 0; sfb < sfbCnt; sfb++ {
		if scf[sfb] != fdkIntMin {
			scfMin = minInt(scfMin, scf[sfb])
			scfMax = maxInt(scfMax, scf[sfb])
		}
	}
	if scfMax == fdkIntMin {
		return
	}

	var scfTmp [maxGroupedSFB]int
	var sfbDistNew [maxGroupedSFB]FixpDBL
	var sfbDistMax [maxGroupedSFB]FixpDBL
	deltaPe := FixpDBL(0)
	stopSfb := 0
	scfAct := fdkIntMin
	for {
		scfPrev := scfAct

		sfb := stopSfb
		for sfb < sfbCnt && scf[sfb] == fdkIntMin {
			sfb++
		}
		startSfb := sfb
		if startSfb >= sfbCnt {
			break
		}
		scfAct = scf[startSfb]
		sfb++
		for sfb < sfbCnt && (scf[sfb] == fdkIntMin || scf[sfb] == scf[startSfb]) {
			sfb++
		}
		stopSfb = sfb

		scfNext := scfAct
		if stopSfb < sfbCnt {
			scfNext = scf[stopSfb]
		}
		if scfPrev == fdkIntMin {
			scfPrev = scfAct
		}

		scfPrevNextMax := maxInt(scfPrev, scfNext)
		scfPrevNextMin := minInt(scfPrev, scfNext)
		scfHi := maxInt(scfPrevNextMax, scfAct)
		scfLo := scfPrevNextMax
		if scfPrevNextMax >= scfAct {
			scfLo = minInt(scfAct, scfPrevNextMin)
		}

		if startSfb < sfbCnt && scfHi-scfLo <= maxScfDelta {
			if scfHi > scf[startSfb] {
				for sfb = startSfb; sfb < stopSfb; sfb++ {
					if scf[sfb] != fdkIntMin {
						sfbDistMax[sfb] = FMultDD(scfDistMaxFac, qcOutChannel.SfbThresholdLdData[sfb]) +
							FMultDD(scfDistMaxFac, sfbDist[sfb]) +
							FMultDD(scfDistMaxFac, sfbDist[sfb])
						sfbDistMax[sfb] = maxFixpDBL(sfbDistMax[sfb], qcOutChannel.SfbEnergyLdData[sfb]-scfLd1e3)
						sfbDistMax[sfb] = minFixpDBL(sfbDistMax[sfb], qcOutChannel.SfbThresholdLdData[sfb])
					}
				}

				bCheckScf := true
				for scfNew := scf[startSfb] + 1; scfNew <= scfHi; scfNew++ {
					copy(scfTmp[:], scf[:])
					for sfb = startSfb; sfb < stopSfb; sfb++ {
						if scfTmp[sfb] != fdkIntMin {
							scfTmp[sfb] = scfNew
						}
					}

					deltaScfBits := FDKaacEncCountScfBitsDiff(scf, scfTmp[:], sfbCnt, startSfb, stopSfb)
					deltaSpecPe := FDKaacEncCalcSpecPeDiff(
						qcOutChannel.SfbEnergyLdData[:], scf, scfTmp[:],
						sfbConstPePart, sfbFormFactorLdData, sfbNRelevantLines, startSfb, stopSfb,
					)
					deltaPeNew := deltaPe + deltaScfBits + deltaSpecPe
					if deltaPeNew < 0 {
						bSuccess := true
						for sfb = startSfb; sfb < stopSfb; sfb++ {
							if scfTmp[sfb] != fdkIntMin {
								sfbOffs := psyOutChan.SfbOffsets[sfb]
								sfbWidth := psyOutChan.SfbOffsets[sfb+1] - sfbOffs
								sfbDistNew[sfb] = FDKaacEncCalcSfbDist(
									qcOutChannel.MdctSpectrum[sfbOffs:],
									quantSpecTmp[sfbOffs:],
									sfbWidth,
									scfNew,
									dZoneQuantEnable,
								)
								if sfbDistNew[sfb] > sfbDistMax[sfb] {
									bSuccess = false
									if sfbDistNew[sfb] == qcOutChannel.SfbEnergyLdData[sfb] {
										bCheckScf = false
									}
									break
								}
							}
						}
						if !bCheckScf {
							break
						}
						if bSuccess {
							deltaPe = deltaPeNew
							for sfb = startSfb; sfb < stopSfb; sfb++ {
								if scf[sfb] != fdkIntMin {
									sfbOffs := psyOutChan.SfbOffsets[sfb]
									sfbWidth := psyOutChan.SfbOffsets[sfb+1] - sfbOffs
									scf[sfb] = scfNew
									sfbDist[sfb] = sfbDistNew[sfb]
									copy(quantSpec[sfbOffs:sfbOffs+sfbWidth], quantSpecTmp[sfbOffs:sfbOffs+sfbWidth])
								}
							}
						}
					}
				}
			}

			if scfAct == scf[startSfb] && scfLo < scfAct && scfMax-scfMin <= maxScfDelta {
				bminScfViolation := false
				copy(scfTmp[:], scf[:])
				scfNew := scfLo
				for sfb = startSfb; sfb < stopSfb; sfb++ {
					if scfTmp[sfb] != fdkIntMin {
						scfTmp[sfb] = scfNew
						if scfNew < minScf[sfb] {
							bminScfViolation = true
						}
					}
				}

				deltaPeNew := FixpDBL(0)
				if !bminScfViolation {
					deltaScfBits := FDKaacEncCountScfBitsDiff(scf, scfTmp[:], sfbCnt, startSfb, stopSfb)
					deltaSpecPe := FDKaacEncCalcSpecPeDiff(
						qcOutChannel.SfbEnergyLdData[:], scf, scfTmp[:],
						sfbConstPePart, sfbFormFactorLdData, sfbNRelevantLines, startSfb, stopSfb,
					)
					deltaPeNew = deltaPe + deltaScfBits + deltaSpecPe
				}

				if !bminScfViolation && deltaPeNew < 0 {
					distOldSum := FixpDBL(0)
					distNewSum := FixpDBL(0)
					for sfb = startSfb; sfb < stopSfb; sfb++ {
						if scfTmp[sfb] != fdkIntMin {
							distOldSum += CalcInvLdData(sfbDist[sfb]) >> distFacShift

							sfbOffs := psyOutChan.SfbOffsets[sfb]
							sfbWidth := psyOutChan.SfbOffsets[sfb+1] - sfbOffs
							sfbDistNew[sfb] = FDKaacEncCalcSfbDist(
								qcOutChannel.MdctSpectrum[sfbOffs:],
								quantSpecTmp[sfbOffs:],
								sfbWidth,
								scfNew,
								dZoneQuantEnable,
							)

							if sfbDistNew[sfb] > qcOutChannel.SfbThresholdLdData[sfb] {
								distNewSum = distOldSum << 1
								break
							}
							distNewSum += CalcInvLdData(sfbDistNew[sfb]) >> distFacShift
						}
					}
					if distNewSum < FMultDD(scfFinerDistRatio, distOldSum) {
						deltaPe = deltaPeNew
						for sfb = startSfb; sfb < stopSfb; sfb++ {
							if scf[sfb] != fdkIntMin {
								sfbOffs := psyOutChan.SfbOffsets[sfb]
								sfbWidth := psyOutChan.SfbOffsets[sfb+1] - sfbOffs
								scf[sfb] = scfNew
								sfbDist[sfb] = sfbDistNew[sfb]
								copy(quantSpec[sfbOffs:sfbOffs+sfbWidth], quantSpecTmp[sfbOffs:sfbOffs+sfbWidth])
							}
						}
					}
				}
			}

			if scfMax-scfMin <= maxScfDelta-3 {
				copy(scfTmp[:], scf[:])
				for i := 0; i < 3; i++ {
					scfNew := scfTmp[startSfb] - 1
					for sfb = startSfb; sfb < stopSfb; sfb++ {
						if scfTmp[sfb] != fdkIntMin {
							scfTmp[sfb] = scfNew
						}
					}
					deltaScfBits := FDKaacEncCountScfBitsDiff(scf, scfTmp[:], sfbCnt, startSfb, stopSfb)
					deltaPeNew := deltaPe + deltaScfBits
					if deltaPeNew <= 0 {
						bSuccess := true
						distOldSum := FixpDBL(0)
						distNewSum := FixpDBL(0)
						for sfb = startSfb; sfb < stopSfb; sfb++ {
							if scfTmp[sfb] != fdkIntMin {
								sfbOffs := psyOutChan.SfbOffsets[sfb]
								sfbWidth := psyOutChan.SfbOffsets[sfb+1] - sfbOffs
								sfbEnQ, sfbDistQ := FDKaacEncCalcSfbQuantEnergyAndDist(
									qcOutChannel.MdctSpectrum[sfbOffs:],
									quantSpec[sfbOffs:],
									sfbWidth,
									scfNew,
								)
								sfbDistNew[sfb] = sfbDistQ

								distOldSum += CalcInvLdData(sfbDist[sfb]) >> distFacShift
								distNewSum += CalcInvLdData(sfbDistNew[sfb]) >> distFacShift
								if sfbDistNew[sfb] > sfbDist[sfb]+scfDistIncreaseLimit ||
									sfbEnQ < qcOutChannel.SfbEnergyLdData[sfb]-scfEnergyDecreaseLimit {
									bSuccess = false
									break
								}
							}
						}
						if distNewSum < distOldSum && bSuccess {
							deltaPe = deltaPeNew
							for sfb = startSfb; sfb < stopSfb; sfb++ {
								if scf[sfb] != fdkIntMin {
									scf[sfb] = scfNew
									sfbDist[sfb] = sfbDistNew[sfb]
								}
							}
						}
					}
				}
			}
		}

		if stopSfb > sfbCnt {
			break
		}
	}
}

func checkFormFactorInputs(sfbFormFactorLdData []FixpDBL, psyOutChan *PsyOutChannel) {
	if psyOutChan == nil {
		panic("fdkaac: nil form-factor psy output")
	}
	if psyOutChan.SfbCnt <= 0 || psyOutChan.SfbCnt > maxGroupedSFB || psyOutChan.SfbPerGroup <= 0 || psyOutChan.SfbCnt%psyOutChan.SfbPerGroup != 0 {
		panic("fdkaac: invalid form-factor band count")
	}
	if psyOutChan.MaxSfbPerGroup <= 0 || psyOutChan.MaxSfbPerGroup > psyOutChan.SfbPerGroup {
		panic("fdkaac: invalid form-factor group width")
	}
	if len(sfbFormFactorLdData) < psyOutChan.SfbCnt {
		panic("fdkaac: short form-factor output")
	}
	prev := psyOutChan.SfbOffsets[0]
	if prev < 0 {
		panic("fdkaac: invalid form-factor offset")
	}
	for i := 0; i < psyOutChan.SfbCnt; i++ {
		next := psyOutChan.SfbOffsets[i+1]
		if next < prev {
			panic("fdkaac: invalid form-factor offset")
		}
		prev = next
	}
	if prev > len(psyOutChan.MdctSpectrum) {
		panic("fdkaac: short form-factor spectrum")
	}
}

func checkSfbRelevantLinesInputs(
	sfbFormFactorLdData []FixpDBL,
	sfbEnergyLdData []FixpDBL,
	sfbThresholdLdData []FixpDBL,
	sfbOffsets []int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	sfbNRelevantLines []FixpDBL,
) {
	if sfbCnt <= 0 || sfbCnt > maxGroupedSFB || sfbPerGroup <= 0 || sfbCnt%sfbPerGroup != 0 {
		panic("fdkaac: invalid relevant-lines band count")
	}
	if maxSfbPerGroup <= 0 || maxSfbPerGroup > sfbPerGroup {
		panic("fdkaac: invalid relevant-lines group width")
	}
	if len(sfbFormFactorLdData) < sfbCnt || len(sfbEnergyLdData) < sfbCnt || len(sfbThresholdLdData) < sfbCnt || len(sfbNRelevantLines) < sfbCnt {
		panic("fdkaac: short relevant-lines data")
	}
	if len(sfbOffsets) < sfbCnt+1 {
		panic("fdkaac: short relevant-lines offsets")
	}
	prev := sfbOffsets[0]
	if prev < 0 {
		panic("fdkaac: invalid relevant-lines offset")
	}
	for i := 0; i < sfbCnt; i++ {
		next := sfbOffsets[i+1]
		if next < prev {
			panic("fdkaac: invalid relevant-lines offset")
		}
		if i%sfbPerGroup < maxSfbPerGroup && next == prev {
			panic("fdkaac: empty relevant-lines active band")
		}
		prev = next
	}
}

func checkScfDiffInputs(scfOld []int, scfNew []int, sfbCnt int, startSfb int, stopSfb int) {
	if sfbCnt <= 0 || sfbCnt > maxGroupedSFB || startSfb < 0 || stopSfb <= startSfb || stopSfb > sfbCnt {
		panic("fdkaac: invalid scalefactor diff range")
	}
	if len(scfOld) < sfbCnt || len(scfNew) < sfbCnt {
		panic("fdkaac: short scalefactor diff data")
	}
}

func checkSpecPeDiffInputs(
	sfbEnergyLdData []FixpDBL,
	scfOld []int,
	scfNew []int,
	sfbConstPePart []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbNRelevantLines []FixpDBL,
	startSfb int,
	stopSfb int,
) {
	if startSfb < 0 || stopSfb <= startSfb || stopSfb > maxGroupedSFB {
		panic("fdkaac: invalid spec-pe diff range")
	}
	if len(sfbEnergyLdData) < stopSfb || len(scfOld) < stopSfb || len(scfNew) < stopSfb ||
		len(sfbConstPePart) < stopSfb || len(sfbFormFactorLdData) < stopSfb || len(sfbNRelevantLines) < stopSfb {
		panic("fdkaac: short spec-pe diff data")
	}
}

func checkImproveScfInputs(
	spec []FixpDBL,
	quantSpec []int16,
	quantSpecTmp []int16,
	sfbWidth int,
	distLdData *FixpDBL,
	minScfCalculated *int,
) {
	if sfbWidth < 0 {
		panic("fdkaac: negative improve-scf width")
	}
	if len(spec) < sfbWidth || len(quantSpec) < sfbWidth || len(quantSpecTmp) < sfbWidth {
		panic("fdkaac: short improve-scf data")
	}
	if distLdData == nil || minScfCalculated == nil {
		panic("fdkaac: nil improve-scf output")
	}
}

func checkEstimateScaleFactorsChannelInputs(
	qcOutChannel *QCOutChannel,
	psyOutChannel *PsyOutChannel,
	scf []int,
	globalGain *int,
	sfbFormFactorLdData []FixpDBL,
	invQuant int,
	quantSpec []int16,
	quantSpecTmp []int16,
) int {
	if qcOutChannel == nil {
		panic("fdkaac: nil estimate-scalefactor qc output")
	}
	if psyOutChannel == nil {
		panic("fdkaac: nil estimate-scalefactor psy output")
	}
	if globalGain == nil {
		panic("fdkaac: nil estimate-scalefactor global gain")
	}
	if psyOutChannel.SfbCnt <= 0 || psyOutChannel.SfbCnt > maxGroupedSFB || psyOutChannel.SfbPerGroup <= 0 || psyOutChannel.SfbCnt%psyOutChannel.SfbPerGroup != 0 {
		panic("fdkaac: invalid estimate-scalefactor band count")
	}
	if psyOutChannel.MaxSfbPerGroup <= 0 || psyOutChannel.MaxSfbPerGroup > psyOutChannel.SfbPerGroup {
		panic("fdkaac: invalid estimate-scalefactor group width")
	}
	if len(scf) < psyOutChannel.SfbCnt || len(sfbFormFactorLdData) < psyOutChannel.SfbCnt {
		panic("fdkaac: short estimate-scalefactor band data")
	}

	prev := psyOutChannel.SfbOffsets[0]
	if prev < 0 {
		panic("fdkaac: invalid estimate-scalefactor offset")
	}
	for i := 0; i < psyOutChannel.SfbCnt; i++ {
		next := psyOutChannel.SfbOffsets[i+1]
		if next < prev {
			panic("fdkaac: invalid estimate-scalefactor offset")
		}
		if i%psyOutChannel.SfbPerGroup < psyOutChannel.MaxSfbPerGroup && next == prev {
			panic("fdkaac: empty estimate-scalefactor active band")
		}
		prev = next
	}
	if prev > len(qcOutChannel.MdctSpectrum) {
		panic("fdkaac: short estimate-scalefactor spectrum")
	}
	if invQuant > 0 && (len(quantSpec) < prev || len(quantSpecTmp) < prev) {
		panic("fdkaac: short estimate-scalefactor quant data")
	}
	return prev
}

func checkEstimateScaleFactorsInputs(
	psyOutChannel []*PsyOutChannel,
	qcOutChannel []*QCOutChannel,
	invQuant int,
	nChannels int,
	quantSpecTmp []int16,
) {
	if nChannels < 0 || len(qcOutChannel) < nChannels || len(psyOutChannel) < nChannels {
		panic("fdkaac: invalid estimate-scalefactor channel count")
	}
	if invQuant > 0 && len(quantSpecTmp) < maxSpectralLines {
		panic("fdkaac: short estimate-scalefactor scratch")
	}
}

func checkAssimilateSingleScfInputs(
	psyOutChan *PsyOutChannel,
	qcOutChannel *QCOutChannel,
	quantSpec []int16,
	quantSpecTmp []int16,
	scf []int,
	minScf []int,
	sfbDist []FixpDBL,
	sfbConstPePart []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbNRelevantLines []FixpDBL,
	minScfCalculated []int,
) {
	if psyOutChan == nil {
		panic("fdkaac: nil assimilate-single psy output")
	}
	if qcOutChannel == nil {
		panic("fdkaac: nil assimilate-single qc output")
	}
	if psyOutChan.SfbCnt <= 0 || psyOutChan.SfbCnt > maxGroupedSFB || psyOutChan.SfbPerGroup <= 0 || psyOutChan.SfbCnt%psyOutChan.SfbPerGroup != 0 {
		panic("fdkaac: invalid assimilate-single band count")
	}
	if psyOutChan.MaxSfbPerGroup <= 0 || psyOutChan.MaxSfbPerGroup > psyOutChan.SfbPerGroup {
		panic("fdkaac: invalid assimilate-single group width")
	}
	if len(scf) < psyOutChan.SfbCnt || len(minScf) < psyOutChan.SfbCnt || len(sfbDist) < psyOutChan.SfbCnt ||
		len(sfbConstPePart) < psyOutChan.SfbCnt || len(sfbFormFactorLdData) < psyOutChan.SfbCnt ||
		len(sfbNRelevantLines) < psyOutChan.SfbCnt || len(minScfCalculated) < psyOutChan.SfbCnt {
		panic("fdkaac: short assimilate-single band data")
	}
	prev := psyOutChan.SfbOffsets[0]
	if prev < 0 {
		panic("fdkaac: invalid assimilate-single offset")
	}
	for i := 0; i < psyOutChan.SfbCnt; i++ {
		next := psyOutChan.SfbOffsets[i+1]
		if next < prev {
			panic("fdkaac: invalid assimilate-single offset")
		}
		prev = next
	}
	if prev > len(qcOutChannel.MdctSpectrum) || len(quantSpec) < prev || len(quantSpecTmp) < prev {
		panic("fdkaac: short assimilate-single spectrum")
	}
}

func checkAssimilateMultipleScfInputs(
	psyOutChan *PsyOutChannel,
	qcOutChannel *QCOutChannel,
	quantSpec []int16,
	quantSpecTmp []int16,
	scf []int,
	minScf []int,
	sfbDist []FixpDBL,
	sfbConstPePart []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbNRelevantLines []FixpDBL,
) {
	if psyOutChan == nil {
		panic("fdkaac: nil assimilate-multiple psy output")
	}
	if qcOutChannel == nil {
		panic("fdkaac: nil assimilate-multiple qc output")
	}
	if psyOutChan.SfbCnt <= 0 || psyOutChan.SfbCnt > maxGroupedSFB || psyOutChan.SfbPerGroup <= 0 || psyOutChan.SfbCnt%psyOutChan.SfbPerGroup != 0 {
		panic("fdkaac: invalid assimilate-multiple band count")
	}
	if psyOutChan.MaxSfbPerGroup <= 0 || psyOutChan.MaxSfbPerGroup > psyOutChan.SfbPerGroup {
		panic("fdkaac: invalid assimilate-multiple group width")
	}
	if len(scf) < psyOutChan.SfbCnt || len(minScf) < psyOutChan.SfbCnt || len(sfbDist) < psyOutChan.SfbCnt ||
		len(sfbConstPePart) < psyOutChan.SfbCnt || len(sfbFormFactorLdData) < psyOutChan.SfbCnt ||
		len(sfbNRelevantLines) < psyOutChan.SfbCnt {
		panic("fdkaac: short assimilate-multiple band data")
	}
	prev := psyOutChan.SfbOffsets[0]
	if prev < 0 {
		panic("fdkaac: invalid assimilate-multiple offset")
	}
	for i := 0; i < psyOutChan.SfbCnt; i++ {
		next := psyOutChan.SfbOffsets[i+1]
		if next < prev {
			panic("fdkaac: invalid assimilate-multiple offset")
		}
		prev = next
	}
	if prev > len(qcOutChannel.MdctSpectrum) || len(quantSpec) < prev || len(quantSpecTmp) < prev {
		panic("fdkaac: short assimilate-multiple spectrum")
	}
}
