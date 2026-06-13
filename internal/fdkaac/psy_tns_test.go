package fdkaac

import "testing"

var psyTnsSink uint64

func TestFDKaacEncPsyAdvanceTnsAndTonalityLongStereoVector(t *testing.T) {
	conf := mustPsyTnsConfig(t, LongWindow)
	var leftData, rightData PsyData
	var leftSpec, rightSpec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&leftData, leftSpec[:], &conf, 2333)
	preparePsyTnsLongData(&rightData, rightSpec[:], &conf, 3779)

	var leftStatic, rightStatic PsyStatic
	leftStatic.BlockSwitchingControl.LastWindowSequence = LongWindow
	rightStatic.BlockSwitchingControl.LastWindowSequence = LongWindow
	leftStatic.BlockSwitchingControl.NoOfGroups = 1
	rightStatic.BlockSwitchingControl.NoOfGroups = 1
	leftStatic.BlockSwitchingControl.GroupLen[0] = 1
	rightStatic.BlockSwitchingControl.GroupLen[0] = 1

	var leftTNS, rightTNS TNSData
	var leftOut, rightOut PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var scratch PsyTnsTonalityScratch

	rc := FDKaacEncPsyAdvanceTnsAndTonality(
		2,
		[]*PsyStatic{&leftStatic, &rightStatic},
		[]*PsyData{&leftData, &rightData},
		[]*TNSData{&leftTNS, &rightTNS},
		[]*PsyConfiguration{&conf, &conf},
		[]*PsyOutChannel{&leftOut, &rightOut},
		&tonality,
		&scratch,
	)
	if rc != AACEncOK {
		t.Fatalf("long psy TNS rc = %#x, want OK", rc)
	}
	if got, want := hashPsyTnsStage(&leftData, &leftTNS, &leftOut, tonality[0][:], &conf), uint64(0x220065d9fd13ca34); got != want {
		t.Fatalf("left long psy TNS hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashPsyTnsStage(&rightData, &rightTNS, &rightOut, tonality[1][:], &conf), uint64(0x9fff560dfed76531); got != want {
		t.Fatalf("right long psy TNS hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPsyAdvanceTnsAndTonalityShortVector(t *testing.T) {
	conf := mustPsyTnsConfig(t, ShortWindow)
	var data PsyData
	var spec [maxSpectralLines]FixpDBL
	preparePsyTnsShortData(&data, spec[:], &conf, 4187)

	var static PsyStatic
	static.BlockSwitchingControl.LastWindowSequence = ShortWindow
	static.BlockSwitchingControl.NoOfGroups = 2
	static.BlockSwitchingControl.GroupLen[0] = 3
	static.BlockSwitchingControl.GroupLen[1] = 5

	var tns TNSData
	var out PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var scratch PsyTnsTonalityScratch

	rc := FDKaacEncPsyAdvanceTnsAndTonality(
		1,
		[]*PsyStatic{&static},
		[]*PsyData{&data},
		[]*TNSData{&tns},
		[]*PsyConfiguration{&conf},
		[]*PsyOutChannel{&out},
		&tonality,
		&scratch,
	)
	if rc != AACEncOK {
		t.Fatalf("short psy TNS rc = %#x, want OK", rc)
	}
	if got, want := hashPsyTnsStage(&data, &tns, &out, tonality[0][:], &conf), uint64(0xdaf14bf966d20ef7); got != want {
		t.Fatalf("short psy TNS hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPsyAdvanceTnsDisabledClearsDynamicData(t *testing.T) {
	conf := mustPsyTnsConfig(t, LongWindow)
	conf.TnsConf.TnsActive = 0
	var data PsyData
	var spec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&data, spec[:], &conf, 4691)

	var static PsyStatic
	static.BlockSwitchingControl.LastWindowSequence = LongWindow
	static.BlockSwitchingControl.NoOfGroups = 1
	static.BlockSwitchingControl.GroupLen[0] = 1

	tns := TNSData{FiltersMerged: 1}
	tns.DataRaw.Long.SubBlockInfo.TnsActive = [maxTnsFilters]int{1, 1}
	tns.DataRaw.Long.SubBlockInfo.PredictionGain = [maxTnsFilters]int{44, 55}

	var out PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var scratch PsyTnsTonalityScratch

	FDKaacEncPsyAdvanceTnsAndTonality(
		1,
		[]*PsyStatic{&static},
		[]*PsyData{&data},
		[]*TNSData{&tns},
		[]*PsyConfiguration{&conf},
		[]*PsyOutChannel{&out},
		&tonality,
		&scratch,
	)

	if got, want := hashPsyTnsDynamic(&tns, &out, conf.TnsConf.MaxOrder), uint64(0xd26aa46908be1db5); got != want {
		t.Fatalf("disabled TNS dynamic hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPsyAdvanceTnsLFELeavesNoActiveFilter(t *testing.T) {
	conf := mustPsyTnsConfig(t, LongWindow)
	var data PsyData
	var spec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&data, spec[:], &conf, 4951)

	var static PsyStatic
	static.IsLFE = 1
	static.BlockSwitchingControl.LastWindowSequence = LongWindow

	tns := TNSData{FiltersMerged: 1}
	tns.DataRaw.Long.SubBlockInfo.TnsActive = [maxTnsFilters]int{1, 1}
	tns.DataRaw.Long.SubBlockInfo.PredictionGain = [maxTnsFilters]int{1234, 5678}

	var out PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var scratch PsyTnsTonalityScratch

	FDKaacEncPsyAdvanceTnsAndTonality(
		1,
		[]*PsyStatic{&static},
		[]*PsyData{&data},
		[]*TNSData{&tns},
		[]*PsyConfiguration{&conf},
		[]*PsyOutChannel{&out},
		&tonality,
		&scratch,
	)

	if tns.DataRaw.Long.SubBlockInfo.TnsActive != [maxTnsFilters]int{} {
		t.Fatalf("LFE TNS active = %v, want zero", tns.DataRaw.Long.SubBlockInfo.TnsActive)
	}
	if tns.DataRaw.Long.SubBlockInfo.PredictionGain != [maxTnsFilters]int{1234, 5678} {
		t.Fatalf("LFE TNS prediction gain was cleared")
	}
}

func TestFDKaacEncPsyAdvanceTnsRejectsInvalid(t *testing.T) {
	conf := mustPsyTnsConfig(t, LongWindow)
	var data PsyData
	var spec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&data, spec[:], &conf, 5279)
	var static PsyStatic
	static.BlockSwitchingControl.LastWindowSequence = LongWindow
	var tns TNSData
	var out PsyOutChannel
	var tonality [2][maxSFBLong]FixpSGL
	var scratch PsyTnsTonalityScratch

	tests := []struct {
		name string
		fn   func()
	}{
		{"zero channels", func() {
			FDKaacEncPsyAdvanceTnsAndTonality(0, nil, nil, nil, nil, nil, &tonality, &scratch)
		}},
		{"nil static", func() {
			FDKaacEncPsyAdvanceTnsAndTonality(1, []*PsyStatic{nil}, []*PsyData{&data}, []*TNSData{&tns}, []*PsyConfiguration{&conf}, []*PsyOutChannel{&out}, &tonality, &scratch)
		}},
		{"nil tonality", func() {
			FDKaacEncPsyAdvanceTnsAndTonality(1, []*PsyStatic{&static}, []*PsyData{&data}, []*TNSData{&tns}, []*PsyConfiguration{&conf}, []*PsyOutChannel{&out}, nil, &scratch)
		}},
		{"short spectrum", func() {
			bad := data
			bad.MdctSpectrum = bad.MdctSpectrum[:conf.GranuleLength-1]
			FDKaacEncPsyAdvanceTnsAndTonality(1, []*PsyStatic{&static}, []*PsyData{&bad}, []*TNSData{&tns}, []*PsyConfiguration{&conf}, []*PsyOutChannel{&out}, &tonality, &scratch)
		}},
		{"bad window", func() {
			bad := static
			bad.BlockSwitchingControl.LastWindowSequence = 99
			FDKaacEncPsyAdvanceTnsAndTonality(1, []*PsyStatic{&bad}, []*PsyData{&data}, []*TNSData{&tns}, []*PsyConfiguration{&conf}, []*PsyOutChannel{&out}, &tonality, &scratch)
		}},
	}

	for _, tt := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		}()
	}
}

func TestFDKaacEncPsyAdvanceTnsAllocs(t *testing.T) {
	conf := mustPsyTnsConfig(t, LongWindow)
	var seedData PsyData
	var seedSpec [maxSpectralLines]FixpDBL
	preparePsyTnsLongData(&seedData, seedSpec[:], &conf, 5801)
	var static PsyStatic
	static.BlockSwitchingControl.LastWindowSequence = LongWindow
	static.BlockSwitchingControl.NoOfGroups = 1
	static.BlockSwitchingControl.GroupLen[0] = 1
	var tonality [2][maxSFBLong]FixpSGL
	var scratch PsyTnsTonalityScratch

	allocs := testing.AllocsPerRun(1000, func() {
		data := seedData
		spec := seedSpec
		data.MdctSpectrum = spec[:]
		var tns TNSData
		var out PsyOutChannel
		rc := FDKaacEncPsyAdvanceTnsAndTonality(
			1,
			[]*PsyStatic{&static},
			[]*PsyData{&data},
			[]*TNSData{&tns},
			[]*PsyConfiguration{&conf},
			[]*PsyOutChannel{&out},
			&tonality,
			&scratch,
		)
		psyTnsSink ^= uint64(rc)
		psyTnsSink ^= hashPsyTnsStage(&data, &tns, &out, tonality[0][:], &conf)
	})
	if allocs != 0 {
		t.Fatalf("psy TNS allocations = %v, want 0", allocs)
	}
}

func mustPsyTnsConfig(t *testing.T, blockType int) PsyConfiguration {
	t.Helper()
	tnsConf, psyConf := mustTnsEncodeConfig(t, blockType)
	psyConf.TnsConf = tnsConf
	if got := FDKaacEncInitPnsConfiguration(&psyConf.PnsConf, 48000, 48000, 1, psyConf.SfbCnt, psyConf.SfbOffset[:], 2, 1); got != AACEncOK {
		t.Fatalf("PNS config rc = %#x, want OK", got)
	}
	return psyConf
}

func preparePsyTnsLongData(data *PsyData, spectrum []FixpDBL, conf *PsyConfiguration, seed uint32) {
	fillTnsDetectSpectrum(spectrum[:conf.GranuleLength], seed)
	for line := conf.LowpassLine; line < conf.GranuleLength; line++ {
		spectrum[line] = 0
	}
	*data = PsyData{MdctSpectrum: spectrum}
	data.SfbActive = conf.SfbActive
	data.LowpassLine = conf.LowpassLine

	FDKaacEncCalcSfbMaxScaleSpec(spectrum, conf.SfbOffset[:], data.SfbMaxScaleSpec.Long[:], data.SfbActive)
	minSpecShift := DfractBits - 2
	for sfb := 0; sfb < data.SfbActive; sfb++ {
		minSpecShift = minInt(minSpecShift, data.SfbMaxScaleSpec.Long[sfb])
	}
	FDKaacEncCheckBandEnergyOptim(
		spectrum,
		data.SfbMaxScaleSpec.Long[:],
		conf.SfbOffset[:],
		data.SfbActive,
		data.SfbEnergy.Long[:],
		data.SfbEnergyLdData.Long[:],
		minSpecShift-4,
	)
	for sfb := 0; sfb < data.SfbActive; sfb++ {
		data.SfbThreshold.Long[sfb] = FMultDD(data.SfbEnergy.Long[sfb], cRatio)
	}
}

func preparePsyTnsShortData(data *PsyData, spectrum []FixpDBL, conf *PsyConfiguration, seed uint32) {
	fillTnsDetectSpectrum(spectrum[:conf.GranuleLength], seed)
	shortLen := conf.GranuleLength / transFac
	for w := 0; w < transFac; w++ {
		base := w * shortLen
		for line := conf.LowpassLine; line < shortLen; line++ {
			spectrum[base+line] = 0
		}
	}
	*data = PsyData{MdctSpectrum: spectrum}
	data.SfbActive = conf.SfbActive
	data.LowpassLine = conf.LowpassLine

	for w := 0; w < transFac; w++ {
		base := w * shortLen
		FDKaacEncCalcSfbMaxScaleSpec(spectrum[base:], conf.SfbOffset[:], data.SfbMaxScaleSpec.Short[w][:], data.SfbActive)
		FDKaacEncCalcBandEnergyOptimShort(spectrum[base:], data.SfbMaxScaleSpec.Short[w][:], conf.SfbOffset[:], data.SfbActive, data.SfbEnergy.Short[w][:])
		for sfb := 0; sfb < data.SfbActive; sfb++ {
			data.SfbThreshold.Short[w][sfb] = FMultDD(data.SfbEnergy.Short[w][sfb], cRatio)
		}
	}
}

func hashPsyTnsStage(data *PsyData, tns *TNSData, out *PsyOutChannel, tonality []FixpSGL, conf *PsyConfiguration) uint64 {
	h := uint64(14695981039346656037)
	h = hashTnsSyncInt(h, data.MdctScale)
	h = hashTnsSyncInt(h, data.SfbActive)
	h = mixPsyTnsHash(h, hashFixpDBL(data.MdctSpectrum[:conf.GranuleLength]))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbEnergy.Long[:maxGroupedSFB]))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbEnergyLdData.Long[:maxGroupedSFB]))
	h = mixPsyTnsHash(h, hashFixpDBL(data.SfbThreshold.Long[:maxGroupedSFB]))
	h = mixPsyTnsHash(h, hashPsyTnsScales(data.SfbMaxScaleSpec.Long[:maxGroupedSFB]))
	h = mixPsyTnsHash(h, hashPsyTnsShortEnergy(&data.SfbEnergy.Short))
	h = mixPsyTnsHash(h, hashPsyTnsShortEnergy(&data.SfbThreshold.Short))
	h = mixPsyTnsHash(h, hashPsyTnsShortScales(&data.SfbMaxScaleSpec.Short))
	h = mixPsyTnsHash(h, hashPsyTnsDynamic(tns, out, conf.TnsConf.MaxOrder))
	h = mixPsyTnsHash(h, hashFixpSGL(tonality[:conf.SfbCnt]))
	return h
}

func hashPsyTnsDynamic(tns *TNSData, out *PsyOutChannel, order int) uint64 {
	h := uint64(14695981039346656037)
	h = hashTnsSyncInt(h, tns.FiltersMerged)
	h = hashTnsSyncInt(h, out.TNSInfo.NumOfFilters[0])
	h = hashTnsSyncInt(h, out.TNSInfo.CoefRes[0])
	for filt := 0; filt < maxTnsFilters; filt++ {
		h = hashTnsSyncInt(h, tns.DataRaw.Long.SubBlockInfo.TnsActive[filt])
		h = hashTnsSyncInt(h, tns.DataRaw.Long.SubBlockInfo.PredictionGain[filt])
		h = hashTnsSyncInt(h, out.TNSInfo.Order[0][filt])
		h = hashTnsSyncInt(h, out.TNSInfo.Length[0][filt])
		h = hashTnsSyncInt(h, out.TNSInfo.Direction[0][filt])
		for i := 0; i < order; i++ {
			h = hashTnsSyncInt(h, out.TNSInfo.Coef[0][filt][i])
		}
	}
	for w := 0; w < transFac; w++ {
		h = hashTnsSyncInt(h, out.TNSInfo.NumOfFilters[w])
		for filt := 0; filt < maxTnsFilters; filt++ {
			h = hashTnsSyncInt(h, tns.DataRaw.Short.SubBlockInfo[w].TnsActive[filt])
			h = hashTnsSyncInt(h, tns.DataRaw.Short.SubBlockInfo[w].PredictionGain[filt])
			h = hashTnsSyncInt(h, out.TNSInfo.Order[w][filt])
			h = hashTnsSyncInt(h, out.TNSInfo.Length[w][filt])
			for i := 0; i < order; i++ {
				h = hashTnsSyncInt(h, out.TNSInfo.Coef[w][filt][i])
			}
		}
	}
	return h
}

func hashPsyTnsScales(scales []int) uint64 {
	h := uint64(14695981039346656037)
	for _, v := range scales {
		h = hashTnsSyncInt(h, v)
	}
	return h
}

func hashPsyTnsShortEnergy(values *[transFac][maxSFBShort]FixpDBL) uint64 {
	h := uint64(14695981039346656037)
	for w := 0; w < transFac; w++ {
		h = mixPsyTnsHash(h, hashFixpDBL(values[w][:]))
	}
	return h
}

func hashPsyTnsShortScales(values *[transFac][maxSFBShort]int) uint64 {
	h := uint64(14695981039346656037)
	for w := 0; w < transFac; w++ {
		h = mixPsyTnsHash(h, hashPsyTnsScales(values[w][:]))
	}
	return h
}

func mixPsyTnsHash(h uint64, v uint64) uint64 {
	for i := 0; i < 8; i++ {
		h = fnv64AddByte(h, byte(v>>(8*i)))
	}
	return h
}
