package fdkaac

func FDKaacEncCalcSfbMaxScaleSpec(mdctSpectrum []FixpDBL, bandOffset []int, sfbMaxScaleSpec []int, numBands int) {
	checkBandEnergySpectrum(mdctSpectrum, bandOffset, numBands)
	checkBandEnergyIntOutput(sfbMaxScaleSpec, numBands)

	for i := 0; i < numBands; i++ {
		maxSpc := FixpDBL(0)
		for j := bandOffset[i]; j < bandOffset[i+1]; j++ {
			tmp := fixpAbsDBL(mdctSpectrum[j])
			if tmp > maxSpc {
				maxSpc = tmp
			}
		}
		leading := FixNormZD(maxSpc) - 1
		if leading > DfractBits-2 {
			leading = DfractBits - 2
		}
		sfbMaxScaleSpec[i] = leading
	}
}

func FDKaacEncCalcBandEnergyOptimShort(mdctSpectrum []FixpDBL, sfbMaxScaleSpec []int, bandOffset []int, numBands int, bandEnergy []FixpDBL) {
	checkBandEnergySpectrum(mdctSpectrum, bandOffset, numBands)
	checkBandEnergyScales(sfbMaxScaleSpec, numBands)
	checkBandEnergyOutput(bandEnergy, numBands)

	for i := 0; i < numBands; i++ {
		leadingBits := sfbMaxScaleSpec[i] - 3
		tmp := FixpDBL(0)
		for j := bandOffset[i]; j < bandOffset[i+1]; j++ {
			spec := ScaleValueDBL(mdctSpectrum[j], leadingBits)
			tmp = FPow2AddDiv2D(tmp, spec)
		}
		bandEnergy[i] = tmp
	}

	for i := 0; i < numBands; i++ {
		scale := (2 * (sfbMaxScaleSpec[i] - 3)) - 1
		scale = maxInt(minInt(scale, DfractBits-1), -(DfractBits - 1))
		bandEnergy[i] = ScaleValueSaturateDBL(bandEnergy[i], -scale)
	}
}

func FDKaacEncCalcBandNrgMSOpt(
	mdctSpectrumLeft []FixpDBL,
	mdctSpectrumRight []FixpDBL,
	sfbMaxScaleSpecLeft []int,
	sfbMaxScaleSpecRight []int,
	bandOffset []int,
	numBands int,
	bandEnergyMid []FixpDBL,
	bandEnergySide []FixpDBL,
	calcLdData bool,
	bandEnergyMidLdData []FixpDBL,
	bandEnergySideLdData []FixpDBL,
) {
	if calcLdData {
		panic("fdkaac: band energy ld-data not supported")
	}
	_ = bandEnergyMidLdData
	_ = bandEnergySideLdData

	checkBandEnergySpectrum(mdctSpectrumLeft, bandOffset, numBands)
	checkBandEnergySpectrum(mdctSpectrumRight, bandOffset, numBands)
	checkBandEnergyScales(sfbMaxScaleSpecLeft, numBands)
	checkBandEnergyScales(sfbMaxScaleSpecRight, numBands)
	checkBandEnergyOutput(bandEnergyMid, numBands)
	checkBandEnergyOutput(bandEnergySide, numBands)

	maxNrg := MaxValDBL >> 1
	for i := 0; i < numBands; i++ {
		nrgMid := FixpDBL(0)
		nrgSide := FixpDBL(0)
		minScale := minInt(sfbMaxScaleSpecLeft[i], sfbMaxScaleSpecRight[i]) - 4
		minScale = maxInt(0, minScale)

		if minScale > 0 {
			shift := uint(minScale - 1)
			for j := bandOffset[i]; j < bandOffset[i+1]; j++ {
				specL := mdctSpectrumLeft[j] << shift
				specR := mdctSpectrumRight[j] << shift
				specM := specL + specR
				specS := specL - specR
				nrgMid = FPow2AddDiv2D(nrgMid, specM)
				nrgSide = FPow2AddDiv2D(nrgSide, specS)
			}
		} else {
			for j := bandOffset[i]; j < bandOffset[i+1]; j++ {
				specL := mdctSpectrumLeft[j] >> 1
				specR := mdctSpectrumRight[j] >> 1
				specM := specL + specR
				specS := specL - specR
				nrgMid = FPow2AddDiv2D(nrgMid, specM)
				nrgSide = FPow2AddDiv2D(nrgSide, specS)
			}
		}

		if nrgMid > maxNrg {
			nrgMid = maxNrg
		}
		if nrgSide > maxNrg {
			nrgSide = maxNrg
		}
		bandEnergyMid[i] = nrgMid << 1
		bandEnergySide[i] = nrgSide << 1
	}

	for i := 0; i < numBands; i++ {
		minScale := minInt(sfbMaxScaleSpecLeft[i], sfbMaxScaleSpecRight[i])
		scale := maxInt(0, 2*(minScale-4))
		scale = minInt(scale, DfractBits-1)
		bandEnergyMid[i] >>= uint(scale)
		bandEnergySide[i] >>= uint(scale)
	}
}

func checkBandEnergySpectrum(mdctSpectrum []FixpDBL, bandOffset []int, numBands int) {
	if numBands < 0 {
		panic("fdkaac: negative band count")
	}
	if len(bandOffset) < numBands+1 {
		panic("fdkaac: short band offsets")
	}
	start := bandOffset[0]
	if start < 0 || start > len(mdctSpectrum) {
		panic("fdkaac: invalid band offset")
	}
	for i := 0; i < numBands; i++ {
		end := bandOffset[i+1]
		if end < start || end > len(mdctSpectrum) {
			panic("fdkaac: invalid band offset")
		}
		start = end
	}
}

func checkBandEnergyScales(sfbMaxScaleSpec []int, numBands int) {
	checkBandEnergyIntOutput(sfbMaxScaleSpec, numBands)
	for i := 0; i < numBands; i++ {
		if sfbMaxScaleSpec[i] < 0 || sfbMaxScaleSpec[i] > DfractBits-2 {
			panic("fdkaac: invalid band scale")
		}
	}
}

func checkBandEnergyIntOutput(output []int, numBands int) {
	if len(output) < numBands {
		panic("fdkaac: short band int output")
	}
}

func checkBandEnergyOutput(output []FixpDBL, numBands int) {
	if len(output) < numBands {
		panic("fdkaac: short band output")
	}
}

func fixpAbsDBL(x FixpDBL) FixpDBL {
	if x == MinValDBL {
		return MaxValDBL
	}
	if x < 0 {
		return -x
	}
	return x
}
