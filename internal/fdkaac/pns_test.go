package fdkaac

import "testing"

var pnsRuntimeHashSink uint64
var pnsRuntimeIntSink int

func TestFDKaacEncNoiseDetectVectors(t *testing.T) {
	c := buildPnsRuntimeCase(t, 211)

	FDKaacEncNoiseDetect(c.spectrum[:], c.maxScale[:], c.conf.SfbActive, c.conf.SfbOffset[:], c.data.NoiseFuzzyMeasure[:], &c.pns.NP, c.tonality[:])

	if got, want := hashFixpSGL(c.data.NoiseFuzzyMeasure[:c.conf.SfbActive]), uint64(0x36a48878bd512e87); got != want {
		t.Fatalf("noise fuzzy hash = %#016x, want %#016x", got, want)
	}
	if got, want := c.data.NoiseFuzzyMeasure[c.pns.NP.StartSfb], FixpSGL(0); got != want {
		t.Fatalf("noise fuzzy start = %d, want %d", got, want)
	}
}

func TestFDKaacEncPnsDetectAndCodeVectors(t *testing.T) {
	c := buildPnsRuntimeCase(t, 313)
	for sfb := 0; sfb < c.conf.SfbActive; sfb++ {
		c.thresholdLD[sfb] = c.energyLD[sfb] - ldDataStep2Over64
	}

	FDKaacEncPnsDetect(&c.pns, &c.data, LongWindow, c.conf.SfbActive, c.conf.SfbActive, c.thresholdLD[:], c.conf.SfbOffset[:], c.spectrum[:], c.maxScale[:], c.tonality[:], 1, 0, 0, c.energyLD[:], c.noise[:])

	if got, want := hashBandEnergyInts(c.data.PNSFlag[:c.conf.SfbActive]), uint64(0xc45ebb8ea3c56624); got != want {
		t.Fatalf("pns flag hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashBandEnergyInts(c.noise[:c.conf.SfbActive]), uint64(0x0051e2fcfce81d9c); got != want {
		t.Fatalf("noise energy hash = %#016x, want %#016x", got, want)
	}

	FDKaacEncCodePnsChannel(c.conf.SfbActive, &c.pns, c.data.PNSFlag[:], c.energyLD[:], c.noise[:], c.thresholdLD[:])

	if got, want := hashBandEnergyInts(c.noise[:c.conf.SfbActive]), uint64(0x0051e2fcfce81d9c); got != want {
		t.Fatalf("coded noise energy hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(c.thresholdLD[:c.conf.SfbActive]), uint64(0x80573a86bc3ab9d0); got != want {
		t.Fatalf("coded threshold hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPnsDetectDisabledAndShortNoops(t *testing.T) {
	c := buildPnsRuntimeCase(t, 419)
	for i := range c.data.PNSFlag {
		c.data.PNSFlag[i] = 1
	}
	for i := range c.noise {
		c.noise[i] = 77
	}
	c.pns.UsePns = 0

	FDKaacEncPnsDetect(&c.pns, &c.data, LongWindow, c.conf.SfbActive, c.conf.SfbActive, nil, nil, nil, nil, nil, 0, 0, 0, nil, c.noise[:])

	if got := hashBandEnergyInts(c.data.PNSFlag[:]); got != hashBandEnergyInts(make([]int, maxGroupedSFB)) {
		t.Fatalf("disabled pns flags were not cleared")
	}
	for i, got := range c.noise {
		if got != noNoisePNS {
			t.Fatalf("disabled noise[%d] = %d, want %d", i, got, noNoisePNS)
		}
	}

	c = buildPnsRuntimeCase(t, 421)
	FDKaacEncPnsDetect(&c.pns, &c.data, ShortWindow, c.conf.SfbActive, c.conf.SfbActive, c.thresholdLD[:], c.conf.SfbOffset[:], c.spectrum[:], c.maxScale[:], c.tonality[:], 0, 0, 0, c.energyLD[:], c.noise[:])
	if got := hashBandEnergyInts(c.data.PNSFlag[:]); got != hashBandEnergyInts(make([]int, maxGroupedSFB)) {
		t.Fatalf("short lc pns flags were not cleared")
	}
}

func TestFDKaacEncPnsChannelPairVectors(t *testing.T) {
	left := buildPnsRuntimeCase(t, 503)
	right := buildPnsRuntimeCase(t, 607)
	var mid [maxSFB]FixpDBL
	var msMask [maxGroupedSFB]int
	msDigest := MsMaskNone
	for i := 0; i < left.conf.SfbActive; i++ {
		mid[i] = (left.energy[i] >> 1) + (right.energy[i] >> 2)
		left.data.PNSFlag[i] = 1
		right.data.PNSFlag[i] = 1
		msMask[i] = i & 1
	}

	FDKaacEncPreProcessPnsChannelPair(left.conf.SfbActive, left.energy[:], right.energy[:], left.energyLD[:], right.energyLD[:], mid[:], &left.pns, &left.data, &right.data)

	if got, want := hashFixpDBL(left.data.NoiseEnergyCorrelation[:left.conf.SfbActive]), uint64(0x3d9a596f8c02c66b); got != want {
		t.Fatalf("noise correlation hash = %#016x, want %#016x", got, want)
	}

	FDKaacEncPostProcessPnsChannelPair(left.conf.SfbActive, &left.pns, &left.data, &right.data, msMask[:], &msDigest)

	if got, want := hashBandEnergyInts(msMask[:left.conf.SfbActive]), uint64(0xf04404e603a47365); got != want {
		t.Fatalf("post ms mask hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashBandEnergyInts(left.data.PNSFlag[:left.conf.SfbActive]), uint64(0x81b67dfb340bd465); got != want {
		t.Fatalf("post left pns hash = %#016x, want %#016x", got, want)
	}
	if msDigest != MsMaskSome {
		t.Fatalf("post ms digest = %d, want %d", msDigest, MsMaskSome)
	}
}

func TestFDKaacEncPnsRuntimeRejectsInvalid(t *testing.T) {
	c := buildPnsRuntimeCase(t, 709)
	tests := []struct {
		name string
		fn   func()
	}{
		{"nil noise params", func() {
			FDKaacEncNoiseDetect(c.spectrum[:], c.maxScale[:], c.conf.SfbActive, c.conf.SfbOffset[:], c.data.NoiseFuzzyMeasure[:], nil, c.tonality[:])
		}},
		{"short fuzzy", func() {
			FDKaacEncNoiseDetect(c.spectrum[:], c.maxScale[:], c.conf.SfbActive, c.conf.SfbOffset[:], c.data.NoiseFuzzyMeasure[:c.conf.SfbActive-1], &c.pns.NP, c.tonality[:])
		}},
		{"nil pns data", func() {
			FDKaacEncPnsDetect(&c.pns, nil, LongWindow, c.conf.SfbActive, c.conf.SfbActive, c.thresholdLD[:], c.conf.SfbOffset[:], c.spectrum[:], c.maxScale[:], c.tonality[:], 0, 0, 0, c.energyLD[:], c.noise[:])
		}},
		{"short noise reset", func() {
			FDKaacEncPnsDetect(&c.pns, &c.data, LongWindow, c.conf.SfbActive, c.conf.SfbActive, c.thresholdLD[:], c.conf.SfbOffset[:], c.spectrum[:], c.maxScale[:], c.tonality[:], 0, 0, 0, c.energyLD[:], c.noise[:maxGroupedSFB-1])
		}},
		{"bad group width", func() {
			FDKaacEncPnsDetect(&c.pns, &c.data, LongWindow, c.conf.SfbActive, 1, c.thresholdLD[:], c.conf.SfbOffset[:], c.spectrum[:], c.maxScale[:], c.tonality[:], 0, 0, 0, c.energyLD[:], c.noise[:])
		}},
		{"short code flags", func() {
			FDKaacEncCodePnsChannel(c.conf.SfbActive, &c.pns, c.data.PNSFlag[:c.conf.SfbActive-1], c.energyLD[:], c.noise[:], c.thresholdLD[:])
		}},
		{"nil pair data", func() {
			FDKaacEncPreProcessPnsChannelPair(c.conf.SfbActive, c.energy[:], c.energy[:], c.energyLD[:], c.energyLD[:], c.energy[:], &c.pns, nil, &c.data)
		}},
		{"nil digest", func() {
			FDKaacEncPostProcessPnsChannelPair(c.conf.SfbActive, &c.pns, &c.data, &c.data, c.data.PNSFlag[:], nil)
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

func TestFDKaacEncPnsRuntimeAllocs(t *testing.T) {
	c := buildPnsRuntimeCase(t, 811)
	for sfb := 0; sfb < c.conf.SfbActive; sfb++ {
		c.thresholdLD[sfb] = c.energyLD[sfb] - ldDataStep2Over64
	}

	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncPnsDetect(&c.pns, &c.data, LongWindow, c.conf.SfbActive, c.conf.SfbActive, c.thresholdLD[:], c.conf.SfbOffset[:], c.spectrum[:], c.maxScale[:], c.tonality[:], 1, 0, 0, c.energyLD[:], c.noise[:])
		FDKaacEncCodePnsChannel(c.conf.SfbActive, &c.pns, c.data.PNSFlag[:], c.energyLD[:], c.noise[:], c.thresholdLD[:])
		pnsRuntimeHashSink ^= hashFixpSGL(c.data.NoiseFuzzyMeasure[:c.conf.SfbActive])
		pnsRuntimeIntSink += c.noise[0]
	})
	if allocs != 0 {
		t.Fatalf("pns runtime allocations = %v, want 0", allocs)
	}
}

type pnsRuntimeCase struct {
	conf        PsyConfiguration
	pns         PNSConfig
	data        PNSData
	spectrum    [maxSpectralLines]FixpDBL
	maxScale    [maxSFB]int
	energy      [maxSFB]FixpDBL
	energyLD    [maxSFB]FixpDBL
	thresholdLD [maxSFB]FixpDBL
	tonality    [maxSFB]FixpSGL
	noise       [maxGroupedSFB]int
	scratch     TonalityScratch
}

func buildPnsRuntimeCase(t *testing.T, seed int32) pnsRuntimeCase {
	t.Helper()
	var c pnsRuntimeCase
	c.conf = mustPsyConfigurationForPNS(t, LongWindow)
	if got := FDKaacEncInitPnsConfiguration(&c.pns, 48000, 48000, 1, c.conf.SfbCnt, c.conf.SfbOffset[:], 2, 1); got != AACEncOK {
		t.Fatalf("PNS config rc = %#x, want OK", got)
	}
	fillTonalitySpectrum(c.spectrum[:], seed)
	FDKaacEncCalcSfbMaxScaleSpec(c.spectrum[:], c.conf.SfbOffset[:], c.maxScale[:], c.conf.SfbCnt)
	FDKaacEncCalcBandEnergyOptimLong(c.spectrum[:], c.maxScale[:], c.conf.SfbOffset[:], c.conf.SfbCnt, c.energy[:], c.energyLD[:])
	FDKaacEncCalculateFullTonality(c.spectrum[:], c.maxScale[:], c.energyLD[:], c.tonality[:], c.conf.SfbCnt, c.conf.SfbOffset[:], 1, &c.scratch)
	for i := 0; i < c.conf.SfbActive; i++ {
		c.thresholdLD[i] = c.energyLD[i] - ldDataStep1Over64
	}
	return c
}
