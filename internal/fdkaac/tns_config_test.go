package fdkaac

import "testing"

var (
	tnsConfigHashSink uint64
	tnsConfigIntSink  int
)

func TestFDKaacEncInitTnsConfigurationAACLCVectors(t *testing.T) {
	longPsy := mustPsyConfigurationForTNS(t, LongWindow)
	var long TNSConfig
	if got := FDKaacEncInitTnsConfiguration(64000, 48000, 2, LongWindow, 1024, 0, 0, &long, &longPsy, 1, 1); got != AACEncOK {
		t.Fatalf("long TNS config rc = %#x, want OK", got)
	}
	if long.TnsActive != 1 || long.IsLowDelay != 0 || long.MaxOrder != 12 || long.CoefRes != 4 {
		t.Fatalf("long TNS core = active:%d ld:%d order:%d coef:%d", long.TnsActive, long.IsLowDelay, long.MaxOrder, long.CoefRes)
	}
	if long.LpcStopBand != 40 || long.LpcStopLine != 672 {
		t.Fatalf("long TNS stop = band:%d line:%d, want 40/672", long.LpcStopBand, long.LpcStopLine)
	}
	if long.LpcStartBand != [maxTnsFilters]int{23, 8} || long.LpcStartLine != [maxTnsFilters]int{176, 32} {
		t.Fatalf("long TNS starts = bands:%v lines:%v", long.LpcStartBand, long.LpcStartLine)
	}
	wantEnabled := [maxTnsFilters]int{1, 1}
	wantThresh := [maxTnsFilters]int{1437, 1500}
	wantLimit := [maxTnsFilters]int{12, 5}
	wantDirection := [maxTnsFilters]int{0, 0}
	wantSplit := [maxTnsFilters]int{-1, -1}
	if long.ConfTab.FilterEnabled != wantEnabled ||
		long.ConfTab.ThreshOn != wantThresh ||
		long.ConfTab.TnsLimitOrder != wantLimit ||
		long.ConfTab.TnsFilterDirection != wantDirection ||
		long.ConfTab.ACFSplit != wantSplit ||
		long.ConfTab.SeperateFiltersAllowed != 1 {
		t.Fatalf("long TNS conf tab = %+v", long.ConfTab)
	}
	if got, want := hashFixpDBL(long.ACFWindow[tnsHiFilt][:]), uint64(0xd4bf806eb7b5c9df); got != want {
		t.Fatalf("long high ACF hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(long.ACFWindow[tnsLoFilt][:]), uint64(0xd4bf806eb7b5c9df); got != want {
		t.Fatalf("long low ACF hash = %#016x, want %#016x", got, want)
	}

	shortPsy := mustPsyConfigurationForTNS(t, ShortWindow)
	var short TNSConfig
	if got := FDKaacEncInitTnsConfiguration(64000, 48000, 2, ShortWindow, 1024, 0, 0, &short, &shortPsy, 1, 1); got != AACEncOK {
		t.Fatalf("short TNS config rc = %#x, want OK", got)
	}
	if short.TnsActive != 1 || short.MaxOrder != 5 || short.CoefRes != 3 {
		t.Fatalf("short TNS core = active:%d order:%d coef:%d", short.TnsActive, short.MaxOrder, short.CoefRes)
	}
	if short.LpcStopBand != 12 || short.LpcStopLine != 96 {
		t.Fatalf("short TNS stop = band:%d line:%d, want 12/96", short.LpcStopBand, short.LpcStopLine)
	}
	if short.LpcStartBand != [maxTnsFilters]int{5, 0} || short.LpcStartLine != [maxTnsFilters]int{20, 0} {
		t.Fatalf("short TNS starts = bands:%v lines:%v", short.LpcStartBand, short.LpcStartLine)
	}
	wantShortLimit := [maxTnsFilters]int{5, 0}
	if short.ConfTab.FilterEnabled != wantEnabled ||
		short.ConfTab.ThreshOn != wantThresh ||
		short.ConfTab.TnsLimitOrder != wantShortLimit ||
		short.ConfTab.TnsFilterDirection != wantDirection ||
		short.ConfTab.ACFSplit != wantSplit ||
		short.ConfTab.SeperateFiltersAllowed != 1 {
		t.Fatalf("short TNS conf tab = %+v", short.ConfTab)
	}
	if got, want := hashFixpDBL(short.ACFWindow[tnsHiFilt][:]), uint64(0x36d6cb1e76643c61); got != want {
		t.Fatalf("short high ACF hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(short.ACFWindow[tnsLoFilt][:]), uint64(0x36d6cb1e76643c61); got != want {
		t.Fatalf("short low ACF hash = %#016x, want %#016x", got, want)
	}
	if short.ACFWindow[tnsHiFilt][8] != 0 || short.ACFWindow[tnsLoFilt][8] != 0 {
		t.Fatalf("short ACF tail not zeroed: %d/%d", short.ACFWindow[tnsHiFilt][8], short.ACFWindow[tnsLoFilt][8])
	}
}

func TestFDKaacEncInitTnsConfigurationLowBitrateAndInactive(t *testing.T) {
	longPsy := mustPsyConfigurationForTNS(t, LongWindow)
	var conf TNSConfig
	if got := FDKaacEncInitTnsConfiguration(12000, 48000, 1, LongWindow, 1024, 0, 0, &conf, &longPsy, 0, 0); got != AACEncOK {
		t.Fatalf("low bitrate TNS config rc = %#x, want OK", got)
	}
	if conf.TnsActive != 0 || conf.MaxOrder != 10 || conf.CoefRes != 4 {
		t.Fatalf("low bitrate core = active:%d order:%d coef:%d", conf.TnsActive, conf.MaxOrder, conf.CoefRes)
	}
	if conf.ConfTab.TnsLimitOrder != [maxTnsFilters]int{10, 3} {
		t.Fatalf("low bitrate TNS limits = %v, want [10 3]", conf.ConfTab.TnsLimitOrder)
	}
}

func TestFDKaacEncFreqToBandWidthRoundingVector(t *testing.T) {
	longPsy := mustPsyConfigurationForTNS(t, LongWindow)
	if got := FDKaacEncFreqToBandWidthRounding(1400, 48000, longPsy.SfbCnt, longPsy.SfbOffset[:]); got != 12 {
		t.Fatalf("1400 Hz rounded band = %d, want 12", got)
	}
	if got := FDKaacEncFreqToBandWidthRounding(600, 48000, longPsy.SfbCnt, longPsy.SfbOffset[:]); got != 6 {
		t.Fatalf("600 Hz rounded band = %d, want 6", got)
	}
	if got := FDKaacEncFreqToBandWidthRounding(96000, 48000, longPsy.SfbCnt, longPsy.SfbOffset[:]); got != longPsy.SfbCnt {
		t.Fatalf("overflow rounded band = %d, want %d", got, longPsy.SfbCnt)
	}
}

func TestFDKaacEncTnsConfigurationRejectsMalformedControls(t *testing.T) {
	longPsy := mustPsyConfigurationForTNS(t, LongWindow)
	var conf TNSConfig
	if got := FDKaacEncInitTnsConfiguration(64000, 48000, 0, LongWindow, 1024, 0, 0, &conf, &longPsy, 1, 1); got != tnsInternalError {
		t.Fatalf("bad channel TNS rc = %#x, want %#x", got, tnsInternalError)
	}
	if got := FDKaacEncInitTnsConfiguration(64000, 48000, 2, LongWindow, 512, 0, 0, &conf, &longPsy, 1, 1); got != AACEncInvalidFrameLength {
		t.Fatalf("unsupported frame-length TNS rc = %#x, want %#x", got, AACEncInvalidFrameLength)
	}
	expectAACEncPanic(t, func() { FDKaacEncInitTnsConfiguration(64000, 48000, 2, -1, 1024, 0, 0, &conf, &longPsy, 1, 1) })
	expectAACEncPanic(t, func() { FDKaacEncInitTnsConfiguration(64000, 48000, 2, LongWindow, 1024, 0, 0, nil, &longPsy, 1, 1) })
	expectAACEncPanic(t, func() { FDKaacEncInitTnsConfiguration(64000, 48000, 2, LongWindow, 1024, 0, 0, &conf, nil, 1, 1) })
	expectAACEncPanic(t, func() { FDKaacEncFreqToBandWidthRounding(-1, 48000, longPsy.SfbCnt, longPsy.SfbOffset[:]) })
	expectAACEncPanic(t, func() { FDKaacEncFreqToBandWidthRounding(1400, 0, longPsy.SfbCnt, longPsy.SfbOffset[:]) })
	expectAACEncPanic(t, func() { FDKaacEncFreqToBandWidthRounding(1400, 48000, longPsy.SfbCnt, longPsy.SfbOffset[:3]) })
}

func TestFDKaacEncInitTnsConfigurationAllocs(t *testing.T) {
	longPsy := mustPsyConfigurationForTNS(t, LongWindow)
	var conf TNSConfig
	allocs := testing.AllocsPerRun(100, func() {
		if got := FDKaacEncInitTnsConfiguration(64000, 48000, 2, LongWindow, 1024, 0, 0, &conf, &longPsy, 1, 1); got != AACEncOK {
			t.Fatalf("TNS config rc = %#x, want OK", got)
		}
		tnsConfigHashSink ^= hashFixpDBL(conf.ACFWindow[tnsHiFilt][:])
		tnsConfigIntSink += conf.LpcStopBand
	})
	if allocs != 0 {
		t.Fatalf("FDKaacEncInitTnsConfiguration allocations = %v, want 0", allocs)
	}
}

func mustPsyConfigurationForTNS(t *testing.T, blockType int) PsyConfiguration {
	t.Helper()
	var conf PsyConfiguration
	useMS := 1
	if blockType == ShortWindow {
		useMS = 0
	}
	if got := FDKaacEncInitPsyConfiguration(64000, 48000, 15500, blockType, 1024, 1, useMS, &conf, FilterbankLC); got != AACEncOK {
		t.Fatalf("psy config rc = %#x, want OK", got)
	}
	return conf
}
