package fdkaac

import "testing"

var formFactorSink FixpDBL
var formFactorHashSink uint64
var relevantLinesSink FixpDBL
var relevantLinesHashSink uint64
var scfPeSink FixpDBL
var scfPeHashSink uint64

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

func TestFDKaacEncCountSingleScfBitsVectors(t *testing.T) {
	input := [...]struct {
		scf      int
		scfLeft  int
		scfRight int
	}{
		{scf: 10, scfLeft: 15, scfRight: 8},
		{scf: 45, scfLeft: 0, scfRight: 60},
		{scf: 0, scfLeft: 60, scfRight: -60},
		{scf: 30, scfLeft: 30, scfRight: 30},
		{scf: 24, scfLeft: -12, scfRight: 54},
	}
	want := [...]FixpDBL{1310720, 3932160, 4980736, 262144, 4063232}

	var got [len(input)]FixpDBL
	for i, tt := range input {
		got[i] = FDKaacEncCountSingleScfBits(tt.scf, tt.scfLeft, tt.scfRight)
	}
	assertFixpDBLSlice(t, "single scalefactor bits", got[:], want[:], 0xebd971fc74afaf8b)
}

func TestFDKaacEncCalcSingleSpecPeVectors(t *testing.T) {
	input := [...]struct {
		scf            int
		sfbConstPePart FixpDBL
		nLines         FixpDBL
	}{
		{scf: 2, sfbConstPePart: 0x08000000, nLines: 7252715},
		{scf: 7, sfbConstPePart: 0x04000000, nLines: 19176032},
		{scf: 0, sfbConstPePart: 0x01500000, nLines: 15006516},
		{scf: 12, sfbConstPePart: 0x09000000, nLines: 3388439},
		{scf: 4, sfbConstPePart: 0x03000000, nLines: 8382327},
	}
	want := [...]FixpDBL{287558, 219285, 168736, 83386, 99059}

	var got [len(input)]FixpDBL
	for i, tt := range input {
		got[i] = FDKaacEncCalcSingleSpecPe(tt.scf, tt.sfbConstPePart, tt.nLines)
	}
	assertFixpDBLSlice(t, "single spec pe", got[:], want[:], 0x7f9608da2bb13e33)
}

func TestFDKaacEncCountScfBitsDiffVectors(t *testing.T) {
	const min = fdkIntMin
	scfOld := [...]int{12, min, 8, 11, min, 14, 18, min, 16, 20}
	scfNew := [...]int{13, min, 10, 9, min, 15, 17, min, 14, 19}
	input := [...]struct {
		startSfb int
		stopSfb  int
	}{
		{startSfb: 2, stopSfb: 8},
		{startSfb: 0, stopSfb: 4},
		{startSfb: 4, stopSfb: 7},
		{startSfb: 7, stopSfb: 10},
		{startSfb: 1, stopSfb: 3},
	}
	want := [...]FixpDBL{-131072, 0, 131072, 131072, -262144}

	var got [len(input)]FixpDBL
	for i, tt := range input {
		got[i] = FDKaacEncCountScfBitsDiff(scfOld[:], scfNew[:], len(scfOld), tt.startSfb, tt.stopSfb)
	}
	assertFixpDBLSlice(t, "scalefactor bits diff", got[:], want[:], 0xe6c995131d2b80ef)
}

func TestFDKaacEncCalcSpecPeDiffVectors(t *testing.T) {
	const min = fdkIntMin
	energy := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000, -260000000, -240000000, -200000000, -330000000}
	form := [...]FixpDBL{-330220071, -277590868, -332606544, -282650384, -338598629, -274335045, -293504191, -323306888}
	lines := [...]FixpDBL{7252715, 0, 5626260, 23182377, 0, 19176032, 15006516, 8382327}
	scfOld := [...]int{12, min, 8, 11, min, 14, 18, 16}
	scfNew := [...]int{13, min, 10, 9, min, 15, 17, 14}
	constPart := [...]FixpDBL{
		FixpDBL(min),
		0x04000000,
		0x03000000,
		FixpDBL(min),
		0x02000000,
		FixpDBL(min),
		0x20000000,
		FixpDBL(min),
	}

	got := FDKaacEncCalcSpecPeDiff(
		energy[:], scfOld[:], scfNew[:], constPart[:], form[:], lines[:], 0, len(scfOld),
	)
	if got != 59962 {
		t.Fatalf("spec PE diff = %d, want 59962", got)
	}
	wantConstPart := [...]FixpDBL{
		-39333917,
		0x04000000,
		0x03000000,
		-68118760,
		0x02000000,
		-37276430,
		0x20000000,
		-57790508,
	}
	assertFixpDBLSlice(t, "spec PE const part", constPart[:], wantConstPart[:], 0x2215c451354b2881)
}

func TestFDKaacEncImproveScfVectors(t *testing.T) {
	spec := improveScfVectorSpec()
	tests := []struct {
		name          string
		thresh        FixpDBL
		scf           int
		minScf        int
		dz            int
		wantBest      int
		wantDist      FixpDBL
		wantMinCalc   int
		wantQuant     [7]int16
		wantTmp       [7]int16
		wantQuantHash uint64
		wantTmpHash   uint64
	}{
		{
			name:          "upward accepted under allowed distortion",
			thresh:        -350000000,
			scf:           -22,
			minScf:        -24,
			wantBest:      -21,
			wantDist:      -402731731,
			wantMinCalc:   -21,
			wantQuant:     [...]int16{0, 0, 0, 0, 1, -2, 3},
			wantTmp:       [...]int16{0, 0, 0, 0, 1, -2, 3},
			wantQuantHash: 0xf4f4754f5f8670f2,
			wantTmpHash:   0xf4f4754f5f8670f2,
		},
		{
			name:          "downward accepted after noisy threshold",
			thresh:        -420000000,
			scf:           -20,
			minScf:        -24,
			wantBest:      -21,
			wantDist:      -402731731,
			wantMinCalc:   -21,
			wantQuant:     [...]int16{0, 0, 0, 0, 1, -2, 3},
			wantTmp:       [...]int16{0, 0, 0, 0, 1, -2, 3},
			wantQuantHash: 0xf4f4754f5f8670f2,
			wantTmpHash:   0xf4f4754f5f8670f2,
		},
		{
			name:          "downward copy changes quantized band",
			thresh:        -380000000,
			scf:           -16,
			minScf:        -20,
			wantBest:      -17,
			wantDist:      -352671313,
			wantMinCalc:   -17,
			wantQuant:     [...]int16{0, 0, 0, 0, 1, -1, 2},
			wantTmp:       [...]int16{0, 0, 0, 0, 1, -1, 2},
			wantQuantHash: 0x7bd72c1d68467362,
			wantTmpHash:   0x7bd72c1d68467362,
		},
		{
			name:          "minimum scalefactor stops downward search",
			thresh:        -420000000,
			scf:           -20,
			minScf:        -20,
			wantBest:      -20,
			wantDist:      -373521274,
			wantMinCalc:   -20,
			wantQuant:     [...]int16{0, 0, 0, 0, 1, -2, 3},
			wantTmp:       [...]int16{0, 0, 0, 0, 1, -1, 2},
			wantQuantHash: 0xf4f4754f5f8670f2,
			wantTmpHash:   0x7bd72c1d68467362,
		},
		{
			name:          "dead-zone quantization path",
			thresh:        -360000000,
			scf:           -18,
			minScf:        -21,
			dz:            1,
			wantBest:      -18,
			wantDist:      -347292759,
			wantMinCalc:   -19,
			wantQuant:     [...]int16{0, 0, 0, 0, 1, -1, 2},
			wantTmp:       [...]int16{0, 0, 0, 0, 1, -1, 2},
			wantQuantHash: 0x7bd72c1d68467362,
			wantTmpHash:   0x7bd72c1d68467362,
		},
	}

	for _, tt := range tests {
		var quant [7]int16
		var quantTmp [7]int16
		fillImproveScfQuant(quant[:], 77)
		fillImproveScfQuant(quantTmp[:], -77)
		dist := FixpDBL(123)
		minCalc := 456

		gotBest := FDKaacEncImproveScf(spec[:], quant[:], quantTmp[:], len(spec), tt.thresh, tt.scf, tt.minScf, &dist, &minCalc, tt.dz)
		if gotBest != tt.wantBest {
			t.Fatalf("%s best = %d, want %d", tt.name, gotBest, tt.wantBest)
		}
		if dist != tt.wantDist {
			t.Fatalf("%s dist = %d, want %d", tt.name, dist, tt.wantDist)
		}
		if minCalc != tt.wantMinCalc {
			t.Fatalf("%s minScfCalculated = %d, want %d", tt.name, minCalc, tt.wantMinCalc)
		}
		assertInt16Slice(t, tt.name+" quant", quant[:], tt.wantQuant[:], tt.wantQuantHash)
		assertInt16Slice(t, tt.name+" temp quant", quantTmp[:], tt.wantTmp[:], tt.wantTmpHash)
	}
}

func TestFDKaacEncAssimilateSingleScfVectors(t *testing.T) {
	const min = fdkIntMin
	tests := []struct {
		name              string
		restart           int
		scf               [5]int
		minScfCalculated  [5]int
		dist              [5]FixpDBL
		wantScf           [5]int
		wantMinCalculated [5]int
		wantDist          [5]FixpDBL
		wantConst         [5]FixpDBL
		wantQuant         [11]int16
		wantTmp           [11]int16
		hashScf           uint64
		hashMinCalculated uint64
		hashDist          uint64
		hashConst         uint64
		hashQuant         uint64
		hashTmp           uint64
	}{
		{
			name:              "single pass improves middle and trailing bands",
			scf:               [5]int{4, 6, 4, min, 5},
			minScfCalculated:  [5]int{4, 6, 4, min, 5},
			dist:              [5]FixpDBL{0, MaxValDBL, 0, 0, 0},
			wantScf:           [5]int{4, 5, 4, min, 4},
			wantMinCalculated: [5]int{4, 4, 4, min, 4},
			wantDist:          [5]FixpDBL{0, -458959796, 0, 0, -352594892},
			wantConst:         [5]FixpDBL{FixpDBL(min), -19333917, FixpDBL(min), FixpDBL(min), -39333917},
			wantQuant:         [11]int16{77, 77, 0, 0, 77, 77, 77, 77, 77, 0, 0},
			wantTmp:           [11]int16{-77, -77, 0, 0, -77, -77, -77, -77, -77, 0, 0},
			hashScf:           0x551b83ab62ffafb4,
			hashMinCalculated: 0xa41999737614c7f5,
			hashDist:          0x0d5abebe85fa03f08,
			hashConst:         0x7226cd060404a62e,
			hashQuant:         0xa70a552e0e56e678,
			hashTmp:           0x15d85de3fe8564b5,
		},
		{
			name:              "restart on success revisits earlier neighbours",
			restart:           1,
			scf:               [5]int{4, 7, 6, 4, 5},
			minScfCalculated:  [5]int{4, 7, 6, 4, 5},
			dist:              [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, 0},
			wantScf:           [5]int{4, 6, 5, 4, 4},
			wantMinCalculated: [5]int{4, 4, 4, 4, 4},
			wantDist:          [5]FixpDBL{0, -458959796, -188162553, 0, -352594892},
			wantConst:         [5]FixpDBL{FixpDBL(min), -19333917, -39333917, FixpDBL(min), -39333917},
			wantQuant:         [11]int16{77, 77, 0, 0, 0, 0, 0, 77, 77, 0, 0},
			wantTmp:           [11]int16{-77, -77, 0, 0, 0, 0, 0, -77, -77, 0, 0},
			hashScf:           0x7176680613cad242,
			hashMinCalculated: 0xc8bdc016486eda71,
			hashDist:          0x5caa914a4e91dd2d,
			hashConst:         0xfe3be15cad7d5dc0,
			hashQuant:         0x6dadcb034cbeb975,
			hashTmp:           0x446522701ec7e195,
		},
		{
			name:              "min calculated skips already checked band",
			scf:               [5]int{4, 6, 4, min, 5},
			minScfCalculated:  [5]int{4, 4, 4, min, 5},
			dist:              [5]FixpDBL{0, MaxValDBL, 0, 0, 0},
			wantScf:           [5]int{4, 6, 4, min, 4},
			wantMinCalculated: [5]int{4, 4, 4, min, 4},
			wantDist:          [5]FixpDBL{0, MaxValDBL, 0, 0, -352594892},
			wantConst:         [5]FixpDBL{FixpDBL(min), -19333917, FixpDBL(min), FixpDBL(min), -39333917},
			wantQuant:         [11]int16{77, 77, 77, 77, 77, 77, 77, 77, 77, 0, 0},
			wantTmp:           [11]int16{-77, -77, -77, -77, -77, -77, -77, -77, -77, 0, 0},
			hashScf:           0x4215c5039c3ef877,
			hashMinCalculated: 0xa41999737614c7f5,
			hashDist:          0x55e9dc04bed6088c,
			hashConst:         0x7226cd060404a62e,
			hashQuant:         0x1379514c462877a8,
			hashTmp:           0xe2b9b38b46ef21f5,
		},
		{
			name:              "flat scalefactors do not move",
			scf:               [5]int{4, 4, 4, min, 4},
			minScfCalculated:  [5]int{4, 4, 4, min, 4},
			dist:              [5]FixpDBL{0, MaxValDBL, 0, 0, MaxValDBL},
			wantScf:           [5]int{4, 4, 4, min, 4},
			wantMinCalculated: [5]int{4, 4, 4, min, 4},
			wantDist:          [5]FixpDBL{0, MaxValDBL, 0, 0, MaxValDBL},
			wantConst:         [5]FixpDBL{FixpDBL(min), FixpDBL(min), FixpDBL(min), FixpDBL(min), FixpDBL(min)},
			wantQuant:         [11]int16{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77},
			wantTmp:           [11]int16{-77, -77, -77, -77, -77, -77, -77, -77, -77, -77, -77},
			hashScf:           0xa41999737614c7f5,
			hashMinCalculated: 0xa41999737614c7f5,
			hashDist:          0xf12ac00f199b9d2d,
			hashConst:         0x1dd8a261589532b5,
			hashQuant:         0x53d5c4a2713d7b78,
			hashTmp:           0x488c9dabb335b135,
		},
	}

	for _, tt := range tests {
		var psy PsyOutChannel
		var qc QCOutChannel
		var quant [11]int16
		var quantTmp [11]int16
		var minScf [5]int
		var constPart [5]FixpDBL
		var form [5]FixpDBL
		var lines [5]FixpDBL
		setupAssimilateSingleVector(&psy, &qc, quant[:], quantTmp[:], minScf[:], constPart[:], form[:])
		scf := tt.scf
		minScfCalculated := tt.minScfCalculated
		dist := tt.dist

		FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], tt.restart)

		assertIntSlice(t, tt.name+" scf", scf[:], tt.wantScf[:], tt.hashScf)
		assertIntSlice(t, tt.name+" min calculated", minScfCalculated[:], tt.wantMinCalculated[:], tt.hashMinCalculated)
		assertFixpDBLSlice(t, tt.name+" distortion", dist[:], tt.wantDist[:], tt.hashDist)
		assertFixpDBLSlice(t, tt.name+" const PE", constPart[:], tt.wantConst[:], tt.hashConst)
		assertInt16Slice(t, tt.name+" quant", quant[:], tt.wantQuant[:], tt.hashQuant)
		assertInt16Slice(t, tt.name+" temp quant", quantTmp[:], tt.wantTmp[:], tt.hashTmp)
	}
}

func TestFDKaacEncAssimilateMultipleScfVectors(t *testing.T) {
	const min = fdkIntMin
	tests := []struct {
		name      string
		scf       [5]int
		minScf    [5]int
		dist      [5]FixpDBL
		threshold [5]FixpDBL
		wantScf   [5]int
		wantDist  [5]FixpDBL
		wantConst [5]FixpDBL
		wantQuant [11]int16
		wantTmp   [11]int16
		hashScf   uint64
		hashDist  uint64
		hashConst uint64
		hashQuant uint64
		hashTmp   uint64
	}{
		{
			name:      "region lowered to shared scalefactor",
			scf:       [5]int{4, 7, 6, min, 5},
			dist:      [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, MaxValDBL},
			threshold: [5]FixpDBL{0, 0, 0, 0, 0},
			wantScf:   [5]int{4, 4, 4, min, 4},
			wantDist:  [5]FixpDBL{0, -458959796, -188162553, 0, -352594892},
			wantConst: [5]FixpDBL{FixpDBL(min), -19333917, -39333917, FixpDBL(min), -39333917},
			wantQuant: [11]int16{77, 77, 0, 0, 0, 0, 0, 77, 77, 0, 0},
			wantTmp:   [11]int16{-77, -77, 0, 0, 0, 0, 0, -77, -77, 0, 0},
			hashScf:   0xa41999737614c7f5,
			hashDist:  0x5caa914a4e91dd2d,
			hashConst: 0xfe3be15cad7d5dc0,
			hashQuant: 0x6dadcb034cbeb975,
			hashTmp:   0x446522701ec7e195,
		},
		{
			name:      "threshold rejects region",
			scf:       [5]int{4, 7, 6, min, 5},
			dist:      [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, MaxValDBL},
			threshold: [5]FixpDBL{-500000000, -500000000, -500000000, -500000000, -500000000},
			wantScf:   [5]int{4, 7, 6, min, 5},
			wantDist:  [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, MaxValDBL},
			wantConst: [5]FixpDBL{FixpDBL(min), -19333917, -39333917, FixpDBL(min), -39333917},
			wantQuant: [11]int16{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77},
			wantTmp:   [11]int16{-77, -77, 0, 0, -77, -77, -77, -77, -77, -77, -77},
			hashScf:   0xf9a82985da58c865,
			hashDist:  0xa9fe0a9be26bce89,
			hashConst: 0xfe3be15cad7d5dc0,
			hashQuant: 0x53d5c4a2713d7b78,
			hashTmp:   0x2b31e8cbfd61b1f5,
		},
		{
			name:      "minimum scalefactors reject region",
			scf:       [5]int{4, 7, 6, min, 5},
			minScf:    [5]int{0, 7, 6, 0, 5},
			dist:      [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, MaxValDBL},
			threshold: [5]FixpDBL{0, 0, 0, 0, 0},
			wantScf:   [5]int{4, 7, 6, min, 5},
			wantDist:  [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, MaxValDBL},
			wantConst: [5]FixpDBL{FixpDBL(min), FixpDBL(min), FixpDBL(min), FixpDBL(min), FixpDBL(min)},
			wantQuant: [11]int16{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77},
			wantTmp:   [11]int16{-77, -77, -77, -77, -77, -77, -77, -77, -77, -77, -77},
			hashScf:   0xf9a82985da58c865,
			hashDist:  0xa9fe0a9be26bce89,
			hashConst: 0x1dd8a261589532b5,
			hashQuant: 0x53d5c4a2713d7b78,
			hashTmp:   0x488c9dabb335b135,
		},
		{
			name:      "irrelevant scalefactors do not move",
			scf:       [5]int{min, min, min, min, min},
			dist:      [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, MaxValDBL},
			threshold: [5]FixpDBL{0, 0, 0, 0, 0},
			wantScf:   [5]int{min, min, min, min, min},
			wantDist:  [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, MaxValDBL},
			wantConst: [5]FixpDBL{FixpDBL(min), FixpDBL(min), FixpDBL(min), FixpDBL(min), FixpDBL(min)},
			wantQuant: [11]int16{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77},
			wantTmp:   [11]int16{-77, -77, -77, -77, -77, -77, -77, -77, -77, -77, -77},
			hashScf:   0x1dd8a261589532b5,
			hashDist:  0xa9fe0a9be26bce89,
			hashConst: 0x1dd8a261589532b5,
			hashQuant: 0x53d5c4a2713d7b78,
			hashTmp:   0x488c9dabb335b135,
		},
	}

	for _, tt := range tests {
		var psy PsyOutChannel
		var qc QCOutChannel
		var quant [11]int16
		var quantTmp [11]int16
		var minScf [5]int
		var constPart [5]FixpDBL
		var form [5]FixpDBL
		var lines [5]FixpDBL
		setupAssimilateSingleVector(&psy, &qc, quant[:], quantTmp[:], minScf[:], constPart[:], form[:])
		if tt.minScf != ([5]int{}) {
			minScf = tt.minScf
		}
		copy(qc.SfbThresholdLdData[:], tt.threshold[:])
		scf := tt.scf
		dist := tt.dist

		FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])

		assertIntSlice(t, tt.name+" scf", scf[:], tt.wantScf[:], tt.hashScf)
		assertFixpDBLSlice(t, tt.name+" distortion", dist[:], tt.wantDist[:], tt.hashDist)
		assertFixpDBLSlice(t, tt.name+" const PE", constPart[:], tt.wantConst[:], tt.hashConst)
		assertInt16Slice(t, tt.name+" quant", quant[:], tt.wantQuant[:], tt.hashQuant)
		assertInt16Slice(t, tt.name+" temp quant", quantTmp[:], tt.wantTmp[:], tt.hashTmp)
	}
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

func TestFDKaacEncCountSingleScfBitsRejectsInvalid(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{name: "left delta out of range", fn: func() {
			FDKaacEncCountSingleScfBits(0, 61, 0)
		}},
		{name: "right delta out of range", fn: func() {
			FDKaacEncCountSingleScfBits(0, 0, -61)
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

func TestFDKaacEncScfPeDiffRejectsInvalid(t *testing.T) {
	const min = fdkIntMin
	scfOld := [...]int{12, min, 8, 11}
	scfNew := [...]int{13, min, 10, 9}
	energy := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000}
	form := [...]FixpDBL{-330220071, -277590868, -332606544, -282650384}
	lines := [...]FixpDBL{7252715, 0, 5626260, 23182377}
	constPart := [...]FixpDBL{FixpDBL(min), 0x04000000, 0x03000000, FixpDBL(min)}

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "bad count", fn: func() {
			FDKaacEncCountScfBitsDiff(scfOld[:], scfNew[:], 0, 0, 1)
		}},
		{name: "bad range", fn: func() {
			FDKaacEncCountScfBitsDiff(scfOld[:], scfNew[:], len(scfOld), 2, 2)
		}},
		{name: "stop past count", fn: func() {
			FDKaacEncCountScfBitsDiff(scfOld[:], scfNew[:], len(scfOld), 0, len(scfOld)+1)
		}},
		{name: "short old", fn: func() {
			FDKaacEncCountScfBitsDiff(scfOld[:3], scfNew[:], len(scfOld), 0, 2)
		}},
		{name: "short new", fn: func() {
			FDKaacEncCountScfBitsDiff(scfOld[:], scfNew[:3], len(scfOld), 0, 2)
		}},
		{name: "empty relevant range", fn: func() {
			FDKaacEncCountScfBitsDiff(scfOld[:], scfNew[:], len(scfOld), 1, 2)
		}},
		{name: "delta out of range", fn: func() {
			old := [...]int{0, 0}
			next := [...]int{0, 70}
			FDKaacEncCountScfBitsDiff(old[:], next[:], len(old), 0, len(old))
		}},
		{name: "bad spec range", fn: func() {
			FDKaacEncCalcSpecPeDiff(energy[:], scfOld[:], scfNew[:], constPart[:], form[:], lines[:], 0, 0)
		}},
		{name: "short energy", fn: func() {
			FDKaacEncCalcSpecPeDiff(energy[:3], scfOld[:], scfNew[:], constPart[:], form[:], lines[:], 0, len(scfOld))
		}},
		{name: "short old spec", fn: func() {
			FDKaacEncCalcSpecPeDiff(energy[:], scfOld[:3], scfNew[:], constPart[:], form[:], lines[:], 0, len(scfOld))
		}},
		{name: "short new spec", fn: func() {
			FDKaacEncCalcSpecPeDiff(energy[:], scfOld[:], scfNew[:3], constPart[:], form[:], lines[:], 0, len(scfOld))
		}},
		{name: "short const", fn: func() {
			FDKaacEncCalcSpecPeDiff(energy[:], scfOld[:], scfNew[:], constPart[:3], form[:], lines[:], 0, len(scfOld))
		}},
		{name: "short form", fn: func() {
			FDKaacEncCalcSpecPeDiff(energy[:], scfOld[:], scfNew[:], constPart[:], form[:3], lines[:], 0, len(scfOld))
		}},
		{name: "short lines", fn: func() {
			FDKaacEncCalcSpecPeDiff(energy[:], scfOld[:], scfNew[:], constPart[:], form[:], lines[:3], 0, len(scfOld))
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

func TestFDKaacEncImproveScfRejectsInvalid(t *testing.T) {
	spec := improveScfVectorSpec()
	var quant [7]int16
	var quantTmp [7]int16
	var dist FixpDBL
	var minCalc int

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "negative width", fn: func() {
			FDKaacEncImproveScf(spec[:], quant[:], quantTmp[:], -1, -350000000, -22, -24, &dist, &minCalc, 0)
		}},
		{name: "short spectrum", fn: func() {
			FDKaacEncImproveScf(spec[:6], quant[:], quantTmp[:], len(spec), -350000000, -22, -24, &dist, &minCalc, 0)
		}},
		{name: "short quant", fn: func() {
			FDKaacEncImproveScf(spec[:], quant[:6], quantTmp[:], len(spec), -350000000, -22, -24, &dist, &minCalc, 0)
		}},
		{name: "short temp quant", fn: func() {
			FDKaacEncImproveScf(spec[:], quant[:], quantTmp[:6], len(spec), -350000000, -22, -24, &dist, &minCalc, 0)
		}},
		{name: "nil dist", fn: func() {
			FDKaacEncImproveScf(spec[:], quant[:], quantTmp[:], len(spec), -350000000, -22, -24, nil, &minCalc, 0)
		}},
		{name: "nil min scalefactor", fn: func() {
			FDKaacEncImproveScf(spec[:], quant[:], quantTmp[:], len(spec), -350000000, -22, -24, &dist, nil, 0)
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

func TestFDKaacEncAssimilateSingleScfRejectsInvalid(t *testing.T) {
	var psy PsyOutChannel
	var qc QCOutChannel
	var quant [11]int16
	var quantTmp [11]int16
	var scf [5]int
	var minScf [5]int
	var dist [5]FixpDBL
	var constPart [5]FixpDBL
	var form [5]FixpDBL
	var lines [5]FixpDBL
	var minScfCalculated [5]int
	setupAssimilateSingleVector(&psy, &qc, quant[:], quantTmp[:], minScf[:], constPart[:], form[:])

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "nil psy", fn: func() {
			FDKaacEncAssimilateSingleScf(nil, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "nil qc", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, nil, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "bad band count", fn: func() {
			bad := psy
			bad.SfbCnt = 0
			FDKaacEncAssimilateSingleScf(&bad, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "bad group width", fn: func() {
			bad := psy
			bad.MaxSfbPerGroup = bad.SfbPerGroup + 1
			FDKaacEncAssimilateSingleScf(&bad, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "short scf", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:4], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "short min scf", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:4], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "short dist", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:4], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "short const", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:4], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "short form", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:4], lines[:], minScfCalculated[:], 0)
		}},
		{name: "short relevant lines", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:4], minScfCalculated[:], 0)
		}},
		{name: "short min calculated", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:4], 0)
		}},
		{name: "negative offset", fn: func() {
			bad := psy
			bad.SfbOffsets[0] = -1
			FDKaacEncAssimilateSingleScf(&bad, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "decreasing offset", fn: func() {
			bad := psy
			bad.SfbOffsets[3] = bad.SfbOffsets[2] - 1
			FDKaacEncAssimilateSingleScf(&bad, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "short quant", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:10], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
		}},
		{name: "short temp quant", fn: func() {
			FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:10], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 0)
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

func TestFDKaacEncAssimilateMultipleScfRejectsInvalid(t *testing.T) {
	var psy PsyOutChannel
	var qc QCOutChannel
	var quant [11]int16
	var quantTmp [11]int16
	var scf [5]int
	var minScf [5]int
	var dist [5]FixpDBL
	var constPart [5]FixpDBL
	var form [5]FixpDBL
	var lines [5]FixpDBL
	setupAssimilateSingleVector(&psy, &qc, quant[:], quantTmp[:], minScf[:], constPart[:], form[:])

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "nil psy", fn: func() {
			FDKaacEncAssimilateMultipleScf(nil, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "nil qc", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, nil, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "bad band count", fn: func() {
			bad := psy
			bad.SfbCnt = 0
			FDKaacEncAssimilateMultipleScf(&bad, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "bad group width", fn: func() {
			bad := psy
			bad.MaxSfbPerGroup = bad.SfbPerGroup + 1
			FDKaacEncAssimilateMultipleScf(&bad, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "short scf", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:4], minScf[:], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "short min scf", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:4], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "short dist", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:4], constPart[:], form[:], lines[:])
		}},
		{name: "short const", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:4], form[:], lines[:])
		}},
		{name: "short form", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:4], lines[:])
		}},
		{name: "short relevant lines", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:4])
		}},
		{name: "negative offset", fn: func() {
			bad := psy
			bad.SfbOffsets[0] = -1
			FDKaacEncAssimilateMultipleScf(&bad, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "decreasing offset", fn: func() {
			bad := psy
			bad.SfbOffsets[3] = bad.SfbOffsets[2] - 1
			FDKaacEncAssimilateMultipleScf(&bad, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "short quant", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:10], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
		}},
		{name: "short temp quant", fn: func() {
			FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:10], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
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

func TestFDKaacEncSingleScfPeAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var got [5]FixpDBL
		got[0] = FDKaacEncCountSingleScfBits(10, 15, 8)
		got[1] = FDKaacEncCountSingleScfBits(30, 30, 30)
		got[2] = FDKaacEncCalcSingleSpecPe(2, 0x08000000, 7252715)
		got[3] = FDKaacEncCalcSingleSpecPe(7, 0x04000000, 19176032)
		got[4] = FDKaacEncCalcSingleSpecPe(4, 0x03000000, 8382327)
		scfPeSink = got[0] + got[1] + got[2] + got[3] + got[4]
		scfPeHashSink = hashFixpDBL(got[:])
	})
	if allocs != 0 {
		t.Fatalf("single scalefactor/PE allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncImproveScfAllocs(t *testing.T) {
	spec := improveScfVectorSpec()
	var quant [7]int16
	var quantTmp [7]int16
	var dist FixpDBL
	var minCalc int

	allocs := testing.AllocsPerRun(1000, func() {
		fillImproveScfQuant(quant[:], 77)
		fillImproveScfQuant(quantTmp[:], -77)
		best := FDKaacEncImproveScf(spec[:], quant[:], quantTmp[:], len(spec), -350000000, -22, -24, &dist, &minCalc, 0)
		scfPeSink = dist + FixpDBL(best) + FixpDBL(minCalc)
		scfPeHashSink = hashInt16AsInt(quant[:])
	})
	if allocs != 0 {
		t.Fatalf("improve-scf allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncAssimilateSingleScfAllocs(t *testing.T) {
	var psy PsyOutChannel
	var qc QCOutChannel
	var quant [11]int16
	var quantTmp [11]int16
	var minScf [5]int
	var constPart [5]FixpDBL
	var form [5]FixpDBL
	var lines [5]FixpDBL
	scfBase := [5]int{4, 7, 6, 4, 5}
	minScfCalculatedBase := [5]int{4, 7, 6, 4, 5}
	distBase := [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, 0}
	var scf [5]int
	var minScfCalculated [5]int
	var dist [5]FixpDBL

	allocs := testing.AllocsPerRun(1000, func() {
		setupAssimilateSingleVector(&psy, &qc, quant[:], quantTmp[:], minScf[:], constPart[:], form[:])
		copy(scf[:], scfBase[:])
		copy(minScfCalculated[:], minScfCalculatedBase[:])
		copy(dist[:], distBase[:])
		FDKaacEncAssimilateSingleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:], minScfCalculated[:], 1)
		scfPeSink = dist[1] + dist[2] + dist[4]
		scfPeHashSink = hashBandEnergyInts(scf[:])
	})
	if allocs != 0 {
		t.Fatalf("assimilate-single allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncAssimilateMultipleScfAllocs(t *testing.T) {
	var psy PsyOutChannel
	var qc QCOutChannel
	var quant [11]int16
	var quantTmp [11]int16
	var minScf [5]int
	var constPart [5]FixpDBL
	var form [5]FixpDBL
	var lines [5]FixpDBL
	scfBase := [5]int{4, 7, 6, fdkIntMin, 5}
	distBase := [5]FixpDBL{0, MaxValDBL, MaxValDBL, 0, MaxValDBL}
	threshold := [5]FixpDBL{0, 0, 0, 0, 0}
	var scf [5]int
	var dist [5]FixpDBL

	allocs := testing.AllocsPerRun(1000, func() {
		setupAssimilateSingleVector(&psy, &qc, quant[:], quantTmp[:], minScf[:], constPart[:], form[:])
		copy(qc.SfbThresholdLdData[:], threshold[:])
		copy(scf[:], scfBase[:])
		copy(dist[:], distBase[:])
		FDKaacEncAssimilateMultipleScf(&psy, &qc, quant[:], quantTmp[:], 0, scf[:], minScf[:], dist[:], constPart[:], form[:], lines[:])
		scfPeSink = dist[1] + dist[2] + dist[4]
		scfPeHashSink = hashBandEnergyInts(scf[:])
	})
	if allocs != 0 {
		t.Fatalf("assimilate-multiple allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncScfPeDiffAllocs(t *testing.T) {
	const min = fdkIntMin
	scfOld := [...]int{12, min, 8, 11, min, 14, 18, min, 16, 20}
	scfNew := [...]int{13, min, 10, 9, min, 15, 17, min, 14, 19}
	energy := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000, -260000000, -240000000, -200000000, -330000000}
	form := [...]FixpDBL{-330220071, -277590868, -332606544, -282650384, -338598629, -274335045, -293504191, -323306888}
	lines := [...]FixpDBL{7252715, 0, 5626260, 23182377, 0, 19176032, 15006516, 8382327}
	constPartBase := [...]FixpDBL{FixpDBL(min), 0x04000000, 0x03000000, FixpDBL(min), 0x02000000, FixpDBL(min), 0x20000000, FixpDBL(min)}
	var constPart [len(constPartBase)]FixpDBL
	var got [2]FixpDBL

	allocs := testing.AllocsPerRun(1000, func() {
		copy(constPart[:], constPartBase[:])
		got[0] = FDKaacEncCountScfBitsDiff(scfOld[:], scfNew[:], len(scfOld), 2, 8)
		got[1] = FDKaacEncCalcSpecPeDiff(
			energy[:], scfOld[:len(energy)], scfNew[:len(energy)], constPart[:], form[:], lines[:], 0, len(energy),
		)
		scfPeSink = got[0] + got[1] + constPart[0] + constPart[7]
		scfPeHashSink = hashFixpDBL(got[:])
	})
	if allocs != 0 {
		t.Fatalf("scalefactor/PE diff allocations = %v, want 0", allocs)
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

func improveScfVectorSpec() [7]FixpDBL {
	return [...]FixpDBL{
		0x00100000,
		-0x00200000,
		0x00800000,
		-0x01000000,
		0x04000000,
		-0x08000000,
		0x10000000,
	}
}

func fillImproveScfQuant(x []int16, v int16) {
	for i := range x {
		x[i] = v
	}
}

func setupAssimilateSingleVector(
	psy *PsyOutChannel,
	qc *QCOutChannel,
	quant []int16,
	quantTmp []int16,
	minScf []int,
	constPart []FixpDBL,
	form []FixpDBL,
) {
	offsets := [...]int{0, 2, 4, 7, 9, 11}
	spec := [...]FixpDBL{
		0x00100000, -0x00200000,
		0x00800000, -0x01000000,
		0x04000000, -0x08000000, 0x10000000,
		-0x18000000, 0x20000000,
		0x01800000, -0x03000000,
	}
	energy := [...]FixpDBL{-300000000, -250000000, -280000000, -310000000, -260000000}
	*psy = PsyOutChannel{
		SfbCnt:         5,
		SfbPerGroup:    5,
		MaxSfbPerGroup: 5,
	}
	copy(psy.SfbOffsets[:], offsets[:])
	copy(qc.MdctSpectrum[:], spec[:])
	copy(qc.SfbEnergyLdData[:], energy[:])
	fillImproveScfQuant(quant, 77)
	fillImproveScfQuant(quantTmp, -77)
	for i := range minScf {
		minScf[i] = 0
	}
	for i := range constPart {
		constPart[i] = FixpDBL(fdkIntMin)
	}
	for i := range form {
		form[i] = -330220071 + FixpDBL(i*10000000)
	}
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
