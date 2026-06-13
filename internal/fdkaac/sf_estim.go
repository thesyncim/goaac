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
const upcountLimit = 1
const distFactorLdData = FixpDBL(-10802114) // FDK FL2FXCONST_DBL(-0.0050301265), ld64(1/1.25).

type QCOutChannel struct {
	MdctSpectrum        [1024]FixpDBL
	SfbFormFactorLdData [maxGroupedSFB]FixpDBL
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
