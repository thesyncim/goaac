package fdkaac

import "testing"

var (
	channelMapModeSink ChannelMode
	channelMapIntSink  int
	channelMapRelSink  FixpDBL
)

func TestFDKaacEncDetermineEncoderModeVectors(t *testing.T) {
	tests := []struct {
		name        string
		startMode   ChannelMode
		nChannels   int
		wantMode    ChannelMode
		wantErrCode int
	}{
		{name: "unknown mono", startMode: ModeUnknown, nChannels: 1, wantMode: Mode1, wantErrCode: AACEncOK},
		{name: "unknown 6.1", startMode: ModeUnknown, nChannels: 7, wantMode: Mode6_1, wantErrCode: AACEncOK},
		{name: "unknown first 7.1 table entry", startMode: ModeUnknown, nChannels: 8, wantMode: Mode1_2_2_2_1, wantErrCode: AACEncOK},
		{name: "explicit stereo", startMode: Mode2, nChannels: 2, wantMode: Mode2, wantErrCode: AACEncOK},
		{name: "explicit mismatched count", startMode: Mode2, nChannels: 1, wantMode: Mode2, wantErrCode: AACEncUnsupportedChannelConf},
		{name: "unknown unsupported count", startMode: ModeUnknown, nChannels: 9, wantMode: ModeInvalid, wantErrCode: AACEncUnsupportedChannelConf},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := tt.startMode
			gotErr := FDKaacEncDetermineEncoderMode(&mode, tt.nChannels)
			if gotErr != tt.wantErrCode || mode != tt.wantMode {
				t.Fatalf("mode/error = %v/%#x, want %v/%#x", mode, gotErr, tt.wantMode, tt.wantErrCode)
			}
		})
	}

	t.Run("nil mode", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = FDKaacEncDetermineEncoderMode(nil, 2)
	})
}

func TestFDKaacEncInitChannelMappingVectors(t *testing.T) {
	tests := []struct {
		name      string
		mode      ChannelMode
		order     ChannelOrder
		wantMode  ChannelMode
		wantCh    int
		wantEffCh int
		want      []channelMapElementWant
	}{
		{
			name:      "mono mpeg",
			mode:      Mode1,
			order:     ChannelOrderMPEG,
			wantMode:  Mode1,
			wantCh:    1,
			wantEffCh: 1,
			want: []channelMapElementWant{
				{elType: idSCE, tag: 0, nCh: 1, ch0: 0, rel: MaxValDBL},
			},
		},
		{
			name:      "stereo mpeg",
			mode:      Mode2,
			order:     ChannelOrderMPEG,
			wantMode:  Mode2,
			wantCh:    2,
			wantEffCh: 2,
			want: []channelMapElementWant{
				{elType: idCPE, tag: 0, nCh: 2, ch0: 0, ch1: 1, rel: MaxValDBL},
			},
		},
		{
			name:      "3.0 wav remap",
			mode:      Mode1_2,
			order:     ChannelOrderWAV,
			wantMode:  Mode1_2,
			wantCh:    3,
			wantEffCh: 3,
			want: []channelMapElementWant{
				{elType: idSCE, tag: 0, nCh: 1, ch0: 2, rel: relBits04},
				{elType: idCPE, tag: 0, nCh: 2, ch0: 0, ch1: 1, rel: relBits06},
			},
		},
		{
			name:      "3.0 mpeg pass-through",
			mode:      Mode1_2,
			order:     ChannelOrderMPEG,
			wantMode:  Mode1_2,
			wantCh:    3,
			wantEffCh: 3,
			want: []channelMapElementWant{
				{elType: idSCE, tag: 0, nCh: 1, ch0: 0, rel: relBits04},
				{elType: idCPE, tag: 0, nCh: 2, ch0: 1, ch1: 2, rel: relBits06},
			},
		},
		{
			name:      "5.1 wav remap",
			mode:      Mode1_2_2_1,
			order:     ChannelOrderWAV,
			wantMode:  Mode1_2_2_1,
			wantCh:    6,
			wantEffCh: 5,
			want: []channelMapElementWant{
				{elType: idSCE, tag: 0, nCh: 1, ch0: 2, rel: relBits024},
				{elType: idCPE, tag: 0, nCh: 2, ch0: 0, ch1: 1, rel: relBits035},
				{elType: idCPE, tag: 1, nCh: 2, ch0: 4, ch1: 5, rel: relBits035},
				{elType: idLFE, tag: 0, nCh: 1, ch0: 3, rel: relBits006},
			},
		},
		{
			name:      "7.1 top-front wav remap",
			mode:      Mode7_1TopFront,
			order:     ChannelOrderWAV,
			wantMode:  Mode7_1TopFront,
			wantCh:    8,
			wantEffCh: 7,
			want: []channelMapElementWant{
				{elType: idSCE, tag: 0, nCh: 1, ch0: 2, rel: relBits018},
				{elType: idCPE, tag: 0, nCh: 2, ch0: 0, ch1: 1, rel: relBits026},
				{elType: idCPE, tag: 1, nCh: 2, ch0: 4, ch1: 5, rel: relBits026},
				{elType: idLFE, tag: 0, nCh: 1, ch0: 3, rel: relBits004},
				{elType: idCPE, tag: 2, nCh: 2, ch0: 6, ch1: 7, rel: relBits026},
			},
		},
		{
			name:      "7.1 front-center alias uses chcfg7 map",
			mode:      Mode7_1FrontCenter,
			order:     ChannelOrderWAV,
			wantMode:  Mode7_1FrontCenter,
			wantCh:    8,
			wantEffCh: 7,
			want: []channelMapElementWant{
				{elType: idSCE, tag: 0, nCh: 1, ch0: 2, rel: relBits018},
				{elType: idCPE, tag: 0, nCh: 2, ch0: 6, ch1: 7, rel: relBits026},
				{elType: idCPE, tag: 1, nCh: 2, ch0: 0, ch1: 1, rel: relBits026},
				{elType: idCPE, tag: 2, nCh: 2, ch0: 4, ch1: 5, rel: relBits026},
				{elType: idLFE, tag: 0, nCh: 1, ch0: 3, rel: relBits004},
			},
		},
		{
			name:      "7.1 rear-surround alias uses chcfg12 map",
			mode:      Mode7_1RearSurround,
			order:     ChannelOrderWAV,
			wantMode:  Mode7_1RearSurround,
			wantCh:    8,
			wantEffCh: 7,
			want: []channelMapElementWant{
				{elType: idSCE, tag: 0, nCh: 1, ch0: 2, rel: relBits018},
				{elType: idCPE, tag: 0, nCh: 2, ch0: 0, ch1: 1, rel: relBits026},
				{elType: idCPE, tag: 1, nCh: 2, ch0: 6, ch1: 7, rel: relBits026},
				{elType: idCPE, tag: 2, nCh: 2, ch0: 4, ch1: 5, rel: relBits026},
				{elType: idLFE, tag: 0, nCh: 1, ch0: 3, rel: relBits004},
			},
		},
		{
			name:      "wg4 follows default table for this source path",
			mode:      Mode1_2,
			order:     ChannelOrderWG4,
			wantMode:  Mode1_2,
			wantCh:    3,
			wantEffCh: 3,
			want: []channelMapElementWant{
				{elType: idSCE, tag: 0, nCh: 1, ch0: 2, rel: relBits04},
				{elType: idCPE, tag: 0, nCh: 2, ch0: 0, ch1: 1, rel: relBits06},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cm ChannelMapping
			gotErr := FDKaacEncInitChannelMapping(tt.mode, tt.order, &cm)
			if gotErr != AACEncOK {
				t.Fatalf("error = %#x, want %#x", gotErr, AACEncOK)
			}
			assertChannelMapping(t, &cm, tt.wantMode, tt.wantCh, tt.wantEffCh, tt.want)
		})
	}
}

func TestFDKaacEncInitChannelMappingRejectsInvalid(t *testing.T) {
	t.Run("nil mapping", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, nil)
	})

	t.Run("unsupported mode clears mapping", func(t *testing.T) {
		cm := ChannelMapping{
			EncMode:      Mode2,
			NChannels:    2,
			NChannelsEff: 2,
			NElements:    1,
			ElInfo:       [maxChannelElements]ElementInfo{{ElType: idCPE, NChannelsInEl: 2}},
		}
		gotErr := FDKaacEncInitChannelMapping(ModeUnknown, ChannelOrderMPEG, &cm)
		if gotErr != AACEncUnsupportedChannelConf {
			t.Fatalf("error = %#x, want %#x", gotErr, AACEncUnsupportedChannelConf)
		}
		if cm != (ChannelMapping{}) {
			t.Fatalf("mapping after unsupported mode = %+v, want zero value", cm)
		}
	})

	t.Run("unsupported element type", func(t *testing.T) {
		var el ElementInfo
		var count int
		var tags [idEnd + 1]int
		gotErr := fdkaacEncInitElement(&el, idFIL, &count, ChannelOrderMPEG, int(Mode2), &tags, MaxValDBL)
		if gotErr != AACEncInvalidElementInfoType {
			t.Fatalf("error = %#x, want %#x", gotErr, AACEncInvalidElementInfoType)
		}
	})
}

func TestFDKaacEncInitElementBitsVectors(t *testing.T) {
	const (
		avgBits = 1024
		maxBits = 6144
	)
	tests := []struct {
		name    string
		mode    ChannelMode
		bitrate int
		want    []elementBitsWant
	}{
		{
			name:    "mono",
			mode:    Mode1,
			bitrate: 64000,
			want: []elementBitsWant{
				{chBitrate: 64000, maxBits: 6144, rel: MaxValDBL},
			},
		},
		{
			name:    "stereo",
			mode:    Mode2,
			bitrate: 128000,
			want: []elementBitsWant{
				{chBitrate: 64000, maxBits: 12288, rel: MaxValDBL},
			},
		},
		{
			name:    "3.0",
			mode:    Mode1_2,
			bitrate: 192000,
			want: []elementBitsWant{
				{chBitrate: 76800, maxBits: 6144, rel: relBits04},
				{chBitrate: 57600, maxBits: 12288, rel: relBits06},
			},
		},
		{
			name:    "4.0",
			mode:    Mode1_2_1,
			bitrate: 256000,
			want: []elementBitsWant{
				{chBitrate: 76800, maxBits: 6144, rel: relBits03},
				{chBitrate: 51200, maxBits: 12288, rel: relBits04},
				{chBitrate: 76800, maxBits: 6144, rel: relBits03},
			},
		},
		{
			name:    "5.0",
			mode:    Mode1_2_2,
			bitrate: 320000,
			want: []elementBitsWant{
				{chBitrate: 83199, maxBits: 6144, rel: relBits026},
				{chBitrate: 59200, maxBits: 12288, rel: relBits037},
				{chBitrate: 59200, maxBits: 12288, rel: relBits037},
			},
		},
		{
			name:    "5.1 LFE reservoir exclusion",
			mode:    Mode1_2_2_1,
			bitrate: 320000,
			want: []elementBitsWant{
				{chBitrate: 76799, maxBits: 5996, rel: relBits024},
				{chBitrate: 55999, maxBits: 11992, rel: relBits035},
				{chBitrate: 55999, maxBits: 11992, rel: relBits035},
				{chBitrate: 19199, maxBits: 736, rel: relBits006},
			},
		},
		{
			name:    "6.1 LFE reservoir exclusion",
			mode:    Mode6_1,
			bitrate: 384000,
			want: []elementBitsWant{
				{chBitrate: 76800, maxBits: 6041, rel: relBits02},
				{chBitrate: 52800, maxBits: 12082, rel: relBits027},
				{chBitrate: 52800, maxBits: 12082, rel: relBits027},
				{chBitrate: 38400, maxBits: 6041, rel: relBits02},
				{chBitrate: 19200, maxBits: 614, rel: relBits005},
			},
		},
		{
			name:    "7.1 back",
			mode:    Mode7_1Back,
			bitrate: 448000,
			want: []elementBitsWant{
				{chBitrate: 80640, maxBits: 6074, rel: relBits018},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
				{chBitrate: 17919, maxBits: 490, rel: relBits004},
			},
		},
		{
			name:    "7.1 top-front LFE index",
			mode:    Mode7_1TopFront,
			bitrate: 448000,
			want: []elementBitsWant{
				{chBitrate: 80640, maxBits: 6074, rel: relBits018},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
				{chBitrate: 17919, maxBits: 490, rel: relBits004},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
			},
		},
		{
			name:    "7.1 front-center alias",
			mode:    Mode7_1FrontCenter,
			bitrate: 448000,
			want: []elementBitsWant{
				{chBitrate: 80640, maxBits: 6074, rel: relBits018},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
				{chBitrate: 58239, maxBits: 12148, rel: relBits026},
				{chBitrate: 17919, maxBits: 490, rel: relBits004},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cm ChannelMapping
			if errCode := FDKaacEncInitChannelMapping(tt.mode, ChannelOrderWAV, &cm); errCode != AACEncOK {
				t.Fatalf("mapping error = %#x", errCode)
			}
			elementBits, elementBitsPtrs := initializedElementBits()
			if errCode := FDKaacEncInitElementBits(elementBitsPtrs[:], &cm, tt.bitrate, avgBits, maxBits); errCode != AACEncOK {
				t.Fatalf("element bits error = %#x, want %#x", errCode, AACEncOK)
			}
			assertElementBits(t, elementBits[:], tt.want)
		})
	}
}

func TestFDKaacEncInitElementBitsRejectsInvalid(t *testing.T) {
	t.Run("nil mapping", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		elementBits, elementBitsPtrs := initializedElementBits()
		_ = elementBits
		_ = FDKaacEncInitElementBits(elementBitsPtrs[:], nil, 64000, 1024, 6144)
	})

	t.Run("negative controls", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		var cm ChannelMapping
		if errCode := FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
			t.Fatalf("mapping error = %#x", errCode)
		}
		elementBits, elementBitsPtrs := initializedElementBits()
		_ = elementBits
		_ = FDKaacEncInitElementBits(elementBitsPtrs[:], &cm, -1, 1024, 6144)
	})

	t.Run("too few element entries", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		var cm ChannelMapping
		if errCode := FDKaacEncInitChannelMapping(Mode1_2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
			t.Fatalf("mapping error = %#x", errCode)
		}
		var bit0 ElementBits
		_ = FDKaacEncInitElementBits([]*ElementBits{&bit0}, &cm, 192000, 1024, 6144)
	})

	t.Run("nil element entry", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		var cm ChannelMapping
		if errCode := FDKaacEncInitChannelMapping(Mode2, ChannelOrderMPEG, &cm); errCode != AACEncOK {
			t.Fatalf("mapping error = %#x", errCode)
		}
		_ = FDKaacEncInitElementBits([]*ElementBits{nil}, &cm, 128000, 1024, 6144)
	})

	t.Run("unsupported mode", func(t *testing.T) {
		var cm ChannelMapping
		elementBits, elementBitsPtrs := initializedElementBits()
		errCode := FDKaacEncInitElementBits(elementBitsPtrs[:], &cm, 128000, 1024, 6144)
		if errCode != AACEncUnsupportedChannelConf {
			t.Fatalf("error = %#x, want %#x", errCode, AACEncUnsupportedChannelConf)
		}
		if elementBits[0].ChBitrateEl != 0 || elementBits[0].MaxBitsEl != 0 || elementBits[0].RelativeBitsEl != 0 {
			t.Fatalf("unsupported mode mutated element bits: %+v", elementBits[0])
		}
	})
}

func TestFDKaacEncChannelMappingAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		mode := ModeUnknown
		err0 := FDKaacEncDetermineEncoderMode(&mode, 8)
		var cm51, cm71 ChannelMapping
		err1 := FDKaacEncInitChannelMapping(Mode1_2_2_1, ChannelOrderWAV, &cm51)
		err2 := FDKaacEncInitChannelMapping(Mode7_1TopFront, ChannelOrderWAV, &cm71)
		channelMapModeSink = mode
		channelMapIntSink = err0 + err1 + err2 + cm51.NChannelsEff + cm71.NElements
		channelMapRelSink = cm71.ElInfo[4].RelativeBits
	})
	if allocs != 0 {
		t.Fatalf("channel mapping allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncInitElementBitsAllocs(t *testing.T) {
	var cm ChannelMapping
	if errCode := FDKaacEncInitChannelMapping(Mode7_1TopFront, ChannelOrderWAV, &cm); errCode != AACEncOK {
		t.Fatalf("mapping error = %#x", errCode)
	}
	elementBits, elementBitsPtrs := initializedElementBits()
	allocs := testing.AllocsPerRun(1000, func() {
		errCode := FDKaacEncInitElementBits(elementBitsPtrs[:], &cm, 448000, 1024, 6144)
		channelMapIntSink = errCode + elementBits[3].MaxBitsEl
		channelMapRelSink = elementBits[4].RelativeBitsEl
	})
	if allocs != 0 {
		t.Fatalf("element-bits allocations = %v, want 0", allocs)
	}
}

type channelMapElementWant struct {
	elType int
	tag    int
	nCh    int
	ch0    int
	ch1    int
	rel    FixpDBL
}

type elementBitsWant struct {
	chBitrate int
	maxBits   int
	rel       FixpDBL
}

func assertChannelMapping(t *testing.T, cm *ChannelMapping, wantMode ChannelMode, wantCh int, wantEffCh int, want []channelMapElementWant) {
	t.Helper()
	if cm.EncMode != wantMode || cm.NChannels != wantCh || cm.NChannelsEff != wantEffCh || cm.NElements != len(want) {
		t.Fatalf("mapping header = mode %v ch %d eff %d elems %d, want mode %v ch %d eff %d elems %d",
			cm.EncMode, cm.NChannels, cm.NChannelsEff, cm.NElements,
			wantMode, wantCh, wantEffCh, len(want))
	}
	for i, w := range want {
		got := cm.ElInfo[i]
		if got.ElType != w.elType ||
			got.InstanceTag != w.tag ||
			got.NChannelsInEl != w.nCh ||
			got.ChannelIndex[0] != w.ch0 ||
			got.ChannelIndex[1] != w.ch1 ||
			got.RelativeBits != w.rel {
			t.Fatalf("element %d = %+v, want type %d tag %d nCh %d ch [%d %d] rel %#x",
				i, got, w.elType, w.tag, w.nCh, w.ch0, w.ch1, int32(w.rel))
		}
	}
}

func initializedElementBits() (*[maxChannelElements]ElementBits, [maxChannelElements]*ElementBits) {
	var elementBits [maxChannelElements]ElementBits
	var elementBitsPtrs [maxChannelElements]*ElementBits
	for i := range elementBits {
		elementBits[i].BitResLevelEl = 700 + i
		elementBits[i].MaxBitResBitsEl = 900 + i
		elementBitsPtrs[i] = &elementBits[i]
	}
	return &elementBits, elementBitsPtrs
}

func assertElementBits(t *testing.T, got []ElementBits, want []elementBitsWant) {
	t.Helper()
	for i, w := range want {
		if got[i].ChBitrateEl != w.chBitrate ||
			got[i].MaxBitsEl != w.maxBits ||
			got[i].RelativeBitsEl != w.rel {
			t.Fatalf("element bits %d = %+v, want chBitrate %d maxBits %d rel %#x",
				i, got[i], w.chBitrate, w.maxBits, int32(w.rel))
		}
		if got[i].BitResLevelEl != 700+i || got[i].MaxBitResBitsEl != 900+i {
			t.Fatalf("element bits %d reservoir fields = %d/%d, want preserved %d/%d",
				i, got[i].BitResLevelEl, got[i].MaxBitResBitsEl, 700+i, 900+i)
		}
	}
}
