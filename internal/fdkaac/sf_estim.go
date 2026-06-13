package fdkaac

const formFacShift = 6

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
