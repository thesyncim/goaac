package fdkaac

const (
	maxQuant              = 8191
	mantDigits            = 9
	mantSize              = 1 << mantDigits
	quantKShift           = 16
	quantDeadZoneK        = FixpDBL(0x1d70a3d7)
	quantNoDeadZoneK      = FixpDBL(0x33e425af)
	quantEnergyLdDataBias = FixpDBL(0x04000000)
)

func FDKaacEncQuantizeLines(gain int, noOfLines int, mdctSpectrum []FixpDBL, quaSpectrum []int16, dZoneQuantEnable int) {
	checkQuantLinesInputs(noOfLines, mdctSpectrum, quaSpectrum)
	for line := 0; line < noOfLines; line++ {
		quaSpectrum[line] = fdkaacEncQuantizeLine(gain, mdctSpectrum[line], dZoneQuantEnable)
	}
}

func FDKaacEncInvQuantizeLines(gain int, noOfLines int, quantSpectrum []int16, mdctSpectrum []FixpDBL) {
	checkInvQuantLinesInputs(noOfLines, quantSpectrum, mdctSpectrum)
	for line := 0; line < noOfLines; line++ {
		mdctSpectrum[line] = fdkaacEncInvQuantizeLine(gain, quantSpectrum[line])
	}
}

func FDKaacEncQuantizeSpectrum(
	sfbCnt int,
	maxSfbPerGroup int,
	sfbPerGroup int,
	sfbOffset []int,
	mdctSpectrum []FixpDBL,
	globalGain int,
	scalefactors []int,
	quantizedSpectrum []int16,
	dZoneQuantEnable int,
) {
	checkQuantizeSpectrumInputs(sfbCnt, maxSfbPerGroup, sfbPerGroup, sfbOffset, mdctSpectrum, scalefactors, quantizedSpectrum)

	for sfbOffs := 0; sfbOffs < sfbCnt; sfbOffs += sfbPerGroup {
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			idx := sfbOffs + sfb
			start := sfbOffset[idx]
			stop := sfbOffset[idx+1]
			FDKaacEncQuantizeLines(globalGain-scalefactors[idx], stop-start, mdctSpectrum[start:stop], quantizedSpectrum[start:stop], dZoneQuantEnable)
		}
	}
}

func FDKaacEncCalcSfbDist(mdctSpectrum []FixpDBL, quantSpectrum []int16, noOfLines int, gain int, dZoneQuantEnable int) FixpDBL {
	checkCalcSfbDistInputs(mdctSpectrum, quantSpectrum, noOfLines)

	xfsf := FixpDBL(0)
	for i := 0; i < noOfLines; i++ {
		quantSpectrum[i] = fdkaacEncQuantizeLine(gain, mdctSpectrum[i], dZoneQuantEnable)
		if absInt16(quantSpectrum[i]) > maxQuant {
			return 0
		}

		invQuantSpec := fdkaacEncInvQuantizeLine(gain, quantSpectrum[i])
		diff := fixpAbsDBL(fixpAbsDBL(invQuantSpec) - fixpAbsDBL(mdctSpectrum[i]>>1))

		scale := CountLeadingBits(diff)
		diff = ScaleValueDBL(diff, scale)
		diff = FixPow2D(diff)
		scale = minInt(2*(scale-1), DfractBits-1)
		diff = ScaleValueDBL(diff, -scale)

		xfsf += diff
	}
	return CalcLdData(xfsf)
}

func FDKaacEncCalcSfbQuantEnergyAndDist(mdctSpectrum []FixpDBL, quantSpectrum []int16, noOfLines int, gain int) (FixpDBL, FixpDBL) {
	checkCalcSfbDistInputs(mdctSpectrum, quantSpectrum, noOfLines)

	energy := FixpDBL(0)
	distortion := FixpDBL(0)
	for i := 0; i < noOfLines; i++ {
		if absInt16(quantSpectrum[i]) > maxQuant {
			return 0, 0
		}

		invQuantSpec := fdkaacEncInvQuantizeLine(gain, quantSpectrum[i])
		energy += FixPow2D(invQuantSpec)

		diff := fixpAbsDBL(fixpAbsDBL(invQuantSpec) - fixpAbsDBL(mdctSpectrum[i]>>1))
		scale := CountLeadingBits(diff)
		diff = ScaleValueDBL(diff, scale)
		diff = FixPow2D(diff)
		scale = minInt(2*(scale-1), DfractBits-1)
		diff = ScaleValueDBL(diff, -scale)

		distortion += diff
	}
	return CalcLdData(energy) + quantEnergyLdDataBias, CalcLdData(distortion)
}

func fdkaacEncQuantizeLine(gain int, mdctSpectrum FixpDBL, dZoneQuantEnable int) int16 {
	k := quantNoDeadZoneK >> quantKShift
	if dZoneQuantEnable != 0 {
		k = quantDeadZoneK >> quantKShift
	}

	quantizer := fdkaacEncQuantTableQ[(-gain)&3]
	quantizerShift := ((-gain) >> 2) + 1
	accu := FMultDiv2DD(mdctSpectrum, quantizer)

	if accu < 0 {
		accu = -accu
		accu = fdkaacEncQuantizeMagnitude(accu, quantizerShift)
		return int16(-int((k + accu) >> (DfractBits - 1 - FractBits)))
	}
	if accu > 0 {
		accu = fdkaacEncQuantizeMagnitude(accu, quantizerShift)
		return int16((k + accu) >> (DfractBits - 1 - FractBits))
	}
	return 0
}

func fdkaacEncQuantizeMagnitude(accu FixpDBL, quantizerShift int) FixpDBL {
	accuShift := FixNormZD(accu) - 1
	accu <<= uint(accuShift)
	tabIndex := int((accu >> (DfractBits - 2 - mantDigits)) &^ FixpDBL(mantSize))
	totalShift := quantizerShift - accuShift + 1
	accu = FMultDiv2DD(fdkaacEncMTab34[tabIndex], fdkaacEncQuantTableE[totalShift&3])
	totalShift = (FractBits - 4) - (3 * (totalShift >> 2))
	if totalShift < 0 {
		panic("fdkaac: quantizer shift underflow")
	}
	accu >>= uint(minInt(totalShift, DfractBits-1))
	return accu
}

func fdkaacEncInvQuantizeLine(gain int, quantSpectrum int16) FixpDBL {
	if quantSpectrum == 0 {
		return 0
	}
	if absInt16(quantSpectrum) > maxQuant {
		panic("fdkaac: quantized spectrum out of range")
	}

	iquantizerMod := gain & 3
	iquantizerShift := gain >> 2
	sign := FixpDBL(1)
	accu := FixpDBL(quantSpectrum)
	if accu < 0 {
		sign = -1
		accu = -accu
	}

	ex := CountLeadingBits(accu)
	accu <<= uint(ex)
	specExp := DfractBits - 1 - ex
	if specExp >= len(fdkaacEncSpecExpTableComb[iquantizerMod]) {
		panic("fdkaac: inverse quantizer exponent out of range")
	}
	tabIndex := int((accu >> (DfractBits - 2 - mantDigits)) &^ FixpDBL(mantSize))

	s := fdkaacEncMTab43Elc[tabIndex]
	t := fdkaacEncSpecExpMantTableCombElc[iquantizerMod][specExp]
	accu = FMultDD(s, t)

	specExp = int(fdkaacEncSpecExpTableComb[iquantizerMod][specExp]) - 1
	shift := -iquantizerShift - specExp
	if shift < 0 {
		accu <<= uint(-shift)
	} else {
		accu >>= uint(shift)
	}

	if sign < 0 {
		return -accu
	}
	return accu
}

func checkQuantLinesInputs(noOfLines int, mdctSpectrum []FixpDBL, quaSpectrum []int16) {
	if noOfLines < 0 {
		panic("fdkaac: negative quantize line count")
	}
	if len(mdctSpectrum) < noOfLines || len(quaSpectrum) < noOfLines {
		panic("fdkaac: short quantize line data")
	}
}

func checkInvQuantLinesInputs(noOfLines int, quantSpectrum []int16, mdctSpectrum []FixpDBL) {
	if noOfLines < 0 {
		panic("fdkaac: negative inverse quantize line count")
	}
	if len(quantSpectrum) < noOfLines || len(mdctSpectrum) < noOfLines {
		panic("fdkaac: short inverse quantize line data")
	}
}

func checkQuantizeSpectrumInputs(
	sfbCnt int,
	maxSfbPerGroup int,
	sfbPerGroup int,
	sfbOffset []int,
	mdctSpectrum []FixpDBL,
	scalefactors []int,
	quantizedSpectrum []int16,
) {
	if sfbCnt <= 0 || sfbCnt > maxGroupedSFB || sfbPerGroup <= 0 || sfbCnt%sfbPerGroup != 0 {
		panic("fdkaac: invalid quantize band count")
	}
	if maxSfbPerGroup <= 0 || maxSfbPerGroup > sfbPerGroup {
		panic("fdkaac: invalid quantize group width")
	}
	if len(scalefactors) < sfbCnt {
		panic("fdkaac: short quantize band data")
	}
	checkGroupedSfbOffsets(
		sfbOffset,
		sfbCnt,
		sfbPerGroup,
		maxSfbPerGroup,
		false,
		"fdkaac: invalid quantize offset",
		"fdkaac: short quantize spectrum",
		len(mdctSpectrum),
		len(quantizedSpectrum),
	)
}

func checkCalcSfbDistInputs(mdctSpectrum []FixpDBL, quantSpectrum []int16, noOfLines int) {
	if noOfLines < 0 {
		panic("fdkaac: negative SFB distortion line count")
	}
	if len(mdctSpectrum) < noOfLines || len(quantSpectrum) < noOfLines {
		panic("fdkaac: short SFB distortion data")
	}
}

func absInt16(x int16) int {
	if x < 0 {
		return -int(x)
	}
	return int(x)
}
