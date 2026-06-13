package fdkaac

import "testing"

var psyLifecycleSink int

func TestFDKaacEncPsyNewVector(t *testing.T) {
	var state PsyInternal
	state.GranuleLength = 77
	state.StaticChannels[0].PsyInputBuffer[4] = 123
	state.PsyElements[0].PsyStatic[0] = &state.StaticChannels[7]

	if errCode := FDKaacEncPsyNew(&state, 4, 6); errCode != AACEncOK {
		t.Fatalf("psy new error = %#x, want %#x", errCode, AACEncOK)
	}
	if state.GranuleLength != 0 || state.StaticChannels[0].PsyInputBuffer[4] != 0 {
		t.Fatalf("psy new did not clear stale state")
	}
	if state.PsyDynamic != &state.Dynamic {
		t.Fatalf("dynamic psy state = %p, want %p", state.PsyDynamic, &state.Dynamic)
	}
	for i := 0; i < 4; i++ {
		if state.PsyElement[i] != &state.PsyElements[i] {
			t.Fatalf("psy element %d pointer = %p, want %p", i, state.PsyElement[i], &state.PsyElements[i])
		}
	}
	if state.PsyElement[4] != nil {
		t.Fatalf("psy new linked element past requested count")
	}
	for ch := 0; ch < 6; ch++ {
		if state.PStaticChannel[ch] != &state.StaticChannels[ch] {
			t.Fatalf("psy static channel %d pointer = %p, want %p", ch, state.PStaticChannel[ch], &state.StaticChannels[ch])
		}
	}
	if state.PStaticChannel[6] != nil {
		t.Fatalf("psy new linked static channel past requested count")
	}
}

func TestFDKaacEncPsyOutNewVector(t *testing.T) {
	var state PsyOutState
	state.PsyOutChannels[0][0].SfbCnt = 7
	state.PsyOutElements[0][0].CommonWindow = 1

	if errCode := FDKaacEncPsyOutNew(&state, 4, 6, 1); errCode != AACEncOK {
		t.Fatalf("psy output new error = %#x, want %#x", errCode, AACEncOK)
	}
	if state.PsyOutPtr[0] != &state.PsyOut[0] {
		t.Fatalf("psy output frame pointer = %p, want %p", state.PsyOutPtr[0], &state.PsyOut[0])
	}
	if state.PsyOutChannels[0][0].SfbCnt != 0 || state.PsyOutElements[0][0].CommonWindow != 0 {
		t.Fatalf("psy output new did not clear stale state")
	}
	for ch := 0; ch < 6; ch++ {
		if state.PsyOut[0].PPsyOutChannel[ch] != &state.PsyOutChannels[0][ch] {
			t.Fatalf("psy output channel %d pointer = %p, want %p", ch, state.PsyOut[0].PPsyOutChannel[ch], &state.PsyOutChannels[0][ch])
		}
	}
	if state.PsyOut[0].PPsyOutChannel[6] != nil {
		t.Fatalf("psy output new linked channel past requested count")
	}
	for i := 0; i < 4; i++ {
		if state.PsyOutElementPtr[0][i] != &state.PsyOutElements[0][i] {
			t.Fatalf("psy output element ptr %d = %p, want %p", i, state.PsyOutElementPtr[0][i], &state.PsyOutElements[0][i])
		}
		if state.PsyOut[0].PsyOutElement[i] != &state.PsyOutElements[0][i] {
			t.Fatalf("psy output element %d = %p, want %p", i, state.PsyOut[0].PsyOutElement[i], &state.PsyOutElements[0][i])
		}
	}
	if state.PsyOut[0].PsyOutElement[4] != nil {
		t.Fatalf("psy output new linked element past requested count")
	}
}

func TestFDKaacEncPsyInitStatesVectors(t *testing.T) {
	var state PsyInternal
	var static PsyStatic
	static.PsyInputBuffer[10] = 99
	static.BlockSwitchingControl.Attack = 1
	if errCode := FDKaacEncPsyInitStates(&state, &static, AOTAACLC); errCode != AACEncOK {
		t.Fatalf("psy init states error = %#x, want %#x", errCode, AACEncOK)
	}
	if static.PsyInputBuffer[10] != 0 {
		t.Fatalf("psy init states did not clear input buffer")
	}
	assertPsyLCBlockSwitch(t, &static.BlockSwitchingControl)

	static.PsyInputBuffer[10] = 99
	if errCode := FDKaacEncPsyInitStates(&state, &static, AOTERAACLD); errCode != AACEncOK {
		t.Fatalf("low-delay psy init states error = %#x, want %#x", errCode, AACEncOK)
	}
	if static.BlockSwitchingControl.NBlockSwitchWindows != 4 ||
		static.BlockSwitchingControl.AllowShortFrames != 0 ||
		static.BlockSwitchingControl.AllowLookAhead != 0 ||
		static.BlockSwitchingControl.WindowShape != WindowShapeSine {
		t.Fatalf("low-delay block switching = %+v", static.BlockSwitchingControl)
	}
}

func TestFDKaacEncPsyInitLinksStateAndOutput(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode1_2_2_1, ChannelOrderWAV, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	var psy PsyInternal
	if errCode := FDKaacEncPsyNew(&psy, cm.NElements, cm.NChannels); errCode != AACEncOK {
		t.Fatalf("psy new error = %#x, want %#x", errCode, AACEncOK)
	}
	var psyOut PsyOutState
	if errCode := FDKaacEncPsyOutNew(&psyOut, cm.NElements, cm.NChannels, 1); errCode != AACEncOK {
		t.Fatalf("psy output new error = %#x, want %#x", errCode, AACEncOK)
	}
	psy.StaticChannels[0].PsyInputBuffer[3] = 31
	psy.StaticChannels[2].PsyInputBuffer[3] = 32
	psy.StaticChannels[3].PsyInputBuffer[3] = 33

	if errCode := FDKaacEncPsyInit(&psy, psyOut.PsyOutPtr[:], 1, cm.NChannels, AOTAACLC, &cm); errCode != AACEncOK {
		t.Fatalf("psy init error = %#x, want %#x", errCode, AACEncOK)
	}

	wantStatic := [][2]int{{0, -1}, {1, 2}, {3, 4}, {5, -1}}
	for el, channels := range wantStatic {
		for ch, global := range channels {
			got := psy.PsyElement[el].PsyStatic[ch]
			if global == -1 {
				if got != nil {
					t.Fatalf("element %d static channel %d = %p, want nil", el, ch, got)
				}
				continue
			}
			wantPtr := &psy.StaticChannels[global]
			if got != wantPtr {
				t.Fatalf("element %d static channel %d = %p, want %p", el, ch, got, wantPtr)
			}
		}
	}
	for ch := 0; ch < 5; ch++ {
		if psy.StaticChannels[ch].IsLFE != 0 {
			t.Fatalf("channel %d LFE flag = %d, want 0", ch, psy.StaticChannels[ch].IsLFE)
		}
	}
	for ch := 0; ch < 3; ch++ {
		if psy.StaticChannels[ch].BlockSwitchingControl.NBlockSwitchWindows != 0 {
			t.Fatalf("preserved static channel %d was reset: %+v", ch, psy.StaticChannels[ch].BlockSwitchingControl)
		}
	}
	for ch := 3; ch < 5; ch++ {
		assertPsyLCBlockSwitch(t, &psy.StaticChannels[ch].BlockSwitchingControl)
	}
	if psy.StaticChannels[0].PsyInputBuffer[3] != 31 ||
		psy.StaticChannels[2].PsyInputBuffer[3] != 32 ||
		psy.StaticChannels[3].PsyInputBuffer[3] != 0 {
		t.Fatalf("psy init reset/preserve inputs = %d/%d/%d, want 31/32/0",
			psy.StaticChannels[0].PsyInputBuffer[3],
			psy.StaticChannels[2].PsyInputBuffer[3],
			psy.StaticChannels[3].PsyInputBuffer[3])
	}
	if psy.StaticChannels[5].IsLFE != 1 {
		t.Fatalf("LFE flag = %d, want 1", psy.StaticChannels[5].IsLFE)
	}
	if psy.StaticChannels[5].BlockSwitchingControl.NBlockSwitchWindows != 0 {
		t.Fatalf("LFE block switching was initialized: %+v", psy.StaticChannels[5].BlockSwitchingControl)
	}

	for el, channels := range wantStatic {
		for ch, global := range channels {
			got := psyOut.PsyOut[0].PsyOutElement[el].PsyOutChannel[ch]
			if global == -1 {
				if got != nil {
					t.Fatalf("output element %d channel %d = %p, want nil", el, ch, got)
				}
				continue
			}
			wantPtr := &psyOut.PsyOutChannels[0][global]
			if got != wantPtr {
				t.Fatalf("output element %d channel %d = %p, want %p", el, ch, got, wantPtr)
			}
		}
	}
}

func TestFDKaacEncPsyInitStereoReconfigOffset(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	var psy PsyInternal
	if errCode := FDKaacEncPsyNew(&psy, cm.NElements, 6); errCode != AACEncOK {
		t.Fatalf("psy new error = %#x, want %#x", errCode, AACEncOK)
	}
	var psyOut PsyOutState
	if errCode := FDKaacEncPsyOutNew(&psyOut, cm.NElements, cm.NChannels, 1); errCode != AACEncOK {
		t.Fatalf("psy output new error = %#x, want %#x", errCode, AACEncOK)
	}
	psy.StaticChannels[0].PsyInputBuffer[1] = 10
	psy.StaticChannels[1].PsyInputBuffer[1] = 11
	psy.StaticChannels[2].PsyInputBuffer[1] = 12

	if errCode := FDKaacEncPsyInit(&psy, psyOut.PsyOutPtr[:], 1, 6, AOTAACLC, &cm); errCode != AACEncOK {
		t.Fatalf("psy init error = %#x, want %#x", errCode, AACEncOK)
	}
	if psy.PsyElement[0].PsyStatic[0] != &psy.StaticChannels[1] ||
		psy.PsyElement[0].PsyStatic[1] != &psy.StaticChannels[2] {
		t.Fatalf("stereo reconfig statics = %p/%p, want %p/%p",
			psy.PsyElement[0].PsyStatic[0], psy.PsyElement[0].PsyStatic[1],
			&psy.StaticChannels[1], &psy.StaticChannels[2])
	}
	if psy.StaticChannels[0].PsyInputBuffer[1] != 0 {
		t.Fatalf("offset spare channel was not reset")
	}
	assertPsyLCBlockSwitch(t, &psy.StaticChannels[0].BlockSwitchingControl)
	if psy.StaticChannels[1].PsyInputBuffer[1] != 11 ||
		psy.StaticChannels[2].PsyInputBuffer[1] != 12 ||
		psy.StaticChannels[1].BlockSwitchingControl.NBlockSwitchWindows != 0 ||
		psy.StaticChannels[2].BlockSwitchingControl.NBlockSwitchWindows != 0 {
		t.Fatalf("stereo reconfig preserved-state branch changed channels 1/2")
	}
	if psyOut.PsyOut[0].PsyOutElement[0].PsyOutChannel[0] != &psyOut.PsyOutChannels[0][0] ||
		psyOut.PsyOut[0].PsyOutElement[0].PsyOutChannel[1] != &psyOut.PsyOutChannels[0][1] {
		t.Fatalf("stereo output channels were offset unexpectedly")
	}
}

func TestFDKaacEncPsyMainInitAACLCVectors(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	var psy PsyInternal
	if errCode := FDKaacEncPsyNew(&psy, cm.NElements, cm.NChannels); errCode != AACEncOK {
		t.Fatalf("psy new error = %#x, want %#x", errCode, AACEncOK)
	}
	var psyOut PsyOutState
	if errCode := FDKaacEncPsyOutNew(&psyOut, cm.NElements, cm.NChannels, 1); errCode != AACEncOK {
		t.Fatalf("psy output new error = %#x, want %#x", errCode, AACEncOK)
	}
	if errCode := FDKaacEncPsyInit(&psy, psyOut.PsyOutPtr[:], 1, cm.NChannels, AOTAACLC, &cm); errCode != AACEncOK {
		t.Fatalf("psy init error = %#x, want %#x", errCode, AACEncOK)
	}
	psy.StaticChannels[0].PsyInputBuffer[7] = 91
	psy.StaticChannels[1].BlockSwitchingControl.Attack = 3

	const (
		sampleRate    = 48000
		granuleLength = 1024
		bitRate       = 96000
		bandwidth     = 15500
	)
	if errCode := FDKaacEncPsyMainInit(&psy, AOTAACLC, &cm, sampleRate, granuleLength, bitRate, tnsEnableMask, bandwidth, 1, 1, 1, 0, 0); errCode != AACEncOK {
		t.Fatalf("psy main init error = %#x, want %#x", errCode, AACEncOK)
	}
	if psy.GranuleLength != granuleLength {
		t.Fatalf("granule length = %d, want %d", psy.GranuleLength, granuleLength)
	}
	if psy.StaticChannels[0].PsyInputBuffer[7] != 91 || psy.StaticChannels[1].BlockSwitchingControl.Attack != 3 {
		t.Fatalf("initFlags=0 reset static state unexpectedly")
	}

	wantLong, wantShort := wantAACLCMainInitConfigs(t, &cm, sampleRate, granuleLength, bitRate, tnsEnableMask, bandwidth, 1, 1, 1, 0)
	if psy.PsyConf[0] != wantLong {
		t.Fatalf("long psy main config did not match direct source sequence")
	}
	if psy.PsyConf[1] != wantShort {
		t.Fatalf("short psy main config did not match direct source sequence")
	}
	if psy.PsyConf[0].TnsConf.TnsActive != 1 || psy.PsyConf[1].TnsConf.TnsActive != 1 {
		t.Fatalf("TNS active flags = %d/%d, want 1/1", psy.PsyConf[0].TnsConf.TnsActive, psy.PsyConf[1].TnsConf.TnsActive)
	}
	if psy.PsyConf[0].PnsConf.UsePns != 1 || psy.PsyConf[1].PnsConf.UsePns != 1 {
		t.Fatalf("PNS use flags = %d/%d, want 1/1", psy.PsyConf[0].PnsConf.UsePns, psy.PsyConf[1].PnsConf.UsePns)
	}
	if got, want := hashFixpSGL(psy.PsyConf[0].PnsConf.NP.PowDistPSDcurve[:psy.PsyConf[0].SfbCnt+1]), uint64(0x1445dfcb53e1d95b); got != want {
		t.Fatalf("long PNS PSD hash = %#016x, want %#016x", got, want)
	}

	for ch := 0; ch < cm.ElInfo[0].NChannelsInEl; ch++ {
		static := psy.PsyElement[0].PsyStatic[ch]
		if static.CalcPreEcho != 1 || static.MdctScaleNm1 != pcmQuantThrScale>>1 {
			t.Fatalf("pre-echo state ch%d = calc:%d scale:%d", ch, static.CalcPreEcho, static.MdctScaleNm1)
		}
		if hashFixpDBL(static.SfbThresholdNm1[:psy.PsyConf[0].SfbCnt]) != hashFixpDBL(psy.PsyConf[0].SfbPcmQuantThreshold[:psy.PsyConf[0].SfbCnt]) {
			t.Fatalf("pre-echo threshold history ch%d does not match PCM thresholds", ch)
		}
	}
}

func TestFDKaacEncPsyMainInitResetAndDisabledToolVectors(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	var psy PsyInternal
	if errCode := FDKaacEncPsyNew(&psy, cm.NElements, cm.NChannels); errCode != AACEncOK {
		t.Fatalf("psy new error = %#x, want %#x", errCode, AACEncOK)
	}
	var psyOut PsyOutState
	if errCode := FDKaacEncPsyOutNew(&psyOut, cm.NElements, cm.NChannels, 1); errCode != AACEncOK {
		t.Fatalf("psy output new error = %#x, want %#x", errCode, AACEncOK)
	}
	if errCode := FDKaacEncPsyInit(&psy, psyOut.PsyOutPtr[:], 1, cm.NChannels, AOTAACLC, &cm); errCode != AACEncOK {
		t.Fatalf("psy init error = %#x, want %#x", errCode, AACEncOK)
	}
	psy.StaticChannels[0].PsyInputBuffer[11] = 77
	psy.StaticChannels[0].BlockSwitchingControl.Attack = 1

	if errCode := FDKaacEncPsyMainInit(&psy, AOTAACLC, &cm, 48000, 1024, 128000, 0, 15500, 0, 0, 0, 0, 1); errCode != AACEncOK {
		t.Fatalf("psy main init error = %#x, want %#x", errCode, AACEncOK)
	}
	if psy.StaticChannels[0].PsyInputBuffer[11] != 0 {
		t.Fatalf("initFlags reset did not clear psy input buffer")
	}
	assertPsyLCBlockSwitch(t, &psy.StaticChannels[0].BlockSwitchingControl)
	if psy.PsyConf[0].AllowIS || psy.PsyConf[0].AllowMS || psy.PsyConf[1].AllowIS || psy.PsyConf[1].AllowMS {
		t.Fatalf("disabled stereo tools were enabled")
	}
	if psy.PsyConf[0].TnsConf.TnsActive != 0 || psy.PsyConf[1].TnsConf.TnsActive != 0 {
		t.Fatalf("disabled TNS active flags = %d/%d", psy.PsyConf[0].TnsConf.TnsActive, psy.PsyConf[1].TnsConf.TnsActive)
	}
	if psy.PsyConf[0].PnsConf.UsePns != 0 || psy.PsyConf[1].PnsConf.UsePns != 0 {
		t.Fatalf("disabled PNS use flags = %d/%d", psy.PsyConf[0].PnsConf.UsePns, psy.PsyConf[1].PnsConf.UsePns)
	}
}

func TestFDKaacEncPsyLifecycleRejectsInvalid(t *testing.T) {
	var psy PsyInternal
	if got := FDKaacEncPsyNew(nil, 1, 1); got != AACEncInvalidHandle {
		t.Fatalf("nil psy state error = %#x, want %#x", got, AACEncInvalidHandle)
	}
	if got := FDKaacEncPsyNew(&psy, maxChannelElements+1, 1); got != AACEncNoMemory {
		t.Fatalf("oversize psy state error = %#x, want %#x", got, AACEncNoMemory)
	}
	expectAACEncPanic(t, func() {
		_ = FDKaacEncPsyNew(&psy, -1, 1)
	})

	var out PsyOutState
	if got := FDKaacEncPsyOutNew(nil, 1, 1, 1); got != AACEncInvalidHandle {
		t.Fatalf("nil psy output state error = %#x, want %#x", got, AACEncInvalidHandle)
	}
	if got := FDKaacEncPsyOutNew(&out, 1, maxAACChannels+1, 1); got != AACEncNoMemory {
		t.Fatalf("oversize psy output error = %#x, want %#x", got, AACEncNoMemory)
	}
	expectAACEncPanic(t, func() {
		_ = FDKaacEncPsyOutNew(&out, 1, 1, -1)
	})
	expectAACEncPanic(t, func() {
		_ = FDKaacEncPsyInit(nil, out.PsyOutPtr[:], 1, 1, AOTAACLC, &ChannelMapping{})
	})

	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	if errCode := FDKaacEncPsyNew(&psy, cm.NElements, cm.NChannels); errCode != AACEncOK {
		t.Fatalf("psy new error = %#x, want %#x", errCode, AACEncOK)
	}
	if errCode := FDKaacEncPsyOutNew(&out, cm.NElements, cm.NChannels, 1); errCode != AACEncOK {
		t.Fatalf("psy output new error = %#x, want %#x", errCode, AACEncOK)
	}
	psy.PStaticChannel[1] = nil
	expectAACEncPanic(t, func() {
		_ = FDKaacEncPsyInit(&psy, out.PsyOutPtr[:], 1, cm.NChannels, AOTAACLC, &cm)
	})

	var validPsy PsyInternal
	if errCode := FDKaacEncPsyNew(&validPsy, cm.NElements, cm.NChannels); errCode != AACEncOK {
		t.Fatalf("psy new error = %#x, want %#x", errCode, AACEncOK)
	}
	if errCode := FDKaacEncPsyInit(&validPsy, out.PsyOutPtr[:], 1, cm.NChannels, AOTAACLC, &cm); errCode != AACEncOK {
		t.Fatalf("psy init error = %#x, want %#x", errCode, AACEncOK)
	}
	if got := FDKaacEncPsyMainInit(&validPsy, AOTAACLC, &cm, 12345, 1024, 96000, tnsEnableMask, 15500, 1, 1, 1, 0, 0); got != AACEncUnsupportedSamplingRate {
		t.Fatalf("unsupported psy main sample rate rc = %#x, want %#x", got, AACEncUnsupportedSamplingRate)
	}
	expectAACEncPanic(t, func() {
		_ = FDKaacEncPsyMainInit(nil, AOTAACLC, &cm, 48000, 1024, 96000, tnsEnableMask, 15500, 1, 1, 1, 0, 0)
	})
	expectAACEncPanic(t, func() {
		_ = FDKaacEncPsyMainInit(&validPsy, AOTAACLC, &cm, 48000, 512, 96000, tnsEnableMask, 15500, 1, 1, 1, 0, 0)
	})
}

func TestFDKaacEncPsyLifecycleAllocs(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode1_2_2_1, ChannelOrderWAV, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	var psy PsyInternal
	var out PsyOutState
	allocs := testing.AllocsPerRun(1000, func() {
		errCode := FDKaacEncPsyNew(&psy, cm.NElements, cm.NChannels)
		errCode += FDKaacEncPsyOutNew(&out, cm.NElements, cm.NChannels, 1)
		errCode += FDKaacEncPsyInit(&psy, out.PsyOutPtr[:], 1, cm.NChannels, AOTAACLC, &cm)
		psyLifecycleSink = errCode + psy.StaticChannels[0].BlockSwitchingControl.NBlockSwitchWindows + len(out.PsyOut[0].PPsyOutChannel)
	})
	if allocs != 0 {
		t.Fatalf("psy lifecycle allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncPsyMainInitAllocs(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	var psy PsyInternal
	var out PsyOutState
	if errCode := FDKaacEncPsyNew(&psy, cm.NElements, cm.NChannels); errCode != AACEncOK {
		t.Fatalf("psy new error = %#x", errCode)
	}
	if errCode := FDKaacEncPsyOutNew(&out, cm.NElements, cm.NChannels, 1); errCode != AACEncOK {
		t.Fatalf("psy output new error = %#x", errCode)
	}
	if errCode := FDKaacEncPsyInit(&psy, out.PsyOutPtr[:], 1, cm.NChannels, AOTAACLC, &cm); errCode != AACEncOK {
		t.Fatalf("psy init error = %#x", errCode)
	}
	allocs := testing.AllocsPerRun(100, func() {
		if errCode := FDKaacEncPsyMainInit(&psy, AOTAACLC, &cm, 48000, 1024, 96000, tnsEnableMask, 15500, 1, 1, 1, 0, 1); errCode != AACEncOK {
			t.Fatalf("psy main init error = %#x", errCode)
		}
		psyLifecycleSink += psy.PsyConf[0].SfbCnt + psy.PsyConf[1].SfbCnt
	})
	if allocs != 0 {
		t.Fatalf("FDKaacEncPsyMainInit allocations = %v, want 0", allocs)
	}
}

func wantAACLCMainInitConfigs(t *testing.T, cm *ChannelMapping, sampleRate int, granuleLength int, bitRate int, tnsMask int, bandwidth int, usePns int, useIS int, useMS int, syntaxFlags uint32) (PsyConfiguration, PsyConfiguration) {
	t.Helper()
	channelsEff := cm.NChannelsEff
	tnsChannels := 2
	if FDKaacEncGetMonoStereoMode(cm.EncMode) == ElementModeMono {
		tnsChannels = 1
	}
	ldSbrPresent := 0
	if syntaxFlags&acSBRPresent != 0 {
		ldSbrPresent = 1
	}

	var long PsyConfiguration
	if got := FDKaacEncInitPsyConfiguration(bitRate/channelsEff, sampleRate, bandwidth, LongWindow, granuleLength, useIS, useMS, &long, FilterbankLC); got != AACEncOK {
		t.Fatalf("direct long psy config rc = %#x", got)
	}
	if got := FDKaacEncInitTnsConfiguration((bitRate*tnsChannels)/channelsEff, sampleRate, tnsChannels, LongWindow, granuleLength, 0, ldSbrPresent, &long.TnsConf, &long, tnsMask&2, tnsMask&8); got != AACEncOK {
		t.Fatalf("direct long TNS config rc = %#x", got)
	}
	if got := FDKaacEncInitPnsConfiguration(&long.PnsConf, bitRate/channelsEff, sampleRate, usePns, long.SfbCnt, long.SfbOffset[:], cm.ElInfo[0].NChannelsInEl, 1); got != AACEncOK {
		t.Fatalf("direct long PNS config rc = %#x", got)
	}

	var short PsyConfiguration
	if granuleLength > 512 {
		if got := FDKaacEncInitPsyConfiguration(bitRate/channelsEff, sampleRate, bandwidth, ShortWindow, granuleLength, useIS, useMS, &short, FilterbankLC); got != AACEncOK {
			t.Fatalf("direct short psy config rc = %#x", got)
		}
		if got := FDKaacEncInitTnsConfiguration((bitRate*tnsChannels)/channelsEff, sampleRate, tnsChannels, ShortWindow, granuleLength, 0, ldSbrPresent, &short.TnsConf, &short, tnsMask&1, tnsMask&4); got != AACEncOK {
			t.Fatalf("direct short TNS config rc = %#x", got)
		}
		if got := FDKaacEncInitPnsConfiguration(&short.PnsConf, bitRate/channelsEff, sampleRate, usePns, short.SfbCnt, short.SfbOffset[:], cm.ElInfo[1].NChannelsInEl, 1); got != AACEncOK {
			t.Fatalf("direct short PNS config rc = %#x", got)
		}
	}
	return long, short
}

func assertPsyLCBlockSwitch(t *testing.T, got *BlockSwitchingControl) {
	t.Helper()
	if got.NBlockSwitchWindows != 8 ||
		got.AllowShortFrames != 1 ||
		got.AllowLookAhead != 1 ||
		got.NoOfGroups != maxNoOfGroups ||
		got.LastWindowSequence != LongWindow ||
		got.WindowShape != WindowShapeKBD {
		t.Fatalf("LC block switching = %+v", *got)
	}
}
