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
