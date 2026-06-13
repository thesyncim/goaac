package fdkaac

const (
	tonalityHalf           FixpDBL = 0x40000000
	tonalityLdConvtone     FixpDBL = 0x06000000
	tonalityChaosThreshold FixpDBL = -111465353
	tonalityNormLog        FixpDBL = -646457015 // 0xd977d949
)

type TonalityScratch struct {
	ChaosMeasurePerLine [maxSpectralLines]FixpDBL
}

func FDKaacEncCalculateChaosMeasure(mdctData []FixpDBL, numberOfLines int, chaosMeasure []FixpDBL) {
	checkChaosMeasureLineInputs(mdctData, numberOfLines, chaosMeasure)
	fdkaacEncCalculateChaosMeasurePeakFast(mdctData, numberOfLines, chaosMeasure)
}

func FDKaacEncCalculateFullTonality(
	spectrum []FixpDBL,
	sfbMaxScaleSpec []int,
	sfbEnergyLD64 []FixpDBL,
	sfbTonality []FixpSGL,
	sfbCnt int,
	sfbOffset []int,
	usePns int,
	scratch *TonalityScratch,
) {
	numberOfLines := checkTonalityMetadata(spectrum, sfbCnt, sfbOffset)
	if usePns == 0 {
		return
	}
	checkTonalityBandInputs(sfbMaxScaleSpec, sfbEnergyLD64, sfbTonality, sfbCnt, scratch)

	chaosMeasurePerLine := scratch.ChaosMeasurePerLine[:]
	FDKaacEncCalculateChaosMeasure(spectrum, numberOfLines, chaosMeasurePerLine)
	fdkaacEncSmoothChaosMeasure(chaosMeasurePerLine, numberOfLines)
	fdkaacEncCalcSfbTonality(spectrum, sfbMaxScaleSpec, chaosMeasurePerLine, sfbTonality, sfbCnt, sfbOffset, sfbEnergyLD64)
}

func fdkaacEncCalculateChaosMeasurePeakFast(mdctData []FixpDBL, numberOfLines int, chaosMeasure []FixpDBL) {
	left0Div2 := fdkaacEncChaosAbsTap(mdctData[0]) >> 1
	left1Div2 := fdkaacEncChaosAbsTap(mdctData[1]) >> 1
	center0 := fdkaacEncChaosAbsTap(mdctData[2])
	center1 := fdkaacEncChaosAbsTap(mdctData[3])

	for j := 2; j < numberOfLines-2; j += 2 {
		right0 := fdkaacEncChaosAbsTap(mdctData[j+2])
		tmp0 := left0Div2 + (right0 >> 1)
		right1 := fdkaacEncChaosAbsTap(mdctData[j+3])
		tmp1 := left1Div2 + (right1 >> 1)

		if tmp0 < center0 {
			leadingBits := FixNormZD(center0) - 1
			tmp0 = schurDiv(tmp0<<uint(leadingBits), center0<<uint(leadingBits), 8)
			tmp0 = FMultDD(tmp0, tmp0)
		} else {
			tmp0 = MaxValDBL
		}
		chaosMeasure[j] = tmp0
		left0Div2 = center0 >> 1
		center0 = right0

		if tmp1 < center1 {
			leadingBits := FixNormZD(center1) - 1
			tmp1 = schurDiv(tmp1<<uint(leadingBits), center1<<uint(leadingBits), 8)
			tmp1 = FMultDD(tmp1, tmp1)
		} else {
			tmp1 = MaxValDBL
		}
		left1Div2 = center1 >> 1
		center1 = right1
		chaosMeasure[j+1] = tmp1
	}

	chaosMeasure[0] = chaosMeasure[2]
	chaosMeasure[1] = chaosMeasure[2]
	for i := numberOfLines - 3; i < numberOfLines; i++ {
		chaosMeasure[i] = tonalityHalf
	}
}

func fdkaacEncSmoothChaosMeasure(chaosMeasure []FixpDBL, numberOfLines int) {
	left := chaosMeasure[0]
	j := 1
	for ; j < numberOfLines-1; j += 2 {
		right := chaosMeasure[j]
		right -= right >> 2
		left = right + (left >> 2)
		chaosMeasure[j] = left

		right = chaosMeasure[j+1]
		right -= right >> 2
		left = right + (left >> 2)
		chaosMeasure[j+1] = left
	}
	if j == numberOfLines-1 {
		right := chaosMeasure[j]
		right -= right >> 2
		left = right + (left >> 2)
		chaosMeasure[j] = left
	}
}

func fdkaacEncCalcSfbTonality(
	spectrum []FixpDBL,
	sfbMaxScaleSpec []int,
	chaosMeasure []FixpDBL,
	sfbTonality []FixpSGL,
	sfbCnt int,
	sfbOffset []int,
	sfbEnergyLD64 []FixpDBL,
) {
	for i := 0; i < sfbCnt; i++ {
		shiftBits := maxInt(0, sfbMaxScaleSpec[i]-4)
		chaosMeasureSfb := FixpDBL(0)
		for line := sfbOffset[i]; line < sfbOffset[i+1]; line++ {
			tmp := spectrum[line] << uint(shiftBits)
			lineNrg := FMultDiv2DD(tmp, tmp)
			chaosMeasureSfb = FMultAddDiv2DD(chaosMeasureSfb, lineNrg, chaosMeasure[line])
		}

		if chaosMeasureSfb == 0 {
			sfbTonality[i] = MaxValSGL
			continue
		}

		chaosMeasureSfbLD64 := CalcLdData(chaosMeasureSfb) - sfbEnergyLD64[i]
		chaosMeasureSfbLD64 += tonalityLdConvtone - (FixpDBL(shiftBits) << (DfractBits - 6))
		if chaosMeasureSfbLD64 > tonalityChaosThreshold {
			if chaosMeasureSfbLD64 <= 0 {
				sfbTonality[i] = FXDBL2FXSGL(FMultDiv2DD(chaosMeasureSfbLD64, tonalityNormLog) << 7)
			} else {
				sfbTonality[i] = 0
			}
		} else {
			sfbTonality[i] = MaxValSGL
		}
	}
}

func fdkaacEncChaosAbsTap(x FixpDBL) FixpDBL {
	return x ^ (x >> (DfractBits - 1))
}

func checkChaosMeasureLineInputs(mdctData []FixpDBL, numberOfLines int, chaosMeasure []FixpDBL) {
	if numberOfLines < 4 || numberOfLines > maxSpectralLines || numberOfLines&1 != 0 {
		panic("fdkaac: invalid chaos measure line count")
	}
	if len(mdctData) < numberOfLines {
		panic("fdkaac: short chaos measure spectrum")
	}
	if len(chaosMeasure) < numberOfLines {
		panic("fdkaac: short chaos measure output")
	}
}

func checkTonalityMetadata(spectrum []FixpDBL, sfbCnt int, sfbOffset []int) int {
	if sfbCnt <= 0 {
		panic("fdkaac: empty tonality bands")
	}
	checkBandEnergySpectrum(spectrum, sfbOffset, sfbCnt)
	numberOfLines := sfbOffset[sfbCnt]
	if numberOfLines < 4 || numberOfLines > maxSpectralLines {
		panic("fdkaac: invalid tonality line count")
	}
	return numberOfLines
}

func checkTonalityBandInputs(
	sfbMaxScaleSpec []int,
	sfbEnergyLD64 []FixpDBL,
	sfbTonality []FixpSGL,
	sfbCnt int,
	scratch *TonalityScratch,
) {
	checkBandEnergyScales(sfbMaxScaleSpec, sfbCnt)
	checkBandEnergyOutput(sfbEnergyLD64, sfbCnt)
	if len(sfbTonality) < sfbCnt {
		panic("fdkaac: short tonality output")
	}
	if scratch == nil {
		panic("fdkaac: missing tonality scratch")
	}
}
