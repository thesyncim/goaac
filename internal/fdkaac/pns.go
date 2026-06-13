package fdkaac

const (
	pnsCodeBookLav    = 60
	pnsFuzzyHalf      = FixpSGL(0x4000)
	pnsThr15Ld        = FixpDBL(0x012b8034)
	pnsHalfOver64     = FixpDBL(0x01000000)
	pnsOneOver64      = FixpDBL(0x02000000)
	pnsMinus32Over64  = FixpDBL(-1073741824)
	pnsPowerLowLimit  = FixpDBL(0x3f5c28f6)
	pnsPowerHighLimit = FixpDBL(0x40a3d70a)
	pnsTonLowLimit    = FixpDBL(0x3999999a)
	pnsTonHighLimit   = FixpDBL(0x46666666)
)

func FDKaacEncNoiseDetect(
	mdctSpectrum []FixpDBL,
	sfbMaxScaleSpec []int,
	sfbActive int,
	sfbOffset []int,
	noiseFuzzyMeasure []FixpSGL,
	np *NoiseParams,
	sfbTonality []FixpSGL,
) {
	checkNoiseDetectInputs(mdctSpectrum, sfbMaxScaleSpec, sfbActive, sfbOffset, noiseFuzzyMeasure, np, sfbTonality)

	for sfb := 0; sfb < sfbActive; sfb++ {
		fuzzyTotal := MaxValSGL
		sfbWidth := sfbOffset[sfb+1] - sfbOffset[sfb]

		if sfb < np.StartSfb || sfbWidth < np.MinSfbWidth {
			noiseFuzzyMeasure[sfb] = 0
			continue
		}

		if np.DetectionAlgorithmFlags&pnsUsePowerDistribution != 0 && fuzzyTotal > pnsFuzzyHalf {
			leadingBits := maxInt(0, sfbMaxScaleSpec[sfb]-3)
			k := sfbWidth >> 2
			fhelp1 := FixpDBL(0)
			fhelp2 := FixpDBL(0)
			fhelp3 := FixpDBL(0)
			fhelp4 := FixpDBL(0)
			for i := sfbOffset[sfb]; i < sfbOffset[sfb]+k; i++ {
				fhelp1 = FPow2AddDiv2D(fhelp1, mdctSpectrum[i]<<uint(leadingBits))
				fhelp2 = FPow2AddDiv2D(fhelp2, mdctSpectrum[i+k]<<uint(leadingBits))
				fhelp3 = FPow2AddDiv2D(fhelp3, mdctSpectrum[i+2*k]<<uint(leadingBits))
				fhelp4 = FPow2AddDiv2D(fhelp4, mdctSpectrum[i+3*k]<<uint(leadingBits))
			}

			maxVal := maxFixpDBL(maxFixpDBL(fhelp1, fhelp2), maxFixpDBL(fhelp3, fhelp4))
			minVal := minFixpDBL(minFixpDBL(fhelp1, fhelp2), minFixpDBL(fhelp3, fhelp4))

			leadingBits = CountLeadingBits(maxVal)
			testVal := maxVal << uint(leadingBits)
			refVal := minVal << uint(leadingBits)
			testVal = FMultDiv2DS(testVal, np.PowDistPSDcurve[sfb])

			fuzzy := fdkaacEncFuzzyIsSmaller(testVal, refVal, pnsPowerLowLimit, pnsPowerHighLimit)
			fuzzyTotal = minFixpSGL(fuzzyTotal, fuzzy)
		}

		if np.DetectionAlgorithmFlags&pnsUsePsychTonality != 0 && fuzzyTotal > pnsFuzzyHalf {
			testVal := FXSGL2FXDBL(sfbTonality[sfb]) >> 1
			refVal := np.RefTonality
			fuzzy := fdkaacEncFuzzyIsSmaller(testVal, refVal, pnsTonLowLimit, pnsTonHighLimit)
			fuzzyTotal = minFixpSGL(fuzzyTotal, fuzzy)
		}

		noiseFuzzyMeasure[sfb] = fuzzyTotal
	}
}

func FDKaacEncPnsDetect(
	pnsConf *PNSConfig,
	pnsData *PNSData,
	lastWindowSequence int,
	sfbActive int,
	maxSfbPerGroup int,
	sfbThresholdLdData []FixpDBL,
	sfbOffset []int,
	mdctSpectrum []FixpDBL,
	sfbMaxScaleSpec []int,
	sfbTonality []FixpSGL,
	tnsOrder int,
	tnsPredictionGain int,
	tnsActive int,
	sfbEnergyLdData []FixpDBL,
	noiseNrg []int,
) {
	checkPnsDetectResetInputs(pnsConf, pnsData, sfbActive, noiseNrg)

	for i := range pnsData.PNSFlag {
		pnsData.PNSFlag[i] = 0
	}
	for i := 0; i < maxGroupedSFB; i++ {
		noiseNrg[i] = noNoisePNS
	}

	if pnsConf.UsePns == 0 {
		return
	}
	if pnsConf.NP.DetectionAlgorithmFlags&pnsIsLowComplexity != 0 && lastWindowSequence == ShortWindow {
		return
	}
	if pnsConf.NP.DetectionAlgorithmFlags&pnsIsLowComplexity == 0 &&
		pnsConf.NP.DetectionAlgorithmFlags&pnsJustLongWindow != 0 &&
		lastWindowSequence != LongWindow {
		return
	}

	checkPnsDetectActiveInputs(
		lastWindowSequence, sfbActive, maxSfbPerGroup, sfbThresholdLdData,
		sfbOffset, mdctSpectrum, sfbMaxScaleSpec, sfbTonality, sfbEnergyLdData,
	)

	fdkaacEncNoiseDetection(
		pnsConf, pnsData, sfbActive, sfbOffset, tnsOrder, tnsPredictionGain,
		tnsActive, mdctSpectrum, sfbMaxScaleSpec, sfbTonality,
	)

	startNoiseSfb := pnsConf.NP.StartSfb
	for sfb := 0; sfb < sfbActive; sfb++ {
		if sfb < startNoiseSfb {
			pnsData.PNSFlag[sfb] = 0
			continue
		}
		if pnsData.NoiseFuzzyMeasure[sfb] > pnsFuzzyHalf &&
			sfbThresholdLdData[sfb]+pnsThr15Ld < sfbEnergyLdData[sfb] {
			pnsData.PNSFlag[sfb] = 1
		} else {
			pnsData.PNSFlag[sfb] = 0
		}
	}

	if pnsData.NoiseFuzzyMeasure[0] > pnsFuzzyHalf && pnsData.PNSFlag[1] != 0 {
		pnsData.PNSFlag[0] = 1
	}
	for sfb := 1; sfb < maxSfbPerGroup-1; sfb++ {
		if pnsData.NoiseFuzzyMeasure[sfb] > pnsConf.NP.GapFillThr &&
			pnsData.PNSFlag[sfb-1] != 0 && pnsData.PNSFlag[sfb+1] != 0 {
			pnsData.PNSFlag[sfb] = 1
		}
	}
	if maxSfbPerGroup > 0 {
		if pnsData.NoiseFuzzyMeasure[maxSfbPerGroup-1] > pnsConf.NP.GapFillThr &&
			pnsData.PNSFlag[maxSfbPerGroup-2] != 0 {
			pnsData.PNSFlag[maxSfbPerGroup-1] = 1
		}
		if pnsData.PNSFlag[maxSfbPerGroup-2] == 0 {
			pnsData.PNSFlag[maxSfbPerGroup-1] = 0
		}
	}
	if pnsData.PNSFlag[1] == 0 {
		pnsData.PNSFlag[0] = 0
	}
	for sfb := 1; sfb < maxSfbPerGroup-1; sfb++ {
		if pnsData.PNSFlag[sfb-1] == 0 && pnsData.PNSFlag[sfb+1] == 0 {
			pnsData.PNSFlag[sfb] = 0
		}
	}

	fdkaacEncCalcNoiseNrgs(sfbActive, pnsData.PNSFlag[:], sfbEnergyLdData, noiseNrg)
}

func FDKaacEncCodePnsChannel(
	sfbActive int,
	pnsConf *PNSConfig,
	pnsFlag []int,
	sfbEnergyLdData []FixpDBL,
	noiseNrg []int,
	sfbThresholdLdData []FixpDBL,
) {
	checkCodePnsChannelInputs(sfbActive, pnsConf, pnsFlag, sfbEnergyLdData, noiseNrg, sfbThresholdLdData)

	if pnsConf.UsePns == 0 {
		for sfb := 0; sfb < sfbActive; sfb++ {
			noiseNrg[sfb] = noNoisePNS
		}
		return
	}

	lastINoiseEnergy := 0
	firstPNSBand := true
	for sfb := 0; sfb < sfbActive; sfb++ {
		if pnsFlag[sfb] != 0 {
			if noiseNrg[sfb] != noNoisePNS {
				sfbThresholdLdData[sfb] = sfbEnergyLdData[sfb] + pnsOneOver64
			}
			if !firstPNSBand {
				deltaINoiseEnergy := noiseNrg[sfb] - lastINoiseEnergy
				if deltaINoiseEnergy > pnsCodeBookLav {
					noiseNrg[sfb] -= deltaINoiseEnergy - pnsCodeBookLav
				} else if deltaINoiseEnergy < -pnsCodeBookLav {
					noiseNrg[sfb] -= deltaINoiseEnergy + pnsCodeBookLav
				}
			} else {
				firstPNSBand = false
			}
			lastINoiseEnergy = noiseNrg[sfb]
		} else {
			noiseNrg[sfb] = noNoisePNS
		}
	}
}

func FDKaacEncPreProcessPnsChannelPair(
	sfbActive int,
	sfbEnergyLeft []FixpDBL,
	sfbEnergyRight []FixpDBL,
	sfbEnergyLeftLD []FixpDBL,
	sfbEnergyRightLD []FixpDBL,
	sfbEnergyMid []FixpDBL,
	pnsConf *PNSConfig,
	pnsDataLeft *PNSData,
	pnsDataRight *PNSData,
) {
	if pnsConf == nil {
		panic("fdkaac: nil pns configuration")
	}
	if pnsConf.UsePns == 0 {
		return
	}
	checkPnsPairInputs(sfbActive, sfbEnergyLeft, sfbEnergyRight, sfbEnergyLeftLD, sfbEnergyRightLD, sfbEnergyMid, pnsDataLeft, pnsDataRight)

	for sfb := 0; sfb < sfbActive; sfb++ {
		quot := (sfbEnergyLeftLD[sfb] >> 1) + (sfbEnergyRightLD[sfb] >> 1)
		ccf := FixpDBL(0)
		if quot >= pnsMinus32Over64 {
			accu := sfbEnergyMid[sfb] - (((sfbEnergyLeft[sfb] >> 1) + (sfbEnergyRight[sfb] >> 1)) >> 1)
			sign := accu < 0
			accu = fixpAbsDBL(accu)
			ccf = CalcLdData(accu) + pnsOneOver64 - quot
			if ccf >= 0 {
				ccf = MaxValDBL
			} else if sign {
				ccf = -CalcInvLdData(ccf)
			} else {
				ccf = CalcInvLdData(ccf)
			}
		}
		pnsDataLeft.NoiseEnergyCorrelation[sfb] = ccf
		pnsDataRight.NoiseEnergyCorrelation[sfb] = ccf
	}
}

func FDKaacEncPostProcessPnsChannelPair(
	sfbActive int,
	pnsConf *PNSConfig,
	pnsDataLeft *PNSData,
	pnsDataRight *PNSData,
	msMask []int,
	msDigest *int,
) {
	if pnsConf == nil {
		panic("fdkaac: nil pns configuration")
	}
	if pnsConf.UsePns == 0 {
		return
	}
	checkPnsPostProcessInputs(sfbActive, pnsDataLeft, pnsDataRight, msMask, msDigest)

	for sfb := 0; sfb < sfbActive; sfb++ {
		if msMask[sfb] != 0 {
			if pnsDataLeft.PNSFlag[sfb] != 0 && pnsDataRight.PNSFlag[sfb] != 0 {
				if pnsDataLeft.NoiseEnergyCorrelation[sfb] <= pnsConf.NoiseCorrelationThresh {
					msMask[sfb] = 0
					*msDigest = MsMaskSome
				}
			} else {
				pnsDataLeft.PNSFlag[sfb] = 0
				pnsDataRight.PNSFlag[sfb] = 0
			}
		}
		if pnsDataLeft.PNSFlag[sfb] != 0 && pnsDataRight.PNSFlag[sfb] != 0 {
			if pnsDataLeft.NoiseEnergyCorrelation[sfb] > pnsConf.NoiseCorrelationThresh {
				msMask[sfb] = 1
				*msDigest = MsMaskSome
			}
		}
	}
}

func fdkaacEncNoiseDetection(
	pnsConf *PNSConfig,
	pnsData *PNSData,
	sfbActive int,
	sfbOffset []int,
	tnsOrder int,
	tnsPredictionGain int,
	tnsActive int,
	mdctSpectrum []FixpDBL,
	sfbMaxScaleSpec []int,
	sfbTonality []FixpSGL,
) {
	condition := true
	if pnsConf.NP.DetectionAlgorithmFlags&pnsIsLowComplexity == 0 {
		condition = tnsOrder > 3
	}
	if pnsConf.NP.DetectionAlgorithmFlags&pnsUseTnsGainThreshold != 0 &&
		tnsPredictionGain >= pnsConf.NP.TnsGainThreshold && condition &&
		!(pnsConf.NP.DetectionAlgorithmFlags&pnsUseTnsPNS != 0 &&
			tnsPredictionGain >= pnsConf.NP.TnsPNSGainThreshold &&
			tnsActive != 0) {
		for sfb := 0; sfb < sfbActive; sfb++ {
			pnsData.NoiseFuzzyMeasure[sfb] = 0
		}
		return
	}
	FDKaacEncNoiseDetect(mdctSpectrum, sfbMaxScaleSpec, sfbActive, sfbOffset, pnsData.NoiseFuzzyMeasure[:], &pnsConf.NP, sfbTonality)
}

func fdkaacEncCalcNoiseNrgs(sfbActive int, pnsFlag []int, sfbEnergyLdData []FixpDBL, noiseNrg []int) {
	tmp := (-logNormPCM) << 2
	for sfb := 0; sfb < sfbActive; sfb++ {
		if pnsFlag[sfb] != 0 {
			nrg := int((-sfbEnergyLdData[sfb] + pnsHalfOver64) >> (DfractBits - 1 - 7))
			noiseNrg[sfb] = tmp - nrg
		}
	}
}

func fdkaacEncFuzzyIsSmaller(testVal FixpDBL, refVal FixpDBL, loLim FixpDBL, hiLim FixpDBL) FixpSGL {
	if refVal <= 0 {
		return 0
	}
	if testVal >= FMultDD((hiLim>>1)+(loLim>>1), refVal) {
		return 0
	}
	return MaxValSGL
}

func minFixpSGL(a, b FixpSGL) FixpSGL {
	if a < b {
		return a
	}
	return b
}

func checkNoiseDetectInputs(
	mdctSpectrum []FixpDBL,
	sfbMaxScaleSpec []int,
	sfbActive int,
	sfbOffset []int,
	noiseFuzzyMeasure []FixpSGL,
	np *NoiseParams,
	sfbTonality []FixpSGL,
) {
	if np == nil {
		panic("fdkaac: nil pns noise params")
	}
	if sfbActive <= 0 || sfbActive > maxGroupedSFB {
		panic("fdkaac: invalid pns active band count")
	}
	checkBandEnergySpectrum(mdctSpectrum, sfbOffset, sfbActive)
	checkBandEnergyScales(sfbMaxScaleSpec, sfbActive)
	if len(noiseFuzzyMeasure) < sfbActive {
		panic("fdkaac: short pns fuzzy output")
	}
	if len(sfbTonality) < sfbActive {
		panic("fdkaac: short pns tonality")
	}
	if np.StartSfb < 0 || np.StartSfb > sfbActive || np.MinSfbWidth < 0 {
		panic("fdkaac: invalid pns noise params")
	}
}

func checkPnsDetectResetInputs(pnsConf *PNSConfig, pnsData *PNSData, sfbActive int, noiseNrg []int) {
	if pnsConf == nil {
		panic("fdkaac: nil pns configuration")
	}
	if pnsData == nil {
		panic("fdkaac: nil pns data")
	}
	if sfbActive <= 0 || sfbActive > maxGroupedSFB {
		panic("fdkaac: invalid pns active band count")
	}
	if len(noiseNrg) < maxGroupedSFB {
		panic("fdkaac: short pns noise energy")
	}
}

func checkPnsDetectActiveInputs(
	lastWindowSequence int,
	sfbActive int,
	maxSfbPerGroup int,
	sfbThresholdLdData []FixpDBL,
	sfbOffset []int,
	mdctSpectrum []FixpDBL,
	sfbMaxScaleSpec []int,
	sfbTonality []FixpSGL,
	sfbEnergyLdData []FixpDBL,
) {
	if !validSFBWindowSequence(lastWindowSequence) {
		panic("fdkaac: invalid pns window sequence")
	}
	if maxSfbPerGroup < 2 || maxSfbPerGroup > sfbActive {
		panic("fdkaac: invalid pns group width")
	}
	checkBandEnergySpectrum(mdctSpectrum, sfbOffset, sfbActive)
	checkBandEnergyScales(sfbMaxScaleSpec, sfbActive)
	if len(sfbTonality) < sfbActive {
		panic("fdkaac: short pns tonality")
	}
	if len(sfbThresholdLdData) < sfbActive {
		panic("fdkaac: short pns threshold ld-data")
	}
	if len(sfbEnergyLdData) < sfbActive {
		panic("fdkaac: short pns energy ld-data")
	}
}

func checkCodePnsChannelInputs(
	sfbActive int,
	pnsConf *PNSConfig,
	pnsFlag []int,
	sfbEnergyLdData []FixpDBL,
	noiseNrg []int,
	sfbThresholdLdData []FixpDBL,
) {
	if pnsConf == nil {
		panic("fdkaac: nil pns configuration")
	}
	if sfbActive < 0 || sfbActive > maxGroupedSFB {
		panic("fdkaac: invalid pns active band count")
	}
	if len(noiseNrg) < sfbActive {
		panic("fdkaac: short pns noise energy")
	}
	if pnsConf.UsePns == 0 {
		return
	}
	if len(pnsFlag) < sfbActive {
		panic("fdkaac: short pns flags")
	}
	if len(sfbEnergyLdData) < sfbActive || len(sfbThresholdLdData) < sfbActive {
		panic("fdkaac: short pns ld-data")
	}
}

func checkPnsPairInputs(
	sfbActive int,
	sfbEnergyLeft []FixpDBL,
	sfbEnergyRight []FixpDBL,
	sfbEnergyLeftLD []FixpDBL,
	sfbEnergyRightLD []FixpDBL,
	sfbEnergyMid []FixpDBL,
	pnsDataLeft *PNSData,
	pnsDataRight *PNSData,
) {
	if sfbActive < 0 || sfbActive > maxGroupedSFB {
		panic("fdkaac: invalid pns active band count")
	}
	if pnsDataLeft == nil || pnsDataRight == nil {
		panic("fdkaac: nil pns data")
	}
	if len(sfbEnergyLeft) < sfbActive || len(sfbEnergyRight) < sfbActive || len(sfbEnergyMid) < sfbActive {
		panic("fdkaac: short pns energy")
	}
	if len(sfbEnergyLeftLD) < sfbActive || len(sfbEnergyRightLD) < sfbActive {
		panic("fdkaac: short pns energy ld-data")
	}
}

func checkPnsPostProcessInputs(sfbActive int, pnsDataLeft *PNSData, pnsDataRight *PNSData, msMask []int, msDigest *int) {
	if sfbActive < 0 || sfbActive > maxGroupedSFB {
		panic("fdkaac: invalid pns active band count")
	}
	if pnsDataLeft == nil || pnsDataRight == nil {
		panic("fdkaac: nil pns data")
	}
	if len(msMask) < sfbActive {
		panic("fdkaac: short pns ms mask")
	}
	if msDigest == nil {
		panic("fdkaac: nil pns ms digest")
	}
}
