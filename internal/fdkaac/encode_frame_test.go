package fdkaac

import "testing"

var encodeFrameHashSink uint64

func TestFDKaacEncEncodeFrameRawStereoVector(t *testing.T) {
	cfg := baseAACLCConfig(48000, 128000, 2, Mode2)
	var state AACEncFrameState
	if errCode := FDKaacEncInitRawFrameState(&state, cfg); errCode != AACEncOK {
		t.Fatalf("raw frame init error = %#x, want OK", errCode)
	}

	var scratch AACEncFrameScratch
	var pcm [2 * maxSpectralLines]int16
	fillPsyMainSmoothPCM(pcm[:], cfg.FrameLength)
	out, result, errCode := FDKaacEncEncodeFrameRaw(&state, nil, pcm[:], cfg.FrameLength, &scratch)
	if errCode != AACEncOK {
		t.Fatalf("raw frame encode error = %#x, want OK", errCode)
	}
	if result.PayloadBytes != len(out) {
		t.Fatalf("payload bytes = %d, want %d", result.PayloadBytes, len(out))
	}
	if result.Write.ChannelElements != 1 || result.Write.GlobalExtensions != 1 {
		t.Fatalf("write result = %+v, want one CPE and fill extension", result.Write)
	}
	if result.QCMain.QuantizationDone != 1 || result.QCMain.QuantizedElements == 0 {
		t.Fatalf("QC result = %+v, want quantized frame", result.QCMain)
	}
	if result.TransportStaticBits != 0 {
		t.Fatalf("raw transport static bits = %d, want 0", result.TransportStaticBits)
	}
	if result.TotalBits != len(out)*8 || result.Write.FrameBits != result.TotalBits {
		t.Fatalf("bits = total %d write %d bytes %d", result.TotalBits, result.Write.FrameBits, len(out))
	}
	if got, want := hashHuffBytes(out), uint64(0x75f983c227eab3f5); got != want {
		t.Fatalf("raw frame hash = %#x, want %#x; len=%d result=%+v", got, want, len(out), result)
	}
}

func TestFDKaacEncEncodeFrameRawRejectsInvalid(t *testing.T) {
	cfg := baseAACLCConfig(48000, 128000, 2, Mode2)
	var state AACEncFrameState
	if errCode := FDKaacEncInitRawFrameState(&state, cfg); errCode != AACEncOK {
		t.Fatalf("raw frame init error = %#x, want OK", errCode)
	}
	var scratch AACEncFrameScratch
	var pcm [2 * maxSpectralLines]int16

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"nil state", func() { FDKaacEncEncodeFrameRaw(nil, nil, pcm[:], cfg.FrameLength, &scratch) }},
		{"nil scratch", func() { FDKaacEncEncodeFrameRaw(&state, nil, pcm[:], cfg.FrameLength, nil) }},
		{"uninitialized", func() {
			var empty AACEncFrameState
			FDKaacEncEncodeFrameRaw(&empty, nil, pcm[:], cfg.FrameLength, &scratch)
		}},
		{"short stride", func() { FDKaacEncEncodeFrameRaw(&state, nil, pcm[:], cfg.FrameLength-1, &scratch) }},
		{"short input", func() { FDKaacEncEncodeFrameRaw(&state, nil, pcm[:cfg.FrameLength], cfg.FrameLength, &scratch) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

func TestFDKaacEncEncodeFrameRawAllocs(t *testing.T) {
	cfg := baseAACLCConfig(48000, 128000, 2, Mode2)
	var state AACEncFrameState
	if errCode := FDKaacEncInitRawFrameState(&state, cfg); errCode != AACEncOK {
		t.Fatalf("raw frame init error = %#x, want OK", errCode)
	}
	var pcm [2 * maxSpectralLines]int16
	fillPsyMainSmoothPCM(pcm[:], cfg.FrameLength)
	var out [2048]byte
	var scratch AACEncFrameScratch

	allocs := testing.AllocsPerRun(100, func() {
		scratch = AACEncFrameScratch{}
		got, result, errCode := FDKaacEncEncodeFrameRaw(&state, out[:0], pcm[:], cfg.FrameLength, &scratch)
		if errCode != AACEncOK {
			t.Fatalf("raw frame encode error = %#x, want OK", errCode)
		}
		encodeFrameHashSink = hashHuffBytes(got) ^ uint64(result.TotalBits)
	})
	if allocs != 0 {
		t.Fatalf("allocs = %.2f, want 0", allocs)
	}
}
