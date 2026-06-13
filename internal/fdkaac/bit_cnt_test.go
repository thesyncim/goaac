package fdkaac

import "testing"

var bitCountSink int
var bitCountHashSink uint64

func TestFDKaacEncBitCountScalefactorDeltaVectors(t *testing.T) {
	if len(fdkaacEncHuffLtabScf) != 121 {
		t.Fatalf("len(fdkaacEncHuffLtabScf) = %d, want 121", len(fdkaacEncHuffLtabScf))
	}
	var tableInts [len(fdkaacEncHuffLtabScf)]int
	for i, v := range fdkaacEncHuffLtabScf {
		tableInts[i] = int(v)
	}
	if h := hashBandEnergyInts(tableInts[:]); h != 0x46152da3c54f4792 {
		t.Fatalf("scalefactor length table hash = %#016x, want %#016x", h, uint64(0x46152da3c54f4792))
	}

	deltas := [...]int{-60, -41, -20, -1, 0, 1, 2, 23, 30, 60}
	want := [...]int{18, 18, 12, 3, 1, 4, 4, 13, 18, 19}
	var got [len(deltas)]int
	for i, delta := range deltas {
		got[i] = FDKaacEncBitCountScalefactorDelta(delta)
	}
	assertIntSlice(t, "scalefactor delta bit counts", got[:], want[:], 0xd2bddbfbeedfcd17)
}

func TestFDKaacEncBitCountScalefactorDeltaRejectsInvalid(t *testing.T) {
	for _, delta := range []int{-61, 61} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("delta %d did not panic", delta)
				}
			}()
			FDKaacEncBitCountScalefactorDelta(delta)
		}()
	}
}

func TestFDKaacEncBitCountScalefactorDeltaAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		sum := 0
		var got [10]int
		for i, delta := range [...]int{-60, -32, -1, 0, 1, 2, 10, 24, 42, 60} {
			got[i] = FDKaacEncBitCountScalefactorDelta(delta)
			sum += got[i]
		}
		bitCountSink = sum
		bitCountHashSink = hashBandEnergyInts(got[:])
	})
	if allocs != 0 {
		t.Fatalf("scalefactor bit-count allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncSpectralHuffmanLengthTables(t *testing.T) {
	if got, want := hashHuffLtab12(), uint64(0x1ce2ee9ace4b3f31); got != want {
		t.Fatalf("huff_ltab1_2 hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashHuffLtab34(), uint64(0x003d5aed96fc04ca04); got != want {
		t.Fatalf("huff_ltab3_4 hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashHuffLtab56(), uint64(0x90f6f28c185d3866); got != want {
		t.Fatalf("huff_ltab5_6 hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashHuffLtab78(), uint64(0x07167551d1b8a6f07); got != want {
		t.Fatalf("huff_ltab7_8 hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashHuffLtab910(), uint64(0xd53ff76a463607fd); got != want {
		t.Fatalf("huff_ltab9_10 hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashHuffLtab11(), uint64(0xc28a4a9d309ab353); got != want {
		t.Fatalf("huff_ltab11 hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncBitCountVectors(t *testing.T) {
	tests := []struct {
		name   string
		values [8]int16
		maxVal int
		want   [12]int
		hash   uint64
	}{
		{
			name:   "zero",
			values: [8]int16{0, 0, 0, 0, 0, 0, 0, 0},
			maxVal: 0,
			want:   [12]int{0, 2, 6, 2, 8, 4, 16, 4, 20, 4, 24, 16},
			hash:   0x1d5aa3971c3a3283,
		},
		{
			name:   "lav1",
			values: [8]int16{-1, 0, 1, 0, 1, -1, 0, 1},
			maxVal: 1,
			want:   [12]int{invalidBitcount, 16, 14, 18, 13, 17, 16, 18, 20, 18, 24, 24},
			hash:   0x293c21c966234825,
		},
		{
			name:   "lav2",
			values: [8]int16{-2, 0, 2, 1, -1, 2, 0, -2},
			maxVal: 2,
			want:   [12]int{invalidBitcount, invalidBitcount, invalidBitcount, 33, 26, 30, 24, 30, 24, 30, 26, 28},
			hash:   0xbe09662510b028ca,
		},
		{
			name:   "lav4",
			values: [8]int16{-4, 3, 0, 4, -2, 1, -3, 2},
			maxVal: 4,
			want:   [12]int{invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, 41, 33, 38, 29, 39, 29, 33},
			hash:   0x66ffa28233c8caa9,
		},
		{
			name:   "lav7",
			values: [8]int16{-7, 6, 0, 5, -4, 3, -2, 1},
			maxVal: 7,
			want:   [12]int{invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, 43, 34, 44, 34, 35},
			hash:   0xb78a6c0796f33c2d,
		},
		{
			name:   "lav12",
			values: [8]int16{-12, 11, 0, 10, -9, 8, -7, 6},
			maxVal: 12,
			want:   [12]int{invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, 58, 49, 47},
			hash:   0x2d8f21b6cab59235,
		},
		{
			name:   "lav15",
			values: [8]int16{-15, 14, 0, 13, -12, 11, -10, 9},
			maxVal: 15,
			want:   [12]int{invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, 51},
			hash:   0xf80a03be174dcaba,
		},
		{
			name:   "escape",
			values: [8]int16{-31, 16, 0, 17, -64, 1, 20, -18},
			maxVal: 31,
			want:   [12]int{invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, invalidBitcount, 69},
			hash:   0x36551e69b484312c,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := [12]int{-99, -99, -99, -99, -99, -99, -99, -99, -99, -99, -99, -99}
			FDKaacEncBitCount(tc.values[:], len(tc.values), tc.maxVal, got[:])
			assertIntSlice(t, "spectral bit counts", got[:], tc.want[:], tc.hash)
		})
	}
}

func TestFDKaacEncCountValuesVectors(t *testing.T) {
	tests := []struct {
		codeBook int
		values   []int16
		want     int
	}{
		{codeBookZeroNo, []int16{31, -16, 0, 1}, 0},
		{codeBook1No, []int16{-1, 0, 1, 0, 1, -1, 0, 1}, 16},
		{codeBook2No, []int16{-1, 0, 1, 0, 1, -1, 0, 1}, 14},
		{codeBook3No, []int16{-2, 0, 2, 1, -1, 2, 0, -2}, 33},
		{codeBook4No, []int16{-2, 0, 2, 1, -1, 2, 0, -2}, 26},
		{codeBook5No, []int16{-4, 3, 0, 4, -2, 1, -3, 2}, 41},
		{codeBook6No, []int16{-4, 3, 0, 4, -2, 1, -3, 2}, 33},
		{codeBook7No, []int16{-7, 6, 0, 5, -4, 3, -2, 1}, 43},
		{codeBook8No, []int16{-7, 6, 0, 5, -4, 3, -2, 1}, 34},
		{codeBook9No, []int16{-12, 11, 0, 10, -9, 8, -7, 6}, 58},
		{codeBook10No, []int16{-12, 11, 0, 10, -9, 8, -7, 6}, 49},
		{codeBookEscNo, []int16{-31, 16, 0, 17, -64, 1, 20, -18}, 69},
	}

	var got [12]int
	for i, tc := range tests {
		got[i] = FDKaacEncCountValues(tc.values, len(tc.values), tc.codeBook)
		if got[i] != tc.want {
			t.Fatalf("codebook %d bit count = %d, want %d", tc.codeBook, got[i], tc.want)
		}
	}
	assertIntSlice(t, "count-values bit counts", got[:], []int{0, 16, 14, 33, 26, 41, 33, 43, 34, 58, 49, 69}, 0xb3ab02ac6ae863ef)
}

func TestFDKaacEncBitCountRejectsInvalid(t *testing.T) {
	values := []int16{0, 0, 0, 0}
	var bitCount [12]int
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"negative width", func() { FDKaacEncBitCount(values, -4, 0, bitCount[:]) }},
		{"unaligned width", func() { FDKaacEncBitCount(values, 2, 0, bitCount[:]) }},
		{"negative max", func() { FDKaacEncBitCount(values, 4, -1, bitCount[:]) }},
		{"short values", func() { FDKaacEncBitCount(values[:3], 4, 0, bitCount[:]) }},
		{"short bitcount", func() { FDKaacEncBitCount(values, 4, 0, bitCount[:11]) }},
		{"negative codebook", func() { FDKaacEncCountValues(values, 4, -1) }},
		{"large codebook", func() { FDKaacEncCountValues(values, 4, 12) }},
		{"unaligned nonescape width", func() { FDKaacEncCountValues(values, 2, codeBook1No) }},
		{"unaligned escape width", func() { FDKaacEncCountValues(values, 3, codeBookEscNo) }},
		{"short count values", func() { FDKaacEncCountValues(values[:3], 4, codeBook1No) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tc.name)
				}
			}()
			tc.fn()
		})
	}
}

func TestFDKaacEncSpectralBitCountAllocs(t *testing.T) {
	values := [8]int16{-31, 16, 0, 17, -64, 1, 20, -18}
	var bitCount [12]int
	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncBitCount(values[:], len(values), 31, bitCount[:])
		bitCountSink = FDKaacEncCountValues(values[:], len(values), codeBookEscNo)
		bitCountHashSink = hashBandEnergyInts(bitCount[:])
	})
	if allocs != 0 {
		t.Fatalf("spectral bit-count allocations = %v, want 0", allocs)
	}
}

func hashHuffLtab12() uint64 {
	var x [3 * 3 * 3 * 3]int
	n := 0
	for i := range fdkaacEncHuffLtab12 {
		for j := range fdkaacEncHuffLtab12[i] {
			for k := range fdkaacEncHuffLtab12[i][j] {
				for l := range fdkaacEncHuffLtab12[i][j][k] {
					x[n] = int(fdkaacEncHuffLtab12[i][j][k][l])
					n++
				}
			}
		}
	}
	return hashBandEnergyInts(x[:])
}

func hashHuffLtab34() uint64 {
	var x [3 * 3 * 3 * 3]int
	n := 0
	for i := range fdkaacEncHuffLtab34 {
		for j := range fdkaacEncHuffLtab34[i] {
			for k := range fdkaacEncHuffLtab34[i][j] {
				for l := range fdkaacEncHuffLtab34[i][j][k] {
					x[n] = int(fdkaacEncHuffLtab34[i][j][k][l])
					n++
				}
			}
		}
	}
	return hashBandEnergyInts(x[:])
}

func hashHuffLtab56() uint64 {
	var x [9 * 9]int
	n := 0
	for i := range fdkaacEncHuffLtab56 {
		for j := range fdkaacEncHuffLtab56[i] {
			x[n] = int(fdkaacEncHuffLtab56[i][j])
			n++
		}
	}
	return hashBandEnergyInts(x[:])
}

func hashHuffLtab78() uint64 {
	var x [8 * 8]int
	n := 0
	for i := range fdkaacEncHuffLtab78 {
		for j := range fdkaacEncHuffLtab78[i] {
			x[n] = int(fdkaacEncHuffLtab78[i][j])
			n++
		}
	}
	return hashBandEnergyInts(x[:])
}

func hashHuffLtab910() uint64 {
	var x [13 * 13]int
	n := 0
	for i := range fdkaacEncHuffLtab910 {
		for j := range fdkaacEncHuffLtab910[i] {
			x[n] = int(fdkaacEncHuffLtab910[i][j])
			n++
		}
	}
	return hashBandEnergyInts(x[:])
}

func hashHuffLtab11() uint64 {
	var x [17 * 17]int
	n := 0
	for i := range fdkaacEncHuffLtab11 {
		for j := range fdkaacEncHuffLtab11[i] {
			x[n] = int(fdkaacEncHuffLtab11[i][j])
			n++
		}
	}
	return hashBandEnergyInts(x[:])
}
