package fdkaac

const (
	maxPsyInputBufferSize = 2 * 1024
	maxPsyOverlapAddSize  = 3 * 512 / 2
)

type PsyElement struct {
	PsyStatic [2]*PsyStatic
}

type PsyData struct {
	MdctSpectrum      []FixpDBL
	SfbThreshold      SFBThreshold
	SfbEnergy         SFBEnergy
	SfbEnergyLdData   SFBEnergy
	SfbMaxScaleSpec   SFBMaxScale
	SfbEnergyMS       SFBEnergy
	SfbEnergyMSLdData [maxGroupedSFB]FixpDBL
	SfbSpreadEnergy   SFBEnergy
	MdctScale         int
	GroupedSfbOffset  [maxGroupedSFB + 1]int
	SfbActive         int
	LowpassLine       int
}

type SFBMaxScale struct {
	Long  [maxGroupedSFB]int
	Short [transFac][maxSFBShort]int
}

type PsyDynamic struct {
	PsyData [2]PsyData
	TNSData [2]TNSData
	PNSData [2]PNSData
}

type PsyStatic struct {
	PsyInputBuffer        [maxPsyInputBufferSize]int16
	OverlapAddBuffer      [maxPsyOverlapAddSize]FixpDBL
	MDCT                  MDCT
	BlockSwitchingControl BlockSwitchingControl
	SfbThresholdNm1       [maxGroupedSFB]FixpDBL
	MdctScaleNm1          int
	CalcPreEcho           int
	IsLFE                 int
}

type PsyInternal struct {
	PsyElement     [maxChannelElements]*PsyElement
	PStaticChannel [maxAACChannels]*PsyStatic
	PsyDynamic     *PsyDynamic
	PsyConf        [2]PsyConfiguration
	GranuleLength  int

	PsyElements    [maxChannelElements]PsyElement
	StaticChannels [maxAACChannels]PsyStatic
	Dynamic        PsyDynamic
}

type PsyOut struct {
	PsyOutElement  [maxChannelElements]*PsyOutElement
	PPsyOutChannel [maxAACChannels]*PsyOutChannel
}

type PsyOutState struct {
	PsyOut           [maxAACSubFrames]PsyOut
	PsyOutPtr        [maxAACSubFrames]*PsyOut
	PsyOutChannels   [maxAACSubFrames][maxAACChannels]PsyOutChannel
	PsyOutElements   [maxAACSubFrames][maxChannelElements]PsyOutElement
	PsyOutElementPtr [maxAACSubFrames][maxChannelElements]*PsyOutElement
}

func FDKaacEncPsyNew(state *PsyInternal, nElements int, nChannels int) int {
	if state == nil {
		return AACEncInvalidHandle
	}
	if nElements < 0 || nChannels < 0 {
		panic("fdkaac: negative psy dimensions")
	}
	if nElements > maxChannelElements || nChannels > maxAACChannels {
		return AACEncNoMemory
	}

	*state = PsyInternal{}
	for i := 0; i < nElements; i++ {
		state.PsyElement[i] = &state.PsyElements[i]
	}
	for i := 0; i < nChannels; i++ {
		state.PStaticChannel[i] = &state.StaticChannels[i]
	}
	state.PsyDynamic = &state.Dynamic
	return AACEncOK
}

func FDKaacEncPsyOutNew(state *PsyOutState, nElements int, nChannels int, nSubFrames int) int {
	if state == nil {
		return AACEncInvalidHandle
	}
	if nElements < 0 || nChannels < 0 || nSubFrames < 0 {
		panic("fdkaac: negative psy output dimensions")
	}
	if nElements > maxChannelElements || nChannels > maxAACChannels || nSubFrames > maxAACSubFrames {
		return AACEncNoMemory
	}

	*state = PsyOutState{}
	for n := 0; n < nSubFrames; n++ {
		state.PsyOutPtr[n] = &state.PsyOut[n]
		psyOut := state.PsyOutPtr[n]

		for i := 0; i < nChannels; i++ {
			psyOut.PPsyOutChannel[i] = &state.PsyOutChannels[n][i]
		}
		for i := 0; i < nElements; i++ {
			state.PsyOutElementPtr[n][i] = &state.PsyOutElements[n][i]
			psyOut.PsyOutElement[i] = state.PsyOutElementPtr[n][i]
		}
	}
	return AACEncOK
}

func FDKaacEncPsyInitStates(hPsy *PsyInternal, psyStatic *PsyStatic, audioObjectType AudioObjectType) int {
	_ = hPsy
	if psyStatic == nil {
		panic("fdkaac: nil psy static state")
	}

	psyStatic.PsyInputBuffer = [maxPsyInputBufferSize]int16{}
	FDKaacEncInitBlockSwitching(&psyStatic.BlockSwitchingControl, isLowDelayAOT(audioObjectType))
	return AACEncOK
}

func FDKaacEncPsyInit(
	hPsy *PsyInternal,
	psyOut []*PsyOut,
	nSubFrames int,
	nMaxChannels int,
	audioObjectType AudioObjectType,
	cm *ChannelMapping,
) int {
	checkPsyInitInputs(hPsy, psyOut, nSubFrames, nMaxChannels, cm)

	chInc := 0
	resetChannels := 3
	if nMaxChannels > 2 && cm.NChannels == 2 {
		chInc = 1
		FDKaacEncPsyInitStates(hPsy, hPsy.PStaticChannel[0], audioObjectType)
	}
	if nMaxChannels == 2 {
		resetChannels = 0
	}

	mappedChannels := 0
	for i := 0; i < cm.NElements; i++ {
		for ch := 0; ch < cm.ElInfo[i].NChannelsInEl; ch++ {
			psyStatic := hPsy.PStaticChannel[chInc]
			hPsy.PsyElement[i].PsyStatic[ch] = psyStatic
			if cm.ElInfo[i].ElType != idLFE {
				if chInc >= resetChannels {
					FDKaacEncPsyInitStates(hPsy, psyStatic, audioObjectType)
				}
				MDCTInit(&psyStatic.MDCT, nil)
				psyStatic.IsLFE = 0
			} else {
				psyStatic.IsLFE = 1
			}
			chInc++
			mappedChannels++
		}
	}
	if mappedChannels != cm.NChannels {
		panic("fdkaac: inconsistent psy channel mapping")
	}

	for n := 0; n < nSubFrames; n++ {
		chInc = 0
		for i := 0; i < cm.NElements; i++ {
			psyOutElement := psyOut[n].PsyOutElement[i]
			for ch := 0; ch < cm.ElInfo[i].NChannelsInEl; ch++ {
				psyOutElement.PsyOutChannel[ch] = psyOut[n].PPsyOutChannel[chInc]
				chInc++
			}
		}
	}
	return AACEncOK
}

func FDKaacEncPsyMainInit(
	hPsy *PsyInternal,
	audioObjectType AudioObjectType,
	cm *ChannelMapping,
	sampleRate int,
	granuleLength int,
	bitRate int,
	tnsMask int,
	bandwidth int,
	usePns int,
	useIS int,
	useMS int,
	syntaxFlags uint32,
	initFlags uint32,
) int {
	checkPsyMainInitInputs(hPsy, cm, sampleRate, granuleLength, bitRate, bandwidth)

	channelsEff := cm.NChannelsEff
	tnsChannels := 0
	switch FDKaacEncGetMonoStereoMode(cm.EncMode) {
	case ElementModeMono:
		tnsChannels = 1
	case ElementModeStereo:
		tnsChannels = 2
	default:
		tnsChannels = 0
	}

	filterbank := FilterbankLC
	switch audioObjectType {
	case AOTERAACLD:
		filterbank = FilterbankLD
	case AOTERAACELD:
		filterbank = FilterbankELD
	}

	hPsy.GranuleLength = granuleLength

	err := FDKaacEncInitPsyConfiguration(
		bitRate/channelsEff,
		sampleRate,
		bandwidth,
		LongWindow,
		hPsy.GranuleLength,
		useIS,
		useMS,
		&hPsy.PsyConf[0],
		filterbank,
	)
	if err != AACEncOK {
		return err
	}

	ldSbrPresent := 0
	if syntaxFlags&acSBRPresent != 0 {
		ldSbrPresent = 1
	}
	isLowDelay := 0
	if isLowDelayAOT(audioObjectType) {
		isLowDelay = 1
	}
	err = FDKaacEncInitTnsConfiguration(
		(bitRate*tnsChannels)/channelsEff,
		sampleRate,
		tnsChannels,
		LongWindow,
		hPsy.GranuleLength,
		isLowDelay,
		ldSbrPresent,
		&hPsy.PsyConf[0].TnsConf,
		&hPsy.PsyConf[0],
		tnsMask&2,
		tnsMask&8,
	)
	if err != AACEncOK {
		return err
	}

	if granuleLength > 512 {
		err = FDKaacEncInitPsyConfiguration(
			bitRate/channelsEff,
			sampleRate,
			bandwidth,
			ShortWindow,
			hPsy.GranuleLength,
			useIS,
			useMS,
			&hPsy.PsyConf[1],
			filterbank,
		)
		if err != AACEncOK {
			return err
		}

		err = FDKaacEncInitTnsConfiguration(
			(bitRate*tnsChannels)/channelsEff,
			sampleRate,
			tnsChannels,
			ShortWindow,
			hPsy.GranuleLength,
			isLowDelay,
			ldSbrPresent,
			&hPsy.PsyConf[1].TnsConf,
			&hPsy.PsyConf[1],
			tnsMask&1,
			tnsMask&4,
		)
		if err != AACEncOK {
			return err
		}
	}

	for i := 0; i < cm.NElements; i++ {
		for ch := 0; ch < cm.ElInfo[i].NChannelsInEl; ch++ {
			psyStatic := hPsy.PsyElement[i].PsyStatic[ch]
			if initFlags != 0 {
				FDKaacEncPsyInitStates(hPsy, psyStatic, audioObjectType)
			}
			FDKaacEncInitPreEchoControl(
				psyStatic.SfbThresholdNm1[:],
				&psyStatic.CalcPreEcho,
				hPsy.PsyConf[0].SfbCnt,
				hPsy.PsyConf[0].SfbPcmQuantThreshold[:],
				&psyStatic.MdctScaleNm1,
			)
		}
	}

	err = FDKaacEncInitPnsConfiguration(
		&hPsy.PsyConf[0].PnsConf,
		bitRate/channelsEff,
		sampleRate,
		usePns,
		hPsy.PsyConf[0].SfbCnt,
		hPsy.PsyConf[0].SfbOffset[:],
		cm.ElInfo[0].NChannelsInEl,
		boolToInt(hPsy.PsyConf[0].Filterbank == FilterbankLC),
	)
	if err != AACEncOK {
		return err
	}

	if granuleLength > 512 {
		err = FDKaacEncInitPnsConfiguration(
			&hPsy.PsyConf[1].PnsConf,
			bitRate/channelsEff,
			sampleRate,
			usePns,
			hPsy.PsyConf[1].SfbCnt,
			hPsy.PsyConf[1].SfbOffset[:],
			cm.ElInfo[1].NChannelsInEl,
			boolToInt(hPsy.PsyConf[1].Filterbank == FilterbankLC),
		)
		if err != AACEncOK {
			return err
		}
	}

	return err
}

func checkPsyInitInputs(hPsy *PsyInternal, psyOut []*PsyOut, nSubFrames int, nMaxChannels int, cm *ChannelMapping) {
	if hPsy == nil || cm == nil {
		panic("fdkaac: nil psy init state")
	}
	if nSubFrames < 0 || nMaxChannels < 0 {
		panic("fdkaac: negative psy init dimensions")
	}
	if nSubFrames > maxAACSubFrames || nMaxChannels > maxAACChannels {
		panic("fdkaac: psy init dimensions exceed fixed storage")
	}
	if len(psyOut) < nSubFrames {
		panic("fdkaac: short psy output frame list")
	}
	if cm.NElements < 0 || cm.NElements > maxChannelElements || cm.NChannels < 0 || cm.NChannels > maxAACChannels {
		panic("fdkaac: invalid psy channel mapping")
	}
	if cm.NChannels > nMaxChannels || (nMaxChannels > 2 && cm.NChannels == 2 && nMaxChannels < 3) {
		panic("fdkaac: invalid psy max-channel configuration")
	}

	staticOffset := 0
	if nMaxChannels > 2 && cm.NChannels == 2 {
		staticOffset = 1
	}
	for ch := 0; ch < cm.NChannels+staticOffset; ch++ {
		if hPsy.PStaticChannel[ch] == nil {
			panic("fdkaac: nil psy static channel")
		}
	}
	for i := 0; i < cm.NElements; i++ {
		if hPsy.PsyElement[i] == nil {
			panic("fdkaac: nil psy element")
		}
		nChannelsInEl := cm.ElInfo[i].NChannelsInEl
		if nChannelsInEl < 0 || nChannelsInEl > len(hPsy.PsyElement[i].PsyStatic) {
			panic("fdkaac: invalid psy element channel count")
		}
	}

	for n := 0; n < nSubFrames; n++ {
		if psyOut[n] == nil {
			panic("fdkaac: nil psy output frame")
		}
		chInc := 0
		for i := 0; i < cm.NElements; i++ {
			if psyOut[n].PsyOutElement[i] == nil {
				panic("fdkaac: nil psy output element")
			}
			nChannelsInEl := cm.ElInfo[i].NChannelsInEl
			if nChannelsInEl < 0 || nChannelsInEl > len(psyOut[n].PsyOutElement[i].PsyOutChannel) {
				panic("fdkaac: invalid psy output element channel count")
			}
			for ch := 0; ch < nChannelsInEl; ch++ {
				if chInc >= len(psyOut[n].PPsyOutChannel) || psyOut[n].PPsyOutChannel[chInc] == nil {
					panic("fdkaac: nil psy output channel")
				}
				chInc++
			}
		}
		if chInc != cm.NChannels {
			panic("fdkaac: inconsistent psy output channel mapping")
		}
	}
}

func checkPsyMainInitInputs(hPsy *PsyInternal, cm *ChannelMapping, sampleRate int, granuleLength int, bitRate int, bandwidth int) {
	if hPsy == nil || cm == nil {
		panic("fdkaac: nil psy main init state")
	}
	if sampleRate <= 0 || granuleLength <= 0 || bitRate <= 0 || bandwidth <= 0 {
		panic("fdkaac: invalid psy main init control")
	}
	if cm.NChannelsEff <= 0 || cm.NChannelsEff > maxAACChannels || cm.NElements < 0 || cm.NElements > maxChannelElements {
		panic("fdkaac: invalid psy main init channel mapping")
	}
	if cm.NChannels < 0 || cm.NChannels > maxAACChannels {
		panic("fdkaac: invalid psy main init channel count")
	}
	for i := 0; i < cm.NElements; i++ {
		if hPsy.PsyElement[i] == nil {
			panic("fdkaac: nil psy main init element")
		}
		nChannelsInEl := cm.ElInfo[i].NChannelsInEl
		if nChannelsInEl < 0 || nChannelsInEl > len(hPsy.PsyElement[i].PsyStatic) {
			panic("fdkaac: invalid psy main init element channel count")
		}
		for ch := 0; ch < nChannelsInEl; ch++ {
			if hPsy.PsyElement[i].PsyStatic[ch] == nil {
				panic("fdkaac: nil psy main init static channel")
			}
		}
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
