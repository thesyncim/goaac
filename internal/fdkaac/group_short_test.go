package fdkaac

import "testing"

var groupShortSink FixpDBL
var groupShortIntSink int
var groupShortHashSink uint64

func TestFDKaacEncGroupShortDataVectors(t *testing.T) {
	var mdctSpectrum [160]FixpDBL
	var sfbThreshold SFBThreshold
	var sfbEnergy SFBEnergy
	var sfbEnergyMS SFBEnergy
	var sfbSpreadEnergy SFBEnergy
	var groupedSfbOffset [21]int
	var groupedSfbMinSnrLdData [20]FixpDBL
	var scratch GroupShortScratch

	fillGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy)
	maxSfbPerGroup := -1
	callGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], &scratch)

	if maxSfbPerGroup != 5 {
		t.Fatalf("maxSfbPerGroup = %d, want 5", maxSfbPerGroup)
	}

	wantGroupedSfbOffset := [...]int{0, 6, 15, 27, 42, 60, 66, 75, 87, 102, 120, 122, 125, 129, 134, 140, 142, 145, 149, 154, 160}
	if groupedSfbOffset != wantGroupedSfbOffset {
		t.Fatalf("grouped sfb offsets = %v, want %v", groupedSfbOffset, wantGroupedSfbOffset)
	}
	if got, want := hashBandEnergyInts(groupedSfbOffset[:]), uint64(0x9e9eeae8b6e593ed); got != want {
		t.Fatalf("grouped sfb offset hash = %#016x, want %#016x", got, want)
	}

	wantGroupedMinSnr := [...]FixpDBL{-1000, -2000, -3000, -4000, -5000, -1000, -2000, -3000, -4000, -5000, -1000, -2000, -3000, -4000, -5000, -1000, -2000, -3000, -4000, -5000}
	if groupedSfbMinSnrLdData != wantGroupedMinSnr {
		t.Fatalf("grouped min-snr = %v, want %v", groupedSfbMinSnrLdData, wantGroupedMinSnr)
	}
	if got, want := hashFixpDBL(groupedSfbMinSnrLdData[:]), uint64(0xcb48ead74aeaabed); got != want {
		t.Fatalf("grouped min-snr hash = %#016x, want %#016x", got, want)
	}

	wantThreshold := [...]FixpDBL{12582912, 18874368, 25165824, 31457280, MaxValDBL, 31457280, 47185920, 62914560, 78643200, 94371840, 14680064, 22020096, 29360128, 36700160, 44040192, 16777216, 25165824, 33554432, 41943040, 50331648}
	assertFixpDBLSlice(t, "threshold", sfbThreshold.Long[:20], wantThreshold[:], 0x1d438b142fde4978)

	wantEnergy := [...]FixpDBL{4718592, 9437184, 14155776, 18874368, 23592960, 9437184, 18874368, MaxValDBL, 37748736, 47185920, 4194304, 8388608, 12582912, 16777216, 20971520, 4718592, 9437184, 14155776, 18874368, 23592960}
	assertFixpDBLSlice(t, "energy", sfbEnergy.Long[:20], wantEnergy[:], 0xf3f2588ba5ea7e3e)

	wantEnergyMS := [...]FixpDBL{6291456, 9437184, 12582912, 15728640, 18874368, 11010048, 16515072, 22020096, 27525120, 33030144, 4718592, 7077888, 9437184, 11796480, 14155776, 5242880, 7864320, 10485760, 13107200, 15728640}
	assertFixpDBLSlice(t, "energyMS", sfbEnergyMS.Long[:20], wantEnergyMS[:], 0x9ceb7fd83d664a4d)

	wantSpreadEnergy := [...]FixpDBL{5898240, 7864320, 9830400, 11796480, 13762560, 9437184, 12582912, 15728640, 18874368, 22020096, 3932160, 5242880, 6553600, 7864320, 9175040, 4325376, 5767168, 7208960, 8650752, 10092544}
	assertFixpDBLSlice(t, "spread energy", sfbSpreadEnergy.Long[:20], wantSpreadEnergy[:], 0x9b33596aaee5cee1)

	wantSpectrumFirst40 := [...]FixpDBL{-2000, -1993, -1000, -993, 0, 7, -1986, -1979, -1972, -986, -979, -972, 14, 21, 28, -1965, -1958, -1951, -1944, -965, -958, -951, -944, 35, 42, 49, 56, -1937, -1930, -1923, -1916, -1909, -937, -930, -923, -916, -909, 63, 70, 77}
	assertFixpDBLSlice(t, "spectrum first40", mdctSpectrum[:40], wantSpectrumFirst40[:], 0x491ee77ba0e45579)
	if got, want := hashFixpDBL(mdctSpectrum[:]), uint64(0x3a7408c73b720f32); got != want {
		t.Fatalf("spectrum hash = %#016x, want %#016x", got, want)
	}

	if sfbThreshold.Long[20] != -10001 || sfbEnergy.Long[20] != -10002 || sfbEnergyMS.Long[20] != -10003 || sfbSpreadEnergy.Long[20] != -10004 {
		t.Fatalf("long energy sentinels changed: threshold=%d energy=%d ms=%d spread=%d", sfbThreshold.Long[20], sfbEnergy.Long[20], sfbEnergyMS.Long[20], sfbSpreadEnergy.Long[20])
	}
}

func TestFDKaacEncGroupShortDataMaxSFBFloor(t *testing.T) {
	var mdctSpectrum [160]FixpDBL
	var sfbThreshold SFBThreshold
	var sfbEnergy SFBEnergy
	var sfbEnergyMS SFBEnergy
	var sfbSpreadEnergy SFBEnergy
	var groupedSfbOffset [21]int
	var groupedSfbMinSnrLdData [20]FixpDBL
	var scratch GroupShortScratch

	fillGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy)
	for i := range mdctSpectrum {
		mdctSpectrum[i] = 0
	}
	maxSfbPerGroup := -1
	callGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], &scratch)

	if maxSfbPerGroup != 1 {
		t.Fatalf("zero spectrum maxSfbPerGroup = %d, want 1", maxSfbPerGroup)
	}
}

func TestFDKaacEncGroupShortDataInactiveBandVector(t *testing.T) {
	var mdctSpectrum [160]FixpDBL
	var sfbThreshold SFBThreshold
	var sfbEnergy SFBEnergy
	var sfbEnergyMS SFBEnergy
	var sfbSpreadEnergy SFBEnergy
	var groupedSfbOffset [21]int
	var groupedSfbMinSnrLdData [20]FixpDBL
	var scratch GroupShortScratch

	fillGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy)
	for i := range groupedSfbMinSnrLdData {
		groupedSfbMinSnrLdData[i] = -7777
	}
	for i := range scratch.Spectrum {
		scratch.Spectrum[i] = 1234567
	}

	maxSfbPerGroup := -1
	callGroupShortInactiveVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], &scratch)

	if maxSfbPerGroup != 4 {
		t.Fatalf("inactive maxSfbPerGroup = %d, want 4", maxSfbPerGroup)
	}

	wantGroupedSfbOffset := [...]int{0, 6, 15, 27, 42, 60, 66, 75, 87, 102, 120, 122, 125, 129, 134, 140, 142, 145, 149, 154, 160}
	if groupedSfbOffset != wantGroupedSfbOffset {
		t.Fatalf("inactive grouped sfb offsets = %v, want %v", groupedSfbOffset, wantGroupedSfbOffset)
	}
	wantGroupedMinSnr := [...]FixpDBL{-1000, -2000, -3000, -4000, -7777, -1000, -2000, -3000, -4000, -7777, -1000, -2000, -3000, -4000, -7777, -1000, -2000, -3000, -4000, -7777}
	assertFixpDBLSlice(t, "inactive grouped min-snr", groupedSfbMinSnrLdData[:], wantGroupedMinSnr[:], 0x86cd6b7d29a64d25)

	wantThreshold := [...]FixpDBL{12582912, 18874368, 25165824, 31457280, -10001, 31457280, 47185920, 62914560, 78643200, -10001, 14680064, 22020096, 29360128, 36700160, -10001, 16777216, 25165824, 33554432, 41943040, -10001}
	assertFixpDBLSlice(t, "inactive threshold", sfbThreshold.Long[:20], wantThreshold[:], 0x0945a881b42e8b10)

	wantEnergy := [...]FixpDBL{4718592, 9437184, 14155776, 18874368, -10002, 9437184, 18874368, MaxValDBL, 37748736, -10002, 4194304, 8388608, 12582912, 16777216, -10002, 4718592, 9437184, 14155776, 18874368, -10002}
	assertFixpDBLSlice(t, "inactive energy", sfbEnergy.Long[:20], wantEnergy[:], 0xee1fcb21c7c68e87)

	wantEnergyMS := [...]FixpDBL{6291456, 9437184, 12582912, 15728640, -10003, 11010048, 16515072, 22020096, 27525120, -10003, 4718592, 7077888, 9437184, 11796480, -10003, 5242880, 7864320, 10485760, 13107200, -10003}
	assertFixpDBLSlice(t, "inactive energyMS", sfbEnergyMS.Long[:20], wantEnergyMS[:], 0x8a8bf3c6a1e23a65)

	wantSpreadEnergy := [...]FixpDBL{5898240, 7864320, 9830400, 11796480, -10004, 9437184, 12582912, 15728640, 18874368, -10004, 3932160, 5242880, 6553600, 7864320, -10004, 4325376, 5767168, 7208960, 8650752, -10004}
	assertFixpDBLSlice(t, "inactive spread energy", sfbSpreadEnergy.Long[:20], wantSpreadEnergy[:], 0x85db8549db19e942)

	wantSpectrumFirst60 := [...]FixpDBL{
		-2000, -1993, -1000, -993, 0, 7,
		-1986, -1979, -1972, -986, -979, -972, 14, 21, 28,
		-1965, -1958, -1951, -1944, -965, -958, -951, -944, 35, 42, 49, 56,
		-1937, -1930, -1923, -1916, -1909, -937, -930, -923, -916, -909, 63, 70, 77, 84, 91,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	assertFixpDBLSlice(t, "inactive spectrum first60", mdctSpectrum[:60], wantSpectrumFirst60[:], 0x7e9ba8babe625a46)
	for _, span := range [...][2]int{{42, 60}, {102, 120}, {134, 140}, {154, 160}} {
		for i := span[0]; i < span[1]; i++ {
			if mdctSpectrum[i] != 0 {
				t.Fatalf("inactive spectrum[%d] = %d, want zero", i, mdctSpectrum[i])
			}
		}
	}
	if got, want := hashFixpDBL(mdctSpectrum[:]), uint64(0x0e1fee5eef478ea1); got != want {
		t.Fatalf("inactive spectrum hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncGroupShortDataRejectsInvalid(t *testing.T) {
	var mdctSpectrum [160]FixpDBL
	var sfbThreshold SFBThreshold
	var sfbEnergy SFBEnergy
	var sfbEnergyMS SFBEnergy
	var sfbSpreadEnergy SFBEnergy
	var groupedSfbOffset [21]int
	var groupedSfbMinSnrLdData [20]FixpDBL
	var scratch GroupShortScratch
	sfbOffset := [...]int{0, 2, 5, 9, 14, 20}
	sfbMinSnr := [...]FixpDBL{-1000, -2000, -3000, -4000, -5000}
	groupLen := [...]int{3, 3, 1, 1}
	maxSfbPerGroup := 0

	fillGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy)

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "nil threshold", fn: func() {
			FDKaacEncGroupShortData(mdctSpectrum[:], nil, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 5, sfbOffset[:], sfbMinSnr[:], groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], 4, groupLen[:], len(mdctSpectrum), &scratch)
		}},
		{name: "nil output", fn: func() {
			FDKaacEncGroupShortData(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 5, sfbOffset[:], sfbMinSnr[:], groupedSfbOffset[:], nil, groupedSfbMinSnrLdData[:], 4, groupLen[:], len(mdctSpectrum), &scratch)
		}},
		{name: "nil scratch", fn: func() {
			FDKaacEncGroupShortData(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 5, sfbOffset[:], sfbMinSnr[:], groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], 4, groupLen[:], len(mdctSpectrum), nil)
		}},
		{name: "bad active count", fn: func() {
			FDKaacEncGroupShortData(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 6, sfbOffset[:], sfbMinSnr[:], groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], 4, groupLen[:], len(mdctSpectrum), &scratch)
		}},
		{name: "bad offsets", fn: func() {
			badOffset := [...]int{0, 2, 5, 4, 14, 20}
			FDKaacEncGroupShortData(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 5, badOffset[:], sfbMinSnr[:], groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], 4, groupLen[:], len(mdctSpectrum), &scratch)
		}},
		{name: "nonzero first offset", fn: func() {
			badOffset := [...]int{1, 2, 5, 9, 14, 20}
			FDKaacEncGroupShortData(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 5, badOffset[:], sfbMinSnr[:], groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], 4, groupLen[:], len(mdctSpectrum), &scratch)
		}},
		{name: "bad group sum", fn: func() {
			badGroupLen := [...]int{3, 3, 1, 2}
			FDKaacEncGroupShortData(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 5, sfbOffset[:], sfbMinSnr[:], groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], 4, badGroupLen[:], len(mdctSpectrum), &scratch)
		}},
		{name: "short grouped offset", fn: func() {
			FDKaacEncGroupShortData(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 5, sfbOffset[:], sfbMinSnr[:], groupedSfbOffset[:20], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], 4, groupLen[:], len(mdctSpectrum), &scratch)
		}},
		{name: "short grouped min-snr", fn: func() {
			FDKaacEncGroupShortData(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, 5, 5, sfbOffset[:], sfbMinSnr[:], groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:19], 4, groupLen[:], len(mdctSpectrum), &scratch)
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

func TestFDKaacEncGroupShortDataAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var mdctSpectrum [160]FixpDBL
		var sfbThreshold SFBThreshold
		var sfbEnergy SFBEnergy
		var sfbEnergyMS SFBEnergy
		var sfbSpreadEnergy SFBEnergy
		var groupedSfbOffset [21]int
		var groupedSfbMinSnrLdData [20]FixpDBL
		var scratch GroupShortScratch

		fillGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy)
		maxSfbPerGroup := 0
		callGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], &scratch)
		fillGroupShortVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy)
		callGroupShortInactiveVector(mdctSpectrum[:], &sfbThreshold, &sfbEnergy, &sfbEnergyMS, &sfbSpreadEnergy, groupedSfbOffset[:], &maxSfbPerGroup, groupedSfbMinSnrLdData[:], &scratch)

		groupShortSink = mdctSpectrum[0] + sfbThreshold.Long[0] + sfbEnergy.Long[7] + sfbEnergyMS.Long[19] + sfbSpreadEnergy.Long[14] + groupedSfbMinSnrLdData[19]
		groupShortIntSink = maxSfbPerGroup + groupedSfbOffset[20]
		groupShortHashSink = hashFixpDBL(mdctSpectrum[:])
	})
	if allocs != 0 {
		t.Fatalf("group short allocations = %v, want 0", allocs)
	}
}

func fillGroupShortVector(mdctSpectrum []FixpDBL, sfbThreshold *SFBThreshold, sfbEnergy *SFBEnergy, sfbEnergyMS *SFBEnergy, sfbSpreadEnergy *SFBEnergy) {
	for i := range sfbThreshold.Long {
		sfbThreshold.Long[i] = -10001
		sfbEnergy.Long[i] = -10002
		sfbEnergyMS.Long[i] = -10003
		sfbSpreadEnergy.Long[i] = -10004
	}
	for wnd := 0; wnd < transFac; wnd++ {
		for line := 0; line < 20; line++ {
			mdctSpectrum[wnd*20+line] = FixpDBL((wnd+1)*1000 + line*7 - 3000)
		}
	}
	for wnd := 0; wnd < 7; wnd++ {
		for line := 14; line < 20; line++ {
			mdctSpectrum[wnd*20+line] = 0
		}
	}
	for wnd := 0; wnd < transFac; wnd++ {
		for sfb := 0; sfb < 5; sfb++ {
			sfbThreshold.Short[wnd][sfb] = FixpDBL((wnd + 1) * (sfb + 2) * 0x00100000)
			sfbEnergy.Short[wnd][sfb] = FixpDBL((wnd + 2) * (sfb + 1) * 0x00080000)
			sfbEnergyMS.Short[wnd][sfb] = FixpDBL((wnd + 3) * (sfb + 2) * 0x00040000)
			sfbSpreadEnergy.Short[wnd][sfb] = FixpDBL((wnd + 4) * (sfb + 3) * 0x00020000)
		}
	}
	sfbThreshold.Short[1][4] = MaxValDBL - 5
	sfbThreshold.Short[2][4] = 100
	sfbEnergy.Short[4][2] = MaxValDBL - 20
	sfbEnergy.Short[5][2] = 25
}

func callGroupShortVector(
	mdctSpectrum []FixpDBL,
	sfbThreshold *SFBThreshold,
	sfbEnergy *SFBEnergy,
	sfbEnergyMS *SFBEnergy,
	sfbSpreadEnergy *SFBEnergy,
	groupedSfbOffset []int,
	maxSfbPerGroup *int,
	groupedSfbMinSnrLdData []FixpDBL,
	scratch *GroupShortScratch,
) {
	sfbOffset := [...]int{0, 2, 5, 9, 14, 20}
	sfbMinSnrLdData := [...]FixpDBL{-1000, -2000, -3000, -4000, -5000}
	groupLen := [...]int{3, 3, 1, 1}
	FDKaacEncGroupShortData(
		mdctSpectrum,
		sfbThreshold,
		sfbEnergy,
		sfbEnergyMS,
		sfbSpreadEnergy,
		5,
		5,
		sfbOffset[:],
		sfbMinSnrLdData[:],
		groupedSfbOffset,
		maxSfbPerGroup,
		groupedSfbMinSnrLdData,
		4,
		groupLen[:],
		len(mdctSpectrum),
		scratch,
	)
}

func callGroupShortInactiveVector(
	mdctSpectrum []FixpDBL,
	sfbThreshold *SFBThreshold,
	sfbEnergy *SFBEnergy,
	sfbEnergyMS *SFBEnergy,
	sfbSpreadEnergy *SFBEnergy,
	groupedSfbOffset []int,
	maxSfbPerGroup *int,
	groupedSfbMinSnrLdData []FixpDBL,
	scratch *GroupShortScratch,
) {
	sfbOffset := [...]int{0, 2, 5, 9, 14, 20}
	sfbMinSnrLdData := [...]FixpDBL{-1000, -2000, -3000, -4000, -5000}
	groupLen := [...]int{3, 3, 1, 1}
	FDKaacEncGroupShortData(
		mdctSpectrum,
		sfbThreshold,
		sfbEnergy,
		sfbEnergyMS,
		sfbSpreadEnergy,
		5,
		4,
		sfbOffset[:],
		sfbMinSnrLdData[:],
		groupedSfbOffset,
		maxSfbPerGroup,
		groupedSfbMinSnrLdData,
		4,
		groupLen[:],
		len(mdctSpectrum),
		scratch,
	)
}

func assertFixpDBLSlice(t *testing.T, name string, got []FixpDBL, want []FixpDBL, wantHash uint64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %d, want %d; got %v want %v", name, i, got[i], want[i], got, want)
		}
	}
	if h := hashFixpDBL(got); h != wantHash {
		t.Fatalf("%s hash = %#016x, want %#016x", name, h, wantHash)
	}
}
