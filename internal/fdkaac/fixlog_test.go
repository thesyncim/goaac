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

func TestLdDataVectorRejectsInvalid(t *testing.T) {
	var src [2]FixpDBL
	var dst [2]FixpDBL
	tests := []struct {
		name string
		fn   func()
	}{
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
		ldDataSink = output[0] + output[1] + output[2] + output[3] + output[4]
	})
	if allocs != 0 {
		t.Fatalf("LdDataVector allocations = %v, want 0", allocs)
	}
}
