package fdkaac

import "testing"

var msStereoSink FixpDBL
var msStereoIntSink int
var msStereoHashSink uint64

func TestFDKaacEncMsStereoProcessingAllPromotionVectors(t *testing.T) {
	var specL, specR [16]FixpDBL
	var enL, enR, enM, enS [8]FixpDBL
	var thrL, thrR [8]FixpDBL
	var spreadL, spreadR [8]FixpDBL
	var enLLd, enRLd, enMLd, enSLd [8]FixpDBL
	var thrLLd, thrRLd [8]FixpDBL
	msMask := [...]int{7, 7, 7, 7, 7, 7, 7, 7}
	sfbOffset := [...]int{0, 2, 4, 6, 8, 10, 12, 14, 16}

	fillMsStereoBase(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:])
	enLLd[1] = 900
	enRLd[1] = 1000
	enMLd[1] = 10000
	enSLd[1] = 9000

	msDigest := -1
	FDKaacEncMsStereoProcessing(
		specL[:], specR[:],
		enL[:], enR[:], enM[:], enS[:],
		thrL[:], thrR[:], spreadL[:], spreadR[:],
		enLLd[:], enRLd[:], enMLd[:], enSLd[:],
		thrLLd[:], thrRLd[:],
		nil, &msDigest, msMask[:],
		1, 8, 4, 3, sfbOffset[:],
	)

	if msDigest != MsMaskAll {
		t.Fatalf("ms digest = %d, want %d", msDigest, MsMaskAll)
	}
	assertIntSlice(t, "all ms mask", msMask[:], []int{1, 1, 1, 7, 1, 1, 1, 7}, 0xbc50a7fe31905ae5)
	assertFixpDBLSlice(t, "all left spectrum", specL[:], []FixpDBL{1333, 1667, 2000, 2334, 2667, 3001, 2000, 3000, 4001, 4335, 4668, 5002, 5335, 5669, 10000, 11000}, 0x329572801743e35a)
	assertFixpDBLSlice(t, "all right spectrum", specR[:], []FixpDBL{-5333, -4667, -4000, -3334, -2667, -2001, 4669, 4336, -1, 665, 1332, 1998, 2665, 3331, 2005, 1672}, 0xcb6369ed788dc601)
	assertFixpDBLSlice(t, "all left threshold", thrL[:], []FixpDBL{25165824, 33554432, 41943040, 83886080, 58720256, 67108864, 75497472, 150994944}, 0xe15848e4e2013c2b)
	assertFixpDBLSlice(t, "all right threshold", thrR[:], []FixpDBL{25165824, 33554432, 41943040, 50331648, 58720256, 67108864, 75497472, 83886080}, 0x88f09d2632fe8a39)
	assertFixpDBLSlice(t, "all left energy", enL[:], []FixpDBL{11534336, 12582912, 13631488, 16777216, 15728640, 16777216, 17825792, 25165824}, 0x9df7c47bca9db015)
	assertFixpDBLSlice(t, "all right energy", enR[:], []FixpDBL{10223616, 11010048, 11796480, 15728640, 13369344, 14155776, 14942208, 22020096}, 0x76e364eea42440a2)
	assertFixpDBLSlice(t, "all left energy ld", enLLd[:], []FixpDBL{800, 10000, 820, 10300, 840, 850, 860, 10700}, 0x771c782851a898b8)
	assertFixpDBLSlice(t, "all right energy ld", enRLd[:], []FixpDBL{900, 9000, 920, 11300, 940, 950, 960, 11700}, 0xc40306306f9673f0)
	assertFixpDBLSlice(t, "all left spread", spreadL[:], []FixpDBL{6291456, 7864320, 9437184, 25165824, 12582912, 14155776, 15728640, 41943040}, 0x5a6c9d613b6c64fc)
	assertFixpDBLSlice(t, "all right spread", spreadR[:], []FixpDBL{6291456, 7864320, 9437184, 22020096, 12582912, 14155776, 15728640, 34603008}, 0x5ea7a7e0ff596fbc)
}

func TestFDKaacEncMsStereoProcessingSomeWithISVectors(t *testing.T) {
	var specL, specR [12]FixpDBL
	var enL, enR, enM, enS [6]FixpDBL
	var thrL, thrR [6]FixpDBL
	var spreadL, spreadR [6]FixpDBL
	var enLLd, enRLd, enMLd, enSLd [6]FixpDBL
	var thrLLd, thrRLd [6]FixpDBL
	isBook := [...]int{0, 1, 0, 0, 0, 0}
	msMask := [...]int{7, 1, 7, 7, 7, 7}
	sfbOffset := [...]int{0, 2, 4, 6, 8, 10, 12}

	fillMsStereoBase(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:])
	for _, i := range [...]int{2, 3, 4} {
		enLLd[i] = 900
		enRLd[i] = 1000
		enMLd[i] = FixpDBL(10000 + i*100)
		enSLd[i] = FixpDBL(9000 + i*100)
	}

	msDigest := -1
	FDKaacEncMsStereoProcessing(
		specL[:], specR[:],
		enL[:], enR[:], enM[:], enS[:],
		thrL[:], thrR[:], spreadL[:], spreadR[:],
		enLLd[:], enRLd[:], enMLd[:], enSLd[:],
		thrLLd[:], thrRLd[:],
		isBook[:], &msDigest, msMask[:],
		1, 6, 6, 5, sfbOffset[:],
	)

	if msDigest != MsMaskSome {
		t.Fatalf("ms digest = %d, want %d", msDigest, MsMaskSome)
	}
	assertIntSlice(t, "some ms mask", msMask[:], []int{1, 1, 0, 0, 0, 7}, 0x5d674e6468687bc2)
	assertFixpDBLSlice(t, "some left spectrum", specL[:], []FixpDBL{1333, 1667, -2000, -1000, 0, 1000, 2000, 3000, 4000, 5000, 6000, 7000}, 0x4bd09b22756f4f0f)
	assertFixpDBLSlice(t, "some right spectrum", specR[:], []FixpDBL{-5333, -4667, 6001, 5668, 5335, 5002, 4669, 4336, 4003, 3670, 3337, 3004}, 0xd013543e5a4d87fd)
	assertFixpDBLSlice(t, "some left threshold", thrL[:], []FixpDBL{25165824, 50331648, 67108864, 83886080, 100663296, 117440512}, 0xfb7f14257bd4246f)
	assertFixpDBLSlice(t, "some right threshold", thrR[:], []FixpDBL{25165824, 33554432, 41943040, 50331648, 58720256, 67108864}, 0xe0c0bab0af157c86)
	assertFixpDBLSlice(t, "some left energy ld", enLLd[:], []FixpDBL{800, 10100, 900, 900, 900, 10500}, 0x142e35ae48f02eaf)
	assertFixpDBLSlice(t, "some right energy ld", enRLd[:], []FixpDBL{900, 11100, 1000, 1000, 1000, 11500}, 0x964f81bbc7d866e4)
	assertFixpDBLSlice(t, "some left spread", spreadL[:], []FixpDBL{6291456, 16777216, 20971520, 25165824, 29360128, 33554432}, 0x1512c967d161285f)
	assertFixpDBLSlice(t, "some right spread", spreadR[:], []FixpDBL{6291456, 15728640, 18874368, 22020096, 25165824, 28311552}, 0x40f39446660a0a75)
}

func TestFDKaacEncMsStereoProcessingDisabledVectors(t *testing.T) {
	var specL, specR [8]FixpDBL
	var enL, enR, enM, enS [4]FixpDBL
	var thrL, thrR [4]FixpDBL
	var spreadL, spreadR [4]FixpDBL
	var enLLd, enRLd, enMLd, enSLd [4]FixpDBL
	var thrLLd, thrRLd [4]FixpDBL
	msMask := [...]int{7, 7, 7, 7}
	sfbOffset := [...]int{0, 2, 4, 6, 8}

	fillMsStereoBase(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:])
	msDigest := -1
	FDKaacEncMsStereoProcessing(
		specL[:], specR[:],
		enL[:], enR[:], enM[:], enS[:],
		thrL[:], thrR[:], spreadL[:], spreadR[:],
		enLLd[:], enRLd[:], enMLd[:], enSLd[:],
		thrLLd[:], thrRLd[:],
		nil, &msDigest, msMask[:],
		0, 4, 4, 4, sfbOffset[:],
	)

	if msDigest != MsMaskNone {
		t.Fatalf("ms digest = %d, want %d", msDigest, MsMaskNone)
	}
	assertIntSlice(t, "disabled ms mask", msMask[:], []int{0, 0, 0, 0}, 0x88201fb960ff6465)
	if got, want := hashFixpDBL(specL[:]), uint64(0xaa30ba553e326ad0); got != want {
		t.Fatalf("disabled left spectrum hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncMsStereoProcessingRejectsInvalid(t *testing.T) {
	var specL, specR [8]FixpDBL
	var enL, enR, enM, enS [4]FixpDBL
	var thrL, thrR [4]FixpDBL
	var spreadL, spreadR [4]FixpDBL
	var enLLd, enRLd, enMLd, enSLd [4]FixpDBL
	var thrLLd, thrRLd [4]FixpDBL
	var isBook [4]int
	var msMask [4]int
	sfbOffset := [...]int{0, 2, 4, 6, 8}
	msDigest := 0

	fillMsStereoBase(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:])
	call := func(
		specLSlice, specRSlice []FixpDBL,
		enLSlice, enRSlice, enMSlice, enSSlice []FixpDBL,
		thrLSlice, thrRSlice []FixpDBL,
		spreadLSlice, spreadRSlice []FixpDBL,
		enLLdSlice, enRLdSlice, enMLdSlice, enSLdSlice []FixpDBL,
		thrLLdSlice, thrRLdSlice []FixpDBL,
		isBookSlice []int,
		digest *int,
		mask []int,
		cnt, perGroup, maxPerGroup int,
		offset []int,
	) {
		FDKaacEncMsStereoProcessing(specLSlice, specRSlice, enLSlice, enRSlice, enMSlice, enSSlice, thrLSlice, thrRSlice, spreadLSlice, spreadRSlice, enLLdSlice, enRLdSlice, enMLdSlice, enSLdSlice, thrLLdSlice, thrRLdSlice, isBookSlice, digest, mask, 1, cnt, perGroup, maxPerGroup, offset)
	}

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "nil digest", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, nil, msMask[:], 4, 4, 4, sfbOffset[:])
		}},
		{name: "bad count", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 3, 2, 2, sfbOffset[:])
		}},
		{name: "bad group", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 4, 4, 5, sfbOffset[:])
		}},
		{name: "short offsets", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 4, 4, 4, sfbOffset[:4])
		}},
		{name: "decreasing offsets", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 4, 4, 4, []int{0, 2, 1, 6, 8})
		}},
		{name: "short spectrum", fn: func() {
			call(specL[:7], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 4, 4, 4, sfbOffset[:])
		}},
		{name: "short is-book", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], isBook[:3], &msDigest, msMask[:], 4, 4, 4, sfbOffset[:])
		}},
		{name: "short mask", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:3], 4, 4, 4, sfbOffset[:])
		}},
		{name: "short energy", fn: func() {
			call(specL[:], specR[:], enL[:3], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 4, 4, 4, sfbOffset[:])
		}},
		{name: "short threshold", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:3], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 4, 4, 4, sfbOffset[:])
		}},
		{name: "short spread", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:3], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 4, 4, 4, sfbOffset[:])
		}},
		{name: "short ld", fn: func() {
			call(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:3], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 4, 4, 4, sfbOffset[:])
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

func TestFDKaacEncMsStereoProcessingAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var specL, specR [16]FixpDBL
		var enL, enR, enM, enS [8]FixpDBL
		var thrL, thrR [8]FixpDBL
		var spreadL, spreadR [8]FixpDBL
		var enLLd, enRLd, enMLd, enSLd [8]FixpDBL
		var thrLLd, thrRLd [8]FixpDBL
		msMask := [...]int{7, 7, 7, 7, 7, 7, 7, 7}
		sfbOffset := [...]int{0, 2, 4, 6, 8, 10, 12, 14, 16}

		fillMsStereoBase(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:])
		enLLd[1] = 900
		enRLd[1] = 1000
		enMLd[1] = 10000
		enSLd[1] = 9000
		msDigest := 0

		FDKaacEncMsStereoProcessing(specL[:], specR[:], enL[:], enR[:], enM[:], enS[:], thrL[:], thrR[:], spreadL[:], spreadR[:], enLLd[:], enRLd[:], enMLd[:], enSLd[:], thrLLd[:], thrRLd[:], nil, &msDigest, msMask[:], 1, 8, 4, 3, sfbOffset[:])
		msStereoSink = specL[0] + specR[1] + enL[2] + thrL[4] + spreadL[5]
		msStereoIntSink = msDigest + msMask[0]
		msStereoHashSink = hashFixpDBL(specL[:])
	})
	if allocs != 0 {
		t.Fatalf("ms stereo allocations = %v, want 0", allocs)
	}
}

func fillMsStereoBase(
	specL []FixpDBL,
	specR []FixpDBL,
	enL []FixpDBL,
	enR []FixpDBL,
	enM []FixpDBL,
	enS []FixpDBL,
	thrL []FixpDBL,
	thrR []FixpDBL,
	spreadL []FixpDBL,
	spreadR []FixpDBL,
	enLLd []FixpDBL,
	enRLd []FixpDBL,
	enMLd []FixpDBL,
	enSLd []FixpDBL,
	thrLLd []FixpDBL,
	thrRLd []FixpDBL,
) {
	for i := range specL {
		specL[i] = FixpDBL((i+1)*1000 - 5000)
	}
	for i := range specR {
		specR[i] = FixpDBL(7000 - (i+1)*333)
	}
	for i := range enL {
		thrL[i] = FixpDBL((i + 2) * 0x01000000)
		thrR[i] = FixpDBL((i + 3) * 0x00800000)
		enL[i] = FixpDBL((i + 5) * 0x00200000)
		enR[i] = FixpDBL((i + 7) * 0x00180000)
		enM[i] = FixpDBL((i + 11) * 0x00100000)
		enS[i] = FixpDBL((i + 13) * 0x000c0000)
		spreadL[i] = FixpDBL((i + 3) * 0x00400000)
		spreadR[i] = FixpDBL((i + 4) * 0x00300000)
		thrLLd[i] = FixpDBL(1000 + i*100)
		thrRLd[i] = FixpDBL(1200 + i*100)
		enLLd[i] = FixpDBL(10000 + i*100)
		enRLd[i] = FixpDBL(11000 + i*100)
		enMLd[i] = FixpDBL(800 + i*10)
		enSLd[i] = FixpDBL(900 + i*10)
	}
}

func assertIntSlice(t *testing.T, name string, got []int, want []int, wantHash uint64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %d, want %d; got %v want %v", name, i, got[i], want[i], got, want)
		}
	}
	if h := hashBandEnergyInts(got); h != wantHash {
		t.Fatalf("%s hash = %#016x, want %#016x", name, h, wantHash)
	}
}
