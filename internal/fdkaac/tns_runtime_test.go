package fdkaac

import "testing"

var tnsRuntimeSink uint64

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
