package fdkaac

import "testing"

var (
	psyConfigHashSink uint64
	psyConfigIntSink  int
)

func TestFDKaacEncInitSfbTableAACLCVectors(t *testing.T) {
	var offsets [maxSFB + 1]int
	cnt := 0
	if got := FDKaacEncInitSfbTable(48000, LongWindow, 1024, offsets[:], &cnt); got != AACEncOK {
		t.Fatalf("long SFB init rc = %#x, want OK", got)
	}
	wantLong := [...]int{
		0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 48, 56, 64, 72, 80, 88,
		96, 108, 120, 132, 144, 160, 176, 196, 216, 240, 264, 292, 320,
		352, 384, 416, 448, 480, 512, 544, 576, 608, 640, 672, 704, 736,
		768, 800, 832, 864, 896, 928, 1024,
	}
	if cnt != 49 {
		t.Fatalf("long SFB count = %d, want 49", cnt)
	}
	for i := range wantLong {
		if offsets[i] != wantLong[i] {
			t.Fatalf("long SFB offset[%d] = %d, want %d", i, offsets[i], wantLong[i])
		}
	}

	offsets = [maxSFB + 1]int{}
	cnt = 0
	if got := FDKaacEncInitSfbTable(44100, ShortWindow, 1024, offsets[:], &cnt); got != AACEncOK {
		t.Fatalf("short SFB init rc = %#x, want OK", got)
	}
	wantShort := [...]int{0, 4, 8, 12, 16, 20, 28, 36, 44, 56, 68, 80, 96, 112, 128}
	if cnt != 14 {
		t.Fatalf("short SFB count = %d, want 14", cnt)
	}
	for i := range wantShort {
		if offsets[i] != wantShort[i] {
			t.Fatalf("short SFB offset[%d] = %d, want %d", i, offsets[i], wantShort[i])
		}
	}

	offsets = [maxSFB + 1]int{}
	cnt = 0
	if got := FDKaacEncInitSfbTable(48000, LongWindow, 960, offsets[:], &cnt); got != AACEncOK {
		t.Fatalf("960 SFB init rc = %#x, want OK", got)
	}
	if cnt != 49 || offsets[48] != 928 || offsets[49] != 960 {
		t.Fatalf("960 SFB clamp count/tail = %d/%d/%d, want 49/928/960", cnt, offsets[48], offsets[49])
	}
}

func TestFDKaacEncInitPsyConfigurationAACLCVectors(t *testing.T) {
	var long PsyConfiguration
	if got := FDKaacEncInitPsyConfiguration(64000, 48000, 15500, LongWindow, 1024, 1, 1, &long, FilterbankLC); got != AACEncOK {
		t.Fatalf("long psy config rc = %#x, want OK", got)
	}
	if long.SfbCnt != 49 || long.SfbActive != 40 || long.SfbActiveLFE != 3 {
		t.Fatalf("long SFB counts = cnt:%d active:%d lfe:%d", long.SfbCnt, long.SfbActive, long.SfbActiveLFE)
	}
	if !long.AllowIS || !long.AllowMS || long.Filterbank != FilterbankLC || long.GranuleLength != 1024 {
		t.Fatalf("long tool/config flags = is:%v ms:%v fb:%d len:%d", long.AllowIS, long.AllowMS, long.Filterbank, long.GranuleLength)
	}
	if long.LowpassLine != 661 || long.LowpassLineLFE != lfeLowpassLine {
		t.Fatalf("long lowpass = %d/%d, want 661/%d", long.LowpassLine, long.LowpassLineLFE, lfeLowpassLine)
	}
	if long.ClipEnergy != psyClipEnergyLong || long.MaxAllowedIncreaseFactor != 2 || long.MinRemainingThresholdFactor != 0x0148 {
		t.Fatalf("long ratio controls = clip:%#x inc:%d remain:%#x", uint32(long.ClipEnergy), long.MaxAllowedIncreaseFactor, uint16(long.MinRemainingThresholdFactor))
	}
	wantPCM := [...]FixpDBL{22135176, 22135176, 22135176, 22135176, 44270352, 44270352, 177081408, 531244224}
	gotPCM := [...]FixpDBL{
		long.SfbPcmQuantThreshold[0],
		long.SfbPcmQuantThreshold[1],
		long.SfbPcmQuantThreshold[8],
		long.SfbPcmQuantThreshold[9],
		long.SfbPcmQuantThreshold[10],
		long.SfbPcmQuantThreshold[11],
		long.SfbPcmQuantThreshold[47],
		long.SfbPcmQuantThreshold[48],
	}
	if gotPCM != wantPCM {
		t.Fatalf("long PCM thresholds = %v, want %v", gotPCM, wantPCM)
	}
	if got, want := hashFixpDBL(long.SfbMaskLowFactor[:long.SfbCnt]), uint64(0x7111cbefa992a814); got != want {
		t.Fatalf("long low spreading hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(long.SfbMaskHighFactor[:long.SfbCnt]), uint64(0x0adf4a1649ae7766); got != want {
		t.Fatalf("long high spreading hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(long.SfbMaskLowFactorSprEn[:long.SfbCnt]), uint64(0x7111cbefa992a814); got != want {
		t.Fatalf("long spread-energy low hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(long.SfbMaskHighFactorSprEn[:long.SfbCnt]), uint64(0x341ab26b8f8c1e85); got != want {
		t.Fatalf("long spread-energy high hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(long.SfbMinSnrLdData[:long.SfbActive]), uint64(0x7123f603a5f99485); got != want {
		t.Fatalf("long min-SNR hash = %#016x, want %#016x", got, want)
	}

	var short PsyConfiguration
	if got := FDKaacEncInitPsyConfiguration(64000, 48000, 15500, ShortWindow, 1024, 1, 0, &short, FilterbankLC); got != AACEncOK {
		t.Fatalf("short psy config rc = %#x, want OK", got)
	}
	if short.SfbCnt != 14 || short.SfbActive != 12 || short.SfbActiveLFE != 0 {
		t.Fatalf("short SFB counts = cnt:%d active:%d lfe:%d", short.SfbCnt, short.SfbActive, short.SfbActiveLFE)
	}
	if !short.AllowIS || short.AllowMS {
		t.Fatalf("short tool flags = is:%v ms:%v", short.AllowIS, short.AllowMS)
	}
	if short.LowpassLine != 82 || short.LowpassLineLFE != 0 || short.ClipEnergy != psyClipEnergyLong>>6 {
		t.Fatalf("short lowpass/clip = %d/%d/%#x", short.LowpassLine, short.LowpassLineLFE, uint32(short.ClipEnergy))
	}
	if got, want := hashFixpDBL(short.SfbMaskLowFactor[:short.SfbCnt]), uint64(0x908de443b85cc476); got != want {
		t.Fatalf("short low spreading hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(short.SfbMaskHighFactor[:short.SfbCnt]), uint64(0x771e7c381824e401); got != want {
		t.Fatalf("short high spreading hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(short.SfbMaskLowFactorSprEn[:short.SfbCnt]), uint64(0x385a17f76bdcb2c3); got != want {
		t.Fatalf("short spread-energy low hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(short.SfbMaskHighFactorSprEn[:short.SfbCnt]), uint64(0x771e7c381824e401); got != want {
		t.Fatalf("short spread-energy high hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(short.SfbMinSnrLdData[:short.SfbActive]), uint64(0x2bc5344f5cf59475); got != want {
		t.Fatalf("short min-SNR hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPsyConfigurationRejectsMalformedControls(t *testing.T) {
	var offsets [maxSFB + 1]int
	cnt := 0
	if got := FDKaacEncInitSfbTable(12345, LongWindow, 1024, offsets[:], &cnt); got != AACEncUnsupportedSamplingRate {
		t.Fatalf("unsupported SFB sample rate rc = %#x, want %#x", got, AACEncUnsupportedSamplingRate)
	}
	if got := FDKaacEncInitSfbTable(48000, LongWindow, 512, offsets[:], &cnt); got != AACEncInvalidFrameLength {
		t.Fatalf("unsupported SFB frame length rc = %#x, want %#x", got, AACEncInvalidFrameLength)
	}
	expectAACEncPanic(t, func() { FDKaacEncInitSfbTable(48000, -1, 1024, offsets[:], &cnt) })
	expectAACEncPanic(t, func() { FDKaacEncInitSfbTable(48000, LongWindow, 1024, offsets[:3], &cnt) })
	expectAACEncPanic(t, func() { FDKaacEncInitSfbTable(48000, LongWindow, 1024, offsets[:], nil) })

	var conf PsyConfiguration
	expectAACEncPanic(t, func() { FDKaacEncInitPsyConfiguration(64000, 48000, 15500, -1, 1024, 1, 1, &conf, FilterbankLC) })
	expectAACEncPanic(t, func() { FDKaacEncInitPsyConfiguration(64000, 48000, 15500, LongWindow, 960, 1, 1, &conf, FilterbankLC) })
	expectAACEncPanic(t, func() { FDKaacEncInitPsyConfiguration(64000, 48000, 0, LongWindow, 1024, 1, 1, &conf, FilterbankLC) })
	expectAACEncPanic(t, func() { FDKaacEncInitPsyConfiguration(64000, 48000, 15500, LongWindow, 1024, 1, 1, nil, FilterbankLC) })
	expectAACEncPanic(t, func() { FDKaacEncInitPsyConfiguration(64000, 48000, 15500, LongWindow, 1024, 1, 1, &conf, 99) })
}

func TestFDKaacEncInitPsyConfigurationAllocs(t *testing.T) {
	var conf PsyConfiguration
	allocs := testing.AllocsPerRun(100, func() {
		if got := FDKaacEncInitPsyConfiguration(64000, 48000, 15500, LongWindow, 1024, 1, 1, &conf, FilterbankLC); got != AACEncOK {
			t.Fatalf("psy config rc = %#x, want OK", got)
		}
		psyConfigHashSink ^= hashFixpDBL(conf.SfbMinSnrLdData[:conf.SfbActive])
		psyConfigIntSink += conf.SfbActive
	})
	if allocs != 0 {
		t.Fatalf("FDKaacEncInitPsyConfiguration allocations = %v, want 0", allocs)
	}
}
