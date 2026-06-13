package fdkaac

import "testing"

var ldDataSink FixpDBL

func TestCalcLdDataVectors(t *testing.T) {
	input := [...]FixpDBL{
		0,
		-1,
		1,
		0x00010000,
		0x00100000,
		0x01000000,
		0x10000000,
		0x40000000,
		0x60000000,
		MaxValDBL,
	}
	want := [...]FixpDBL{
		-2147483648,
		-2147483648,
		-1040183407,
		-503312495,
		-369094767,
		-234877039,
		-100659311,
		-33550447,
		-13926347,
		-1,
	}

	var got [len(input)]FixpDBL
	for i, v := range input {
		got[i] = CalcLdData(v)
	}
	if got != want {
		t.Fatalf("CalcLdData = %v, want %v", got, want)
	}
	if h, wantHash := hashFixpDBL(got[:]), uint64(0xa8c1deb02695f42e); h != wantHash {
		t.Fatalf("CalcLdData hash = %#016x, want %#016x", h, wantHash)
	}

	var vector [len(input)]FixpDBL
	LdDataVector(input[:], vector[:], len(input))
	if vector != want {
		t.Fatalf("LdDataVector = %v, want %v", vector, want)
	}

	inPlace := input
	LdDataVector(inPlace[:], inPlace[:], len(inPlace))
	if inPlace != want {
		t.Fatalf("in-place LdDataVector = %v, want %v", inPlace, want)
	}
}

func TestCalcInvLdDataVectors(t *testing.T) {
	if len(exp2TabLong) != 32 || len(exp2WTabLong) != 32 || len(exp2XTabLong) != 32 {
		t.Fatalf("unexpected exp2 table lengths: %d/%d/%d", len(exp2TabLong), len(exp2WTabLong), len(exp2XTabLong))
	}
	if exp2TabLong[0] != 0x40000000 || exp2TabLong[31] != 0x7D41D96E {
		t.Fatalf("exp2TabLong edge entries = %#x/%#x", exp2TabLong[0], exp2TabLong[31])
	}
	if exp2WTabLong[0] != 0x40000000 || exp2WTabLong[31] != 0x415B6EFB {
		t.Fatalf("exp2WTabLong edge entries = %#x/%#x", exp2WTabLong[0], exp2WTabLong[31])
	}
	if exp2XTabLong[0] != 0x40000000 || exp2XTabLong[31] != 0x400ABF4F {
		t.Fatalf("exp2XTabLong edge entries = %#x/%#x", exp2XTabLong[0], exp2XTabLong[31])
	}

	input := [...]FixpDBL{
		-1040187393,
		-1040187392,
		-805306368,
		-268435456,
		-1,
		0,
		1,
		16777216,
		268435456,
		1040187391,
		1040187392,
		MaxValDBL,
	}
	want := [...]FixpDBL{
		0,
		1,
		128,
		8388608,
		2147483584,
		MaxValDBL,
		1,
		1,
		256,
		2147483584,
		MaxValDBL,
		MaxValDBL,
	}

	var got [len(input)]FixpDBL
	for i, v := range input {
		got[i] = CalcInvLdData(v)
	}
	assertFixpDBLSlice(t, "CalcInvLdData", got[:], want[:], 0x335c2d95e0481cab)
}

func TestCalcLdIntVectors(t *testing.T) {
	if len(ldIntCoeff) != 193 {
		t.Fatalf("ldIntCoeff length = %d, want 193", len(ldIntCoeff))
	}
	if ldIntCoeff[0] != MinValDBL+1 || ldIntCoeff[192] != 0x0f2b8034 {
		t.Fatalf("ldIntCoeff edge entries = %d/%d", ldIntCoeff[0], ldIntCoeff[192])
	}

	input := [...]int{0, 1, 2, 3, 4, 8, 64, 120, 192, 193}
	want := [...]FixpDBL{0, 0, 33554432, 53182516, 67108864, 100663296, 201326592, 231756791, 254509108, 0}
	var got [len(input)]FixpDBL
	for i, v := range input {
		got[i] = CalcLdInt(v)
	}
	assertFixpDBLSlice(t, "CalcLdInt", got[:], want[:], 0x5548a115d7d858d0)
}

func TestFDivNormVectors(t *testing.T) {
	input := [...]struct {
		num   FixpDBL
		denom FixpDBL
	}{
		{0, 9},
		{1, 2},
		{3, 4},
		{5, 8},
		{7, 9},
		{123, 456},
		{1023, 1024},
		{0x10000000, 0x40000000},
	}
	want := [...]FixpDBL{0, 1073741824, 1610612736, 1342177280, 1670250496, 579239936, 2145386496, 536870912}
	var got [len(input)]FixpDBL
	for i, tt := range input {
		got[i] = fDivNorm(tt.num, tt.denom)
	}
	if got != want {
		t.Fatalf("fDivNorm = %v, want %v", got, want)
	}
	assertFixpDBLSlice(t, "fDivNorm", got[:], want[:], 0x58a03ac09c95c12f)
}

func TestLdDataVectorRejectsInvalid(t *testing.T) {
	var src [2]FixpDBL
	var dst [2]FixpDBL
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "negative fDivNorm numerator",
			fn: func() {
				fDivNorm(-1, 2)
			},
		},
		{
			name: "zero fDivNorm denominator",
			fn: func() {
				fDivNorm(1, 0)
			},
		},
		{
			name: "fDivNorm numerator larger than denominator",
			fn: func() {
				fDivNorm(3, 2)
			},
		},
		{
			name: "bad Schur division count",
			fn: func() {
				schurDiv(1, 2, 0)
			},
		},
		{
			name: "negative count",
			fn: func() {
				LdDataVector(src[:], dst[:], -1)
			},
		},
		{
			name: "short source",
			fn: func() {
				LdDataVector(src[:1], dst[:], 2)
			},
		},
		{
			name: "short destination",
			fn: func() {
				LdDataVector(src[:], dst[:1], 2)
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

func TestCalcLdDataAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		input := [...]FixpDBL{0, 1, 0x100000, 0x10000000, MaxValDBL}
		var output [len(input)]FixpDBL
		LdDataVector(input[:], output[:], len(output))
		ldDataSink = output[0] + output[1] + output[2] + output[3] + output[4] + CalcInvLdData(-1)
	})
	if allocs != 0 {
		t.Fatalf("LdDataVector allocations = %v, want 0", allocs)
	}
}
