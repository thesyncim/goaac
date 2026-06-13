package fdkaac

import "testing"

var tnsRuntimeSink uint64

func TestFDKaacEncTnsCoefficientQuantizerVectors(t *testing.T) {
	index4 := [...]int{-8, -6, -4, -2, 0, 1, 3, 5, 7}
	var parcor4 [tnsMaxOrder]FixpSGL
	var roundTrip4 [tnsMaxOrder]int
	fdkaacEncIndex2Parcor(index4[:], parcor4[:], len(index4), 4)
	fdkaacEncParcor2Index(parcor4[:], roundTrip4[:], len(index4), 4)

	if got, want := hashFixpSGL(parcor4[:len(index4)]), uint64(0xa3d041a6f97f5ef2); got != want {
		t.Fatalf("4-bit parcor hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashBandEnergyInts(roundTrip4[:len(index4)]), uint64(0xf40a715ea257a11d); got != want {
		t.Fatalf("4-bit parcor roundtrip hash = %#016x, want %#016x", got, want)
	}

	index3 := [...]int{-4, -3, -2, -1, 0, 1, 2, 3}
	var parcor3 [tnsMaxOrder]FixpSGL
	var roundTrip3 [tnsMaxOrder]int
	fdkaacEncIndex2Parcor(index3[:], parcor3[:], len(index3), 3)
	fdkaacEncParcor2Index(parcor3[:], roundTrip3[:], len(index3), 3)

	if got, want := hashFixpSGL(parcor3[:len(index3)]), uint64(0x49dfd6a5ff79fd96); got != want {
		t.Fatalf("3-bit parcor hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashBandEnergyInts(roundTrip3[:len(index3)]), uint64(0xec17d0c1e7adb595); got != want {
		t.Fatalf("3-bit parcor roundtrip hash = %#016x, want %#016x", got, want)
	}
}

func TestCLpcAutoToParcorVectors(t *testing.T) {
	acorr := [...]FixpDBL{
		0x50000000, -0x08000000, 0x04000000, -0x02000000, 0x01800000,
		-0x01000000, 0x00c00000, -0x00800000, 0x00600000,
	}
	var refl [tnsMaxOrder]FixpSGL
	var predictionGainM FixpDBL
	var predictionGainE int

	clpcAutoToParcor(acorr[:], 0, refl[:], len(acorr)-1, &predictionGainM, &predictionGainE)

	if got, want := hashFixpSGL(refl[:len(acorr)-1]), uint64(0x586d9133718b6177); got != want {
		t.Fatalf("autocorr parcor hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(acorr[:]), uint64(0x48ea96339ca70952); got != want {
		t.Fatalf("mutated autocorr hash = %#016x, want %#016x", got, want)
	}
	if predictionGainM != 1086914560 || predictionGainE != 1 {
		t.Fatalf("prediction gain = %d/%d, want 1086914560/1", predictionGainM, predictionGainE)
	}

	for i := range refl {
		refl[i] = 123
	}
	zero := [tnsMaxOrder + 1]FixpDBL{}
	clpcAutoToParcor(zero[:], 0, refl[:], 4, &predictionGainM, &predictionGainE)
	if got, want := hashFixpSGL(refl[:4]), uint64(0x47fe0d7eaf8e51e3); got != want {
		t.Fatalf("zero autocorr parcor hash = %#016x, want %#016x", got, want)
	}
	if predictionGainM != halfDBL || predictionGainE != 1 {
		t.Fatalf("zero autocorr prediction gain = %d/%d, want %d/1", predictionGainM, predictionGainE, halfDBL)
	}
	if refl[4] != 123 {
		t.Fatalf("zero autocorr cleared past order")
	}
}

func TestFDKaacEncTnsMergedAutoCorrelationVectors(t *testing.T) {
	longConf, _ := mustTnsEncodeConfig(t, LongWindow)
	var longSpectrum [maxSpectralLines]FixpDBL
	var rxx1, rxx2 [tnsMaxOrder + 1]FixpDBL
	fillTnsDetectSpectrum(longSpectrum[:], 811)

	fdkaacEncMergedAutoCorrelation(longSpectrum[:], &longConf, rxx1[:], rxx2[:])
	if got, want := hashFixpDBL(rxx1[:longConf.MaxOrder+1]), uint64(0x363dd50e6c4473dc); got != want {
		t.Fatalf("long merged autocorr low hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(rxx2[:longConf.MaxOrder+1]), uint64(0xb67d26c740d106d7); got != want {
		t.Fatalf("long merged autocorr high hash = %#016x, want %#016x", got, want)
	}

	shortConf, _ := mustTnsEncodeConfig(t, ShortWindow)
	var shortSpectrum [128]FixpDBL
	rxx1 = [tnsMaxOrder + 1]FixpDBL{}
	rxx2 = [tnsMaxOrder + 1]FixpDBL{}
	fillTnsDetectSpectrum(shortSpectrum[:], 1201)

	fdkaacEncMergedAutoCorrelation(shortSpectrum[:], &shortConf, rxx1[:], rxx2[:])
	if got, want := hashFixpDBL(rxx1[:shortConf.MaxOrder+1]), uint64(0x1540eff77bce1b23); got != want {
		t.Fatalf("short merged autocorr low hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(rxx2[:shortConf.MaxOrder+1]), uint64(0x98db6b1909425a6e); got != want {
		t.Fatalf("short merged autocorr high hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncTnsDetectLongVectors(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL
	fillTnsDetectSpectrum(spectrum[:], 1447)

	data.FiltersMerged = 1
	data.DataRaw.Long.SubBlockInfo.TnsActive = [maxTnsFilters]int{1, 1}
	info.NumOfFilters[0] = 2
	info.Coef[0][tnsHiFilt][0] = 7
	info.Coef[0][tnsLoFilt][0] = -7

	rc := FDKaacEncTnsDetect(&data, &conf, &info, psy.SfbCnt, spectrum[:], 0, LongWindow)
	if rc != 0 {
		t.Fatalf("long tns detect rc = %d, want 0", rc)
	}
	if got, want := hashTnsDetectState(&data, &info, 0, conf.MaxOrder), uint64(0x2e521fc24a29680c); got != want {
		t.Fatalf("long tns detect state hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncTnsDetectShortVectors(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, ShortWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [128]FixpDBL
	fillTnsDetectSpectrum(spectrum[:], 1693)

	const subBlock = 5
	rc := FDKaacEncTnsDetect(&data, &conf, &info, psy.SfbCnt, spectrum[:], subBlock, ShortWindow)
	if rc != 0 {
		t.Fatalf("short tns detect rc = %d, want 0", rc)
	}
	if got, want := hashTnsDetectState(&data, &info, subBlock, conf.MaxOrder), uint64(0x082438d1adcc19ad); got != want {
		t.Fatalf("short tns detect state hash = %#016x, want %#016x", got, want)
	}
	if data.DataRaw.Short.SubBlockInfo[subBlock].TnsActive[tnsLoFilt] != 0 ||
		info.NumOfFilters[subBlock] > 1 {
		t.Fatalf("short tns detect enabled low filter")
	}
}

func TestFDKaacEncTnsDetectInactiveResetsState(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	conf.TnsActive = 0
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL

	data.FiltersMerged = 1
	data.DataRaw.Long.SubBlockInfo.TnsActive = [maxTnsFilters]int{1, 1}
	data.DataRaw.Long.SubBlockInfo.PredictionGain = [maxTnsFilters]int{22, 33}
	info.NumOfFilters[0] = 2
	info.Order[0] = [maxTnsFilters]int{4, 3}
	info.Length[0] = [maxTnsFilters]int{9, 8}
	info.Coef[0][tnsHiFilt][0] = 5
	info.Coef[0][tnsLoFilt][0] = -5

	rc := FDKaacEncTnsDetect(&data, &conf, &info, psy.SfbCnt, spectrum[:], 0, LongWindow)
	if rc != 0 {
		t.Fatalf("inactive tns detect rc = %d, want 0", rc)
	}
	if got, want := hashTnsDetectState(&data, &info, 0, conf.MaxOrder), uint64(0x58bd901001c70a61); got != want {
		t.Fatalf("inactive tns detect state hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncTnsEncodeLongVectors(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL
	fillTnsEncodeSpectrum(spectrum[:], 131)

	beforeLow := hashFixpDBL(spectrum[:conf.LpcStartLine[tnsHiFilt]])
	beforeHigh := hashFixpDBL(spectrum[conf.LpcStopLine:])

	data.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	info.NumOfFilters[0] = 1
	info.CoefRes[0] = conf.CoefRes
	info.Order[0][tnsHiFilt] = 4
	info.Coef[0][tnsHiFilt] = [tnsMaxOrder]int{2, -1, 1, -2}

	rc := FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
	if rc != 0 {
		t.Fatalf("long tns encode rc = %d, want 0", rc)
	}
	if hashFixpDBL(spectrum[:conf.LpcStartLine[tnsHiFilt]]) != beforeLow {
		t.Fatalf("long tns encode changed spectrum before high filter start")
	}
	if hashFixpDBL(spectrum[conf.LpcStopLine:]) != beforeHigh {
		t.Fatalf("long tns encode changed spectrum after stop line")
	}
	if got, want := hashFixpDBL(spectrum[conf.LpcStartLine[tnsHiFilt]:conf.LpcStopLine]), uint64(0x1ad37e9c8d97f37b); got != want {
		t.Fatalf("long tns filtered hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncTnsEncodeLongTwoFilterVector(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL
	fillTnsEncodeSpectrum(spectrum[:], 257)

	data.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	data.DataRaw.Long.SubBlockInfo.TnsActive[tnsLoFilt] = 1
	info.NumOfFilters[0] = 2
	info.CoefRes[0] = conf.CoefRes
	info.Order[0][tnsHiFilt] = 4
	info.Order[0][tnsLoFilt] = 3
	info.Coef[0][tnsHiFilt] = [tnsMaxOrder]int{3, -2, 1, -1}
	info.Coef[0][tnsLoFilt] = [tnsMaxOrder]int{-2, 1, 2}

	rc := FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, StopWindow)
	if rc != 0 {
		t.Fatalf("two-filter tns encode rc = %d, want 0", rc)
	}
	if got, want := hashFixpDBL(spectrum[conf.LpcStartLine[tnsLoFilt]:conf.LpcStartLine[tnsHiFilt]]), uint64(0x491dc79247d170f4); got != want {
		t.Fatalf("two-filter low region hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashFixpDBL(spectrum[conf.LpcStartLine[tnsHiFilt]:conf.LpcStopLine]), uint64(0x89ce43513d23d884); got != want {
		t.Fatalf("two-filter high region hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncTnsEncodeShortVector(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, ShortWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [128]FixpDBL
	fillTnsEncodeSpectrum(spectrum[:], 389)

	const subBlock = 3
	data.DataRaw.Short.SubBlockInfo[subBlock].TnsActive[tnsHiFilt] = 1
	info.NumOfFilters[subBlock] = 1
	info.CoefRes[subBlock] = conf.CoefRes
	info.Order[subBlock][tnsHiFilt] = 3
	info.Coef[subBlock][tnsHiFilt] = [tnsMaxOrder]int{2, -1, 1}

	rc := FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], subBlock, ShortWindow)
	if rc != 0 {
		t.Fatalf("short tns encode rc = %d, want 0", rc)
	}
	if got, want := hashFixpDBL(spectrum[conf.LpcStartLine[tnsHiFilt]:conf.LpcStopLine]), uint64(0xe89831a344b4c9a2); got != want {
		t.Fatalf("short tns filtered hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncTnsEncodeInactiveNoop(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL
	fillTnsEncodeSpectrum(spectrum[:], 401)
	before := hashFixpDBL(spectrum[:])

	info.NumOfFilters[0] = maxTnsFilters + 1

	rc := FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
	if rc != 1 {
		t.Fatalf("inactive tns encode rc = %d, want 1", rc)
	}
	if got := hashFixpDBL(spectrum[:]); got != before {
		t.Fatalf("inactive tns encode changed spectrum: %#016x -> %#016x", before, got)
	}
}

func TestFDKaacEncTnsSyncLongVectors(t *testing.T) {
	var destData, srcData TNSData
	var destInfo, srcInfo TNSInfo
	conf := TNSConfig{MaxOrder: 5}

	srcData.FiltersMerged = 1
	srcData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	srcInfo.NumOfFilters[0] = 1
	srcInfo.Order[0][tnsHiFilt] = 3
	srcInfo.Length[0][tnsHiFilt] = 24
	srcInfo.Direction[0][tnsHiFilt] = 1
	srcInfo.CoefCompress[0][tnsHiFilt] = 1
	srcInfo.Coef[0][tnsHiFilt] = [tnsMaxOrder]int{1, 0, 0, 0, 0, 7}

	destInfo.NumOfFilters[0] = 2
	destInfo.Order[0][tnsHiFilt] = 9
	destInfo.Length[0][tnsHiFilt] = 9
	destInfo.Coef[0][tnsHiFilt] = [tnsMaxOrder]int{0, 0, 0, 0, 0, 99}

	FDKaacEncTnsSync(&destData, &srcData, &destInfo, &srcInfo, LongWindow, StartWindow, &conf)

	if destData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] != 1 || destInfo.NumOfFilters[0] != 1 {
		t.Fatalf("long sync active/filter count = %d/%d, want 1/1",
			destData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt], destInfo.NumOfFilters[0])
	}
	if destData.FiltersMerged != 1 {
		t.Fatalf("long sync filtersMerged = %d, want 1", destData.FiltersMerged)
	}
	if destInfo.Order[0][tnsHiFilt] != 3 ||
		destInfo.Length[0][tnsHiFilt] != 24 ||
		destInfo.Direction[0][tnsHiFilt] != 1 ||
		destInfo.CoefCompress[0][tnsHiFilt] != 1 {
		t.Fatalf("long sync metadata = order:%d length:%d direction:%d compress:%d",
			destInfo.Order[0][tnsHiFilt], destInfo.Length[0][tnsHiFilt],
			destInfo.Direction[0][tnsHiFilt], destInfo.CoefCompress[0][tnsHiFilt])
	}
	if got, want := hashBandEnergyInts(destInfo.Coef[0][tnsHiFilt][:conf.MaxOrder]), uint64(0x0c593f1b0d139064); got != want {
		t.Fatalf("long synced coefficients hash = %#016x, want %#016x", got, want)
	}
	if destInfo.Coef[0][tnsHiFilt][5] != 99 {
		t.Fatalf("long sync copied past max order")
	}
}

func TestFDKaacEncTnsSyncInactiveAndDivergentVectors(t *testing.T) {
	conf := TNSConfig{MaxOrder: 5}

	var destData, srcData TNSData
	var destInfo, srcInfo TNSInfo
	destData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	destInfo.NumOfFilters[0] = 1
	destInfo.Order[0][tnsHiFilt] = 4

	FDKaacEncTnsSync(&destData, &srcData, &destInfo, &srcInfo, LongWindow, LongWindow, &conf)

	if destData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] != 0 || destInfo.NumOfFilters[0] != 0 {
		t.Fatalf("inactive source did not clear dest: active/filter count = %d/%d",
			destData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt], destInfo.NumOfFilters[0])
	}
	if destInfo.Order[0][tnsHiFilt] != 4 {
		t.Fatalf("inactive source unexpectedly cleared order")
	}

	destData = TNSData{}
	srcData = TNSData{}
	destInfo = TNSInfo{}
	srcInfo = TNSInfo{}
	destData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	srcData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	destInfo.NumOfFilters[0] = 1
	srcInfo.NumOfFilters[0] = 1
	destInfo.Coef[0][tnsHiFilt][0] = 4
	srcInfo.Order[0][tnsHiFilt] = 2
	srcInfo.Length[0][tnsHiFilt] = 18

	FDKaacEncTnsSync(&destData, &srcData, &destInfo, &srcInfo, LongWindow, LongWindow, &conf)

	if destInfo.Order[0][tnsHiFilt] != 0 || destInfo.Length[0][tnsHiFilt] != 0 || destInfo.Coef[0][tnsHiFilt][0] != 4 {
		t.Fatalf("divergent sync changed destination: order:%d length:%d coef0:%d",
			destInfo.Order[0][tnsHiFilt], destInfo.Length[0][tnsHiFilt], destInfo.Coef[0][tnsHiFilt][0])
	}
}

func TestFDKaacEncTnsSyncShortAndMismatchVectors(t *testing.T) {
	conf := TNSConfig{MaxOrder: 5}
	var destData, srcData TNSData
	var destInfo, srcInfo TNSInfo

	for w := 0; w < transFac; w++ {
		srcData.DataRaw.Short.SubBlockInfo[w].TnsActive[tnsHiFilt] = 1
		srcInfo.NumOfFilters[w] = 1
		srcInfo.Order[w][tnsHiFilt] = w%3 + 1
		srcInfo.Length[w][tnsHiFilt] = 10 + w
		srcInfo.Direction[w][tnsHiFilt] = w & 1
		srcInfo.CoefCompress[w][tnsHiFilt] = (w + 1) & 1
		srcInfo.Coef[w][tnsHiFilt][0] = w & 1
	}
	destInfo.Coef[5][tnsHiFilt][0] = 7

	FDKaacEncTnsSync(&destData, &srcData, &destInfo, &srcInfo, ShortWindow, ShortWindow, &conf)

	if got, want := hashTnsSyncShort(&destData, &destInfo), uint64(0x8bf1a277f0947d8d); got != want {
		t.Fatalf("short sync hash = %#016x, want %#016x", got, want)
	}
	if destInfo.Coef[5][tnsHiFilt][0] != 7 {
		t.Fatalf("divergent short window was synchronized")
	}

	before := hashTnsSyncShort(&destData, &destInfo)
	FDKaacEncTnsSync(&destData, &srcData, &destInfo, &srcInfo, ShortWindow, LongWindow, &conf)
	if got := hashTnsSyncShort(&destData, &destInfo); got != before {
		t.Fatalf("mixed short/long sync changed destination: %#016x -> %#016x", before, got)
	}
}

func TestFDKaacEncTnsSyncRejectsInvalid(t *testing.T) {
	var data TNSData
	var info TNSInfo
	conf := TNSConfig{MaxOrder: 5}

	tests := []struct {
		name string
		fn   func()
	}{
		{"nil dest data", func() { FDKaacEncTnsSync(nil, &data, &info, &info, LongWindow, LongWindow, &conf) }},
		{"nil source data", func() { FDKaacEncTnsSync(&data, nil, &info, &info, LongWindow, LongWindow, &conf) }},
		{"nil dest info", func() { FDKaacEncTnsSync(&data, &data, nil, &info, LongWindow, LongWindow, &conf) }},
		{"nil source info", func() { FDKaacEncTnsSync(&data, &data, &info, nil, LongWindow, LongWindow, &conf) }},
		{"nil config", func() { FDKaacEncTnsSync(&data, &data, &info, &info, LongWindow, LongWindow, nil) }},
		{"bad dest block", func() { FDKaacEncTnsSync(&data, &data, &info, &info, 99, LongWindow, &conf) }},
		{"bad source block", func() { FDKaacEncTnsSync(&data, &data, &info, &info, LongWindow, 99, &conf) }},
		{"bad order", func() {
			bad := TNSConfig{MaxOrder: tnsMaxOrder + 1}
			FDKaacEncTnsSync(&data, &data, &info, &info, LongWindow, LongWindow, &bad)
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

func TestFDKaacEncTnsEncodeRejectsInvalid(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL
	data.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	info.NumOfFilters[0] = 1
	info.Order[0][tnsHiFilt] = 1

	tests := []struct {
		name string
		fn   func()
	}{
		{"nil info", func() {
			FDKaacEncTnsEncode(nil, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
		}},
		{"nil data", func() {
			FDKaacEncTnsEncode(&info, nil, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
		}},
		{"nil config", func() {
			FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, nil, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
		}},
		{"bad block", func() {
			FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, 99)
		}},
		{"bad subblock", func() {
			FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], transFac, LongWindow)
		}},
		{"bad coef resolution", func() {
			bad := conf
			bad.CoefRes = 5
			FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &bad, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
		}},
		{"bad lines", func() {
			bad := conf
			bad.LpcStartLine[tnsHiFilt] = bad.LpcStopLine + 1
			FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &bad, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
		}},
		{"bad filter count", func() {
			badInfo := info
			badInfo.NumOfFilters[0] = maxTnsFilters + 1
			FDKaacEncTnsEncode(&badInfo, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
		}},
		{"bad coefficient", func() {
			badInfo := info
			badInfo.Coef[0][tnsHiFilt][0] = 8
			FDKaacEncTnsEncode(&badInfo, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:], 0, LongWindow)
		}},
		{"short spectrum", func() {
			FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], spectrum[:conf.LpcStopLine-1], 0, LongWindow)
		}},
		{"bad parcor order", func() {
			var parcor [tnsMaxOrder]FixpSGL
			var index [tnsMaxOrder]int
			fdkaacEncParcor2Index(parcor[:], index[:], tnsMaxOrder+1, 4)
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

func TestFDKaacEncTnsDetectRejectsInvalid(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL

	tests := []struct {
		name string
		fn   func()
	}{
		{"nil data", func() { FDKaacEncTnsDetect(nil, &conf, &info, psy.SfbCnt, spectrum[:], 0, LongWindow) }},
		{"nil config", func() { FDKaacEncTnsDetect(&data, nil, &info, psy.SfbCnt, spectrum[:], 0, LongWindow) }},
		{"nil info", func() { FDKaacEncTnsDetect(&data, &conf, nil, psy.SfbCnt, spectrum[:], 0, LongWindow) }},
		{"bad block", func() { FDKaacEncTnsDetect(&data, &conf, &info, psy.SfbCnt, spectrum[:], 0, 99) }},
		{"bad subblock", func() { FDKaacEncTnsDetect(&data, &conf, &info, psy.SfbCnt, spectrum[:], transFac, LongWindow) }},
		{"bad sfb count", func() { FDKaacEncTnsDetect(&data, &conf, &info, -1, spectrum[:], 0, LongWindow) }},
		{"bad coef resolution", func() {
			bad := conf
			bad.CoefRes = 5
			FDKaacEncTnsDetect(&data, &bad, &info, psy.SfbCnt, spectrum[:], 0, LongWindow)
		}},
		{"bad limit order", func() {
			bad := conf
			bad.ConfTab.TnsLimitOrder[tnsHiFilt] = bad.MaxOrder + 1
			FDKaacEncTnsDetect(&data, &bad, &info, psy.SfbCnt, spectrum[:], 0, LongWindow)
		}},
		{"bad lines", func() {
			bad := conf
			bad.LpcStartLine[tnsLoFilt] = bad.LpcStopLine + 1
			FDKaacEncTnsDetect(&data, &bad, &info, psy.SfbCnt, spectrum[:], 0, LongWindow)
		}},
		{"bad split", func() {
			bad := conf
			bad.ConfTab.ACFSplit = [maxTnsFilters]int{2, 3}
			FDKaacEncTnsDetect(&data, &bad, &info, psy.SfbCnt, spectrum[:], 0, LongWindow)
		}},
		{"short autocorr output", func() {
			var rxx [tnsMaxOrder + 1]FixpDBL
			fdkaacEncMergedAutoCorrelation(spectrum[:], &conf, rxx[:conf.MaxOrder], rxx[:])
		}},
		{"short lpc autocorr", func() {
			var refl [tnsMaxOrder]FixpSGL
			clpcAutoToParcor(spectrum[:2], 0, refl[:], 2, nil, nil)
		}},
		{"partial gain output", func() {
			var refl [tnsMaxOrder]FixpSGL
			var gain FixpDBL
			clpcAutoToParcor(spectrum[:3], 0, refl[:], 2, &gain, nil)
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

func TestFDKaacEncTnsSyncAllocs(t *testing.T) {
	conf := TNSConfig{MaxOrder: 5}
	var destData, srcData TNSData
	var destInfo, srcInfo TNSInfo
	srcData.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	srcInfo.NumOfFilters[0] = 1
	srcInfo.Order[0][tnsHiFilt] = 2
	srcInfo.Length[0][tnsHiFilt] = 23
	srcInfo.Coef[0][tnsHiFilt][0] = 1

	allocs := testing.AllocsPerRun(1000, func() {
		FDKaacEncTnsSync(&destData, &srcData, &destInfo, &srcInfo, LongWindow, LongWindow, &conf)
		tnsRuntimeSink ^= hashBandEnergyInts(destInfo.Coef[0][tnsHiFilt][:conf.MaxOrder])
	})
	if allocs != 0 {
		t.Fatalf("tns sync allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncTnsDetectAllocs(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL
	fillTnsDetectSpectrum(spectrum[:], 2053)

	allocs := testing.AllocsPerRun(1000, func() {
		data = TNSData{}
		info = TNSInfo{}
		working := spectrum
		rc := FDKaacEncTnsDetect(&data, &conf, &info, psy.SfbCnt, working[:], 0, LongWindow)
		tnsRuntimeSink ^= uint64(rc)
		tnsRuntimeSink ^= hashTnsDetectState(&data, &info, 0, conf.MaxOrder)
	})
	if allocs != 0 {
		t.Fatalf("tns detect allocations = %v, want 0", allocs)
	}
}

func TestFDKaacEncTnsEncodeAllocs(t *testing.T) {
	conf, psy := mustTnsEncodeConfig(t, LongWindow)
	var data TNSData
	var info TNSInfo
	var spectrum [maxSpectralLines]FixpDBL
	fillTnsEncodeSpectrum(spectrum[:], 557)

	data.DataRaw.Long.SubBlockInfo.TnsActive[tnsHiFilt] = 1
	info.NumOfFilters[0] = 1
	info.Order[0][tnsHiFilt] = 4
	info.Coef[0][tnsHiFilt] = [tnsMaxOrder]int{2, -1, 1, -2}

	allocs := testing.AllocsPerRun(1000, func() {
		working := spectrum
		rc := FDKaacEncTnsEncode(&info, &data, psy.SfbCnt, &conf, psy.SfbOffset[psy.SfbActive], working[:], 0, LongWindow)
		tnsRuntimeSink ^= uint64(rc)
		tnsRuntimeSink ^= hashFixpDBL(working[conf.LpcStartLine[tnsHiFilt]:conf.LpcStopLine])
	})
	if allocs != 0 {
		t.Fatalf("tns encode allocations = %v, want 0", allocs)
	}
}

func mustTnsEncodeConfig(t *testing.T, blockType int) (TNSConfig, PsyConfiguration) {
	t.Helper()
	psy := mustPsyConfigurationForTNS(t, blockType)
	var conf TNSConfig
	if got := FDKaacEncInitTnsConfiguration(96000, 48000, 2, blockType, 1024, 0, 0, &conf, &psy, 1, 1); got != AACEncOK {
		t.Fatalf("TNS config rc = %#x, want OK", got)
	}
	return conf, psy
}

func fillTnsEncodeSpectrum(spectrum []FixpDBL, seed uint32) {
	state := seed
	for i := range spectrum {
		state = state*1664525 + 1013904223
		v := int32((state >> 8) & 0x1fffff)
		v -= 0x100000
		spectrum[i] = FixpDBL(v << 7)
	}
}

func fillTnsDetectSpectrum(spectrum []FixpDBL, seed uint32) {
	state := seed
	acc := int32(0)
	for i := range spectrum {
		state = state*1103515245 + 12345
		target := int32((state >> 10) & 0x3ffff)
		target -= 0x20000
		acc = (acc*7 + target) >> 3
		spectrum[i] = FixpDBL(acc << 9)
	}
}

func hashTnsDetectState(data *TNSData, info *TNSInfo, subBlock int, order int) uint64 {
	h := uint64(14695981039346656037)
	sbInfo := tnsSubblockInfo(data, LongWindow, 0)
	if subBlock != 0 {
		sbInfo = tnsSubblockInfo(data, ShortWindow, subBlock)
	}
	h = hashTnsSyncInt(h, data.FiltersMerged)
	h = hashTnsSyncInt(h, info.NumOfFilters[subBlock])
	h = hashTnsSyncInt(h, info.CoefRes[subBlock])
	for filt := 0; filt < maxTnsFilters; filt++ {
		h = hashTnsSyncInt(h, sbInfo.TnsActive[filt])
		h = hashTnsSyncInt(h, sbInfo.PredictionGain[filt])
		h = hashTnsSyncInt(h, info.Order[subBlock][filt])
		h = hashTnsSyncInt(h, info.Length[subBlock][filt])
		h = hashTnsSyncInt(h, info.Direction[subBlock][filt])
		for i := 0; i < order; i++ {
			h = hashTnsSyncInt(h, info.Coef[subBlock][filt][i])
		}
	}
	return h
}

func hashTnsSyncShort(data *TNSData, info *TNSInfo) uint64 {
	h := uint64(14695981039346656037)
	for w := 0; w < transFac; w++ {
		h = hashTnsSyncInt(h, data.DataRaw.Short.SubBlockInfo[w].TnsActive[tnsHiFilt])
		h = hashTnsSyncInt(h, info.NumOfFilters[w])
		h = hashTnsSyncInt(h, info.Order[w][tnsHiFilt])
		h = hashTnsSyncInt(h, info.Length[w][tnsHiFilt])
		h = hashTnsSyncInt(h, info.Direction[w][tnsHiFilt])
		h = hashTnsSyncInt(h, info.CoefCompress[w][tnsHiFilt])
		for i := 0; i < 2; i++ {
			h = hashTnsSyncInt(h, info.Coef[w][tnsHiFilt][i])
		}
	}
	return h
}

func hashTnsSyncInt(h uint64, v int) uint64 {
	u := uint32(v)
	h = fnv64AddByte(h, byte(u))
	h = fnv64AddByte(h, byte(u>>8))
	h = fnv64AddByte(h, byte(u>>16))
	h = fnv64AddByte(h, byte(u>>24))
	return h
}
