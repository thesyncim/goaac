package fdkaac

import (
	"bytes"
	"testing"
)

func TestFDKaacEncEncodeGlobalGainAndICSVectors(t *testing.T) {
	runBitencVector(t, "global-gain", func(bs *BitStream) int {
		return FDKaacEncEncodeGlobalGain(80, 100, bs, 0)
	}, 8, []byte{0x8c})

	runBitencVector(t, "long-ics", func(bs *BitStream) int {
		return FDKaacEncEncodeIcsInfo(LongWindow, WindowShapeKBD, 0, 8, bs, 0)
	}, 11, []byte{0x12, 0x00})

	runBitencVector(t, "short-ics", func(bs *BitStream) int {
		return FDKaacEncEncodeIcsInfo(ShortWindow, WindowShapeSine, 0x6c, 4, bs, 0)
	}, 15, []byte{0x44, 0xd8})

	if got := FDKaacEncEncodeGlobalGain(80, 100, nil, 0); got != 8 {
		t.Fatalf("nil global-gain bit count = %d, want 8", got)
	}
	if got := FDKaacEncEncodeIcsInfo(LongWindow, WindowShapeKBD, 0, 8, nil, 0); got != 11 {
		t.Fatalf("nil long ICS bit count = %d, want 11", got)
	}
}

func TestFDKaacEncEncodeSectionDataVectors(t *testing.T) {
	_, longSD := buildBitencSectionData(t, fillDynLongCase)
	runBitencVector(t, "long-section", func(bs *BitStream) int {
		return FDKaacEncEncodeSectionData(&longSD, bs, false)
	}, 18, []byte{0x11, 0xd9, 0x40})

	_, shortSD := buildBitencSectionData(t, fillDynShortCase)
	runBitencVector(t, "short-section", func(bs *BitStream) int {
		return FDKaacEncEncodeSectionData(&shortSD, bs, false)
	}, 21, []byte{0x48, 0x06, 0x98})

	_, pnsISSD := buildBitencSectionData(t, fillDynPNSIntensityCase)
	runBitencVector(t, "pns-is-section", func(bs *BitStream) int {
		return FDKaacEncEncodeSectionData(&pnsISSD, bs, false)
	}, 45, []byte{0x00, 0xe8, 0xb8, 0x3e, 0x12, 0x08})

	if got := FDKaacEncEncodeSectionData(&longSD, nil, false); got != 0 {
		t.Fatalf("nil section-data bit count = %d, want 0", got)
	}
}

func TestFDKaacEncEncodeScaleFactorDataVectors(t *testing.T) {
	longTC, longSD := buildBitencSectionData(t, fillDynLongCase)
	runBitencVector(t, "long-scalefactor", func(bs *BitStream) int {
		return FDKaacEncEncodeScaleFactorData(
			longTC.maxValue[:longTC.sfbCnt], &longSD, longTC.scf[:longTC.sfbCnt],
			bs, longTC.noise[:longTC.sfbCnt], longTC.isScale[:longTC.sfbCnt], 120,
		)
	}, 29, []byte{0x55, 0x55, 0x55, 0x50})

	shortTC, shortSD := buildBitencSectionData(t, fillDynShortCase)
	runBitencVector(t, "short-scalefactor", func(bs *BitStream) int {
		return FDKaacEncEncodeScaleFactorData(
			shortTC.maxValue[:shortTC.sfbCnt], &shortSD, shortTC.scf[:shortTC.sfbCnt],
			bs, shortTC.noise[:shortTC.sfbCnt], shortTC.isScale[:shortTC.sfbCnt], 100,
		)
	}, 25, []byte{0x55, 0x56, 0x55, 0x00})

	pnsISTC, pnsISSD := buildBitencSectionData(t, fillDynPNSIntensityCase)
	runBitencVector(t, "pns-is-scalefactor", func(bs *BitStream) int {
		return FDKaacEncEncodeScaleFactorData(
			pnsISTC.maxValue[:pnsISTC.sfbCnt], &pnsISSD, pnsISTC.scf[:pnsISTC.sfbCnt],
			bs, pnsISTC.noise[:pnsISTC.sfbCnt], pnsISTC.isScale[:pnsISTC.sfbCnt], 100,
		)
	}, 31, []byte{0x5c, 0x67, 0x79, 0xee})

	if got := FDKaacEncEncodeScaleFactorData(
		longTC.maxValue[:longTC.sfbCnt], &longSD, longTC.scf[:longTC.sfbCnt],
		nil, longTC.noise[:longTC.sfbCnt], longTC.isScale[:longTC.sfbCnt], 120,
	); got != 0 {
		t.Fatalf("nil scale-factor bit count = %d, want 0", got)
	}
}

func TestFDKaacEncEncodeMSInfoVectors(t *testing.T) {
	runBitencVector(t, "ms-none", func(bs *BitStream) int {
		return FDKaacEncEncodeMSInfo(0, 0, 0, MsMaskNone, nil, bs)
	}, 2, []byte{0x00})

	runBitencVector(t, "ms-all", func(bs *BitStream) int {
		return FDKaacEncEncodeMSInfo(0, 0, 0, MsMaskAll, nil, bs)
	}, 2, []byte{0x80})

	jsFlags := [...]int{1, 0, 1, 0, 0, 1, 0, 0}
	runBitencVector(t, "ms-some", func(bs *BitStream) int {
		return FDKaacEncEncodeMSInfo(8, 4, 3, MsMaskSome, jsFlags[:], bs)
	}, 8, []byte{0x6a})

	if got := FDKaacEncEncodeMSInfo(8, 4, 3, MsMaskSome, jsFlags[:], nil); got != 8 {
		t.Fatalf("nil MS bit count = %d, want 8", got)
	}
}

func TestFDKaacEncEncodeTnsDataPresentVectors(t *testing.T) {
	var inactive TNSInfo
	runBitencVector(t, "tns-present-inactive", func(bs *BitStream) int {
		return FDKaacEncEncodeTnsDataPresent(&inactive, LongWindow, bs)
	}, 1, []byte{0x00})

	active := longTNS4BitCase()
	runBitencVector(t, "tns-present-active", func(bs *BitStream) int {
		return FDKaacEncEncodeTnsDataPresent(&active, LongWindow, bs)
	}, 1, []byte{0x80})

	if got := FDKaacEncEncodeTnsDataPresent(&active, LongWindow, nil); got != 1 {
		t.Fatalf("nil TNS-present bit count = %d, want 1", got)
	}
	if got := FDKaacEncEncodeTnsDataPresent(nil, LongWindow, nil); got != 1 {
		t.Fatalf("nil TNS info present count = %d, want 1", got)
	}
}

func TestFDKaacEncEncodeTnsDataVectors(t *testing.T) {
	var inactive TNSInfo
	runBitencVector(t, "tns-inactive", func(bs *BitStream) int {
		return FDKaacEncEncodeTnsData(&inactive, LongWindow, bs)
	}, 0, []byte{})

	long4 := longTNS4BitCase()
	runBitencVector(t, "tns-long-4bit", func(bs *BitStream) int {
		return FDKaacEncEncodeTnsData(&long4, LongWindow, bs)
	}, 32, []byte{0x6f, 0x12, 0xcf, 0x24})

	long3 := longTNS3BitCase()
	runBitencVector(t, "tns-long-3bit", func(bs *BitStream) int {
		return FDKaacEncEncodeTnsData(&long3, LongWindow, bs)
	}, 25, []byte{0x68, 0x0d, 0x81, 0x80})

	short := shortTNSCase()
	runBitencVector(t, "tns-short", func(bs *BitStream) int {
		return FDKaacEncEncodeTnsData(&short, ShortWindow, bs)
	}, 27, []byte{0x95, 0xd8, 0xa0, 0x00})

	if got := FDKaacEncEncodeTnsData(&long4, LongWindow, nil); got != 32 {
		t.Fatalf("nil TNS data bit count = %d, want 32", got)
	}
	if got := FDKaacEncEncodeTnsData(nil, LongWindow, nil); got != 0 {
		t.Fatalf("nil TNS data count = %d, want 0", got)
	}
}

func TestFDKaacEncEncodeOneBitToolPlaceholders(t *testing.T) {
	runBitencVector(t, "pulse", FDKaacEncEncodePulseData, 1, []byte{0x00})
	runBitencVector(t, "gain-control", FDKaacEncEncodeGainControlData, 1, []byte{0x00})
}

func TestFDKaacEncWriteExtensionPayloadVectors(t *testing.T) {
	payload := [...]byte{0xab, 0xcd}
	runBitencVector(t, "extension-dynamic-range", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionPayload(bs, ExtDynamicRange, payload[:], 12)
	}, 16, []byte{0xba, 0xbc})

	ldsacPayload := [...]byte{0x0f, 0x12, 0x34}
	runBitencVector(t, "extension-ldsac", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionPayload(bs, ExtLDSACData, ldsacPayload[:], 12)
	}, 20, []byte{0x9f, 0x12, 0x30})

	dataElementPayload := [...]byte{0xde, 0xad, 0xbe}
	runBitencVector(t, "extension-data-element", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionPayload(bs, ExtDataElement, dataElementPayload[:], 20)
	}, 40, []byte{0x20, 0x03, 0xde, 0xad, 0xbe})

	runBitencVector(t, "extension-fill-data", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionPayload(bs, ExtFillData, nil, 24)
	}, 24, []byte{0x10, 0xa5, 0xa5})

	runBitencVector(t, "extension-fill", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionPayload(bs, ExtFIL, nil, 24)
	}, 24, []byte{0x00, 0x00, 0x00})

	if got := FDKaacEncWriteExtensionPayload(nil, ExtDynamicRange, payload[:], 12); got != 16 {
		t.Fatalf("nil extension-payload count = %d, want 16", got)
	}
	if got := FDKaacEncWriteExtensionPayload(nil, ExtFIL, nil, 3); got != 0 {
		t.Fatalf("too-small extension-payload count = %d, want 0", got)
	}
}

func TestFDKaacEncWriteDataStreamElementVectors(t *testing.T) {
	payload := [...]byte{0xab, 0xcd}
	runBitencVector(t, "dse-small", func(bs *BitStream) int {
		return FDKaacEncWriteDataStreamElement(bs, 3, 2, payload[:], 0)
	}, 32, []byte{0x86, 0x02, 0xab, 0xcd})

	var large [260]byte
	if got := FDKaacEncWriteDataStreamElement(nil, 0, len(large), large[:], 0); got != 2104 {
		t.Fatalf("large DSE count = %d, want 2104", got)
	}
}

func TestFDKaacEncWriteExtensionDataVectors(t *testing.T) {
	fill := QCOutExtension{Type: ExtFillData, PayloadBits: 31}
	runBitencVector(t, "ga-fill-data", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionData(bs, &fill, 0, 0, 0, 2, 0)
	}, 31, []byte{0xc6, 0x21, 0x4b, 0x4a})

	dynamicPayload := [...]byte{0xab, 0xcd}
	dynamic := QCOutExtension{Type: ExtDynamicRange, PayloadBits: 12, Payload: dynamicPayload[:]}
	runBitencVector(t, "ga-dynamic-range", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionData(bs, &dynamic, 0, 0, 0, 2, 0)
	}, 23, []byte{0xc5, 0x75, 0x78})

	runBitencVector(t, "er-dynamic-range", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionData(bs, &dynamic, 0, 0, acER, 2, 0)
	}, 16, []byte{0xba, 0xbc})

	sbr := QCOutExtension{Type: ExtSBRData, PayloadBits: 12, Payload: dynamicPayload[:]}
	runBitencVector(t, "eld-sbr-direct", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionData(bs, &sbr, 0, 0, acER|acELD, 2, 0)
	}, 12, []byte{0xab, 0xc0})

	dataElementPayload := [...]byte{0xab, 0xcd}
	dataElement := QCOutExtension{Type: ExtDataElement, PayloadBits: 16, Payload: dataElementPayload[:]}
	runBitencVector(t, "ga-data-element", func(bs *BitStream) int {
		return FDKaacEncWriteExtensionData(bs, &dataElement, 3, 0, 0, 2, 0)
	}, 32, []byte{0x86, 0x02, 0xab, 0xcd})

	if got := FDKaacEncWriteExtensionData(nil, &fill, 0, 0, 0, 2, 0); got != 31 {
		t.Fatalf("nil GA fill-data count = %d, want 31", got)
	}
}

func TestFDKaacEncBitstreamElementListVectors(t *testing.T) {
	sce := FDKaacEncGetBitstreamElementList(aotAACLC, -1, 1, 0, 0)
	if sce == nil {
		t.Fatal("nil AAC-LC SCE element list")
	}
	assertRBDSequence(t, "aac-sce", sce.ID, []rbdID{
		rbdADTSCRCStartReg1, rbdElementInstanceTag, rbdGlobalGain, rbdICSInfo,
		rbdSectionData, rbdScaleFactorData, rbdPulse, rbdTNSDataPresent,
		rbdTNSData, rbdGainControlDataPresent, rbdSpectralData,
		rbdADTSCRCEndReg1, rbdEndOfSequence,
	}, 0xb3d2c11ed7345352)

	cpe := FDKaacEncGetBitstreamElementList(aotAACLC, -1, 2, 0, 0)
	if cpe == nil || cpe.Next[0] == nil || cpe.Next[1] == nil {
		t.Fatal("incomplete AAC-LC CPE element list")
	}
	assertRBDSequence(t, "aac-cpe-root", cpe.ID, []rbdID{
		rbdADTSCRCStartReg1, rbdElementInstanceTag, rbdCommonWindow, rbdLinkSequence,
	}, 0xc297db298c27d64c)
	assertRBDSequence(t, "aac-cpe-common0", cpe.Next[0].ID, []rbdID{
		rbdGlobalGain, rbdICSInfo, rbdSectionData, rbdScaleFactorData, rbdPulse,
		rbdTNSDataPresent, rbdTNSData, rbdGainControlDataPresent, rbdSpectralData,
		rbdNextChannel, rbdADTSCRCStartReg2, rbdGlobalGain, rbdICSInfo,
		rbdSectionData, rbdScaleFactorData, rbdPulse, rbdTNSDataPresent,
		rbdTNSData, rbdGainControlDataPresent, rbdSpectralData,
		rbdADTSCRCEndReg1, rbdADTSCRCEndReg2, rbdEndOfSequence,
	}, 0xb4996b2e34bc0ddb)
	assertRBDSequence(t, "aac-cpe-common1", cpe.Next[1].ID, []rbdID{
		rbdICSInfo, rbdMS, rbdGlobalGain, rbdSectionData, rbdScaleFactorData,
		rbdPulse, rbdTNSDataPresent, rbdTNSData, rbdGainControlDataPresent,
		rbdSpectralData, rbdNextChannel, rbdADTSCRCStartReg2, rbdGlobalGain,
		rbdSectionData, rbdScaleFactorData, rbdPulse, rbdTNSDataPresent,
		rbdTNSData, rbdGainControlDataPresent, rbdSpectralData,
		rbdADTSCRCEndReg1, rbdADTSCRCEndReg2, rbdEndOfSequence,
	}, 0xe5e42f6ce8b8b2d)

	if got := FDKaacEncGetBitstreamElementList(99, -1, 1, 0, 0); got != nil {
		t.Fatalf("unsupported AOT list = %#v, want nil", got)
	}
	if got := FDKaacEncGetBitstreamElementList(aotAACLC, 0, 1, 0, 0); got != nil {
		t.Fatalf("unsupported epConfig list = %#v, want nil", got)
	}
}

func TestFDKaacEncChannelElementWriteSCEVector(t *testing.T) {
	qc, psy := buildLongChannelElementCase(t)
	element := ElementInfo{ElType: idSCE, InstanceTag: 2}
	psyElement := PsyOutElement{}
	runBitencVector(t, "sce-long-element", func(bs *BitStream) int {
		bits, errCode := FDKaacEncChannelElementWrite(
			bs, &element, []*QCOutChannel{qc}, &psyElement, []*PsyOutChannel{psy},
			0, aotAACLC, -1, 0,
		)
		if errCode != AACEncOK {
			t.Fatalf("SCE write error = %#x, want OK", errCode)
		}
		return bits
	}, 202, []byte{
		0x05, 0x18, 0x24, 0x04, 0x76, 0x55, 0x55, 0x55,
		0x55, 0x06, 0xcf, 0x72, 0xf3, 0x8e, 0xa9, 0xc6,
		0x2d, 0x8c, 0xfc, 0x97, 0xe6, 0x12, 0x78, 0x38,
		0xe0, 0x40,
	})
}

func TestFDKaacEncChannelElementWriteStaticDryRunVector(t *testing.T) {
	_, psy := buildLongChannelElementCase(t)
	element := ElementInfo{ElType: idSCE, InstanceTag: 2}
	psyElement := PsyOutElement{}
	bits, errCode := FDKaacEncChannelElementWrite(
		nil, &element, nil, &psyElement, []*PsyOutChannel{psy},
		0, aotAACLC, -1, 0,
	)
	if errCode != AACEncOK {
		t.Fatalf("static dry-run error = %#x, want OK", errCode)
	}
	if bits != 29 {
		t.Fatalf("static dry-run bits = %d, want 29", bits)
	}
}

func TestFDKaacEncChannelElementWriteCPECommonWindowVector(t *testing.T) {
	qc0, psy0 := buildLongChannelElementCase(t)
	qc1, psy1 := buildLongChannelElementCase(t)
	element := ElementInfo{ElType: idCPE, InstanceTag: 5}
	psyElement := PsyOutElement{CommonWindow: 1, ToolsInfo: ToolsInfo{MsDigest: MsMaskSome}}
	copy(psyElement.ToolsInfo.MsMask[:], []int{1, 0, 1, 0, 0, 1, 0, 0})

	runBitencVector(t, "cpe-common-long-element", func(bs *BitStream) int {
		bits, errCode := FDKaacEncChannelElementWrite(
			bs, &element, []*QCOutChannel{qc0, qc1}, &psyElement, []*PsyOutChannel{psy0, psy1},
			0, aotAACLC, -1, 0,
		)
		if errCode != AACEncOK {
			t.Fatalf("CPE write error = %#x, want OK", errCode)
		}
		return bits
	}, 397, []byte{
		0x2b, 0x12, 0x0d, 0x24, 0x60, 0x8e, 0xca, 0xaa,
		0xaa, 0xaa, 0xa0, 0xd9, 0xee, 0x5e, 0x71, 0xd5,
		0x38, 0xc5, 0xb1, 0x9f, 0x92, 0xfc, 0xc2, 0x4f,
		0x07, 0x1c, 0x0c, 0x60, 0x8e, 0xca, 0xaa, 0xaa,
		0xaa, 0xa0, 0xd9, 0xee, 0x5e, 0x71, 0xd5, 0x38,
		0xc5, 0xb1, 0x9f, 0x92, 0xfc, 0xc2, 0x4f, 0x07,
		0x1c, 0x08,
	})
}

func TestFDKaacEncChannelElementWriteErrors(t *testing.T) {
	qc, psy := buildLongChannelElementCase(t)
	element := ElementInfo{ElType: idSCE, InstanceTag: 0}
	psyElement := PsyOutElement{}
	var storage [256]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}

	badSection := *qc
	badSection.SectionData.SideInfoBits++
	if _, errCode := FDKaacEncChannelElementWrite(&bs, &element, []*QCOutChannel{&badSection}, &psyElement, []*PsyOutChannel{psy}, 0, aotAACLC, -1, 0); errCode != AACEncWriteSecError {
		t.Fatalf("section mismatch error = %#x, want %#x", errCode, AACEncWriteSecError)
	}

	ResetBitStream(&bs, BSWriter)
	badScale := *qc
	badScale.SectionData.ScalefacBits++
	if _, errCode := FDKaacEncChannelElementWrite(&bs, &element, []*QCOutChannel{&badScale}, &psyElement, []*PsyOutChannel{psy}, 0, aotAACLC, -1, 0); errCode != AACEncWriteScalError {
		t.Fatalf("scale mismatch error = %#x, want %#x", errCode, AACEncWriteScalError)
	}

	ResetBitStream(&bs, BSWriter)
	badSpectral := *qc
	badSpectral.SectionData.HuffmanBits++
	if _, errCode := FDKaacEncChannelElementWrite(&bs, &element, []*QCOutChannel{&badSpectral}, &psyElement, []*PsyOutChannel{psy}, 0, aotAACLC, -1, 0); errCode != AACEncWriteSpecError {
		t.Fatalf("spectral mismatch error = %#x, want %#x", errCode, AACEncWriteSpecError)
	}
}

func TestFDKaacEncEncodeSpectralDataVectors(t *testing.T) {
	longTC, longSD := buildBitencSectionData(t, fillDynLongCase)
	runBitencVector(t, "long-spectral", func(bs *BitStream) int {
		return FDKaacEncEncodeSpectralData(longTC.offsets[:longTC.sfbCnt+1], &longSD, longTC.quant[:], bs)
	}, 126, []byte{0x6c, 0xf7, 0x2f, 0x38, 0xea, 0x9c, 0x62, 0xd8, 0xcf, 0xc9, 0x7e, 0x61, 0x27, 0x83, 0x8e, 0x04})

	shortTC, shortSD := buildBitencSectionData(t, fillDynShortCase)
	runBitencVector(t, "short-spectral", func(bs *BitStream) int {
		return FDKaacEncEncodeSpectralData(shortTC.offsets[:shortTC.sfbCnt+1], &shortSD, shortTC.quant[:], bs)
	}, 94, []byte{0x9a, 0x5f, 0xce, 0x7f, 0x12, 0x3a, 0xfb, 0x8d, 0x69, 0xbc, 0x7c, 0xa4})
}

func TestFDKaacEncEncodeChannelDataSequenceVector(t *testing.T) {
	tc, sectionData := buildBitencSectionData(t, fillDynLongCase)
	runBitencVector(t, "long-channel-data", func(bs *BitStream) int {
		bits := 0
		bits += FDKaacEncEncodeGlobalGain(80, 100, bs, 0)
		bits += FDKaacEncEncodeIcsInfo(LongWindow, WindowShapeKBD, 0, 8, bs, 0)
		bits += FDKaacEncEncodeSectionData(&sectionData, bs, false)
		bits += FDKaacEncEncodeScaleFactorData(
			tc.maxValue[:tc.sfbCnt], &sectionData, tc.scf[:tc.sfbCnt],
			bs, tc.noise[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 120,
		)
		bits += FDKaacEncEncodeSpectralData(tc.offsets[:tc.sfbCnt+1], &sectionData, tc.quant[:], bs)
		return bits
	}, 192, []byte{
		0x8c, 0x12, 0x02, 0x3b, 0x2a, 0xaa, 0xaa, 0xaa,
		0x9b, 0x3d, 0xcb, 0xce, 0x3a, 0xa7, 0x18, 0xb6,
		0x33, 0xf2, 0x5f, 0x98, 0x49, 0xe0, 0xe3, 0x81,
	})
}

func TestFDKaacEncBitencRejectsInvalid(t *testing.T) {
	tc, sectionData := buildBitencSectionData(t, fillDynLongCase)
	var storage [64]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"bad ICS block type", func() { FDKaacEncEncodeIcsInfo(-1, WindowShapeKBD, 0, 8, &bs, 0) }},
		{"bad ICS window shape", func() { FDKaacEncEncodeIcsInfo(LongWindow, -1, 0, 8, &bs, 0) }},
		{"bad ICS grouping mask", func() { FDKaacEncEncodeIcsInfo(ShortWindow, WindowShapeSine, 0x80, 4, &bs, 0) }},
		{"nil spectral bitstream", func() {
			FDKaacEncEncodeSpectralData(tc.offsets[:tc.sfbCnt+1], &sectionData, tc.quant[:], nil)
		}},
		{"nil spectral section data", func() {
			FDKaacEncEncodeSpectralData(tc.offsets[:tc.sfbCnt+1], nil, tc.quant[:], &bs)
		}},
		{"short spectral offsets", func() {
			FDKaacEncEncodeSpectralData(tc.offsets[:tc.sfbCnt], &sectionData, tc.quant[:], &bs)
		}},
		{"malformed spectral offsets", func() {
			bad := tc
			bad.offsets[2] = bad.offsets[1] - 1
			FDKaacEncEncodeSpectralData(bad.offsets[:bad.sfbCnt+1], &sectionData, bad.quant[:], &bs)
		}},
		{"short spectral values", func() {
			FDKaacEncEncodeSpectralData(tc.offsets[:tc.sfbCnt+1], &sectionData, tc.quant[:tc.offsets[tc.sfbCnt]-1], &bs)
		}},
		{"nil section data", func() { FDKaacEncEncodeSectionData(nil, &bs, false) }},
		{"invalid section codebook", func() {
			bad := sectionData
			bad.Huffsection[0].CodeBook = codeBookISInPhaseNo + 1
			FDKaacEncEncodeSectionData(&bad, &bs, false)
		}},
		{"nil scalefactor section data", func() {
			FDKaacEncEncodeScaleFactorData(tc.maxValue[:tc.sfbCnt], nil, tc.scf[:tc.sfbCnt], &bs, tc.noise[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 120)
		}},
		{"short scalefactors", func() {
			FDKaacEncEncodeScaleFactorData(tc.maxValue[:tc.sfbCnt], &sectionData, tc.scf[:tc.sfbCnt-1], &bs, tc.noise[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 120)
		}},
		{"invalid first scalefactor", func() {
			bad := sectionData
			bad.FirstScf = tc.sfbCnt
			FDKaacEncEncodeScaleFactorData(tc.maxValue[:tc.sfbCnt], &bad, tc.scf[:tc.sfbCnt], &bs, tc.noise[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 120)
		}},
		{"invalid MS digest", func() { FDKaacEncEncodeMSInfo(0, 0, 0, -1, nil, &bs) }},
		{"invalid MS sfb count", func() { FDKaacEncEncodeMSInfo(-1, 0, 0, MsMaskNone, nil, &bs) }},
		{"invalid MS group", func() {
			jsFlags := [...]int{1, 0, 1, 0}
			FDKaacEncEncodeMSInfo(4, 3, 2, MsMaskSome, jsFlags[:], &bs)
		}},
		{"invalid MS max sfb", func() {
			jsFlags := [...]int{1, 0, 1, 0}
			FDKaacEncEncodeMSInfo(4, 4, 5, MsMaskSome, jsFlags[:], &bs)
		}},
		{"short MS mask", func() {
			jsFlags := [...]int{1, 0, 1}
			FDKaacEncEncodeMSInfo(4, 4, 2, MsMaskSome, jsFlags[:], &bs)
		}},
		{"invalid TNS block type", func() {
			tns := longTNS4BitCase()
			FDKaacEncEncodeTnsData(&tns, -1, &bs)
		}},
		{"invalid TNS filter count", func() {
			tns := longTNS4BitCase()
			tns.NumOfFilters[0] = maxTnsFilters + 1
			FDKaacEncEncodeTnsData(&tns, LongWindow, &bs)
		}},
		{"invalid TNS coef resolution", func() {
			tns := longTNS4BitCase()
			tns.CoefRes[0] = 2
			FDKaacEncEncodeTnsData(&tns, LongWindow, &bs)
		}},
		{"invalid TNS length", func() {
			tns := longTNS4BitCase()
			tns.Length[0][0] = 64
			FDKaacEncEncodeTnsData(&tns, LongWindow, &bs)
		}},
		{"invalid short TNS order", func() {
			tns := shortTNSCase()
			tns.Order[0][0] = 8
			FDKaacEncEncodeTnsData(&tns, ShortWindow, &bs)
		}},
		{"invalid TNS direction", func() {
			tns := longTNS4BitCase()
			tns.Direction[0][0] = 2
			FDKaacEncEncodeTnsData(&tns, LongWindow, &bs)
		}},
		{"invalid TNS coefficient", func() {
			tns := longTNS4BitCase()
			tns.Coef[0][0][0] = 8
			FDKaacEncEncodeTnsData(&tns, LongWindow, &bs)
		}},
		{"invalid extension type", func() {
			FDKaacEncWriteExtensionPayload(&bs, 16, nil, 8)
		}},
		{"negative extension bits", func() {
			FDKaacEncWriteExtensionPayload(&bs, ExtFIL, nil, -1)
		}},
		{"short extension payload", func() {
			payload := [...]byte{0xab}
			FDKaacEncWriteExtensionPayload(&bs, ExtDynamicRange, payload[:], 12)
		}},
		{"invalid DSE tag", func() {
			payload := [...]byte{0xab}
			FDKaacEncWriteDataStreamElement(&bs, 16, 1, payload[:], 0)
		}},
		{"negative DSE length", func() {
			FDKaacEncWriteDataStreamElement(&bs, 0, -1, nil, 0)
		}},
		{"short DSE payload", func() {
			payload := [...]byte{0xab}
			FDKaacEncWriteDataStreamElement(&bs, 0, 2, payload[:], 0)
		}},
		{"nil extension data", func() {
			FDKaacEncWriteExtensionData(&bs, nil, 0, 0, 0, 2, 0)
		}},
		{"short data-element extension", func() {
			payload := [...]byte{0xab}
			extension := QCOutExtension{Type: ExtDataElement, PayloadBits: 16, Payload: payload[:]}
			FDKaacEncWriteExtensionData(&bs, &extension, 0, 0, 0, 2, 0)
		}},
		{"nil element info", func() {
			FDKaacEncChannelElementWrite(&bs, nil, nil, &PsyOutElement{}, nil, 0, aotAACLC, -1, 0)
		}},
		{"invalid element type", func() {
			qc, psy := buildLongChannelElementCase(t)
			element := ElementInfo{ElType: idDSE, InstanceTag: 0}
			FDKaacEncChannelElementWrite(&bs, &element, []*QCOutChannel{qc}, &PsyOutElement{}, []*PsyOutChannel{psy}, 0, aotAACLC, -1, 0)
		}},
		{"invalid element tag", func() {
			qc, psy := buildLongChannelElementCase(t)
			element := ElementInfo{ElType: idSCE, InstanceTag: 16}
			FDKaacEncChannelElementWrite(&bs, &element, []*QCOutChannel{qc}, &PsyOutElement{}, []*PsyOutChannel{psy}, 0, aotAACLC, -1, 0)
		}},
		{"invalid common window", func() {
			qc0, psy0 := buildLongChannelElementCase(t)
			qc1, psy1 := buildLongChannelElementCase(t)
			element := ElementInfo{ElType: idCPE, InstanceTag: 0}
			psyElement := PsyOutElement{CommonWindow: 2}
			FDKaacEncChannelElementWrite(&bs, &element, []*QCOutChannel{qc0, qc1}, &psyElement, []*PsyOutChannel{psy0, psy1}, 0, aotAACLC, -1, 0)
		}},
		{"unsupported channel sequence", func() {
			qc, psy := buildLongChannelElementCase(t)
			element := ElementInfo{ElType: idSCE, InstanceTag: 0}
			FDKaacEncChannelElementWrite(&bs, &element, []*QCOutChannel{qc}, &PsyOutElement{}, []*PsyOutChannel{psy}, 0, 99, -1, 0)
		}},
		{"short channel inputs", func() {
			qc0, psy0 := buildLongChannelElementCase(t)
			element := ElementInfo{ElType: idCPE, InstanceTag: 0}
			FDKaacEncChannelElementWrite(&bs, &element, []*QCOutChannel{qc0}, &PsyOutElement{}, []*PsyOutChannel{psy0}, 0, aotAACLC, -1, 0)
		}},
		{"nil QC with bitstream", func() {
			_, psy := buildLongChannelElementCase(t)
			element := ElementInfo{ElType: idSCE, InstanceTag: 0}
			FDKaacEncChannelElementWrite(&bs, &element, nil, &PsyOutElement{}, []*PsyOutChannel{psy}, 0, aotAACLC, -1, 0)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			ResetBitStream(&bs, BSWriter)
			tt.fn()
		})
	}
}

func TestFDKaacEncChannelElementWriteAllocs(t *testing.T) {
	qc, psy := buildLongChannelElementCase(t)
	element := ElementInfo{ElType: idSCE, InstanceTag: 2}
	psyElement := PsyOutElement{}
	qcChannels := []*QCOutChannel{qc}
	psyChannels := []*PsyOutChannel{psy}
	var storage [256]byte
	var out [256]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		clear(storage[:])
		clear(out[:])
		ResetBitStream(&bs, BSWriter)
		bits, errCode := FDKaacEncChannelElementWrite(&bs, &element, qcChannels, &psyElement, psyChannels, 0, aotAACLC, -1, 0)
		if errCode != AACEncOK {
			t.Fatalf("channel element write error = %#x", errCode)
		}
		ByteAlign(&bs, 0)
		n := FetchBuffer(&bs, out[:])
		bitCountSink = bits + n
		bitCountHashSink = hashHuffBytes(out[:n])
	})
	if allocs != 0 {
		t.Fatalf("channel element writer allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncExtensionWritersAllocs(t *testing.T) {
	payload := [...]byte{0xab, 0xcd}
	dynamic := QCOutExtension{Type: ExtDynamicRange, PayloadBits: 12, Payload: payload[:]}
	fill := QCOutExtension{Type: ExtFillData, PayloadBits: 31}
	dataElement := QCOutExtension{Type: ExtDataElement, PayloadBits: 16, Payload: payload[:]}
	var storage [128]byte
	var out [128]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		clear(storage[:])
		clear(out[:])
		ResetBitStream(&bs, BSWriter)
		FDKaacEncWriteExtensionPayload(&bs, ExtDynamicRange, payload[:], 12)
		FDKaacEncWriteExtensionData(&bs, &dynamic, 0, 0, 0, 2, 0)
		FDKaacEncWriteExtensionData(&bs, &fill, 0, 0, 0, 2, 0)
		FDKaacEncWriteExtensionData(&bs, &dataElement, 3, 0, 0, 2, 0)
		ByteAlign(&bs, 0)
		n := FetchBuffer(&bs, out[:])
		bitCountSink = n
		bitCountHashSink = hashHuffBytes(out[:n])
	})
	if allocs != 0 {
		t.Fatalf("extension writer allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncSideInfoWritersAllocs(t *testing.T) {
	jsFlags := [...]int{1, 0, 1, 0, 0, 1, 0, 0}
	tns := longTNS4BitCase()
	var storage [64]byte
	var out [64]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		clear(storage[:])
		clear(out[:])
		ResetBitStream(&bs, BSWriter)
		FDKaacEncEncodeMSInfo(8, 4, 3, MsMaskSome, jsFlags[:], &bs)
		FDKaacEncEncodeTnsDataPresent(&tns, LongWindow, &bs)
		FDKaacEncEncodeTnsData(&tns, LongWindow, &bs)
		FDKaacEncEncodePulseData(&bs)
		FDKaacEncEncodeGainControlData(&bs)
		ByteAlign(&bs, 0)
		n := FetchBuffer(&bs, out[:])
		bitCountSink = n
		bitCountHashSink = hashHuffBytes(out[:n])
	})
	if allocs != 0 {
		t.Fatalf("side-info writer allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncBitencAllocs(t *testing.T) {
	tc, sectionData := buildBitencSectionData(t, fillDynLongCase)
	var storage [256]byte
	var out [256]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		clear(storage[:])
		clear(out[:])
		ResetBitStream(&bs, BSWriter)
		FDKaacEncEncodeGlobalGain(80, 100, &bs, 0)
		FDKaacEncEncodeIcsInfo(LongWindow, WindowShapeKBD, 0, 8, &bs, 0)
		FDKaacEncEncodeSectionData(&sectionData, &bs, false)
		FDKaacEncEncodeScaleFactorData(
			tc.maxValue[:tc.sfbCnt], &sectionData, tc.scf[:tc.sfbCnt],
			&bs, tc.noise[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 120,
		)
		FDKaacEncEncodeSpectralData(tc.offsets[:tc.sfbCnt+1], &sectionData, tc.quant[:], &bs)
		ByteAlign(&bs, 0)
		n := FetchBuffer(&bs, out[:])
		bitCountSink = n
		bitCountHashSink = hashHuffBytes(out[:n])
	})
	if allocs != 0 {
		t.Fatalf("bitstream helper allocations = %v, want 0", allocs)
	}
}

func longTNS4BitCase() TNSInfo {
	var tns TNSInfo
	tns.NumOfFilters[0] = 1
	tns.CoefRes[0] = 4
	tns.Length[0][0] = 30
	tns.Order[0][0] = 4
	tns.Direction[0][0] = 1
	copy(tns.Coef[0][0][:], []int{-4, -1, 2, 4})
	return tns
}

func longTNS3BitCase() TNSInfo {
	var tns TNSInfo
	tns.NumOfFilters[0] = 1
	tns.CoefRes[0] = 4
	tns.Length[0][0] = 16
	tns.Order[0][0] = 3
	tns.Direction[0][0] = 0
	copy(tns.Coef[0][0][:], []int{-4, 0, 3})
	return tns
}

func shortTNSCase() TNSInfo {
	var tns TNSInfo
	tns.NumOfFilters[0] = 1
	tns.CoefRes[0] = 3
	tns.Length[0][0] = 5
	tns.Order[0][0] = 3
	tns.Direction[0][0] = 1
	copy(tns.Coef[0][0][:], []int{-2, 1, 2})
	return tns
}

func buildLongChannelElementCase(t *testing.T) (*QCOutChannel, *PsyOutChannel) {
	t.Helper()
	tc, sectionData := buildBitencSectionData(t, fillDynLongCase)
	qc := &QCOutChannel{GlobalGain: 80, SectionData: sectionData}
	copy(qc.QuantSpec[:], tc.quant[:])
	copy(qc.MaxValueInSfb[:], tc.maxValue[:])
	copy(qc.Scf[:], tc.scf[:])

	psy := &PsyOutChannel{
		SfbCnt:             tc.sfbCnt,
		SfbPerGroup:        tc.sfbPerGroup,
		MaxSfbPerGroup:     tc.maxSfbPerGroup,
		LastWindowSequence: tc.blockType,
		WindowShape:        WindowShapeKBD,
		GroupingMask:       0,
		MdctScale:          0,
	}
	copy(psy.SfbOffsets[:], tc.offsets[:])
	copy(psy.NoiseNrg[:], tc.noise[:])
	copy(psy.IsScale[:], tc.isScale[:])
	return qc, psy
}

func assertRBDSequence(t *testing.T, name string, got []rbdID, want []rbdID, hash uint64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	var gotInts [64]int
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %d, want %d", name, i, got[i], want[i])
		}
		gotInts[i] = int(got[i])
	}
	if h := hashBandEnergyInts(gotInts[:len(want)]); h != hash {
		t.Fatalf("%s hash = %#016x, want %#016x", name, h, hash)
	}
}

func buildBitencSectionData(t *testing.T, fill func(*dynBitCountCase)) (dynBitCountCase, SectionData) {
	t.Helper()
	var tc dynBitCountCase
	fill(&tc)
	var state BitCounterState
	var sectionData SectionData
	FDKaacEncDynBitCount(
		&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt],
		tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup,
		tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt],
		tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0,
	)
	return tc, sectionData
}

func runBitencVector(t *testing.T, name string, encode func(*BitStream) int, wantBits int, wantBytes []byte) {
	t.Helper()
	var storage [256]byte
	var out [256]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}
	gotBits := encode(&bs)
	if gotBits != wantBits {
		t.Fatalf("%s returned bits = %d, want %d", name, gotBits, wantBits)
	}
	if validBits := int(BitStreamValidBits(&bs)); validBits != wantBits {
		t.Fatalf("%s valid bits = %d, want %d", name, validBits, wantBits)
	}
	ByteAlign(&bs, 0)
	n := FetchBuffer(&bs, out[:])
	if !bytes.Equal(out[:n], wantBytes) {
		t.Fatalf("%s bytes = % x, want % x", name, out[:n], wantBytes)
	}
}
