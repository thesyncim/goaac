package fdkaac

import "testing"

var quantizeSink FixpDBL
var quantizeHashSink uint64

func TestFDKaacEncQuantizerTableVectors(t *testing.T) {
	if len(fdkaacEncMTab34) != 512 || len(fdkaacEncMTab43Elc) != 512 {
		t.Fatalf("quantizer table lengths = %d/%d, want 512/512", len(fdkaacEncMTab34), len(fdkaacEncMTab43Elc))
	}
	if fdkaacEncMTab34[0] != 0x4c1bf829 || fdkaacEncMTab34[511] != 0x7fe7ff40 {
		t.Fatalf("mTab34 edge entries = %#x/%#x", uint32(fdkaacEncMTab34[0]), uint32(fdkaacEncMTab34[511]))
	}
	if fdkaacEncMTab43Elc[0] != 0x32cbfd4a || fdkaacEncMTab43Elc[511] != 0x7fd5571d {
		t.Fatalf("mTab43 edge entries = %#x/%#x", uint32(fdkaacEncMTab43Elc[0]), uint32(fdkaacEncMTab43Elc[511]))
	}
	assertFixpDBLSlice(t, "mTab34", fdkaacEncMTab34[:], fdkaacEncMTab34[:], 0xba6d3e66d1dc61eb)
	assertFixpDBLSlice(t, "quantTableQ", fdkaacEncQuantTableQ[:], fdkaacEncQuantTableQ[:], 0x2a4b4b98c071f5ca)
	assertFixpDBLSlice(t, "quantTableE", fdkaacEncQuantTableE[:], fdkaacEncQuantTableE[:], 0x68456ec8ceedd78b)
	assertFixpDBLSlice(t, "mTab43", fdkaacEncMTab43Elc[:], fdkaacEncMTab43Elc[:], 0x569201da133c0943)

	var mantFlat [56]FixpDBL
	var expFlat [56]int
	k := 0
	for i := range fdkaacEncSpecExpMantTableCombElc {
		for j := range fdkaacEncSpecExpMantTableCombElc[i] {
			mantFlat[k] = fdkaacEncSpecExpMantTableCombElc[i][j]
			expFlat[k] = int(fdkaacEncSpecExpTableComb[i][j])
			k++
		}
	}
	assertFixpDBLSlice(t, "spec exponent mantissa", mantFlat[:], mantFlat[:], 0x5e4c4eef6ac68b6c)
	assertIntSlice(t, "spec exponent shifts", expFlat[:], expFlat[:], 0x58cc4c5392b0b6d4)
}

func TestFDKaacEncQuantizeLinesVectors(t *testing.T) {
	spec := quantizeVectorSpectrum()
	var gotNoDeadZone [len(spec)]int16
	var gotDeadZone [len(spec)]int16

	FDKaacEncQuantizeLines(-20, len(spec), spec[:], gotNoDeadZone[:], 0)
	FDKaacEncQuantizeLines(-16, len(spec), spec[:], gotDeadZone[:], 1)

	wantNoDeadZone := [...]int16{0, 0, 0, 0, 0, 1, -2, 3, -4, 5}
	wantDeadZone := [...]int16{0, 0, 0, 0, 0, 0, -1, 1, -2, 3}
	assertInt16Slice(t, "quantized no-dead-zone lines", gotNoDeadZone[:], wantNoDeadZone[:], 0x7cc28575a454dd28)
	assertInt16Slice(t, "quantized dead-zone lines", gotDeadZone[:], wantDeadZone[:], 0x3b4b0d50625f13fe)
}

func TestFDKaacEncInvQuantizeLinesVectors(t *testing.T) {
	quant := [...]int16{0, 1, -1, 3, -7, 15, -63, 511, -8191}
	var got [len(quant)]FixpDBL

	FDKaacEncInvQuantizeLines(-20, len(quant), quant[:], got[:])

	want := [...]FixpDBL{0, 33554431, -33554431, 145181595, -449311235, 1241285180, 178489312, -357797376, -2059579392}
	assertFixpDBLSlice(t, "inverse quantized lines", got[:], want[:], 0x02e3fcd5bc21618b)
}

func TestFDKaacEncQuantizeSpectrumVectors(t *testing.T) {
	spec := quantizeVectorSpectrum()
	offsets := [...]int{0, 2, 5, 8, 10}
	scalefactors := [...]int{0, -4, 4, 9}
	quant := [...]int16{1234, 1234, 1234, 1234, 1234, 1234, 1234, 1234, 1234, 1234}

	FDKaacEncQuantizeSpectrum(4, 3, 4, offsets[:], spec[:], -20, scalefactors[:], quant[:], 0)

	want := [...]int16{0, 0, 0, 0, 0, 2, -3, 5, 1234, 1234}
	assertInt16Slice(t, "quantized spectrum", quant[:], want[:], 0x7f8599aa1f3e4e2c)
}

func TestFDKaacEncCalcSfbDistVectors(t *testing.T) {
	spec := quantizeVectorSpectrum()
	var quant [7]int16

	got := FDKaacEncCalcSfbDist(spec[1:8], quant[:], len(quant), -20, 0)

	if got != -373521274 {
		t.Fatalf("SFB distortion = %d, want -373521274", got)
	}
	wantQuant := [...]int16{0, 0, 0, 0, 1, -2, 3}
	assertInt16Slice(t, "SFB distortion quant", quant[:], wantQuant[:], 0xf4f4754f5f8670f2)
}

func TestFDKaacEncCalcSfbQuantEnergyAndDistVectors(t *testing.T) {
	spec := quantizeVectorSpectrum()
	quant := [...]int16{0, 0, 0, 0, 1, -2, 3}

	energy, dist := FDKaacEncCalcSfbQuantEnergyAndDist(spec[1:8], quant[:], len(quant), -20)

	got := [...]FixpDBL{energy, dist}
	want := [...]FixpDBL{-177692928, -373521274}
	assertFixpDBLSlice(t, "quantized SFB energy/distortion", got[:], want[:], 0x89e2e99fcfeba676)

	badQuant := [...]int16{9000, 1}
	energy, dist = FDKaacEncCalcSfbQuantEnergyAndDist(spec[:2], badQuant[:], len(badQuant), -20)
	if energy != 0 || dist != 0 {
		t.Fatalf("out-of-range quantized energy/distortion = %d/%d, want 0/0", energy, dist)
	}
}

func TestFDKaacEncQuantizeRejectsInvalid(t *testing.T) {
	spec := quantizeVectorSpectrum()
	var quant [len(spec)]int16
	var inv [len(spec)]FixpDBL
	offsets := [...]int{0, 2, 5, 8, 10}
	scalefactors := [...]int{0, -4, 4, 9}

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "negative quantize lines", fn: func() {
			FDKaacEncQuantizeLines(-20, -1, spec[:], quant[:], 0)
		}},
		{name: "short quantize spectrum input", fn: func() {
			FDKaacEncQuantizeLines(-20, len(spec), spec[:len(spec)-1], quant[:], 0)
		}},
		{name: "short quantize spectrum output", fn: func() {
			FDKaacEncQuantizeLines(-20, len(spec), spec[:], quant[:len(quant)-1], 0)
		}},
		{name: "negative inverse quantize lines", fn: func() {
			FDKaacEncInvQuantizeLines(-20, -1, quant[:], inv[:])
		}},
		{name: "short inverse quantize input", fn: func() {
			FDKaacEncInvQuantizeLines(-20, len(quant), quant[:len(quant)-1], inv[:])
		}},
		{name: "short inverse quantize output", fn: func() {
			FDKaacEncInvQuantizeLines(-20, len(quant), quant[:], inv[:len(inv)-1])
		}},
		{name: "inverse quantize out of range", fn: func() {
			bad := [...]int16{9000}
			FDKaacEncInvQuantizeLines(-20, 1, bad[:], inv[:1])
		}},
		{name: "bad quantize band count", fn: func() {
			FDKaacEncQuantizeSpectrum(0, 3, 4, offsets[:], spec[:], -20, scalefactors[:], quant[:], 0)
		}},
		{name: "bad quantize group multiple", fn: func() {
			FDKaacEncQuantizeSpectrum(4, 2, 3, offsets[:], spec[:], -20, scalefactors[:], quant[:], 0)
		}},
		{name: "bad quantize group width", fn: func() {
			FDKaacEncQuantizeSpectrum(4, 5, 4, offsets[:], spec[:], -20, scalefactors[:], quant[:], 0)
		}},
		{name: "short quantize offsets", fn: func() {
			FDKaacEncQuantizeSpectrum(4, 3, 4, offsets[:4], spec[:], -20, scalefactors[:], quant[:], 0)
		}},
		{name: "short quantize scalefactors", fn: func() {
			FDKaacEncQuantizeSpectrum(4, 3, 4, offsets[:], spec[:], -20, scalefactors[:3], quant[:], 0)
		}},
		{name: "decreasing quantize offset", fn: func() {
			bad := offsets
			bad[2] = bad[1] - 1
			FDKaacEncQuantizeSpectrum(4, 3, 4, bad[:], spec[:], -20, scalefactors[:], quant[:], 0)
		}},
		{name: "short quantize mdct", fn: func() {
			FDKaacEncQuantizeSpectrum(4, 3, 4, offsets[:], spec[:9], -20, scalefactors[:], quant[:], 0)
		}},
		{name: "short quantize output", fn: func() {
			FDKaacEncQuantizeSpectrum(4, 3, 4, offsets[:], spec[:], -20, scalefactors[:], quant[:9], 0)
		}},
		{name: "negative SFB distortion lines", fn: func() {
			FDKaacEncCalcSfbDist(spec[:], quant[:], -1, -20, 0)
		}},
		{name: "short SFB distortion input", fn: func() {
			FDKaacEncCalcSfbDist(spec[:len(spec)-1], quant[:], len(spec), -20, 0)
		}},
		{name: "short SFB distortion output", fn: func() {
			FDKaacEncCalcSfbDist(spec[:], quant[:len(quant)-1], len(spec), -20, 0)
		}},
	}

	for _, tt := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		}()
	}
}

func TestFDKaacEncQuantizeAllocs(t *testing.T) {
	spec := quantizeVectorSpectrum()
	offsets := [...]int{0, 2, 5, 8, 10}
	scalefactors := [...]int{0, -4, 4, 9}
	var quant [len(spec)]int16
	var inv [len(spec)]FixpDBL

	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncQuantizeLines(-20, len(spec), spec[:], quant[:], 0)
		FDKaacEncInvQuantizeLines(-20, len(spec), quant[:], inv[:])
		FDKaacEncQuantizeSpectrum(4, 3, 4, offsets[:], spec[:], -20, scalefactors[:], quant[:], 0)
		dist := FDKaacEncCalcSfbDist(spec[1:8], quant[:7], 7, -20, 0)
		energy, qdist := FDKaacEncCalcSfbQuantEnergyAndDist(spec[1:8], quant[:7], 7, -20)
		quantizeSink = dist + energy + qdist + inv[3]
		quantizeHashSink = hashInt16AsInt(quant[:])
	})
	if allocs != 0 {
		t.Fatalf("quantize allocations = %v, want 0", allocs)
	}
}

func quantizeVectorSpectrum() [10]FixpDBL {
	return [...]FixpDBL{
		0,
		0x00100000,
		-0x00200000,
		0x00800000,
		-0x01000000,
		0x04000000,
		-0x08000000,
		0x10000000,
		-0x18000000,
		0x20000000,
	}
}

func assertInt16Slice(t *testing.T, name string, got []int16, want []int16, wantHash uint64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %d, want %d; got %v want %v", name, i, got[i], want[i], got, want)
		}
	}
	if h := hashInt16AsInt(got); h != wantHash {
		t.Fatalf("%s hash = %#016x, want %#016x", name, h, wantHash)
	}
}

func hashInt16AsInt(x []int16) uint64 {
	h := uint64(14695981039346656037)
	for _, v := range x {
		u := uint32(int32(v))
		h = fnv64AddByte(h, byte(u))
		h = fnv64AddByte(h, byte(u>>8))
		h = fnv64AddByte(h, byte(u>>16))
		h = fnv64AddByte(h, byte(u>>24))
	}
	return h
}
