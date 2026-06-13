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
	var midLd [8]FixpDBL
	var sideLd [8]FixpDBL

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

	FDKaacEncCalcBandNrgMSOpt(left[:], right[:], leftScales[:], rightScales[:], bandOffset[:], len(leftScales), mid[:], side[:], true, midLd[:], sideLd[:])

	wantMid := [...]FixpDBL{18347, 149626, 78810, 40555, 93577, 350318, 378596, 208659}
	wantSide := [...]FixpDBL{115634, 48598, 63691, 72185, 287306, 99470, 445417, 439751}
	wantMidLd := [...]FixpDBL{-564944250, -463352363, -494387184, -526548844, -486073593, -422171559, -418413597, -447254035}
	wantSideLd := [...]FixpDBL{-475828249, -517790832, -504698384, -498636955, -431769422, -483117437, -410545189, -411164931}
	if mid != wantMid {
		t.Fatalf("mid band energy = %v, want %v", mid, wantMid)
	}
	if side != wantSide {
		t.Fatalf("side band energy = %v, want %v", side, wantSide)
	}
	if midLd != wantMidLd {
		t.Fatalf("mid ld band energy = %v, want %v", midLd, wantMidLd)
	}
	if sideLd != wantSideLd {
		t.Fatalf("side ld band energy = %v, want %v", sideLd, wantSideLd)
	}
	if got, want := hashFixpDBL(mid[:]), uint64(0x429d3767236ea3aa); got != want {
		t.Fatalf("mid band energy hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(side[:]), uint64(0xa04b55300f81b455); got != want {
		t.Fatalf("side band energy hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(midLd[:]), uint64(0xf6d1e9828a0edebb); got != want {
		t.Fatalf("mid ld band energy hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(sideLd[:]), uint64(0xb4212193c5ea1ed8); got != want {
		t.Fatalf("side ld band energy hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncBandEnergyLongVectors(t *testing.T) {
	bandOffset := [...]int{0, 4, 9, 16, 28, 40, 56, 72, 96}
	var spec [96]FixpDBL
	var scales [8]int
	var checkEnergy [8]FixpDBL
	var checkLd [8]FixpDBL
	var longEnergy [8]FixpDBL
	var longLd [8]FixpDBL

	fillBandEnergySpec(spec[:], 41)
	FDKaacEncCalcSfbMaxScaleSpec(spec[:], bandOffset[:], scales[:], len(scales))

	wantScales := [...]int{7, 7, 6, 7, 7, 6, 6, 6}
	if scales != wantScales {
		t.Fatalf("long band scales = %v, want %v", scales, wantScales)
	}
	if got, want := hashBandEnergyInts(scales[:]), uint64(0x33212a8cd2b60d05); got != want {
		t.Fatalf("long band scale hash = %#016x, want %#016x", got, want)
	}

	maxNrg := FDKaacEncCheckBandEnergyOptim(spec[:], scales[:], bandOffset[:], len(scales), checkEnergy[:], checkLd[:], 2)
	if maxNrg != 16491088 {
		t.Fatalf("check max energy = %d, want 16491088", maxNrg)
	}

	wantCheckEnergy := [...]FixpDBL{13069862, 6696762, 6669916, 25121236, 30283786, 9514870, 8834672, 16491088}
	wantLd := [...]FixpDBL{-448295917, -480666156, -413751744, -416665376, -407617823, -396553851, -400143065, -369931465}
	if checkEnergy != wantCheckEnergy {
		t.Fatalf("check band energy = %v, want %v", checkEnergy, wantCheckEnergy)
	}
	if checkLd != wantLd {
		t.Fatalf("check ld band energy = %v, want %v", checkLd, wantLd)
	}
	if got, want := hashFixpDBL(checkEnergy[:]), uint64(0x90bb4783283b7e47); got != want {
		t.Fatalf("check band energy hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(checkLd[:]), uint64(0x71285fe8b16f8e02); got != want {
		t.Fatalf("check ld band energy hash = %#016x, want %#016x", got, want)
	}

	shiftBits := FDKaacEncCalcBandEnergyOptimLong(spec[:], scales[:], bandOffset[:], len(scales), longEnergy[:], longLd[:])
	if shiftBits != 0 {
		t.Fatalf("long shift bits = %d, want 0", shiftBits)
	}

	wantLongEnergy := [...]FixpDBL{204216, 104636, 416869, 392519, 473184, 594679, 552167, 1030693}
	if longEnergy != wantLongEnergy {
		t.Fatalf("long band energy = %v, want %v", longEnergy, wantLongEnergy)
	}
	if longLd != wantLd {
		t.Fatalf("long ld band energy = %v, want %v", longLd, wantLd)
	}
	if got, want := hashFixpDBL(longEnergy[:]), uint64(0xa58c7b1540170696); got != want {
		t.Fatalf("long band energy hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(longLd[:]), uint64(0x71285fe8b16f8e02); got != want {
		t.Fatalf("long ld band energy hash = %#016x, want %#016x", got, want)
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
			name: "short ld-data output",
			fn: func() {
				FDKaacEncCalcBandNrgMSOpt(spec[:], spec[:], validScales[:], validScales[:], validOffsets[:], 2, energy[:], energy[:], true, energy[:1], energy[:])
			},
		},
		{
			name: "missing ld-data output",
			fn: func() {
				FDKaacEncCalcBandNrgMSOpt(spec[:], spec[:], validScales[:], validScales[:], validOffsets[:], 2, energy[:], energy[:], true, nil, nil)
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
		var ld [8]FixpDBL
		var midLd [8]FixpDBL
		var sideLd [8]FixpDBL

		fillBandEnergySpec(spec[:], 11)
		fillBandEnergySpec(left[:], 17)
		fillBandEnergySpec(right[:], 29)
		FDKaacEncCalcSfbMaxScaleSpec(spec[:], bandOffset[:], scales[:], len(scales))
		FDKaacEncCalcBandEnergyOptimShort(spec[:], scales[:], bandOffset[:], len(scales), energy[:])
		_ = FDKaacEncCheckBandEnergyOptim(spec[:], scales[:], bandOffset[:], len(scales), energy[:], ld[:], 2)
		_ = FDKaacEncCalcBandEnergyOptimLong(spec[:], scales[:], bandOffset[:], len(scales), energy[:], ld[:])
		FDKaacEncCalcSfbMaxScaleSpec(left[:], bandOffset[:], leftScales[:], len(leftScales))
		FDKaacEncCalcSfbMaxScaleSpec(right[:], bandOffset[:], rightScales[:], len(rightScales))
		FDKaacEncCalcBandNrgMSOpt(left[:], right[:], leftScales[:], rightScales[:], bandOffset[:], len(leftScales), mid[:], side[:], true, midLd[:], sideLd[:])
		bandEnergySink = energy[0] + mid[0] + side[0] + ld[0] + midLd[0] + sideLd[0]
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
