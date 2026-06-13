package fdkaac

import "testing"

var intensityHashSink uint64

func TestFDKaacEncIntensityStereoProcessingVector(t *testing.T) {
	var c intensityStereoCase
	prepareIntensityStereoCase(&c)

	c.run(1)

	if c.msDigest != MsMaskSome {
		t.Fatalf("ms digest = %d, want %d", c.msDigest, MsMaskSome)
	}
	for sfb := 0; sfb < intensityVectorSfbCnt; sfb++ {
		if c.isBook[sfb] != codeBookISInPhaseNo {
			t.Fatalf("isBook[%d] = %d, want %d", sfb, c.isBook[sfb], codeBookISInPhaseNo)
		}
		wantMS := 0
		if sfb >= intensityVectorSfbCnt-2 {
			wantMS = 1
		}
		if c.msMask[sfb] != wantMS {
			t.Fatalf("msMask[%d] = %d, want %d", sfb, c.msMask[sfb], wantMS)
		}
		if c.energyRight[sfb] != 0 || c.energyLdRight[sfb] != ldDataMinusOne ||
			c.thresholdRight[sfb] != 0 || c.thresholdLdRight[sfb] != psyPostMinThresholdLdData ||
			c.spreadRight[sfb] != 0 {
			t.Fatalf("right sfb %d not suppressed by intensity stereo", sfb)
		}
		if c.pns[0].PNSFlag[sfb] != 0 || c.pns[1].PNSFlag[sfb] != 0 {
			t.Fatalf("PNS flags[%d] = %d/%d, want 0/0", sfb, c.pns[0].PNSFlag[sfb], c.pns[1].PNSFlag[sfb])
		}
		for line := c.offsets[sfb]; line < c.offsets[sfb+1]; line++ {
			if c.specRight[line] != 0 {
				t.Fatalf("right spectrum[%d] = %#x, want 0", line, c.specRight[line])
			}
		}
	}
	if got, want := hashIntensityStereoCase(&c), uint64(0x858e6116927f45b5); got != want {
		t.Fatalf("intensity stereo hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncIntensityStereoDisabledClearsScalesOnly(t *testing.T) {
	var c intensityStereoCase
	prepareIntensityStereoCase(&c)
	before := hashIntensityStereoCase(&c)

	c.run(0)

	if c.msDigest != MsMaskAll {
		t.Fatalf("disabled ms digest = %d, want %d", c.msDigest, MsMaskAll)
	}
	for sfb := 0; sfb < intensityVectorSfbCnt; sfb++ {
		if c.isBook[sfb] != 0 || c.isScale[sfb] != 0 {
			t.Fatalf("disabled is state[%d] = %d/%d, want 0/0", sfb, c.isBook[sfb], c.isScale[sfb])
		}
		if c.msMask[sfb] != 7 || c.pns[0].PNSFlag[sfb] != 1 || c.pns[1].PNSFlag[sfb] != 1 {
			t.Fatalf("disabled control state[%d] changed", sfb)
		}
	}
	after := hashIntensityStereoCase(&c)
	if before == after {
		t.Fatalf("disabled hash did not observe cleared IS scales")
	}
}

func TestFDKaacEncIntensityStereoRejectsInvalidControls(t *testing.T) {
	var c intensityStereoCase
	prepareIntensityStereoCase(&c)
	tests := []struct {
		name string
		run  func()
	}{
		{"nil digest", func() {
			FDKaacEncIntensityStereoProcessing(c.energyLeft[:], c.energyRight[:], c.specLeft[:], c.specRight[:], c.thresholdLeft[:], c.thresholdRight[:], c.thresholdLdRight[:], c.spreadLeft[:], c.spreadRight[:], c.energyLdLeft[:], c.energyLdRight[:], nil, c.msMask[:], intensityVectorSfbCnt, intensityVectorSfbPerGroup, intensityVectorMaxSfbPerGroup, c.offsets[:], 1, c.isBook[:], c.isScale[:], c.pnsPtrs[:])
		}},
		{"zero sfb count", func() {
			FDKaacEncIntensityStereoProcessing(c.energyLeft[:], c.energyRight[:], c.specLeft[:], c.specRight[:], c.thresholdLeft[:], c.thresholdRight[:], c.thresholdLdRight[:], c.spreadLeft[:], c.spreadRight[:], c.energyLdLeft[:], c.energyLdRight[:], &c.msDigest, c.msMask[:], 0, intensityVectorSfbPerGroup, intensityVectorMaxSfbPerGroup, c.offsets[:], 1, c.isBook[:], c.isScale[:], c.pnsPtrs[:])
		}},
		{"bad group width", func() {
			FDKaacEncIntensityStereoProcessing(c.energyLeft[:], c.energyRight[:], c.specLeft[:], c.specRight[:], c.thresholdLeft[:], c.thresholdRight[:], c.thresholdLdRight[:], c.spreadLeft[:], c.spreadRight[:], c.energyLdLeft[:], c.energyLdRight[:], &c.msDigest, c.msMask[:], intensityVectorSfbCnt, intensityVectorSfbPerGroup, intensityVectorSfbPerGroup+1, c.offsets[:], 1, c.isBook[:], c.isScale[:], c.pnsPtrs[:])
		}},
		{"short offsets", func() {
			FDKaacEncIntensityStereoProcessing(c.energyLeft[:], c.energyRight[:], c.specLeft[:], c.specRight[:], c.thresholdLeft[:], c.thresholdRight[:], c.thresholdLdRight[:], c.spreadLeft[:], c.spreadRight[:], c.energyLdLeft[:], c.energyLdRight[:], &c.msDigest, c.msMask[:], intensityVectorSfbCnt, intensityVectorSfbPerGroup, intensityVectorMaxSfbPerGroup, c.offsets[:intensityVectorSfbCnt], 1, c.isBook[:], c.isScale[:], c.pnsPtrs[:])
		}},
		{"decreasing offsets", func() {
			badOffsets := c.offsets
			badOffsets[3] = badOffsets[2] - 1
			FDKaacEncIntensityStereoProcessing(c.energyLeft[:], c.energyRight[:], c.specLeft[:], c.specRight[:], c.thresholdLeft[:], c.thresholdRight[:], c.thresholdLdRight[:], c.spreadLeft[:], c.spreadRight[:], c.energyLdLeft[:], c.energyLdRight[:], &c.msDigest, c.msMask[:], intensityVectorSfbCnt, intensityVectorSfbPerGroup, intensityVectorMaxSfbPerGroup, badOffsets[:], 1, c.isBook[:], c.isScale[:], c.pnsPtrs[:])
		}},
		{"short spectrum", func() {
			FDKaacEncIntensityStereoProcessing(c.energyLeft[:], c.energyRight[:], c.specLeft[:c.offsets[intensityVectorSfbCnt]-1], c.specRight[:], c.thresholdLeft[:], c.thresholdRight[:], c.thresholdLdRight[:], c.spreadLeft[:], c.spreadRight[:], c.energyLdLeft[:], c.energyLdRight[:], &c.msDigest, c.msMask[:], intensityVectorSfbCnt, intensityVectorSfbPerGroup, intensityVectorMaxSfbPerGroup, c.offsets[:], 1, c.isBook[:], c.isScale[:], c.pnsPtrs[:])
		}},
		{"short vectors", func() {
			FDKaacEncIntensityStereoProcessing(c.energyLeft[:intensityVectorSfbCnt-1], c.energyRight[:], c.specLeft[:], c.specRight[:], c.thresholdLeft[:], c.thresholdRight[:], c.thresholdLdRight[:], c.spreadLeft[:], c.spreadRight[:], c.energyLdLeft[:], c.energyLdRight[:], &c.msDigest, c.msMask[:], intensityVectorSfbCnt, intensityVectorSfbPerGroup, intensityVectorMaxSfbPerGroup, c.offsets[:], 1, c.isBook[:], c.isScale[:], c.pnsPtrs[:])
		}},
		{"short pns pair", func() {
			FDKaacEncIntensityStereoProcessing(c.energyLeft[:], c.energyRight[:], c.specLeft[:], c.specRight[:], c.thresholdLeft[:], c.thresholdRight[:], c.thresholdLdRight[:], c.spreadLeft[:], c.spreadRight[:], c.energyLdLeft[:], c.energyLdRight[:], &c.msDigest, c.msMask[:], intensityVectorSfbCnt, intensityVectorSfbPerGroup, intensityVectorMaxSfbPerGroup, c.offsets[:], 1, c.isBook[:], c.isScale[:], c.pnsPtrs[:1])
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectAACEncPanic(t, tt.run)
		})
	}
}

func TestFDKaacEncIntensityStereoAllocs(t *testing.T) {
	var c intensityStereoCase
	allocs := testing.AllocsPerRun(1000, func() {
		prepareIntensityStereoCase(&c)
		c.run(1)
		intensityHashSink ^= hashIntensityStereoCase(&c)
	})
	if allocs != 0 {
		t.Fatalf("intensity stereo allocations = %v, want 0", allocs)
	}
}

const (
	intensityVectorSfbCnt          = 8
	intensityVectorSfbPerGroup     = 8
	intensityVectorMaxSfbPerGroup  = 8
	intensityVectorLinesPerBand    = 8
	intensityVectorFirstOutOfPhase = 6
)

type intensityStereoCase struct {
	specLeft         [maxSpectralLines]FixpDBL
	specRight        [maxSpectralLines]FixpDBL
	offsets          [maxGroupedSFB + 1]int
	scaleLeft        [maxGroupedSFB]int
	scaleRight       [maxGroupedSFB]int
	energyLeft       [maxGroupedSFB]FixpDBL
	energyRight      [maxGroupedSFB]FixpDBL
	energyLdLeft     [maxGroupedSFB]FixpDBL
	energyLdRight    [maxGroupedSFB]FixpDBL
	thresholdLeft    [maxGroupedSFB]FixpDBL
	thresholdRight   [maxGroupedSFB]FixpDBL
	thresholdLdRight [maxGroupedSFB]FixpDBL
	spreadLeft       [maxGroupedSFB]FixpDBL
	spreadRight      [maxGroupedSFB]FixpDBL
	msDigest         int
	msMask           [maxGroupedSFB]int
	isBook           [maxGroupedSFB]int
	isScale          [maxGroupedSFB]int
	pns              [2]PNSData
	pnsPtrs          [2]*PNSData
}

func prepareIntensityStereoCase(c *intensityStereoCase) {
	*c = intensityStereoCase{}
	for sfb := 0; sfb <= intensityVectorSfbCnt; sfb++ {
		c.offsets[sfb] = sfb * intensityVectorLinesPerBand
	}
	for sfb := 0; sfb < intensityVectorSfbCnt; sfb++ {
		for line := c.offsets[sfb]; line < c.offsets[sfb+1]; line++ {
			n := line - c.offsets[sfb]
			v := FixpDBL(0x01800000 + ((sfb + 3) * (n + 5) * 0x00012000))
			if (n+sfb)&1 != 0 {
				v = -v
			}
			c.specLeft[line] = v
			c.specRight[line] = v >> 2
			if sfb >= intensityVectorFirstOutOfPhase {
				c.specRight[line] = -c.specRight[line]
			}
		}
	}

	FDKaacEncCalcSfbMaxScaleSpec(c.specLeft[:], c.offsets[:], c.scaleLeft[:], intensityVectorSfbCnt)
	FDKaacEncCalcSfbMaxScaleSpec(c.specRight[:], c.offsets[:], c.scaleRight[:], intensityVectorSfbCnt)
	FDKaacEncCalcBandEnergyOptimLong(c.specLeft[:], c.scaleLeft[:], c.offsets[:], intensityVectorSfbCnt, c.energyLeft[:], c.energyLdLeft[:])
	FDKaacEncCalcBandEnergyOptimLong(c.specRight[:], c.scaleRight[:], c.offsets[:], intensityVectorSfbCnt, c.energyRight[:], c.energyLdRight[:])

	for sfb := 0; sfb < intensityVectorSfbCnt; sfb++ {
		c.thresholdLeft[sfb] = FMultDD(c.energyLeft[sfb], cRatio)
		c.thresholdRight[sfb] = FMultDD(c.energyRight[sfb], cRatio)
		c.thresholdLdRight[sfb] = maxFixpDBL(CalcLdData(c.thresholdRight[sfb]), psyPostMinThresholdLdData)
		c.spreadLeft[sfb] = c.energyLeft[sfb]
		c.spreadRight[sfb] = c.energyRight[sfb]
		c.msMask[sfb] = 7
		c.isBook[sfb] = -11
		c.isScale[sfb] = 23
		c.pns[0].PNSFlag[sfb] = 1
		c.pns[1].PNSFlag[sfb] = 1
	}
	c.msDigest = MsMaskAll
	c.pnsPtrs[0] = &c.pns[0]
	c.pnsPtrs[1] = &c.pns[1]
}

func (c *intensityStereoCase) run(allowIS int) {
	FDKaacEncIntensityStereoProcessing(
		c.energyLeft[:],
		c.energyRight[:],
		c.specLeft[:],
		c.specRight[:],
		c.thresholdLeft[:],
		c.thresholdRight[:],
		c.thresholdLdRight[:],
		c.spreadLeft[:],
		c.spreadRight[:],
		c.energyLdLeft[:],
		c.energyLdRight[:],
		&c.msDigest,
		c.msMask[:],
		intensityVectorSfbCnt,
		intensityVectorSfbPerGroup,
		intensityVectorMaxSfbPerGroup,
		c.offsets[:],
		allowIS,
		c.isBook[:],
		c.isScale[:],
		c.pnsPtrs[:],
	)
}

func hashIntensityStereoCase(c *intensityStereoCase) uint64 {
	lines := c.offsets[intensityVectorSfbCnt]
	h := uint64(14695981039346656037)
	h = hashTnsSyncInt(h, c.msDigest)
	h = mixPsyTnsHash(h, hashFixpDBL(c.specLeft[:lines]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.specRight[:lines]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.energyLeft[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.energyRight[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.energyLdLeft[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.energyLdRight[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.thresholdLeft[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.thresholdRight[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.thresholdLdRight[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.spreadLeft[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashFixpDBL(c.spreadRight[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(c.msMask[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(c.isBook[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(c.isScale[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(c.pns[0].PNSFlag[:intensityVectorSfbCnt]))
	h = mixPsyTnsHash(h, hashPsyPostIntSlice(c.pns[1].PNSFlag[:intensityVectorSfbCnt]))
	return h
}
