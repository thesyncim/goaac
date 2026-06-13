package fdkaac

import "testing"

var psyMainHashSink uint64

func TestFDKaacEncPsyMainLongStereoVector(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 1, 1)
	confs := [2]PsyConfiguration{longConf, shortConf}

	var leftStatic, rightStatic PsyStatic
	initPsyMainStatic(&leftStatic, &longConf)
	initPsyMainStatic(&rightStatic, &longConf)

	var dyn PsyDynamic
	var leftQC, rightQC QCOutChannel
	leftOut := PsyOutChannel{MdctSpectrum: leftQC.MdctSpectrum[:]}
	rightOut := PsyOutChannel{MdctSpectrum: rightQC.MdctSpectrum[:]}
	psyElement := PsyElement{PsyStatic: [2]*PsyStatic{&leftStatic, &rightStatic}}
	outElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&leftOut, &rightOut}}

	var pcm [2 * maxSpectralLines]int16
	fillPsyMainSmoothPCM(pcm[:], longConf.GranuleLength)
	chIdx := [2]int{0, 1}
	var scratch PsyMainScratch

	rc := FDKaacEncPsyMain(2, &psyElement, &dyn, &confs, &outElement, pcm[:], longConf.GranuleLength, chIdx[:], 2, &scratch)
	if rc != AACEncOK {
		t.Fatalf("psy main rc = %#x, want OK", rc)
	}

	if outElement.CommonWindow != 1 {
		t.Fatalf("common window = %d, want 1", outElement.CommonWindow)
	}
	if leftOut.LastWindowSequence != LongWindow || rightOut.LastWindowSequence != LongWindow {
		t.Fatalf("window sequence = %d/%d, want long", leftOut.LastWindowSequence, rightOut.LastWindowSequence)
	}
	if len(leftOut.MdctSpectrum) != longConf.GranuleLength || len(rightOut.MdctSpectrum) != longConf.GranuleLength {
		t.Fatalf("mdct lengths = %d/%d, want %d", len(leftOut.MdctSpectrum), len(rightOut.MdctSpectrum), longConf.GranuleLength)
	}
	if leftStatic.PsyInputBuffer[longConf.GranuleLength] != pcm[448] ||
		leftStatic.PsyInputBuffer[longConf.GranuleLength+575] != pcm[1023] {
		t.Fatalf("input history rotation mismatch")
	}
	if got, want := hashPsyPostStage(2, []*PsyData{&dyn.PsyData[0], &dyn.PsyData[1]}, []*TNSData{&dyn.TNSData[0], &dyn.TNSData[1]}, []*PNSData{&dyn.PNSData[0], &dyn.PNSData[1]}, &outElement, &longConf), uint64(0x8912655256a7e261); got != want {
		t.Fatalf("long stereo psy main hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPsyMainBlockSwitchTransitionVector(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 0, 1)
	confs := [2]PsyConfiguration{longConf, shortConf}

	var static PsyStatic
	initPsyMainStatic(&static, &longConf)

	var dyn PsyDynamic
	var qc QCOutChannel
	out := PsyOutChannel{MdctSpectrum: qc.MdctSpectrum[:]}
	psyElement := PsyElement{PsyStatic: [2]*PsyStatic{&static}}
	outElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&out}}

	var pcm [maxSpectralLines]int16
	fillPsyMainAttackPCM(pcm[:])
	chIdx := [1]int{0}
	var scratch PsyMainScratch

	rc := FDKaacEncPsyMain(1, &psyElement, &dyn, &confs, &outElement, pcm[:], longConf.GranuleLength, chIdx[:], 1, &scratch)
	if rc != AACEncOK {
		t.Fatalf("transition psy main rc = %#x, want OK", rc)
	}
	if out.LastWindowSequence != StartWindow {
		t.Fatalf("transition window sequence = %d, want start", out.LastWindowSequence)
	}
	if static.BlockSwitchingControl.Attack == 0 {
		t.Fatalf("block-switch attack flag was not raised")
	}
	if got, want := hashPsyPostStage(1, []*PsyData{&dyn.PsyData[0]}, []*TNSData{&dyn.TNSData[0]}, []*PNSData{&dyn.PNSData[0]}, &outElement, &longConf), uint64(0xc4f627cd11dd428a); got != want {
		t.Fatalf("transition psy main hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPsyMainRejectsInvalidControls(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 0, 1)
	confs := [2]PsyConfiguration{longConf, shortConf}
	var static PsyStatic
	initPsyMainStatic(&static, &longConf)
	var dyn PsyDynamic
	var qc QCOutChannel
	out := PsyOutChannel{MdctSpectrum: qc.MdctSpectrum[:]}
	psyElement := PsyElement{PsyStatic: [2]*PsyStatic{&static}}
	outElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&out}}
	var pcm [maxSpectralLines]int16
	fillPsyMainSmoothPCM(pcm[:], longConf.GranuleLength)
	chIdx := [1]int{0}
	var scratch PsyMainScratch

	badFilter := confs
	badFilter[0].Filterbank = FilterbankLD
	if got := FDKaacEncPsyMain(1, &psyElement, &dyn, &badFilter, &outElement, pcm[:], longConf.GranuleLength, chIdx[:], 1, &scratch); got != AACEncUnsupportedFilterbank {
		t.Fatalf("unsupported filterbank rc = %#x, want %#x", got, AACEncUnsupportedFilterbank)
	}

	tests := []struct {
		name string
		run  func()
	}{
		{"nil scratch", func() {
			FDKaacEncPsyMain(1, &psyElement, &dyn, &confs, &outElement, pcm[:], longConf.GranuleLength, chIdx[:], 1, nil)
		}},
		{"short input", func() {
			FDKaacEncPsyMain(1, &psyElement, &dyn, &confs, &outElement, pcm[:longConf.GranuleLength-1], longConf.GranuleLength, chIdx[:], 1, &scratch)
		}},
		{"nil spectrum", func() {
			badOut := PsyOutChannel{}
			badElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&badOut}}
			FDKaacEncPsyMain(1, &psyElement, &dyn, &confs, &badElement, pcm[:], longConf.GranuleLength, chIdx[:], 1, &scratch)
		}},
		{"mismatched short config", func() {
			badConf := confs
			badConf[1].GranuleLength--
			FDKaacEncPsyMain(1, &psyElement, &dyn, &badConf, &outElement, pcm[:], longConf.GranuleLength, chIdx[:], 1, &scratch)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectAACEncPanic(t, tt.run)
		})
	}
}

func TestFDKaacEncPsyMainAllocs(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 0, 1)
	confs := [2]PsyConfiguration{longConf, shortConf}
	var seedStatic PsyStatic
	initPsyMainStatic(&seedStatic, &longConf)

	var pcm [maxSpectralLines]int16
	fillPsyMainSmoothPCM(pcm[:], longConf.GranuleLength)
	chIdx := [1]int{0}

	var static PsyStatic
	var dyn PsyDynamic
	var qc QCOutChannel
	var out PsyOutChannel
	var psyElement PsyElement
	var outElement PsyOutElement
	var scratch PsyMainScratch

	allocs := testing.AllocsPerRun(1000, func() {
		static = seedStatic
		dyn = PsyDynamic{}
		qc = QCOutChannel{}
		out = PsyOutChannel{MdctSpectrum: qc.MdctSpectrum[:]}
		psyElement = PsyElement{PsyStatic: [2]*PsyStatic{&static}}
		outElement = PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&out}}
		scratch = PsyMainScratch{}

		rc := FDKaacEncPsyMain(1, &psyElement, &dyn, &confs, &outElement, pcm[:], longConf.GranuleLength, chIdx[:], 1, &scratch)
		psyMainHashSink ^= uint64(rc)
		psyMainHashSink ^= hashPsyPostStage(1, []*PsyData{&dyn.PsyData[0]}, []*TNSData{&dyn.TNSData[0]}, []*PNSData{&dyn.PNSData[0]}, &outElement, &longConf)
	})
	if allocs != 0 {
		t.Fatalf("psy main allocations = %v, want 0", allocs)
	}
}

func initPsyMainStatic(static *PsyStatic, longConf *PsyConfiguration) {
	*static = PsyStatic{}
	FDKaacEncInitBlockSwitching(&static.BlockSwitchingControl, false)
	MDCTInit(&static.MDCT, nil)
	FDKaacEncInitPreEchoControl(
		static.SfbThresholdNm1[:],
		&static.CalcPreEcho,
		longConf.SfbCnt,
		longConf.SfbPcmQuantThreshold[:],
		&static.MdctScaleNm1,
	)
}

func fillPsyMainSmoothPCM(dst []int16, stride int) {
	for ch := 0; ch < len(dst)/stride; ch++ {
		base := ch * stride
		for i := 0; i < stride; i++ {
			v := ((i + ch*7) % 64) - 32
			dst[base+i] = int16(v / 8)
		}
	}
}

func fillPsyMainAttackPCM(dst []int16) {
	clear(dst)
	for i := 736; i < 832 && i < len(dst); i++ {
		if i&1 == 0 {
			dst[i] = 28000
		} else {
			dst[i] = -28000
		}
	}
}
