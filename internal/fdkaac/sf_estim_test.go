package fdkaac

import "testing"

var formFactorSink FixpDBL
var formFactorHashSink uint64
var relevantLinesSink FixpDBL
var relevantLinesHashSink uint64

func TestFDKaacEncCalcFormFactorLongVectors(t *testing.T) {
	var qc QCOutChannel
	var psy PsyOutChannel
	fillLongFormFactorInput(&qc, &psy)

	FDKaacEncCalcFormFactor([]*QCOutChannel{&qc}, []*PsyOutChannel{&psy}, 1)

	want := [...]FixpDBL{
		-330220071,
		-277590868,
		-332606544,
		-282650384,
		-338598629,
		-274335045,
		MinValDBL,
		MinValDBL,
	}
	assertFixpDBLSlice(t, "long form factor", qc.SfbFormFactorLdData[:psy.SfbCnt], want[:], 0x9e129645cc524391)
}

func TestFDKaacEncCalcFormFactorShortVectors(t *testing.T) {
	var qc QCOutChannel
	var psy PsyOutChannel
	fillShortFormFactorInput(&qc, &psy)

	FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:], &psy)

	want := [...]FixpDBL{
		-384653471,
		-315286532,
		-286561148,
		MinValDBL,
		-291398406,
		-305775759,
		-293504191,
		MinValDBL,
		-338330422,
		-323306888,
		-296050596,
		MinValDBL,
	}
	assertFixpDBLSlice(t, "short form factor", qc.SfbFormFactorLdData[:psy.SfbCnt], want[:], 0xecf998598cb5b807)
}

func TestFDKaacEncCalcSfbRelevantLinesLongVectors(t *testing.T) {
	form := [...]FixpDBL{-330220071, -277590868, -332606544, -282650384, -338598629, -274335045, MinValDBL, MinValDBL}
	energy := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000, -260000000, -240000000, -200000000, -200000000}
	threshold := [...]FixpDBL{-320000000, -220000000, -300000000, -330000000, -250000000, -260000000, -210000000, -210000000}
	offsets := [...]int{0, 3, 7, 9, 14, 16, 21, 24, 26}
	lines := [...]FixpDBL{1, 2, 3, 4, 5, 6, 7, 8}

	FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], offsets[:], 8, 8, 6, lines[:])

	want := [...]FixpDBL{7252715, 0, 5626260, 23182377, 0, 19176032, 0, 0}
	assertFixpDBLSlice(t, "long relevant lines", lines[:], want[:], 0x05e8cb5ca16f289e)
}

func TestFDKaacEncCalcSfbRelevantLinesShortVectors(t *testing.T) {
	form := [...]FixpDBL{-384653471, -315286532, -286561148, MinValDBL, -291398406, -305775759, -293504191, MinValDBL, -338330422, -323306888, -296050596, MinValDBL}
	energy := [...]FixpDBL{-390000000, -300000000, -260000000, -200000000, -290000000, -330000000, -280000000, -210000000, -340000000, -320000000, -260000000, -220000000}
	threshold := [...]FixpDBL{-410000000, -290000000, -270000000, -190000000, -300000000, -310000000, -300000000, -200000000, -330000000, -340000000, -250000000, -230000000}
	offsets := [...]int{0, 2, 5, 9, 12, 17, 20, 24, 28, 31, 33, 36, 40}
	lines := [...]FixpDBL{11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22}

	FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], offsets[:], 12, 4, 3, lines[:])

	want := [...]FixpDBL{3388439, 0, 15621102, 0, 17450949, 0, 15006516, 0, 0, 8382327, 0, 0}
	assertFixpDBLSlice(t, "short relevant lines", lines[:], want[:], 0x17456b49ebcd2b7d)
}

func TestFDKaacEncCalcFormFactorRejectsInvalid(t *testing.T) {
	var qc QCOutChannel
	var psy PsyOutChannel
	fillLongFormFactorInput(&qc, &psy)

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "bad channel count", fn: func() {
			FDKaacEncCalcFormFactor([]*QCOutChannel{&qc}, []*PsyOutChannel{&psy}, 2)
		}},
		{name: "nil qc output", fn: func() {
			FDKaacEncCalcFormFactor([]*QCOutChannel{nil}, []*PsyOutChannel{&psy}, 1)
		}},
		{name: "nil psy output", fn: func() {
			FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:], nil)
		}},
		{name: "short output", fn: func() {
			FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:psy.SfbCnt-1], &psy)
		}},
		{name: "bad band count", fn: func() {
			bad := psy
			bad.SfbCnt = 0
			FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:], &bad)
		}},
		{name: "bad group multiple", fn: func() {
			bad := psy
			bad.SfbPerGroup = 5
			FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:], &bad)
		}},
		{name: "bad group width", fn: func() {
			bad := psy
			bad.MaxSfbPerGroup = bad.SfbPerGroup + 1
			FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:], &bad)
		}},
		{name: "negative offset", fn: func() {
			bad := psy
			bad.SfbOffsets[0] = -1
			FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:], &bad)
		}},
		{name: "decreasing offset", fn: func() {
			bad := psy
			bad.SfbOffsets[4] = bad.SfbOffsets[3] - 1
			FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:], &bad)
		}},
		{name: "short spectrum", fn: func() {
			bad := psy
			bad.MdctSpectrum = bad.MdctSpectrum[:bad.SfbOffsets[bad.SfbCnt]-1]
			FDKaacEncCalcFormFactorChannel(qc.SfbFormFactorLdData[:], &bad)
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

func TestFDKaacEncCalcSfbRelevantLinesRejectsInvalid(t *testing.T) {
	form := [...]FixpDBL{-330220071, -277590868, -332606544, -282650384, -338598629, -274335045, MinValDBL, MinValDBL}
	energy := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000, -260000000, -240000000, -200000000, -200000000}
	threshold := [...]FixpDBL{-320000000, -220000000, -300000000, -330000000, -250000000, -260000000, -210000000, -210000000}
	offsets := [...]int{0, 3, 7, 9, 14, 16, 21, 24, 26}
	var lines [8]FixpDBL

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "bad count", fn: func() {
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], offsets[:], 0, 8, 6, lines[:])
		}},
		{name: "bad group multiple", fn: func() {
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], offsets[:], 8, 5, 4, lines[:])
		}},
		{name: "bad group width", fn: func() {
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], offsets[:], 8, 8, 9, lines[:])
		}},
		{name: "short form", fn: func() {
			FDKaacEncCalcSfbRelevantLines(form[:7], energy[:], threshold[:], offsets[:], 8, 8, 6, lines[:])
		}},
		{name: "short energy", fn: func() {
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:7], threshold[:], offsets[:], 8, 8, 6, lines[:])
		}},
		{name: "short threshold", fn: func() {
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:7], offsets[:], 8, 8, 6, lines[:])
		}},
		{name: "short output", fn: func() {
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], offsets[:], 8, 8, 6, lines[:7])
		}},
		{name: "short offsets", fn: func() {
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], offsets[:8], 8, 8, 6, lines[:])
		}},
		{name: "negative offset", fn: func() {
			bad := offsets
			bad[0] = -1
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], bad[:], 8, 8, 6, lines[:])
		}},
		{name: "decreasing offset", fn: func() {
			bad := offsets
			bad[3] = bad[2] - 1
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], bad[:], 8, 8, 6, lines[:])
		}},
		{name: "empty active band", fn: func() {
			bad := offsets
			bad[3] = bad[2]
			FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], bad[:], 8, 8, 6, lines[:])
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

func TestFDKaacEncCalcFormFactorAllocs(t *testing.T) {
	var qc QCOutChannel
	var psy PsyOutChannel
	qcChannels := []*QCOutChannel{&qc}
	psyChannels := []*PsyOutChannel{&psy}

	allocs := testing.AllocsPerRun(1000, func() {
		fillShortFormFactorInput(&qc, &psy)
		FDKaacEncCalcFormFactor(qcChannels, psyChannels, 1)
		formFactorSink = qc.SfbFormFactorLdData[0] + qc.SfbFormFactorLdData[6] + qc.SfbFormFactorLdData[11]
		formFactorHashSink = hashFixpDBL(qc.SfbFormFactorLdData[:psy.SfbCnt])
	})
	if allocs != 0 {
		t.Fatalf("form-factor allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncCalcSfbRelevantLinesAllocs(t *testing.T) {
	form := [...]FixpDBL{-384653471, -315286532, -286561148, MinValDBL, -291398406, -305775759, -293504191, MinValDBL, -338330422, -323306888, -296050596, MinValDBL}
	energy := [...]FixpDBL{-390000000, -300000000, -260000000, -200000000, -290000000, -330000000, -280000000, -210000000, -340000000, -320000000, -260000000, -220000000}
	threshold := [...]FixpDBL{-410000000, -290000000, -270000000, -190000000, -300000000, -310000000, -300000000, -200000000, -330000000, -340000000, -250000000, -230000000}
	offsets := [...]int{0, 2, 5, 9, 12, 17, 20, 24, 28, 31, 33, 36, 40}
	var lines [12]FixpDBL

	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncCalcSfbRelevantLines(form[:], energy[:], threshold[:], offsets[:], 12, 4, 3, lines[:])
		relevantLinesSink = lines[0] + lines[4] + lines[9]
		relevantLinesHashSink = hashFixpDBL(lines[:])
	})
	if allocs != 0 {
		t.Fatalf("relevant-lines allocations = %v, want 0", allocs)
	}
}

func fillLongFormFactorInput(qc *QCOutChannel, psy *PsyOutChannel) {
	spec := [...]FixpDBL{
		0, 2109497, 3170418, 4231339, -5292260, 6353181, 7348566, 0,
		-9470408, 1094145, 2155066, 3150451, -4211372, 5272293, 0,
		7394135, -8389520, 9450441, 1074178, 2135099, -3196020, 0,
		5252326, 6313247, -7374168, 8435089,
	}
	offsets := [...]int{0, 3, 7, 9, 14, 16, 21, 24, 26}
	*psy = PsyOutChannel{
		SfbCnt:         8,
		SfbPerGroup:    8,
		MaxSfbPerGroup: 6,
		MdctSpectrum:   qc.MdctSpectrum[:],
	}
	copy(qc.MdctSpectrum[:], spec[:])
	copy(psy.SfbOffsets[:], offsets[:])
}

func fillShortFormFactorInput(qc *QCOutChannel, psy *PsyOutChannel) {
	spec := [...]FixpDBL{
		0, -1102897, 1681506, 2129043, -2707652, 3155189, 3733798, -4312407,
		4759944, 5338553, 0, 6364699, 6943308, -575101, 1153710, 1601247,
		-2179856, 2627393, 3206002, -3784611, 0, 4810757, -5258294, 5836903,
		6415512, -6863049, 625914, 1073451, -1652060, 2099597, 0, -3256815,
		3704352, 4282961, -4730498, 5309107, 5887716, -6335253, 6913862, 545655,
	}
	offsets := [...]int{0, 2, 5, 9, 12, 17, 20, 24, 28, 31, 33, 36, 40}
	*psy = PsyOutChannel{
		SfbCnt:         12,
		SfbPerGroup:    4,
		MaxSfbPerGroup: 3,
		MdctSpectrum:   qc.MdctSpectrum[:],
	}
	copy(qc.MdctSpectrum[:], spec[:])
	copy(psy.SfbOffsets[:], offsets[:])
}
