package fdkaac

type QCState struct {
	Kernel              QCKernel
	AdjThrState         AdjThrState
	AdjThrStateElements [maxChannelElements]ATSElement
	BitCounter          BitCounterState
	ElementBits         [maxChannelElements]ElementBits
	ElementBitsPtr      [maxChannelElements]*ElementBits
}

type QCOutState struct {
	QCOut           [maxAACSubFrames]QCOut
	QCOutPtr        [maxAACSubFrames]*QCOut
	QCOutChannels   [maxAACSubFrames][maxAACChannels]QCOutChannel
	QCOutElements   [maxAACSubFrames][maxChannelElements]QCOutElement
	QCOutElementPtr [maxAACSubFrames][maxChannelElements]*QCOutElement
}

func FDKaacEncQCNew(state *QCState, nElements int) int {
	if state == nil {
		return AACEncInvalidHandle
	}
	if nElements < 0 {
		panic("fdkaac: negative QC element count")
	}
	if nElements > maxChannelElements {
		return AACEncNoMemory
	}

	*state = QCState{}
	for i := 0; i < nElements; i++ {
		state.AdjThrState.AdjThrStateElem[i] = &state.AdjThrStateElements[i]
		state.ElementBitsPtr[i] = &state.ElementBits[i]
	}
	return AACEncOK
}

func FDKaacEncQCOutNew(state *QCOutState, nElements int, nChannels int, nSubFrames int) int {
	if state == nil {
		return AACEncInvalidHandle
	}
	if nElements < 0 || nChannels < 0 || nSubFrames < 0 {
		panic("fdkaac: negative QC output dimensions")
	}
	if nElements > maxChannelElements || nChannels > maxAACChannels || nSubFrames > maxAACSubFrames {
		return AACEncNoMemory
	}

	*state = QCOutState{}
	for n := 0; n < nSubFrames; n++ {
		state.QCOutPtr[n] = &state.QCOut[n]
		qcOut := state.QCOutPtr[n]

		for i := 0; i < nChannels; i++ {
			qcOut.PQCOutChannels[i] = &state.QCOutChannels[n][i]
		}
		for i := 0; i < nElements; i++ {
			state.QCOutElementPtr[n][i] = &state.QCOutElements[n][i]
			qcOut.QCElement[i] = state.QCOutElementPtr[n][i]
		}
	}
	return AACEncOK
}

func FDKaacEncQCOutInit(qcOut []*QCOut, nSubFrames int, cm *ChannelMapping) int {
	checkQCOutInitInputs(qcOut, nSubFrames, cm)

	for n := 0; n < nSubFrames; n++ {
		chInc := 0
		for i := 0; i < cm.NElements; i++ {
			qcElement := qcOut[n].QCElement[i]
			for ch := 0; ch < cm.ElInfo[i].NChannelsInEl; ch++ {
				qcElement.QCOutChannel[ch] = qcOut[n].PQCOutChannels[chInc]
				chInc++
			}
		}
	}
	return AACEncOK
}

func checkQCOutInitInputs(qcOut []*QCOut, nSubFrames int, cm *ChannelMapping) {
	if cm == nil {
		panic("fdkaac: nil QC output channel mapping")
	}
	if nSubFrames < 0 {
		panic("fdkaac: negative QC output subframe count")
	}
	if len(qcOut) < nSubFrames {
		panic("fdkaac: short QC output frame list")
	}
	if cm.NElements < 0 || cm.NElements > maxChannelElements || cm.NChannels < 0 || cm.NChannels > maxAACChannels {
		panic("fdkaac: invalid QC output channel mapping")
	}

	for n := 0; n < nSubFrames; n++ {
		if qcOut[n] == nil {
			panic("fdkaac: nil QC output frame")
		}
		chInc := 0
		for i := 0; i < cm.NElements; i++ {
			qcElement := qcOut[n].QCElement[i]
			if qcElement == nil {
				panic("fdkaac: nil QC output element")
			}
			nChannelsInEl := cm.ElInfo[i].NChannelsInEl
			if nChannelsInEl < 0 || nChannelsInEl > len(qcElement.QCOutChannel) {
				panic("fdkaac: invalid QC output element channel count")
			}
			for ch := 0; ch < nChannelsInEl; ch++ {
				if chInc >= len(qcOut[n].PQCOutChannels) || qcOut[n].PQCOutChannels[chInc] == nil {
					panic("fdkaac: nil QC output channel")
				}
				chInc++
			}
		}
		if chInc != cm.NChannels {
			panic("fdkaac: inconsistent QC output channel mapping")
		}
	}
}
