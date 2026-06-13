package fdkaac

import "testing"

var adjThrSink FixpDBL
var adjThrHashSink uint64

func TestFDKaacEncPECalculationLongPatchVector(t *testing.T) {
	peData, psy, qc, tools, state := buildAdjThrLongPatchCase()
	FDKaacEncPECalculation(&peData, psy[:], qc[:], &tools, &state, 1)

	gotTotals := [...]FixpDBL{peData.Pe, peData.ConstPart, peData.NActiveLines, state.ChaosMeasureEnFac[0], FixpDBL(state.LastEnFacPatch[0])}
	wantTotals := [...]FixpDBL{205, -125, 26, MaxValDBL, 1}
	wantEnFac := [...]FixpDBL{-130934720, -93434720, -115934720, -138434720, -100934720, -85934720, -55934720, -153434720}
	wantWeighted := [...]FixpDBL{-169065280, -156565280, -164065280, -171565280, -159065280, -154065280, -144065280, -176565280}
	wantThreshold := [...]FixpDBL{-389065280, -406565280, -394065280, -391565280, -399065280, -419065280, -434065280, -386565280}
	assertFixpDBLSlice(t, "long PE totals", gotTotals[:], wantTotals[:], 0xe40bea3822b7fdb7)
	assertFixpDBLSlice(t, "long PE enFac", qc[0].SfbEnFacLd[:8], wantEnFac[:], 0x9031a2f33d0d11d1)
	assertFixpDBLSlice(t, "long PE weighted", qc[0].SfbWeightedEnergyLdData[:8], wantWeighted[:], 0xa3f83c9644d7dc8b)
	assertFixpDBLSlice(t, "long PE threshold", qc[0].SfbThresholdLdData[:8], wantThreshold[:], 0x1bbb56261a694aa5)
}

func TestFDKaacEncPECalculationShortTransitionVector(t *testing.T) {
	peData, psy, qc, tools, state := buildAdjThrShortCase()
	FDKaacEncPECalculation(&peData, psy[:], qc[:], &tools, &state, 1)

	gotTotals := [...]FixpDBL{peData.Pe, peData.ConstPart, peData.NActiveLines, state.ChaosMeasureEnFac[0], FixpDBL(state.LastEnFacPatch[0])}
	wantTotals := [...]FixpDBL{153, -143, 19, chaosMeasureShort, 1}
	wantEnFac := [...]FixpDBL{0, 0, 0, 0, 0, 0, 0, 0}
	wantWeighted := [...]FixpDBL{-300000000, -250000000, -280000000, 0, -260000000, -240000000, -200000000, 0}
	wantThreshold := [...]FixpDBL{-520000000, -500000000, -510000000, -530000000, -500000000, -505000000, -490000000, -540000000}
	assertFixpDBLSlice(t, "short PE totals", gotTotals[:], wantTotals[:], 0x5452a186aa54edbc)
	assertFixpDBLSlice(t, "short PE enFac", qc[0].SfbEnFacLd[:8], wantEnFac[:], 0x0c8210784d8af5a5)
	assertFixpDBLSlice(t, "short PE weighted", qc[0].SfbWeightedEnergyLdData[:8], wantWeighted[:], 0x00c577ce0fd0eaea)
	assertFixpDBLSlice(t, "short PE threshold", qc[0].SfbThresholdLdData[:8], wantThreshold[:], 0x73a0d18b723496d1)
}

func TestFDKaacEncAdjThrInitStateVectors(t *testing.T) {
	var state AdjThrState

	FDKaacEncInitAdjThrState(&state, 1, 0, 0)
	gotSingle := [...]int{
		state.BitDistributionMode,
		state.MaxIter2ndGuess,
		int(state.BRESParamLong.MinBitSave),
		int(state.BRESParamLong.MaxBitSpend),
		int(state.BRESParamShort.MinBitSpend),
		int(state.BRESParamShort.MaxBitSpend),
	}
	wantSingle := [...]int{
		BitDistributionModeInterElement,
		1,
		-0x06666666,
		0x33333333,
		-0x06666668,
		0x40000000,
	}
	if gotSingle != wantSingle {
		t.Fatalf("single-element adjustment state = %v, want %v", gotSingle, wantSingle)
	}

	FDKaacEncInitAdjThrState(&state, 2, 1, 1)
	gotMulti := [...]int{
		state.BitDistributionMode,
		state.MaxIter2ndGuess,
		int(state.BRESParamLong.ClipSaveLow),
		int(state.BRESParamLong.ClipSaveHigh),
		int(state.BRESParamShort.ClipSaveLow),
		int(state.BRESParamShort.ClipSaveHigh),
	}
	wantMulti := [...]int{
		BitDistributionModeIntraElement,
		3,
		0x1999999a,
		0x7999999a,
		0x199999a0,
		0x5fffffff,
	}
	if gotMulti != wantMulti {
		t.Fatalf("multi-element adjustment state = %v, want %v", gotMulti, wantMulti)
	}
}

func TestFDKaacEncInitATSElementVectors(t *testing.T) {
	stereo := ATSElement{
		ChaosMeasureEnFac: [2]FixpDBL{1234, 5678},
		LastEnFacPatch:    [2]int{9, 8},
	}
	FDKaacEncInitATSElement(&stereo, 1000, 64000, 2, 48000, 0, 0, 0, peCorrectionHalf)

	gotStereo := [...]int{
		stereo.PeMin,
		stereo.PeMax,
		stereo.PeOffset,
		stereo.AHParam.ModifyMinSnr,
		stereo.AHParam.StartSfbL,
		stereo.AHParam.StartSfbS,
		int(stereo.ChaosMeasureOld),
		int(stereo.VBRQualFactor),
		int(stereo.MinSNRAdaptParam.MaxRed),
		int(stereo.MinSNRAdaptParam.StartRatio),
		int(stereo.MinSNRAdaptParam.MaxRatio),
		int(stereo.MinSNRAdaptParam.RedRatioFac),
		int(stereo.MinSNRAdaptParam.RedOffs),
		int(stereo.PeCorrectionFactorM),
		stereo.PeCorrectionFactorE,
		stereo.DynBitsLast,
		stereo.PeLast,
		int(stereo.Bits2PeFactorM),
		stereo.Bits2PeFactorE,
		int(stereo.ChaosMeasureEnFac[0]),
		stereo.LastEnFacPatch[0],
	}
	wantStereo := [...]int{
		400,
		600,
		0,
		adjThrTrue,
		15,
		3,
		int(peCorrection03),
		int(peCorrectionHalf),
		int(minSnrAdaptDefaultMaxRed),
		int(minSnrAdaptDefaultStartRatio),
		0,
		int(minSnrAdaptDefaultRedRatioFac),
		int(minSnrAdaptDefaultRedOffs),
		int(peCorrectionHalf),
		1,
		-1,
		0,
		int(adjThrDefaultBits2PEFactorM),
		adjThrDefaultBits2PEFactorE,
		0,
		0,
	}
	if gotStereo != wantStereo {
		t.Fatalf("stereo adjustment element = %v, want %v", gotStereo, wantStereo)
	}

	var lowBitrate ATSElement
	FDKaacEncInitATSElement(&lowBitrate, 1000, 16000, 1, 48000, 0, 0, 0, peCorrectionHalf)
	gotLowBitrate := [...]int{
		lowBitrate.PeMin,
		lowBitrate.PeMax,
		lowBitrate.PeOffset,
		lowBitrate.AHParam.ModifyMinSnr,
		lowBitrate.AHParam.StartSfbL,
		lowBitrate.AHParam.StartSfbS,
		int(lowBitrate.Bits2PeFactorM),
		lowBitrate.Bits2PeFactorE,
	}
	wantLowBitrate := [...]int{
		400,
		600,
		50,
		adjThrFalse,
		0,
		0,
		int(adjThrDefaultBits2PEFactorM),
		adjThrDefaultBits2PEFactorE,
	}
	if gotLowBitrate != wantLowBitrate {
		t.Fatalf("low-bitrate adjustment element = %v, want %v", gotLowBitrate, wantLowBitrate)
	}

	defaultM, defaultE := FDKaacEncInitBits2PeFactor(128000, 3, 48000, 1, 1, 1)
	if defaultM != adjThrDefaultBits2PEFactorM || defaultE != adjThrDefaultBits2PEFactorE {
		t.Fatalf("default bits-to-PE = (%d, %d), want (%d, %d)", defaultM, defaultE, adjThrDefaultBits2PEFactorM, adjThrDefaultBits2PEFactorE)
	}
}

func TestFDKaacEncAdjustThresholdsCBRSingleElementVector(t *testing.T) {
	const desiredPe = 160

	var directPE PEData
	var directPsy PsyOutChannel
	var directQC QCOutChannel
	var directTools ToolsInfo
	var directElemState ATSElement
	var directAdjState AdjThrState
	var directQCElement QCOutElement
	var directQCOut QCOut
	var directPsyElement PsyOutElement
	var directCM ChannelMapping
	fillAdjustThresholdsCBRCase(
		&directPE,
		&directPsy,
		&directQC,
		&directTools,
		&directElemState,
		&directAdjState,
		&directQCElement,
		&directQCOut,
		&directPsyElement,
		&directCM,
		desiredPe,
	)
	var directScratch AdaptThresholdsToPeScratch
	directResult := FDKaacEncAdaptThresholdsToPeCBR(
		&directQCElement.PEData,
		directQCElement.QCOutChannel[:],
		directPsyElement.PsyOutChannel[:],
		&directPsyElement.ToolsInfo,
		directAdjState.AdjThrStateElem[0],
		&directScratch,
		1,
		desiredPe,
		directAdjState.MaxIter2ndGuess,
	)
	directQCElements := [1]*QCOutElement{&directQCElement}
	directPsyElements := [1]*PsyOutElement{&directPsyElement}
	fdkaacEncUnweightThresholds(directQCElements[:], directPsyElements[:], &directCM)

	var wrappedPE PEData
	var wrappedPsy PsyOutChannel
	var wrappedQC QCOutChannel
	var wrappedTools ToolsInfo
	var wrappedElemState ATSElement
	var wrappedAdjState AdjThrState
	var wrappedQCElement QCOutElement
	var wrappedQCOut QCOut
	var wrappedPsyElement PsyOutElement
	var wrappedCM ChannelMapping
	fillAdjustThresholdsCBRCase(
		&wrappedPE,
		&wrappedPsy,
		&wrappedQC,
		&wrappedTools,
		&wrappedElemState,
		&wrappedAdjState,
		&wrappedQCElement,
		&wrappedQCOut,
		&wrappedPsyElement,
		&wrappedCM,
		desiredPe,
	)
	var wrappedScratch AdjustThresholdsScratch
	wrappedQCElements := [1]*QCOutElement{&wrappedQCElement}
	wrappedPsyElements := [1]*PsyOutElement{&wrappedPsyElement}
	wrappedResult := FDKaacEncAdjustThresholds(
		&wrappedAdjState,
		wrappedQCElements[:],
		&wrappedQCOut,
		wrappedPsyElements[:],
		1,
		&wrappedCM,
		&wrappedScratch,
	)

	if wrappedResult.AdaptedElements != 1 || wrappedResult.Iterations != directResult.Iterations || wrappedResult.RedPe != directResult.RedPe {
		t.Fatalf("wrapped CBR result = %+v, direct = %+v", wrappedResult, directResult)
	}
	if wrappedQCElement.PEData.Pe != directQCElement.PEData.Pe ||
		hashFixpDBL(wrappedQC.SfbThresholdLdData[:8]) != hashFixpDBL(directQC.SfbThresholdLdData[:8]) ||
		hashFixpDBL(wrappedQC.SfbMinSnrLdData[:8]) != hashFixpDBL(directQC.SfbMinSnrLdData[:8]) {
		t.Fatalf("wrapped CBR state diverged: pe %d/%d threshold %#016x/%#016x minsnr %#016x/%#016x",
			wrappedQCElement.PEData.Pe,
			directQCElement.PEData.Pe,
			hashFixpDBL(wrappedQC.SfbThresholdLdData[:8]),
			hashFixpDBL(directQC.SfbThresholdLdData[:8]),
			hashFixpDBL(wrappedQC.SfbMinSnrLdData[:8]),
			hashFixpDBL(directQC.SfbMinSnrLdData[:8]),
		)
	}
}

func TestFDKaacEncAdjustThresholdsCBRInterElementMultiVector(t *testing.T) {
	var pe0, pe1 PEData
	var psy0, psy1 PsyOutChannel
	var qc0, qc1 QCOutChannel
	var tools0, tools1 ToolsInfo
	var elemState0, elemState1 ATSElement
	var adjState0, adjState1 AdjThrState
	var qcElement0, qcElement1 QCOutElement
	var qcOut0, qcOut1 QCOut
	var psyElement0, psyElement1 PsyOutElement
	var cm0, cm1 ChannelMapping
	fillAdjustThresholdsCBRCase(&pe0, &psy0, &qc0, &tools0, &elemState0, &adjState0, &qcElement0, &qcOut0, &psyElement0, &cm0, 160)
	fillAdjustThresholdsCBRCase(&pe1, &psy1, &qc1, &tools1, &elemState1, &adjState1, &qcElement1, &qcOut1, &psyElement1, &cm1, 160)

	adjState0.AdjThrStateElem[1] = &elemState1
	cm0.NElements = 2
	cm0.ElInfo[0].ChannelIndex[0] = 0
	cm0.ElInfo[1] = ElementInfo{ElType: idSCE, NChannelsInEl: 1, ChannelIndex: [2]int{1, 0}}
	qcOut0.TotalNoRedPe = int(qcElement0.PEData.Pe + qcElement1.PEData.Pe)
	qcOut0.TotalGrantedPeCorr = 260
	qcElements := [2]*QCOutElement{&qcElement0, &qcElement1}
	psyElements := [2]*PsyOutElement{&psyElement0, &psyElement1}
	var scratch AdjustThresholdsScratch

	result := FDKaacEncAdjustThresholds(&adjState0, qcElements[:], &qcOut0, psyElements[:], 1, &cm0, &scratch)
	want := AdjustThresholdsResult{AdaptedElements: 2, RedPe: 260, LastReductionValueM: 154368336}
	if result != want {
		t.Fatalf("multi-element inter-CBR result = %+v, want %+v", result, want)
	}
	if got, want := hashFixpDBL(qc0.SfbThresholdLdData[:8]), uint64(0xdc3cf7e4216d9d7b); got != want {
		t.Fatalf("element0 threshold hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(qc1.SfbThresholdLdData[:8]), uint64(0xdc3cf7e4216d9d7b); got != want {
		t.Fatalf("element1 threshold hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(qc0.SfbMinSnrLdData[:8]), uint64(0x0c8210784d8af5a5); got != want {
		t.Fatalf("element0 min-SNR hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(qc1.SfbMinSnrLdData[:8]), uint64(0x0c8210784d8af5a5); got != want {
		t.Fatalf("element1 min-SNR hash = %#016x, want %#016x", got, want)
	}
	if qcElement0.PEData.Pe != 130 || qcElement1.PEData.Pe != 130 {
		t.Fatalf("multi-element PE = %d/%d, want 130/130", qcElement0.PEData.Pe, qcElement1.PEData.Pe)
	}
}

func TestFDKaacEncAdjustThresholdsVBRSingleElementVector(t *testing.T) {
	var directPsy PsyOutChannel
	var directQC QCOutChannel
	var directState ATSElement
	var directTools ToolsInfo
	var directScratch AdaptThresholdsVBRScratch
	fillAdjustThresholdsVBRCase(&directPsy, &directQC, &directState, &directTools)
	directPsyChannels := [1]*PsyOutChannel{&directPsy}
	directQCChannels := [1]*QCOutChannel{&directQC}
	FDKaacEncAdaptThresholdsVBR(directQCChannels[:], directPsyChannels[:], &directTools, &directState, &directScratch, 1)
	directQCElement := QCOutElement{QCOutChannel: [2]*QCOutChannel{&directQC}}
	directPsyElement := PsyOutElement{ToolsInfo: directTools, PsyOutChannel: [2]*PsyOutChannel{&directPsy}}
	directCM := ChannelMapping{NElements: 1}
	directCM.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	directQCElements := [1]*QCOutElement{&directQCElement}
	directPsyElements := [1]*PsyOutElement{&directPsyElement}
	fdkaacEncUnweightThresholds(directQCElements[:], directPsyElements[:], &directCM)

	var wrappedPsy PsyOutChannel
	var wrappedQC QCOutChannel
	var wrappedState ATSElement
	var wrappedTools ToolsInfo
	fillAdjustThresholdsVBRCase(&wrappedPsy, &wrappedQC, &wrappedState, &wrappedTools)
	wrappedAdjState := AdjThrState{BitDistributionMode: BitDistributionModeInterElement, MaxIter2ndGuess: 1}
	wrappedAdjState.AdjThrStateElem[0] = &wrappedState
	wrappedQCElement := QCOutElement{QCOutChannel: [2]*QCOutChannel{&wrappedQC}}
	wrappedQCOut := QCOut{}
	wrappedPsyElement := PsyOutElement{ToolsInfo: wrappedTools, PsyOutChannel: [2]*PsyOutChannel{&wrappedPsy}}
	wrappedCM := ChannelMapping{NElements: 1}
	wrappedCM.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	wrappedQCElements := [1]*QCOutElement{&wrappedQCElement}
	wrappedPsyElements := [1]*PsyOutElement{&wrappedPsyElement}
	var wrappedScratch AdjustThresholdsScratch

	wrappedResult := FDKaacEncAdjustThresholds(
		&wrappedAdjState,
		wrappedQCElements[:],
		&wrappedQCOut,
		wrappedPsyElements[:],
		0,
		&wrappedCM,
		&wrappedScratch,
	)
	if wrappedResult.AdaptedElements != 1 {
		t.Fatalf("wrapped VBR adapted elements = %d, want 1", wrappedResult.AdaptedElements)
	}
	if wrappedState.ChaosMeasureOld != directState.ChaosMeasureOld ||
		hashFixpDBL(wrappedQC.SfbThresholdLdData[:8]) != hashFixpDBL(directQC.SfbThresholdLdData[:8]) {
		t.Fatalf("wrapped VBR state diverged: chaos %d/%d threshold %#016x/%#016x",
			wrappedState.ChaosMeasureOld,
			directState.ChaosMeasureOld,
			hashFixpDBL(wrappedQC.SfbThresholdLdData[:8]),
			hashFixpDBL(directQC.SfbThresholdLdData[:8]),
		)
	}
}

func TestFDKaacEncBitresCalcBitFacVectors(t *testing.T) {
	var state AdjThrState
	FDKaacEncInitBitresState(&state)

	longElem := buildBitresElement()
	longFactor, longFactorE := FDKaacEncBitresCalcBitFac(500, 1200, 430, LongWindow, 720, MaxValDBL, &state, &longElem)
	longGot := [...]int{int(longFactor), longFactorE, longElem.PeMin, longElem.PeMax}
	longWant := [...]int{1008633107, 1, 255, 607}
	if longGot != longWant {
		t.Fatalf("long bitres factor = %v, want %v", longGot, longWant)
	}

	shortElem := buildBitresElement()
	shortElem.PeMin = 160
	shortElem.PeMax = 500
	shortFactor, shortFactorE := FDKaacEncBitresCalcBitFac(900, 1000, 620, ShortWindow, 512, MaxValDBL, &state, &shortElem)
	shortGot := [...]int{int(shortFactor), shortFactorE, shortElem.PeMin, shortElem.PeMax}
	shortWant := [...]int{1610612730, 1, 196, 620}
	if shortGot != shortWant {
		t.Fatalf("short bitres factor = %v, want %v", shortGot, shortWant)
	}
}

func TestFDKaacEncDistributeBitsVectors(t *testing.T) {
	for _, tt := range []struct {
		name          string
		mode          BitresMode
		nChannels     int
		window        [2]int
		pe            int
		grantedDyn    int
		bitresBits    int
		maxBitresBits int
		element       ATSElement
		want          [8]int
	}{
		{
			name:          "full long",
			mode:          BitresModeFull,
			nChannels:     1,
			window:        [2]int{LongWindow, LongWindow},
			pe:            430,
			grantedDyn:    720,
			bitresBits:    500,
			maxBitresBits: 1200,
			element:       buildBitresElementWithHistory(420, 650),
			want:          [8]int{798, 798, 255, 607, 798, -1, int(peCorrectionHalf), 1},
		},
		{
			name:          "reduced short pair",
			mode:          BitresModeReduced,
			nChannels:     2,
			window:        [2]int{LongWindow, ShortWindow},
			pe:            500,
			grantedDyn:    640,
			bitresBits:    35,
			maxBitresBits: 1200,
			element:       buildBitresElementWithHistory(390, 550),
			want:          [8]int{755, 748, 180, 620, 755, -1, 1064152270, 1},
		},
		{
			name:          "disabled zero grant",
			mode:          BitresModeDisabled,
			nChannels:     1,
			window:        [2]int{StopWindow, LongWindow},
			pe:            320,
			grantedDyn:    0,
			bitresBits:    0,
			maxBitresBits: 0,
			element:       buildBitresElementWithHistory(0, -1),
			want:          [8]int{0, 0, 180, 620, 0, -1, int(lowBitresCorrectionMin), 1},
		},
		{
			name:          "full negative grant",
			mode:          BitresModeFull,
			nChannels:     1,
			window:        [2]int{LongWindow, LongWindow},
			pe:            320,
			grantedDyn:    -56,
			bitresBits:    0,
			maxBitresBits: 1200,
			element:       buildBitresElementWithHistory(0, -1),
			want:          [8]int{0, 0, 180, 620, 0, -1, int(peCorrectionHalf), 1},
		},
		{
			name:          "full short pair",
			mode:          BitresModeFull,
			nChannels:     2,
			window:        [2]int{LongWindow, ShortWindow},
			pe:            620,
			grantedDyn:    512,
			bitresBits:    900,
			maxBitresBits: 1000,
			element:       buildBitresElementWithHistoryAndBounds(300, 520, 160, 500),
			want:          [8]int{906, 906, 196, 620, 906, -1, int(peCorrectionHalf), 1},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var state AdjThrState
			FDKaacEncInitBitresState(&state)
			element := tt.element
			psy0 := PsyOutChannel{LastWindowSequence: tt.window[0]}
			psy1 := PsyOutChannel{LastWindowSequence: tt.window[1]}
			psy := [2]*PsyOutChannel{&psy0, &psy1}
			peData := PEData{Pe: FixpDBL(tt.pe)}

			grantedPE, grantedPECorr := FDKaacEncDistributeBits(
				&state,
				&element,
				psy[:],
				&peData,
				tt.nChannels,
				0,
				tt.grantedDyn,
				tt.bitresBits,
				tt.maxBitresBits,
				MaxValDBL,
				tt.mode,
			)

			got := [...]int{
				grantedPE,
				grantedPECorr,
				element.PeMin,
				element.PeMax,
				element.PeLast,
				element.DynBitsLast,
				int(element.PeCorrectionFactorM),
				element.PeCorrectionFactorE,
			}
			if got != tt.want {
				t.Fatalf("distribution state = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFDKaacEncPECorrectionHighBitresVector(t *testing.T) {
	gotM, gotE := fdkaacEncCalcPECorrection(
		peCorrection115Over2,
		650,
		700,
		650,
		defaultBitresBits2PEFactorM,
		defaultBitresBits2PEFactorE,
	)
	got := [...]int{int(gotM), gotE}
	want := [...]int{1186484695, 1}
	if got != want {
		t.Fatalf("high bitres PE correction = %v, want %v", got, want)
	}

	resetM, resetE := fdkaacEncCalcPECorrection(
		peCorrection115Over2,
		650,
		700,
		0,
		defaultBitresBits2PEFactorM,
		defaultBitresBits2PEFactorE,
	)
	if resetM != peCorrectionHalf || resetE != 1 {
		t.Fatalf("high bitres PE correction reset = (%d,%d), want (%d,1)", resetM, resetE, peCorrectionHalf)
	}
}

func TestFDKaacEncThresholdExpVector(t *testing.T) {
	_, psy, qc, _, _ := buildAdjThrLongPatchCase()
	var thrExp [2][maxGroupedSFB]FixpDBL
	thrExp[0][8] = 12345

	FDKaacEncCalcThresholdExp(&thrExp, qc[:], psy[:], 1)

	want := [...]FixpDBL{146436310, 162369982, 154197474, 139065786, 162369982, 158230974, 170975635, 132066240}
	assertFixpDBLSlice(t, "threshold exponent", thrExp[0][:8], want[:], 0xe9a52d022f9509a2)
	if thrExp[0][8] != 12345 {
		t.Fatalf("threshold exponent touched inactive band = %d", thrExp[0][8])
	}
}

func TestFDKaacEncAdaptMinSnrVector(t *testing.T) {
	var param MinSNRAdaptParam
	FDKaacEncInitMinSnrAdaptParam(&param)
	gotParam := [...]FixpDBL{param.MaxRed, param.StartRatio, param.RedRatioFac, param.RedOffs}
	wantParam := [...]FixpDBL{0x00800000, 0x06a4d3c0, -0x30000000, 0x02c00000}
	if gotParam != wantParam {
		t.Fatalf("min-SNR defaults = %v, want %v", gotParam, wantParam)
	}

	psyStorage, qcStorage := buildMinSnrAdaptCase()
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	FDKaacEncAdaptMinSnr(qc[:], psy[:], &param, 1)

	wantMinSnr := [...]FixpDBL{-90000000, -109772544, -120282752, -80000000, -70000000, -110239808, -95471104, -100000000}
	assertFixpDBLSlice(t, "adapted min-SNR", qc[0].SfbMinSnrLdData[:8], wantMinSnr[:], 0xe5ce61fd11345447)

	zeroPsy, zeroQC := buildMinSnrAdaptCase()
	for i := 0; i < 8; i++ {
		zeroPsy.SfbEnergy[i] = 0
		zeroQC.SfbEnergyLdData[i] = MinValDBL
	}
	before := zeroQC.SfbMinSnrLdData
	zeroPsyPtrs := [1]*PsyOutChannel{&zeroPsy}
	zeroQCPtrs := [1]*QCOutChannel{&zeroQC}
	FDKaacEncAdaptMinSnr(zeroQCPtrs[:], zeroPsyPtrs[:], &param, 1)
	if zeroQC.SfbMinSnrLdData != before {
		t.Fatalf("zero-energy min-SNR changed = %v, want %v", zeroQC.SfbMinSnrLdData[:8], before[:8])
	}
}

func TestFDKaacEncInitAvoidHoleFlagLongVector(t *testing.T) {
	if got, want := [...]uint8{AvoidHoleNone, AvoidHoleInactive, AvoidHoleActive}, [...]uint8{0, 1, 2}; got != want {
		t.Fatalf("avoid-hole constants = %v, want %v", got, want)
	}

	psyStorage, qcStorage := buildMinSnrAdaptCase()
	copy(qcStorage.SfbEnergy[:], psyStorage.SfbEnergy[:])
	spreadEnergy := [...]FixpDBL{100000000, 2000000, 10000000, 200000000, 50000000, 3000000, 2000000, 40000000}
	copy(qcStorage.SfbSpreadEnergy[:], spreadEnergy[:])
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	var tools ToolsInfo
	var ahFlag [2][maxGroupedSFB]uint8
	ahParam := AHParam{ModifyMinSnr: 1}

	FDKaacEncInitAvoidHoleFlag(qc[:], psy[:], &ahFlag, &tools, 1, &ahParam)

	wantSpread := [...]FixpDBL{50000000, 1000000, 5000000, 100000000, 25000000, 1500000, 1000000, 20000000}
	wantMinSnr := [...]FixpDBL{-90000000, -64302174, -94302174, -80000000, -70000000, -104302174, -54302174, -100000000}
	wantFlags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleNone, AvoidHoleInactive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive}
	assertFixpDBLSlice(t, "long avoid-hole spread", qc[0].SfbSpreadEnergy[:8], wantSpread[:], 0x753349bb575847e5)
	assertFixpDBLSlice(t, "long avoid-hole min-SNR", qc[0].SfbMinSnrLdData[:8], wantMinSnr[:], 0xd66b4eb205740560)
	assertUint8Slice(t, "long avoid-hole flags", ahFlag[0][:8], wantFlags[:])
}

func TestFDKaacEncInitAvoidHoleFlagShortVector(t *testing.T) {
	psyStorage, qcStorage := buildMinSnrAdaptCase()
	psyStorage.LastWindowSequence = ShortWindow
	psyStorage.MaxSfbPerGroup = 3
	copy(qcStorage.SfbEnergy[:], psyStorage.SfbEnergy[:])
	spreadEnergy := [...]FixpDBL{400000000, 3000000, 2000000, 123456789, 40000000, 1000000, 200000000, 987654321}
	copy(qcStorage.SfbSpreadEnergy[:], spreadEnergy[:])
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	var tools ToolsInfo
	var ahFlag [2][maxGroupedSFB]uint8
	ahFlag[0][3] = AvoidHoleActive
	ahFlag[0][7] = AvoidHoleActive
	ahParam := AHParam{ModifyMinSnr: 0}

	FDKaacEncInitAvoidHoleFlag(qc[:], psy[:], &ahFlag, &tools, 1, &ahParam)

	wantSpread := [...]FixpDBL{251999998, 1889999, 1259999, 123456789, 25199999, 629999, 125999999, 987654321}
	wantMinSnr := [...]FixpDBL{-90000000, -120000000, -150000000, -80000000, -70000000, -160000000, -110000000, -100000000}
	wantFlags := [...]uint8{AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleActive}
	assertFixpDBLSlice(t, "short avoid-hole spread", qc[0].SfbSpreadEnergy[:8], wantSpread[:], 0xfdf217b5498431a2)
	assertFixpDBLSlice(t, "short avoid-hole min-SNR", qc[0].SfbMinSnrLdData[:8], wantMinSnr[:], 0xe3f3bae079921845)
	assertUint8Slice(t, "short avoid-hole flags", ahFlag[0][:8], wantFlags[:])
}

func TestFDKaacEncInitAvoidHoleFlagStereoMSVector(t *testing.T) {
	leftPsy, leftQC := buildMinSnrAdaptCase()
	rightPsy := leftPsy
	var rightQC QCOutChannel

	rightEnergy := [...]FixpDBL{100000000, 8000000, 6000000, 20000000, 160000000, 500000, 6000000, 10000000}
	rightEnergyLd := [...]FixpDBL{-148464108, -270731634, -284657976, -226374761, -125711465, -404949362, -284657976, -259929193}
	leftSpread := [...]FixpDBL{20000000, 12000000, 100000000, 10000000, 400000000, 2000000, 1000000, 300000000}
	rightSpread := [...]FixpDBL{50000000, 1000000, 2000000, 80000000, 60000000, 1000000, 10000000, 20000000}
	rightMinSnr := [...]FixpDBL{-95000000, -100000000, -130000000, -60000000, -85000000, -140000000, -90000000, -75000000}

	copy(leftQC.SfbEnergy[:], leftPsy.SfbEnergy[:])
	copy(leftQC.SfbSpreadEnergy[:], leftSpread[:])
	copy(rightPsy.SfbEnergy[:], rightEnergy[:])
	copy(rightQC.SfbEnergy[:], rightEnergy[:])
	copy(rightQC.SfbEnergyLdData[:], rightEnergyLd[:])
	copy(rightQC.SfbMinSnrLdData[:], rightMinSnr[:])
	copy(rightQC.SfbSpreadEnergy[:], rightSpread[:])

	psy := [2]*PsyOutChannel{&leftPsy, &rightPsy}
	qc := [2]*QCOutChannel{&leftQC, &rightQC}
	tools := ToolsInfo{MsMask: [maxGroupedSFB]int{1, 0, 1, 0, 1, 1, 0, 1}}
	var ahFlag [2][maxGroupedSFB]uint8
	ahParam := AHParam{ModifyMinSnr: 0}

	FDKaacEncInitAvoidHoleFlag(qc[:], psy[:], &ahFlag, &tools, 2, &ahParam)

	wantLeftSpread := [...]FixpDBL{179999995, 6000000, 1799999, 5000000, 161999995, 1000000, 500000, 150000000}
	wantRightSpread := [...]FixpDBL{89999997, 500000, 1000000, 40000000, 30000000, 500000, 5000000, 10000000}
	wantLeftMinSnr := [...]FixpDBL{-90000000, -120000000, -150000000, -80000000, -70000000, -160000000, -110000000, -100000000}
	wantRightMinSnr := [...]FixpDBL{-95000000, -100000000, -130000000, -60000000, -85000000, -140000000, -90000000, -60744127}
	wantLeftFlags := [...]uint8{AvoidHoleInactive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone}
	wantRightFlags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleInactive}
	assertFixpDBLSlice(t, "stereo avoid-hole left spread", qc[0].SfbSpreadEnergy[:8], wantLeftSpread[:], 0x4182af6c7e1a8aa3)
	assertFixpDBLSlice(t, "stereo avoid-hole right spread", qc[1].SfbSpreadEnergy[:8], wantRightSpread[:], 0x8f718c737e7686a3)
	assertFixpDBLSlice(t, "stereo avoid-hole left min-SNR", qc[0].SfbMinSnrLdData[:8], wantLeftMinSnr[:], 0xe3f3bae079921845)
	assertFixpDBLSlice(t, "stereo avoid-hole right min-SNR", qc[1].SfbMinSnrLdData[:8], wantRightMinSnr[:], 0xc4ec53be2857d015)
	assertUint8Slice(t, "stereo avoid-hole left flags", ahFlag[0][:8], wantLeftFlags[:])
	assertUint8Slice(t, "stereo avoid-hole right flags", ahFlag[1][:8], wantRightFlags[:])
}

func TestFDKaacEncCalcPENoAHVector(t *testing.T) {
	peData, psyStorage, ahFlag := buildPENoAHCase()
	psy := [1]*PsyOutChannel{&psyStorage}

	pe, constPart, nActiveLines := FDKaacEncCalcPENoAH(&peData, &ahFlag, psy[:], 1)
	got := [...]int{pe, constPart, nActiveLines}
	want := [...]int{179, 1790, 179}
	if got != want {
		t.Fatalf("no-AH PE totals = %v, want %v", got, want)
	}
}

func TestFDKaacEncCalcRedValPowerVectors(t *testing.T) {
	input := [...]struct {
		num   FixpDBL
		denum FixpDBL
	}{
		{0, 40},
		{25, 100},
		{-25, 100},
		{90, 45},
		{-75, 320},
		{-290, 104},
		{-80, 104},
		{200, 160},
	}
	want := [...]FixpDBL{
		1073741824, 1,
		1276901414, 1,
		902905648, 1,
		1073741824, 3,
		912737645, 1,
		1243422908, -2,
		1259997673, 0,
		1276901414, 2,
	}

	var got [len(input) * 2]FixpDBL
	for i, tt := range input {
		m, e := FDKaacEncCalcRedValPower(tt.num, tt.denum)
		got[2*i] = m
		got[2*i+1] = FixpDBL(e)
	}
	assertFixpDBLSlice(t, "red-value power mantissa/exponent", got[:], want[:], 0xf05ed6c375a81fcb)
}

func TestFDKaacEncReductionValueVectors(t *testing.T) {
	initialInput := [...]struct {
		constPart int
		noRedPe   int
		desiredPe int
		active    int
	}{
		{-125, 205, 120, 26},
		{1790, 179, 90, 179},
		{350, 260, 120, 40},
		{0, 400, 320, 64},
	}
	secondInput := [...]struct {
		redM      FixpDBL
		redE      int
		constPart int
		redPe     int
		desiredPe int
		active    int
	}{
		{181429942, 0, 100, 160, 80, 20},
		{181429942, 0, 1790, 179, 90, 179},
		{0, 1, 350, 260, 120, 40},
	}
	want := [...]FixpDBL{
		181429942, 0,
		57431353, 4,
		330627911, 3,
		87922620, 1,
		182291419, 3,
		34385362, 5,
		165313956, 4,
	}

	var got [len(initialInput)*2 + len(secondInput)*2]FixpDBL
	out := 0
	for _, tt := range initialInput {
		m, e := fdkaacEncInitialReductionValue(tt.constPart, tt.noRedPe, tt.desiredPe, tt.active)
		got[out] = m
		got[out+1] = FixpDBL(e)
		out += 2
	}
	for _, tt := range secondInput {
		m, e := fdkaacEncSecondGuessReductionValue(tt.redM, tt.redE, tt.constPart, tt.redPe, tt.desiredPe, tt.active)
		got[out] = m
		got[out+1] = FixpDBL(e)
		out += 2
	}
	assertFixpDBLSlice(t, "CBR reduction values", got[:], want[:], 0x179c5b2fd6463248)
}

func TestFDKaacEncReduceThresholdsCBRVector(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	fillThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, &thrExp, 1, 0x20000000, 0)

	wantThreshold := [...]FixpDBL{-221732784, -500000000, -219545908, -253192572, -217269500, -505000000, -214901004, -253564456}
	wantFlags := [...]uint8{AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleActive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleActive}
	assertFixpDBLSlice(t, "CBR reduced thresholds", qc[0].SfbThresholdLdData[:8], wantThreshold[:], 0xfb17d2f72a7cf5df)
	assertUint8Slice(t, "CBR avoid-hole flags", ahFlag[0][:8], wantFlags[:])
}

func TestFDKaacEncReduceThresholdsCBRMinimumRatioVector(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	fillThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, &thrExp, 1, 0x02000000, 0)

	wantThreshold := [...]FixpDBL{-438160343, -500000000, -471875624, -488146048, -443260691, -505000000, -455301736, -476815123}
	assertFixpDBLSlice(t, "CBR minimum-ratio thresholds", qc[0].SfbThresholdLdData[:8], wantThreshold[:], 0xef89e842df8e6d3a)
}

func TestFDKaacEncReduceThresholdsCBRZeroReduction(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	fillThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
	beforeThreshold := qcStorage.SfbThresholdLdData
	beforeFlags := ahFlag
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, &thrExp, 1, 0, 0)

	if qcStorage.SfbThresholdLdData != beforeThreshold {
		t.Fatalf("zero-reduction thresholds changed = %v, want %v", qcStorage.SfbThresholdLdData[:8], beforeThreshold[:8])
	}
	if ahFlag != beforeFlags {
		t.Fatalf("zero-reduction flags changed = %v, want %v", ahFlag[0][:8], beforeFlags[0][:8])
	}
}

func TestFDKaacEncVBRReductionTables(t *testing.T) {
	wantInvInt := [...]FixpDBL{MaxValDBL, MaxValDBL, 0x40000000, 0x2aaaaaaa, 0x20000000, 0x19999999, 0x15555555, 0x12492492}
	wantInvSqrt4 := [...]FixpDBL{MaxValDBL, MaxValDBL, 0x6ba27e65, 0x61424bb5, 0x5a827999, 0x55994845, 0x51c8e33c, 0x4eb160d1}
	assertFixpDBLSlice(t, "VBR inverse-int table", adjThrInvInt[:], wantInvInt[:], 0xd99e30ed42afba6c)
	assertFixpDBLSlice(t, "VBR inverse-sqrt4 table", adjThrInvSqrt4[:], wantInvSqrt4[:], 0x1deaf605797b3cf9)
}

func TestFDKaacEncCalcChaosMeasureVector(t *testing.T) {
	_, psy, qc, _, _ := buildAdjThrLongPatchCase()
	got := FDKaacEncCalcChaosMeasure(psy[0], qc[0].SfbFormFactorLdData[:])
	if got != 8 {
		t.Fatalf("chaos measure = %d, want 8", got)
	}

	noActive := *psy[0]
	for i := 0; i < 8; i++ {
		noActive.SfbEnergyLdData[i] = -600000000
		noActive.SfbThresholdLdData[i] = -100000000
	}
	if got := FDKaacEncCalcChaosMeasure(&noActive, qc[0].SfbFormFactorLdData[:]); got != MaxValDBL {
		t.Fatalf("inactive chaos measure = %d, want %d", got, MaxValDBL)
	}
}

func TestFDKaacEncReduceThresholdsVBRLongVector(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	fillVBRLongThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	chaosOld := vbrChaosHalf

	FDKaacEncReduceThresholdsVBR(qc[:], psy[:], &ahFlag, &thrExp, 1, peCorrectionHalf, &chaosOld)

	wantThreshold := [...]FixpDBL{-464694604, -500000000, -457131312, -472161452, -449474100, -505000000, -441725520, -479529344}
	wantFlags := [...]uint8{AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive}
	assertFixpDBLSlice(t, "VBR long thresholds", qc[0].SfbThresholdLdData[:8], wantThreshold[:], 0x09e6d3e3b556f9bd)
	assertUint8Slice(t, "VBR long avoid-hole flags", ahFlag[0][:8], wantFlags[:])
	if chaosOld != 0 {
		t.Fatalf("VBR long chaos history = %d, want 0", chaosOld)
	}
}

func TestFDKaacEncReduceThresholdsVBRShortVector(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	fillVBRShortThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	chaosOld := FixpDBL(0x30000000)

	FDKaacEncReduceThresholdsVBR(qc[:], psy[:], &ahFlag, &thrExp, 1, 0x50000000, &chaosOld)

	wantThreshold := [...]FixpDBL{-242726640, -500000000, -240298184, -530000000, -237773484, -505000000, -235150056, -540000000}
	wantFlags := [...]uint8{AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, 0, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, 0}
	assertFixpDBLSlice(t, "VBR short thresholds", qc[0].SfbThresholdLdData[:8], wantThreshold[:], 0x38db12f3606c9c20)
	assertUint8Slice(t, "VBR short avoid-hole flags", ahFlag[0][:8], wantFlags[:])
	if chaosOld != 0x34000000 {
		t.Fatalf("VBR short chaos history = %#x, want 0x34000000", chaosOld)
	}
}

func TestFDKaacEncCorrectThresholdsVectors(t *testing.T) {
	for _, tt := range []struct {
		name          string
		deltaPe       int
		wantThreshold [8]FixpDBL
		wantPEFactors [8]FixpDBL
		thresholdHash uint64
		factorHash    uint64
	}{
		{
			name:          "negative delta",
			deltaPe:       -34,
			wantThreshold: [8]FixpDBL{-462798369, -444410394, -510000000, -530000000, -444410394, -505000000, -435243923, -481257449},
			wantPEFactors: [8]FixpDBL{66016450, 117815144, MinValDBL, MinValDBL, 158831828, MinValDBL, 131010137, 154040199},
			thresholdHash: 0x5b793ef08820a954,
			factorHash:    0x0e7edda028d26d3e,
		},
		{
			name:          "positive delta",
			deltaPe:       29,
			wantThreshold: [8]FixpDBL{-556590981, -535559795, -510000000, -567088414, -535559795, -540822049, -525026599, -577576684},
			wantPEFactors: [8]FixpDBL{66016450, 117815144, MinValDBL, 144581084, 158831828, 98542766, 131010137, 154040199},
			thresholdHash: 0x18c733ed143dd661,
			factorHash:    0x21e09ec057373cb5,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var psyStorage PsyOutChannel
			var qcStorage QCOutChannel
			var peData PEData
			var thrExp [2][maxGroupedSFB]FixpDBL
			var ahFlag [2][maxGroupedSFB]uint8
			var scratch CorrectThresholdScratch
			fillCorrectThresholdCase(&psyStorage, &qcStorage, &peData, &thrExp, &ahFlag)
			psy := [1]*PsyOutChannel{&psyStorage}
			qc := [1]*QCOutChannel{&qcStorage}

			FDKaacEncCorrectThresholds(qc[:], psy[:], &peData, &ahFlag, &thrExp, &scratch, 1, 0x18000000, 0, tt.deltaPe)

			wantFlags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleActive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive}
			wantActiveLd := [...]FixpDBL{0, 53182516, MinValDBL, 77910978, 94199200, 33554432, 67108864, 86736948}
			assertFixpDBLSlice(t, tt.name+" thresholds", qc[0].SfbThresholdLdData[:8], tt.wantThreshold[:], tt.thresholdHash)
			assertFixpDBLSlice(t, tt.name+" PE factors", scratch.SfbPEFactorsLdData[0][:8], tt.wantPEFactors[:], tt.factorHash)
			assertFixpDBLSlice(t, tt.name+" active-line LD", scratch.SfbNActiveLinesLdData[0][:8], wantActiveLd[:], 0xfa1e0c5c4038d0c3)
			assertUint8Slice(t, tt.name+" avoid-hole flags", ahFlag[0][:8], wantFlags[:])
		})
	}
}

func TestFDKaacEncReduceMinSnrVector(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var peData PEData
	var ahFlag [2][maxGroupedSFB]uint8
	fillReduceMinSnrCase(&psyStorage, &qcStorage, &peData, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	redPeGlobal := 160

	FDKaacEncReduceMinSnr(qc[:], psy[:], &peData, &ahFlag, 1, 90, &redPeGlobal)

	gotTotals := [...]FixpDBL{peData.Pe, peData.PEChannelData[0].Pe, FixpDBL(redPeGlobal)}
	wantTotals := [...]FixpDBL{76, 76, 76}
	wantThreshold := [...]FixpDBL{-520000000, -500000000, -510000000, -320802114, -500000000, -505000000, -210802114, -340802114}
	wantMinSnr := [...]FixpDBL{-90000000, -120000000, -150000000, snrLdFac, -70000000, -160000000, snrLdFac, snrLdFac}
	wantSfbPe := [...]FixpDBL{720896, 1441792, 5242880, 1966080, 983040, 1638400, 983040, 2359296}
	assertFixpDBLSlice(t, "reduce-min-SNR totals", gotTotals[:], wantTotals[:], 0xbe4f7b423a8dd959)
	assertFixpDBLSlice(t, "reduce-min-SNR thresholds", qc[0].SfbThresholdLdData[:8], wantThreshold[:], 0xe05c86648ca3f4a1)
	assertFixpDBLSlice(t, "reduce-min-SNR min-SNR", qc[0].SfbMinSnrLdData[:8], wantMinSnr[:], 0x05d1661497c1a708)
	assertFixpDBLSlice(t, "reduce-min-SNR sfb PE", peData.PEChannelData[0].SfbPe[:8], wantSfbPe[:], 0x3391ec27ed897f9b)
}

func TestFDKaacEncReduceMinSnrNoOp(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var peData PEData
	var ahFlag [2][maxGroupedSFB]uint8
	fillReduceMinSnrCase(&psyStorage, &qcStorage, &peData, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	redPeGlobal := 80
	beforeThreshold := qcStorage.SfbThresholdLdData
	beforePe := peData.PEChannelData[0].SfbPe

	FDKaacEncReduceMinSnr(qc[:], psy[:], &peData, &ahFlag, 1, 90, &redPeGlobal)

	if redPeGlobal != 80 {
		t.Fatalf("no-op reduce-min-SNR PE = %d, want 80", redPeGlobal)
	}
	if qcStorage.SfbThresholdLdData != beforeThreshold {
		t.Fatalf("no-op reduce-min-SNR threshold changed = %v, want %v", qcStorage.SfbThresholdLdData[:8], beforeThreshold[:8])
	}
	if peData.PEChannelData[0].SfbPe != beforePe {
		t.Fatalf("no-op reduce-min-SNR PE bands changed = %v, want %v", peData.PEChannelData[0].SfbPe[:8], beforePe[:8])
	}
}

func TestFDKaacEncAllowMoreHolesMSVector(t *testing.T) {
	var leftPsy PsyOutChannel
	var rightPsy PsyOutChannel
	var leftQC QCOutChannel
	var rightQC QCOutChannel
	var peData PEData
	var tools ToolsInfo
	var state ATSElement
	var ahFlag [2][maxGroupedSFB]uint8
	fillAllowMoreHolesMSCase(&leftPsy, &rightPsy, &leftQC, &rightQC, &peData, &tools, &state, &ahFlag)
	psy := [2]*PsyOutChannel{&leftPsy, &rightPsy}
	qc := [2]*QCOutChannel{&leftQC, &rightQC}

	FDKaacEncAllowMoreHoles(qc[:], psy[:], &peData, &tools, &state, &ahFlag, 2, 70, 120)

	wantLeftThreshold := [...]FixpDBL{-520000000, -510000000, -500000000, -490000000}
	wantRightThreshold := [...]FixpDBL{-530000000, -520000000, -366445568, -500000000}
	wantLeftFlags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive}
	wantRightFlags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleActive}
	wantLeftPE := [...]FixpDBL{655360, 1310720, 1966080, 0}
	wantRightPE := [...]FixpDBL{983040, 1638400, 3932160, 0}
	assertFixpDBLSlice(t, "allow-more-holes MS left threshold", qc[0].SfbThresholdLdData[:4], wantLeftThreshold[:], 0xe569c44c5bbc4aff)
	assertFixpDBLSlice(t, "allow-more-holes MS right threshold", qc[1].SfbThresholdLdData[:4], wantRightThreshold[:], 0x17b3f12a8fcd925f)
	assertUint8Slice(t, "allow-more-holes MS left flags", ahFlag[0][:4], wantLeftFlags[:])
	assertUint8Slice(t, "allow-more-holes MS right flags", ahFlag[1][:4], wantRightFlags[:])
	assertFixpDBLSlice(t, "allow-more-holes MS left PE", peData.PEChannelData[0].SfbPe[:4], wantLeftPE[:], 0x670a2d6276fde545)
	assertFixpDBLSlice(t, "allow-more-holes MS right PE", peData.PEChannelData[1].SfbPe[:4], wantRightPE[:], 0xca3908e3c9e67e7f)
}

func TestFDKaacEncAllowMoreHolesEnergyVector(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var peData PEData
	var tools ToolsInfo
	var state ATSElement
	var ahFlag [2][maxGroupedSFB]uint8
	fillAllowMoreHolesEnergyCase(&psyStorage, &qcStorage, &peData, &tools, &state, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	FDKaacEncAllowMoreHoles(qc[:], psy[:], &peData, &tools, &state, &ahFlag, 1, 70, 100)

	wantThreshold := [...]FixpDBL{-520000000, -500000000, -510000000, -296445568, -500000000, -505000000, -490000000, -540000000}
	wantFlags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleNone, AvoidHoleInactive, AvoidHoleActive, AvoidHoleInactive, AvoidHoleInactive}
	wantPE := [...]FixpDBL{720896, 1441792, 5242880, 2621440, 983040, 1638400, 2293760, 4587520}
	assertFixpDBLSlice(t, "allow-more-holes energy threshold", qc[0].SfbThresholdLdData[:8], wantThreshold[:], 0xd5c50105c1114255)
	assertUint8Slice(t, "allow-more-holes energy flags", ahFlag[0][:8], wantFlags[:])
	assertFixpDBLSlice(t, "allow-more-holes energy PE", peData.PEChannelData[0].SfbPe[:8], wantPE[:], 0xbddc50616fc42913)
}

func TestFDKaacEncAllowMoreHolesNoOp(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var peData PEData
	var tools ToolsInfo
	var state ATSElement
	var ahFlag [2][maxGroupedSFB]uint8
	fillAllowMoreHolesEnergyCase(&psyStorage, &qcStorage, &peData, &tools, &state, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	beforeThreshold := qcStorage.SfbThresholdLdData
	beforeFlags := ahFlag

	FDKaacEncAllowMoreHoles(qc[:], psy[:], &peData, &tools, &state, &ahFlag, 1, 100, 90)

	if qcStorage.SfbThresholdLdData != beforeThreshold {
		t.Fatalf("no-op allow-more-holes threshold changed = %v, want %v", qcStorage.SfbThresholdLdData[:8], beforeThreshold[:8])
	}
	if ahFlag != beforeFlags {
		t.Fatalf("no-op allow-more-holes flags changed = %v, want %v", ahFlag[0][:8], beforeFlags[0][:8])
	}
}

func TestFDKaacEncResetAHFlagsVector(t *testing.T) {
	psy0 := PsyOutChannel{
		SfbCnt:             8,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     3,
		LastWindowSequence: LongWindow,
	}
	psy1 := PsyOutChannel{
		SfbCnt:             8,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     2,
		LastWindowSequence: LongWindow,
	}
	for i := 0; i <= 8; i++ {
		psy0.SfbOffsets[i] = i
		psy1.SfbOffsets[i] = i
	}
	psy := [2]*PsyOutChannel{&psy0, &psy1}
	ahFlag := [2][maxGroupedSFB]uint8{
		{AvoidHoleActive, AvoidHoleNone, AvoidHoleActive, AvoidHoleActive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleActive},
		{AvoidHoleActive, AvoidHoleActive, AvoidHoleActive, AvoidHoleNone, AvoidHoleActive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleActive},
	}

	FDKaacEncResetAHFlags(&ahFlag, psy[:], 2)

	want0 := [...]uint8{AvoidHoleInactive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleActive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleActive}
	want1 := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleActive}
	assertUint8Slice(t, "reset AH flags ch0", ahFlag[0][:8], want0[:])
	assertUint8Slice(t, "reset AH flags ch1", ahFlag[1][:8], want1[:])
}

func TestFDKaacEncResetAHFlagsRejectsInvalid(t *testing.T) {
	psy0 := PsyOutChannel{
		SfbCnt:             8,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     3,
		LastWindowSequence: LongWindow,
	}
	for i := 0; i <= 8; i++ {
		psy0.SfbOffsets[i] = i
	}
	psy := [1]*PsyOutChannel{&psy0}
	var ahFlag [2][maxGroupedSFB]uint8

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil flags", func() { FDKaacEncResetAHFlags(nil, psy[:], 1) }},
		{"bad channel count", func() { FDKaacEncResetAHFlags(&ahFlag, psy[:], 0) }},
		{"short psy", func() { FDKaacEncResetAHFlags(&ahFlag, psy[:0], 1) }},
		{"nil psy", func() {
			bad := [1]*PsyOutChannel{nil}
			FDKaacEncResetAHFlags(&ahFlag, bad[:], 1)
		}},
		{"bad band shape", func() {
			bad := psy0
			bad.SfbPerGroup = 3
			FDKaacEncResetAHFlags(&ahFlag, []*PsyOutChannel{&bad}, 1)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncPECalculationRejectsInvalid(t *testing.T) {
	peData, psy, qc, tools, state := buildAdjThrLongPatchCase()
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil PE data", func() { FDKaacEncPECalculation(nil, psy[:], qc[:], &tools, &state, 1) }},
		{"nil tools", func() { FDKaacEncPECalculation(&peData, psy[:], qc[:], nil, &state, 1) }},
		{"nil state", func() { FDKaacEncPECalculation(&peData, psy[:], qc[:], &tools, nil, 1) }},
		{"bad channel count", func() { FDKaacEncPECalculation(&peData, psy[:], qc[:], &tools, &state, 0) }},
		{"short channels", func() { FDKaacEncPECalculation(&peData, psy[:0], qc[:], &tools, &state, 1) }},
		{"nil psy channel", func() {
			bad := psy
			bad[0] = nil
			FDKaacEncPECalculation(&peData, bad[:], qc[:], &tools, &state, 1)
		}},
		{"nil qc channel", func() {
			bad := qc
			bad[0] = nil
			FDKaacEncPECalculation(&peData, psy[:], bad[:], &tools, &state, 1)
		}},
		{"bad group multiple", func() {
			bad := *psy[0]
			bad.SfbPerGroup = 3
			FDKaacEncPECalculation(&peData, []*PsyOutChannel{&bad}, qc[:], &tools, &state, 1)
		}},
		{"empty spectrum", func() {
			bad := *psy[0]
			bad.SfbOffsets[bad.SfbCnt] = 0
			FDKaacEncPECalculation(&peData, []*PsyOutChannel{&bad}, qc[:], &tools, &state, 1)
		}},
		{"bad block type", func() {
			bad := *psy[0]
			bad.LastWindowSequence = -1
			FDKaacEncPECalculation(&peData, []*PsyOutChannel{&bad}, qc[:], &tools, &state, 1)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncMinSnrAdjustmentRejectsInvalid(t *testing.T) {
	var param MinSNRAdaptParam
	FDKaacEncInitMinSnrAdaptParam(&param)
	var thrExp [2][maxGroupedSFB]FixpDBL
	psyStorage, qcStorage := buildMinSnrAdaptCase()
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil threshold scratch", func() { FDKaacEncCalcThresholdExp(nil, qc[:], psy[:], 1) }},
		{"nil min snr param", func() { FDKaacEncAdaptMinSnr(qc[:], psy[:], nil, 1) }},
		{"nil init param", func() { FDKaacEncInitMinSnrAdaptParam(nil) }},
		{"bad channel count", func() { FDKaacEncCalcThresholdExp(&thrExp, qc[:], psy[:], 0) }},
		{"short inputs", func() { FDKaacEncAdaptMinSnr(qc[:0], psy[:], &param, 1) }},
		{"nil qc", func() {
			bad := [1]*QCOutChannel{nil}
			FDKaacEncAdaptMinSnr(bad[:], psy[:], &param, 1)
		}},
		{"nil psy", func() {
			bad := [1]*PsyOutChannel{nil}
			FDKaacEncCalcThresholdExp(&thrExp, qc[:], bad[:], 1)
		}},
		{"bad band shape", func() {
			badPsy := psyStorage
			badPsy.SfbPerGroup = 3
			bad := [1]*PsyOutChannel{&badPsy}
			FDKaacEncAdaptMinSnr(qc[:], bad[:], &param, 1)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncInitAvoidHoleFlagRejectsInvalid(t *testing.T) {
	psyStorage, qcStorage := buildMinSnrAdaptCase()
	copy(qcStorage.SfbEnergy[:], psyStorage.SfbEnergy[:])
	copy(qcStorage.SfbSpreadEnergy[:], psyStorage.SfbEnergy[:])
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	var ahFlag [2][maxGroupedSFB]uint8
	var tools ToolsInfo
	ahParam := AHParam{ModifyMinSnr: 1}

	rightPsy := psyStorage
	rightQC := qcStorage
	stereoPsy := [2]*PsyOutChannel{&psyStorage, &rightPsy}
	stereoQC := [2]*QCOutChannel{&qcStorage, &rightQC}

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil flag scratch", func() { FDKaacEncInitAvoidHoleFlag(qc[:], psy[:], nil, &tools, 1, &ahParam) }},
		{"nil tools", func() { FDKaacEncInitAvoidHoleFlag(qc[:], psy[:], &ahFlag, nil, 1, &ahParam) }},
		{"nil parameter", func() { FDKaacEncInitAvoidHoleFlag(qc[:], psy[:], &ahFlag, &tools, 1, nil) }},
		{"bad channel count", func() { FDKaacEncInitAvoidHoleFlag(qc[:], psy[:], &ahFlag, &tools, 0, &ahParam) }},
		{"short inputs", func() { FDKaacEncInitAvoidHoleFlag(qc[:0], psy[:], &ahFlag, &tools, 1, &ahParam) }},
		{"nil qc", func() {
			bad := [1]*QCOutChannel{nil}
			FDKaacEncInitAvoidHoleFlag(bad[:], psy[:], &ahFlag, &tools, 1, &ahParam)
		}},
		{"nil psy", func() {
			bad := [1]*PsyOutChannel{nil}
			FDKaacEncInitAvoidHoleFlag(qc[:], bad[:], &ahFlag, &tools, 1, &ahParam)
		}},
		{"bad band shape", func() {
			badPsy := psyStorage
			badPsy.SfbPerGroup = 3
			bad := [1]*PsyOutChannel{&badPsy}
			FDKaacEncInitAvoidHoleFlag(qc[:], bad[:], &ahFlag, &tools, 1, &ahParam)
		}},
		{"mismatched stereo bands", func() {
			rightPsy.MaxSfbPerGroup = 3
			FDKaacEncInitAvoidHoleFlag(stereoQC[:], stereoPsy[:], &ahFlag, &tools, 2, &ahParam)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncThresholdReductionRejectsInvalid(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	fillThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	peData, pePsy, peFlags := buildPENoAHCase()
	chaosOld := vbrChaosHalf
	var correctScratch CorrectThresholdScratch
	adaptPE := PEData{Pe: 205, ConstPart: -125, NActiveLines: 26}
	adaptState := ATSElement{AHParam: AHParam{ModifyMinSnr: 1, StartSfbL: 4, StartSfbS: 3}}
	FDKaacEncInitMinSnrAdaptParam(&adaptState.MinSNRAdaptParam)
	var adaptTools ToolsInfo
	var adaptScratch AdaptThresholdsToPeScratch
	adaptVBRState := ATSElement{AHParam: AHParam{ModifyMinSnr: 1}, VBRQualFactor: peCorrectionHalf, ChaosMeasureOld: vbrChaosHalf}
	FDKaacEncInitMinSnrAdaptParam(&adaptVBRState.MinSNRAdaptParam)
	var adaptVBRScratch AdaptThresholdsVBRScratch

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil no-AH PE data", func() { FDKaacEncCalcPENoAH(nil, &peFlags, []*PsyOutChannel{&pePsy}, 1) }},
		{"nil no-AH flags", func() { FDKaacEncCalcPENoAH(&peData, nil, []*PsyOutChannel{&pePsy}, 1) }},
		{"bad no-AH channel count", func() { FDKaacEncCalcPENoAH(&peData, &peFlags, []*PsyOutChannel{&pePsy}, 0) }},
		{"nil no-AH psy", func() { FDKaacEncCalcPENoAH(&peData, &peFlags, []*PsyOutChannel{nil}, 1) }},
		{"zero red-power denominator", func() { FDKaacEncCalcRedValPower(1, 0) }},
		{"min red-power numerator", func() { FDKaacEncCalcRedValPower(MinValDBL, 1) }},
		{"nil adapt PE", func() {
			FDKaacEncAdaptThresholdsToPeCBR(nil, qc[:], psy[:], &adaptTools, &adaptState, &adaptScratch, 1, 120, 1)
		}},
		{"nil adapt tools", func() {
			FDKaacEncAdaptThresholdsToPeCBR(&adaptPE, qc[:], psy[:], nil, &adaptState, &adaptScratch, 1, 120, 1)
		}},
		{"nil adapt state", func() {
			FDKaacEncAdaptThresholdsToPeCBR(&adaptPE, qc[:], psy[:], &adaptTools, nil, &adaptScratch, 1, 120, 1)
		}},
		{"nil adapt scratch", func() {
			FDKaacEncAdaptThresholdsToPeCBR(&adaptPE, qc[:], psy[:], &adaptTools, &adaptState, nil, 1, 120, 1)
		}},
		{"bad adapt desired PE", func() {
			FDKaacEncAdaptThresholdsToPeCBR(&adaptPE, qc[:], psy[:], &adaptTools, &adaptState, &adaptScratch, 1, 0, 1)
		}},
		{"bad adapt iteration count", func() {
			FDKaacEncAdaptThresholdsToPeCBR(&adaptPE, qc[:], psy[:], &adaptTools, &adaptState, &adaptScratch, 1, 120, -1)
		}},
		{"negative adapt PE state", func() {
			bad := adaptPE
			bad.Pe = -1
			FDKaacEncAdaptThresholdsToPeCBR(&bad, qc[:], psy[:], &adaptTools, &adaptState, &adaptScratch, 1, 120, 1)
		}},
		{"nil VBR adapt tools", func() {
			FDKaacEncAdaptThresholdsVBR(qc[:], psy[:], nil, &adaptVBRState, &adaptVBRScratch, 1)
		}},
		{"nil VBR adapt state", func() {
			FDKaacEncAdaptThresholdsVBR(qc[:], psy[:], &adaptTools, nil, &adaptVBRScratch, 1)
		}},
		{"nil VBR adapt scratch", func() {
			FDKaacEncAdaptThresholdsVBR(qc[:], psy[:], &adaptTools, &adaptVBRState, nil, 1)
		}},
		{"negative VBR adapt quality", func() {
			bad := adaptVBRState
			bad.VBRQualFactor = -1
			FDKaacEncAdaptThresholdsVBR(qc[:], psy[:], &adaptTools, &bad, &adaptVBRScratch, 1)
		}},
		{"nil reduction flags", func() { FDKaacEncReduceThresholdsCBR(qc[:], psy[:], nil, &thrExp, 1, 0x20000000, 0) }},
		{"nil threshold exponent", func() { FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, nil, 1, 0x20000000, 0) }},
		{"nil reduction qc", func() {
			bad := [1]*QCOutChannel{nil}
			FDKaacEncReduceThresholdsCBR(bad[:], psy[:], &ahFlag, &thrExp, 1, 0x20000000, 0)
		}},
		{"negative reduction value", func() { FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, &thrExp, 1, -1, 0) }},
		{"bad reduction exponent", func() { FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, &thrExp, 1, 0x20000000, DfractBits+1) }},
		{"nil chaos-measure psy", func() { FDKaacEncCalcChaosMeasure(nil, qcStorage.SfbFormFactorLdData[:]) }},
		{"short chaos-measure form", func() { FDKaacEncCalcChaosMeasure(&psyStorage, qcStorage.SfbFormFactorLdData[:2]) }},
		{"nil VBR flags", func() { FDKaacEncReduceThresholdsVBR(qc[:], psy[:], nil, &thrExp, 1, peCorrectionHalf, &chaosOld) }},
		{"nil VBR exponent", func() { FDKaacEncReduceThresholdsVBR(qc[:], psy[:], &ahFlag, nil, 1, peCorrectionHalf, &chaosOld) }},
		{"nil VBR chaos history", func() { FDKaacEncReduceThresholdsVBR(qc[:], psy[:], &ahFlag, &thrExp, 1, peCorrectionHalf, nil) }},
		{"negative VBR quality", func() { FDKaacEncReduceThresholdsVBR(qc[:], psy[:], &ahFlag, &thrExp, 1, -1, &chaosOld) }},
		{"bad VBR group length", func() {
			var shortPsy PsyOutChannel
			var shortQC QCOutChannel
			var shortThrExp [2][maxGroupedSFB]FixpDBL
			var shortAHFlag [2][maxGroupedSFB]uint8
			fillVBRShortThresholdReductionCase(&shortPsy, &shortQC, &shortThrExp, &shortAHFlag)
			shortPsy.GroupLen[0] = len(adjThrInvInt)
			FDKaacEncReduceThresholdsVBR([]*QCOutChannel{&shortQC}, []*PsyOutChannel{&shortPsy}, &shortAHFlag, &shortThrExp, 1, peCorrectionHalf, &chaosOld)
		}},
		{"nil correct-threshold PE", func() {
			FDKaacEncCorrectThresholds(qc[:], psy[:], nil, &ahFlag, &thrExp, &correctScratch, 1, 0x18000000, 0, -1)
		}},
		{"nil correct-threshold flags", func() {
			FDKaacEncCorrectThresholds(qc[:], psy[:], &peData, nil, &thrExp, &correctScratch, 1, 0x18000000, 0, -1)
		}},
		{"nil correct-threshold exponent", func() {
			FDKaacEncCorrectThresholds(qc[:], psy[:], &peData, &ahFlag, nil, &correctScratch, 1, 0x18000000, 0, -1)
		}},
		{"nil correct-threshold scratch", func() {
			FDKaacEncCorrectThresholds(qc[:], psy[:], &peData, &ahFlag, &thrExp, nil, 1, 0x18000000, 0, -1)
		}},
		{"negative correct-threshold reduction", func() {
			FDKaacEncCorrectThresholds(qc[:], psy[:], &peData, &ahFlag, &thrExp, &correctScratch, 1, -1, 0, -1)
		}},
		{"bad correct-threshold exponent", func() {
			FDKaacEncCorrectThresholds(qc[:], psy[:], &peData, &ahFlag, &thrExp, &correctScratch, 1, 0x18000000, DfractBits+1, -1)
		}},
		{"negative correct-threshold active lines", func() {
			badPE := peData
			badPE.PEChannelData[0].SfbNActiveLines[0] = -1
			FDKaacEncCorrectThresholds(qc[:], psy[:], &badPE, &ahFlag, &thrExp, &correctScratch, 1, 0x18000000, 0, -1)
		}},
		{"zero correct-threshold norm", func() {
			var emptyPE PEData
			FDKaacEncCorrectThresholds(qc[:], psy[:], &emptyPE, &ahFlag, &thrExp, &correctScratch, 1, 0x18000000, 0, 1)
		}},
		{"nil reduce-min-SNR PE", func() {
			redPeGlobal := 160
			FDKaacEncReduceMinSnr(qc[:], psy[:], nil, &ahFlag, 1, 90, &redPeGlobal)
		}},
		{"nil reduce-min-SNR flags", func() {
			redPeGlobal := 160
			FDKaacEncReduceMinSnr(qc[:], psy[:], &peData, nil, 1, 90, &redPeGlobal)
		}},
		{"nil reduce-min-SNR total", func() {
			FDKaacEncReduceMinSnr(qc[:], psy[:], &peData, &ahFlag, 1, 90, nil)
		}},
		{"negative reduce-min-SNR target", func() {
			redPeGlobal := 160
			FDKaacEncReduceMinSnr(qc[:], psy[:], &peData, &ahFlag, 1, -1, &redPeGlobal)
		}},
		{"negative reduce-min-SNR band PE", func() {
			var badPsy PsyOutChannel
			var badQC QCOutChannel
			var badPE PEData
			var badFlags [2][maxGroupedSFB]uint8
			fillReduceMinSnrCase(&badPsy, &badQC, &badPE, &badFlags)
			badPE.PEChannelData[0].SfbPe[3] = -1
			redPeGlobal := 160
			FDKaacEncReduceMinSnr([]*QCOutChannel{&badQC}, []*PsyOutChannel{&badPsy}, &badPE, &badFlags, 1, 90, &redPeGlobal)
		}},
		{"nil allow-more-holes PE", func() {
			var amhPsy PsyOutChannel
			var amhQC QCOutChannel
			var amhPE PEData
			var amhTools ToolsInfo
			var amhState ATSElement
			var amhFlags [2][maxGroupedSFB]uint8
			fillAllowMoreHolesEnergyCase(&amhPsy, &amhQC, &amhPE, &amhTools, &amhState, &amhFlags)
			FDKaacEncAllowMoreHoles([]*QCOutChannel{&amhQC}, []*PsyOutChannel{&amhPsy}, nil, &amhTools, &amhState, &amhFlags, 1, 70, 100)
		}},
		{"nil allow-more-holes tools", func() {
			var amhPsy PsyOutChannel
			var amhQC QCOutChannel
			var amhPE PEData
			var amhTools ToolsInfo
			var amhState ATSElement
			var amhFlags [2][maxGroupedSFB]uint8
			fillAllowMoreHolesEnergyCase(&amhPsy, &amhQC, &amhPE, &amhTools, &amhState, &amhFlags)
			FDKaacEncAllowMoreHoles([]*QCOutChannel{&amhQC}, []*PsyOutChannel{&amhPsy}, &amhPE, nil, &amhState, &amhFlags, 1, 70, 100)
		}},
		{"nil allow-more-holes state", func() {
			var amhPsy PsyOutChannel
			var amhQC QCOutChannel
			var amhPE PEData
			var amhTools ToolsInfo
			var amhState ATSElement
			var amhFlags [2][maxGroupedSFB]uint8
			fillAllowMoreHolesEnergyCase(&amhPsy, &amhQC, &amhPE, &amhTools, &amhState, &amhFlags)
			FDKaacEncAllowMoreHoles([]*QCOutChannel{&amhQC}, []*PsyOutChannel{&amhPsy}, &amhPE, &amhTools, nil, &amhFlags, 1, 70, 100)
		}},
		{"nil allow-more-holes flags", func() {
			var amhPsy PsyOutChannel
			var amhQC QCOutChannel
			var amhPE PEData
			var amhTools ToolsInfo
			var amhState ATSElement
			var amhFlags [2][maxGroupedSFB]uint8
			fillAllowMoreHolesEnergyCase(&amhPsy, &amhQC, &amhPE, &amhTools, &amhState, &amhFlags)
			FDKaacEncAllowMoreHoles([]*QCOutChannel{&amhQC}, []*PsyOutChannel{&amhPsy}, &amhPE, &amhTools, &amhState, nil, 1, 70, 100)
		}},
		{"negative allow-more-holes current PE", func() {
			var amhPsy PsyOutChannel
			var amhQC QCOutChannel
			var amhPE PEData
			var amhTools ToolsInfo
			var amhState ATSElement
			var amhFlags [2][maxGroupedSFB]uint8
			fillAllowMoreHolesEnergyCase(&amhPsy, &amhQC, &amhPE, &amhTools, &amhState, &amhFlags)
			FDKaacEncAllowMoreHoles([]*QCOutChannel{&amhQC}, []*PsyOutChannel{&amhPsy}, &amhPE, &amhTools, &amhState, &amhFlags, 1, 70, -1)
		}},
		{"bad allow-more-holes start band", func() {
			var amhPsy PsyOutChannel
			var amhQC QCOutChannel
			var amhPE PEData
			var amhTools ToolsInfo
			var amhState ATSElement
			var amhFlags [2][maxGroupedSFB]uint8
			fillAllowMoreHolesEnergyCase(&amhPsy, &amhQC, &amhPE, &amhTools, &amhState, &amhFlags)
			amhState.AHParam.StartSfbL = -1
			FDKaacEncAllowMoreHoles([]*QCOutChannel{&amhQC}, []*PsyOutChannel{&amhPsy}, &amhPE, &amhTools, &amhState, &amhFlags, 1, 70, 100)
		}},
		{"negative allow-more-holes band PE", func() {
			var amhPsy PsyOutChannel
			var amhQC QCOutChannel
			var amhPE PEData
			var amhTools ToolsInfo
			var amhState ATSElement
			var amhFlags [2][maxGroupedSFB]uint8
			fillAllowMoreHolesEnergyCase(&amhPsy, &amhQC, &amhPE, &amhTools, &amhState, &amhFlags)
			amhPE.PEChannelData[0].SfbPe[3] = -1
			FDKaacEncAllowMoreHoles([]*QCOutChannel{&amhQC}, []*PsyOutChannel{&amhPsy}, &amhPE, &amhTools, &amhState, &amhFlags, 1, 70, 100)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncDistributeBitsRejectsInvalid(t *testing.T) {
	var state AdjThrState
	FDKaacEncInitBitresState(&state)
	element := buildBitresElement()
	psy0 := PsyOutChannel{LastWindowSequence: LongWindow}
	psy := [1]*PsyOutChannel{&psy0}
	peData := PEData{Pe: 430}

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil state", func() {
			FDKaacEncDistributeBits(nil, &element, psy[:], &peData, 1, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"nil element", func() {
			FDKaacEncDistributeBits(&state, nil, psy[:], &peData, 1, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"nil PE data", func() {
			FDKaacEncDistributeBits(&state, &element, psy[:], nil, 1, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"bad channel count", func() {
			FDKaacEncDistributeBits(&state, &element, psy[:], &peData, 0, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"short psy", func() {
			FDKaacEncDistributeBits(&state, &element, psy[:], &peData, 2, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"nil psy", func() {
			bad := [1]*PsyOutChannel{nil}
			FDKaacEncDistributeBits(&state, &element, bad[:], &peData, 1, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"bad window", func() {
			badPsy := PsyOutChannel{LastWindowSequence: -1}
			bad := [1]*PsyOutChannel{&badPsy}
			FDKaacEncDistributeBits(&state, &element, bad[:], &peData, 1, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"negative reservoir", func() {
			FDKaacEncDistributeBits(&state, &element, psy[:], &peData, 1, 0, 720, -1, 1200, MaxValDBL, BitresModeFull)
		}},
		{"bad mode", func() {
			FDKaacEncDistributeBits(&state, &element, psy[:], &peData, 1, 0, 720, 500, 1200, MaxValDBL, BitresMode(99))
		}},
		{"bad factor", func() {
			bad := element
			bad.Bits2PeFactorM = 0
			FDKaacEncDistributeBits(&state, &bad, psy[:], &peData, 1, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"bad bounds", func() {
			bad := element
			bad.PeMax = bad.PeMin
			FDKaacEncDistributeBits(&state, &bad, psy[:], &peData, 1, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
		{"uninitialized reservoir params", func() {
			var badState AdjThrState
			FDKaacEncDistributeBits(&badState, &element, psy[:], &peData, 1, 0, 720, 500, 1200, MaxValDBL, BitresModeFull)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncAdjThrInitRejectsInvalid(t *testing.T) {
	var state AdjThrState
	var element ATSElement

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil adjustment state", func() {
			FDKaacEncInitAdjThrState(nil, 1, 0, 0)
		}},
		{"bad element count", func() {
			FDKaacEncInitAdjThrState(&state, 0, 0, 0)
		}},
		{"nil element", func() {
			FDKaacEncInitATSElement(nil, 1000, 64000, 2, 48000, 0, 0, 0, peCorrectionHalf)
		}},
		{"bad mean PE", func() {
			FDKaacEncInitATSElement(&element, 0, 64000, 2, 48000, 0, 0, 0, peCorrectionHalf)
		}},
		{"bad bitrate", func() {
			FDKaacEncInitATSElement(&element, 1000, 0, 2, 48000, 0, 0, 0, peCorrectionHalf)
		}},
		{"bad channel count", func() {
			FDKaacEncInitATSElement(&element, 1000, 64000, 0, 48000, 0, 0, 0, peCorrectionHalf)
		}},
		{"bad sample rate", func() {
			FDKaacEncInitATSElement(&element, 1000, 64000, 2, 0, 0, 0, 0, peCorrectionHalf)
		}},
		{"bad VBR quality", func() {
			FDKaacEncInitATSElement(&element, 1000, 64000, 2, 48000, 0, 0, 0, -1)
		}},
		{"bad bits-to-PE configuration", func() {
			FDKaacEncInitBits2PeFactor(0, 2, 48000, 0, 0, 0)
		}},
		{"unsupported advanced bits-to-PE", func() {
			FDKaacEncInitBits2PeFactor(64000, 2, 48000, 1, 0, 0)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncAdjustThresholdsRejectsInvalid(t *testing.T) {
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil adjustment state", func() {
			var peData PEData
			var psy PsyOutChannel
			var qc QCOutChannel
			var tools ToolsInfo
			var elemState ATSElement
			var adjState AdjThrState
			var qcElement QCOutElement
			var qcOut QCOut
			var psyElement PsyOutElement
			var cm ChannelMapping
			var scratch AdjustThresholdsScratch
			fillAdjustThresholdsCBRCase(&peData, &psy, &qc, &tools, &elemState, &adjState, &qcElement, &qcOut, &psyElement, &cm, 160)
			qcElements := [1]*QCOutElement{&qcElement}
			psyElements := [1]*PsyOutElement{&psyElement}
			FDKaacEncAdjustThresholds(nil, qcElements[:], &qcOut, psyElements[:], 1, &cm, &scratch)
		}},
		{"nil qc output", func() {
			var peData PEData
			var psy PsyOutChannel
			var qc QCOutChannel
			var tools ToolsInfo
			var elemState ATSElement
			var adjState AdjThrState
			var qcElement QCOutElement
			var qcOut QCOut
			var psyElement PsyOutElement
			var cm ChannelMapping
			var scratch AdjustThresholdsScratch
			fillAdjustThresholdsCBRCase(&peData, &psy, &qc, &tools, &elemState, &adjState, &qcElement, &qcOut, &psyElement, &cm, 160)
			qcElements := [1]*QCOutElement{&qcElement}
			psyElements := [1]*PsyOutElement{&psyElement}
			FDKaacEncAdjustThresholds(&adjState, qcElements[:], nil, psyElements[:], 1, &cm, &scratch)
		}},
		{"nil channel mapping", func() {
			var peData PEData
			var psy PsyOutChannel
			var qc QCOutChannel
			var tools ToolsInfo
			var elemState ATSElement
			var adjState AdjThrState
			var qcElement QCOutElement
			var qcOut QCOut
			var psyElement PsyOutElement
			var cm ChannelMapping
			var scratch AdjustThresholdsScratch
			fillAdjustThresholdsCBRCase(&peData, &psy, &qc, &tools, &elemState, &adjState, &qcElement, &qcOut, &psyElement, &cm, 160)
			qcElements := [1]*QCOutElement{&qcElement}
			psyElements := [1]*PsyOutElement{&psyElement}
			FDKaacEncAdjustThresholds(&adjState, qcElements[:], &qcOut, psyElements[:], 1, nil, &scratch)
		}},
		{"nil scratch", func() {
			var peData PEData
			var psy PsyOutChannel
			var qc QCOutChannel
			var tools ToolsInfo
			var elemState ATSElement
			var adjState AdjThrState
			var qcElement QCOutElement
			var qcOut QCOut
			var psyElement PsyOutElement
			var cm ChannelMapping
			fillAdjustThresholdsCBRCase(&peData, &psy, &qc, &tools, &elemState, &adjState, &qcElement, &qcOut, &psyElement, &cm, 160)
			qcElements := [1]*QCOutElement{&qcElement}
			psyElements := [1]*PsyOutElement{&psyElement}
			FDKaacEncAdjustThresholds(&adjState, qcElements[:], &qcOut, psyElements[:], 1, &cm, nil)
		}},
		{"invalid element PE budget", func() {
			var peData PEData
			var psy PsyOutChannel
			var qc QCOutChannel
			var tools ToolsInfo
			var elemState ATSElement
			var adjState AdjThrState
			var qcElement QCOutElement
			var qcOut QCOut
			var psyElement PsyOutElement
			var cm ChannelMapping
			var scratch AdjustThresholdsScratch
			fillAdjustThresholdsCBRCase(&peData, &psy, &qc, &tools, &elemState, &adjState, &qcElement, &qcOut, &psyElement, &cm, 160)
			qcOut.TotalGrantedPeCorr = qcOut.TotalNoRedPe
			qcElement.StaticBitsUsed = minBufSizePerEffChan + 1
			qcElements := [1]*QCOutElement{&qcElement}
			psyElements := [1]*PsyOutElement{&psyElement}
			FDKaacEncAdjustThresholds(&adjState, qcElements[:], &qcOut, psyElements[:], 1, &cm, &scratch)
		}},
		{"nil element state", func() {
			var peData PEData
			var psy PsyOutChannel
			var qc QCOutChannel
			var tools ToolsInfo
			var elemState ATSElement
			var adjState AdjThrState
			var qcElement QCOutElement
			var qcOut QCOut
			var psyElement PsyOutElement
			var cm ChannelMapping
			var scratch AdjustThresholdsScratch
			fillAdjustThresholdsCBRCase(&peData, &psy, &qc, &tools, &elemState, &adjState, &qcElement, &qcOut, &psyElement, &cm, 160)
			adjState.AdjThrStateElem[0] = nil
			qcElements := [1]*QCOutElement{&qcElement}
			psyElements := [1]*PsyOutElement{&psyElement}
			FDKaacEncAdjustThresholds(&adjState, qcElements[:], &qcOut, psyElements[:], 1, &cm, &scratch)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncAdjThrInitAllocs(t *testing.T) {
	var state AdjThrState
	var element ATSElement

	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncInitAdjThrState(&state, 2, 0, 1)
		FDKaacEncInitATSElement(&element, 1000, 64000, 2, 48000, 0, 0, 0, peCorrectionHalf)
		adjThrSink = state.BRESParamLong.MaxBitSpend + element.Bits2PeFactorM
		adjThrHashSink = uint64(element.PeMin + state.MaxIter2ndGuess)
	})
	if allocs != 0 {
		t.Fatalf("adjustment init allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncAdjustThresholdsAllocs(t *testing.T) {
	var peData PEData
	var peData1 PEData
	var psy PsyOutChannel
	var psy1 PsyOutChannel
	var qc QCOutChannel
	var qc1 QCOutChannel
	var tools ToolsInfo
	var tools1 ToolsInfo
	var elemState ATSElement
	var elemState1 ATSElement
	var adjState AdjThrState
	var adjState1 AdjThrState
	var qcElement QCOutElement
	var qcElement1 QCOutElement
	var qcOut QCOut
	var qcOut1 QCOut
	var psyElement PsyOutElement
	var psyElement1 PsyOutElement
	var cm ChannelMapping
	var cm1 ChannelMapping
	var scratch AdjustThresholdsScratch
	qcElements := [1]*QCOutElement{&qcElement}
	psyElements := [1]*PsyOutElement{&psyElement}
	qcElements2 := [2]*QCOutElement{&qcElement, &qcElement1}
	psyElements2 := [2]*PsyOutElement{&psyElement, &psyElement1}

	allocs := testing.AllocsPerRun(1000, func() {
		scratch = AdjustThresholdsScratch{}
		fillAdjustThresholdsCBRCase(&peData, &psy, &qc, &tools, &elemState, &adjState, &qcElement, &qcOut, &psyElement, &cm, 160)
		result := FDKaacEncAdjustThresholds(&adjState, qcElements[:], &qcOut, psyElements[:], 1, &cm, &scratch)
		fillAdjustThresholdsCBRCase(&peData, &psy, &qc, &tools, &elemState, &adjState, &qcElement, &qcOut, &psyElement, &cm, 160)
		fillAdjustThresholdsCBRCase(&peData1, &psy1, &qc1, &tools1, &elemState1, &adjState1, &qcElement1, &qcOut1, &psyElement1, &cm1, 160)
		adjState.AdjThrStateElem[1] = &elemState1
		cm.NElements = 2
		cm.ElInfo[0].ChannelIndex[0] = 0
		cm.ElInfo[1] = ElementInfo{ElType: idSCE, NChannelsInEl: 1, ChannelIndex: [2]int{1, 0}}
		qcOut.TotalNoRedPe = int(qcElement.PEData.Pe + qcElement1.PEData.Pe)
		qcOut.TotalGrantedPeCorr = 260
		multiResult := FDKaacEncAdjustThresholds(&adjState, qcElements2[:], &qcOut, psyElements2[:], 1, &cm, &scratch)
		adjThrSink = qc.SfbThresholdLdData[0] + FixpDBL(result.RedPe)
		adjThrHashSink = hashFixpDBL(qc.SfbThresholdLdData[:8]) ^ uint64(multiResult.RedPe)
	})
	if allocs != 0 {
		t.Fatalf("adjust thresholds allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncMinSnrAdjustmentAllocs(t *testing.T) {
	var param MinSNRAdaptParam
	var thrExp [2][maxGroupedSFB]FixpDBL
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncInitMinSnrAdaptParam(&param)
		psyStorage, qcStorage = buildMinSnrAdaptCase()
		FDKaacEncCalcThresholdExp(&thrExp, qc[:], psy[:], 1)
		FDKaacEncAdaptMinSnr(qc[:], psy[:], &param, 1)
		adjThrSink = thrExp[0][0] + qcStorage.SfbMinSnrLdData[7]
	})
	if allocs != 0 {
		t.Fatalf("min-SNR adaptation allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncAvoidHoleAllocs(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var tools ToolsInfo
	var ahFlag [2][maxGroupedSFB]uint8
	ahParam := AHParam{ModifyMinSnr: 1}
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	spreadEnergy := [...]FixpDBL{100000000, 2000000, 10000000, 200000000, 50000000, 3000000, 2000000, 40000000}

	allocs := testing.AllocsPerRun(1000, func() {
		ahFlag = [2][maxGroupedSFB]uint8{}
		psyStorage, qcStorage = buildMinSnrAdaptCase()
		copy(qcStorage.SfbEnergy[:], psyStorage.SfbEnergy[:])
		copy(qcStorage.SfbSpreadEnergy[:], spreadEnergy[:])
		FDKaacEncInitAvoidHoleFlag(qc[:], psy[:], &ahFlag, &tools, 1, &ahParam)
		adjThrSink = qcStorage.SfbSpreadEnergy[2] + qcStorage.SfbMinSnrLdData[6]
		adjThrHashSink = uint64(ahFlag[0][2])
	})
	if allocs != 0 {
		t.Fatalf("avoid-hole allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncThresholdReductionAllocs(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var peData PEData
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	allocs := testing.AllocsPerRun(1000, func() {
		fillThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
		FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, &thrExp, 1, 0x20000000, 0)
		fillPENoAHCase(&peData, &psyStorage, &ahFlag)
		pe, constPart, nActiveLines := FDKaacEncCalcPENoAH(&peData, &ahFlag, psy[:], 1)
		redM, redE := FDKaacEncCalcRedValPower(FixpDBL(constPart-pe), FixpDBL(4*nActiveLines))
		reductionM, reductionE := fdkaacEncInitialReductionValue(constPart, pe, pe/2, maxInt(nActiveLines, 1))
		adjThrSink = qcStorage.SfbThresholdLdData[0] + FixpDBL(pe+constPart+nActiveLines) + redM + FixpDBL(redE) + reductionM + FixpDBL(reductionE)
		adjThrHashSink = uint64(ahFlag[0][3])
	})
	if allocs != 0 {
		t.Fatalf("threshold reduction allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncVBRThresholdReductionAllocs(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	var tools ToolsInfo
	var state ATSElement
	var scratch AdaptThresholdsVBRScratch
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	allocs := testing.AllocsPerRun(1000, func() {
		fillVBRLongThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
		chaosOld := vbrChaosHalf
		FDKaacEncReduceThresholdsVBR(qc[:], psy[:], &ahFlag, &thrExp, 1, peCorrectionHalf, &chaosOld)
		fillVBRLongThresholdReductionCase(&psyStorage, &qcStorage, &thrExp, &ahFlag)
		state = ATSElement{AHParam: AHParam{ModifyMinSnr: 1}, VBRQualFactor: peCorrectionHalf, ChaosMeasureOld: vbrChaosHalf}
		FDKaacEncInitMinSnrAdaptParam(&state.MinSNRAdaptParam)
		FDKaacEncAdaptThresholdsVBR(qc[:], psy[:], &tools, &state, &scratch, 1)
		adjThrSink = qcStorage.SfbThresholdLdData[0] + chaosOld + state.ChaosMeasureOld + scratch.ThrExp[0][0] + FDKaacEncCalcChaosMeasure(&psyStorage, qcStorage.SfbFormFactorLdData[:])
		adjThrHashSink = uint64(ahFlag[0][3])
	})
	if allocs != 0 {
		t.Fatalf("VBR threshold reduction allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncCorrectThresholdsAllocs(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var peData PEData
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	var scratch CorrectThresholdScratch
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	allocs := testing.AllocsPerRun(1000, func() {
		fillCorrectThresholdCase(&psyStorage, &qcStorage, &peData, &thrExp, &ahFlag)
		FDKaacEncCorrectThresholds(qc[:], psy[:], &peData, &ahFlag, &thrExp, &scratch, 1, 0x18000000, 0, -34)
		adjThrSink = qcStorage.SfbThresholdLdData[0] + scratch.SfbPEFactorsLdData[0][0] + scratch.SfbNActiveLinesLdData[0][1]
		adjThrHashSink = hashFixpDBL(qcStorage.SfbThresholdLdData[:8])
	})
	if allocs != 0 {
		t.Fatalf("correct-threshold allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncReduceMinSnrAllocs(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var peData PEData
	var ahFlag [2][maxGroupedSFB]uint8
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	allocs := testing.AllocsPerRun(1000, func() {
		fillReduceMinSnrCase(&psyStorage, &qcStorage, &peData, &ahFlag)
		redPeGlobal := 160
		FDKaacEncReduceMinSnr(qc[:], psy[:], &peData, &ahFlag, 1, 90, &redPeGlobal)
		adjThrSink = qcStorage.SfbThresholdLdData[6] + peData.PEChannelData[0].SfbPe[7] + FixpDBL(redPeGlobal)
		adjThrHashSink = hashFixpDBL(qcStorage.SfbMinSnrLdData[:8])
	})
	if allocs != 0 {
		t.Fatalf("reduce-min-SNR allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncAllowMoreHolesAllocs(t *testing.T) {
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var peData PEData
	var tools ToolsInfo
	var state ATSElement
	var ahFlag [2][maxGroupedSFB]uint8
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}

	allocs := testing.AllocsPerRun(1000, func() {
		fillAllowMoreHolesEnergyCase(&psyStorage, &qcStorage, &peData, &tools, &state, &ahFlag)
		FDKaacEncAllowMoreHoles(qc[:], psy[:], &peData, &tools, &state, &ahFlag, 1, 70, 100)
		adjThrSink = qcStorage.SfbThresholdLdData[3] + FixpDBL(ahFlag[0][3]) + peData.PEChannelData[0].SfbPe[3]
		adjThrHashSink = hashFixpDBL(qcStorage.SfbThresholdLdData[:8])
	})
	if allocs != 0 {
		t.Fatalf("allow-more-holes allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncResetAHFlagsAllocs(t *testing.T) {
	psy0 := PsyOutChannel{
		SfbCnt:             8,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     3,
		LastWindowSequence: LongWindow,
	}
	for i := 0; i <= 8; i++ {
		psy0.SfbOffsets[i] = i
	}
	psy := [1]*PsyOutChannel{&psy0}
	var ahFlag [2][maxGroupedSFB]uint8

	allocs := testing.AllocsPerRun(1000, func() {
		ahFlag[0] = [maxGroupedSFB]uint8{AvoidHoleActive, AvoidHoleNone, AvoidHoleActive, AvoidHoleActive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleActive}
		FDKaacEncResetAHFlags(&ahFlag, psy[:], 1)
		adjThrHashSink = uint64(ahFlag[0][0])<<8 | uint64(ahFlag[0][3])
	})
	if allocs != 0 {
		t.Fatalf("reset AH flag allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncPECalculationAllocs(t *testing.T) {
	var peData PEData
	var psyStorage PsyOutChannel
	var qcStorage QCOutChannel
	var tools ToolsInfo
	var state ATSElement
	psy := [1]*PsyOutChannel{&psyStorage}
	qc := [1]*QCOutChannel{&qcStorage}
	allocs := testing.AllocsPerRun(1000, func() {
		fillAdjThrLongPatchCase(&peData, &psyStorage, &qcStorage, &tools, &state)
		FDKaacEncPECalculation(&peData, psy[:], qc[:], &tools, &state, 1)
		adjThrSink = peData.Pe + peData.ConstPart + peData.NActiveLines + qc[0].SfbEnFacLd[0]
		adjThrHashSink = hashFixpDBL(qc[0].SfbWeightedEnergyLdData[:8])
	})
	if allocs != 0 {
		t.Fatalf("PE calculation allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncDistributeBitsAllocs(t *testing.T) {
	var state AdjThrState
	var element ATSElement
	var psy0 PsyOutChannel
	var psy1 PsyOutChannel
	var peData PEData
	psy := [2]*PsyOutChannel{&psy0, &psy1}

	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncInitBitresState(&state)
		element = buildBitresElementWithHistoryAndBounds(300, 520, 160, 500)
		psy0 = PsyOutChannel{LastWindowSequence: LongWindow}
		psy1 = PsyOutChannel{LastWindowSequence: ShortWindow}
		peData = PEData{Pe: 620}

		grantedPE, grantedPECorr := FDKaacEncDistributeBits(
			&state,
			&element,
			psy[:],
			&peData,
			2,
			0,
			512,
			900,
			1000,
			MaxValDBL,
			BitresModeFull,
		)
		adjThrSink = FixpDBL(grantedPE + grantedPECorr + element.PeMin + element.PeMax)
	})
	if allocs != 0 {
		t.Fatalf("bit distribution allocations = %v, want 0", allocs)
	}
}

func assertUint8Slice(t *testing.T, name string, got []uint8, want []uint8) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %d, want %d; got %v want %v", name, i, got[i], want[i], got, want)
		}
	}
}

func buildAdjThrLongPatchCase() (PEData, [1]*PsyOutChannel, [1]*QCOutChannel, ToolsInfo, ATSElement) {
	var peData PEData
	var psy PsyOutChannel
	var qc QCOutChannel
	var tools ToolsInfo
	var state ATSElement
	fillAdjThrLongPatchCase(&peData, &psy, &qc, &tools, &state)
	return peData, [1]*PsyOutChannel{&psy}, [1]*QCOutChannel{&qc}, tools, state
}

func fillAdjThrLongPatchCase(peData *PEData, psy *PsyOutChannel, qc *QCOutChannel, tools *ToolsInfo, state *ATSElement) {
	*peData = PEData{}
	*psy = PsyOutChannel{}
	*qc = QCOutChannel{}
	*tools = ToolsInfo{}
	*state = ATSElement{PeOffset: 17, LastEnFacPatch: [2]int{1, 1}}

	energy, _, form, offsets, isBook, isScale := linePeVectorInputs()
	threshold := [...]FixpDBL{-520000000, -500000000, -510000000, -530000000, -500000000, -505000000, -490000000, -540000000}

	psy.SfbCnt = 8
	psy.SfbPerGroup = 4
	psy.MaxSfbPerGroup = 4
	psy.LastWindowSequence = LongWindow
	copy(psy.SfbOffsets[:], offsets[:])
	copy(psy.SfbEnergyLdData[:], energy[:])
	copy(psy.SfbThresholdLdData[:], threshold[:])
	copy(psy.IsBook[:], isBook[:])
	copy(psy.IsScale[:], isScale[:])
	for i := 0; i < 8; i++ {
		psy.SfbEnergy[i] = CalcInvLdData(energy[i])
	}

	for i := 0; i < 8; i++ {
		qc.SfbFormFactorLdData[i] = form[i] + 180000000
	}
	copy(qc.SfbEnergyLdData[:], energy[:])
	copy(qc.SfbThresholdLdData[:], threshold[:])
}

func buildAdjThrShortCase() (PEData, [1]*PsyOutChannel, [1]*QCOutChannel, ToolsInfo, ATSElement) {
	peData, psy, qc, tools, state := buildAdjThrLongPatchCase()
	peData = PEData{}
	state = ATSElement{PeOffset: 11}
	psy[0].LastWindowSequence = ShortWindow
	psy[0].SfbCnt = 8
	psy[0].SfbPerGroup = 4
	psy[0].MaxSfbPerGroup = 3
	return peData, psy, qc, tools, state
}

const (
	defaultBitresBits2PEFactorM FixpDBL = 0x4b851e80
	defaultBitresBits2PEFactorE         = 1
)

func buildBitresElement() ATSElement {
	return buildBitresElementWithHistoryAndBounds(0, -1, 180, 620)
}

func buildBitresElementWithHistory(peLast int, dynBitsLast int) ATSElement {
	return buildBitresElementWithHistoryAndBounds(peLast, dynBitsLast, 180, 620)
}

func buildBitresElementWithHistoryAndBounds(peLast int, dynBitsLast int, peMin int, peMax int) ATSElement {
	return ATSElement{
		PeMin:               peMin,
		PeMax:               peMax,
		Bits2PeFactorM:      defaultBitresBits2PEFactorM,
		Bits2PeFactorE:      defaultBitresBits2PEFactorE,
		PeLast:              peLast,
		DynBitsLast:         dynBitsLast,
		PeCorrectionFactorM: peCorrectionHalf,
		PeCorrectionFactorE: 1,
	}
}

func primeAdjustThresholdElement(state *ATSElement) {
	if state.PeMin == 0 && state.PeMax == 0 {
		state.PeMin = 180
		state.PeMax = 620
	}
	state.Bits2PeFactorM = defaultBitresBits2PEFactorM
	state.Bits2PeFactorE = defaultBitresBits2PEFactorE
	state.PeCorrectionFactorM = peCorrectionHalf
	state.PeCorrectionFactorE = 1
	state.DynBitsLast = -1
	state.AHParam = AHParam{ModifyMinSnr: 1, StartSfbL: 4, StartSfbS: 3}
	state.VBRQualFactor = peCorrectionHalf
	if state.ChaosMeasureOld == 0 {
		state.ChaosMeasureOld = vbrChaosHalf
	}
	FDKaacEncInitMinSnrAdaptParam(&state.MinSNRAdaptParam)
}

func fillAdjustThresholdsCBRCase(
	peData *PEData,
	psy *PsyOutChannel,
	qc *QCOutChannel,
	tools *ToolsInfo,
	elemState *ATSElement,
	adjState *AdjThrState,
	qcElement *QCOutElement,
	qcOut *QCOut,
	psyElement *PsyOutElement,
	cm *ChannelMapping,
	desiredPe int,
) {
	fillAdjThrLongPatchCase(peData, psy, qc, tools, elemState)
	primeAdjustThresholdElement(elemState)
	psyChannels := [1]*PsyOutChannel{psy}
	qcChannels := [1]*QCOutChannel{qc}
	FDKaacEncPECalculation(peData, psyChannels[:], qcChannels[:], tools, elemState, 1)

	*adjState = AdjThrState{BitDistributionMode: BitDistributionModeInterElement, MaxIter2ndGuess: 1}
	adjState.AdjThrStateElem[0] = elemState
	*qcElement = QCOutElement{
		GrantedPeCorr: desiredPe,
		PEData:        *peData,
		QCOutChannel:  [2]*QCOutChannel{qc},
	}
	*qcOut = QCOut{
		TotalNoRedPe:       int(peData.Pe),
		TotalGrantedPeCorr: desiredPe,
	}
	*psyElement = PsyOutElement{
		ToolsInfo:     *tools,
		PsyOutChannel: [2]*PsyOutChannel{psy},
	}
	*cm = ChannelMapping{NElements: 1}
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
}

func fillAdjustThresholdsVBRCase(psy *PsyOutChannel, qc *QCOutChannel, state *ATSElement, tools *ToolsInfo) {
	var thrExp [2][maxGroupedSFB]FixpDBL
	var ahFlag [2][maxGroupedSFB]uint8
	fillVBRLongThresholdReductionCase(psy, qc, &thrExp, &ahFlag)
	*tools = ToolsInfo{}
	*state = ATSElement{VBRQualFactor: peCorrectionHalf, ChaosMeasureOld: vbrChaosHalf}
	primeAdjustThresholdElement(state)
}

func buildMinSnrAdaptCase() (PsyOutChannel, QCOutChannel) {
	var psy PsyOutChannel
	var qc QCOutChannel
	psy.SfbCnt = 8
	psy.SfbPerGroup = 4
	psy.MaxSfbPerGroup = 4
	psy.LastWindowSequence = LongWindow
	for i := 0; i <= 8; i++ {
		psy.SfbOffsets[i] = i
	}

	energy := [...]FixpDBL{200000000, 4000000, 2000000, 60000000, 180000000, 1000000, 3000000, 90000000}
	energyLd := [...]FixpDBL{-114909676, -304286066, -337840498, -173192572, -120010024, -371394930, -318212408, -153564456}
	threshold := [...]FixpDBL{-520000000, -500000000, -510000000, -530000000, -500000000, -505000000, -490000000, -540000000}
	minSnr := [...]FixpDBL{-90000000, -120000000, -150000000, -80000000, -70000000, -160000000, -110000000, -100000000}
	copy(psy.SfbEnergy[:], energy[:])
	copy(qc.SfbEnergyLdData[:], energyLd[:])
	copy(qc.SfbThresholdLdData[:], threshold[:])
	copy(qc.SfbMinSnrLdData[:], minSnr[:])
	return psy, qc
}

func buildPENoAHCase() (PEData, PsyOutChannel, [2][maxGroupedSFB]uint8) {
	var peData PEData
	var psy PsyOutChannel
	var ahFlag [2][maxGroupedSFB]uint8
	fillPENoAHCase(&peData, &psy, &ahFlag)
	return peData, psy, ahFlag
}

func fillPENoAHCase(peData *PEData, psy *PsyOutChannel, ahFlag *[2][maxGroupedSFB]uint8) {
	*peData = PEData{Offset: 17}
	*psy = PsyOutChannel{
		SfbCnt:             8,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     4,
		LastWindowSequence: LongWindow,
	}
	*ahFlag = [2][maxGroupedSFB]uint8{}
	for i := 0; i <= 8; i++ {
		psy.SfbOffsets[i] = i
	}

	peValues := [...]int{1, 2, 4, 8, 16, 32, 64, 128}
	constValues := [...]int{10, 20, 40, 80, 160, 320, 640, 1280}
	activeLines := [...]int{1, 2, 4, 8, 16, 32, 64, 128}
	flags := [...]uint8{AvoidHoleNone, AvoidHoleInactive, AvoidHoleActive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone}
	for i := 0; i < 8; i++ {
		peData.PEChannelData[0].SfbPe[i] = FixpDBL(peValues[i] << peConstPartShift)
		peData.PEChannelData[0].SfbConstPart[i] = FixpDBL(constValues[i] << peConstPartShift)
		peData.PEChannelData[0].SfbNActiveLines[i] = FixpDBL(activeLines[i])
		ahFlag[0][i] = flags[i]
	}
}

func fillThresholdReductionCase(
	psy *PsyOutChannel,
	qc *QCOutChannel,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	ahFlag *[2][maxGroupedSFB]uint8,
) {
	*psy, *qc = buildMinSnrAdaptCase()
	copy(qc.SfbWeightedEnergyLdData[:], qc.SfbEnergyLdData[:])
	*thrExp = [2][maxGroupedSFB]FixpDBL{}
	*ahFlag = [2][maxGroupedSFB]uint8{}
	flags := [...]uint8{AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive}
	copy(ahFlag[0][:], flags[:])

	psyPtrs := [1]*PsyOutChannel{psy}
	qcPtrs := [1]*QCOutChannel{qc}
	FDKaacEncCalcThresholdExp(thrExp, qcPtrs[:], psyPtrs[:], 1)
}

func fillVBRLongThresholdReductionCase(
	psy *PsyOutChannel,
	qc *QCOutChannel,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	ahFlag *[2][maxGroupedSFB]uint8,
) {
	var peData PEData
	var tools ToolsInfo
	var state ATSElement
	fillAdjThrLongPatchCase(&peData, psy, qc, &tools, &state)
	copy(qc.SfbWeightedEnergyLdData[:], qc.SfbEnergyLdData[:])
	*thrExp = [2][maxGroupedSFB]FixpDBL{}
	*ahFlag = [2][maxGroupedSFB]uint8{}
	flags := [...]uint8{AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive}
	copy(ahFlag[0][:], flags[:])

	psyLocal := [1]*PsyOutChannel{psy}
	qcLocal := [1]*QCOutChannel{qc}
	FDKaacEncCalcThresholdExp(thrExp, qcLocal[:], psyLocal[:], 1)
}

func fillVBRShortThresholdReductionCase(
	psy *PsyOutChannel,
	qc *QCOutChannel,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	ahFlag *[2][maxGroupedSFB]uint8,
) {
	*psy, *qc = buildMinSnrAdaptCase()
	psy.LastWindowSequence = ShortWindow
	psy.MaxSfbPerGroup = 3
	psy.GroupLen = [maxNoOfGroups]int{3, 3, 0, 0}
	copy(psy.SfbEnergyLdData[:], qc.SfbEnergyLdData[:])
	copy(psy.SfbThresholdLdData[:], qc.SfbThresholdLdData[:])
	copy(qc.SfbWeightedEnergyLdData[:], qc.SfbEnergyLdData[:])
	*thrExp = [2][maxGroupedSFB]FixpDBL{}
	*ahFlag = [2][maxGroupedSFB]uint8{}
	flags := [...]uint8{AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, 0, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, 0}
	copy(ahFlag[0][:], flags[:])

	psyLocal := [1]*PsyOutChannel{psy}
	qcLocal := [1]*QCOutChannel{qc}
	FDKaacEncCalcThresholdExp(thrExp, qcLocal[:], psyLocal[:], 1)
}

func fillCorrectThresholdCase(
	psy *PsyOutChannel,
	qc *QCOutChannel,
	peData *PEData,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	ahFlag *[2][maxGroupedSFB]uint8,
) {
	*psy = PsyOutChannel{
		SfbCnt:             8,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     4,
		LastWindowSequence: LongWindow,
	}
	*qc = QCOutChannel{}
	*peData = PEData{}
	*thrExp = [2][maxGroupedSFB]FixpDBL{}
	*ahFlag = [2][maxGroupedSFB]uint8{}
	for i := 0; i <= 8; i++ {
		psy.SfbOffsets[i] = i
	}

	threshold := [...]FixpDBL{-520000000, -500000000, -510000000, -530000000, -500000000, -505000000, -490000000, -540000000}
	weighted := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000, -260000000, -240000000, -200000000, -330000000}
	minSnr := [...]FixpDBL{-90000000, -120000000, -150000000, -80000000, -70000000, -160000000, -110000000, -100000000}
	activeLines := [...]FixpDBL{1, 3, 0, 5, 7, 2, 4, 6}
	flags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleActive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleNone, AvoidHoleInactive}
	copy(qc.SfbThresholdLdData[:], threshold[:])
	copy(qc.SfbWeightedEnergyLdData[:], weighted[:])
	copy(qc.SfbMinSnrLdData[:], minSnr[:])
	copy(peData.PEChannelData[0].SfbNActiveLines[:], activeLines[:])
	copy(ahFlag[0][:], flags[:])

	psyLocal := [1]*PsyOutChannel{psy}
	qcLocal := [1]*QCOutChannel{qc}
	FDKaacEncCalcThresholdExp(thrExp, qcLocal[:], psyLocal[:], 1)
}

func fillReduceMinSnrCase(
	psy *PsyOutChannel,
	qc *QCOutChannel,
	peData *PEData,
	ahFlag *[2][maxGroupedSFB]uint8,
) {
	*psy = PsyOutChannel{
		SfbCnt:             8,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     4,
		LastWindowSequence: LongWindow,
	}
	*qc = QCOutChannel{}
	*peData = PEData{Pe: 160}
	peData.PEChannelData[0].Pe = 160
	*ahFlag = [2][maxGroupedSFB]uint8{}
	for i := 0; i <= 8; i++ {
		psy.SfbOffsets[i] = i
	}

	threshold := [...]FixpDBL{-520000000, -500000000, -510000000, -530000000, -500000000, -505000000, -490000000, -540000000}
	weighted := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000, -260000000, -240000000, -200000000, -330000000}
	minSnr := [...]FixpDBL{-90000000, -120000000, -150000000, -80000000, -70000000, -160000000, -110000000, -100000000}
	nLines := [...]int{2, 3, 4, 20, 5, 6, 10, 24}
	pe := [...]FixpDBL{
		11 << peConstPartShift,
		22 << peConstPartShift,
		80 << peConstPartShift,
		60 << peConstPartShift,
		15 << peConstPartShift,
		25 << peConstPartShift,
		35 << peConstPartShift,
		70 << peConstPartShift,
	}
	flags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleInactive, AvoidHoleInactive}
	copy(qc.SfbThresholdLdData[:], threshold[:])
	copy(qc.SfbWeightedEnergyLdData[:], weighted[:])
	copy(qc.SfbMinSnrLdData[:], minSnr[:])
	copy(peData.PEChannelData[0].SfbNLines[:], nLines[:])
	copy(peData.PEChannelData[0].SfbPe[:], pe[:])
	copy(ahFlag[0][:], flags[:])
}

func fillAllowMoreHolesMSCase(
	leftPsy *PsyOutChannel,
	rightPsy *PsyOutChannel,
	leftQC *QCOutChannel,
	rightQC *QCOutChannel,
	peData *PEData,
	tools *ToolsInfo,
	state *ATSElement,
	ahFlag *[2][maxGroupedSFB]uint8,
) {
	*leftPsy = PsyOutChannel{
		SfbCnt:             4,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     3,
		LastWindowSequence: LongWindow,
	}
	*rightPsy = *leftPsy
	*leftQC = QCOutChannel{}
	*rightQC = QCOutChannel{}
	*peData = PEData{}
	*tools = ToolsInfo{}
	*state = ATSElement{AHParam: AHParam{StartSfbL: 1, StartSfbS: 1}}
	*ahFlag = [2][maxGroupedSFB]uint8{}
	for i := 0; i <= 4; i++ {
		leftPsy.SfbOffsets[i] = i
		rightPsy.SfbOffsets[i] = i
	}

	leftThreshold := [...]FixpDBL{-520000000, -510000000, -500000000, -490000000}
	rightThreshold := [...]FixpDBL{-530000000, -520000000, -510000000, -500000000}
	leftWeighted := [...]FixpDBL{-260000000, -240000000, -200000000, -220000000}
	rightWeighted := [...]FixpDBL{-300000000, -350000000, -400000000, -450000000}
	minSnr := [...]FixpDBL{-90000000, -90000000, -90000000, -90000000}
	leftPE := [...]FixpDBL{
		10 << peConstPartShift,
		20 << peConstPartShift,
		30 << peConstPartShift,
		40 << peConstPartShift,
	}
	rightPE := [...]FixpDBL{
		15 << peConstPartShift,
		25 << peConstPartShift,
		60 << peConstPartShift,
		35 << peConstPartShift,
	}
	flags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive}
	copy(leftQC.SfbThresholdLdData[:], leftThreshold[:])
	copy(rightQC.SfbThresholdLdData[:], rightThreshold[:])
	copy(leftQC.SfbWeightedEnergyLdData[:], leftWeighted[:])
	copy(rightQC.SfbWeightedEnergyLdData[:], rightWeighted[:])
	copy(leftQC.SfbMinSnrLdData[:], minSnr[:])
	copy(rightQC.SfbMinSnrLdData[:], minSnr[:])
	copy(peData.PEChannelData[0].SfbPe[:], leftPE[:])
	copy(peData.PEChannelData[1].SfbPe[:], rightPE[:])
	copy(tools.MsMask[:], []int{1, 1, 1, 0})
	copy(ahFlag[0][:], flags[:])
	copy(ahFlag[1][:], flags[:])
}

func fillAllowMoreHolesEnergyCase(
	psy *PsyOutChannel,
	qc *QCOutChannel,
	peData *PEData,
	tools *ToolsInfo,
	state *ATSElement,
	ahFlag *[2][maxGroupedSFB]uint8,
) {
	*psy = PsyOutChannel{
		SfbCnt:             8,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     4,
		LastWindowSequence: LongWindow,
	}
	*qc = QCOutChannel{}
	*peData = PEData{}
	*tools = ToolsInfo{}
	*state = ATSElement{AHParam: AHParam{StartSfbL: 1, StartSfbS: 1}}
	*ahFlag = [2][maxGroupedSFB]uint8{}
	for i := 0; i <= 8; i++ {
		psy.SfbOffsets[i] = i
	}

	threshold := [...]FixpDBL{-520000000, -500000000, -510000000, -530000000, -500000000, -505000000, -490000000, -540000000}
	weighted := [...]FixpDBL{-300000000, -250000000, -280000000, -330000000, -260000000, -240000000, -200000000, -330000000}
	energyLd := [...]FixpDBL{-100000000, -160000000, -120000000, -800000000, -180000000, -140000000, -130000000, -170000000}
	energy := [...]FixpDBL{120000000, 110000000, 100000000, 90000000, 80000000, 70000000, 60000000, 50000000}
	pe := [...]FixpDBL{
		11 << peConstPartShift,
		22 << peConstPartShift,
		80 << peConstPartShift,
		40 << peConstPartShift,
		15 << peConstPartShift,
		25 << peConstPartShift,
		35 << peConstPartShift,
		70 << peConstPartShift,
	}
	flags := [...]uint8{AvoidHoleInactive, AvoidHoleInactive, AvoidHoleNone, AvoidHoleInactive, AvoidHoleInactive, AvoidHoleActive, AvoidHoleInactive, AvoidHoleInactive}
	copy(qc.SfbThresholdLdData[:], threshold[:])
	copy(qc.SfbWeightedEnergyLdData[:], weighted[:])
	copy(qc.SfbEnergyLdData[:], energyLd[:])
	copy(qc.SfbEnergy[:], energy[:])
	copy(peData.PEChannelData[0].SfbPe[:], pe[:])
	copy(ahFlag[0][:], flags[:])
}
