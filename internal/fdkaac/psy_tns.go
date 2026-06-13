package fdkaac

type PsyTnsTonalityScratch struct {
	Tonality [2]TonalityScratch
}

func FDKaacEncPsyAdvanceTnsAndTonality(
	channels int,
	psyStatic []*PsyStatic,
	psyData []*PsyData,
	tnsData []*TNSData,
	psyConf []*PsyConfiguration,
	psyOutChannel []*PsyOutChannel,
	sfbTonality *[2][maxSFBLong]FixpSGL,
	scratch *PsyTnsTonalityScratch,
) int {
	checkPsyAdvanceTnsInputs(channels, psyStatic, psyData, tnsData, psyConf, psyOutChannel, sfbTonality, scratch)

	if channels >= 1 && psyStatic[0].IsLFE != 0 {
		tnsData[0].DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 0
		tnsData[0].DataRaw.Long.SubBlockInfo.TnsActive[tnsLoFilt] = 0
		return AACEncOK
	}

	var isShortWindow [2]bool
	var nWindows [2]int
	var windowLength [2]int
	for ch := 0; ch < channels; ch++ {
		isShortWindow[ch] = psyStatic[ch].BlockSwitchingControl.LastWindowSequence == ShortWindow
		if isShortWindow[ch] {
			nWindows[ch] = transFac
			windowLength[ch] = psyConf[ch].GranuleLength / transFac
		} else {
			nWindows[ch] = 1
			windowLength[ch] = psyConf[ch].GranuleLength
		}
	}

	for ch := 0; ch < channels; ch++ {
		if !isShortWindow[ch] {
			FDKaacEncCalculateFullTonality(
				psyData[ch].MdctSpectrum,
				psyData[ch].SfbMaxScaleSpec.Long[:],
				psyData[ch].SfbEnergyLdData.Long[:],
				sfbTonality[ch][:],
				psyData[ch].SfbActive,
				psyConf[ch].SfbOffset[:],
				psyConf[ch].PnsConf.UsePns,
				&scratch.Tonality[ch],
			)
		}
	}

	if psyConf[0].TnsConf.TnsActive == 0 && (channels == 1 || psyConf[1].TnsConf.TnsActive == 0) {
		for ch := 0; ch < channels; ch++ {
			*tnsData[ch] = TNSData{}
		}
		return AACEncOK
	}

	var tnsActive [transFac]int
	var nrgScaling [2]int
	tnsSpecShift := 0

	for ch := 0; ch < channels; ch++ {
		for w := 0; w < nWindows[ch]; w++ {
			wOffset := w * windowLength[ch]
			FDKaacEncTnsDetect(
				tnsData[ch],
				&psyConf[ch].TnsConf,
				&psyOutChannel[ch].TNSInfo,
				psyConf[ch].SfbCnt,
				psyData[ch].MdctSpectrum[wOffset:],
				w,
				psyStatic[ch].BlockSwitchingControl.LastWindowSequence,
			)
		}
	}

	if channels == 2 {
		FDKaacEncTnsSync(
			tnsData[1],
			tnsData[0],
			&psyOutChannel[1].TNSInfo,
			&psyOutChannel[0].TNSInfo,
			psyStatic[1].BlockSwitchingControl.LastWindowSequence,
			psyStatic[0].BlockSwitchingControl.LastWindowSequence,
			&psyConf[1].TnsConf,
		)
	}

	if channels >= 1 {
		for w := 0; w < nWindows[0]; w++ {
			if isShortWindow[0] {
				tnsActive[w] = boolToInt(
					tnsData[0].DataRaw.Short.SubBlockInfo[w].TnsActive[tnsHiFilt] != 0 ||
						tnsData[0].DataRaw.Short.SubBlockInfo[w].TnsActive[tnsLoFilt] != 0 ||
						tnsData[channels-1].DataRaw.Short.SubBlockInfo[w].TnsActive[tnsHiFilt] != 0 ||
						tnsData[channels-1].DataRaw.Short.SubBlockInfo[w].TnsActive[tnsLoFilt] != 0,
				)
			} else {
				tnsActive[w] = boolToInt(
					tnsData[0].DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] != 0 ||
						tnsData[0].DataRaw.Long.SubBlockInfo.TnsActive[tnsLoFilt] != 0 ||
						tnsData[channels-1].DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] != 0 ||
						tnsData[channels-1].DataRaw.Long.SubBlockInfo.TnsActive[tnsLoFilt] != 0,
				)
			}
		}
	}

	for ch := 0; ch < channels; ch++ {
		if tnsActive[0] != 0 && !isShortWindow[ch] {
			for line := 0; line < psyConf[ch].LowpassLine; line++ {
				psyData[ch].MdctSpectrum[line] >>= 1
			}
			for sfb := 0; sfb < psyData[ch].SfbActive; sfb++ {
				psyData[ch].SfbThreshold.Long[sfb] >>= 2
			}
			psyData[ch].MdctScale++
		}
	}

	for ch := 0; ch < channels; ch++ {
		for w := 0; w < nWindows[ch]; w++ {
			wOffset := w * windowLength[ch]
			FDKaacEncTnsEncode(
				&psyOutChannel[ch].TNSInfo,
				tnsData[ch],
				psyConf[ch].SfbCnt,
				&psyConf[ch].TnsConf,
				psyConf[ch].SfbOffset[psyData[ch].SfbActive],
				psyData[ch].MdctSpectrum[wOffset:],
				w,
				psyStatic[ch].BlockSwitchingControl.LastWindowSequence,
			)

			if tnsActive[w] != 0 {
				if isShortWindow[ch] {
					FDKaacEncCalcSfbMaxScaleSpec(
						psyData[ch].MdctSpectrum[wOffset:],
						psyConf[ch].SfbOffset[:],
						psyData[ch].SfbMaxScaleSpec.Short[w][:],
						psyData[ch].SfbActive,
					)
				} else {
					FDKaacEncCalcSfbMaxScaleSpec(
						psyData[ch].MdctSpectrum[wOffset:],
						psyConf[ch].SfbOffset[:],
						psyData[ch].SfbMaxScaleSpec.Long[:],
						psyData[ch].SfbActive,
					)
				}
			}
		}
	}

	for ch := 0; ch < channels; ch++ {
		for w := 0; w < nWindows[ch]; w++ {
			if tnsActive[w] == 0 {
				continue
			}
			wOffset := w * windowLength[ch]
			if isShortWindow[ch] {
				FDKaacEncCalcBandEnergyOptimShort(
					psyData[ch].MdctSpectrum[wOffset:],
					psyData[ch].SfbMaxScaleSpec.Short[w][:],
					psyConf[ch].SfbOffset[:],
					psyData[ch].SfbActive,
					psyData[ch].SfbEnergy.Short[w][:],
				)
			} else {
				nrgScaling[ch] = FDKaacEncCalcBandEnergyOptimLong(
					psyData[ch].MdctSpectrum,
					psyData[ch].SfbMaxScaleSpec.Long[:],
					psyConf[ch].SfbOffset[:],
					psyData[ch].SfbActive,
					psyData[ch].SfbEnergy.Long[:],
					psyData[ch].SfbEnergyLdData.Long[:],
				)
				tnsSpecShift = maxInt(tnsSpecShift, nrgScaling[ch])
			}
		}
	}

	for ch := 0; ch < channels; ch++ {
		if tnsSpecShift == 0 || isShortWindow[ch] {
			continue
		}
		for line := 0; line < psyConf[ch].LowpassLine; line++ {
			psyData[ch].MdctSpectrum[line] >>= uint(tnsSpecShift)
		}
		scale := (tnsSpecShift - nrgScaling[ch]) << 1
		for sfb := 0; sfb < psyData[ch].SfbActive; sfb++ {
			psyData[ch].SfbEnergyLdData.Long[sfb] -= FixpDBL(scale) * ldDataStep1Over64
			psyData[ch].SfbEnergy.Long[sfb] >>= uint(scale)
			psyData[ch].SfbThreshold.Long[sfb] >>= uint(tnsSpecShift << 1)
		}
		psyData[ch].MdctScale += tnsSpecShift
	}

	return AACEncOK
}

func checkPsyAdvanceTnsInputs(
	channels int,
	psyStatic []*PsyStatic,
	psyData []*PsyData,
	tnsData []*TNSData,
	psyConf []*PsyConfiguration,
	psyOutChannel []*PsyOutChannel,
	sfbTonality *[2][maxSFBLong]FixpSGL,
	scratch *PsyTnsTonalityScratch,
) {
	if channels <= 0 || channels > 2 {
		panic("fdkaac: invalid psy TNS channel count")
	}
	if len(psyStatic) < channels || len(psyData) < channels || len(tnsData) < channels ||
		len(psyConf) < channels || len(psyOutChannel) < channels {
		panic("fdkaac: short psy TNS channel list")
	}
	if sfbTonality == nil || scratch == nil {
		panic("fdkaac: nil psy TNS scratch")
	}
	for ch := 0; ch < channels; ch++ {
		if psyStatic[ch] == nil || psyData[ch] == nil || tnsData[ch] == nil ||
			psyConf[ch] == nil || psyOutChannel[ch] == nil {
			panic("fdkaac: nil psy TNS channel state")
		}
		if psyConf[ch].GranuleLength <= 0 || psyConf[ch].GranuleLength > maxSpectralLines {
			panic("fdkaac: invalid psy TNS granule length")
		}
		if len(psyData[ch].MdctSpectrum) < psyConf[ch].GranuleLength {
			panic("fdkaac: short psy TNS spectrum")
		}
		if !validSFBWindowSequence(psyStatic[ch].BlockSwitchingControl.LastWindowSequence) {
			panic("fdkaac: invalid psy TNS window sequence")
		}
		if psyData[ch].SfbActive <= 0 || psyData[ch].SfbActive > psyConf[ch].SfbCnt {
			panic("fdkaac: invalid psy TNS active band count")
		}
		if psyConf[ch].SfbCnt <= 0 || psyConf[ch].SfbCnt > maxSFB {
			panic("fdkaac: invalid psy TNS sfb count")
		}
		if psyConf[ch].LowpassLine < 0 || psyConf[ch].LowpassLine > psyConf[ch].GranuleLength {
			panic("fdkaac: invalid psy TNS lowpass line")
		}
		if psyStatic[ch].BlockSwitchingControl.LastWindowSequence == ShortWindow &&
			psyConf[ch].GranuleLength%transFac != 0 {
			panic("fdkaac: invalid short psy TNS granule length")
		}
		checkBandEnergySpectrum(psyData[ch].MdctSpectrum, psyConf[ch].SfbOffset[:], psyData[ch].SfbActive)
	}
}
