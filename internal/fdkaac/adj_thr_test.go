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
			maxBitresBits: 1200,
			element:       buildBitresElementWithHistory(0, -1),
			want:          [8]int{0, 0, 180, 620, 0, -1, int(lowBitresCorrectionMin), 1},
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

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil no-AH PE data", func() { FDKaacEncCalcPENoAH(nil, &peFlags, []*PsyOutChannel{&pePsy}, 1) }},
		{"nil no-AH flags", func() { FDKaacEncCalcPENoAH(&peData, nil, []*PsyOutChannel{&pePsy}, 1) }},
		{"bad no-AH channel count", func() { FDKaacEncCalcPENoAH(&peData, &peFlags, []*PsyOutChannel{&pePsy}, 0) }},
		{"nil no-AH psy", func() { FDKaacEncCalcPENoAH(&peData, &peFlags, []*PsyOutChannel{nil}, 1) }},
		{"nil reduction flags", func() { FDKaacEncReduceThresholdsCBR(qc[:], psy[:], nil, &thrExp, 1, 0x20000000, 0) }},
		{"nil threshold exponent", func() { FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, nil, 1, 0x20000000, 0) }},
		{"nil reduction qc", func() {
			bad := [1]*QCOutChannel{nil}
			FDKaacEncReduceThresholdsCBR(bad[:], psy[:], &ahFlag, &thrExp, 1, 0x20000000, 0)
		}},
		{"negative reduction value", func() { FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, &thrExp, 1, -1, 0) }},
		{"bad reduction exponent", func() { FDKaacEncReduceThresholdsCBR(qc[:], psy[:], &ahFlag, &thrExp, 1, 0x20000000, DfractBits+1) }},
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
		{"negative grant", func() {
			FDKaacEncDistributeBits(&state, &element, psy[:], &peData, 1, 0, -1, 500, 1200, MaxValDBL, BitresModeFull)
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
		adjThrSink = qcStorage.SfbThresholdLdData[0] + FixpDBL(pe+constPart+nActiveLines)
		adjThrHashSink = uint64(ahFlag[0][3])
	})
	if allocs != 0 {
		t.Fatalf("threshold reduction allocations = %v, want 0", allocs)
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
