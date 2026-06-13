package fdkaac

import "testing"

var sqrtFixpSink FixpDBL
var invSqrtShiftSink int

func TestSqrtFixpVectors(t *testing.T) {
	input := [...]FixpDBL{
		0,
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
		0,
		46340,
		11863280,
		47453132,
		189812528,
		759250124,
		1518500248,
		1859775392,
		2147483644,
	}

	var got [len(input)]FixpDBL
	for i, v := range input {
		got[i] = sqrtFixp(v)
	}
	assertFixpDBLSlice(t, "sqrtFixp", got[:], want[:], 0x43fd5f8114cec8d8)
}

func TestInvSqrtNorm2Vectors(t *testing.T) {
	if len(invSqrtTab) != 130 {
		t.Fatalf("len(invSqrtTab) = %d, want 130", len(invSqrtTab))
	}

	input := [...]FixpDBL{
		0,
		1,
		0x00010000,
		0x00100000,
		0x01000000,
		0x10000000,
		0x40000000,
		0x60000000,
		MaxValDBL,
	}
	wantMant := [...]FixpDBL{
		MaxValDBL,
		1518500249,
		1518500249,
		1518500249,
		1518500249,
		1518500249,
		1518500249,
		1239850262,
		1073741823,
	}
	wantShift := [...]int{16, 16, 8, 6, 4, 2, 1, 1, 1}

	var gotMant [len(input)]FixpDBL
	var gotShift [len(input)]int
	for i, v := range input {
		gotMant[i] = invSqrtNorm2(v, &gotShift[i])
	}
	assertFixpDBLSlice(t, "invSqrtNorm2 mantissa", gotMant[:], wantMant[:], 0x67ed0063aecc6e59)
	assertIntSlice(t, "invSqrtNorm2 shift", gotShift[:], wantShift[:], 0x38f1fec28ce6c1fc)
}

func TestSqrtFixpRejectsInvalid(t *testing.T) {
	shift := 0
	tests := []struct {
		name string
		fn   func()
	}{
		{name: "negative sqrt", fn: func() {
			sqrtFixp(-1)
		}},
		{name: "negative inverse sqrt", fn: func() {
			invSqrtNorm2(-1, &shift)
		}},
		{name: "nil inverse sqrt shift", fn: func() {
			invSqrtNorm2(1, nil)
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

func TestSqrtFixpAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		input := [...]FixpDBL{0, 1, 0x10000, 0x100000, 0x10000000, MaxValDBL}
		sum := FixpDBL(0)
		shiftSum := 0
		for _, v := range input {
			shift := 0
			sum += sqrtFixp(v)
			sum += invSqrtNorm2(v, &shift)
			shiftSum += shift
		}
		sqrtFixpSink = sum
		invSqrtShiftSink = shiftSum
	})
	if allocs != 0 {
		t.Fatalf("sqrtFixp allocations = %v, want 0", allocs)
	}
}
