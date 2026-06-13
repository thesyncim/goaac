package fdkaac

import "testing"

var psyOutputSink FixpDBL
var psyOutputIntSink int
var psyOutputHashSink uint64

func TestFDKaacEncBuildPsyOutChannelLongVectors(t *testing.T) {
	var sfbEnergy SFBEnergy
	var sfbSpreadEnergy SFBEnergy
	fillPsyOutputEnergy(&sfbEnergy, &sfbSpreadEnergy)
	blockSwitch := BlockSwitchingControl{
		LastWindowSequence: StopWindow,
		WindowShape:        WindowShapeKBD,
		NoOfGroups:         1,
		GroupLen:           [maxNoOfGroups]int{1, 0, 0, 0},
	}
	out := PsyOutChannel{GroupingMask: -1}

	FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &blockSwitch, 7, 11, false, 12, 0)

	if out.SfbCnt != 12 || out.SfbPerGroup != 12 || out.MaxSfbPerGroup != 7 {
		t.Fatalf("long sfb fields = cnt:%d per:%d max:%d", out.SfbCnt, out.SfbPerGroup, out.MaxSfbPerGroup)
	}
	if out.LastWindowSequence != StopWindow || out.WindowShape != WindowShapeKBD || out.MdctScale != 11 || out.GroupingMask != 0 {
		t.Fatalf("long state = seq:%d shape:%d scale:%d mask:%d", out.LastWindowSequence, out.WindowShape, out.MdctScale, out.GroupingMask)
	}
	assertIntSlice(t, "long group len", out.GroupLen[:], []int{1, 0, 0, 0}, 0x392209f14dea4c24)
	assertPsyOutputEnergy(t, &out)
}

func TestFDKaacEncBuildPsyOutChannelShortVectors(t *testing.T) {
	var sfbEnergy SFBEnergy
	var sfbSpreadEnergy SFBEnergy
	fillPsyOutputEnergy(&sfbEnergy, &sfbSpreadEnergy)
	blockSwitch := BlockSwitchingControl{
		LastWindowSequence: ShortWindow,
		WindowShape:        WindowShapeKBD,
		NoOfGroups:         maxNoOfGroups,
		GroupLen:           [maxNoOfGroups]int{3, 3, 1, 1},
	}
	out := PsyOutChannel{LastWindowSequence: LongWindow, WindowShape: WindowShapeKBD}

	FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &blockSwitch, 5, 8, true, 0, 5)

	if out.SfbCnt != 20 || out.SfbPerGroup != 5 || out.MaxSfbPerGroup != 5 {
		t.Fatalf("short sfb fields = cnt:%d per:%d max:%d", out.SfbCnt, out.SfbPerGroup, out.MaxSfbPerGroup)
	}
	if out.LastWindowSequence != ShortWindow || out.WindowShape != WindowShapeSine || out.MdctScale != 8 || out.GroupingMask != 108 {
		t.Fatalf("short state = seq:%d shape:%d scale:%d mask:%d", out.LastWindowSequence, out.WindowShape, out.MdctScale, out.GroupingMask)
	}
	assertIntSlice(t, "short group len", out.GroupLen[:], []int{3, 3, 1, 1}, 0x62da96e0648cb665)
	assertPsyOutputEnergy(t, &out)
}

func TestFDKaacEncBuildPsyOutGroupingMasks(t *testing.T) {
	var sfbEnergy SFBEnergy
	var sfbSpreadEnergy SFBEnergy
	fillPsyOutputEnergy(&sfbEnergy, &sfbSpreadEnergy)

	tests := []struct {
		name     string
		groupLen [maxNoOfGroups]int
		groups   int
		wantMask int
	}{
		{name: "attack0", groupLen: [maxNoOfGroups]int{1, 3, 3, 1}, groups: 4, wantMask: 54},
		{name: "attack2", groupLen: [maxNoOfGroups]int{2, 1, 3, 2}, groups: 4, wantMask: 77},
		{name: "two-groups", groupLen: [maxNoOfGroups]int{4, 4, 0, 0}, groups: 2, wantMask: 119},
	}

	for _, tt := range tests {
		blockSwitch := BlockSwitchingControl{
			LastWindowSequence: ShortWindow,
			WindowShape:        WindowShapeSine,
			NoOfGroups:         tt.groups,
			GroupLen:           tt.groupLen,
		}
		var out PsyOutChannel
		FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &blockSwitch, 5, 0, true, 0, 5)
		if out.GroupingMask != tt.wantMask {
			t.Fatalf("%s grouping mask = %d, want %d", tt.name, out.GroupingMask, tt.wantMask)
		}
	}
}

func TestFDKaacEncBuildPsyOutChannelRejectsInvalid(t *testing.T) {
	var out PsyOutChannel
	var sfbEnergy SFBEnergy
	var sfbSpreadEnergy SFBEnergy
	validLong := BlockSwitchingControl{LastWindowSequence: LongWindow, WindowShape: WindowShapeSine, NoOfGroups: 1, GroupLen: [maxNoOfGroups]int{1, 0, 0, 0}}
	validShort := BlockSwitchingControl{LastWindowSequence: ShortWindow, WindowShape: WindowShapeSine, NoOfGroups: maxNoOfGroups, GroupLen: [maxNoOfGroups]int{3, 3, 1, 1}}

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "nil output", fn: func() {
			FDKaacEncBuildPsyOutChannel(nil, &sfbEnergy, &sfbSpreadEnergy, &validLong, 5, 0, false, 10, 0)
		}},
		{name: "nil energy", fn: func() {
			FDKaacEncBuildPsyOutChannel(&out, nil, &sfbSpreadEnergy, &validLong, 5, 0, false, 10, 0)
		}},
		{name: "nil block switching", fn: func() {
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, nil, 5, 0, false, 10, 0)
		}},
		{name: "bad groups", fn: func() {
			bad := validLong
			bad.NoOfGroups = 0
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &bad, 5, 0, false, 10, 0)
		}},
		{name: "bad group length", fn: func() {
			bad := validShort
			bad.GroupLen[1] = 0
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &bad, 5, 0, true, 0, 5)
		}},
		{name: "short sequence mismatch", fn: func() {
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &validLong, 5, 0, true, 0, 5)
		}},
		{name: "short bad sum", fn: func() {
			bad := validShort
			bad.GroupLen = [maxNoOfGroups]int{3, 3, 1, 2}
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &bad, 5, 0, true, 0, 5)
		}},
		{name: "short bad count", fn: func() {
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &validShort, 5, 0, true, 0, 16)
		}},
		{name: "short bad max", fn: func() {
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &validShort, 6, 0, true, 0, 5)
		}},
		{name: "long short sequence", fn: func() {
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &validShort, 5, 0, false, 10, 0)
		}},
		{name: "long bad grouping", fn: func() {
			bad := validLong
			bad.NoOfGroups = 2
			bad.GroupLen = [maxNoOfGroups]int{1, 1, 0, 0}
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &bad, 5, 0, false, 10, 0)
		}},
		{name: "long bad active", fn: func() {
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &validLong, 5, 0, false, 0, 0)
		}},
		{name: "long bad max", fn: func() {
			FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &validLong, 11, 0, false, 10, 0)
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

func TestFDKaacEncBuildPsyOutChannelAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var out PsyOutChannel
		var sfbEnergy SFBEnergy
		var sfbSpreadEnergy SFBEnergy
		fillPsyOutputEnergy(&sfbEnergy, &sfbSpreadEnergy)
		blockSwitch := BlockSwitchingControl{
			LastWindowSequence: ShortWindow,
			WindowShape:        WindowShapeKBD,
			NoOfGroups:         maxNoOfGroups,
			GroupLen:           [maxNoOfGroups]int{3, 3, 1, 1},
		}
		FDKaacEncBuildPsyOutChannel(&out, &sfbEnergy, &sfbSpreadEnergy, &blockSwitch, 5, 8, true, 0, 5)
		psyOutputSink = out.SfbEnergy[0] + out.SfbSpreadEnergy[23] + FixpDBL(out.GroupingMask)
		psyOutputIntSink = out.SfbCnt + out.SfbPerGroup + out.MaxSfbPerGroup + out.LastWindowSequence + out.WindowShape
		psyOutputHashSink = hashFixpDBL(out.SfbEnergy[:])
	})
	if allocs != 0 {
		t.Fatalf("psy output allocations = %v, want 0", allocs)
	}
}

func fillPsyOutputEnergy(sfbEnergy *SFBEnergy, sfbSpreadEnergy *SFBEnergy) {
	for i := 0; i < maxGroupedSFB; i++ {
		sfbEnergy.Long[i] = FixpDBL((i+3)*0x00100000 + (i%5)*1234 - 5000)
		sfbSpreadEnergy.Long[i] = FixpDBL((i+5)*0x000c0000 - (i%7)*321 + 7000)
	}
}

func assertPsyOutputEnergy(t *testing.T, out *PsyOutChannel) {
	t.Helper()
	if got, want := hashFixpDBL(out.SfbEnergy[:]), uint64(0x6c2dbdc3a4c2d851); got != want {
		t.Fatalf("psy output energy hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(out.SfbSpreadEnergy[:]), uint64(0x8b6915e66204ffcf); got != want {
		t.Fatalf("psy output spread energy hash = %#016x, want %#016x", got, want)
	}
	if out.SfbEnergy[0] != 3140728 || out.SfbEnergy[17] != 20968988 || out.SfbEnergy[59] != 65011648 {
		t.Fatalf("psy output energy samples = %d/%d/%d", out.SfbEnergy[0], out.SfbEnergy[17], out.SfbEnergy[59])
	}
	if out.SfbSpreadEnergy[0] != 3939160 || out.SfbSpreadEnergy[23] != 22026454 || out.SfbSpreadEnergy[59] != 50337685 {
		t.Fatalf("psy output spread samples = %d/%d/%d", out.SfbSpreadEnergy[0], out.SfbSpreadEnergy[23], out.SfbSpreadEnergy[59])
	}
}
