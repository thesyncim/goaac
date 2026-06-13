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

type channelMapElementWant struct {
	elType int
	tag    int
	nCh    int
	ch0    int
	ch1    int
	rel    FixpDBL
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
