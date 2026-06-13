package fdkaac

import "testing"

var bandEnergySink FixpDBL
var bandEnergyIntSink int

func TestFDKaacEncBandEnergyShortVectors(t *testing.T) {
	bandOffset := [...]int{0, 3, 7, 12, 16, 24, 32, 48, 64}
	var spec [64]FixpDBL
	var scales [8]int
	var energy [8]FixpDBL

	fillBandEnergySpec(spec[:], 11)
	FDKaacEncCalcSfbMaxScaleSpec(spec[:], bandOffset[:], scales[:], len(scales))

	wantScales := [...]int{7, 7, 7, 7, 7, 7, 7, 6}
	if scales != wantScales {
		t.Fatalf("short band scales = %v, want %v", scales, wantScales)
	}
	if got, want := hashBandEnergyInts(scales[:]), uint64(0x091f8a1c8f6fd254); got != want {
		t.Fatalf("short band scale hash = %#016x, want %#016x", got, want)
	}

	FDKaacEncCalcBandEnergyOptimShort(spec[:], scales[:], bandOffset[:], len(scales), energy[:])

	wantEnergy := [...]FixpDBL{55459, 107404, 284854, 207515, 178367, 359516, 541924, 841488}
	if energy != wantEnergy {
		t.Fatalf("short band energy = %v, want %v", energy, wantEnergy)
	}
	if got, want := hashFixpDBL(energy[:]), uint64(0x20b6123dcc1dd2c3); got != want {
		t.Fatalf("short band energy hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncBandNrgMSOptVectors(t *testing.T) {
	bandOffset := [...]int{0, 3, 7, 12, 16, 24, 32, 48, 64}
	var left [64]FixpDBL
	var right [64]FixpDBL
	var leftScales [8]int
	var rightScales [8]int
	var mid [8]FixpDBL
	var side [8]FixpDBL

	fillBandEnergySpec(left[:], 17)
	fillBandEnergySpec(right[:], 29)
	FDKaacEncCalcSfbMaxScaleSpec(left[:], bandOffset[:], leftScales[:], len(leftScales))
	FDKaacEncCalcSfbMaxScaleSpec(right[:], bandOffset[:], rightScales[:], len(rightScales))

	wantLeftScales := [...]int{7, 7, 7, 7, 6, 7, 7, 6}
	wantRightScales := [...]int{7, 7, 7, 7, 7, 7, 6, 7}
	if leftScales != wantLeftScales {
		t.Fatalf("left scales = %v, want %v", leftScales, wantLeftScales)
	}
	if rightScales != wantRightScales {
		t.Fatalf("right scales = %v, want %v", rightScales, wantRightScales)
	}
	if got, want := hashBandEnergyInts(leftScales[:]), uint64(0xab8c26b602c70d75); got != want {
		t.Fatalf("left scale hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashBandEnergyInts(rightScales[:]), uint64(0x881fa51d44295fe4); got != want {
		t.Fatalf("right scale hash = %#016x, want %#016x", got, want)
	}

	FDKaacEncCalcBandNrgMSOpt(left[:], right[:], leftScales[:], rightScales[:], bandOffset[:], len(leftScales), mid[:], side[:], false, nil, nil)

	wantMid := [...]FixpDBL{18347, 149626, 78810, 40555, 93577, 350318, 378596, 208659}
	wantSide := [...]FixpDBL{115634, 48598, 63691, 72185, 287306, 99470, 445417, 439751}
	if mid != wantMid {
		t.Fatalf("mid band energy = %v, want %v", mid, wantMid)
	}
	if side != wantSide {
		t.Fatalf("side band energy = %v, want %v", side, wantSide)
	}
	if got, want := hashFixpDBL(mid[:]), uint64(0x429d3767236ea3aa); got != want {
		t.Fatalf("mid band energy hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(side[:]), uint64(0xa04b55300f81b455); got != want {
		t.Fatalf("side band energy hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncBandEnergyRejectsUnsupported(t *testing.T) {
	var spec [4]FixpDBL
	var energy [2]FixpDBL
	validOffsets := [...]int{0, 2, 4}
	validScales := [...]int{2, 2}

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "negative band count",
			fn: func() {
				FDKaacEncCalcSfbMaxScaleSpec(spec[:], validOffsets[:], validScales[:], -1)
			},
		},
		{
			name: "short offsets",
			fn: func() {
				FDKaacEncCalcSfbMaxScaleSpec(spec[:], validOffsets[:1], validScales[:], 2)
			},
		},
		{
			name: "decreasing offsets",
			fn: func() {
				FDKaacEncCalcSfbMaxScaleSpec(spec[:], []int{0, 3, 2}, validScales[:], 2)
			},
		},
		{
			name: "short output",
			fn: func() {
				FDKaacEncCalcSfbMaxScaleSpec(spec[:], validOffsets[:], validScales[:1], 2)
			},
		},
		{
			name: "invalid scale",
			fn: func() {
				FDKaacEncCalcBandEnergyOptimShort(spec[:], []int{31, 2}, validOffsets[:], 2, energy[:])
			},
		},
		{
			name: "ld-data branch",
			fn: func() {
				FDKaacEncCalcBandNrgMSOpt(spec[:], spec[:], validScales[:], validScales[:], validOffsets[:], 2, energy[:], energy[:], true, energy[:], energy[:])
			},
		},
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

func TestFDKaacEncBandEnergyAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		bandOffset := [...]int{0, 3, 7, 12, 16, 24, 32, 48, 64}
		var spec [64]FixpDBL
		var left [64]FixpDBL
		var right [64]FixpDBL
		var scales [8]int
		var leftScales [8]int
		var rightScales [8]int
		var energy [8]FixpDBL
		var mid [8]FixpDBL
		var side [8]FixpDBL

		fillBandEnergySpec(spec[:], 11)
		fillBandEnergySpec(left[:], 17)
		fillBandEnergySpec(right[:], 29)
		FDKaacEncCalcSfbMaxScaleSpec(spec[:], bandOffset[:], scales[:], len(scales))
		FDKaacEncCalcBandEnergyOptimShort(spec[:], scales[:], bandOffset[:], len(scales), energy[:])
		FDKaacEncCalcSfbMaxScaleSpec(left[:], bandOffset[:], leftScales[:], len(leftScales))
		FDKaacEncCalcSfbMaxScaleSpec(right[:], bandOffset[:], rightScales[:], len(rightScales))
		FDKaacEncCalcBandNrgMSOpt(left[:], right[:], leftScales[:], rightScales[:], bandOffset[:], len(leftScales), mid[:], side[:], false, nil, nil)
		bandEnergySink = energy[0] + mid[0] + side[0]
		bandEnergyIntSink = scales[0] + leftScales[0] + rightScales[0]
	})
	if allocs != 0 {
		t.Fatalf("band energy allocations = %v, want 0", allocs)
	}
}

func fillBandEnergySpec(dst []FixpDBL, seed int) {
	x := int32(uint32(0x9e3779b9) ^ uint32(seed*0x45d9f3b))
	for i := range dst {
		x = int32(int64(x)*1664525 + 1013904223 + int64(i*97+seed))
		bucket := int(uint32(x)%97) - 48
		amp := 0x00041000 + ((i+seed)%7)*0x00005000
		dst[i] = FixpDBL(bucket * amp)
	}
}

func hashBandEnergyInts(x []int) uint64 {
	h := uint64(14695981039346656037)
	for _, v := range x {
		u := uint32(v)
		h = fnv64AddByte(h, byte(u))
		h = fnv64AddByte(h, byte(u>>8))
		h = fnv64AddByte(h, byte(u>>16))
		h = fnv64AddByte(h, byte(u>>24))
	}
	return h
}
