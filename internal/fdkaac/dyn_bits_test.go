package fdkaac

import "testing"

func TestFDKaacEncSideInfoTables(t *testing.T) {
	if got, want := hashBandEnergyInts(fdkaacEncSideInfoTabLong[:]), uint64(0xf333f0b1a056e532); got != want {
		t.Fatalf("long side-info table hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashBandEnergyInts(fdkaacEncSideInfoTabShort[:]), uint64(0xe7e7bc982540ebc8); got != want {
		t.Fatalf("short side-info table hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncDynBitCountLongVector(t *testing.T) {
	var tc dynBitCountCase
	fillDynLongCase(&tc)

	var state BitCounterState
	var sectionData SectionData
	got := FDKaacEncDynBitCount(
		&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt],
		tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup,
		tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt],
		tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0,
	)
	assertSectionData(t, "long", got, &sectionData, dynTotals{
		total: 173, huffman: 126, sideInfo: 18, scalefac: 29, noise: 0, firstScf: 0,
		sections: 2, totalHash: 0x14312f4880e8007b, sectionHash: 0xa9d04f00f65e51d6,
	})
}

func TestFDKaacEncDynBitCountVCB11Vector(t *testing.T) {
	var tc dynBitCountCase
	fillDynLongCase(&tc)

	var state BitCounterState
	var sectionData SectionData
	got := FDKaacEncDynBitCount(
		&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt],
		tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup,
		tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt],
		tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], acErVCB11,
	)
	assertSectionData(t, "long-vcb11", got, &sectionData, dynTotals{
		total: 173, huffman: 130, sideInfo: 14, scalefac: 29, noise: 0, firstScf: 0,
		sections: 2, totalHash: 0xb363368d2fc7f9db, sectionHash: 0xa9d04f00f65e51d6,
	})
}

func TestFDKaacEncDynBitCountShortGroupedVector(t *testing.T) {
	var tc dynBitCountCase
	fillDynShortCase(&tc)

	var state BitCounterState
	var sectionData SectionData
	got := FDKaacEncDynBitCount(
		&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt],
		tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup,
		tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt],
		tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0,
	)
	assertSectionData(t, "short", got, &sectionData, dynTotals{
		total: 140, huffman: 94, sideInfo: 21, scalefac: 25, noise: 0, firstScf: 0,
		sections: 3, totalHash: 0x0eb540911e804b08, sectionHash: 0x6ea1c9e13ba4b539,
	})
}

func TestFDKaacEncDynBitCountPNSAndIntensityVector(t *testing.T) {
	var tc dynBitCountCase
	fillDynPNSIntensityCase(&tc)

	var state BitCounterState
	var sectionData SectionData
	got := FDKaacEncDynBitCount(
		&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt],
		tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup,
		tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt],
		tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0,
	)
	assertSectionData(t, "pns-is", got, &sectionData, dynTotals{
		total: 82, huffman: 6, sideInfo: 45, scalefac: 18, noise: 13, firstScf: 1,
		sections: 5, totalHash: 0x9132989ca8e4c127, sectionHash: 0xd311b27cd88cb0bc,
	})
}

func TestFDKaacEncDynBitCountRejectsInvalid(t *testing.T) {
	var tc dynBitCountCase
	fillDynLongCase(&tc)
	var state BitCounterState
	var sectionData SectionData

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil state", func() {
			FDKaacEncDynBitCount(nil, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt], tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup, tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt], tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0)
		}},
		{"nil section data", func() {
			FDKaacEncDynBitCount(&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt], tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup, tc.offsets[:tc.sfbCnt+1], nil, tc.noise[:tc.sfbCnt], tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0)
		}},
		{"bad group shape", func() {
			FDKaacEncDynBitCount(&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt], tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, 3, tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt], tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0)
		}},
		{"max larger than group", func() {
			FDKaacEncDynBitCount(&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt], tc.blockType, tc.sfbCnt, tc.sfbPerGroup+1, tc.sfbPerGroup, tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt], tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0)
		}},
		{"short scalefactors", func() {
			FDKaacEncDynBitCount(&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt-1], tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup, tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt], tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0)
		}},
		{"short offsets", func() {
			FDKaacEncDynBitCount(&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt], tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup, tc.offsets[:tc.sfbCnt], &sectionData, tc.noise[:tc.sfbCnt], tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0)
		}},
		{"malformed offsets", func() {
			bad := tc
			bad.offsets[3] = bad.offsets[2] - 1
			FDKaacEncDynBitCount(&state, bad.quant[:], bad.maxValue[:bad.sfbCnt], bad.scf[:bad.sfbCnt], bad.blockType, bad.sfbCnt, bad.maxSfbPerGroup, bad.sfbPerGroup, bad.offsets[:bad.sfbCnt+1], &sectionData, bad.noise[:bad.sfbCnt], bad.isBook[:bad.sfbCnt], bad.isScale[:bad.sfbCnt], 0)
		}},
		{"short spectrum", func() {
			FDKaacEncDynBitCount(&state, tc.quant[:tc.offsets[tc.sfbCnt]-1], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt], tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup, tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt], tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0)
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

func TestFDKaacEncDynBitCountAllocs(t *testing.T) {
	var tc dynBitCountCase
	fillDynLongCase(&tc)
	var state BitCounterState
	var sectionData SectionData
	var flat [maxSections * 4]int

	allocs := testing.AllocsPerRun(1000, func() {
		got := FDKaacEncDynBitCount(
			&state, tc.quant[:], tc.maxValue[:tc.sfbCnt], tc.scf[:tc.sfbCnt],
			tc.blockType, tc.sfbCnt, tc.maxSfbPerGroup, tc.sfbPerGroup,
			tc.offsets[:tc.sfbCnt+1], &sectionData, tc.noise[:tc.sfbCnt],
			tc.isBook[:tc.sfbCnt], tc.isScale[:tc.sfbCnt], 0,
		)
		n := flattenSectionData(&sectionData, flat[:])
		bitCountSink = got + n
		bitCountHashSink = hashBandEnergyInts(flat[:n])
	})
	if allocs != 0 {
		t.Fatalf("dynamic bit-count allocations = %v, want 0", allocs)
	}
}

type dynBitCountCase struct {
	blockType      int
	sfbCnt         int
	maxSfbPerGroup int
	sfbPerGroup    int
	quant          [maxSpectralLines]int16
	offsets        [maxGroupedSFB + 1]int
	maxValue       [maxGroupedSFB]uint32
	scf            [maxGroupedSFB]int
	noise          [maxGroupedSFB]int
	isBook         [maxGroupedSFB]int
	isScale        [maxGroupedSFB]int
}

type dynTotals struct {
	total       int
	huffman     int
	sideInfo    int
	scalefac    int
	noise       int
	firstScf    int
	sections    int
	totalHash   uint64
	sectionHash uint64
}

func fillDynLongCase(tc *dynBitCountCase) {
	*tc = dynBitCountCase{blockType: LongWindow, sfbCnt: 8, maxSfbPerGroup: 8, sfbPerGroup: 8}
	values := [...]int16{
		0, 0, 0, 0, -1, 0, 1, 0, 1, -1, 0, 1, -2, 0, 2, 1,
		-4, 3, 0, 4, -7, 6, 0, 5, -12, 11, 0, 10, -31, 16, 0, 17,
	}
	copy(tc.quant[:], values[:])
	fillDynOffsetsAndMax(tc, 4)
	copy(tc.scf[:], []int{100, 99, 98, 97, 96, 95, 94, 93})
	for i := 0; i < tc.sfbCnt; i++ {
		tc.noise[i] = noNoisePNS
	}
}

func fillDynShortCase(tc *dynBitCountCase) {
	*tc = dynBitCountCase{blockType: ShortWindow, sfbCnt: 8, maxSfbPerGroup: 4, sfbPerGroup: 4}
	values := [...]int16{
		-1, 0, 1, 0, 1, 0, -1, 0, -2, 0, 2, 1, -2, 1, 0, 2,
		0, 0, 0, 0, -4, 3, 0, 4, -7, 6, 0, 5, -7, 0, 6, -5,
	}
	copy(tc.quant[:], values[:])
	fillDynOffsetsAndMax(tc, 4)
	copy(tc.scf[:], []int{80, 79, 78, 77, 76, 75, 74, 73})
	for i := 0; i < tc.sfbCnt; i++ {
		tc.noise[i] = noNoisePNS
	}
}

func fillDynPNSIntensityCase(tc *dynBitCountCase) {
	*tc = dynBitCountCase{blockType: LongWindow, sfbCnt: 6, maxSfbPerGroup: 6, sfbPerGroup: 6}
	values := [...]int16{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, -1, 0, 1, 0,
	}
	copy(tc.quant[:], values[:])
	fillDynOffsetsAndMax(tc, 4)
	copy(tc.scf[:], []int{90, 88, 87, 86, 85, 80})
	for i := 0; i < tc.sfbCnt; i++ {
		tc.noise[i] = noNoisePNS
	}
	tc.noise[1] = 10
	tc.noise[2] = 12
	tc.isBook[3] = codeBookISOutOfPhaseNo
	tc.isBook[4] = codeBookISInPhaseNo
	tc.isScale[3] = 5
	tc.isScale[4] = 7
}

func fillDynOffsetsAndMax(tc *dynBitCountCase, width int) {
	for i := 0; i <= tc.sfbCnt; i++ {
		tc.offsets[i] = i * width
	}
	for i := 0; i < tc.sfbCnt; i++ {
		maxValue := 0
		for j := tc.offsets[i]; j < tc.offsets[i+1]; j++ {
			maxValue = maxInt(maxValue, absInt16(tc.quant[j]))
		}
		tc.maxValue[i] = uint32(maxValue)
	}
}

func assertSectionData(t *testing.T, name string, got int, sectionData *SectionData, want dynTotals) {
	t.Helper()
	if got != want.total {
		t.Fatalf("%s total = %d, want %d", name, got, want.total)
	}
	if sectionData.HuffmanBits != want.huffman || sectionData.SideInfoBits != want.sideInfo ||
		sectionData.ScalefacBits != want.scalefac || sectionData.NoiseNrgBits != want.noise ||
		sectionData.FirstScf != want.firstScf || sectionData.NoOfSections != want.sections {
		t.Fatalf("%s state = huff:%d side:%d scf:%d noise:%d first:%d sections:%d, want huff:%d side:%d scf:%d noise:%d first:%d sections:%d",
			name, sectionData.HuffmanBits, sectionData.SideInfoBits, sectionData.ScalefacBits,
			sectionData.NoiseNrgBits, sectionData.FirstScf, sectionData.NoOfSections,
			want.huffman, want.sideInfo, want.scalefac, want.noise, want.firstScf, want.sections)
	}
	totalVector := [...]int{
		got, sectionData.HuffmanBits, sectionData.SideInfoBits, sectionData.ScalefacBits,
		sectionData.NoiseNrgBits, sectionData.FirstScf, sectionData.NoOfSections,
	}
	if h := hashBandEnergyInts(totalVector[:]); h != want.totalHash {
		t.Fatalf("%s total hash = %#016x, want %#016x", name, h, want.totalHash)
	}
	var flat [maxSections * 4]int
	n := flattenSectionData(sectionData, flat[:])
	if h := hashBandEnergyInts(flat[:n]); h != want.sectionHash {
		t.Fatalf("%s section hash = %#016x, want %#016x; flat=%v", name, h, want.sectionHash, flat[:n])
	}
}

func flattenSectionData(sectionData *SectionData, dst []int) int {
	n := 0
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		dst[n+0] = section.CodeBook
		dst[n+1] = section.SfbStart
		dst[n+2] = section.SfbCnt
		dst[n+3] = section.SectionBits
		n += 4
	}
	return n
}
