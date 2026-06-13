package fdkaac

import "testing"

var advanceThresholdSink FixpDBL
var advanceThresholdIntSink int
var advanceThresholdHashSink uint64

func TestFDKaacEncAdvanceThresholdsLongVectors(t *testing.T) {
	const (
		sfbActive = 6
		maxSfb    = 8
	)
	threshold := [...]FixpDBL{0x50000000, 0x01000000, 0x08000000, 0x02000000, 0x00600000, 0x04000000, -777, -888}
	energy := [...]FixpDBL{0x00800000, 0x02000000, 0x00400000, 0x03000000, 0x00200000, 0x00100000, -111, -222}
	spreadEnergy := [...]FixpDBL{-333, -333, -333, -333, -333, -333, -333, -333}
	conf := makeAdvanceThresholdConfig(sfbActive)
	state := PsyThresholdState{
		SfbThresholdNm1: [maxGroupedSFB]FixpDBL{0x20000000, 0x10000000, 0x04000000, 0x02000000, 0x01000000, 0x00800000},
		CalcPreEcho:     1,
		MdctScalenm1:    3,
	}

	FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], sfbActive, 1, maxSfb, 3, false, false, LongWindow, &conf, &state)

	wantThreshold := [...]FixpDBL{8388608, 5242880, 8388608, 4194304, 2359296, 4194304, -777, -888}
	assertFixpDBLSlice(t, "long threshold", threshold[:], wantThreshold[:], 0x8ff6efbee40b96b2)
	wantSpreadEnergy := [...]FixpDBL{8388608, 33554432, 31457280, 50331648, 25165824, 14155776, -333, -333}
	assertFixpDBLSlice(t, "long spread energy", spreadEnergy[:], wantSpreadEnergy[:], 0x1ddc4ed1e00aee02)
	wantHistory := [...]FixpDBL{8388608, 5242880, 8388608, 4194304, 2359296, 4194304}
	assertFixpDBLSlice(t, "long history", state.SfbThresholdNm1[:sfbActive], wantHistory[:], 0x74261c1a81040cf1)
	if state.CalcPreEcho != 1 || state.MdctScalenm1 != 3 {
		t.Fatalf("long state calcPreEcho=%d mdctScalenm1=%d, want 1/3", state.CalcPreEcho, state.MdctScalenm1)
	}
}

func TestFDKaacEncAdvanceThresholdsShortLFEVectors(t *testing.T) {
	const (
		sfbActive = 5
		maxSfb    = 6
		nWindows  = 3
	)
	var threshold [nWindows * maxSfb]FixpDBL
	var energy [nWindows * maxSfb]FixpDBL
	var spreadEnergy [nWindows * maxSfb]FixpDBL
	for w := 0; w < nWindows; w++ {
		for i := 0; i < maxSfb; i++ {
			base := w * maxSfb
			if i < sfbActive {
				threshold[base+i] = FixpDBL((w + 1) * (i + 2) * 0x01000000)
				energy[base+i] = FixpDBL((w + 3) * (i + 1) * 0x00300000)
				spreadEnergy[base+i] = -7000
			} else {
				threshold[base+i] = FixpDBL(-1000 - w)
				energy[base+i] = FixpDBL(-2000 - w)
				spreadEnergy[base+i] = FixpDBL(-3000 - w)
			}
		}
	}
	conf := makeAdvanceThresholdConfig(sfbActive)
	conf.ClipEnergy = 0x01000000

	FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], sfbActive, nWindows, maxSfb, -9, true, true, ShortWindow, &conf, nil)

	wantThreshold := [...]FixpDBL{2097152, 1610612736, 805306368, 402653184, 201326592, -1000, 4194304, 6291456, 8388608, 10485760, 12582912, -1001, 6291456, 9437184, 12582912, 15728640, 18874368, -1002}
	assertFixpDBLSlice(t, "short threshold", threshold[:], wantThreshold[:], 0x2061d881940a3029)
	wantSpreadEnergy := [...]FixpDBL{9437184, 19464192, 28311552, 37748736, 47185920, -3000, 12582912, 25952256, 37748736, 50331648, 62914560, -3001, 15728640, 32440320, 47185920, 62914560, 78643200, -3002}
	assertFixpDBLSlice(t, "short spread energy", spreadEnergy[:], wantSpreadEnergy[:], 0x2531395ef1a9bc39)
}

func TestFDKaacEncAdvanceThresholdsStartStopStateVectors(t *testing.T) {
	t.Run("start resets history after control", func(t *testing.T) {
		const sfbActive = 4
		threshold := [...]FixpDBL{0x01000000, 0x02000000, 0x00400000, 0x00100000}
		energy := [...]FixpDBL{0x01000000, 0x00300000, 0x02000000, 0x00100000}
		var spreadEnergy [sfbActive]FixpDBL
		conf := makeAdvanceThresholdConfig(sfbActive)
		conf.ClipEnergy = 0x04000000
		state := PsyThresholdState{
			SfbThresholdNm1: [maxGroupedSFB]FixpDBL{0x10000000, 0x04000000, 0x01000000, 0x00800000},
			CalcPreEcho:     1,
			MdctScalenm1:    2,
		}

		FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], sfbActive, 1, sfbActive, 2, false, false, StartWindow, &conf, &state)

		wantThreshold := [...]FixpDBL{1048576, 2097152, 917504, 458752}
		assertFixpDBLSlice(t, "start threshold", threshold[:], wantThreshold[:], 0x15ffc92a0a754804)
		wantSpreadEnergy := [...]FixpDBL{16777216, 23068672, 33554432, 14680064}
		assertFixpDBLSlice(t, "start spread energy", spreadEnergy[:], wantSpreadEnergy[:], 0xaee17598eb88dcef)
		wantHistory := [...]FixpDBL{MaxValDBL, MaxValDBL, MaxValDBL, MaxValDBL}
		assertFixpDBLSlice(t, "start history", state.SfbThresholdNm1[:sfbActive], wantHistory[:], 0x60c02e02c2c4a555)
		if state.CalcPreEcho != 0 || state.MdctScalenm1 != 0 {
			t.Fatalf("start state calcPreEcho=%d mdctScalenm1=%d, want 0/0", state.CalcPreEcho, state.MdctScalenm1)
		}
	})

	t.Run("stop resets history before control", func(t *testing.T) {
		const sfbActive = 4
		threshold := [...]FixpDBL{0x04000000, 0x02000000, 0x01000000, 0x00800000}
		energy := [...]FixpDBL{0x00800000, 0x00600000, 0x00300000, 0x00100000}
		var spreadEnergy [sfbActive]FixpDBL
		conf := makeAdvanceThresholdConfig(sfbActive)
		conf.ClipEnergy = 0x02000000
		state := PsyThresholdState{
			SfbThresholdNm1: [maxGroupedSFB]FixpDBL{1, 2, 3, 4},
			CalcPreEcho:     1,
			MdctScalenm1:    9,
		}

		FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], sfbActive, 1, sfbActive, 1, false, false, StopWindow, &conf, &state)

		wantThreshold := [...]FixpDBL{4194304, 2097152, 1048576, 524288}
		assertFixpDBLSlice(t, "stop threshold", threshold[:], wantThreshold[:], 0xe066de81aa7e8f4d)
		wantSpreadEnergy := [...]FixpDBL{8388608, 6291456, 3145728, 1376256}
		assertFixpDBLSlice(t, "stop spread energy", spreadEnergy[:], wantSpreadEnergy[:], 0x1ac2f70e95036038)
		assertFixpDBLSlice(t, "stop history", state.SfbThresholdNm1[:sfbActive], wantThreshold[:], 0xe066de81aa7e8f4d)
		if state.CalcPreEcho != 1 || state.MdctScalenm1 != 1 {
			t.Fatalf("stop state calcPreEcho=%d mdctScalenm1=%d, want 1/1", state.CalcPreEcho, state.MdctScalenm1)
		}
	})
}

func TestFDKaacEncAdvanceThresholdsRejectsInvalid(t *testing.T) {
	var threshold [8]FixpDBL
	var energy [8]FixpDBL
	var spreadEnergy [8]FixpDBL
	conf := makeAdvanceThresholdConfig(4)
	var state PsyThresholdState

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "nil config", fn: func() {
			FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], 4, 1, 4, 0, false, false, LongWindow, nil, &state)
		}},
		{name: "nil state", fn: func() {
			FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], 4, 1, 4, 0, false, false, LongWindow, &conf, nil)
		}},
		{name: "zero bands", fn: func() {
			FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], 0, 1, 4, 0, false, false, LongWindow, &conf, &state)
		}},
		{name: "bad stride", fn: func() {
			FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], 4, 1, 3, 0, false, false, LongWindow, &conf, &state)
		}},
		{name: "bad long windows", fn: func() {
			FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], 4, 2, 4, 0, false, false, LongWindow, &conf, &state)
		}},
		{name: "bad short windows", fn: func() {
			FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], 4, 9, 4, 0, true, true, ShortWindow, &conf, nil)
		}},
		{name: "bad sequence", fn: func() {
			FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], 4, 1, 4, 0, false, false, WrongWindow, &conf, &state)
		}},
		{name: "short threshold", fn: func() {
			FDKaacEncAdvanceThresholds(threshold[:3], energy[:], spreadEnergy[:], 4, 1, 4, 0, false, false, LongWindow, &conf, &state)
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

func TestFDKaacEncAdvanceThresholdsAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		const (
			sfbActive = 6
			maxSfb    = 8
		)
		threshold := [...]FixpDBL{0x50000000, 0x01000000, 0x08000000, 0x02000000, 0x00600000, 0x04000000, -777, -888}
		energy := [...]FixpDBL{0x00800000, 0x02000000, 0x00400000, 0x03000000, 0x00200000, 0x00100000, -111, -222}
		spreadEnergy := [...]FixpDBL{-333, -333, -333, -333, -333, -333, -333, -333}
		conf := makeAdvanceThresholdConfig(sfbActive)
		state := PsyThresholdState{
			SfbThresholdNm1: [maxGroupedSFB]FixpDBL{0x20000000, 0x10000000, 0x04000000, 0x02000000, 0x01000000, 0x00800000},
			CalcPreEcho:     1,
			MdctScalenm1:    3,
		}

		FDKaacEncAdvanceThresholds(threshold[:], energy[:], spreadEnergy[:], sfbActive, 1, maxSfb, 3, false, false, LongWindow, &conf, &state)
		advanceThresholdSink = threshold[0] + spreadEnergy[1] + state.SfbThresholdNm1[2]
		advanceThresholdIntSink = state.CalcPreEcho + state.MdctScalenm1
		advanceThresholdHashSink = hashFixpDBL(threshold[:])
	})
	if allocs != 0 {
		t.Fatalf("advance thresholds allocations = %v, want 0", allocs)
	}
}

func makeAdvanceThresholdConfig(sfbActive int) PsyThresholdConfig {
	conf := PsyThresholdConfig{
		ClipEnergy:                  0x20000000,
		MaxAllowedIncreaseFactor:    2,
		MinRemainingThresholdFactor: 0x0148,
	}
	pcmThresholds := [...]FixpDBL{0x30000000, 0x18000000, 0x0c000000, 0x06000000, 0x03000000, 0x01800000}
	maskLow := [...]FixpDBL{0, 0x50000000, 0x48000000, 0x40000000, 0x38000000, 0x30000000}
	maskHigh := [...]FixpDBL{0, 0x30000000, 0x38000000, 0x40000000, 0x48000000, 0x50000000}
	maskLowSpr := [...]FixpDBL{0, 0x58000000, 0x50000000, 0x48000000, 0x40000000, 0x38000000}
	maskHighSpr := [...]FixpDBL{0, 0x28000000, 0x30000000, 0x38000000, 0x40000000, 0x48000000}
	for i := 0; i < sfbActive; i++ {
		conf.SfbPcmQuantThreshold[i] = pcmThresholds[i]
		conf.SfbMaskLowFactor[i] = maskLow[i]
		conf.SfbMaskHighFactor[i] = maskHigh[i]
		conf.SfbMaskLowFactorSprEn[i] = maskLowSpr[i]
		conf.SfbMaskHighFactorSprEn[i] = maskHighSpr[i]
	}
	return conf
}
