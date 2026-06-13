package fdkaac

import (
	"bytes"
	"testing"
)

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

func TestFDKaacEncHuffmanCodewordTables(t *testing.T) {
	if got, n := hashHuffCtab1234(fdkaacEncHuffCtab1); n != 81 || got != 0x506876e813765fb5 {
		t.Fatalf("huff_ctab1 = len:%d hash:%#016x, want len:81 hash:%#016x", n, got, uint64(0x506876e813765fb5))
	}
	if got, n := hashHuffCtab1234(fdkaacEncHuffCtab2); n != 81 || got != 0xe0c1946d42d01641 {
		t.Fatalf("huff_ctab2 = len:%d hash:%#016x, want len:81 hash:%#016x", n, got, uint64(0xe0c1946d42d01641))
	}
	if got, n := hashHuffCtab1234(fdkaacEncHuffCtab3); n != 81 || got != 0xc70ceef4750f7ca1 {
		t.Fatalf("huff_ctab3 = len:%d hash:%#016x, want len:81 hash:%#016x", n, got, uint64(0xc70ceef4750f7ca1))
	}
	if got, n := hashHuffCtab1234(fdkaacEncHuffCtab4); n != 81 || got != 0xc82636a7300f8179 {
		t.Fatalf("huff_ctab4 = len:%d hash:%#016x, want len:81 hash:%#016x", n, got, uint64(0xc82636a7300f8179))
	}
	if got, n := hashHuffCtab56(fdkaacEncHuffCtab5); n != 81 || got != 0x2d3668b451382f89 {
		t.Fatalf("huff_ctab5 = len:%d hash:%#016x, want len:81 hash:%#016x", n, got, uint64(0x2d3668b451382f89))
	}
	if got, n := hashHuffCtab56(fdkaacEncHuffCtab6); n != 81 || got != 0x3831594defb6f9bd {
		t.Fatalf("huff_ctab6 = len:%d hash:%#016x, want len:81 hash:%#016x", n, got, uint64(0x3831594defb6f9bd))
	}
	if got, n := hashHuffCtab78(fdkaacEncHuffCtab7); n != 64 || got != 0xda7a9ca97d724b7e {
		t.Fatalf("huff_ctab7 = len:%d hash:%#016x, want len:64 hash:%#016x", n, got, uint64(0xda7a9ca97d724b7e))
	}
	if got, n := hashHuffCtab78(fdkaacEncHuffCtab8); n != 64 || got != 0x7d83d3f78aa60229 {
		t.Fatalf("huff_ctab8 = len:%d hash:%#016x, want len:64 hash:%#016x", n, got, uint64(0x7d83d3f78aa60229))
	}
	if got, n := hashHuffCtab910(fdkaacEncHuffCtab9); n != 169 || got != 0xd63e800a68ccbc71 {
		t.Fatalf("huff_ctab9 = len:%d hash:%#016x, want len:169 hash:%#016x", n, got, uint64(0xd63e800a68ccbc71))
	}
	if got, n := hashHuffCtab910(fdkaacEncHuffCtab10); n != 169 || got != 0x2f0343f9318fc689 {
		t.Fatalf("huff_ctab10 = len:%d hash:%#016x, want len:169 hash:%#016x", n, got, uint64(0x2f0343f9318fc689))
	}
	if got, n := hashHuffCtab11(); n != 357 || got != 0xcdbaf324c918934d {
		t.Fatalf("huff_ctab11 = len:%d hash:%#016x, want len:357 hash:%#016x", n, got, uint64(0xcdbaf324c918934d))
	}
	if got, n := hashHuffCtabScf(); n != 121 || got != 0x18dfd429c9508539 {
		t.Fatalf("huff_ctabscf = len:%d hash:%#016x, want len:121 hash:%#016x", n, got, uint64(0x18dfd429c9508539))
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

func TestFDKaacEncBitCountShortBandVector(t *testing.T) {
	values := [...]int16{0, 0, 7, -7}
	got := [12]int{-99, -99, -99, -99, -99, -99, -99, -99, -99, -99, -99, -99}
	FDKaacEncBitCount(values[:], 2, 0, got[:])
	want := [12]int{0, 1, 3, 1, 4, 2, 8, 2, 10, 2, 12, 8}
	if got != want {
		t.Fatalf("short band bit counts = %v, want %v", got, want)
	}
}

func TestFDKaacEncBitCountShortEscapeBandVector(t *testing.T) {
	values := [...]int16{16, 99}
	got := [12]int{-99, -99, -99, -99, -99, -99, -99, -99, -99, -99, -99, -99}
	FDKaacEncBitCount(values[:], 1, 16, got[:])

	want := [12]int{
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		invalidBitcount,
		int(fdkaacEncHuffLtab11[16][0]) + 6,
	}
	if got != want {
		t.Fatalf("short escape band bit counts = %v, want %v", got, want)
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

func TestFDKaacEncCodeValuesVectors(t *testing.T) {
	tests := []struct {
		codeBook int
		values   []int16
		bits     uint32
		want     []byte
	}{
		{codeBookZeroNo, []int16{31, -16, 0, 1}, 0, []byte{}},
		{codeBook1No, []int16{-1, 0, 1, 0, 1, -1, 0, 1}, 16, []byte{0xd9, 0xee}},
		{codeBook2No, []int16{-1, 0, 1, 0, 1, -1, 0, 1}, 14, []byte{0xa3, 0x98}},
		{codeBook3No, []int16{-2, 0, 2, 1, -1, 2, 0, -2}, 33, []byte{0xff, 0xea, 0x7f, 0xf2, 0x80}},
		{codeBook4No, []int16{-2, 0, 2, 1, -1, 2, 0, -2}, 26, []byte{0xfc, 0xe7, 0xed, 0x40}},
		{codeBook5No, []int16{-4, 3, 0, 4, -2, 1, -3, 2}, 41, []byte{0xff, 0x8f, 0xd3, 0xd9, 0xf7, 0x00}},
		{codeBook6No, []int16{-4, 3, 0, 4, -2, 1, -3, 2}, 33, []byte{0xff, 0x3f, 0x53, 0x78, 0x00}},
		{codeBook7No, []int16{-7, 6, 0, 5, -4, 3}, 35, []byte{0xff, 0xcb, 0xd6, 0xf7, 0xc0}},
		{codeBook8No, []int16{-7, 6, 0, 5, -4, 3}, 28, []byte{0xfe, 0x5e, 0x2c, 0x60}},
		{codeBook9No, []int16{-12, 11, 0, 10, -9, 8}, 44, []byte{0xff, 0xf6, 0xfd, 0xd7, 0xfa, 0xe0}},
		{codeBook10No, []int16{-12, 11, 0, 10, -9, 8}, 38, []byte{0xff, 0xcb, 0xf8, 0x3e, 0x08}},
		{codeBookEscNo, []int16{-31, 16, 0, 17, -64, 1, 20, -18}, 69, []byte{0x24, 0xf0, 0x71, 0xc0, 0xda, 0xd8, 0x02, 0x24, 0x10}},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			var storage [64]byte
			var out [64]byte
			var bs BitStream
			if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
				t.Fatal(err)
			}
			if got := FDKaacEncCodeValues(tc.values, len(tc.values), tc.codeBook, &bs); got != 0 {
				t.Fatalf("FDKaacEncCodeValues returned %d, want 0", got)
			}
			if got := BitStreamValidBits(&bs); got != tc.bits {
				t.Fatalf("codebook %d bits = %d, want %d", tc.codeBook, got, tc.bits)
			}
			ByteAlign(&bs, 0)
			n := FetchBuffer(&bs, out[:])
			if !bytes.Equal(out[:n], tc.want) {
				t.Fatalf("codebook %d bytes = % x, want % x", tc.codeBook, out[:n], tc.want)
			}
		})
	}
}

func TestFDKaacEncCodeScalefactorDeltaVectors(t *testing.T) {
	var storage [16]byte
	var out [16]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}
	for _, delta := range []int{-60, -1, 0, 1, 2, 23, 60} {
		if got := FDKaacEncCodeScalefactorDelta(delta, &bs); got != 0 {
			t.Fatalf("FDKaacEncCodeScalefactorDelta(%d) = %d, want 0", delta, got)
		}
	}
	if got := BitStreamValidBits(&bs); got != 62 {
		t.Fatalf("scalefactor code bits = %d, want 62", got)
	}
	ByteAlign(&bs, 0)
	n := FetchBuffer(&bs, out[:])
	if want := []byte{0xff, 0xfa, 0x22, 0xb3, 0xff, 0x1f, 0xff, 0xcc}; !bytes.Equal(out[:n], want) {
		t.Fatalf("scalefactor code bytes = % x, want % x", out[:n], want)
	}

	ResetBitStream(&bs, BSWriter)
	if got := FDKaacEncCodeScalefactorDelta(61, &bs); got != 1 {
		t.Fatalf("out-of-range scalefactor code result = %d, want 1", got)
	}
	if got := BitStreamValidBits(&bs); got != 0 {
		t.Fatalf("out-of-range scalefactor wrote %d bits, want 0", got)
	}
}

func TestFDKaacEncBitCountRejectsInvalid(t *testing.T) {
	values := []int16{0, 0, 0, 0}
	tooWideValues := make([]int16, maxSpectralLines+1)
	var bitCount [12]int
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"negative width", func() { FDKaacEncBitCount(values, -4, 0, bitCount[:]) }},
		{"too wide", func() { FDKaacEncBitCount(tooWideValues, maxSpectralLines+1, 0, bitCount[:]) }},
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

func TestFDKaacEncCodeValuesRejectsInvalid(t *testing.T) {
	values := []int16{0, 0, 0, 0}
	var storage [16]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"negative width", func() { FDKaacEncCodeValues(values, -4, codeBook1No, &bs) }},
		{"invalid codebook", func() { FDKaacEncCodeValues(values, 4, 12, &bs) }},
		{"unaligned quad width", func() { FDKaacEncCodeValues(values, 2, codeBook1No, &bs) }},
		{"unaligned pair width", func() { FDKaacEncCodeValues(values, 3, codeBook7No, &bs) }},
		{"unaligned escape width", func() { FDKaacEncCodeValues(values, 3, codeBookEscNo, &bs) }},
		{"short values", func() { FDKaacEncCodeValues(values[:3], 4, codeBook1No, &bs) }},
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

func TestFDKaacEncSpectralBitCountShortBandAllocs(t *testing.T) {
	values := [4]int16{0, 0, 7, -7}
	var bitCount [12]int
	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncBitCount(values[:], 2, 0, bitCount[:])
		bitCountHashSink = hashBandEnergyInts(bitCount[:])
	})
	if allocs != 0 {
		t.Fatalf("short-band spectral bit-count allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncCodeValuesAllocs(t *testing.T) {
	values := [8]int16{-31, 16, 0, 17, -64, 1, 20, -18}
	var storage [64]byte
	var out [64]byte
	var bs BitStream
	allocs := testing.AllocsPerRun(1000, func() {
		clear(storage[:])
		clear(out[:])
		if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
			t.Fatal(err)
		}
		FDKaacEncCodeValues(values[:], len(values), codeBookEscNo, &bs)
		FDKaacEncCodeScalefactorDelta(23, &bs)
		ByteAlign(&bs, 0)
		n := FetchBuffer(&bs, out[:])
		bitCountSink = n
		bitCountHashSink = hashHuffBytes(out[:n])
	})
	if allocs != 0 {
		t.Fatalf("spectral code-values allocations = %v, want 0", allocs)
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

func hashHuffCtab1234(table [3][3][3][3]uint16) (uint64, int) {
	h := uint64(14695981039346656037)
	n := 0
	for i := range table {
		for j := range table[i] {
			for k := range table[i][j] {
				for l := range table[i][j][k] {
					h = hashHuffUint32(h, uint32(table[i][j][k][l]))
					n++
				}
			}
		}
	}
	return h, n
}

func hashHuffCtab56(table [9][9]uint16) (uint64, int) {
	h := uint64(14695981039346656037)
	n := 0
	for i := range table {
		for j := range table[i] {
			h = hashHuffUint32(h, uint32(table[i][j]))
			n++
		}
	}
	return h, n
}

func hashHuffCtab78(table [8][8]uint16) (uint64, int) {
	h := uint64(14695981039346656037)
	n := 0
	for i := range table {
		for j := range table[i] {
			h = hashHuffUint32(h, uint32(table[i][j]))
			n++
		}
	}
	return h, n
}

func hashHuffCtab910(table [13][13]uint16) (uint64, int) {
	h := uint64(14695981039346656037)
	n := 0
	for i := range table {
		for j := range table[i] {
			h = hashHuffUint32(h, uint32(table[i][j]))
			n++
		}
	}
	return h, n
}

func hashHuffCtab11() (uint64, int) {
	h := uint64(14695981039346656037)
	n := 0
	for i := range fdkaacEncHuffCtab11 {
		for j := range fdkaacEncHuffCtab11[i] {
			h = hashHuffUint32(h, uint32(fdkaacEncHuffCtab11[i][j]))
			n++
		}
	}
	return h, n
}

func hashHuffCtabScf() (uint64, int) {
	h := uint64(14695981039346656037)
	for _, v := range fdkaacEncHuffCtabScf {
		h = hashHuffUint32(h, v)
	}
	return h, len(fdkaacEncHuffCtabScf)
}

func hashHuffUint32(h uint64, v uint32) uint64 {
	h = fnv64AddByte(h, byte(v))
	h = fnv64AddByte(h, byte(v>>8))
	h = fnv64AddByte(h, byte(v>>16))
	h = fnv64AddByte(h, byte(v>>24))
	return h
}

func hashHuffBytes(x []byte) uint64 {
	h := uint64(14695981039346656037)
	for _, b := range x {
		h = fnv64AddByte(h, b)
	}
	return h
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
