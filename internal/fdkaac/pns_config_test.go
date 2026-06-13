package fdkaac

import "testing"

var (
	pnsConfigHashSink uint64
	pnsConfigIntSink  int
)

func TestFDKaacEncLookUpPnsUseVectors(t *testing.T) {
	tests := []struct {
		name       string
		bitRate    int
		sampleRate int
		numChan    int
		isLC       int
		want       int
	}{
		{"LC enabled exact 48k", 48000, 48000, 2, 1, 4},
		{"LC disabled high", 64000, 48000, 2, 1, 0},
		{"LC default sample-rate fallback", 48000, 12345, 2, 1, 4},
		{"stereo non-LC mid", 30000, 44100, 2, 0, 5},
		{"mono non-LC low", 17000, 22050, 1, 0, 1},
		{"non-LC unsupported sample-rate", 17000, 12345, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FDKaacEncLookUpPnsUse(tt.bitRate, tt.sampleRate, tt.numChan, tt.isLC); got != tt.want {
				t.Fatalf("PNS use = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFDKaacEncInitPnsConfigurationAACLCVectors(t *testing.T) {
	psy := mustPsyConfigurationForPNS(t, LongWindow)
	var conf PNSConfig
	if got := FDKaacEncInitPnsConfiguration(&conf, 48000, 48000, 1, psy.SfbCnt, psy.SfbOffset[:], 2, 1); got != AACEncOK {
		t.Fatalf("PNS config rc = %#x, want OK", got)
	}
	if conf.UsePns != 1 {
		t.Fatalf("UsePns = %d, want 1", conf.UsePns)
	}
	if conf.MinCorrelationEnergy != 0 || conf.NoiseCorrelationThresh != pnsNoiseCorrelationThr {
		t.Fatalf("correlation controls = %#x/%#x", uint32(conf.MinCorrelationEnergy), uint32(conf.NoiseCorrelationThresh))
	}
	wantFlags := uint16(pnsIsLowComplexity | pnsUsePowerDistribution | pnsUsePsychTonality | pnsUseTnsGainThreshold | pnsUseTnsPNS | pnsJustLongWindow)
	if conf.NP.DetectionAlgorithmFlags != wantFlags {
		t.Fatalf("PNS flags = %#x, want %#x", conf.NP.DetectionAlgorithmFlags, wantFlags)
	}
	if conf.NP.StartSfb != 23 {
		t.Fatalf("start SFB = %d, want 23", conf.NP.StartSfb)
	}
	if conf.NP.RefPower != FXSGL2FXDBL(0x199a) || conf.NP.RefTonality != FXSGL2FXDBL(0x0ccd) {
		t.Fatalf("ref power/tonality = %#x/%#x", uint32(conf.NP.RefPower), uint32(conf.NP.RefTonality))
	}
	if conf.NP.TnsGainThreshold != 1410 || conf.NP.TnsPNSGainThreshold != 1400 || conf.NP.MinSfbWidth != 16 || conf.NP.GapFillThr != 0x4000 {
		t.Fatalf("PNS scalar params = tns:%d pns:%d width:%d gap:%#x", conf.NP.TnsGainThreshold, conf.NP.TnsPNSGainThreshold, conf.NP.MinSfbWidth, uint16(conf.NP.GapFillThr))
	}
	if got, want := hashFixpSGL(conf.NP.PowDistPSDcurve[:psy.SfbCnt+1]), uint64(0x1445dfcb53e1d95b); got != want {
		t.Fatalf("PSD curve hash = %#016x, want %#016x", got, want)
	}
	if conf.NP.PowDistPSDcurve[psy.SfbCnt-1] != 0 || conf.NP.PowDistPSDcurve[psy.SfbCnt] != 0 {
		t.Fatalf("PSD curve tail = %d/%d, want 0/0", conf.NP.PowDistPSDcurve[psy.SfbCnt-1], conf.NP.PowDistPSDcurve[psy.SfbCnt])
	}
}

func TestFDKaacEncInitPnsConfigurationDisabledAACLCVector(t *testing.T) {
	psy := mustPsyConfigurationForPNS(t, LongWindow)
	var conf PNSConfig
	if got := FDKaacEncInitPnsConfiguration(&conf, 64000, 48000, 1, psy.SfbCnt, psy.SfbOffset[:], 2, 1); got != AACEncOK {
		t.Fatalf("PNS config rc = %#x, want OK", got)
	}
	if conf.UsePns != 0 {
		t.Fatalf("UsePns = %d, want 0", conf.UsePns)
	}
	if conf.NP.DetectionAlgorithmFlags != pnsIsLowComplexity {
		t.Fatalf("disabled flags = %#x, want low-complexity flag", conf.NP.DetectionAlgorithmFlags)
	}
	var zero [maxGroupedSFB]FixpSGL
	if hashFixpSGL(conf.NP.PowDistPSDcurve[:]) != hashFixpSGL(zero[:]) {
		t.Fatalf("disabled PSD curve not zero")
	}
}

func TestFDKaacEncPnsConfigurationRejectsMalformedControls(t *testing.T) {
	psy := mustPsyConfigurationForPNS(t, LongWindow)
	var conf PNSConfig
	expectAACEncPanic(t, func() { FDKaacEncInitPnsConfiguration(nil, 48000, 48000, 1, psy.SfbCnt, psy.SfbOffset[:], 2, 1) })
	expectAACEncPanic(t, func() { FDKaacEncInitPnsConfiguration(&conf, -1, 48000, 1, psy.SfbCnt, psy.SfbOffset[:], 2, 1) })
	expectAACEncPanic(t, func() { FDKaacEncInitPnsConfiguration(&conf, 48000, 0, 1, psy.SfbCnt, psy.SfbOffset[:], 2, 1) })
	expectAACEncPanic(t, func() { FDKaacEncInitPnsConfiguration(&conf, 48000, 48000, 1, 0, psy.SfbOffset[:], 2, 1) })
	expectAACEncPanic(t, func() { FDKaacEncInitPnsConfiguration(&conf, 48000, 48000, 1, psy.SfbCnt, psy.SfbOffset[:3], 2, 1) })

	var np NoiseParams
	usePns := 1
	expectAACEncPanic(t, func() { FDKaacEncGetPnsParam(nil, 48000, 48000, psy.SfbCnt, psy.SfbOffset[:], &usePns, 2, 1) })
	expectAACEncPanic(t, func() { FDKaacEncGetPnsParam(&np, 48000, 48000, psy.SfbCnt, psy.SfbOffset[:], nil, 2, 1) })
}

func TestFDKaacEncInitPnsConfigurationAllocs(t *testing.T) {
	psy := mustPsyConfigurationForPNS(t, LongWindow)
	var conf PNSConfig
	allocs := testing.AllocsPerRun(100, func() {
		if got := FDKaacEncInitPnsConfiguration(&conf, 48000, 48000, 1, psy.SfbCnt, psy.SfbOffset[:], 2, 1); got != AACEncOK {
			t.Fatalf("PNS config rc = %#x, want OK", got)
		}
		pnsConfigHashSink ^= hashFixpSGL(conf.NP.PowDistPSDcurve[:psy.SfbCnt+1])
		pnsConfigIntSink += conf.NP.StartSfb
	})
	if allocs != 0 {
		t.Fatalf("FDKaacEncInitPnsConfiguration allocations = %v, want 0", allocs)
	}
}

func mustPsyConfigurationForPNS(t *testing.T, blockType int) PsyConfiguration {
	t.Helper()
	var conf PsyConfiguration
	useMS := 1
	if blockType == ShortWindow {
		useMS = 0
	}
	if got := FDKaacEncInitPsyConfiguration(48000, 48000, 15500, blockType, 1024, 1, useMS, &conf, FilterbankLC); got != AACEncOK {
		t.Fatalf("psy config rc = %#x, want OK", got)
	}
	return conf
}

func hashFixpSGL(x []FixpSGL) uint64 {
	h := uint64(1469598103934665603)
	for _, v := range x {
		u := uint16(v)
		h = fnv64AddByte(h, byte(u))
		h = fnv64AddByte(h, byte(u>>8))
	}
	return h
}
