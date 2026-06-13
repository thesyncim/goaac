package fdkaac

import "testing"

var bandwidthSink int
var bandwidthModeSink QCBitrateMode

func TestFDKaacEncDetermineBandWidthCBRVectors(t *testing.T) {
	tests := []struct {
		name       string
		cm         ChannelMapping
		mode       ChannelMode
		bitrate    int
		sampleRate int
		frameLen   int
		proposed   int
		wantBW     int
		wantErr    int
	}{
		{
			name:       "mono stepped table",
			cm:         bandwidthMonoMapping(),
			mode:       Mode1,
			bitrate:    64000,
			sampleRate: 48000,
			frameLen:   1024,
			wantBW:     13950,
			wantErr:    AACEncOK,
		},
		{
			name:       "stereo stepped table",
			cm:         bandwidthStereoMapping(),
			mode:       Mode2,
			bitrate:    96000,
			sampleRate: 48000,
			frameLen:   1024,
			wantBW:     14260,
			wantErr:    AACEncOK,
		},
		{
			name:       "LFE excluded from effective channels",
			cm:         bandwidthFiveOneMapping(),
			mode:       Mode1_2_2_1,
			bitrate:    320000,
			sampleRate: 48000,
			frameLen:   1024,
			wantBW:     15500,
			wantErr:    AACEncOK,
		},
		{
			name:       "proposed CBR clips to Nyquist",
			cm:         bandwidthStereoMapping(),
			mode:       Mode2,
			bitrate:    128000,
			sampleRate: 32000,
			frameLen:   1024,
			proposed:   30000,
			wantBW:     16000,
			wantErr:    AACEncOK,
		},
		{
			name:       "invalid channel bitrate",
			cm:         bandwidthStereoMapping(),
			mode:       Mode2,
			bitrate:    1200000,
			sampleRate: 48000,
			frameLen:   1024,
			wantBW:     -1,
			wantErr:    AACEncInvalidChannelBitrate,
		},
		{
			name:       "low-delay mono interpolates with FDK fixed point",
			cm:         bandwidthMonoMapping(),
			mode:       Mode1,
			bitrate:    26000,
			sampleRate: 22050,
			frameLen:   128,
			wantBW:     6987,
			wantErr:    AACEncOK,
		},
		{
			name:       "low-delay stereo interpolates with FDK fixed point",
			cm:         bandwidthStereoMapping(),
			mode:       Mode2,
			bitrate:    52000,
			sampleRate: 22050,
			frameLen:   128,
			wantBW:     8025,
			wantErr:    AACEncOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBW, gotErr := FDKaacEncDetermineBandWidth(tt.proposed, tt.bitrate, QCBitrateModeCBR, tt.sampleRate, tt.frameLen, &tt.cm, tt.mode)
			if gotBW != tt.wantBW || gotErr != tt.wantErr {
				t.Fatalf("bandwidth/error = %d/%#x, want %d/%#x", gotBW, gotErr, tt.wantBW, tt.wantErr)
			}
		})
	}
}

func TestFDKaacEncDetermineBandWidthVBRVectors(t *testing.T) {
	cm := bandwidthStereoMapping()
	tests := []struct {
		name       string
		mode       QCBitrateMode
		sampleRate int
		proposed   int
		wantBW     int
		wantErr    int
	}{
		{name: "VBR5 table", mode: QCBitrateModeVBR5, sampleRate: 48000, wantBW: 19293, wantErr: AACEncOK},
		{name: "VBR5 clips to Nyquist", mode: QCBitrateModeVBR5, sampleRate: 32000, wantBW: 16000, wantErr: AACEncOK},
		{name: "VBR proposed bypasses 20 kHz CBR cap", mode: QCBitrateModeVBR4, sampleRate: 48000, proposed: 22000, wantBW: 22000, wantErr: AACEncOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBW, gotErr := FDKaacEncDetermineBandWidth(tt.proposed, 96000, tt.mode, tt.sampleRate, 1024, &cm, Mode2)
			if gotBW != tt.wantBW || gotErr != tt.wantErr {
				t.Fatalf("VBR bandwidth/error = %d/%#x, want %d/%#x", gotBW, gotErr, tt.wantBW, tt.wantErr)
			}
		})
	}
}

func TestFDKaacEncVBRBitrateHelpers(t *testing.T) {
	if got := FDKaacEncGetVBRBitrate(QCBitrateModeVBR3, Mode1); got != 56000 {
		t.Fatalf("VBR3 mono bitrate = %d, want 56000", got)
	}
	if got := FDKaacEncGetVBRBitrate(QCBitrateModeVBR3, Mode2); got != 96000 {
		t.Fatalf("VBR3 stereo bitrate = %d, want 96000", got)
	}
	if got := FDKaacEncGetVBRBitrate(QCBitrateModeVBR2, Mode1_2_2_1); got != 160000 {
		t.Fatalf("VBR2 5.1 effective bitrate = %d, want 160000", got)
	}
	if got := FDKaacEncAdjustVBRBitrateMode(QCBitrateModeVBR5, -1, Mode2); got != QCBitrateModeVBR5 {
		t.Fatalf("unchanged VBR mode = %v, want VBR5", got)
	}
	if got := FDKaacEncAdjustVBRBitrateMode(QCBitrateModeVBR5, 100000, Mode2); got != QCBitrateModeVBR3 {
		t.Fatalf("lowered VBR mode = %v, want VBR3", got)
	}
	if got := FDKaacEncAdjustVBRBitrateMode(QCBitrateModeVBR3, 300000, Mode2); got != QCBitrateModeVBR3 {
		t.Fatalf("non-increasing VBR mode = %v, want VBR3", got)
	}
	if got := FDKaacEncAdjustVBRBitrateMode(QCBitrateModeVBR1, 1000, Mode1); got != QCBitrateModeInvalid {
		t.Fatalf("underflow VBR mode = %v, want invalid", got)
	}
	if got := FDKaacEncAdjustVBRBitrateMode(QCBitrateModeCBR, -1, Mode2); got != QCBitrateModeInvalid {
		t.Fatalf("CBR adjust mode = %v, want invalid", got)
	}
}

func TestFDKaacEncBandwidthChannelModeConfig(t *testing.T) {
	cfg, ok := FDKaacEncGetChannelModeConfiguration(Mode1_2_2_1)
	if !ok || cfg.NChannels != 6 || cfg.NChannelsEff != 5 || cfg.NElements != 4 {
		t.Fatalf("5.1 config = %+v/%v, want 6 channels, 5 effective, 4 elements", cfg, ok)
	}
	cfg, ok = FDKaacEncGetChannelModeConfiguration(Mode7_1FrontCenter)
	if !ok || cfg.NChannels != 8 || cfg.NChannelsEff != 7 || cfg.NElements != 5 {
		t.Fatalf("7.1 front config = %+v/%v, want 8 channels, 7 effective, 5 elements", cfg, ok)
	}
	if got := FDKaacEncGetMonoStereoMode(Mode1); got != ElementModeMono {
		t.Fatalf("mono mode = %v, want mono", got)
	}
	if got := FDKaacEncGetMonoStereoMode(Mode7_1RearSurround); got != ElementModeStereo {
		t.Fatalf("7.1 rear mode = %v, want stereo", got)
	}
	if got := FDKaacEncGetMonoStereoMode(ModeInvalid); got != ElementModeInvalid {
		t.Fatalf("invalid mode = %v, want invalid", got)
	}
}

func TestFDKaacEncDetermineBandWidthRejectsInvalid(t *testing.T) {
	t.Run("nil mapping", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		FDKaacEncDetermineBandWidth(0, 96000, QCBitrateModeCBR, 48000, 1024, nil, Mode2)
	})
	t.Run("unsupported channel config", func(t *testing.T) {
		cm := bandwidthStereoMapping()
		gotBW, gotErr := FDKaacEncDetermineBandWidth(0, 96000, QCBitrateModeCBR, 48000, 1024, &cm, ModeUnknown)
		if gotBW != 0 || gotErr != AACEncUnsupportedChannelConf {
			t.Fatalf("unsupported channel config = %d/%#x, want 0/%#x", gotBW, gotErr, AACEncUnsupportedChannelConf)
		}
	})
	t.Run("unsupported bitrate mode", func(t *testing.T) {
		cm := bandwidthStereoMapping()
		gotBW, gotErr := FDKaacEncDetermineBandWidth(0, 96000, QCBitrateModeInvalid, 48000, 1024, &cm, Mode2)
		if gotBW != 0 || gotErr != AACEncUnsupportedBitrateMode {
			t.Fatalf("unsupported bitrate mode = %d/%#x, want 0/%#x", gotBW, gotErr, AACEncUnsupportedBitrateMode)
		}
	})
	t.Run("invalid VBR channel mode", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = FDKaacEncGetVBRBitrate(QCBitrateModeVBR3, ModeUnknown)
	})
	t.Run("invalid VBR adjustment channel mode", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = FDKaacEncAdjustVBRBitrateMode(QCBitrateModeVBR3, 96000, ModeUnknown)
	})
}

func TestFDKaacEncBandwidthAllocs(t *testing.T) {
	cm := bandwidthFiveOneMapping()
	allocs := testing.AllocsPerRun(1000, func() {
		bw, errCode := FDKaacEncDetermineBandWidth(0, 320000, QCBitrateModeCBR, 48000, 1024, &cm, Mode1_2_2_1)
		vbr := FDKaacEncGetVBRBitrate(QCBitrateModeVBR4, Mode1_2_2_1)
		mode := FDKaacEncAdjustVBRBitrateMode(QCBitrateModeVBR5, 100000, Mode2)
		bandwidthSink = bw + errCode + vbr
		bandwidthModeSink = mode
	})
	if allocs != 0 {
		t.Fatalf("bandwidth allocations = %v, want 0", allocs)
	}
}

func bandwidthMonoMapping() ChannelMapping {
	var cm ChannelMapping
	cm.NElements = 1
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	return cm
}

func bandwidthStereoMapping() ChannelMapping {
	var cm ChannelMapping
	cm.NElements = 1
	cm.ElInfo[0] = ElementInfo{ElType: idCPE, NChannelsInEl: 2}
	return cm
}

func bandwidthFiveOneMapping() ChannelMapping {
	var cm ChannelMapping
	cm.NElements = 4
	cm.ElInfo[0] = ElementInfo{ElType: idSCE, NChannelsInEl: 1}
	cm.ElInfo[1] = ElementInfo{ElType: idCPE, NChannelsInEl: 2}
	cm.ElInfo[2] = ElementInfo{ElType: idCPE, NChannelsInEl: 2}
	cm.ElInfo[3] = ElementInfo{ElType: idLFE, NChannelsInEl: 1}
	return cm
}
