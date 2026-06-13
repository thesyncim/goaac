package fdkaac

import "testing"

var psyPostHashSink uint64

func TestFDKaacEncPsyPostTnsToolsAndOutputLongStereoVector(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 0, 1)
	var leftData, rightData PsyData
	var leftSpec, rightSpec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&leftData, leftSpec[:], &longConf, 6101)
	preparePsyTnsLongData(&rightData, rightSpec[:], &longConf, 7103)

	var leftStatic, rightStatic PsyStatic
	initPsyPostStatic(&leftStatic, LongWindow, []int{1}, &longConf)
	initPsyPostStatic(&rightStatic, LongWindow, []int{1}, &longConf)

	var leftTNS, rightTNS TNSData
	var leftOut, rightOut PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var tnsScratch PsyTnsTonalityScratch

	rc := FDKaacEncPsyAdvanceTnsAndTonality(
		2,
		[]*PsyStatic{&leftStatic, &rightStatic},
		[]*PsyData{&leftData, &rightData},
		[]*TNSData{&leftTNS, &rightTNS},
		[]*PsyConfiguration{&longConf, &longConf},
		[]*PsyOutChannel{&leftOut, &rightOut},
		&tonality,
		&tnsScratch,
	)
	if rc != AACEncOK {
		t.Fatalf("TNS rc = %#x, want OK", rc)
	}

	var leftPNS, rightPNS PNSData
	element := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&leftOut, &rightOut}}
	var postScratch PsyPostTnsScratch
	rc = FDKaacEncPsyPostTnsToolsAndOutput(
		2,
		[]*PsyStatic{&leftStatic, &rightStatic},
		[]*PsyData{&leftData, &rightData},
		[]*TNSData{&leftTNS, &rightTNS},
		[]*PNSData{&leftPNS, &rightPNS},
		&longConf,
		&shortConf,
		&element,
		&tonality,
		&postScratch,
	)
	if rc != AACEncOK {
		t.Fatalf("post-TNS rc = %#x, want OK", rc)
	}

	if element.CommonWindow != 1 {
		t.Fatalf("common window = %d, want 1", element.CommonWindow)
	}
	if leftOut.SfbCnt != longConf.SfbActive || rightOut.SfbPerGroup != longConf.SfbActive {
		t.Fatalf("long output sfb fields = %d/%d, want %d", leftOut.SfbCnt, rightOut.SfbPerGroup, longConf.SfbActive)
	}
	if len(leftOut.MdctSpectrum) != longConf.GranuleLength || len(rightOut.MdctSpectrum) != longConf.GranuleLength {
		t.Fatalf("output mdct lengths = %d/%d, want %d", len(leftOut.MdctSpectrum), len(rightOut.MdctSpectrum), longConf.GranuleLength)
	}
	if got, want := hashPsyPostStage(2, []*PsyData{&leftData, &rightData}, []*TNSData{&leftTNS, &rightTNS}, []*PNSData{&leftPNS, &rightPNS}, &element, &longConf), uint64(0xf42934541271e50f); got != want {
		t.Fatalf("long stereo post-TNS hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPsyPostTnsToolsAndOutputShortVector(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 0, 1)
	var data PsyData
	var spec [maxSpectralLines]FixpDBL
	preparePsyTnsShortData(&data, spec[:], &shortConf, 8111)

	var static PsyStatic
	initPsyPostStatic(&static, ShortWindow, []int{3, 5}, &longConf)

	var tns TNSData
	var out PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var tnsScratch PsyTnsTonalityScratch
	rc := FDKaacEncPsyAdvanceTnsAndTonality(
		1,
		[]*PsyStatic{&static},
		[]*PsyData{&data},
		[]*TNSData{&tns},
		[]*PsyConfiguration{&shortConf},
		[]*PsyOutChannel{&out},
		&tonality,
		&tnsScratch,
	)
	if rc != AACEncOK {
		t.Fatalf("short TNS rc = %#x, want OK", rc)
	}

	var pns PNSData
	element := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&out}}
	var postScratch PsyPostTnsScratch
	rc = FDKaacEncPsyPostTnsToolsAndOutput(
		1,
		[]*PsyStatic{&static},
		[]*PsyData{&data},
		[]*TNSData{&tns},
		[]*PNSData{&pns},
		&longConf,
		&shortConf,
		&element,
		&tonality,
		&postScratch,
	)
	if rc != AACEncOK {
		t.Fatalf("short post-TNS rc = %#x, want OK", rc)
	}

	if out.SfbCnt != static.BlockSwitchingControl.NoOfGroups*shortConf.SfbCnt || out.SfbPerGroup != shortConf.SfbCnt {
		t.Fatalf("short output sfb fields = %d/%d", out.SfbCnt, out.SfbPerGroup)
	}
	if out.LastWindowSequence != ShortWindow || out.WindowShape != WindowShapeSine || len(out.MdctSpectrum) != shortConf.GranuleLength {
		t.Fatalf("short output state = seq:%d shape:%d len:%d", out.LastWindowSequence, out.WindowShape, len(out.MdctSpectrum))
	}
	if got, want := hashPsyPostStage(1, []*PsyData{&data}, []*TNSData{&tns}, []*PNSData{&pns}, &element, &shortConf), uint64(0xb49e3977aa19fdde); got != want {
		t.Fatalf("short post-TNS hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPsyPostTnsRejectsIntensityStereoUntilPorted(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 1, 1)
	var leftData, rightData PsyData
	var leftSpec, rightSpec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&leftData, leftSpec[:], &longConf, 9103)
	preparePsyTnsLongData(&rightData, rightSpec[:], &longConf, 9109)

	var leftStatic, rightStatic PsyStatic
	initPsyPostStatic(&leftStatic, LongWindow, []int{1}, &longConf)
	initPsyPostStatic(&rightStatic, LongWindow, []int{1}, &longConf)

	var leftTNS, rightTNS TNSData
	var leftOut, rightOut PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var tnsScratch PsyTnsTonalityScratch
	FDKaacEncPsyAdvanceTnsAndTonality(
		2,
		[]*PsyStatic{&leftStatic, &rightStatic},
		[]*PsyData{&leftData, &rightData},
		[]*TNSData{&leftTNS, &rightTNS},
		[]*PsyConfiguration{&longConf, &longConf},
		[]*PsyOutChannel{&leftOut, &rightOut},
		&tonality,
		&tnsScratch,
	)

	var leftPNS, rightPNS PNSData
	element := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&leftOut, &rightOut}}
	var postScratch PsyPostTnsScratch
	expectAACEncPanic(t, func() {
		FDKaacEncPsyPostTnsToolsAndOutput(
			2,
			[]*PsyStatic{&leftStatic, &rightStatic},
			[]*PsyData{&leftData, &rightData},
			[]*TNSData{&leftTNS, &rightTNS},
			[]*PNSData{&leftPNS, &rightPNS},
			&longConf,
			&shortConf,
			&element,
			&tonality,
			&postScratch,
		)
	})
}

func TestFDKaacEncPsyPostTnsRejectsInvalidControls(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 0, 1)
	var data PsyData
	var spec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&data, spec[:], &longConf, 10103)
	var static PsyStatic
	initPsyPostStatic(&static, LongWindow, []int{1}, &longConf)
	var tns TNSData
	var pns PNSData
	var out PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var scratch PsyPostTnsScratch
	element := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&out}}

	tests := []struct {
		name string
		run  func()
	}{
		{"zero channels", func() {
			FDKaacEncPsyPostTnsToolsAndOutput(0, nil, nil, nil, nil, &longConf, &shortConf, &element, &tonality, &scratch)
		}},
		{"nil scratch", func() {
			FDKaacEncPsyPostTnsToolsAndOutput(1, []*PsyStatic{&static}, []*PsyData{&data}, []*TNSData{&tns}, []*PNSData{&pns}, &longConf, &shortConf, &element, &tonality, nil)
		}},
		{"nil output channel", func() {
			badElement := PsyOutElement{}
			FDKaacEncPsyPostTnsToolsAndOutput(1, []*PsyStatic{&static}, []*PsyData{&data}, []*TNSData{&tns}, []*PNSData{&pns}, &longConf, &shortConf, &badElement, &tonality, &scratch)
		}},
		{"stereo window mismatch", func() {
			var rightData PsyData
			var rightSpec [maxSpectralLines]FixpDBL
			preparePsyTnsLongData(&rightData, rightSpec[:], &longConf, 10111)
			var rightStatic PsyStatic
			initPsyPostStatic(&rightStatic, StartWindow, []int{1}, &longConf)
			var rightTNS TNSData
			var rightPNS PNSData
			var rightOut PsyOutChannel
			badElement := PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&out, &rightOut}}
			FDKaacEncPsyPostTnsToolsAndOutput(2, []*PsyStatic{&static, &rightStatic}, []*PsyData{&data, &rightData}, []*TNSData{&tns, &rightTNS}, []*PNSData{&pns, &rightPNS}, &longConf, &shortConf, &badElement, &tonality, &scratch)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectAACEncPanic(t, tt.run)
		})
	}
}

func TestFDKaacEncPsyPostTnsAllocs(t *testing.T) {
	longConf, shortConf := mustPsyPostConfig(t, 0, 1)
	var seedData PsyData
	var seedSpec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&seedData, seedSpec[:], &longConf, 11113)
	var seedStatic PsyStatic
	initPsyPostStatic(&seedStatic, LongWindow, []int{1}, &longConf)
	var seedTNS TNSData
	var seedOut PsyOutChannel
	var seedTonality [2][maxSFBLong]FixpSGL
	var seedTnsScratch PsyTnsTonalityScratch
	FDKaacEncPsyAdvanceTnsAndTonality(
		1,
		[]*PsyStatic{&seedStatic},
		[]*PsyData{&seedData},
		[]*TNSData{&seedTNS},
		[]*PsyConfiguration{&longConf},
		[]*PsyOutChannel{&seedOut},
		&seedTonality,
		&seedTnsScratch,
	)

	staticPtrs := [1]*PsyStatic{}
	dataPtrs := [1]*PsyData{}
	tnsPtrs := [1]*TNSData{}
	pnsPtrs := [1]*PNSData{}
	var postScratch PsyPostTnsScratch
	var data PsyData
	var spec [maxSpectralLines]FixpDBL
	var static PsyStatic
	var tns TNSData
	var pns PNSData
	var out PsyOutChannel
	var element PsyOutElement
	var tonality [2][maxSFBLong]FixpSGL
	staticPtrs[0] = &static
	dataPtrs[0] = &data
	tnsPtrs[0] = &tns
	pnsPtrs[0] = &pns
	allocs := testing.AllocsPerRun(1000, func() {
		data = seedData
		spec = seedSpec
		data.MdctSpectrum = spec[:]
		static = seedStatic
		tns = seedTNS
		pns = PNSData{}
		out = PsyOutChannel{}
		element = PsyOutElement{PsyOutChannel: [2]*PsyOutChannel{&out}}
		tonality = seedTonality

		rc := FDKaacEncPsyPostTnsToolsAndOutput(1, staticPtrs[:], dataPtrs[:], tnsPtrs[:], pnsPtrs[:], &longConf, &shortConf, &element, &tonality, &postScratch)
		psyPostHashSink ^= uint64(rc)
		psyPostHashSink ^= hashPsyPostStage(1, dataPtrs[:], tnsPtrs[:], pnsPtrs[:], &element, &longConf)
	})
	if allocs != 0 {
		t.Fatalf("post-TNS allocations = %v, want 0", allocs)
	}
}

func mustPsyPostConfig(t *testing.T, useIS int, useMS int) (PsyConfiguration, PsyConfiguration) {
	t.Helper()
	var longConf, shortConf PsyConfiguration
	if got := FDKaacEncInitPsyConfiguration(48000, 48000, 15500, LongWindow, 1024, useIS, useMS, &longConf, FilterbankLC); got != AACEncOK {
		t.Fatalf("long psy config rc = %#x, want OK", got)
	}
	if got := FDKaacEncInitTnsConfiguration(96000, 48000, 2, LongWindow, 1024, 0, 0, &longConf.TnsConf, &longConf, 1, 1); got != AACEncOK {
		t.Fatalf("long TNS config rc = %#x, want OK", got)
	}
	if got := FDKaacEncInitPnsConfiguration(&longConf.PnsConf, 48000, 48000, 1, longConf.SfbCnt, longConf.SfbOffset[:], 2, 1); got != AACEncOK {
		t.Fatalf("long PNS config rc = %#x, want OK", got)
	}
	if got := FDKaacEncInitPsyConfiguration(48000, 48000, 15500, ShortWindow, 1024, useIS, useMS, &shortConf, FilterbankLC); got != AACEncOK {
		t.Fatalf("short psy config rc = %#x, want OK", got)
	}
	if got := FDKaacEncInitTnsConfiguration(96000, 48000, 2, ShortWindow, 1024, 0, 0, &shortConf.TnsConf, &shortConf, 1, 1); got != AACEncOK {
		t.Fatalf("short TNS config rc = %#x, want OK", got)
	}
	if got := FDKaacEncInitPnsConfiguration(&shortConf.PnsConf, 48000, 48000, 1, shortConf.SfbCnt, shortConf.SfbOffset[:], 2, 1); got != AACEncOK {
		t.Fatalf("short PNS config rc = %#x, want OK", got)
	}
	return longConf, shortConf
}

func initPsyPostStatic(static *PsyStatic, sequence int, groupLen []int, longConf *PsyConfiguration) {
	*static = PsyStatic{}
	static.BlockSwitchingControl.LastWindowSequence = sequence
	static.BlockSwitchingControl.WindowShape = WindowShapeKBD
	static.BlockSwitchingControl.NoOfGroups = len(groupLen)
	for i, n := range groupLen {
		static.BlockSwitchingControl.GroupLen[i] = n
	}
	FDKaacEncInitPreEchoControl(
		static.SfbThresholdNm1[:],
		&static.CalcPreEcho,
		longConf.SfbCnt,
		longConf.SfbPcmQuantThreshold[:],
		&static.MdctScaleNm1,
	)
}

func hashPsyPostStage(
	channels int,
	psyData []*PsyData,
	tnsData []*TNSData,
	pnsData []*PNSData,
	element *PsyOutElement,
	conf *PsyConfiguration,
) uint64 {
	h := uint64(14695981039346656037)
	h = hashTnsSyncInt(h, element.CommonWindow)
	h = hashTnsSyncInt(h, element.ToolsInfo.MsDigest)
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(element.ToolsInfo.MsMask[:]))
	for ch := 0; ch < channels; ch++ {
		h = mixPsyTnsHash(h, hashPsyPostChannel(psyData[ch], tnsData[ch], pnsData[ch], element.PsyOutChannel[ch], conf))
	}
	return h
}

func hashPsyPostChannel(data *PsyData, tns *TNSData, pns *PNSData, out *PsyOutChannel, conf *PsyConfiguration) uint64 {
	h := uint64(14695981039346656037)
	h = hashTnsSyncInt(h, data.MdctScale)
	h = hashTnsSyncInt(h, data.SfbActive)
	h = mixPsyTnsHash(h, hashFixpDBL(data.MdctSpectrum[:conf.GranuleLength]))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbThreshold.Long[:]))
	h = mixPsyTnsHash(h, hashPsyTnsShortEnergy(&data.SfbThreshold.Short))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbEnergy.Long[:]))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbEnergyLdData.Long[:]))
	h = mixPsyTnsHash(h, hashPsyTnsShortEnergy(&data.SfbEnergy.Short))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbEnergyMS.Long[:]))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbEnergyMSLdData[:]))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbSpreadEnergy.Long[:]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(data.GroupedSfbOffset[:]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(pns.PNSFlag[:]))
	h = mixPsyTnsHash(h, hashFixpSGL(pns.NoiseFuzzyMeasure[:]))
	h = mixPsyTnsHash(h, hashFixpDBL(pns.NoiseEnergyCorrelation[:]))
	h = mixPsyTnsHash(h, hashPsyPostOut(out))
	h = mixPsyTnsHash(h, hashPsyTnsDynamic(tns, out, conf.TnsConf.MaxOrder))
	return h
}

func hashPsyPostOut(out *PsyOutChannel) uint64 {
	h := uint64(14695981039346656037)
	h = hashTnsSyncInt(h, out.SfbCnt)
	h = hashTnsSyncInt(h, out.SfbPerGroup)
	h = hashTnsSyncInt(h, out.MaxSfbPerGroup)
	h = hashTnsSyncInt(h, out.LastWindowSequence)
	h = hashTnsSyncInt(h, out.WindowShape)
	h = hashTnsSyncInt(h, out.GroupingMask)
	h = hashTnsSyncInt(h, out.MdctScale)
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(out.SfbOffsets[:]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(out.GroupLen[:]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(out.NoiseNrg[:]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(out.IsBook[:]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(out.IsScale[:]))
	h = mixPsyTnsHash(h, hashFixpDBL(out.MdctSpectrum))
	h = mixPsyTnsHash(h, hashFixpDBL(out.SfbEnergy[:]))
	h = mixPsyTnsHash(h, hashFixpDBL(out.SfbEnergyLdData[:]))
	h = mixPsyTnsHash(h, hashFixpDBL(out.SfbThresholdLdData[:]))
	h = mixPsyTnsHash(h, hashFixpDBL(out.SfbSpreadEnergy[:]))
	return h
}

func hashPsyPostIntSlice(values []int) uint64 {
	h := uint64(14695981039346656037)
	for _, v := range values {
		h = hashTnsSyncInt(h, v)
	}
	return h
}
