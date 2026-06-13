package fdkaac

const psyPostMinThresholdLdData = FixpDBL(-0x42000000)

type PsyPostTnsScratch struct {
	GroupShort [2]GroupShortScratch
}

func FDKaacEncPsyPostTnsToolsAndOutput(
	channels int,
	psyStatic []*PsyStatic,
	psyData []*PsyData,
	tnsData []*TNSData,
	pnsData []*PNSData,
	psyConfLong *PsyConfiguration,
	psyConfShort *PsyConfiguration,
	psyOutElement *PsyOutElement,
	sfbTonality *[2][maxSFBLong]FixpSGL,
	scratch *PsyPostTnsScratch,
) int {
	checkPsyPostTnsInputs(
		channels, psyStatic, psyData, tnsData, pnsData,
		psyConfLong, psyConfShort, psyOutElement, sfbTonality, scratch,
	)

	var isShortWindow [2]bool
	var nWindows [2]int
	var windowLength [2]int
	var maxSfbPerGroup [2]int
	for ch := 0; ch < channels; ch++ {
		isShortWindow[ch] = psyStatic[ch].BlockSwitchingControl.LastWindowSequence == ShortWindow
		conf := psyPostConf(isShortWindow[ch], psyConfLong, psyConfShort)
		if isShortWindow[ch] {
			nWindows[ch] = transFac
			windowLength[ch] = conf.GranuleLength / transFac
		} else {
			nWindows[ch] = 1
			windowLength[ch] = conf.GranuleLength
		}
	}

	for ch := 0; ch < channels; ch++ {
		conf := psyPostConf(isShortWindow[ch], psyConfLong, psyConfShort)
		psyPostAdvanceThresholds(psyData[ch], psyStatic[ch], conf, isShortWindow[ch], nWindows[ch])
	}

	if channels == 2 {
		for w := 0; w < nWindows[1]; w++ {
			wOffset := w * windowLength[1]
			if isShortWindow[1] {
				FDKaacEncCalcBandNrgMSOpt(
					psyData[0].MdctSpectrum[wOffset:],
					psyData[1].MdctSpectrum[wOffset:],
					psyData[0].SfbMaxScaleSpec.Short[w][:],
					psyData[1].SfbMaxScaleSpec.Short[w][:],
					psyConfShort.SfbOffset[:],
					psyData[0].SfbActive,
					psyData[0].SfbEnergyMS.Short[w][:],
					psyData[1].SfbEnergyMS.Short[w][:],
					false,
					nil,
					nil,
				)
			} else {
				FDKaacEncCalcBandNrgMSOpt(
					psyData[0].MdctSpectrum[wOffset:],
					psyData[1].MdctSpectrum[wOffset:],
					psyData[0].SfbMaxScaleSpec.Long[:],
					psyData[1].SfbMaxScaleSpec.Long[:],
					psyConfLong.SfbOffset[:],
					psyData[0].SfbActive,
					psyData[0].SfbEnergyMS.Long[:],
					psyData[1].SfbEnergyMS.Long[:],
					true,
					psyData[0].SfbEnergyMSLdData[:],
					psyData[1].SfbEnergyMSLdData[:],
				)
			}
		}
	}

	for ch := 0; ch < channels; ch++ {
		out := psyOutElement.PsyOutChannel[ch]
		if isShortWindow[ch] {
			psyPostGroupShort(ch, channels, psyStatic[ch], psyData[ch], psyConfShort, out, &maxSfbPerGroup[ch], scratch)
		} else {
			psyPostPrepareLong(psyData[ch], psyConfLong, out, &maxSfbPerGroup[ch])
		}
	}

	for ch := 0; ch < channels; ch++ {
		clear(psyOutElement.PsyOutChannel[ch].IsBook[:])
		clear(psyOutElement.PsyOutChannel[ch].IsScale[:])
	}

	for ch := 0; ch < channels; ch++ {
		if psyStatic[ch].IsLFE != 0 {
			continue
		}
		pnsOffsetConf := psyConfLong
		if isShortWindow[ch] {
			pnsOffsetConf = psyConfShort
		}
		FDKaacEncPnsDetect(
			&psyConfLong.PnsConf,
			pnsData[ch],
			psyStatic[ch].BlockSwitchingControl.LastWindowSequence,
			psyData[ch].SfbActive,
			maxSfbPerGroup[ch],
			psyOutElement.PsyOutChannel[ch].SfbThresholdLdData[:],
			pnsOffsetConf.SfbOffset[:],
			psyData[ch].MdctSpectrum,
			psyData[ch].SfbMaxScaleSpec.Long[:],
			sfbTonality[ch][:],
			psyOutElement.PsyOutChannel[ch].TNSInfo.Order[0][0],
			tnsData[ch].DataRaw.Long.SubBlockInfo.PredictionGain[tnsHiFilt],
			tnsData[ch].DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt],
			psyOutElement.PsyOutChannel[ch].SfbEnergyLdData[:],
			psyOutElement.PsyOutChannel[ch].NoiseNrg[:],
		)
	}

	if channels == 2 {
		psyPostStereo(psyStatic, psyData, pnsData, psyConfLong, psyConfShort, psyOutElement, isShortWindow[0], &maxSfbPerGroup)
	}

	for ch := 0; ch < channels; ch++ {
		out := psyOutElement.PsyOutChannel[ch]
		if psyStatic[ch].IsLFE != 0 {
			for sfb := 0; sfb < psyData[ch].SfbActive; sfb++ {
				out.NoiseNrg[sfb] = noNoisePNS
			}
		} else {
			conf := psyPostConf(isShortWindow[ch], psyConfLong, psyConfShort)
			FDKaacEncCodePnsChannel(
				psyData[ch].SfbActive,
				&conf.PnsConf,
				pnsData[ch].PNSFlag[:],
				psyData[ch].SfbEnergyLdData.Long[:],
				out.NoiseNrg[:],
				out.SfbThresholdLdData[:],
			)
		}
	}

	for ch := 0; ch < channels; ch++ {
		conf := psyPostConf(isShortWindow[ch], psyConfLong, psyConfShort)
		out := psyOutElement.PsyOutChannel[ch]
		FDKaacEncBuildPsyOutChannel(
			out,
			&psyData[ch].SfbEnergy,
			&psyData[ch].SfbSpreadEnergy,
			&psyStatic[ch].BlockSwitchingControl,
			maxSfbPerGroup[ch],
			psyData[ch].MdctScale,
			isShortWindow[ch],
			psyConfLong.SfbActive,
			psyConfShort.SfbCnt,
		)
		out.MdctSpectrum = psyData[ch].MdctSpectrum[:conf.GranuleLength]
	}

	return AACEncOK
}

func psyPostConf(isShort bool, longConf *PsyConfiguration, shortConf *PsyConfiguration) *PsyConfiguration {
	if isShort {
		return shortConf
	}
	return longConf
}

func psyPostAdvanceThresholds(data *PsyData, static *PsyStatic, conf *PsyConfiguration, isShortWindow bool, nWindows int) {
	energyShift := data.MdctScale * 2
	clipNrgShift := energyShift - thrShiftBits
	headroom := 0
	if isShortWindow {
		headroom = 6
	}

	var clipEnergy FixpDBL
	switch {
	case clipNrgShift >= 0:
		clipEnergy = conf.ClipEnergy >> uint(clipNrgShift)
	case clipNrgShift >= -headroom:
		clipEnergy = conf.ClipEnergy << uint(-clipNrgShift)
	default:
		clipEnergy = MaxValDBL
	}

	for w := 0; w < nWindows; w++ {
		threshold := data.SfbThreshold.Long[:]
		energy := data.SfbEnergy.Long[:]
		spreadEnergy := data.SfbSpreadEnergy.Long[:]
		if isShortWindow {
			threshold = data.SfbThreshold.Short[w][:]
			energy = data.SfbEnergy.Short[w][:]
			spreadEnergy = data.SfbSpreadEnergy.Short[w][:]
		}

		for i := 0; i < data.SfbActive; i++ {
			threshold[i] = minFixpDBL(threshold[i], clipEnergy)
		}

		FDKaacEncSpreadingMax(data.SfbActive, conf.SfbMaskLowFactor[:], conf.SfbMaskHighFactor[:], threshold)

		energyShift += pcmQuantThrScale
		if energyShift >= 0 {
			energyShift = minInt(DfractBits-1, energyShift)
			shift := uint(energyShift)
			for i := 0; i < data.SfbActive; i++ {
				threshold[i] >>= thrShiftBits
				threshold[i] = maxFixpDBL(threshold[i], conf.SfbPcmQuantThreshold[i]>>shift)
			}
		} else {
			energyShift = minInt(DfractBits-1, -energyShift)
			shift := uint(energyShift)
			for i := 0; i < data.SfbActive; i++ {
				threshold[i] >>= thrShiftBits
				threshold[i] = maxFixpDBL(threshold[i], conf.SfbPcmQuantThreshold[i]<<shift)
			}
		}

		if static.IsLFE == 0 {
			if static.BlockSwitchingControl.LastWindowSequence == StopWindow {
				for i := 0; i < data.SfbActive; i++ {
					static.SfbThresholdNm1[i] = MaxValDBL
				}
				static.MdctScaleNm1 = 0
				static.CalcPreEcho = 0
			}

			FDKaacEncPreEchoControl(
				static.SfbThresholdNm1[:],
				static.CalcPreEcho,
				data.SfbActive,
				conf.MaxAllowedIncreaseFactor,
				conf.MinRemainingThresholdFactor,
				threshold,
				data.MdctScale,
				&static.MdctScaleNm1,
			)
			static.CalcPreEcho = 1

			if static.BlockSwitchingControl.LastWindowSequence == StartWindow {
				for i := 0; i < data.SfbActive; i++ {
					static.SfbThresholdNm1[i] = MaxValDBL
				}
				static.MdctScaleNm1 = 0
				static.CalcPreEcho = 0
			}
		}

		copy(spreadEnergy[:data.SfbActive], energy[:data.SfbActive])
		FDKaacEncSpreadingMax(data.SfbActive, conf.SfbMaskLowFactorSprEn[:], conf.SfbMaskHighFactorSprEn[:], spreadEnergy)
	}
}

func psyPostGroupShort(
	ch int,
	channels int,
	static *PsyStatic,
	data *PsyData,
	conf *PsyConfiguration,
	out *PsyOutChannel,
	maxSfbPerGroup *int,
	scratch *PsyPostTnsScratch,
) {
	var groupedSfbMinSnrLdData [maxGroupedSFB]FixpDBL
	noSfb := static.BlockSwitchingControl.NoOfGroups * conf.SfbCnt
	FDKaacEncGroupShortData(
		data.MdctSpectrum,
		&data.SfbThreshold,
		&data.SfbEnergy,
		&data.SfbEnergyMS,
		&data.SfbSpreadEnergy,
		conf.SfbCnt,
		data.SfbActive,
		conf.SfbOffset[:],
		conf.SfbMinSnrLdData[:],
		data.GroupedSfbOffset[:],
		maxSfbPerGroup,
		groupedSfbMinSnrLdData[:],
		static.BlockSwitchingControl.NoOfGroups,
		static.BlockSwitchingControl.GroupLen[:],
		conf.GranuleLength,
		&scratch.GroupShort[ch],
	)

	for sfbGrp := 0; sfbGrp < noSfb; sfbGrp += conf.SfbCnt {
		LdDataVector(data.SfbEnergy.Long[sfbGrp:], out.SfbEnergyLdData[sfbGrp:], data.SfbActive)
		LdDataVector(data.SfbThreshold.Long[sfbGrp:], out.SfbThresholdLdData[sfbGrp:], data.SfbActive)
		psyPostClampThresholdLdData(out.SfbThresholdLdData[sfbGrp:], data.SfbActive)
		if channels == 2 {
			LdDataVector(data.SfbEnergyMS.Long[sfbGrp:], data.SfbEnergyMSLdData[sfbGrp:], data.SfbActive)
		}
	}

	clear(out.SfbOffsets[:])
	copy(out.SfbOffsets[:], data.GroupedSfbOffset[:])
}

func psyPostPrepareLong(data *PsyData, conf *PsyConfiguration, out *PsyOutChannel, maxSfbPerGroup *int) {
	sfb := data.SfbActive - 1
	for ; sfb >= 0; sfb-- {
		line := conf.SfbOffset[sfb+1] - 1
		for ; line >= conf.SfbOffset[sfb]; line-- {
			if data.MdctSpectrum[line] != 0 {
				break
			}
		}
		if line > conf.SfbOffset[sfb] {
			break
		}
	}
	*maxSfbPerGroup = maxInt(minInt(5, data.SfbActive), sfb+1)

	copy(out.SfbEnergyLdData[:data.SfbActive], data.SfbEnergyLdData.Long[:data.SfbActive])
	clear(out.SfbOffsets[:])
	copy(out.SfbOffsets[:], conf.SfbOffset[:])
	LdDataVector(data.SfbThreshold.Long[:], out.SfbThresholdLdData[:], data.SfbActive)
	psyPostClampThresholdLdData(out.SfbThresholdLdData[:], data.SfbActive)
}

func psyPostClampThresholdLdData(values []FixpDBL, n int) {
	for i := 0; i < n; i++ {
		values[i] = maxFixpDBL(values[i], psyPostMinThresholdLdData)
	}
}

func psyPostStereo(
	psyStatic []*PsyStatic,
	psyData []*PsyData,
	pnsData []*PNSData,
	psyConfLong *PsyConfiguration,
	psyConfShort *PsyConfiguration,
	psyOutElement *PsyOutElement,
	isShortWindow bool,
	maxSfbPerGroup *[2]int,
) {
	psyOutElement.ToolsInfo.MsDigest = MsMaskNone
	clear(psyOutElement.ToolsInfo.MsMask[:])
	psyOutElement.CommonWindow = 1
	maxSfb := maxInt(maxSfbPerGroup[0], maxSfbPerGroup[1])
	maxSfbPerGroup[0] = maxSfb
	maxSfbPerGroup[1] = maxSfb

	if !isShortWindow {
		FDKaacEncPreProcessPnsChannelPair(
			psyData[0].SfbActive,
			psyData[0].SfbEnergy.Long[:],
			psyData[1].SfbEnergy.Long[:],
			psyOutElement.PsyOutChannel[0].SfbEnergyLdData[:],
			psyOutElement.PsyOutChannel[1].SfbEnergyLdData[:],
			psyData[0].SfbEnergyMS.Long[:],
			&psyConfLong.PnsConf,
			pnsData[0],
			pnsData[1],
		)
		FDKaacEncIntensityStereoProcessing(
			psyData[0].SfbEnergy.Long[:],
			psyData[1].SfbEnergy.Long[:],
			psyData[0].MdctSpectrum,
			psyData[1].MdctSpectrum,
			psyData[0].SfbThreshold.Long[:],
			psyData[1].SfbThreshold.Long[:],
			psyOutElement.PsyOutChannel[1].SfbThresholdLdData[:],
			psyData[0].SfbSpreadEnergy.Long[:],
			psyData[1].SfbSpreadEnergy.Long[:],
			psyOutElement.PsyOutChannel[0].SfbEnergyLdData[:],
			psyOutElement.PsyOutChannel[1].SfbEnergyLdData[:],
			&psyOutElement.ToolsInfo.MsDigest,
			psyOutElement.ToolsInfo.MsMask[:],
			psyConfLong.SfbCnt,
			psyConfLong.SfbCnt,
			maxSfbPerGroup[0],
			psyConfLong.SfbOffset[:],
			boolToInt(psyConfLong.AllowIS && psyOutElement.CommonWindow != 0),
			psyOutElement.PsyOutChannel[1].IsBook[:],
			psyOutElement.PsyOutChannel[1].IsScale[:],
			pnsData,
		)
		FDKaacEncMsStereoProcessing(
			psyData[0].MdctSpectrum,
			psyData[1].MdctSpectrum,
			psyData[0].SfbEnergy.Long[:],
			psyData[1].SfbEnergy.Long[:],
			psyData[0].SfbEnergyMS.Long[:],
			psyData[1].SfbEnergyMS.Long[:],
			psyData[0].SfbThreshold.Long[:],
			psyData[1].SfbThreshold.Long[:],
			psyData[0].SfbSpreadEnergy.Long[:],
			psyData[1].SfbSpreadEnergy.Long[:],
			psyOutElement.PsyOutChannel[0].SfbEnergyLdData[:],
			psyOutElement.PsyOutChannel[1].SfbEnergyLdData[:],
			psyData[0].SfbEnergyMSLdData[:],
			psyData[1].SfbEnergyMSLdData[:],
			psyOutElement.PsyOutChannel[0].SfbThresholdLdData[:],
			psyOutElement.PsyOutChannel[1].SfbThresholdLdData[:],
			psyOutElement.PsyOutChannel[1].IsBook[:],
			&psyOutElement.ToolsInfo.MsDigest,
			psyOutElement.ToolsInfo.MsMask[:],
			boolToInt(psyConfLong.AllowMS),
			psyData[0].SfbActive,
			psyData[0].SfbActive,
			maxSfbPerGroup[0],
			psyOutElement.PsyOutChannel[0].SfbOffsets[:],
		)
		FDKaacEncPostProcessPnsChannelPair(
			psyData[0].SfbActive,
			&psyConfLong.PnsConf,
			pnsData[0],
			pnsData[1],
			psyOutElement.ToolsInfo.MsMask[:],
			&psyOutElement.ToolsInfo.MsDigest,
		)
		return
	}

	sfbCnt := psyStatic[0].BlockSwitchingControl.NoOfGroups * psyConfShort.SfbCnt
	FDKaacEncIntensityStereoProcessing(
		psyData[0].SfbEnergy.Long[:],
		psyData[1].SfbEnergy.Long[:],
		psyData[0].MdctSpectrum,
		psyData[1].MdctSpectrum,
		psyData[0].SfbThreshold.Long[:],
		psyData[1].SfbThreshold.Long[:],
		psyOutElement.PsyOutChannel[1].SfbThresholdLdData[:],
		psyData[0].SfbSpreadEnergy.Long[:],
		psyData[1].SfbSpreadEnergy.Long[:],
		psyOutElement.PsyOutChannel[0].SfbEnergyLdData[:],
		psyOutElement.PsyOutChannel[1].SfbEnergyLdData[:],
		&psyOutElement.ToolsInfo.MsDigest,
		psyOutElement.ToolsInfo.MsMask[:],
		sfbCnt,
		psyConfShort.SfbCnt,
		maxSfbPerGroup[0],
		psyData[0].GroupedSfbOffset[:],
		boolToInt(psyConfLong.AllowIS && psyOutElement.CommonWindow != 0),
		psyOutElement.PsyOutChannel[1].IsBook[:],
		psyOutElement.PsyOutChannel[1].IsScale[:],
		pnsData,
	)
	FDKaacEncMsStereoProcessing(
		psyData[0].MdctSpectrum,
		psyData[1].MdctSpectrum,
		psyData[0].SfbEnergy.Long[:],
		psyData[1].SfbEnergy.Long[:],
		psyData[0].SfbEnergyMS.Long[:],
		psyData[1].SfbEnergyMS.Long[:],
		psyData[0].SfbThreshold.Long[:],
		psyData[1].SfbThreshold.Long[:],
		psyData[0].SfbSpreadEnergy.Long[:],
		psyData[1].SfbSpreadEnergy.Long[:],
		psyOutElement.PsyOutChannel[0].SfbEnergyLdData[:],
		psyOutElement.PsyOutChannel[1].SfbEnergyLdData[:],
		psyData[0].SfbEnergyMSLdData[:],
		psyData[1].SfbEnergyMSLdData[:],
		psyOutElement.PsyOutChannel[0].SfbThresholdLdData[:],
		psyOutElement.PsyOutChannel[1].SfbThresholdLdData[:],
		psyOutElement.PsyOutChannel[1].IsBook[:],
		&psyOutElement.ToolsInfo.MsDigest,
		psyOutElement.ToolsInfo.MsMask[:],
		boolToInt(psyConfShort.AllowMS),
		sfbCnt,
		psyConfShort.SfbCnt,
		maxSfbPerGroup[0],
		psyOutElement.PsyOutChannel[0].SfbOffsets[:],
	)
}

func checkPsyPostTnsInputs(
	channels int,
	psyStatic []*PsyStatic,
	psyData []*PsyData,
	tnsData []*TNSData,
	pnsData []*PNSData,
	psyConfLong *PsyConfiguration,
	psyConfShort *PsyConfiguration,
	psyOutElement *PsyOutElement,
	sfbTonality *[2][maxSFBLong]FixpSGL,
	scratch *PsyPostTnsScratch,
) {
	if channels <= 0 || channels > 2 {
		panic("fdkaac: invalid psy post-TNS channel count")
	}
	if len(psyStatic) < channels || len(psyData) < channels || len(tnsData) < channels || len(pnsData) < channels {
		panic("fdkaac: short psy post-TNS channel list")
	}
	if psyConfLong == nil || psyConfShort == nil || psyOutElement == nil || sfbTonality == nil || scratch == nil {
		panic("fdkaac: nil psy post-TNS state")
	}
	if psyConfLong.GranuleLength <= 0 || psyConfLong.GranuleLength > maxSpectralLines ||
		psyConfShort.GranuleLength <= 0 || psyConfShort.GranuleLength > maxSpectralLines ||
		psyConfShort.GranuleLength%transFac != 0 {
		panic("fdkaac: invalid psy post-TNS granule length")
	}
	for ch := 0; ch < channels; ch++ {
		if psyStatic[ch] == nil || psyData[ch] == nil || tnsData[ch] == nil || pnsData[ch] == nil || psyOutElement.PsyOutChannel[ch] == nil {
			panic("fdkaac: nil psy post-TNS channel state")
		}
		if !validSFBWindowSequence(psyStatic[ch].BlockSwitchingControl.LastWindowSequence) {
			panic("fdkaac: invalid psy post-TNS window sequence")
		}
		isShort := psyStatic[ch].BlockSwitchingControl.LastWindowSequence == ShortWindow
		conf := psyPostConf(isShort, psyConfLong, psyConfShort)
		if len(psyData[ch].MdctSpectrum) < conf.GranuleLength {
			panic("fdkaac: short psy post-TNS spectrum")
		}
		if psyData[ch].SfbActive <= 0 || psyData[ch].SfbActive > conf.SfbCnt {
			panic("fdkaac: invalid psy post-TNS active band count")
		}
		if conf.SfbCnt <= 0 || conf.SfbCnt > maxSFB {
			panic("fdkaac: invalid psy post-TNS sfb count")
		}
	}
	if channels == 2 {
		leftShort := psyStatic[0].BlockSwitchingControl.LastWindowSequence == ShortWindow
		rightShort := psyStatic[1].BlockSwitchingControl.LastWindowSequence == ShortWindow
		if leftShort != rightShort || psyStatic[0].BlockSwitchingControl.LastWindowSequence != psyStatic[1].BlockSwitchingControl.LastWindowSequence {
			panic("fdkaac: psy post-TNS stereo requires common window")
		}
		if psyData[0].SfbActive != psyData[1].SfbActive {
			panic("fdkaac: psy post-TNS stereo active band mismatch")
		}
	}
}
