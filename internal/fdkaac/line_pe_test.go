package fdkaac

import "testing"

var linePeSink FixpDBL
var linePeHashSink uint64

func TestFDKaacEncPrepareSfbPeVectors(t *testing.T) {
	energy, threshold, form, offsets, _, _ := linePeVectorInputs()
	var pe PEChannelData
	fillLinePeSentinels(&pe)

	FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], offsets[:], 8, 4, 3)

	want := [...]int{3, 0, 2, 99, 0, 1, 3, 99}
	assertIntSlice(t, "prepared PE nlines", pe.SfbNLines[:8], want[:], 0x3d1001271f038d96)
}

func TestFDKaacEncCalcSfbPeVectors(t *testing.T) {
	energy, threshold, form, offsets, isBook, isScale := linePeVectorInputs()
	var pe PEChannelData
	fillLinePeSentinels(&pe)

	FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], offsets[:], 8, 4, 3)
	FDKaacEncCalcSfbPe(&pe, energy[:], threshold[:], 8, 4, 3, isBook[:], isScale[:])

	wantPe := [...]FixpDBL{325451, 0, 546875, -77, 0, 253906, 292676, -77}
	wantConstPart := [...]FixpDBL{-723344, 0, -1093750, -88, 0, -468750, -395596, -88}
	wantActiveLines := [...]FixpDBL{2, 0, 2, -99, 0, 1, 2, -99}
	wantTotals := [...]FixpDBL{21, -41, 7}

	assertFixpDBLSlice(t, "sfb PE", pe.SfbPe[:8], wantPe[:], 0xe5275f6030e5dfdd)
	assertFixpDBLSlice(t, "sfb const PE", pe.SfbConstPart[:8], wantConstPart[:], 0xef5c3379db22b82a)
	assertFixpDBLSlice(t, "sfb active lines", pe.SfbNActiveLines[:8], wantActiveLines[:], 0x618ad57c274fa72e)
	gotTotals := [...]FixpDBL{pe.Pe, pe.ConstPart, pe.NActiveLines}
	assertFixpDBLSlice(t, "PE totals", gotTotals[:], wantTotals[:], 0x86fb9984ea4993a3)
}

func TestFDKaacEncLinePeRejectsInvalid(t *testing.T) {
	energy, threshold, form, offsets, isBook, isScale := linePeVectorInputs()
	var pe PEChannelData

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "nil prepare channel", fn: func() {
			FDKaacEncPrepareSfbPe(nil, energy[:], threshold[:], form[:], offsets[:], 8, 4, 3)
		}},
		{name: "bad prepare count", fn: func() {
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], offsets[:], 0, 4, 3)
		}},
		{name: "bad prepare group multiple", fn: func() {
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], offsets[:], 8, 3, 2)
		}},
		{name: "bad prepare group width", fn: func() {
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], offsets[:], 8, 4, 5)
		}},
		{name: "short prepare energy", fn: func() {
			FDKaacEncPrepareSfbPe(&pe, energy[:7], threshold[:], form[:], offsets[:], 8, 4, 3)
		}},
		{name: "short prepare threshold", fn: func() {
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:7], form[:], offsets[:], 8, 4, 3)
		}},
		{name: "short prepare form", fn: func() {
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:7], offsets[:], 8, 4, 3)
		}},
		{name: "short prepare offsets", fn: func() {
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], offsets[:8], 8, 4, 3)
		}},
		{name: "negative offset", fn: func() {
			bad := offsets
			bad[0] = -1
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], bad[:], 8, 4, 3)
		}},
		{name: "decreasing offset", fn: func() {
			bad := offsets
			bad[2] = bad[1] - 1
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], bad[:], 8, 4, 3)
		}},
		{name: "empty active band", fn: func() {
			bad := offsets
			bad[1] = bad[0]
			FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], bad[:], 8, 4, 3)
		}},
		{name: "nil calc channel", fn: func() {
			FDKaacEncCalcSfbPe(nil, energy[:], threshold[:], 8, 4, 3, isBook[:], isScale[:])
		}},
		{name: "short calc energy", fn: func() {
			FDKaacEncCalcSfbPe(&pe, energy[:7], threshold[:], 8, 4, 3, isBook[:], isScale[:])
		}},
		{name: "short calc threshold", fn: func() {
			FDKaacEncCalcSfbPe(&pe, energy[:], threshold[:7], 8, 4, 3, isBook[:], isScale[:])
		}},
		{name: "short is book", fn: func() {
			FDKaacEncCalcSfbPe(&pe, energy[:], threshold[:], 8, 4, 3, isBook[:7], isScale[:])
		}},
		{name: "short is scale", fn: func() {
			FDKaacEncCalcSfbPe(&pe, energy[:], threshold[:], 8, 4, 3, isBook[:], isScale[:7])
		}},
		{name: "intensity delta out of range", fn: func() {
			var badBook [8]int
			var badScale [8]int
			badBook[1] = 1
			badScale[1] = 61
			FDKaacEncCalcSfbPe(&pe, energy[:], threshold[:], 8, 4, 3, badBook[:], badScale[:])
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

func TestFDKaacEncLinePeAllocs(t *testing.T) {
	energy, threshold, form, offsets, isBook, isScale := linePeVectorInputs()
	var pe PEChannelData

	allocs := testing.AllocsPerRun(1000, func() {
		fillLinePeSentinels(&pe)
		FDKaacEncPrepareSfbPe(&pe, energy[:], threshold[:], form[:], offsets[:], 8, 4, 3)
		FDKaacEncCalcSfbPe(&pe, energy[:], threshold[:], 8, 4, 3, isBook[:], isScale[:])
		linePeSink = pe.Pe + pe.ConstPart + pe.NActiveLines + pe.SfbPe[0] + pe.SfbConstPart[6]
		linePeHashSink = hashFixpDBL(pe.SfbPe[:8])
	})
	if allocs != 0 {
		t.Fatalf("line PE allocations = %v, want 0", allocs)
	}
}

func linePeVectorInputs() (
	[8]FixpDBL,
	[8]FixpDBL,
	[8]FixpDBL,
	[9]int,
	[8]int,
	[8]int,
) {
	energy := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000, -260000000, -240000000, -200000000, -330000000}
	threshold := [...]FixpDBL{-320000000, -220000000, -420000000, -330000000, -250000000, -370000000, -210000000, -210000000}
	form := [...]FixpDBL{-330220071, -277590868, -332606544, -282650384, -338598629, -274335045, -293504191, -323306888}
	offsets := [...]int{0, 3, 7, 9, 14, 16, 21, 24, 26}
	isBook := [...]int{0, 1, 0, 0, 1, 0, 0, 0}
	isScale := [...]int{0, 5, 0, 0, 9, 0, 0, 0}
	return energy, threshold, form, offsets, isBook, isScale
}

func fillLinePeSentinels(pe *PEChannelData) {
	for i := range pe.SfbNLines {
		pe.SfbNLines[i] = 99
		pe.SfbPe[i] = -77
		pe.SfbConstPart[i] = -88
		pe.SfbNActiveLines[i] = -99
	}
	pe.Pe = -1
	pe.ConstPart = -2
	pe.NActiveLines = -3
}
