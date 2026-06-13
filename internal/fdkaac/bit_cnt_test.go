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
