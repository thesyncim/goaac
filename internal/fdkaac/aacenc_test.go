package fdkaac

import "testing"

var aacencConfigSink int

func TestFDKaacEncAacInitDefaultConfigVector(t *testing.T) {
	var cfg AACEncConfig
	FDKaacEncAacInitDefaultConfig(&cfg)
	want := AACEncConfig{
		BitRate:         -1,
		AverageBits:     -1,
		BitrateMode:     QCBitrateModeCBR,
		UseTns:          tnsEnableMask,
		UsePns:          1,
		UseIS:           1,
		UseMS:           1,
		FrameLength:     -1,
		EpConfig:        -1,
		NSubFrames:      1,
		ChannelOrder:    ChannelOrderMPEG,
		ChannelMode:     ModeUnknown,
		MinBitsPerFrame: -1,
		MaxBitsPerFrame: -1,
		AudioMuxVersion: -1,
		DownscaleFactor: 1,
	}
	if cfg != want {
		t.Fatalf("default config = %+v, want %+v", cfg, want)
	}
}

func TestFDKaacEncFrameBitrateMathVectors(t *testing.T) {
	if got := FDKaacEncCalcBitsPerFrame(128000, 1024, 48000); got != 2730 {
		t.Fatalf("bits/frame = %d, want 2730", got)
	}
	if got := FDKaacEncCalcBitrate(2730, 1024, 48000); got != 127968 {
		t.Fatalf("bitrate = %d, want 127968", got)
	}
	if got := FDKaacEncCalcBitsPerFrame(96000, 960, 44100); got != 2089 {
		t.Fatalf("960-frame bits/frame = %d, want 2089", got)
	}
	if got := FDKaacEncCalcBitrate(2089, 960, 44100); got != 95963 {
		t.Fatalf("960-frame bitrate = %d, want 95963", got)
	}
}

func TestFDKaacEncLimitBitrateVectors(t *testing.T) {
	gotRate, gotAvg := FDKaacEncLimitBitrate(staticBits56, AOTAACLC, 48000, 1024, 2, 2, 128000, -1, QCBitrateModeCBR, 1)
	if gotRate != 128000 || gotAvg != 2730 {
		t.Fatalf("normal limit = %d/%d, want 128000/2730", gotRate, gotAvg)
	}
	gotRate, gotAvg = FDKaacEncLimitBitrate(staticBits208, AOTAACLC, 48000, 1024, 2, 2, 1000, -1, QCBitrateModeCBR, 1)
	if gotRate != 13500 || gotAvg != 288 {
		t.Fatalf("low limit = %d/%d, want 13500/288", gotRate, gotAvg)
	}
	gotRate, gotAvg = FDKaacEncLimitBitrate(staticBits56, AOTAACLC, 48000, 1024, 2, 2, 1000000, -1, QCBitrateModeCBR, 1)
	if gotRate != 576000 || gotAvg != 12288 {
		t.Fatalf("high limit = %d/%d, want 576000/12288", gotRate, gotAvg)
	}
}

func TestFDKaacEncPrepareQCInitFromConfigCBRVector(t *testing.T) {
	cfg := baseAACLCConfig(48000, 128000, 2, Mode2)
	var state AACEncInitState
	if errCode := FDKaacEncPrepareQCInitFromConfig(&state, &cfg, staticBits56); errCode != AACEncOK {
		t.Fatalf("prepare QC init error = %#x, want %#x", errCode, AACEncOK)
	}
	if state.AverageBitsPerFrame != 2730 ||
		state.PsyBitrate != 128000 ||
		state.Bandwidth90dB != 15500 ||
		state.TNSMask != tnsEnableMask ||
		state.InternalAOT != AOTAACLC {
		t.Fatalf("CBR init state = %+v", state)
	}
	wantQC := QCInit{
		ChannelMapping:      &state.ChannelMapping,
		MaxBits:             12288,
		AverageBits:         2736,
		BitRes:              9552,
		SampleRate:          48000,
		StaticBits:          56,
		BitrateMode:         QCBitrateModeCBR,
		MeanPe:              6613,
		InvQuant:            2,
		MaxIterations:       5,
		MaxBitFac:           75350016,
		Bitrate:             128000,
		NSubFrames:          1,
		BitResMode:          BitresModeFull,
		BitDistributionMode: BitDistributionModeIntraElement,
		PaddingRest:         48000,
	}
	if state.QCInit != wantQC {
		t.Fatalf("CBR QC init = %+v, want %+v", state.QCInit, wantQC)
	}
	if cfg.BandWidth != 15500 || cfg.MaxAncBytesPerAU != 256 {
		t.Fatalf("mutated CBR config bandwidth/maxAnc = %d/%d, want 15500/256", cfg.BandWidth, cfg.MaxAncBytesPerAU)
	}
}

func TestFDKaacEncPrepareQCInitFromConfigVBRVector(t *testing.T) {
	cfg := baseAACLCConfig(48000, 240000, 6, Mode1_2_2_1)
	cfg.BitrateMode = QCBitrateModeVBR3
	cfg.ChannelOrder = ChannelOrderWAV
	var state AACEncInitState
	if errCode := FDKaacEncPrepareQCInitFromConfig(&state, &cfg, staticBits64); errCode != AACEncOK {
		t.Fatalf("prepare VBR QC init error = %#x, want %#x", errCode, AACEncOK)
	}
	if state.AverageBitsPerFrame != 5120 ||
		state.PsyBitrate != 240000 ||
		state.Bandwidth90dB != 15750 ||
		state.ChannelMapping.NElements != 4 ||
		state.ChannelMapping.NChannelsEff != 5 {
		t.Fatalf("VBR init state = %+v", state)
	}
	wantQC := QCInit{
		ChannelMapping:      &state.ChannelMapping,
		MaxBits:             30720,
		AverageBits:         5120,
		BitRes:              30720,
		SampleRate:          48000,
		StaticBits:          64,
		BitrateMode:         QCBitrateModeVBR3,
		MeanPe:              6720,
		InvQuant:            2,
		MaxIterations:       5,
		MaxBitFac:           100663296,
		Bitrate:             240000,
		NSubFrames:          1,
		BitResMode:          BitresModeFull,
		BitDistributionMode: BitDistributionModeInterElement,
		PaddingRest:         48000,
	}
	if state.QCInit != wantQC {
		t.Fatalf("VBR QC init = %+v, want %+v", state.QCInit, wantQC)
	}
	if cfg.BandWidth != 15750 || cfg.MaxAncBytesPerAU != 256 {
		t.Fatalf("mutated VBR config bandwidth/maxAnc = %d/%d, want 15750/256", cfg.BandWidth, cfg.MaxAncBytesPerAU)
	}
}

func TestFDKaacEncPrepareQCInitFeedsQCInit(t *testing.T) {
	cfg := baseAACLCConfig(48000, 128000, 2, Mode2)
	var state AACEncInitState
	if errCode := FDKaacEncPrepareQCInitFromConfig(&state, &cfg, staticBits56); errCode != AACEncOK {
		t.Fatalf("prepare QC init error = %#x, want %#x", errCode, AACEncOK)
	}
	elementBits, elementBitsPtrs := initializedElementBits()
	adjState, _ := initializedQCInitAdjThrState()
	var kernel QCKernel
	if errCode := FDKaacEncQCInit(&kernel, adjState, elementBitsPtrs[:], &state.QCInit, 1); errCode != AACEncOK {
		t.Fatalf("QC init error = %#x, want %#x", errCode, AACEncOK)
	}
	if kernel.GlobHdrBits != 56 ||
		kernel.MaxBitsPerFrame != 12288 ||
		kernel.BitResTot != 9552 ||
		kernel.BitResTotMax != 9552 ||
		kernel.MaxIterations != 5 ||
		kernel.InvQuant != 2 ||
		kernel.MaxBitFac != 75350016 {
		t.Fatalf("kernel from prepared QC init = %+v", kernel)
	}
	assertElementBits(t, elementBits[:], []elementBitsWant{
		{chBitrate: 64000, maxBits: 12288, rel: MaxValDBL},
	})
	if adjState.AdjThrStateElem[0].PeMin != 2645 || adjState.AdjThrStateElem[0].PeMax != 3968 {
		t.Fatalf("threshold element from prepared QC init = %+v", *adjState.AdjThrStateElem[0])
	}
}

func TestFDKaacEncPrepareQCInitRejectsInvalid(t *testing.T) {
	valid := baseAACLCConfig(48000, 128000, 2, Mode2)
	tests := []struct {
		name string
		cfg  AACEncConfig
		want int
	}{
		{name: "bad sample rate", cfg: withAACLCConfig(valid, func(c *AACEncConfig) { c.SampleRate = 12345 }), want: AACEncUnsupportedSamplingRate},
		{name: "unset bitrate", cfg: withAACLCConfig(valid, func(c *AACEncConfig) { c.BitRate = -1 }), want: AACEncUnsupportedBitrate},
		{name: "unsupported cbr bitrate", cfg: withAACLCConfig(valid, func(c *AACEncConfig) { c.BitRate = 1000 }), want: AACEncUnsupportedBitrate},
		{name: "channel mismatch", cfg: withAACLCConfig(valid, func(c *AACEncConfig) { c.NChannels = 1 }), want: AACEncUnsupportedChannelConf},
		{name: "er vcb11", cfg: withAACLCConfig(valid, func(c *AACEncConfig) { c.SyntaxFlags = acERVCB11 }), want: AACEncUnsupportedERFormat},
		{name: "er hcr", cfg: withAACLCConfig(valid, func(c *AACEncConfig) { c.SyntaxFlags = acERHCR }), want: AACEncUnsupportedERFormat},
		{name: "bad frame length", cfg: withAACLCConfig(valid, func(c *AACEncConfig) { c.FrameLength = 512 }), want: AACEncInvalidFrameLength},
		{name: "bad bitrate mode", cfg: withAACLCConfig(valid, func(c *AACEncConfig) { c.BitrateMode = QCBitrateModeInvalid }), want: AACEncUnsupportedBitrateMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var state AACEncInitState
			if got := FDKaacEncPrepareQCInitFromConfig(&state, &tt.cfg, staticBits56); got != tt.want {
				t.Fatalf("error = %#x, want %#x", got, tt.want)
			}
		})
	}
	var cfg AACEncConfig
	var state AACEncInitState
	if got := FDKaacEncPrepareQCInitFromConfig(nil, &valid, staticBits56); got != AACEncInvalidHandle {
		t.Fatalf("nil state error = %#x, want %#x", got, AACEncInvalidHandle)
	}
	if got := FDKaacEncPrepareQCInitFromConfig(&state, nil, staticBits56); got != AACEncInvalidHandle {
		t.Fatalf("nil config error = %#x, want %#x", got, AACEncInvalidHandle)
	}
	if got := FDKaacEncPrepareQCInitFromConfig(&state, &cfg, staticBits56); got != AACEncUnsupportedChannelConf {
		t.Fatalf("zero config error = %#x, want %#x", got, AACEncUnsupportedChannelConf)
	}
}

func TestFDKaacEncAncillaryVectors(t *testing.T) {
	var bits int
	if errCode := FDKaacEncInitCheckAncillary(128000, 1024, -1, &bits, 48000); errCode != AACEncOK {
		t.Fatalf("default ancillary error = %#x, want %#x", errCode, AACEncOK)
	}
	if bits != 272 {
		t.Fatalf("default ancillary bits = %d, want 272", bits)
	}
	if errCode := FDKaacEncInitCheckAncillary(128000, 1024, 16000, &bits, 48000); errCode != AACEncOK {
		t.Fatalf("explicit ancillary error = %#x, want %#x", errCode, AACEncOK)
	}
	if bits != 336 {
		t.Fatalf("explicit ancillary bits = %d, want 336", bits)
	}
	if errCode := FDKaacEncInitCheckAncillary(128000, 1024, 19200, &bits, 48000); errCode != AACEncUnsupportedAncBitrate {
		t.Fatalf("high ancillary error = %#x, want %#x", errCode, AACEncUnsupportedAncBitrate)
	}
}

func TestFDKaacEncEncBitresToTpBitresVectors(t *testing.T) {
	kernel := QCKernel{BitResTot: 4096}
	if got := FDKaacEncEncBitresToTpBitres(&kernel, QCBitrateModeCBR, -1, 2); got != 4096 {
		t.Fatalf("CBR transport reservoir = %d, want 4096", got)
	}
	if got := FDKaacEncEncBitresToTpBitres(&kernel, QCBitrateModeVBR4, -1, 2); got != fdkIntMax {
		t.Fatalf("VBR transport reservoir = %d, want %d", got, fdkIntMax)
	}
	if got := FDKaacEncEncBitresToTpBitres(&kernel, QCBitrateModeSFR, -1, 2); got != 0 {
		t.Fatalf("SFR transport reservoir = %d, want 0", got)
	}
	if got := FDKaacEncEncBitresToTpBitres(&kernel, QCBitrateModeCBR, 2, 5); got != minBufSizePerEffChan*5 {
		t.Fatalf("audioMuxVersion 2 reservoir = %d, want %d", got, minBufSizePerEffChan*5)
	}
	expectAACEncPanic(t, func() {
		_ = FDKaacEncEncBitresToTpBitres(nil, QCBitrateModeCBR, -1, 2)
	})
}

func TestFDKaacEncPrepareQCInitAllocs(t *testing.T) {
	cfg := baseAACLCConfig(48000, 128000, 2, Mode2)
	var state AACEncInitState
	allocs := testing.AllocsPerRun(1000, func() {
		errCode := FDKaacEncPrepareQCInitFromConfig(&state, &cfg, staticBits56)
		aacencConfigSink = errCode + state.QCInit.MaxBits + int(cfg.MaxAncBytesPerAU)
	})
	if allocs != 0 {
		t.Fatalf("prepare QC init allocations = %v, want 0", allocs)
	}
}

func baseAACLCConfig(sampleRate int, bitrate int, channels int, mode ChannelMode) AACEncConfig {
	var cfg AACEncConfig
	FDKaacEncAacInitDefaultConfig(&cfg)
	cfg.SampleRate = sampleRate
	cfg.BitRate = bitrate
	cfg.AOT = AOTAACLC
	cfg.NChannels = channels
	cfg.ChannelMode = mode
	cfg.FrameLength = 1024
	cfg.UseRequant = 1
	return cfg
}

func withAACLCConfig(cfg AACEncConfig, edit func(*AACEncConfig)) AACEncConfig {
	edit(&cfg)
	return cfg
}

func staticBits56(int) int  { return 56 }
func staticBits64(int) int  { return 64 }
func staticBits208(int) int { return 208 }

func expectAACEncPanic(t *testing.T, run func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	run()
}
