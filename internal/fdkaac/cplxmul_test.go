package fdkaac

import "testing"

var cplxSink FixpDBL

func TestCplxMultDDVectors(t *testing.T) {
	// Expected values were checked against pinned FDK-AAC v2.0.3 on arm64.
	tests := []struct {
		aRe, aIm, bRe, bIm FixpDBL
		divRe, divIm       FixpDBL
		subRe, subIm       FixpDBL
		mulRe, mulIm       FixpDBL
	}{
		{0, 0, 0, 0, 0, 0, 305419896, -19088743, 0, 0},
		{1, -1, 1, -1, 0, -1, 305419896, -19088741, 0, -2},
		{0x40000000, 0x20000000, 0x40000000, -0x20000000, 335544320, 0, -30124424, -19088743, 671088640, 0},
		{MaxValDBL, MinValDBL, MaxValDBL, MinValDBL, -1, -2147483647, 305419897, 2128394905, -2, 2},
		{0x12345678, -0x1234567, 0x76543210, -0x2345678, 141007487, -11453246, 164412408, -7635496, 282014976, -22906493},
		{-0x40000000, 0x40000000, 0x20000000, 0x60000000, -536870912, -268435456, 842290808, 249346713, -1073741824, -536870912},
	}
	for _, tt := range tests {
		re, im := CplxMultDiv2DD(tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if re != tt.divRe || im != tt.divIm {
			t.Fatalf("CplxMultDiv2DD(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.divRe, tt.divIm)
		}
		re, im = CplxMultDiv2DPK(tt.aRe, tt.aIm, FixpDPK{Re: tt.bRe, Im: tt.bIm})
		if re != tt.divRe || im != tt.divIm {
			t.Fatalf("CplxMultDiv2DPK(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.divRe, tt.divIm)
		}
		re, im = CplxMultSubDiv2DD(0x12345678, -0x1234567, tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if re != tt.subRe || im != tt.subIm {
			t.Fatalf("CplxMultSubDiv2DD(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.subRe, tt.subIm)
		}
		re, im = CplxMultDD(tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if re != tt.mulRe || im != tt.mulIm {
			t.Fatalf("CplxMultDD(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.mulRe, tt.mulIm)
		}
		re, im = CplxMultDPK(tt.aRe, tt.aIm, FixpDPK{Re: tt.bRe, Im: tt.bIm})
		if re != tt.mulRe || im != tt.mulIm {
			t.Fatalf("CplxMultDPK(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.mulRe, tt.mulIm)
		}
	}
}

func TestCplxMultDSVectors(t *testing.T) {
	tests := []struct {
		aRe, aIm     FixpDBL
		bRe, bIm     FixpSGL
		divRe, divIm FixpDBL
		subRe, subIm FixpDBL
		mulRe, mulIm FixpDBL
	}{
		{0, 0, 0, 0, 0, 0, 305419896, -19088743, 0, 0},
		{1, -1, 1, -1, 0, -1, 305419896, -19088741, 0, -4},
		{0x40000000, 0x20000000, 0x4000, -0x2000, 335544320, 0, -30124424, -19088743, 671088640, 0},
		{MaxValDBL, MinValDBL, MaxValSGL, MinValSGL, -32769, -2147450880, 305452665, 2128362137, -65538, 65536},
		{0x12345678, -0x1234567, 12345, -23456, 50699814, -112908625, 254720081, 93819882, 101399630, -225817250},
		{-0x40000000, 0x40000000, 0x2000, 0x6000, -536870912, -268435456, 842290808, 249346713, -1073741824, -536870912},
	}
	for _, tt := range tests {
		re, im := CplxMultDiv2DS(tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if re != tt.divRe || im != tt.divIm {
			t.Fatalf("CplxMultDiv2DS(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.divRe, tt.divIm)
		}
		re, im = CplxMultDiv2SPK(tt.aRe, tt.aIm, FixpSPK{Re: tt.bRe, Im: tt.bIm})
		if re != tt.divRe || im != tt.divIm {
			t.Fatalf("CplxMultDiv2SPK(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.divRe, tt.divIm)
		}
		re, im = CplxMultSubDiv2DS(0x12345678, -0x1234567, tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if re != tt.subRe || im != tt.subIm {
			t.Fatalf("CplxMultSubDiv2DS(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.subRe, tt.subIm)
		}
		re, im = CplxMultDS(tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if re != tt.mulRe || im != tt.mulIm {
			t.Fatalf("CplxMultDS(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.mulRe, tt.mulIm)
		}
		re, im = CplxMultSPK(tt.aRe, tt.aIm, FixpSPK{Re: tt.bRe, Im: tt.bIm})
		if re != tt.mulRe || im != tt.mulIm {
			t.Fatalf("CplxMultSPK(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.mulRe, tt.mulIm)
		}
	}
}

func TestCplxMultSSVectors(t *testing.T) {
	tests := []struct {
		aRe, aIm, bRe, bIm FixpSGL
		divRe, divIm       FixpDBL
		mulRe, mulIm       FixpDBL
		divSRe, divSIm     FixpSGL
		mulSRe, mulSIm     FixpSGL
	}{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, -1, 1, -1, 0, -2, 0, -4, 0, -1, 0, -1},
		{0x4000, 0x2000, 0x4000, -0x2000, 335544320, 0, 10240, 0, 5120, 0, 10240, 0},
		{MaxValSGL, MinValSGL, MaxValSGL, MinValSGL, -65535, -2147418112, -2, -65536, -1, -32767, -2, 2},
		{12345, -23456, 23456, -12345, 0, -702582961, 0, -21444, 0, -10721, 0, -21442},
		{MinValSGL, 0x4000, 0x2000, 0x6000, -671088640, -671088640, -20480, -20480, -10240, -10240, -20480, -20480},
	}
	for _, tt := range tests {
		re, im := CplxMultDiv2SS(tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if re != tt.divRe || im != tt.divIm {
			t.Fatalf("CplxMultDiv2SS(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.divRe, tt.divIm)
		}
		re, im = CplxMultDiv2SSPK(tt.aRe, tt.aIm, FixpSPK{Re: tt.bRe, Im: tt.bIm})
		if re != tt.divRe || im != tt.divIm {
			t.Fatalf("CplxMultDiv2SSPK(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.divRe, tt.divIm)
		}
		sre, sim := CplxMultDiv2SSSGL(tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if sre != tt.divSRe || sim != tt.divSIm {
			t.Fatalf("CplxMultDiv2SSSGL(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, sre, sim, tt.divSRe, tt.divSIm)
		}
		sre, sim = CplxMultDiv2SSPKSGL(tt.aRe, tt.aIm, FixpSPK{Re: tt.bRe, Im: tt.bIm})
		if sre != tt.divSRe || sim != tt.divSIm {
			t.Fatalf("CplxMultDiv2SSPKSGL(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, sre, sim, tt.divSRe, tt.divSIm)
		}
		re, im = CplxMultDS(FixpDBL(tt.aRe), FixpDBL(tt.aIm), tt.bRe, tt.bIm)
		if re != tt.mulRe || im != tt.mulIm {
			t.Fatalf("CplxMultDS(promoted %d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, re, im, tt.mulRe, tt.mulIm)
		}
		sre, sim = CplxMultSS(tt.aRe, tt.aIm, tt.bRe, tt.bIm)
		if sre != tt.mulSRe || sim != tt.mulSIm {
			t.Fatalf("CplxMultSS(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tt.aRe, tt.aIm, tt.bRe, tt.bIm, sre, sim, tt.mulSRe, tt.mulSIm)
		}
	}
}

func TestCplxMultAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		re, im := CplxMultDiv2DD(0x12345678, -0x1234567, 0x76543210, -0x2345678)
		re, im = CplxMultSubDiv2DS(re, im, re, im, 12345, -23456)
		re, im = CplxMultSPK(re, im, FixpSPK{Re: 0x4000, Im: -0x2000})
		cplxSink = re ^ im
	})
	if allocs != 0 {
		t.Fatalf("complex multiply helpers allocations = %v, want 0", allocs)
	}
}
