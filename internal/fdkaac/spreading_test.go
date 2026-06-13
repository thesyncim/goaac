package fdkaac

import "testing"

var spreadingSink FixpDBL

func TestFDKaacEncSpreadingMaxVectors(t *testing.T) {
	maskLow := [...]FixpDBL{
		0x00000000, 0x50000000, 0x4a000000, 0x42000000,
		0x38000000, 0x30000000, 0x28000000, 0x20000000,
	}
	maskHigh := [...]FixpDBL{
		0x00000000, 0x30000000, 0x36000000, 0x40000000,
		0x48000000, 0x50000000, 0x58000000, 0x60000000,
	}
	energy := [...]FixpDBL{
		0x01000000, 0x00200000, 0x04000000, 0x00100000,
		0x00600000, 0x08000000, 0x00300000, 0x00280000,
	}

	FDKaacEncSpreadingMax(len(energy), maskLow[:], maskHigh[:], energy[:])

	want := [...]FixpDBL{16777216, 41943040, 67108864, 33554432, 58720256, 134217728, 92274688, 69206016}
	if energy != want {
		t.Fatalf("spread energy = %v, want %v", energy, want)
	}
	if got, wantHash := hashFixpDBL(energy[:]), uint64(0x780425f1a5ee9668); got != wantHash {
		t.Fatalf("spread energy hash = %#016x, want %#016x", got, wantHash)
	}
}

func TestFDKaacEncSpreadingMaxRejectsInvalid(t *testing.T) {
	var x [2]FixpDBL
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "zero bands",
			fn: func() {
				FDKaacEncSpreadingMax(0, x[:], x[:], x[:])
			},
		},
		{
			name: "negative bands",
			fn: func() {
				FDKaacEncSpreadingMax(-1, x[:], x[:], x[:])
			},
		},
		{
			name: "short low factors",
			fn: func() {
				FDKaacEncSpreadingMax(2, x[:1], x[:], x[:])
			},
		},
		{
			name: "short high factors",
			fn: func() {
				FDKaacEncSpreadingMax(2, x[:], x[:1], x[:])
			},
		},
		{
			name: "short energy",
			fn: func() {
				FDKaacEncSpreadingMax(2, x[:], x[:], x[:1])
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

func TestFDKaacEncSpreadingMaxAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		maskLow := [...]FixpDBL{0, 0x50000000, 0x4a000000, 0x42000000, 0x38000000, 0x30000000, 0x28000000, 0x20000000}
		maskHigh := [...]FixpDBL{0, 0x30000000, 0x36000000, 0x40000000, 0x48000000, 0x50000000, 0x58000000, 0x60000000}
		energy := [...]FixpDBL{0x01000000, 0x00200000, 0x04000000, 0x00100000, 0x00600000, 0x08000000, 0x00300000, 0x00280000}
		FDKaacEncSpreadingMax(len(energy), maskLow[:], maskHigh[:], energy[:])
		spreadingSink = energy[0] + energy[7]
	})
	if allocs != 0 {
		t.Fatalf("FDKaacEncSpreadingMax allocations = %v, want 0", allocs)
	}
}
