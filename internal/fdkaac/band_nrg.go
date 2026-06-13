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

func FDKaacEncCheckBandEnergyOptim(mdctSpectrum []FixpDBL, sfbMaxScaleSpec []int, bandOffset []int, numBands int, bandEnergy []FixpDBL, bandEnergyLdData []FixpDBL, minSpecShift int) FixpDBL {
	if numBands <= 0 {
		panic("fdkaac: empty band energy")
	}
	checkBandEnergySpectrum(mdctSpectrum, bandOffset, numBands)
	checkBandEnergyScales(sfbMaxScaleSpec, numBands)
	checkBandEnergyOutput(bandEnergy, numBands)
	checkBandEnergyOutput(bandEnergyLdData, numBands)

	nr := 0
	maxNrgLd := ldDataMinusOne
	for i := 0; i < numBands; i++ {
		scale := maxInt(0, sfbMaxScaleSpec[i]-4)
		tmp := FixpDBL(0)
		for j := bandOffset[i]; j < bandOffset[i+1]; j++ {
			spec := mdctSpectrum[j] << uint(scale)
			tmp = FPow2AddDiv2D(tmp, spec)
		}
		bandEnergy[i] = tmp << 1

		bandEnergyLdData[i] = CalcLdData(bandEnergy[i])
		if bandEnergyLdData[i] != ldDataMinusOne {
			bandEnergyLdData[i] -= FixpDBL(scale) * ldDataStep2Over64
		}
		if bandEnergyLdData[i] > maxNrgLd {
			maxNrgLd = bandEnergyLdData[i]
			nr = i
		}
	}

	scale := maxInt(0, sfbMaxScaleSpec[nr]-4)
	scale = maxInt(2*(minSpecShift-scale), -(DfractBits - 1))
	return ScaleValueDBL(bandEnergy[nr], scale)
}

func FDKaacEncCalcBandEnergyOptimLong(mdctSpectrum []FixpDBL, sfbMaxScaleSpec []int, bandOffset []int, numBands int, bandEnergy []FixpDBL, bandEnergyLdData []FixpDBL) int {
	checkBandEnergySpectrum(mdctSpectrum, bandOffset, numBands)
	checkBandEnergyScales(sfbMaxScaleSpec, numBands)
	checkBandEnergyOutput(bandEnergy, numBands)
	checkBandEnergyOutput(bandEnergyLdData, numBands)

	shiftBits := 0
	maxNrgLd := FixpDBL(0)

	for i := 0; i < numBands; i++ {
		leadingBits := sfbMaxScaleSpec[i] - 4
		tmp := FixpDBL(0)
		if leadingBits >= 0 {
			shift := uint(leadingBits)
			for j := bandOffset[i]; j < bandOffset[i+1]; j++ {
				spec := mdctSpectrum[j] << shift
				tmp = FPow2AddDiv2D(tmp, spec)
			}
		} else {
			shift := uint(-leadingBits)
			for j := bandOffset[i]; j < bandOffset[i+1]; j++ {
				spec := mdctSpectrum[j] >> shift
				tmp = FPow2AddDiv2D(tmp, spec)
			}
		}
		bandEnergy[i] = tmp << 1
	}

	LdDataVector(bandEnergy, bandEnergyLdData, numBands)
	for i := numBands - 1; i >= 0; i-- {
		scaleDiff := FixpDBL(sfbMaxScaleSpec[i]-4) * ldDataStep2Over64
		if bandEnergyLdData[i] >= ((ldDataMinusOne >> 1) + (scaleDiff >> 1)) {
			bandEnergyLdData[i] -= scaleDiff
		} else {
			bandEnergyLdData[i] = ldDataMinusOne
		}
		if bandEnergyLdData[i] > maxNrgLd {
			maxNrgLd = bandEnergyLdData[i]
		}
	}

	if maxNrgLd <= 0 {
		for i := numBands - 1; i >= 0; i-- {
			scale := minInt((sfbMaxScaleSpec[i]-4)<<1, DfractBits-1)
			bandEnergy[i] = ScaleValueDBL(bandEnergy[i], -scale)
		}
		return 0
	}

	for maxNrgLd > 0 {
		maxNrgLd -= ldDataStep2Over64
		shiftBits++
	}
	for i := numBands - 1; i >= 0; i-- {
		scale := minInt(((sfbMaxScaleSpec[i]-4)+shiftBits)<<1, DfractBits-1)
		bandEnergyLdData[i] -= FixpDBL(shiftBits) * ldDataStep2Over64
		bandEnergy[i] = ScaleValueDBL(bandEnergy[i], -scale)
	}
	return shiftBits
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
	checkBandEnergySpectrum(mdctSpectrumLeft, bandOffset, numBands)
	checkBandEnergySpectrum(mdctSpectrumRight, bandOffset, numBands)
	checkBandEnergyScales(sfbMaxScaleSpecLeft, numBands)
	checkBandEnergyScales(sfbMaxScaleSpecRight, numBands)
	checkBandEnergyOutput(bandEnergyMid, numBands)
	checkBandEnergyOutput(bandEnergySide, numBands)
	if calcLdData {
		checkBandEnergyOutput(bandEnergyMidLdData, numBands)
		checkBandEnergyOutput(bandEnergySideLdData, numBands)
	}

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

	if calcLdData {
		LdDataVector(bandEnergyMid, bandEnergyMidLdData, numBands)
		LdDataVector(bandEnergySide, bandEnergySideLdData, numBands)
	}

	for i := 0; i < numBands; i++ {
		minScale := minInt(sfbMaxScaleSpecLeft[i], sfbMaxScaleSpecRight[i])
		scale := maxInt(0, 2*(minScale-4))

		if calcLdData {
			minus := FixpDBL(scale) * ldDataStep1Over64
			if bandEnergyMidLdData[i] != ldDataMinusOne {
				bandEnergyMidLdData[i] -= minus
			}
			if bandEnergySideLdData[i] != ldDataMinusOne {
				bandEnergySideLdData[i] -= minus
			}
		}

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
