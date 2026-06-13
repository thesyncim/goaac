package fdkaac

import "testing"

var qcMainPrepareSink int
var qcMainPrepareHashSink uint64

func TestFDKaacEncCalcMaxValueInSfbVector(t *testing.T) {
	quant := [...]int16{0, -2, 5, -7, 4, 1, 12, -3}
	offsets := [...]int{0, 3, 6, 8, 8}
	maxValue := [...]uint32{99, 99, 99, 99}

	got := FDKaacEncCalcMaxValueInSfb(4, 2, 4, offsets[:], quant[:], maxValue[:])
	if got != 7 {
		t.Fatalf("max value = %d, want 7", got)
	}
	want := [...]uint32{5, 7, 99, 99}
	if maxValue != want {
		t.Fatalf("max-value bands = %v, want %v", maxValue, want)
	}
}

func TestFDKaacEncReduceBitConsumptionVectors(t *testing.T) {
	var left QCOutChannel
	var right QCOutChannel
	element := QCOutElement{
		DynBitsUsed: 90,
		QCOutChannel: [2]*QCOutChannel{
			&left,
			&right,
		},
	}
	left.GlobalGain = 50
	right.GlobalGain = 70
	iterations := 0
	chConstraints := [...]int{0, 1}
	calculateQuant := [...]int{0, 0}
	elBits := ElementBits{MaxBitsEl: 1024, BitResLevelEl: 128}

	result, errCode := FDKaacEncReduceBitConsumption(
		&iterations, 4, 1, chConstraints[:], calculateQuant[:], 2, nil, nil, &element, &elBits, aotAACLC, 0, -1,
	)
	if errCode != AACEncOK {
		t.Fatalf("reduce-bit normal error = %#x, want OK", errCode)
	}
	if iterations != 1 || result.IterationsReached != 1 || result.AdjustedChannels != 1 {
		t.Fatalf("reduce-bit iteration result = %+v iterations=%d", result, iterations)
	}
	if left.GlobalGain != 51 || right.GlobalGain != 70 {
		t.Fatalf("global gains = %d,%d want 51,70", left.GlobalGain, right.GlobalGain)
	}
	if calculateQuant != [...]int{1, 0} {
		t.Fatalf("calculate quant = %v, want [1 0]", calculateQuant)
	}

	element.GrantedDynBits = 256
	element.StaticBitsUsed = 9
	left.GlobalGain = 51
	right.GlobalGain = 70
	iterations = 4
	chConstraints = [...]int{1, 1}
	calculateQuant = [...]int{0, 0}
	elBits = ElementBits{MaxBitsEl: 2048, BitResLevelEl: 128}
	result, errCode = FDKaacEncReduceBitConsumption(
		&iterations, 4, -1, chConstraints[:], calculateQuant[:], 2, nil, nil, &element, &elBits, aotAACLC, 0, -1,
	)
	if errCode != AACEncOK {
		t.Fatalf("reduce-bit max-iteration non-crash error = %#x, want OK", errCode)
	}
	if iterations != 5 || result.MaxIterationsHit != 1 || result.AdjustedChannels != 2 {
		t.Fatalf("max-iteration result = %+v iterations=%d", result, iterations)
	}
	if left.GlobalGain != 52 || right.GlobalGain != 71 {
		t.Fatalf("max-iteration global gains = %d,%d want 52,71", left.GlobalGain, right.GlobalGain)
	}
	if calculateQuant != [...]int{1, 1} {
		t.Fatalf("max-iteration calculate quant = %v, want [1 1]", calculateQuant)
	}
}

func TestFDKaacEncReduceBitConsumptionCrashRecoveryBoundary(t *testing.T) {
	var qc QCOutChannel
	var psy PsyOutChannel
	var psyElement PsyOutElement
	var qcOut QCOut
	fillQCCrashRecoveryCase(&psy, &qc, &psyElement, &qcOut)
	element := QCOutElement{
		StaticBitsUsed: qcOut.StaticBits,
		DynBitsUsed:    300,
		GrantedDynBits: 10,
		QCOutChannel:   [2]*QCOutChannel{&qc},
	}
	elBits := ElementBits{MaxBitsEl: 128, BitResLevelEl: 0}
	iterations := 3
	chConstraints := [...]int{1}
	calculateQuant := [...]int{0}

	result, errCode := FDKaacEncReduceBitConsumption(
		&iterations, 3, 1, chConstraints[:], calculateQuant[:], 1, &psyElement, &qcOut, &element, &elBits, aotAACLC, 0, -1,
	)
	if errCode != AACEncOK {
		t.Fatalf("crash recovery boundary error = %#x, want OK", errCode)
	}
	if result.CrashRecoveryNeeded != 1 || result.BitsToSave != 298 || result.SavedBits != 19 || result.StopSfb != 0 {
		t.Fatalf("crash recovery result = %+v, want needed with 298 bits and 19 static bits saved", result)
	}
	if iterations != 4 || qc.GlobalGain != 0 || calculateQuant[0] != 1 {
		t.Fatalf("crash recovery retry state iterations=%d gain=%d calc=%d", iterations, qc.GlobalGain, calculateQuant[0])
	}
	if qc.SectionData.MaxSfbPerGroup != 0 || psy.MaxSfbPerGroup != 0 || qc.SectionData.Huffsection[0].SfbCnt != 0 {
		t.Fatalf("crash recovery band state max=%d psy=%d section=%d", qc.SectionData.MaxSfbPerGroup, psy.MaxSfbPerGroup, qc.SectionData.Huffsection[0].SfbCnt)
	}
	if element.StaticBitsUsed != 29 || element.GrantedDynBits != 29 || qcOut.StaticBits != 29 || qcOut.GrantedDynBits != 29 || qcOut.MaxDynBits != 119 {
		t.Fatalf("crash recovery bit budgets element static/dyn=%d/%d frame static/dyn/max=%d/%d/%d", element.StaticBitsUsed, element.GrantedDynBits, qcOut.StaticBits, qcOut.GrantedDynBits, qcOut.MaxDynBits)
	}
	if psy.TNSInfo != (TNSInfo{}) || psyElement.ToolsInfo != (ToolsInfo{}) {
		t.Fatalf("crash recovery did not clear zero-spectrum tools")
	}
}

func TestFDKaacEncCrashRecoveryVector(t *testing.T) {
	var qc QCOutChannel
	var psy PsyOutChannel
	var psyElement PsyOutElement
	var qcOut QCOut
	fillQCCrashRecoveryCase(&psy, &qc, &psyElement, &qcOut)
	qcElement := QCOutElement{
		StaticBitsUsed: qcOut.StaticBits,
		GrantedDynBits: 10,
		QCOutChannel:   [2]*QCOutChannel{&qc},
	}

	result, errCode := FDKaacEncCrashRecovery(1, &psyElement, &qcOut, &qcElement, 10000, aotAACLC, 0, -1)
	if errCode != AACEncOK {
		t.Fatalf("crash recovery error = %#x, want OK", errCode)
	}
	if result != (QCCrashRecoveryResult{BitsToSave: 10000, SavedBits: 19, StopSfb: 0, StaticBitsNew: 29}) {
		t.Fatalf("crash recovery result = %+v", result)
	}
	if got := [...]int{
		qc.SectionData.MaxSfbPerGroup,
		psy.MaxSfbPerGroup,
		qc.SectionData.Huffsection[0].SfbCnt,
		qcElement.StaticBitsUsed,
		qcElement.GrantedDynBits,
		qcOut.StaticBits,
		qcOut.GrantedDynBits,
		qcOut.MaxDynBits,
	}; got != [...]int{0, 0, 0, 29, 29, 29, 29, 119} {
		t.Fatalf("crash recovery state = %v", got)
	}
	if psy.TNSInfo != (TNSInfo{}) || psyElement.ToolsInfo != (ToolsInfo{}) {
		t.Fatalf("crash recovery did not clear tools")
	}
}

func TestFDKaacEncQCAccountingVectors(t *testing.T) {
	cm := ChannelMapping{NElements: 3}
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	cm.ElInfo[1] = ElementInfo{ElType: idDSE}
	cm.ElInfo[2] = ElementInfo{ElType: idCPE, NChannelsInEl: 2}
	qcOut0 := QCOut{GlobalExtBits: 5}
	el0 := QCOutElement{DynBitsUsed: 101, StaticBitsUsed: 29, ExtBitsUsed: 6}
	el2 := QCOutElement{DynBitsUsed: 201, StaticBitsUsed: 44, ExtBitsUsed: 8}
	qcOutFrames := [1]*QCOut{&qcOut0}
	var qcElements [1][maxChannelElements]*QCOutElement
	qcElements[0][0] = &el0
	qcElements[0][2] = &el2

	if got := FDKaacEncGetTotalConsumedBits(qcOutFrames[:], qcElements[:], &cm, 56, 1); got != 456 {
		t.Fatalf("total consumed bits = %d, want 456", got)
	}

	vbrKernel := QCKernel{BitrateMode: QCBitrateModeVBR3, MinBitsPerFrame: 900}
	vbrOut := QCOut{GrantedDynBits: 1000, UsedDynBits: 813, StaticBits: 50, ElementExtBits: 3, GlobalExtBits: 5}
	vbrFrames := [1]*QCOut{&vbrOut}
	if errCode := FDKaacEncUpdateFillBits(&vbrKernel, vbrFrames[:]); errCode != AACEncOK {
		t.Fatalf("VBR fill error = %#x, want OK", errCode)
	}
	if vbrOut.TotFillBits != 35 || vbrOut.TotalBits != 874 {
		t.Fatalf("VBR fill = %d total = %d, want 35 and source-shaped pre-extra total 874", vbrOut.TotFillBits, vbrOut.TotalBits)
	}

	cbrKernel := QCKernel{BitrateMode: QCBitrateModeCBR, BitResTotMax: 400, BitResTot: 360}
	cbrOut := QCOut{GrantedDynBits: 1000, UsedDynBits: 920, StaticBits: 60, ElementExtBits: 4, GlobalExtBits: 2}
	cbrFrames := [1]*QCOut{&cbrOut}
	if errCode := FDKaacEncUpdateFillBits(&cbrKernel, cbrFrames[:]); errCode != AACEncOK {
		t.Fatalf("CBR fill error = %#x, want OK", errCode)
	}
	if cbrOut.TotFillBits != 48 || cbrOut.TotalBits != 1034 {
		t.Fatalf("CBR fill = %d total = %d, want 48 and 1034", cbrOut.TotFillBits, cbrOut.TotalBits)
	}
}

func TestFDKaacEncFinalizeAndBitresVectors(t *testing.T) {
	kernel := QCKernel{
		GlobHdrBits:     64,
		MaxBitsPerFrame: 1024,
		BitrateMode:     QCBitrateModeCBR,
		BitResTot:       100,
		BitResTotMax:    1000,
	}
	qcOut := QCOut{
		GrantedDynBits: 500,
		StaticBits:     29,
		UsedDynBits:    211,
		TotFillBits:    31,
		ElementExtBits: 4,
		GlobalExtBits:  6,
	}
	if errCode := FDKaacEncFinalizeBitConsumption(&kernel, &qcOut, 56, aotAACLC, 0, -1); errCode != AACEncOK {
		t.Fatalf("finalize error = %#x, want OK", errCode)
	}
	if kernel.GlobHdrBits != 56 || kernel.BitResTot != 108 {
		t.Fatalf("finalize kernel = %+v, want header 56 bitres 108", kernel)
	}
	if qcOut.TotFillBits != 27 || qcOut.AlignBits != 3 || qcOut.TotalBits != 280 || qcOut.GrantedDynBits != 500 {
		t.Fatalf("finalized frame = %+v, want fill 27 align 3 total 280 granted 500", qcOut)
	}

	frames := [1]*QCOut{&qcOut}
	FDKaacEncUpdateBitres(&kernel, frames[:])
	if kernel.BitResTot != 367 {
		t.Fatalf("CBR bit reservoir = %d, want 367", kernel.BitResTot)
	}

	vbrKernel := QCKernel{BitrateMode: QCBitrateModeVBR5, MaxBitsPerFrame: 700, BitResTotMax: 650}
	vbrOut := QCOut{UsedDynBits: 200}
	vbrFrames := [1]*QCOut{&vbrOut}
	FDKaacEncUpdateBitres(&vbrKernel, vbrFrames[:])
	if vbrKernel.BitResTot != 650 {
		t.Fatalf("VBR bit reservoir = %d, want 650", vbrKernel.BitResTot)
	}
}

func TestFDKaacEncElementBitDistributionVectors(t *testing.T) {
	cm := ChannelMapping{NElements: 3}
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	cm.ElInfo[1] = ElementInfo{ElType: idCPE, NChannelsInEl: 2}
	cm.ElInfo[2] = ElementInfo{ElType: idDSE}
	el0 := QCOutElement{}
	el1 := QCOutElement{}
	qcElements := [maxChannelElements]*QCOutElement{&el0, &el1}
	bits0 := ElementBits{RelativeBitsEl: 0x40000000}
	bits1 := ElementBits{RelativeBitsEl: 0}
	elementBits := [maxChannelElements]*ElementBits{&bits0, &bits1}

	if errCode := FDKaacEncDistributeElementDynBits(qcElements[:], &cm, elementBits[:], 1000); errCode != AACEncOK {
		t.Fatalf("element bit distribution error = %#x, want OK", errCode)
	}
	if el0.GrantedDynBits != 500 || el1.GrantedDynBits != 500 {
		t.Fatalf("distributed bits = %d,%d want 500,500", el0.GrantedDynBits, el1.GrantedDynBits)
	}
}

func TestFDKaacEncBitResRedistributionVectors(t *testing.T) {
	cm := ChannelMapping{NElements: 2}
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	cm.ElInfo[1] = ElementInfo{ElType: idCPE, NChannelsInEl: 2}
	kernel := QCKernel{MaxBitsPerFrame: 2000, BitResTot: 800, BitResTotMax: 1200}
	bits0 := ElementBits{RelativeBitsEl: 0x20000000}
	bits1 := ElementBits{RelativeBitsEl: 0x60000000}
	elementBits := [maxChannelElements]*ElementBits{&bits0, &bits1}

	if errCode := FDKaacEncBitResRedistribution(&kernel, &cm, elementBits[:], 1000); errCode != AACEncOK {
		t.Fatalf("bit reservoir redistribution error = %#x, want OK", errCode)
	}
	if got := [...]int{bits0.BitResLevelEl, bits1.BitResLevelEl, bits0.MaxBitResBitsEl, bits1.MaxBitResBitsEl}; got != [...]int{200, 600, 250, 750} {
		t.Fatalf("redistributed reservoir = %v, want [200 600 250 750]", got)
	}

	low := QCKernel{MaxBitsPerFrame: 2000, BitResTot: -1, BitResTotMax: 1200}
	if errCode := FDKaacEncBitResRedistribution(&low, &cm, elementBits[:], 1000); errCode != AACEncBitresTooLow {
		t.Fatalf("low reservoir error = %#x, want %#x", errCode, AACEncBitresTooLow)
	}
	high := QCKernel{MaxBitsPerFrame: 2000, BitResTot: 1300, BitResTotMax: 1200}
	if errCode := FDKaacEncBitResRedistribution(&high, &cm, elementBits[:], 1000); errCode != AACEncBitresTooHigh {
		t.Fatalf("high reservoir error = %#x, want %#x", errCode, AACEncBitresTooHigh)
	}
}

func TestFDKaacEncPrepareBitDistributionVector(t *testing.T) {
	var state AdjThrState
	FDKaacEncInitBitresState(&state)
	elementState := buildBitresElementWithHistory(420, 650)
	state.AdjThrStateElem[0] = &elementState

	psy := PsyOutChannel{LastWindowSequence: LongWindow}
	psyElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&psy}}
	qcElement := QCOutElement{PEData: PEData{Pe: 430}}
	qcOut := QCOut{}
	cm := ChannelMapping{NElements: 1}
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	elBits := ElementBits{RelativeBitsEl: 0x40000000, BitResLevelEl: 500, MaxBitResBitsEl: 1200}
	kernel := QCKernel{MaxBitsPerFrame: 2000, MaxBitFac: MaxValDBL, BitResMode: BitresModeFull}
	psyElements := [1]*PsyOutElement{&psyElement}
	qcOutFrames := [1]*QCOut{&qcOut}
	var qcElements [1][maxChannelElements]*QCOutElement
	qcElements[0][0] = &qcElement
	elementBits := [maxChannelElements]*ElementBits{&elBits}

	result, errCode := FDKaacEncPrepareBitDistribution(&kernel, &state, psyElements[:], qcOutFrames[:], qcElements[:], &cm, elementBits[:], 720)
	if errCode != AACEncOK {
		t.Fatalf("prepare bit distribution error = %#x, want OK", errCode)
	}
	if got := [...]int{
		result.TotalAvailableBits,
		result.AvgTotalDynBits,
		result.DistributedBits,
		result.DistributedElements,
		result.TotalGrantedPeCorr,
		qcOut.GrantedDynBits,
		qcOut.MaxDynBits,
		qcElement.GrantedDynBits,
		qcElement.GrantedPe,
		qcElement.GrantedPeCorr,
		elementState.PeMin,
		elementState.PeMax,
	}; got != [...]int{1220, 0, 720, 1, 798, 720, 2000, 720, 798, 798, 255, 607} {
		t.Fatalf("prepare bit distribution vector = %v", got)
	}
}

func TestFDKaacEncGetMinimalStaticBitDemandVector(t *testing.T) {
	var psy PsyOutChannel
	var qc QCOutChannel
	var psyElement PsyOutElement
	var qcOut QCOut
	fillQCCrashRecoveryCase(&psy, &qc, &psyElement, &qcOut)

	cm := ChannelMapping{NElements: 2}
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	cm.ElInfo[1] = ElementInfo{ElType: idDSE}
	psyElements := [2]*PsyOutElement{&psyElement, nil}

	bits, errCode := FDKaacEncGetMinimalStaticBitDemand(&cm, psyElements[:])
	if errCode != AACEncOK {
		t.Fatalf("minimal static demand error = %#x, want OK", errCode)
	}
	if bits != 29 {
		t.Fatalf("minimal static demand = %d, want 29", bits)
	}
}

func TestFDKaacEncPrepareBitDistributionLowReservoirMinimalStaticVector(t *testing.T) {
	var state AdjThrState
	FDKaacEncInitBitresState(&state)
	elementState := buildBitresElementWithHistory(0, -1)
	state.AdjThrStateElem[0] = &elementState

	psy := PsyOutChannel{LastWindowSequence: LongWindow, WindowShape: WindowShapeKBD}
	psyElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&psy}}
	qcElement := QCOutElement{PEData: PEData{Pe: 10}}
	qcOut := QCOut{StaticBits: 96}
	cm := ChannelMapping{NElements: 1}
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	elBits := ElementBits{RelativeBitsEl: 0x40000000, BitResLevelEl: 0, MaxBitResBitsEl: 1200}
	kernel := QCKernel{MaxBitsPerFrame: 2000, MaxBitFac: MaxValDBL, BitResMode: BitresModeFull}
	psyElements := [1]*PsyOutElement{&psyElement}
	qcOutFrames := [1]*QCOut{&qcOut}
	var qcElements [1][maxChannelElements]*QCOutElement
	qcElements[0][0] = &qcElement
	elementBits := [maxChannelElements]*ElementBits{&elBits}

	result, errCode := FDKaacEncPrepareBitDistribution(&kernel, &state, psyElements[:], qcOutFrames[:], qcElements[:], &cm, elementBits[:], 40)
	if errCode != AACEncOK {
		t.Fatalf("low-reservoir prepare error = %#x, want OK", errCode)
	}
	if got := [...]int{
		result.TotalAvailableBits,
		result.AvgTotalDynBits,
		result.DistributedBits,
		result.DistributedElements,
		result.TotalGrantedPeCorr,
		qcOut.GrantedDynBits,
		qcOut.MaxDynBits,
		qcElement.GrantedDynBits,
		qcElement.GrantedPe,
		qcElement.GrantedPeCorr,
	}; got != [...]int{40, 0, -56, 1, 0, -56, 1904, -56, 0, 0} {
		t.Fatalf("low-reservoir prepare vector = %v", got)
	}
}

func TestFDKaacEncQCMainQuantizationDecisionVectors(t *testing.T) {
	for _, tt := range []struct {
		name                   string
		usedDynBits            int
		maxDynBits             int
		totalAvailableBits     int
		sumDynBitsConsumed     int
		decreaseBitConsumption int
		iterations             int
		want                   QCMainQuantizationDecisionResult
	}{
		{
			name:                   "saving direction exits under budget",
			usedDynBits:            100,
			maxDynBits:             1000,
			totalAvailableBits:     512,
			sumDynBitsConsumed:     100,
			decreaseBitConsumption: 1,
			iterations:             0,
			want: QCMainQuantizationDecisionResult{
				QuantizationDone:       1,
				SumDynBitsConsumed:     100,
				SumBitsConsumed:        192,
				DecreaseBitConsumption: 1,
			},
		},
		{
			name:                   "dynamic overshoot forces saving",
			usedDynBits:            700,
			maxDynBits:             600,
			totalAvailableBits:     1200,
			sumDynBitsConsumed:     700,
			decreaseBitConsumption: -1,
			iterations:             0,
			want: QCMainQuantizationDecisionResult{
				SumDynBitsConsumed:     700,
				SumBitsConsumed:        792,
				DecreaseBitConsumption: 1,
				DynBitsOvershoot:       1,
				ConstraintsReset:       1,
			},
		},
		{
			name:                   "over budget keeps saving",
			usedDynBits:            700,
			maxDynBits:             1000,
			totalAvailableBits:     512,
			sumDynBitsConsumed:     700,
			decreaseBitConsumption: -1,
			iterations:             0,
			want: QCMainQuantizationDecisionResult{
				SumDynBitsConsumed:     700,
				SumBitsConsumed:        792,
				DecreaseBitConsumption: 1,
				ConstraintsReset:       1,
			},
		},
		{
			name:                   "emergency iteration under budget spends",
			usedDynBits:            100,
			maxDynBits:             1000,
			totalAvailableBits:     512,
			sumDynBitsConsumed:     100,
			decreaseBitConsumption: -1,
			iterations:             4,
			want: QCMainQuantizationDecisionResult{
				QuantizationDone:       1,
				SumDynBitsConsumed:     100,
				SumBitsConsumed:        192,
				DecreaseBitConsumption: 0,
				EmergencyIterations:    1,
				ConstraintsReset:       1,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			kernel := QCKernel{GlobHdrBits: 56}
			cm := ChannelMapping{NElements: 1}
			cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
			qcOut := QCOut{UsedDynBits: tt.usedDynBits, MaxDynBits: tt.maxDynBits}
			qcElement := QCOutElement{DynBitsUsed: tt.usedDynBits, StaticBitsUsed: 29}
			qcOutFrames := [1]*QCOut{&qcOut}
			var qcElements [1][maxChannelElements]*QCOutElement
			qcElements[0][0] = &qcElement
			iterations := [maxChannelElements]int{tt.iterations}

			got := FDKaacEncQCMainQuantizationDecision(
				&kernel,
				qcOutFrames[:],
				qcElements[:],
				&cm,
				tt.totalAvailableBits,
				0,
				tt.sumDynBitsConsumed,
				tt.decreaseBitConsumption,
				iterations[:],
				4,
			)
			if got != tt.want {
				t.Fatalf("QC decision = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFDKaacEncQCMainFrameOnePassVector(t *testing.T) {
	var cm ChannelMapping
	var adj AdjThrState
	var state ATSElement
	var psy PsyOutChannel
	var qc QCOutChannel
	var psyElement PsyOutElement
	var qcElement QCOutElement
	var elementBits ElementBits
	var qcOut QCOut
	var kernel QCKernel
	fillQCMainFrameCase(&cm, &adj, &state, &psy, &qc, &psyElement, &qcElement, &elementBits, &qcOut, &kernel)
	psyElements := [1]*PsyOutElement{&psyElement}
	qcOutFrames := [1]*QCOut{&qcOut}
	var qcElements [1][maxChannelElements]*QCOutElement
	qcElements[0][0] = &qcElement
	elementBitsSlice := [maxChannelElements]*ElementBits{&elementBits}
	var scratch QCMainFrameScratch

	result, errCode := FDKaacEncQCMainFrame(
		&kernel,
		&adj,
		psyElements[:],
		qcOutFrames[:],
		qcElements[:],
		&cm,
		elementBitsSlice[:],
		&scratch,
		720,
		2,
		0,
		4,
		aotAACLC,
		0,
		-1,
	)
	if errCode != AACEncOK {
		t.Fatalf("QC main frame error = %#x, want OK", errCode)
	}
	if got := [...]int{
		result.IsCBRAdjustment,
		result.BitresAvgTotalBits,
		result.TotalAvailableBits,
		result.QuantizedElements,
		result.QuantizationPasses,
		result.QuantizationDone,
		result.DecreaseBitConsumption,
		qcOut.GrantedDynBits,
		qcOut.UsedDynBits,
		qcElement.GrantedDynBits,
	}; got != [...]int{1, 720, 720, 1, 1, 1, 0, 691, 22, 691} {
		t.Fatalf("QC main frame vector = %v", got)
	}
}

func TestFDKaacEncQCMainFrameRejectsInvalid(t *testing.T) {
	var cm ChannelMapping
	var adj AdjThrState
	var state ATSElement
	var psy PsyOutChannel
	var qc QCOutChannel
	var psyElement PsyOutElement
	var qcElement QCOutElement
	var elementBits ElementBits
	var qcOut QCOut
	var kernel QCKernel
	fillQCMainFrameCase(&cm, &adj, &state, &psy, &qc, &psyElement, &qcElement, &elementBits, &qcOut, &kernel)
	psyElements := [1]*PsyOutElement{&psyElement}
	qcOutFrames := [1]*QCOut{&qcOut}
	var qcElements [1][maxChannelElements]*QCOutElement
	qcElements[0][0] = &qcElement
	elementBitsSlice := [maxChannelElements]*ElementBits{&elementBits}
	var scratch QCMainFrameScratch

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil kernel", func() {
			FDKaacEncQCMainFrame(nil, &adj, psyElements[:], qcOutFrames[:], qcElements[:], &cm, elementBitsSlice[:], &scratch, 720, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"nil threshold state", func() {
			FDKaacEncQCMainFrame(&kernel, nil, psyElements[:], qcOutFrames[:], qcElements[:], &cm, elementBitsSlice[:], &scratch, 720, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"nil frame", func() {
			badFrames := qcOutFrames
			badFrames[0] = nil
			FDKaacEncQCMainFrame(&kernel, &adj, psyElements[:], badFrames[:], qcElements[:], &cm, elementBitsSlice[:], &scratch, 720, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"nil element", func() {
			badElements := qcElements
			badElements[0][0] = nil
			FDKaacEncQCMainFrame(&kernel, &adj, psyElements[:], qcOutFrames[:], badElements[:], &cm, elementBitsSlice[:], &scratch, 720, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"nil scratch", func() {
			FDKaacEncQCMainFrame(&kernel, &adj, psyElements[:], qcOutFrames[:], qcElements[:], &cm, elementBitsSlice[:], nil, 720, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"negative average bits", func() {
			FDKaacEncQCMainFrame(&kernel, &adj, psyElements[:], qcOutFrames[:], qcElements[:], &cm, elementBitsSlice[:], &scratch, -1, 2, 0, 4, aotAACLC, 0, -1)
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

func TestFDKaacEncQCMainPrepareSCEVector(t *testing.T) {
	var directElement ElementInfo
	var directState ATSElement
	var directPsy PsyOutChannel
	var directQC QCOutChannel
	var directTools ToolsInfo
	var directPsyElement PsyOutElement
	var directQCElement QCOutElement
	var directMDCT [maxSpectralLines]FixpDBL
	fillQCMainPrepareCase(
		&directElement, &directState, &directPsy, &directQC, &directTools,
		&directPsyElement, &directQCElement, &directMDCT,
	)

	directPsyChannels := directPsyElement.PsyOutChannel[:directElement.NChannelsInEl]
	directQCChannels := directQCElement.QCOutChannel[:directElement.NChannelsInEl]
	FDKaacEncCalcFormFactor(directQCChannels, directPsyChannels, directElement.NChannelsInEl)
	FDKaacEncPECalculation(
		&directQCElement.PEData, directPsyChannels, directQCChannels,
		&directPsyElement.ToolsInfo, &directState, directElement.NChannelsInEl,
	)
	directBits, directErr := FDKaacEncChannelElementWrite(
		nil, &directElement, nil, &directPsyElement, directPsyChannels,
		0, aotAACLC, -1, 0,
	)
	directQCElement.StaticBitsUsed = directBits
	if directErr != AACEncOK {
		t.Fatalf("direct QC prepare sequence error = %#x, want OK", directErr)
	}

	var wrappedElement ElementInfo
	var wrappedState ATSElement
	var wrappedPsy PsyOutChannel
	var wrappedQC QCOutChannel
	var wrappedTools ToolsInfo
	var wrappedPsyElement PsyOutElement
	var wrappedQCElement QCOutElement
	var wrappedMDCT [maxSpectralLines]FixpDBL
	fillQCMainPrepareCase(
		&wrappedElement, &wrappedState, &wrappedPsy, &wrappedQC, &wrappedTools,
		&wrappedPsyElement, &wrappedQCElement, &wrappedMDCT,
	)

	gotErr := FDKaacEncQCMainPrepare(&wrappedElement, &wrappedState, &wrappedPsyElement, &wrappedQCElement, aotAACLC, 0, -1)
	if gotErr != AACEncOK {
		t.Fatalf("QC prepare error = %#x, want OK", gotErr)
	}
	if wrappedQCElement.StaticBitsUsed != directQCElement.StaticBitsUsed || wrappedQCElement.StaticBitsUsed != 29 {
		t.Fatalf("static bits = %d, want direct %d and vector 29", wrappedQCElement.StaticBitsUsed, directQCElement.StaticBitsUsed)
	}
	if wrappedQCElement.PEData.Pe != directQCElement.PEData.Pe ||
		wrappedQCElement.PEData.ConstPart != directQCElement.PEData.ConstPart ||
		wrappedQCElement.PEData.NActiveLines != directQCElement.PEData.NActiveLines {
		t.Fatalf("PE data = (%d,%d,%d), want (%d,%d,%d)",
			wrappedQCElement.PEData.Pe, wrappedQCElement.PEData.ConstPart, wrappedQCElement.PEData.NActiveLines,
			directQCElement.PEData.Pe, directQCElement.PEData.ConstPart, directQCElement.PEData.NActiveLines,
		)
	}
	if hashFixpDBL(wrappedQC.SfbFormFactorLdData[:8]) != hashFixpDBL(directQC.SfbFormFactorLdData[:8]) ||
		hashFixpDBL(wrappedQC.SfbThresholdLdData[:8]) != hashFixpDBL(directQC.SfbThresholdLdData[:8]) ||
		hashFixpDBL(wrappedQC.SfbWeightedEnergyLdData[:8]) != hashFixpDBL(directQC.SfbWeightedEnergyLdData[:8]) ||
		hashFixpDBL(wrappedQC.SfbEnFacLd[:8]) != hashFixpDBL(directQC.SfbEnFacLd[:8]) {
		t.Fatalf("QC prepare channel hashes differ from direct sequence")
	}
}

func TestFDKaacEncQCMainQuantizeFrameSCEVector(t *testing.T) {
	var directCM ChannelMapping
	var directAdj AdjThrState
	var directState ATSElement
	var directPsy PsyOutChannel
	var directQC QCOutChannel
	var directPsyElement PsyOutElement
	var directQCElement QCOutElement
	var directQCOut QCOut
	var directElementBits ElementBits
	var directScratch QCMainQuantizeScratch
	fillQCMainQuantizeCase(&directCM, &directAdj, &directState, &directPsy, &directQC, &directPsyElement, &directQCElement, &directElementBits, &directQCOut)

	directPsyElements := [1]*PsyOutElement{&directPsyElement}
	directQCElements := [1]*QCOutElement{&directQCElement}
	directElementBitsSlice := [1]*ElementBits{&directElementBits}
	directResult, directErr := runDirectQCMainQuantizeFrame(
		&directCM, directPsyElements[:], &directQCOut, directQCElements[:], &directAdj, directElementBitsSlice[:], &directScratch, 2, 0, 4, 0,
	)
	if directErr != AACEncOK {
		t.Fatalf("direct QC quantize error = %#x, want OK", directErr)
	}

	var wrappedCM ChannelMapping
	var wrappedAdj AdjThrState
	var wrappedState ATSElement
	var wrappedPsy PsyOutChannel
	var wrappedQC QCOutChannel
	var wrappedPsyElement PsyOutElement
	var wrappedQCElement QCOutElement
	var wrappedQCOut QCOut
	var wrappedElementBits ElementBits
	var wrappedScratch QCMainQuantizeScratch
	fillQCMainQuantizeCase(&wrappedCM, &wrappedAdj, &wrappedState, &wrappedPsy, &wrappedQC, &wrappedPsyElement, &wrappedQCElement, &wrappedElementBits, &wrappedQCOut)

	wrappedPsyElements := [1]*PsyOutElement{&wrappedPsyElement}
	wrappedQCElements := [1]*QCOutElement{&wrappedQCElement}
	wrappedElementBitsSlice := [1]*ElementBits{&wrappedElementBits}
	wrappedResult, wrappedErr := FDKaacEncQCMainQuantizeFrame(
		&wrappedCM, wrappedPsyElements[:], &wrappedQCOut, wrappedQCElements[:], &wrappedAdj, wrappedElementBitsSlice[:], &wrappedScratch, 2, 0, 4, aotAACLC, 0, -1,
	)
	if wrappedErr != AACEncOK {
		t.Fatalf("QC quantize error = %#x, want OK", wrappedErr)
	}
	if wrappedResult != directResult {
		t.Fatalf("QC quantize result = %+v, want %+v", wrappedResult, directResult)
	}
	if wrappedQCOut.UsedDynBits != directQCOut.UsedDynBits ||
		wrappedQCElement.DynBitsUsed != directQCElement.DynBitsUsed ||
		wrappedState.DynBitsLast != directState.DynBitsLast {
		t.Fatalf("dynamic bits = frame %d/%d element %d/%d history %d/%d",
			wrappedQCOut.UsedDynBits, directQCOut.UsedDynBits,
			wrappedQCElement.DynBitsUsed, directQCElement.DynBitsUsed,
			wrappedState.DynBitsLast, directState.DynBitsLast,
		)
	}
	if wrappedQC.GlobalGain != directQC.GlobalGain ||
		hashBandEnergyInts(wrappedQC.Scf[:5]) != hashBandEnergyInts(directQC.Scf[:5]) ||
		hashInt16(wrappedQC.QuantSpec[:11]) != hashInt16(directQC.QuantSpec[:11]) ||
		hashUint32(wrappedQC.MaxValueInSfb[:5]) != hashUint32(directQC.MaxValueInSfb[:5]) {
		t.Fatalf("QC quantize channel state differs from direct sequence")
	}
	if wrappedQC.SectionData.HuffmanBits != directQC.SectionData.HuffmanBits ||
		wrappedQC.SectionData.SideInfoBits != directQC.SectionData.SideInfoBits ||
		wrappedQC.SectionData.ScalefacBits != directQC.SectionData.ScalefacBits ||
		wrappedQC.SectionData.NoOfSections != directQC.SectionData.NoOfSections {
		t.Fatalf("section data differs from direct sequence")
	}
}

func TestFDKaacEncQCMainQuantizeFrameResetsRetryScratch(t *testing.T) {
	var cm ChannelMapping
	var adj AdjThrState
	var state ATSElement
	var psy PsyOutChannel
	var qc QCOutChannel
	var psyElement PsyOutElement
	var qcElement QCOutElement
	var elementBits ElementBits
	var qcOut QCOut
	var scratch QCMainQuantizeScratch
	fillQCMainQuantizeCase(&cm, &adj, &state, &psy, &qc, &psyElement, &qcElement, &elementBits, &qcOut)
	for i := 0; i < maxChannelElements; i++ {
		scratch.Iterations[i] = 99
		scratch.ConstraintsFulfilled[i] = 0
		for ch := 0; ch < 2; ch++ {
			scratch.ChConstraintsFulfilled[i][ch] = 0
			scratch.CalculateQuant[i][ch] = 0
		}
	}

	psyElements := [1]*PsyOutElement{&psyElement}
	qcElements := [1]*QCOutElement{&qcElement}
	elementBitsSlice := [1]*ElementBits{&elementBits}
	result, errCode := FDKaacEncQCMainQuantizeFrame(
		&cm, psyElements[:], &qcOut, qcElements[:], &adj, elementBitsSlice[:], &scratch, 2, 0, 4, aotAACLC, 0, -1,
	)
	if errCode != AACEncOK {
		t.Fatalf("QC quantize with stale scratch error = %#x, want OK", errCode)
	}
	if result.ReductionIterations != 0 || result.GainAdjustments != 0 || scratch.Iterations[0] != 0 {
		t.Fatalf("stale scratch leaked into retry state result=%+v iterations=%d", result, scratch.Iterations[0])
	}
	if qcOut.UsedDynBits <= 0 || qcElement.DynBitsUsed != qcOut.UsedDynBits {
		t.Fatalf("stale scratch dynamic bits = element %d frame %d", qcElement.DynBitsUsed, qcOut.UsedDynBits)
	}
}

func TestFDKaacEncQCMainQuantizePassReentersAfterConstraintReset(t *testing.T) {
	var cm ChannelMapping
	var adj AdjThrState
	var state ATSElement
	var psy PsyOutChannel
	var qc QCOutChannel
	var psyElement PsyOutElement
	var qcElement QCOutElement
	var elementBits ElementBits
	var qcOut QCOut
	var scratch QCMainQuantizeScratch
	fillQCMainQuantizeCase(&cm, &adj, &state, &psy, &qc, &psyElement, &qcElement, &elementBits, &qcOut)

	psyElements := [1]*PsyOutElement{&psyElement}
	qcElements := [1]*QCOutElement{&qcElement}
	elementBitsSlice := [1]*ElementBits{&elementBits}
	fdkaacEncQCMainQuantizeSetup(&cm, psyElements[:], &qcOut, qcElements[:], &scratch, 2, 0)
	initialGain := qc.GlobalGain
	fdkaacEncQCMainResetConstraints(&scratch)

	result, errCode := fdkaacEncQCMainQuantizePass(
		&cm,
		psyElements[:],
		&qcOut,
		qcElements[:],
		&adj,
		elementBitsSlice[:],
		&scratch,
		1,
		0,
		4,
		aotAACLC,
		0,
		-1,
	)
	if errCode != AACEncOK {
		t.Fatalf("QC quantize re-entry error = %#x, want OK", errCode)
	}
	want := QCMainQuantizeResult{
		QuantizedElements:      1,
		DynBitsConsumed:        22,
		SumDynBitsConsumed:     22,
		MaxValueAll:            4,
		DecreaseBitConsumption: 1,
		ConstraintsFulfilled:   1,
		ReductionIterations:    1,
		GainAdjustments:        1,
	}
	if result != want {
		t.Fatalf("QC quantize re-entry result = %+v, want %+v", result, want)
	}
	if scratch.Iterations[0] != 1 || qc.GlobalGain != initialGain+1 {
		t.Fatalf("QC quantize re-entry controls gain=%d initial=%d iterations=%d",
			qc.GlobalGain, initialGain, scratch.Iterations[0])
	}
	if qcOut.UsedDynBits != 22 || qcElement.DynBitsUsed != 22 {
		t.Fatalf("QC quantize re-entry accounting frame=%d element=%d", qcOut.UsedDynBits, qcElement.DynBitsUsed)
	}
}

func TestFDKaacEncQCMainPrepareRejectsInvalid(t *testing.T) {
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil element info", func() {
			var state ATSElement
			var psyElement PsyOutElement
			var qcElement QCOutElement
			FDKaacEncQCMainPrepare(nil, &state, &psyElement, &qcElement, aotAACLC, 0, -1)
		}},
		{"nil threshold state", func() {
			element, _, psyElement, qcElement := buildQCMainPrepareValidInput()
			FDKaacEncQCMainPrepare(&element, nil, &psyElement, &qcElement, aotAACLC, 0, -1)
		}},
		{"nil psy element", func() {
			element, state, _, qcElement := buildQCMainPrepareValidInput()
			FDKaacEncQCMainPrepare(&element, &state, nil, &qcElement, aotAACLC, 0, -1)
		}},
		{"nil QC element", func() {
			element, state, psyElement, _ := buildQCMainPrepareValidInput()
			FDKaacEncQCMainPrepare(&element, &state, &psyElement, nil, aotAACLC, 0, -1)
		}},
		{"invalid channel count", func() {
			element, state, psyElement, qcElement := buildQCMainPrepareValidInput()
			element.NChannelsInEl = 2
			FDKaacEncQCMainPrepare(&element, &state, &psyElement, &qcElement, aotAACLC, 0, -1)
		}},
		{"nil psy channel", func() {
			element, state, psyElement, qcElement := buildQCMainPrepareValidInput()
			psyElement.PsyOutChannel[0] = nil
			FDKaacEncQCMainPrepare(&element, &state, &psyElement, &qcElement, aotAACLC, 0, -1)
		}},
		{"nil QC channel", func() {
			element, state, psyElement, qcElement := buildQCMainPrepareValidInput()
			qcElement.QCOutChannel[0] = nil
			FDKaacEncQCMainPrepare(&element, &state, &psyElement, &qcElement, aotAACLC, 0, -1)
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

func TestFDKaacEncQCMainQuantizeRejectsInvalid(t *testing.T) {
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil mapping", func() {
			var qcOut QCOut
			var adj AdjThrState
			var scratch QCMainQuantizeScratch
			FDKaacEncQCMainQuantizeFrame(nil, nil, &qcOut, nil, &adj, nil, &scratch, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"nil output", func() {
			cm, adj, psyElements, qcElements, elementBits, _ := buildQCMainQuantizeValidInput()
			var scratch QCMainQuantizeScratch
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], nil, qcElements[:], &adj, elementBits[:], &scratch, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"nil scratch", func() {
			cm, adj, psyElements, qcElements, elementBits, qcOut := buildQCMainQuantizeValidInput()
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, elementBits[:], nil, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"nil threshold element", func() {
			cm, adj, psyElements, qcElements, elementBits, qcOut := buildQCMainQuantizeValidInput()
			var scratch QCMainQuantizeScratch
			adj.AdjThrStateElem[0] = nil
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, elementBits[:], &scratch, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"nil element bits", func() {
			cm, adj, psyElements, qcElements, elementBits, qcOut := buildQCMainQuantizeValidInput()
			var scratch QCMainQuantizeScratch
			elementBits[0] = nil
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, elementBits[:], &scratch, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"negative static bits", func() {
			cm, adj, psyElements, qcElements, elementBits, qcOut := buildQCMainQuantizeValidInput()
			var scratch QCMainQuantizeScratch
			qcElements[0].StaticBitsUsed = -1
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, elementBits[:], &scratch, 2, 0, 4, aotAACLC, 0, -1)
		}},
		{"negative max iterations", func() {
			cm, adj, psyElements, qcElements, elementBits, qcOut := buildQCMainQuantizeValidInput()
			var scratch QCMainQuantizeScratch
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, elementBits[:], &scratch, 2, 0, -1, aotAACLC, 0, -1)
		}},
		{"short max-value offsets", func() {
			var maxValue [4]uint32
			var quant [8]int16
			offsets := [...]int{0, 2, 4, 8}
			FDKaacEncCalcMaxValueInSfb(4, 2, 4, offsets[:], quant[:], maxValue[:])
		}},
		{"nil dynamic-bit sum", func() {
			cm, _, _, qcElements, _, _ := buildQCMainQuantizeValidInput()
			FDKaacEncUpdateUsedDynBits(nil, qcElements[:], &cm)
		}},
		{"nil consumed frame", func() {
			FDKaacEncGetTotalConsumedDynBits([]*QCOut{nil}, 1)
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

func TestFDKaacEncReduceBitConsumptionRejectsInvalid(t *testing.T) {
	var qc QCOutChannel
	element := QCOutElement{QCOutChannel: [2]*QCOutChannel{&qc}}
	elBits := ElementBits{MaxBitsEl: 128}
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil iterations", func() {
			chConstraints := [...]int{1}
			calculateQuant := [...]int{0}
			FDKaacEncReduceBitConsumption(nil, 4, 1, chConstraints[:], calculateQuant[:], 1, nil, nil, &element, &elBits, aotAACLC, 0, -1)
		}},
		{"negative iterations", func() {
			iterations := -1
			chConstraints := [...]int{1}
			calculateQuant := [...]int{0}
			FDKaacEncReduceBitConsumption(&iterations, 4, 1, chConstraints[:], calculateQuant[:], 1, nil, nil, &element, &elBits, aotAACLC, 0, -1)
		}},
		{"bad gain adjustment", func() {
			iterations := 0
			chConstraints := [...]int{1}
			calculateQuant := [...]int{0}
			FDKaacEncReduceBitConsumption(&iterations, 4, 0, chConstraints[:], calculateQuant[:], 1, nil, nil, &element, &elBits, aotAACLC, 0, -1)
		}},
		{"short constraints", func() {
			iterations := 0
			calculateQuant := [...]int{0}
			FDKaacEncReduceBitConsumption(&iterations, 4, 1, nil, calculateQuant[:], 1, nil, nil, &element, &elBits, aotAACLC, 0, -1)
		}},
		{"short calculate flags", func() {
			iterations := 0
			chConstraints := [...]int{1}
			FDKaacEncReduceBitConsumption(&iterations, 4, 1, chConstraints[:], nil, 1, nil, nil, &element, &elBits, aotAACLC, 0, -1)
		}},
		{"nil element", func() {
			iterations := 0
			chConstraints := [...]int{1}
			calculateQuant := [...]int{0}
			FDKaacEncReduceBitConsumption(&iterations, 4, 1, chConstraints[:], calculateQuant[:], 1, nil, nil, nil, &elBits, aotAACLC, 0, -1)
		}},
		{"nil element bits", func() {
			iterations := 0
			chConstraints := [...]int{1}
			calculateQuant := [...]int{0}
			FDKaacEncReduceBitConsumption(&iterations, 4, 1, chConstraints[:], calculateQuant[:], 1, nil, nil, &element, nil, aotAACLC, 0, -1)
		}},
		{"nil output channel", func() {
			iterations := 0
			chConstraints := [...]int{1}
			calculateQuant := [...]int{0}
			bad := QCOutElement{}
			FDKaacEncReduceBitConsumption(&iterations, 4, 1, chConstraints[:], calculateQuant[:], 1, nil, nil, &bad, &elBits, aotAACLC, 0, -1)
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

func TestFDKaacEncQCAccountingRejectsInvalid(t *testing.T) {
	var cm ChannelMapping
	cm.NElements = 1
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	var qcOut QCOut
	var qcElement QCOutElement
	qcOutFrames := [1]*QCOut{&qcOut}
	var qcElements [1][maxChannelElements]*QCOutElement
	qcElements[0][0] = &qcElement
	kernel := QCKernel{MaxBitsPerFrame: 1024, BitResTotMax: 1000}
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil consumed mapping", func() {
			FDKaacEncGetTotalConsumedBits(qcOutFrames[:], qcElements[:], nil, 56, 1)
		}},
		{"negative consumed header", func() {
			FDKaacEncGetTotalConsumedBits(qcOutFrames[:], qcElements[:], &cm, -1, 1)
		}},
		{"nil consumed element", func() {
			badElements := qcElements
			badElements[0][0] = nil
			FDKaacEncGetTotalConsumedBits(qcOutFrames[:], badElements[:], &cm, 56, 1)
		}},
		{"nil fill kernel", func() {
			FDKaacEncUpdateFillBits(nil, qcOutFrames[:])
		}},
		{"nil fill frame", func() {
			FDKaacEncUpdateFillBits(&kernel, nil)
		}},
		{"negative fill bits", func() {
			bad := QCOut{UsedDynBits: -1}
			badFrames := [1]*QCOut{&bad}
			FDKaacEncUpdateFillBits(&kernel, badFrames[:])
		}},
		{"nil finalize kernel", func() {
			FDKaacEncFinalizeBitConsumption(nil, &qcOut, 56, aotAACLC, 0, -1)
		}},
		{"nil finalize frame", func() {
			FDKaacEncFinalizeBitConsumption(&kernel, nil, 56, aotAACLC, 0, -1)
		}},
		{"negative finalize transport", func() {
			FDKaacEncFinalizeBitConsumption(&kernel, &qcOut, -1, aotAACLC, 0, -1)
		}},
		{"transport header grew", func() {
			cbrKernel := QCKernel{GlobHdrBits: 40, MaxBitsPerFrame: 1024, BitrateMode: QCBitrateModeCBR}
			FDKaacEncFinalizeBitConsumption(&cbrKernel, &qcOut, 56, aotAACLC, 0, -1)
		}},
		{"nil bitres kernel", func() {
			FDKaacEncUpdateBitres(nil, qcOutFrames[:])
		}},
		{"nil bitres frame", func() {
			FDKaacEncUpdateBitres(&kernel, nil)
		}},
		{"negative bitres frame", func() {
			bad := QCOut{UsedDynBits: -1}
			badFrames := [1]*QCOut{&bad}
			FDKaacEncUpdateBitres(&kernel, badFrames[:])
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

	overflowKernel := QCKernel{MaxBitsPerFrame: 100, MinBitsPerFrame: 0}
	overflowOut := QCOut{StaticBits: 80, UsedDynBits: 80}
	if errCode := FDKaacEncFinalizeBitConsumption(&overflowKernel, &overflowOut, 56, aotAACLC, 0, -1); errCode != AACEncQuantError {
		t.Fatalf("overflow finalize error = %#x, want quant error", errCode)
	}
}

func TestFDKaacEncBitDistributionRejectsInvalid(t *testing.T) {
	var cm ChannelMapping
	cm.NElements = 1
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	var qcElement QCOutElement
	var bits ElementBits
	qcElements := [maxChannelElements]*QCOutElement{&qcElement}
	elementBits := [maxChannelElements]*ElementBits{&bits}
	var state AdjThrState
	FDKaacEncInitBitresState(&state)
	elementState := buildBitresElementWithHistory(420, 650)
	state.AdjThrStateElem[0] = &elementState
	var psy PsyOutChannel
	psy.LastWindowSequence = LongWindow
	psyElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&psy}}
	var qcOut QCOut
	psyElements := [1]*PsyOutElement{&psyElement}
	qcOutFrames := [1]*QCOut{&qcOut}
	var qcElementFrames [1][maxChannelElements]*QCOutElement
	qcElementFrames[0][0] = &qcElement
	kernel := QCKernel{MaxBitsPerFrame: 1024, MaxBitFac: MaxValDBL, BitResMode: BitresModeFull}

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil distribute mapping", func() {
			FDKaacEncDistributeElementDynBits(qcElements[:], nil, elementBits[:], 100)
		}},
		{"nil distribute element", func() {
			badElements := qcElements
			badElements[0] = nil
			FDKaacEncDistributeElementDynBits(badElements[:], &cm, elementBits[:], 100)
		}},
		{"nil distribute element bits", func() {
			badBits := elementBits
			badBits[0] = nil
			FDKaacEncDistributeElementDynBits(qcElements[:], &cm, badBits[:], 100)
		}},
		{"nil redistribution kernel", func() {
			FDKaacEncBitResRedistribution(nil, &cm, elementBits[:], 100)
		}},
		{"nil redistribution mapping", func() {
			FDKaacEncBitResRedistribution(&kernel, nil, elementBits[:], 100)
		}},
		{"nil prepare kernel", func() {
			FDKaacEncPrepareBitDistribution(nil, &state, psyElements[:], qcOutFrames[:], qcElementFrames[:], &cm, elementBits[:], 720)
		}},
		{"nil prepare state", func() {
			FDKaacEncPrepareBitDistribution(&kernel, nil, psyElements[:], qcOutFrames[:], qcElementFrames[:], &cm, elementBits[:], 720)
		}},
		{"nil prepare frame", func() {
			FDKaacEncPrepareBitDistribution(&kernel, &state, psyElements[:], nil, qcElementFrames[:], &cm, elementBits[:], 720)
		}},
		{"nil prepare psy element", func() {
			badPsy := psyElements
			badPsy[0] = nil
			FDKaacEncPrepareBitDistribution(&kernel, &state, badPsy[:], qcOutFrames[:], qcElementFrames[:], &cm, elementBits[:], 720)
		}},
		{"bad prepare bitres mode", func() {
			badKernel := kernel
			badKernel.BitResMode = BitresMode(99)
			FDKaacEncPrepareBitDistribution(&badKernel, &state, psyElements[:], qcOutFrames[:], qcElementFrames[:], &cm, elementBits[:], 720)
		}},
		{"bad prepare element max reservoir", func() {
			badBits := bits
			badBits.MaxBitResBitsEl = 0
			badElementBits := [maxChannelElements]*ElementBits{&badBits}
			FDKaacEncPrepareBitDistribution(&kernel, &state, psyElements[:], qcOutFrames[:], qcElementFrames[:], &cm, badElementBits[:], 720)
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

	lowKernel := QCKernel{GlobHdrBits: 800, MaxBitsPerFrame: 1024, BitResTot: 0, MaxBitFac: MaxValDBL, BitResMode: BitresModeFull}
	bits = ElementBits{RelativeBitsEl: 0x40000000, BitResLevelEl: 500, MaxBitResBitsEl: 1200}
	elementBits = [maxChannelElements]*ElementBits{&bits}
	_, errCode := FDKaacEncPrepareBitDistribution(&lowKernel, &state, psyElements[:], qcOutFrames[:], qcElementFrames[:], &cm, elementBits[:], 720)
	if errCode != AACEncBitresTooLow {
		t.Fatalf("low prepare reservoir error = %#x, want %#x", errCode, AACEncBitresTooLow)
	}
}

func TestFDKaacEncQCAccountingAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		cm := ChannelMapping{NElements: 1}
		cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
		qcOut := QCOut{
			GrantedDynBits: 500,
			StaticBits:     29,
			UsedDynBits:    211,
			TotFillBits:    31,
			ElementExtBits: 4,
			GlobalExtBits:  6,
		}
		element := QCOutElement{DynBitsUsed: 211, StaticBitsUsed: 29, ExtBitsUsed: 4}
		frames := [1]*QCOut{&qcOut}
		var elements [1][maxChannelElements]*QCOutElement
		elements[0][0] = &element
		kernel := QCKernel{
			GlobHdrBits:     64,
			MaxBitsPerFrame: 1024,
			BitrateMode:     QCBitrateModeCBR,
			BitResTot:       100,
			BitResTotMax:    1000,
		}
		qcMainPrepareSink = FDKaacEncGetTotalConsumedBits(frames[:], elements[:], &cm, 56, 1)
		if errCode := FDKaacEncUpdateFillBits(&kernel, frames[:]); errCode != AACEncOK {
			t.Fatalf("fill error = %#x", errCode)
		}
		if errCode := FDKaacEncFinalizeBitConsumption(&kernel, &qcOut, 56, aotAACLC, 0, -1); errCode != AACEncOK {
			t.Fatalf("finalize error = %#x", errCode)
		}
		FDKaacEncUpdateBitres(&kernel, frames[:])
		qcMainPrepareSink += qcOut.TotalBits + kernel.BitResTot
	})
	if allocs != 0 {
		t.Fatalf("QC accounting allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncBitDistributionAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var state AdjThrState
		FDKaacEncInitBitresState(&state)
		elementState := buildBitresElementWithHistory(420, 650)
		state.AdjThrStateElem[0] = &elementState
		psy := PsyOutChannel{LastWindowSequence: LongWindow}
		psyElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&psy}}
		qcElement := QCOutElement{PEData: PEData{Pe: 430}}
		qcOut := QCOut{}
		cm := ChannelMapping{NElements: 1}
		cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
		elBits := ElementBits{RelativeBitsEl: 0x40000000, BitResLevelEl: 500, MaxBitResBitsEl: 1200}
		kernel := QCKernel{MaxBitsPerFrame: 2000, MaxBitFac: MaxValDBL, BitResMode: BitresModeFull}
		psyElements := [1]*PsyOutElement{&psyElement}
		qcOutFrames := [1]*QCOut{&qcOut}
		var qcElements [1][maxChannelElements]*QCOutElement
		qcElements[0][0] = &qcElement
		elementBits := [maxChannelElements]*ElementBits{&elBits}

		result, errCode := FDKaacEncPrepareBitDistribution(&kernel, &state, psyElements[:], qcOutFrames[:], qcElements[:], &cm, elementBits[:], 720)
		if errCode != AACEncOK {
			t.Fatalf("prepare bit distribution error = %#x", errCode)
		}
		qcMainPrepareSink = result.TotalAvailableBits + result.TotalGrantedPeCorr + qcElement.GrantedDynBits
	})
	if allocs != 0 {
		t.Fatalf("bit distribution allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncCrashRecoveryAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var qc QCOutChannel
		var psy PsyOutChannel
		var psyElement PsyOutElement
		var qcOut QCOut
		fillQCCrashRecoveryCase(&psy, &qc, &psyElement, &qcOut)
		qcElement := QCOutElement{
			StaticBitsUsed: qcOut.StaticBits,
			GrantedDynBits: 10,
			QCOutChannel:   [2]*QCOutChannel{&qc},
		}

		result, errCode := FDKaacEncCrashRecovery(1, &psyElement, &qcOut, &qcElement, 10000, aotAACLC, 0, -1)
		if errCode != AACEncOK {
			t.Fatalf("crash recovery error = %#x", errCode)
		}
		qcMainPrepareSink = result.SavedBits + qcElement.GrantedDynBits + qcOut.MaxDynBits
	})
	if allocs != 0 {
		t.Fatalf("crash recovery allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncQCMainFrameAllocs(t *testing.T) {
	var baseCM ChannelMapping
	var baseAdj AdjThrState
	var baseState ATSElement
	var basePsy PsyOutChannel
	var baseQC QCOutChannel
	var basePsyElement PsyOutElement
	var baseQCElement QCOutElement
	var baseElementBits ElementBits
	var baseQCOut QCOut
	var baseKernel QCKernel
	fillQCMainFrameCase(
		&baseCM,
		&baseAdj,
		&baseState,
		&basePsy,
		&baseQC,
		&basePsyElement,
		&baseQCElement,
		&baseElementBits,
		&baseQCOut,
		&baseKernel,
	)

	allocs := testing.AllocsPerRun(1000, func() {
		cm := baseCM
		adj := baseAdj
		state := baseState
		psy := basePsy
		qc := baseQC
		psyElement := basePsyElement
		qcElement := baseQCElement
		elementBits := baseElementBits
		qcOut := baseQCOut
		kernel := baseKernel
		var scratch QCMainFrameScratch
		adj.AdjThrStateElem[0] = &state
		psy.MdctSpectrum = qc.MdctSpectrum[:]
		psyElement.PsyOutChannel[0] = &psy
		qcElement.QCOutChannel[0] = &qc
		psyElements := [1]*PsyOutElement{&psyElement}
		qcOutFrames := [1]*QCOut{&qcOut}
		var qcElements [1][maxChannelElements]*QCOutElement
		qcElements[0][0] = &qcElement
		elementBitsSlice := [maxChannelElements]*ElementBits{&elementBits}

		result, errCode := FDKaacEncQCMainFrame(&kernel, &adj, psyElements[:], qcOutFrames[:], qcElements[:], &cm, elementBitsSlice[:], &scratch, 720, 2, 0, 4, aotAACLC, 0, -1)
		if errCode != AACEncOK {
			t.Fatalf("QC main frame error = %#x", errCode)
		}
		qcMainPrepareSink = result.SumDynBitsConsumed + result.SumBitsConsumed + qcOut.GrantedDynBits
	})
	if allocs != 0 {
		t.Fatalf("QC main frame allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncQCMainQuantizeAllocs(t *testing.T) {
	var baseCM ChannelMapping
	var baseAdj AdjThrState
	var baseState ATSElement
	var basePsy PsyOutChannel
	var baseQC QCOutChannel
	var basePsyElement PsyOutElement
	var baseQCElement QCOutElement
	var baseElementBits ElementBits
	var baseQCOut QCOut
	fillQCMainQuantizeCase(
		&baseCM,
		&baseAdj,
		&baseState,
		&basePsy,
		&baseQC,
		&basePsyElement,
		&baseQCElement,
		&baseElementBits,
		&baseQCOut,
	)

	allocs := testing.AllocsPerRun(1000, func() {
		cm := baseCM
		adj := baseAdj
		state := baseState
		psy := basePsy
		qc := baseQC
		psyElement := basePsyElement
		qcElement := baseQCElement
		elementBits := baseElementBits
		qcOut := baseQCOut
		var scratch QCMainQuantizeScratch
		adj.AdjThrStateElem[0] = &state
		psy.MdctSpectrum = qc.MdctSpectrum[:]
		psyElement.PsyOutChannel[0] = &psy
		qcElement.QCOutChannel[0] = &qc
		psyElements := [1]*PsyOutElement{&psyElement}
		qcElements := [1]*QCOutElement{&qcElement}
		elementBitsSlice := [1]*ElementBits{&elementBits}
		result, errCode := FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, elementBitsSlice[:], &scratch, 2, 0, 4, aotAACLC, 0, -1)
		if errCode != AACEncOK {
			t.Fatalf("QC quantize error = %#x", errCode)
		}
		qcMainPrepareSink = result.DynBitsConsumed + qcOut.UsedDynBits + state.DynBitsLast
		qcMainPrepareHashSink = hashInt16(qc.QuantSpec[:11]) ^ hashUint32(qc.MaxValueInSfb[:5])
	})
	if allocs != 0 {
		t.Fatalf("QC quantize allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncQCMainPrepareAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var element ElementInfo
		var state ATSElement
		var psy PsyOutChannel
		var qc QCOutChannel
		var tools ToolsInfo
		var psyElement PsyOutElement
		var qcElement QCOutElement
		var mdct [maxSpectralLines]FixpDBL
		fillQCMainPrepareCase(
			&element, &state, &psy, &qc, &tools, &psyElement, &qcElement, &mdct,
		)
		errCode := FDKaacEncQCMainPrepare(&element, &state, &psyElement, &qcElement, aotAACLC, 0, -1)
		if errCode != AACEncOK {
			t.Fatalf("QC prepare error = %#x", errCode)
		}
		qcMainPrepareSink = qcElement.StaticBitsUsed + int(qcElement.PEData.Pe)
		qcMainPrepareHashSink = hashFixpDBL(qc.SfbFormFactorLdData[:8]) ^ hashFixpDBL(qc.SfbThresholdLdData[:8])
	})
	if allocs != 0 {
		t.Fatalf("QC prepare allocations = %v, want 0", allocs)
	}
}

func buildQCMainPrepareValidInput() (ElementInfo, ATSElement, PsyOutElement, QCOutElement) {
	var element ElementInfo
	var state ATSElement
	var psy PsyOutChannel
	var qc QCOutChannel
	var tools ToolsInfo
	var psyElement PsyOutElement
	var qcElement QCOutElement
	var mdct [maxSpectralLines]FixpDBL
	fillQCMainPrepareCase(&element, &state, &psy, &qc, &tools, &psyElement, &qcElement, &mdct)
	return element, state, psyElement, qcElement
}

func buildQCMainQuantizeValidInput() (ChannelMapping, AdjThrState, [1]*PsyOutElement, [1]*QCOutElement, [1]*ElementBits, QCOut) {
	var cm ChannelMapping
	var adj AdjThrState
	var state ATSElement
	var psy PsyOutChannel
	var qc QCOutChannel
	var psyElement PsyOutElement
	var qcElement QCOutElement
	var elementBits ElementBits
	var qcOut QCOut
	fillQCMainQuantizeCase(&cm, &adj, &state, &psy, &qc, &psyElement, &qcElement, &elementBits, &qcOut)
	return cm, adj, [1]*PsyOutElement{&psyElement}, [1]*QCOutElement{&qcElement}, [1]*ElementBits{&elementBits}, qcOut
}

func fillQCMainFrameCase(
	cm *ChannelMapping,
	adj *AdjThrState,
	state *ATSElement,
	psy *PsyOutChannel,
	qc *QCOutChannel,
	psyElement *PsyOutElement,
	qcElement *QCOutElement,
	elementBits *ElementBits,
	qcOut *QCOut,
	kernel *QCKernel,
) {
	fillQCMainQuantizeCase(cm, adj, state, psy, qc, psyElement, qcElement, elementBits, qcOut)
	FDKaacEncInitBitresState(adj)
	*state = buildBitresElementWithHistory(0, -1)
	adj.AdjThrStateElem[0] = state
	qcElement.PEData = PEData{Pe: 10}
	qcOut.StaticBits = qcElement.StaticBitsUsed
	elementBits.RelativeBitsEl = 0x40000000
	elementBits.BitResLevelEl = 0
	elementBits.MaxBitResBitsEl = 0
	*kernel = QCKernel{
		MaxBitsPerFrame: 2000,
		BitrateMode:     QCBitrateModeCBR,
		BitResMode:      BitresModeDisabled,
		MaxBitFac:       MaxValDBL,
	}
}

func fillQCCrashRecoveryCase(
	psy *PsyOutChannel,
	qc *QCOutChannel,
	psyElement *PsyOutElement,
	qcOut *QCOut,
) {
	*psy = PsyOutChannel{
		SfbCnt:             4,
		SfbPerGroup:        4,
		MaxSfbPerGroup:     4,
		LastWindowSequence: LongWindow,
		WindowShape:        WindowShapeKBD,
	}
	copy(psy.SfbOffsets[:], []int{0, 4, 8, 12, 16})
	psy.TNSInfo.NumOfFilters[0] = 1
	psy.TNSInfo.CoefRes[0] = 4
	psy.TNSInfo.Length[0][0] = 4
	psy.TNSInfo.Order[0][0] = 1
	psy.TNSInfo.Coef[0][0][0] = 2

	*qc = QCOutChannel{GlobalGain: 100}
	qc.SectionData = SectionData{
		BlockType:      LongWindow,
		NoOfGroups:     1,
		SfbCnt:         4,
		MaxSfbPerGroup: 4,
		SfbPerGroup:    4,
		NoOfSections:   1,
		FirstScf:       0,
	}
	qc.SectionData.Huffsection[0] = SectionInfo{CodeBook: codeBook1No, SfbStart: 0, SfbCnt: 4}
	for i := 0; i < 4; i++ {
		qc.Scf[i] = 100
	}
	for i := 0; i < 16; i++ {
		qc.QuantSpec[i] = 1
	}

	*psyElement = PsyOutElement{
		ToolsInfo:     ToolsInfo{MsDigest: MsMaskSome, MsMask: [maxGroupedSFB]int{1, 1, 1, 1}},
		PsyOutChannel: [2]*PsyOutChannel{psy},
	}
	elInfo := ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	psyChannels := [1]*PsyOutChannel{psy}
	qcChannels := [1]*QCOutChannel{qc}
	staticBits, errCode := FDKaacEncChannelElementWrite(nil, &elInfo, qcChannels[:], psyElement, psyChannels[:], 0, aotAACLC, -1, 0)
	if errCode != AACEncOK {
		panic("fdkaac test: crash-recovery fixture failed")
	}
	*qcOut = QCOut{StaticBits: staticBits, GrantedDynBits: 10, MaxDynBits: 100}
}

func runDirectQCMainQuantizeFrame(
	cm *ChannelMapping,
	psyOutElement []*PsyOutElement,
	qcOut *QCOut,
	qcElement []*QCOutElement,
	adjThrState *AdjThrState,
	elementBits []*ElementBits,
	scratch *QCMainQuantizeScratch,
	invQuant int,
	dZoneQuantEnable int,
	maxIterations int,
	syntaxFlags uint32,
) (QCMainQuantizeResult, int) {
	_ = elementBits
	_ = maxIterations
	result := QCMainQuantizeResult{ConstraintsFulfilled: 1, DecreaseBitConsumption: -1}
	for i := 0; i < cm.NElements; i++ {
		elInfo := cm.ElInfo[i]
		if !fdkaacEncIsAdjustableElement(elInfo.ElType) {
			continue
		}
		nChannels := elInfo.NChannelsInEl
		psyChannels := psyOutElement[i].PsyOutChannel[:nChannels]
		qcChannels := qcElement[i].QCOutChannel[:nChannels]

		FDKaacEncEstimateScaleFactors(psyChannels, qcChannels, invQuant, dZoneQuantEnable, nChannels, scratch.QuantSpecTmp[:])
		for ch := 0; ch < nChannels; ch++ {
			psyCh := psyChannels[ch]
			qcCh := qcChannels[ch]
			FDKaacEncQuantizeSpectrum(psyCh.SfbCnt, psyCh.MaxSfbPerGroup, psyCh.SfbPerGroup, psyCh.SfbOffsets[:], qcCh.MdctSpectrum[:], qcCh.GlobalGain, qcCh.Scf[:], qcCh.QuantSpec[:], dZoneQuantEnable)
			maxValue := FDKaacEncCalcMaxValueInSfb(psyCh.SfbCnt, psyCh.MaxSfbPerGroup, psyCh.SfbPerGroup, psyCh.SfbOffsets[:], qcCh.QuantSpec[:], qcCh.MaxValueInSfb[:])
			if maxValue > result.MaxValueAll {
				result.MaxValueAll = maxValue
			}
			if maxValue > maxQuant {
				result.ConstraintsFulfilled = 0
				result.DecreaseBitConsumption = 1
				return result, AACEncQuantError
			}
		}

		qcElement[i].DynBitsUsed = 0
		for ch := 0; ch < nChannels; ch++ {
			psyCh := psyChannels[ch]
			qcCh := qcChannels[ch]
			chDynBits := FDKaacEncDynBitCount(&scratch.BitCounter, qcCh.QuantSpec[:], qcCh.MaxValueInSfb[:], qcCh.Scf[:], psyCh.LastWindowSequence, psyCh.SfbCnt, psyCh.MaxSfbPerGroup, psyCh.SfbPerGroup, psyCh.SfbOffsets[:], &qcCh.SectionData, psyCh.NoiseNrg[:], psyCh.IsBook[:], psyCh.IsScale[:], syntaxFlags)
			qcElement[i].DynBitsUsed += chDynBits
		}
		if adjThrState.AdjThrStateElem[i].DynBitsLast == -1 {
			adjThrState.AdjThrStateElem[i].DynBitsLast = qcElement[i].DynBitsUsed
		}
		result.DynBitsConsumed += qcElement[i].DynBitsUsed
		result.QuantizedElements++
		maxElementDynBits := nChannels*minBufSizePerEffChan - qcElement[i].StaticBitsUsed - qcElement[i].ExtBitsUsed
		if qcElement[i].DynBitsUsed > maxElementDynBits {
			result.ConstraintsFulfilled = 0
			result.DecreaseBitConsumption = 1
			return result, AACEncQuantError
		}
	}
	FDKaacEncUpdateUsedDynBits(&qcOut.UsedDynBits, qcElement, cm)
	result.SumDynBitsConsumed = FDKaacEncGetTotalConsumedDynBits([]*QCOut{qcOut}, 1)
	return result, AACEncOK
}

func fillQCMainQuantizeCase(
	cm *ChannelMapping,
	adj *AdjThrState,
	state *ATSElement,
	psy *PsyOutChannel,
	qc *QCOutChannel,
	psyElement *PsyOutElement,
	qcElement *QCOutElement,
	elementBits *ElementBits,
	qcOut *QCOut,
) {
	var quantTmp [11]int16
	var minScf [5]int
	var constPart [5]FixpDBL
	activeThreshold := [5]FixpDBL{-400000000, -400000000, -400000000, -400000000, -400000000}
	setupAssimilateSingleVector(psy, qc, qc.QuantSpec[:11], quantTmp[:], minScf[:], constPart[:], qc.SfbFormFactorLdData[:5])
	copy(qc.SfbThresholdLdData[:], activeThreshold[:])
	psy.LastWindowSequence = LongWindow
	psy.WindowShape = WindowShapeKBD
	psy.MdctSpectrum = qc.MdctSpectrum[:]

	*state = ATSElement{DynBitsLast: -1}
	*adj = AdjThrState{}
	adj.AdjThrStateElem[0] = state
	*cm = ChannelMapping{NElements: 1}
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, InstanceTag: 2, NChannelsInEl: 1}
	*psyElement = PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{psy}}
	*qcElement = QCOutElement{StaticBitsUsed: 29, QCOutChannel: [2]*QCOutChannel{qc}}
	*elementBits = ElementBits{MaxBitsEl: minBufSizePerEffChan, MaxBitResBitsEl: minBufSizePerEffChan}
	*qcOut = QCOut{UsedDynBits: -1}
}

func fillQCMainPrepareCase(
	element *ElementInfo,
	state *ATSElement,
	psy *PsyOutChannel,
	qc *QCOutChannel,
	tools *ToolsInfo,
	psyElement *PsyOutElement,
	qcElement *QCOutElement,
	mdct *[maxSpectralLines]FixpDBL,
) {
	var peData PEData
	fillAdjThrLongPatchCase(&peData, psy, qc, tools, state)
	psy.MdctSpectrum = mdct[:]
	psy.WindowShape = WindowShapeKBD
	for i := 0; i < psy.SfbOffsets[psy.SfbCnt]; i++ {
		v := FixpDBL((i + 1) * 0x00080000)
		if i&1 != 0 {
			v = -v
		}
		mdct[i] = v
	}

	*element = ElementInfo{ElType: idSCE, InstanceTag: 2, NChannelsInEl: 1}
	*psyElement = PsyOutElement{
		ToolsInfo:     *tools,
		PsyOutChannel: [2]*PsyOutChannel{psy},
	}
	*qcElement = QCOutElement{
		QCOutChannel: [2]*QCOutChannel{qc},
	}
}

func hashInt16(x []int16) uint64 {
	h := uint64(1469598103934665603)
	for _, v := range x {
		h ^= uint64(uint16(v))
		h *= 1099511628211
	}
	return h
}

func hashUint32(x []uint32) uint64 {
	h := uint64(1469598103934665603)
	for _, v := range x {
		h ^= uint64(v)
		h *= 1099511628211
	}
	return h
}
