package fdkaac

type PsyMainScratch struct {
	TimeSignal [2][maxSpectralLines]int16
	Tonality   [2][maxSFBLong]FixpSGL
	TNS        PsyTnsTonalityScratch
	Post       PsyPostTnsScratch
}

func FDKaacEncPsyMain(
	channels int,
	psyElement *PsyElement,
	psyDynamic *PsyDynamic,
	psyConf *[2]PsyConfiguration,
	psyOutElement *PsyOutElement,
	input []int16,
	inputBufSize int,
	chIdx []int,
	totalChannels int,
	scratch *PsyMainScratch,
) int {
	checkPsyMainInputs(channels, psyElement, psyDynamic, psyConf, psyOutElement, input, inputBufSize, chIdx, totalChannels, scratch)
	if psyConf[0].Filterbank != FilterbankLC || psyConf[1].Filterbank != FilterbankLC {
		return AACEncUnsupportedFilterbank
	}

	const commonWindow = 1
	nTimeSamples := psyConf[0].GranuleLength
	blockSwitchingOffset := nTimeSamples + (9*nTimeSamples)/(2*transFac)
	psyOutElement.CommonWindow = commonWindow

	var psyStatic [2]*PsyStatic
	var psyData [2]*PsyData
	var tnsData [2]*TNSData
	var pnsData [2]*PNSData
	var outChannel [2]*PsyOutChannel
	var conf [2]*PsyConfiguration
	var windowLength [2]int
	var nWindows [2]int
	var isShortWindow [2]bool
	zeroSpec := true

	for ch := 0; ch < channels; ch++ {
		psyStatic[ch] = psyElement.PsyStatic[ch]
		psyData[ch] = &psyDynamic.PsyData[ch]
		tnsData[ch] = &psyDynamic.TNSData[ch]
		pnsData[ch] = &psyDynamic.PNSData[ch]
		outChannel[ch] = psyOutElement.PsyOutChannel[ch]
		psyData[ch].MdctSpectrum = outChannel[ch].MdctSpectrum[:nTimeSamples]
	}

	for ch := 0; ch < channels; ch++ {
		base := chIdx[ch] * inputBufSize
		timeSignal := scratch.TimeSignal[ch][:]
		copy(timeSignal[:nTimeSamples], input[base:base+nTimeSamples])

		FDKaacEncBlockSwitching(
			&psyStatic[ch].BlockSwitchingControl,
			nTimeSamples,
			psyStatic[ch].IsLFE != 0,
			timeSignal[:nTimeSamples],
		)

		visible := 2*nTimeSamples - blockSwitchingOffset
		copy(psyStatic[ch].PsyInputBuffer[blockSwitchingOffset:blockSwitchingOffset+visible], timeSignal[:visible])
	}

	var rightBlock *BlockSwitchingControl
	if channels > 1 {
		rightBlock = &psyStatic[1].BlockSwitchingControl
	}
	if FDKaacEncSyncBlockSwitching(&psyStatic[0].BlockSwitchingControl, rightBlock, channels, true) != 0 {
		return AACEncUnsupportedAOT
	}

	for ch := 0; ch < channels; ch++ {
		isShortWindow[ch] = psyStatic[ch].BlockSwitchingControl.LastWindowSequence == ShortWindow
		if isShortWindow[ch] {
			conf[ch] = &psyConf[1]
			windowLength[ch] = nTimeSamples / transFac
			nWindows[ch] = transFac
		} else {
			conf[ch] = &psyConf[0]
			windowLength[ch] = nTimeSamples
			nWindows[ch] = 1
		}
	}

	for ch := 0; ch < channels; ch++ {
		if psyStatic[ch].IsLFE != 0 {
			psyData[ch].SfbActive = conf[ch].SfbActiveLFE
			psyData[ch].LowpassLine = conf[ch].LowpassLineLFE
		} else {
			psyData[ch].SfbActive = conf[ch].SfbActive
			psyData[ch].LowpassLine = conf[ch].LowpassLine
		}

		mdctSpectrumE := 0
		if FDKaacEncTransformReal(
			psyStatic[ch].PsyInputBuffer[:],
			psyData[ch].MdctSpectrum,
			psyStatic[ch].BlockSwitchingControl.LastWindowSequence,
			psyStatic[ch].BlockSwitchingControl.WindowShape,
			&psyStatic[ch].BlockSwitchingControl.LastWindowShape,
			&psyStatic[ch].MDCT,
			nTimeSamples,
			&mdctSpectrumE,
			conf[ch].Filterbank,
		) != 0 {
			return AACEncUnsupportedFilterbank
		}

		for w := 0; w < nWindows[ch]; w++ {
			wOffset := w * windowLength[ch]
			clear(psyData[ch].MdctSpectrum[wOffset+psyData[ch].LowpassLine : wOffset+windowLength[ch]])
			for line := 0; line < psyData[ch].LowpassLine && zeroSpec; line++ {
				if psyData[ch].MdctSpectrum[wOffset+line] != 0 {
					zeroSpec = false
					break
				}
			}
		}

		psyData[ch].MdctScale = mdctSpectrumE

		base := chIdx[ch] * inputBufSize
		copy(psyStatic[ch].PsyInputBuffer[:nTimeSamples], psyStatic[ch].PsyInputBuffer[nTimeSamples:2*nTimeSamples])
		remainingStart := 2*nTimeSamples - blockSwitchingOffset
		remaining := blockSwitchingOffset - nTimeSamples
		copy(psyStatic[ch].PsyInputBuffer[nTimeSamples:nTimeSamples+remaining], input[base+remainingStart:base+remainingStart+remaining])
	}

	if zeroSpec {
		psyMainClearZeroEnergy(channels, psyData[:], isShortWindow, nWindows, conf[:])
	} else {
		psyMainPrepareEnergy(channels, psyData[:], isShortWindow, windowLength, nWindows, conf[:])
	}

	clear(scratch.Tonality[:])
	if rc := FDKaacEncPsyAdvanceTnsAndTonality(
		channels,
		psyStatic[:],
		psyData[:],
		tnsData[:],
		conf[:],
		outChannel[:],
		&scratch.Tonality,
		&scratch.TNS,
	); rc != AACEncOK {
		return rc
	}

	return FDKaacEncPsyPostTnsToolsAndOutput(
		channels,
		psyStatic[:],
		psyData[:],
		tnsData[:],
		pnsData[:],
		&psyConf[0],
		&psyConf[1],
		psyOutElement,
		&scratch.Tonality,
		&scratch.Post,
	)
}

func psyMainPrepareEnergy(
	channels int,
	psyData []*PsyData,
	isShortWindow [2]bool,
	windowLength [2]int,
	nWindows [2]int,
	psyConf []*PsyConfiguration,
) {
	minSpecShift := DfractBits - 2
	nrgShift := DfractBits - 2
	finalShift := DfractBits - 2
	maxNrg := FixpDBL(0)

	for ch := 0; ch < channels; ch++ {
		for w := 0; w < nWindows[ch]; w++ {
			wOffset := w * windowLength[ch]
			scales := psyMainScaleSpec(psyData[ch], isShortWindow[ch], w)
			FDKaacEncCalcSfbMaxScaleSpec(
				psyData[ch].MdctSpectrum[wOffset:],
				psyConf[ch].SfbOffset[:],
				scales,
				psyData[ch].SfbActive,
			)
			for sfb := 0; sfb < psyData[ch].SfbActive; sfb++ {
				minSpecShift = minInt(minSpecShift, scales[sfb])
			}
		}
	}

	for ch := 0; ch < channels; ch++ {
		for w := 0; w < nWindows[ch]; w++ {
			wOffset := w * windowLength[ch]
			currNrg := FDKaacEncCheckBandEnergyOptim(
				psyData[ch].MdctSpectrum[wOffset:],
				psyMainScaleSpec(psyData[ch], isShortWindow[ch], w),
				psyConf[ch].SfbOffset[:],
				psyData[ch].SfbActive,
				psyMainEnergy(psyData[ch], isShortWindow[ch], w),
				psyMainEnergyLd(psyData[ch], isShortWindow[ch], w),
				minSpecShift-4,
			)
			if currNrg > maxNrg {
				maxNrg = currNrg
			}
		}
	}

	if maxNrg != 0 {
		nrgShift = (CountLeadingBits(maxNrg) >> 1) + (minSpecShift - 4)
	}
	if isShortWindow[0] {
		nrgShift--
	}
	finalShift = minInt(minSpecShift, nrgShift)
	if finalShift > psyData[0].MdctScale+3 {
		finalShift = psyData[0].MdctScale + 3
	}
	if finalShift < 0 {
		panic("fdkaac: negative psy main final shift")
	}

	ldShift := FixpDBL(finalShift) * ldDataStep2Over64
	for ch := 0; ch < channels; ch++ {
		for w := 0; w < nWindows[ch]; w++ {
			scales := psyMainScaleSpec(psyData[ch], isShortWindow[ch], w)
			energy := psyMainEnergy(psyData[ch], isShortWindow[ch], w)
			energyLd := psyMainEnergyLd(psyData[ch], isShortWindow[ch], w)
			threshold := psyMainThreshold(psyData[ch], isShortWindow[ch], w)
			for sfb := 0; sfb < psyData[ch].SfbActive; sfb++ {
				scale := maxInt(0, scales[sfb]-4)
				scale = minInt((scale-finalShift)<<1, DfractBits-1)
				if scale >= 0 {
					energy[sfb] >>= uint(scale)
				} else {
					energy[sfb] <<= uint(-scale)
				}
				threshold[sfb] = FMultDD(energy[sfb], cRatio)
				energyLd[sfb] += ldShift
			}
		}
	}

	if finalShift == 0 {
		return
	}
	for ch := 0; ch < channels; ch++ {
		for w := 0; w < nWindows[ch]; w++ {
			wOffset := w * windowLength[ch]
			for line := 0; line < psyData[ch].LowpassLine; line++ {
				psyData[ch].MdctSpectrum[wOffset+line] <<= uint(finalShift)
			}
			scales := psyMainScaleSpec(psyData[ch], isShortWindow[ch], w)
			for sfb := 0; sfb < psyData[ch].SfbActive; sfb++ {
				scales[sfb] -= finalShift
			}
		}
		psyData[ch].MdctScale -= finalShift
	}
}

func psyMainClearZeroEnergy(
	channels int,
	psyData []*PsyData,
	isShortWindow [2]bool,
	nWindows [2]int,
	psyConf []*PsyConfiguration,
) {
	_ = psyConf
	for ch := 0; ch < channels; ch++ {
		psyData[ch].MdctScale = 0
		for w := 0; w < nWindows[ch]; w++ {
			scales := psyMainScaleSpec(psyData[ch], isShortWindow[ch], w)
			energy := psyMainEnergy(psyData[ch], isShortWindow[ch], w)
			energyLd := psyMainEnergyLd(psyData[ch], isShortWindow[ch], w)
			threshold := psyMainThreshold(psyData[ch], isShortWindow[ch], w)
			for sfb := 0; sfb < psyData[ch].SfbActive; sfb++ {
				scales[sfb] = 0
				energy[sfb] = 0
				energyLd[sfb] = ldDataMinusOne
				threshold[sfb] = 0
			}
		}
	}
}

func psyMainScaleSpec(data *PsyData, isShort bool, window int) []int {
	if isShort {
		return data.SfbMaxScaleSpec.Short[window][:]
	}
	return data.SfbMaxScaleSpec.Long[:]
}

func psyMainEnergy(data *PsyData, isShort bool, window int) []FixpDBL {
	if isShort {
		return data.SfbEnergy.Short[window][:]
	}
	return data.SfbEnergy.Long[:]
}

func psyMainEnergyLd(data *PsyData, isShort bool, window int) []FixpDBL {
	if isShort {
		return data.SfbEnergyLdData.Short[window][:]
	}
	return data.SfbEnergyLdData.Long[:]
}

func psyMainThreshold(data *PsyData, isShort bool, window int) []FixpDBL {
	if isShort {
		return data.SfbThreshold.Short[window][:]
	}
	return data.SfbThreshold.Long[:]
}

func checkPsyMainInputs(
	channels int,
	psyElement *PsyElement,
	psyDynamic *PsyDynamic,
	psyConf *[2]PsyConfiguration,
	psyOutElement *PsyOutElement,
	input []int16,
	inputBufSize int,
	chIdx []int,
	totalChannels int,
	scratch *PsyMainScratch,
) {
	if channels <= 0 || channels > 2 {
		panic("fdkaac: invalid psy main channel count")
	}
	if psyElement == nil || psyDynamic == nil || psyConf == nil || psyOutElement == nil || scratch == nil {
		panic("fdkaac: nil psy main state")
	}
	if len(chIdx) < channels {
		panic("fdkaac: short psy main channel index list")
	}
	if totalChannels < channels {
		panic("fdkaac: invalid psy main total channel count")
	}
	nTimeSamples := psyConf[0].GranuleLength
	if nTimeSamples <= 0 || nTimeSamples > maxSpectralLines || nTimeSamples%transFac != 0 {
		panic("fdkaac: invalid psy main granule length")
	}
	if psyConf[1].GranuleLength != nTimeSamples {
		panic("fdkaac: mismatched psy main short granule length")
	}
	if inputBufSize < nTimeSamples {
		panic("fdkaac: short psy main input stride")
	}
	blockSwitchingOffset := nTimeSamples + (9*nTimeSamples)/(2*transFac)
	if blockSwitchingOffset < nTimeSamples || blockSwitchingOffset > 2*nTimeSamples {
		panic("fdkaac: invalid psy main block switching offset")
	}
	for ch := 0; ch < channels; ch++ {
		if psyElement.PsyStatic[ch] == nil || psyOutElement.PsyOutChannel[ch] == nil {
			panic("fdkaac: nil psy main channel state")
		}
		if len(psyOutElement.PsyOutChannel[ch].MdctSpectrum) < nTimeSamples {
			panic("fdkaac: short psy main output spectrum")
		}
		if chIdx[ch] < 0 {
			panic("fdkaac: negative psy main channel index")
		}
		base := chIdx[ch] * inputBufSize
		if base < 0 || base+nTimeSamples > len(input) {
			panic("fdkaac: short psy main input")
		}
		remainingStart := 2*nTimeSamples - blockSwitchingOffset
		remaining := blockSwitchingOffset - nTimeSamples
		if base+remainingStart+remaining > len(input) {
			panic("fdkaac: short psy main lookahead input")
		}
		if psyConf[0].SfbCnt <= 0 || psyConf[0].SfbCnt > maxSFB ||
			psyConf[1].SfbCnt <= 0 || psyConf[1].SfbCnt > maxSFBShort {
			panic("fdkaac: invalid psy main sfb count")
		}
	}
}
