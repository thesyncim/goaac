package fdkaac

const aacEncRawFrameBufferBytes = 8192

type AACEncFrameState struct {
	Config     AACEncConfig
	Init       AACEncInitState
	Psy        PsyInternal
	PsyOut     PsyOutState
	QC         QCState
	QCOut      QCOutState
	StaticBits FDKaacEncStaticBitsFunc
	Ready      bool
}

type AACEncFrameScratch struct {
	PsyMain   PsyMainScratch
	QCMain    QCMainFrameScratch
	QCElement [1][maxChannelElements]*QCOutElement
	BitBuffer [aacEncRawFrameBufferBytes]byte
	Fetch     [aacEncRawFrameBufferBytes]byte
	BitStream BitStream
}

type AACEncFrameResult struct {
	AvgTotalBits        int
	TransportStaticBits int
	QCMain              QCMainFrameResult
	Write               WriteBitstreamResult
	PayloadBytes        int
	TotalBits           int
	UsedDynBits         int
	FillBits            int
	AlignBits           int
	BitReservoir        int
}

func FDKaacEncInitRawFrameState(state *AACEncFrameState, config AACEncConfig) int {
	return FDKaacEncInitFrameState(state, config, fdkaacEncStaticBitsZero)
}

func FDKaacEncInitFrameState(state *AACEncFrameState, config AACEncConfig, staticBits FDKaacEncStaticBitsFunc) int {
	if state == nil {
		return AACEncInvalidHandle
	}
	if staticBits == nil {
		staticBits = fdkaacEncStaticBitsZero
	}

	*state = AACEncFrameState{Config: config, StaticBits: staticBits}
	if state.Config.NSubFrames == 0 {
		state.Config.NSubFrames = 1
	}

	if errCode := FDKaacEncPrepareQCInitFromConfig(&state.Init, &state.Config, staticBits); errCode != AACEncOK {
		return errCode
	}

	cm := &state.Init.ChannelMapping
	if errCode := FDKaacEncPsyNew(&state.Psy, cm.NElements, cm.NChannels); errCode != AACEncOK {
		return errCode
	}
	if errCode := FDKaacEncPsyOutNew(&state.PsyOut, cm.NElements, cm.NChannels, state.Config.NSubFrames); errCode != AACEncOK {
		return errCode
	}
	if errCode := FDKaacEncPsyInit(&state.Psy, state.PsyOut.PsyOutPtr[:], state.Config.NSubFrames, cm.NChannels, state.Config.AOT, cm); errCode != AACEncOK {
		return errCode
	}
	if errCode := FDKaacEncPsyMainInit(
		&state.Psy,
		state.Config.AOT,
		cm,
		state.Config.SampleRate,
		state.Config.FrameLength,
		state.Init.PsyBitrate,
		state.Init.TNSMask,
		state.Config.BandWidth,
		state.Config.UsePns,
		state.Config.UseIS,
		state.Config.UseMS,
		state.Config.SyntaxFlags,
		1,
	); errCode != AACEncOK {
		return errCode
	}

	if errCode := FDKaacEncQCNew(&state.QC, cm.NElements); errCode != AACEncOK {
		return errCode
	}
	if errCode := FDKaacEncQCOutNew(&state.QCOut, cm.NElements, cm.NChannels, state.Config.NSubFrames); errCode != AACEncOK {
		return errCode
	}
	if errCode := FDKaacEncQCOutInit(state.QCOut.QCOutPtr[:], state.Config.NSubFrames, cm); errCode != AACEncOK {
		return errCode
	}
	if errCode := FDKaacEncQCInit(&state.QC.Kernel, &state.QC.AdjThrState, state.QC.ElementBitsPtr[:], &state.Init.QCInit, 1); errCode != AACEncOK {
		return errCode
	}

	state.Ready = true
	return AACEncOK
}

func FDKaacEncEncodeFrameRaw(
	state *AACEncFrameState,
	dst []byte,
	input []int16,
	inputBufSize int,
	scratch *AACEncFrameScratch,
) ([]byte, AACEncFrameResult, int) {
	checkEncodeFrameRawInputs(state, input, inputBufSize, scratch)

	cm := &state.Init.ChannelMapping
	psyOut := state.PsyOut.PsyOutPtr[0]
	qcOut := state.QCOut.QCOutPtr[0]
	fdkaacEncResetFrameOutputs(psyOut, qcOut, cm)

	result := AACEncFrameResult{}
	for el := 0; el < cm.NElements; el++ {
		elInfo := &cm.ElInfo[el]
		if elInfo.ElType != idSCE && elInfo.ElType != idCPE && elInfo.ElType != idLFE {
			return dst, result, AACEncInvalidElementInfoType
		}

		nChannels := elInfo.NChannelsInEl
		psyOutElement := psyOut.PsyOutElement[el]
		qcOutElement := qcOut.QCElement[el]
		for ch := 0; ch < nChannels; ch++ {
			psyOutChannel := psyOutElement.PsyOutChannel[ch]
			qcOutChannel := qcOutElement.QCOutChannel[ch]
			psyOutChannel.MdctSpectrum = qcOutChannel.MdctSpectrum[:state.Config.FrameLength]
		}

		if errCode := FDKaacEncPsyMain(
			nChannels,
			state.Psy.PsyElement[el],
			state.Psy.PsyDynamic,
			&state.Psy.PsyConf,
			psyOutElement,
			input,
			inputBufSize,
			elInfo.ChannelIndex[:],
			cm.NChannels,
			&scratch.PsyMain,
		); errCode != AACEncOK {
			return dst, result, errCode
		}

		for ch := 0; ch < nChannels; ch++ {
			fdkaacEncMirrorPsyToQC(psyOutElement.PsyOutChannel[ch], qcOutElement.QCOutChannel[ch])
		}

		if errCode := FDKaacEncQCMainPrepare(
			elInfo,
			state.QC.AdjThrState.AdjThrStateElem[el],
			psyOutElement,
			qcOutElement,
			int(state.Init.InternalAOT),
			state.Config.SyntaxFlags,
			state.Config.EpConfig,
		); errCode != AACEncOK {
			return dst, result, errCode
		}

		qcOut.ElementExtBits += qcOutElement.ExtBitsUsed
		qcOut.StaticBits += qcOutElement.StaticBitsUsed
		qcOut.TotalNoRedPe += int(qcOutElement.PEData.Pe)
	}

	if state.Config.SyntaxFlags&(acScalable|acER) == 0 {
		qcOut.GlobalExtBits += elIDBits
	}

	avgTotalBits := 0
	if errCode := FDKaacEncAdjustBitrate(&state.QC.Kernel, cm, &avgTotalBits, state.Config.BitRate, state.Config.SampleRate, state.Config.FrameLength); errCode != AACEncOK {
		return dst, result, errCode
	}
	result.AvgTotalBits = avgTotalBits

	state.QC.Kernel.GlobHdrBits = fdkaacEncStaticBits(state.StaticBits, avgTotalBits+state.QC.Kernel.BitResTot)
	result.TransportStaticBits = state.QC.Kernel.GlobHdrBits

	clear(scratch.QCElement[:])
	for el := 0; el < cm.NElements; el++ {
		scratch.QCElement[0][el] = qcOut.QCElement[el]
	}
	qcMain, errCode := FDKaacEncQCMainFrame(
		&state.QC.Kernel,
		&state.QC.AdjThrState,
		psyOut.PsyOutElement[:],
		state.QCOut.QCOutPtr[:],
		scratch.QCElement[:],
		cm,
		state.QC.ElementBitsPtr[:],
		&scratch.QCMain,
		avgTotalBits,
		state.QC.Kernel.InvQuant,
		state.QC.Kernel.DZoneQuantEnable,
		state.QC.Kernel.MaxIterations,
		int(state.Init.InternalAOT),
		state.Config.SyntaxFlags,
		state.Config.EpConfig,
	)
	result.QCMain = qcMain
	if errCode != AACEncOK {
		return dst, result, errCode
	}

	if errCode := FDKaacEncUpdateFillBits(&state.QC.Kernel, state.QCOut.QCOutPtr[:]); errCode != AACEncOK {
		return dst, result, errCode
	}

	exactStaticBits := fdkaacEncStaticBits(state.StaticBits, qcOut.TotalBits+state.QC.Kernel.GlobHdrBits)
	result.TransportStaticBits = exactStaticBits
	if errCode := FDKaacEncFinalizeBitConsumption(
		&state.QC.Kernel,
		qcOut,
		exactStaticBits,
		int(state.Init.InternalAOT),
		state.Config.SyntaxFlags,
		state.Config.EpConfig,
	); errCode != AACEncOK {
		return dst, result, errCode
	}
	FDKaacEncUpdateBitres(&state.QC.Kernel, state.QCOut.QCOutPtr[:])

	clear(scratch.BitBuffer[:])
	if err := InitBitStream(&scratch.BitStream, scratch.BitBuffer[:], 0, BSWriter); err != nil {
		return dst, result, AACEncInvalidHandle
	}
	writeResult, errCode := FDKaacEncWriteBitstream(
		&scratch.BitStream,
		cm,
		qcOut,
		qcOut.QCElement[:cm.NElements],
		psyOut.PsyOutElement[:cm.NElements],
		&state.QC.Kernel,
		int(state.Init.InternalAOT),
		state.Config.SyntaxFlags,
		state.Config.EpConfig,
	)
	result.Write = writeResult
	if errCode != AACEncOK {
		return dst, result, errCode
	}

	n := FetchBuffer(&scratch.BitStream, scratch.Fetch[:])
	result.PayloadBytes = n
	result.TotalBits = qcOut.TotalBits
	result.UsedDynBits = qcOut.UsedDynBits
	result.FillBits = qcOut.TotFillBits
	result.AlignBits = qcOut.AlignBits
	result.BitReservoir = state.QC.Kernel.BitResTot
	return append(dst, scratch.Fetch[:n]...), result, AACEncOK
}

func fdkaacEncResetFrameOutputs(psyOut *PsyOut, qcOut *QCOut, cm *ChannelMapping) {
	qcChannels := qcOut.PQCOutChannels
	qcElements := qcOut.QCElement
	*qcOut = QCOut{}
	qcOut.PQCOutChannels = qcChannels
	qcOut.QCElement = qcElements

	for el := 0; el < cm.NElements; el++ {
		nChannels := cm.ElInfo[el].NChannelsInEl
		psyOutElement := psyOut.PsyOutElement[el]
		qcOutElement := qcOut.QCElement[el]

		psyChannels := psyOutElement.PsyOutChannel
		qcElementChannels := qcOutElement.QCOutChannel
		*psyOutElement = PsyOutElement{}
		*qcOutElement = QCOutElement{}
		psyOutElement.PsyOutChannel = psyChannels
		qcOutElement.QCOutChannel = qcElementChannels

		for ch := 0; ch < nChannels; ch++ {
			*psyOutElement.PsyOutChannel[ch] = PsyOutChannel{}
			*qcOutElement.QCOutChannel[ch] = QCOutChannel{}
		}
	}
}

func fdkaacEncMirrorPsyToQC(psy *PsyOutChannel, qc *QCOutChannel) {
	copy(qc.SfbEnergy[:], psy.SfbEnergy[:])
	copy(qc.SfbEnergyLdData[:], psy.SfbEnergyLdData[:])
	copy(qc.SfbThresholdLdData[:], psy.SfbThresholdLdData[:])
	copy(qc.SfbMinSnrLdData[:], psy.SfbMinSnrLdData[:])
	copy(qc.SfbSpreadEnergy[:], psy.SfbSpreadEnergy[:])
}

func fdkaacEncStaticBitsZero(int) int {
	return 0
}

func checkEncodeFrameRawInputs(state *AACEncFrameState, input []int16, inputBufSize int, scratch *AACEncFrameScratch) {
	if state == nil || scratch == nil {
		panic("fdkaac: nil encoder frame state")
	}
	if !state.Ready || state.Psy.PsyDynamic == nil {
		panic("fdkaac: uninitialized encoder frame state")
	}
	if state.Config.NSubFrames != 1 {
		panic("fdkaac: raw frame encoder supports one subframe")
	}
	if state.Config.FrameLength <= 0 || state.Config.FrameLength > maxSpectralLines {
		panic("fdkaac: invalid raw frame length")
	}
	if inputBufSize < state.Config.FrameLength {
		panic("fdkaac: short raw frame input stride")
	}
	cm := &state.Init.ChannelMapping
	if cm.NElements <= 0 || cm.NElements > maxChannelElements || cm.NChannels <= 0 || cm.NChannels > maxAACChannels {
		panic("fdkaac: invalid raw frame channel mapping")
	}
	if len(input) < inputBufSize*cm.NChannels {
		panic("fdkaac: short raw frame input")
	}
	for el := 0; el < cm.NElements; el++ {
		psyOutElement := state.PsyOut.PsyOutPtr[0].PsyOutElement[el]
		qcOutElement := state.QCOut.QCOutPtr[0].QCElement[el]
		if state.Psy.PsyElement[el] == nil || state.QC.AdjThrState.AdjThrStateElem[el] == nil ||
			psyOutElement == nil || qcOutElement == nil || state.QC.ElementBitsPtr[el] == nil {
			panic("fdkaac: incomplete raw frame element state")
		}
		nChannels := cm.ElInfo[el].NChannelsInEl
		if nChannels != channelElementCount(cm.ElInfo[el].ElType) {
			panic("fdkaac: invalid raw frame element channels")
		}
		for ch := 0; ch < nChannels; ch++ {
			if psyOutElement.PsyOutChannel[ch] == nil || qcOutElement.QCOutChannel[ch] == nil {
				panic("fdkaac: incomplete raw frame channel state")
			}
			idx := cm.ElInfo[el].ChannelIndex[ch]
			if idx < 0 || idx >= cm.NChannels || (idx+1)*inputBufSize > len(input) {
				panic("fdkaac: invalid raw frame channel index")
			}
		}
	}
}
