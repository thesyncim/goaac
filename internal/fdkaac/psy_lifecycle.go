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
