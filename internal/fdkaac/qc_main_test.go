package fdkaac

import "testing"

var qcMainPrepareSink int
var qcMainPrepareHashSink uint64

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
