package fdkaac

import "testing"

var tonalityHashSink uint64
var tonalitySGLHashSink uint64

func TestFDKaacEncCalculateChaosMeasureVectors(t *testing.T) {
	conf := mustPsyConfigurationForPNS(t, LongWindow)
	var spectrum [maxSpectralLines]FixpDBL
	var chaos [maxSpectralLines]FixpDBL
	fillTonalitySpectrum(spectrum[:], 73)

	lines := conf.SfbOffset[conf.SfbCnt]
	FDKaacEncCalculateChaosMeasure(spectrum[:], lines, chaos[:])

	if got, want := chaos[0], chaos[2]; got != want {
		t.Fatalf("head chaos copy = %d, want %d", got, want)
	}
	for i := lines - 3; i < lines; i++ {
		if chaos[i] != tonalityHalf {
			t.Fatalf("tail chaos[%d] = %d, want %d", i, chaos[i], tonalityHalf)
		}
	}
	if got, want := hashFixpDBL(chaos[:lines]), uint64(0x2003934ff7180aa5); got != want {
		t.Fatalf("chaos hash = %#016x, want %#016x", got, want)
	}
	if got, want := chaos[96], FixpDBL(18874368); got != want {
		t.Fatalf("chaos[96] = %d, want %d", got, want)
	}
}

func TestFDKaacEncCalculateFullTonalityVectors(t *testing.T) {
	conf := mustPsyConfigurationForPNS(t, LongWindow)
	var spectrum [maxSpectralLines]FixpDBL
	var maxScale [maxSFB]int
	var bandEnergy [maxSFB]FixpDBL
	var bandEnergyLD64 [maxSFB]FixpDBL
	var tonality [maxSFB]FixpSGL
	var scratch TonalityScratch
	fillTonalitySpectrum(spectrum[:], 97)

	FDKaacEncCalcSfbMaxScaleSpec(spectrum[:], conf.SfbOffset[:], maxScale[:], conf.SfbCnt)
	if shift := FDKaacEncCalcBandEnergyOptimLong(spectrum[:], maxScale[:], conf.SfbOffset[:], conf.SfbCnt, bandEnergy[:], bandEnergyLD64[:]); shift != 0 {
		t.Fatalf("band energy shift = %d, want 0", shift)
	}

	FDKaacEncCalculateFullTonality(spectrum[:], maxScale[:], bandEnergyLD64[:], tonality[:], conf.SfbCnt, conf.SfbOffset[:], 1, &scratch)

	if got, want := hashFixpSGL(tonality[:conf.SfbCnt]), uint64(0x80064dbe68e0b94d); got != want {
		t.Fatalf("tonality hash = %#016x, want %#016x", got, want)
	}
	wantHead := [...]FixpSGL{32767, 0, 0, 1905, 0, 0, 9375, 4054, 0, 6621, 1528, 24}
	for i, want := range wantHead {
		if tonality[i] != want {
			t.Fatalf("tonality[%d] = %d, want %d", i, tonality[i], want)
		}
	}
	if got, want := hashFixpDBL(scratch.ChaosMeasurePerLine[:conf.SfbOffset[conf.SfbCnt]]), uint64(0xd0c793a53bfbd4ff); got != want {
		t.Fatalf("smoothed chaos hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncCalculateFullTonalityUsePnsZeroNoops(t *testing.T) {
	conf := mustPsyConfigurationForPNS(t, LongWindow)
	var spectrum [maxSpectralLines]FixpDBL
	var maxScale [maxSFB]int
	var energyLD64 [maxSFB]FixpDBL
	var tonality [maxSFB]FixpSGL
	for i := range tonality {
		tonality[i] = FixpSGL(i*17 - 301)
	}
	want := tonality

	FDKaacEncCalculateFullTonality(spectrum[:], maxScale[:], energyLD64[:], tonality[:], conf.SfbCnt, conf.SfbOffset[:], 0, nil)

	if tonality != want {
		t.Fatalf("usePns=0 tonality mutated")
	}
}

func TestFDKaacEncCalculateTonalityRejectsInvalid(t *testing.T) {
	conf := mustPsyConfigurationForPNS(t, LongWindow)
	var spectrum [maxSpectralLines]FixpDBL
	var chaos [maxSpectralLines]FixpDBL
	var maxScale [maxSFB]int
	var energyLD64 [maxSFB]FixpDBL
	var tonality [maxSFB]FixpSGL
	var scratch TonalityScratch

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "short chaos input",
			fn:   func() { FDKaacEncCalculateChaosMeasure(spectrum[:3], 4, chaos[:]) },
		},
		{
			name: "short chaos output",
			fn:   func() { FDKaacEncCalculateChaosMeasure(spectrum[:], 4, chaos[:3]) },
		},
		{
			name: "too few chaos lines",
			fn:   func() { FDKaacEncCalculateChaosMeasure(spectrum[:], 3, chaos[:]) },
		},
		{
			name: "odd chaos lines",
			fn:   func() { FDKaacEncCalculateChaosMeasure(spectrum[:], 5, chaos[:]) },
		},
		{
			name: "missing scratch",
			fn: func() {
				FDKaacEncCalculateFullTonality(spectrum[:], maxScale[:], energyLD64[:], tonality[:], conf.SfbCnt, conf.SfbOffset[:], 1, nil)
			},
		},
		{
			name: "short max scale",
			fn: func() {
				FDKaacEncCalculateFullTonality(spectrum[:], maxScale[:conf.SfbCnt-1], energyLD64[:], tonality[:], conf.SfbCnt, conf.SfbOffset[:], 1, &scratch)
			},
		},
		{
			name: "short tonality output",
			fn: func() {
				FDKaacEncCalculateFullTonality(spectrum[:], maxScale[:], energyLD64[:], tonality[:conf.SfbCnt-1], conf.SfbCnt, conf.SfbOffset[:], 1, &scratch)
			},
		},
		{
			name: "invalid offset",
			fn: func() {
				bad := conf.SfbOffset
				bad[conf.SfbCnt] = maxSpectralLines + 1
				FDKaacEncCalculateFullTonality(spectrum[:], maxScale[:], energyLD64[:], tonality[:], conf.SfbCnt, bad[:], 1, &scratch)
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

func TestFDKaacEncCalculateFullTonalityAllocs(t *testing.T) {
	conf := mustPsyConfigurationForPNS(t, LongWindow)
	var spectrum [maxSpectralLines]FixpDBL
	var maxScale [maxSFB]int
	var bandEnergy [maxSFB]FixpDBL
	var bandEnergyLD64 [maxSFB]FixpDBL
	var tonality [maxSFB]FixpSGL
	var scratch TonalityScratch
	fillTonalitySpectrum(spectrum[:], 113)
	FDKaacEncCalcSfbMaxScaleSpec(spectrum[:], conf.SfbOffset[:], maxScale[:], conf.SfbCnt)
	FDKaacEncCalcBandEnergyOptimLong(spectrum[:], maxScale[:], conf.SfbOffset[:], conf.SfbCnt, bandEnergy[:], bandEnergyLD64[:])

	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncCalculateFullTonality(spectrum[:], maxScale[:], bandEnergyLD64[:], tonality[:], conf.SfbCnt, conf.SfbOffset[:], 1, &scratch)
		tonalityHashSink ^= hashFixpDBL(scratch.ChaosMeasurePerLine[:conf.SfbOffset[conf.SfbCnt]])
		tonalitySGLHashSink ^= hashFixpSGL(tonality[:conf.SfbCnt])
	})
	if allocs != 0 {
		t.Fatalf("full tonality allocations = %v, want 0", allocs)
	}
}

func fillTonalitySpectrum(spectrum []FixpDBL, seed int32) {
	state := uint32(seed)
	for i := range spectrum {
		state = state*1664525 + 1013904223
		mag := FixpDBL((state>>8)&0x001fffff) + FixpDBL((i%19+1)<<13)
		if i%7 == 0 {
			mag >>= 3
		}
		if (state >> 31) != 0 {
			mag = -mag
		}
		spectrum[i] = mag
	}
}
