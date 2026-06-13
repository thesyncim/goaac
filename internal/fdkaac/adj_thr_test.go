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
