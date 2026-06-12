package fdkaac

import "testing"

var dctSink FixpDBL
var dctScaleSink int
var dctWindowSink []FixpSPK

func TestDCTGetTablesRadix2AACLC(t *testing.T) {
	tests := []struct {
		length  int
		window  []FixpSPK
		sinStep int
	}{
		{length: 128, window: SineWindow128[:], sinStep: 16},
		{length: 1024, window: SineWindow1024[:], sinStep: 2},
	}
	for _, tt := range tests {
		twiddle, sinTwiddle, sinStep := dctGetTables(tt.length)
		if !sameWindowTable(twiddle, tt.window) {
			t.Fatalf("dctGetTables(%d) returned unexpected twiddle table", tt.length)
		}
		if !sameWindowTable(sinTwiddle, SineTable1024[:]) {
			t.Fatalf("dctGetTables(%d) returned unexpected sine table", tt.length)
		}
		if sinStep != tt.sinStep {
			t.Fatalf("dctGetTables(%d) sin step = %d, want %d", tt.length, sinStep, tt.sinStep)
		}
	}
}

func TestDCTIVRadix2AACLCVectors(t *testing.T) {
	// Expected hashes are source-derived from pinned FDK-AAC v2.0.3 dct_IV
	// equations, AAC-LC window slopes, SINETABLE_16BIT, and radix-2 FFT.
	tests := []struct {
		length     int
		scaleDelta int
		hash       uint64
	}{
		{length: 128, scaleDelta: fftScaleFactor64 + 2, hash: 0x5689d357fa6f295e},
		{length: 1024, scaleDelta: fftScaleFactor512 + 2, hash: 0x1ec3ff46cba4b5da},
	}
	for _, tt := range tests {
		x := make([]FixpDBL, tt.length)
		fillDCTIVInput(x)
		scale := 3
		DCTIV(x, tt.length, &scale)
		if got := hashFixpDBL(x); got != tt.hash {
			t.Fatalf("DCTIV length %d hash = %#016x, want %#016x", tt.length, got, tt.hash)
		}
		if scale != 3+tt.scaleDelta {
			t.Fatalf("DCTIV length %d scale = %d, want %d", tt.length, scale, 3+tt.scaleDelta)
		}
	}
}

func TestDCTIVRejectsUnsupportedLengths(t *testing.T) {
	for _, length := range []int{0, 2, 4, 64, 120, 256, 512, 960} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("DCTIV length %d did not panic", length)
				}
			}()
			var x [1024]FixpDBL
			scale := 0
			DCTIV(x[:], length, &scale)
		}()
	}

	func() {
		defer func() {
			if recover() == nil {
				t.Fatalf("DCTIV accepted a short slice")
			}
		}()
		var x [127]FixpDBL
		scale := 0
		DCTIV(x[:], 128, &scale)
	}()
}

func TestDCTGetTablesAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		dctWindowSink, _, dctScaleSink = dctGetTables(1024)
	})
	if allocs != 0 {
		t.Fatalf("dctGetTables allocations = %v, want 0", allocs)
	}
}

func TestDCTIVAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var x [128]FixpDBL
		fillDCTIVInput(x[:])
		scale := 3
		DCTIV(x[:], 128, &scale)
		dctSink = x[0]
		dctScaleSink = scale
	})
	if allocs != 0 {
		t.Fatalf("DCTIV allocations = %v, want 0", allocs)
	}
}

func fillDCTIVInput(x []FixpDBL) {
	for i := range x {
		x[i] = FixpDBL(((i % 29) - 14) * 0x00040000)
	}
}
