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
