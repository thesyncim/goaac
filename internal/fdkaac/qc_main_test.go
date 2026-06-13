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
	var directScratch QCMainQuantizeScratch
	fillQCMainQuantizeCase(&directCM, &directAdj, &directState, &directPsy, &directQC, &directPsyElement, &directQCElement, &directQCOut)

	directPsyElements := [1]*PsyOutElement{&directPsyElement}
	directQCElements := [1]*QCOutElement{&directQCElement}
	directResult, directErr := runDirectQCMainQuantizeFrame(
		&directCM, directPsyElements[:], &directQCOut, directQCElements[:], &directAdj, &directScratch, 2, 0, 0,
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
	var wrappedScratch QCMainQuantizeScratch
	fillQCMainQuantizeCase(&wrappedCM, &wrappedAdj, &wrappedState, &wrappedPsy, &wrappedQC, &wrappedPsyElement, &wrappedQCElement, &wrappedQCOut)

	wrappedPsyElements := [1]*PsyOutElement{&wrappedPsyElement}
	wrappedQCElements := [1]*QCOutElement{&wrappedQCElement}
	wrappedResult, wrappedErr := FDKaacEncQCMainQuantizeFrame(
		&wrappedCM, wrappedPsyElements[:], &wrappedQCOut, wrappedQCElements[:], &wrappedAdj, &wrappedScratch, 2, 0, 0,
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
			FDKaacEncQCMainQuantizeFrame(nil, nil, &qcOut, nil, &adj, &scratch, 2, 0, 0)
		}},
		{"nil output", func() {
			cm, adj, psyElements, qcElements, _ := buildQCMainQuantizeValidInput()
			var scratch QCMainQuantizeScratch
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], nil, qcElements[:], &adj, &scratch, 2, 0, 0)
		}},
		{"nil scratch", func() {
			cm, adj, psyElements, qcElements, qcOut := buildQCMainQuantizeValidInput()
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, nil, 2, 0, 0)
		}},
		{"nil threshold element", func() {
			cm, adj, psyElements, qcElements, qcOut := buildQCMainQuantizeValidInput()
			var scratch QCMainQuantizeScratch
			adj.AdjThrStateElem[0] = nil
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, &scratch, 2, 0, 0)
		}},
		{"negative static bits", func() {
			cm, adj, psyElements, qcElements, qcOut := buildQCMainQuantizeValidInput()
			var scratch QCMainQuantizeScratch
			qcElements[0].StaticBitsUsed = -1
			FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, &scratch, 2, 0, 0)
		}},
		{"short max-value offsets", func() {
			var maxValue [4]uint32
			var quant [8]int16
			offsets := [...]int{0, 2, 4, 8}
			FDKaacEncCalcMaxValueInSfb(4, 2, 4, offsets[:], quant[:], maxValue[:])
		}},
		{"nil dynamic-bit sum", func() {
			cm, _, _, qcElements, _ := buildQCMainQuantizeValidInput()
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

func TestFDKaacEncQCMainQuantizeAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var cm ChannelMapping
		var adj AdjThrState
		var state ATSElement
		var psy PsyOutChannel
		var qc QCOutChannel
		var psyElement PsyOutElement
		var qcElement QCOutElement
		var qcOut QCOut
		var scratch QCMainQuantizeScratch
		fillQCMainQuantizeCase(&cm, &adj, &state, &psy, &qc, &psyElement, &qcElement, &qcOut)
		psyElements := [1]*PsyOutElement{&psyElement}
		qcElements := [1]*QCOutElement{&qcElement}
		result, errCode := FDKaacEncQCMainQuantizeFrame(&cm, psyElements[:], &qcOut, qcElements[:], &adj, &scratch, 2, 0, 0)
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

func buildQCMainQuantizeValidInput() (ChannelMapping, AdjThrState, [1]*PsyOutElement, [1]*QCOutElement, QCOut) {
	var cm ChannelMapping
	var adj AdjThrState
	var state ATSElement
	var psy PsyOutChannel
	var qc QCOutChannel
	var psyElement PsyOutElement
	var qcElement QCOutElement
	var qcOut QCOut
	fillQCMainQuantizeCase(&cm, &adj, &state, &psy, &qc, &psyElement, &qcElement, &qcOut)
	return cm, adj, [1]*PsyOutElement{&psyElement}, [1]*QCOutElement{&qcElement}, qcOut
}

func runDirectQCMainQuantizeFrame(
	cm *ChannelMapping,
	psyOutElement []*PsyOutElement,
	qcOut *QCOut,
	qcElement []*QCOutElement,
	adjThrState *AdjThrState,
	scratch *QCMainQuantizeScratch,
	invQuant int,
	dZoneQuantEnable int,
	syntaxFlags uint32,
) (QCMainQuantizeResult, int) {
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
