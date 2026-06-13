package fdkaac

const formFacShift = 6
const asPeFacShift = 7
const asPeFacLdData = FixpDBL(0x0e000000)

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
