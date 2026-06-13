package fdkaac

import "testing"

var qcLifecycleSink int

func TestFDKaacEncQCNewVector(t *testing.T) {
	var state QCState
	state.Kernel.GlobHdrBits = 77
	state.AdjThrStateElements[0].PeMin = 99
	state.ElementBits[0].ChBitrateEl = 123

	if errCode := FDKaacEncQCNew(&state, 4); errCode != AACEncOK {
		t.Fatalf("QC new error = %#x, want %#x", errCode, AACEncOK)
	}
	if state.Kernel != (QCKernel{}) || state.BitCounter != (BitCounterState{}) {
		t.Fatalf("QC new did not clear kernel/bit-counter state: %+v / %+v", state.Kernel, state.BitCounter)
	}
	for i := 0; i < 4; i++ {
		if state.AdjThrState.AdjThrStateElem[i] != &state.AdjThrStateElements[i] {
			t.Fatalf("threshold element %d pointer = %p, want %p", i, state.AdjThrState.AdjThrStateElem[i], &state.AdjThrStateElements[i])
		}
		if state.ElementBitsPtr[i] != &state.ElementBits[i] {
			t.Fatalf("element bits %d pointer = %p, want %p", i, state.ElementBitsPtr[i], &state.ElementBits[i])
		}
	}
	if state.AdjThrState.AdjThrStateElem[4] != nil || state.ElementBitsPtr[4] != nil {
		t.Fatalf("QC new linked element past requested count")
	}
}

func TestFDKaacEncQCOutNewVector(t *testing.T) {
	var state QCOutState
	state.QCOut[0].TotalBits = 77
	state.QCOutChannels[0][0].GlobalGain = 88
	state.QCOutElements[0][0].StaticBitsUsed = 99

	if errCode := FDKaacEncQCOutNew(&state, 4, 6, 1); errCode != AACEncOK {
		t.Fatalf("QC output new error = %#x, want %#x", errCode, AACEncOK)
	}
	if state.QCOutPtr[0] != &state.QCOut[0] {
		t.Fatalf("QC output frame pointer = %p, want %p", state.QCOutPtr[0], &state.QCOut[0])
	}
	if state.QCOut[0].TotalBits != 0 || state.QCOutChannels[0][0].GlobalGain != 0 || state.QCOutElements[0][0].StaticBitsUsed != 0 {
		t.Fatalf("QC output new did not clear stale state")
	}
	for ch := 0; ch < 6; ch++ {
		if state.QCOut[0].PQCOutChannels[ch] != &state.QCOutChannels[0][ch] {
			t.Fatalf("QC output channel %d pointer = %p, want %p", ch, state.QCOut[0].PQCOutChannels[ch], &state.QCOutChannels[0][ch])
		}
	}
	if state.QCOut[0].PQCOutChannels[6] != nil {
		t.Fatalf("QC output new linked channel past requested count")
	}
	for i := 0; i < 4; i++ {
		if state.QCOutElementPtr[0][i] != &state.QCOutElements[0][i] {
			t.Fatalf("QC output element ptr %d = %p, want %p", i, state.QCOutElementPtr[0][i], &state.QCOutElements[0][i])
		}
		if state.QCOut[0].QCElement[i] != &state.QCOutElements[0][i] {
			t.Fatalf("QC output element %d = %p, want %p", i, state.QCOut[0].QCElement[i], &state.QCOutElements[0][i])
		}
	}
	if state.QCOut[0].QCElement[4] != nil {
		t.Fatalf("QC output new linked element past requested count")
	}
}

func TestFDKaacEncQCOutInitChannelLinks(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode1_2_2_1, ChannelOrderWAV, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	var state QCOutState
	if errCode := FDKaacEncQCOutNew(&state, cm.NElements, cm.NChannels, 1); errCode != AACEncOK {
		t.Fatalf("QC output new error = %#x, want %#x", errCode, AACEncOK)
	}
	if errCode := FDKaacEncQCOutInit(state.QCOutPtr[:], 1, &cm); errCode != AACEncOK {
		t.Fatalf("QC output init error = %#x, want %#x", errCode, AACEncOK)
	}

	want := [][2]int{{0, -1}, {1, 2}, {3, 4}, {5, -1}}
	for el, channels := range want {
		for ch, global := range channels {
			got := state.QCOut[0].QCElement[el].QCOutChannel[ch]
			if global == -1 {
				if got != nil {
					t.Fatalf("element %d channel %d pointer = %p, want nil", el, ch, got)
				}
				continue
			}
			wantPtr := &state.QCOutChannels[0][global]
			if got != wantPtr {
				t.Fatalf("element %d channel %d pointer = %p, want global channel %d at %p", el, ch, got, global, wantPtr)
			}
		}
	}
}

func TestFDKaacEncQCLifecycleFeedsQCInit(t *testing.T) {
	cfg := baseAACLCConfig(48000, 240000, 6, Mode1_2_2_1)
	cfg.BitrateMode = QCBitrateModeVBR3
	cfg.ChannelOrder = ChannelOrderWAV

	var initState AACEncInitState
	if errCode := FDKaacEncPrepareQCInitFromConfig(&initState, &cfg, staticBits64); errCode != AACEncOK {
		t.Fatalf("prepare QC init error = %#x, want %#x", errCode, AACEncOK)
	}
	var qc QCState
	if errCode := FDKaacEncQCNew(&qc, initState.ChannelMapping.NElements); errCode != AACEncOK {
		t.Fatalf("QC new error = %#x, want %#x", errCode, AACEncOK)
	}
	var qcOut QCOutState
	if errCode := FDKaacEncQCOutNew(&qcOut, initState.ChannelMapping.NElements, initState.ChannelMapping.NChannels, initState.QCInit.NSubFrames); errCode != AACEncOK {
		t.Fatalf("QC output new error = %#x, want %#x", errCode, AACEncOK)
	}
	if errCode := FDKaacEncQCOutInit(qcOut.QCOutPtr[:], initState.QCInit.NSubFrames, &initState.ChannelMapping); errCode != AACEncOK {
		t.Fatalf("QC output init error = %#x, want %#x", errCode, AACEncOK)
	}
	if errCode := FDKaacEncQCInit(&qc.Kernel, &qc.AdjThrState, qc.ElementBitsPtr[:], &initState.QCInit, 1); errCode != AACEncOK {
		t.Fatalf("QC init error = %#x, want %#x", errCode, AACEncOK)
	}

	if qc.Kernel.NElements != 4 ||
		qc.Kernel.BitrateMode != QCBitrateModeVBR3 ||
		qc.Kernel.BitResTot != 30720 ||
		qc.Kernel.MaxBitsPerFrame != 30720 ||
		qc.Kernel.GlobHdrBits != 64 ||
		qc.Kernel.VBRQualFactor != vbrQualFactor3 {
		t.Fatalf("lifecycle QC kernel = %+v", qc.Kernel)
	}
	assertQCLifecycleElementBits(t, qc.ElementBits[:], []elementBitsWant{
		{chBitrate: 57599, maxBits: 5996, rel: relBits024},
		{chBitrate: 41999, maxBits: 11992, rel: relBits035},
		{chBitrate: 41999, maxBits: 11992, rel: relBits035},
		{chBitrate: 14399, maxBits: 736, rel: relBits006},
	})
	if qc.AdjThrState.AdjThrStateElem[3].AHParam.ModifyMinSnr != adjThrFalse {
		t.Fatalf("lifecycle LFE threshold element = %+v", *qc.AdjThrState.AdjThrStateElem[3])
	}
	if qcOut.QCOut[0].QCElement[3].QCOutChannel[0] != &qcOut.QCOutChannels[0][5] {
		t.Fatalf("lifecycle LFE channel link = %p, want %p", qcOut.QCOut[0].QCElement[3].QCOutChannel[0], &qcOut.QCOutChannels[0][5])
	}
}

func TestFDKaacEncQCLifecycleRejectsInvalid(t *testing.T) {
	var qc QCState
	if got := FDKaacEncQCNew(nil, 1); got != AACEncInvalidHandle {
		t.Fatalf("nil QC state error = %#x, want %#x", got, AACEncInvalidHandle)
	}
	if got := FDKaacEncQCNew(&qc, maxChannelElements+1); got != AACEncNoMemory {
		t.Fatalf("oversize QC state error = %#x, want %#x", got, AACEncNoMemory)
	}
	expectAACEncPanic(t, func() {
		_ = FDKaacEncQCNew(&qc, -1)
	})

	var out QCOutState
	if got := FDKaacEncQCOutNew(nil, 1, 1, 1); got != AACEncInvalidHandle {
		t.Fatalf("nil QC output state error = %#x, want %#x", got, AACEncInvalidHandle)
	}
	if got := FDKaacEncQCOutNew(&out, 1, maxAACChannels+1, 1); got != AACEncNoMemory {
		t.Fatalf("oversize QC output error = %#x, want %#x", got, AACEncNoMemory)
	}
	expectAACEncPanic(t, func() {
		_ = FDKaacEncQCOutNew(&out, 1, 1, -1)
	})
	expectAACEncPanic(t, func() {
		_ = FDKaacEncQCOutInit(out.QCOutPtr[:], 1, nil)
	})

	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	if errCode := FDKaacEncQCOutNew(&out, cm.NElements, cm.NChannels, 1); errCode != AACEncOK {
		t.Fatalf("QC output new error = %#x, want %#x", errCode, AACEncOK)
	}
	out.QCOut[0].PQCOutChannels[1] = nil
	expectAACEncPanic(t, func() {
		_ = FDKaacEncQCOutInit(out.QCOutPtr[:], 1, &cm)
	})
}

func TestFDKaacEncQCLifecycleAllocs(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode1_2_2_1, ChannelOrderWAV, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	var qc QCState
	var out QCOutState
	allocs := testing.AllocsPerRun(1000, func() {
		errCode := FDKaacEncQCNew(&qc, cm.NElements)
		errCode += FDKaacEncQCOutNew(&out, cm.NElements, cm.NChannels, 1)
		errCode += FDKaacEncQCOutInit(out.QCOutPtr[:], 1, &cm)
		qcLifecycleSink = errCode + len(out.QCOut[0].PQCOutChannels) + len(qc.ElementBitsPtr)
	})
	if allocs != 0 {
		t.Fatalf("QC lifecycle allocations = %v, want 0", allocs)
	}
}

func assertQCLifecycleElementBits(t *testing.T, got []ElementBits, want []elementBitsWant) {
	t.Helper()
	for i, w := range want {
		if got[i].ChBitrateEl != w.chBitrate ||
			got[i].MaxBitsEl != w.maxBits ||
			got[i].RelativeBitsEl != w.rel {
			t.Fatalf("element bits %d = %+v, want chBitrate %d maxBits %d rel %#x",
				i, got[i], w.chBitrate, w.maxBits, int32(w.rel))
		}
		if got[i].BitResLevelEl != 0 || got[i].MaxBitResBitsEl != 0 {
			t.Fatalf("fresh element bits %d reservoir fields = %d/%d, want 0/0",
				i, got[i].BitResLevelEl, got[i].MaxBitResBitsEl)
		}
	}
}
