package fdkaac

import "testing"

func TestFixMAddDDVectors(t *testing.T) {
	// Expected values were checked against pinned FDK-AAC v2.0.3 on arm64.
	tests := []struct {
		x, a, b                          FixpDBL
		addDiv2, subDiv2, add, sub       FixpDBL
		addDiv2BitExact, subDiv2BitExact FixpDBL
	}{
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1, 2, 2, 1, 1},
		{-1, -1, 1, -2, 0, -4, 0, -2, 0},
		{0x10000000, 0x40000000, 0x40000000, 536870912, 0, 1073741824, 0, 536870912, 0},
		{0x70000000, MaxValDBL, MaxValDBL, -1342177281, 805306369, 1610612734, 1610612738, -1342177281, 805306369},
		{MinValDBL, MinValDBL, MinValDBL, -1073741824, 1073741824, MinValDBL, MinValDBL, -1073741824, 1073741824},
		{0x12345678, 0x12345678, -0x01234567, 304062474, 306777318, 608124948, 613554636, 304062474, 306777318},
		{-0x1234567, -0x1234567, 0x76543210, -27911985, -10265501, -55823970, -20531002, -27911985, -10265501},
		{MaxValDBL, MaxValDBL, MinValDBL, 1073741823, -1073741825, 2147483646, 2147483646, 1073741823, -1073741825},
		{MinValDBL, MinValDBL, MaxValDBL, 1073741824, -1073741824, MinValDBL, MinValDBL, 1073741824, -1073741824},
	}
	for _, tt := range tests {
		if got := FixMAddDiv2DD(tt.x, tt.a, tt.b); got != tt.addDiv2 {
			t.Fatalf("FixMAddDiv2DD(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.addDiv2)
		}
		if got := FixMSubDiv2DD(tt.x, tt.a, tt.b); got != tt.subDiv2 {
			t.Fatalf("FixMSubDiv2DD(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.subDiv2)
		}
		if got := FixMAddDD(tt.x, tt.a, tt.b); got != tt.add {
			t.Fatalf("FixMAddDD(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.add)
		}
		if got := FixMSubDD(tt.x, tt.a, tt.b); got != tt.sub {
			t.Fatalf("FixMSubDD(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.sub)
		}
		if got := FixMAddDiv2BitExactDD(tt.x, tt.a, tt.b); got != tt.addDiv2BitExact {
			t.Fatalf("FixMAddDiv2BitExactDD(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.addDiv2BitExact)
		}
		if got := FixMSubDiv2BitExactDD(tt.x, tt.a, tt.b); got != tt.subDiv2BitExact {
			t.Fatalf("FixMSubDiv2BitExactDD(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.subDiv2BitExact)
		}
	}
}

func TestFixMAddMixedVectors(t *testing.T) {
	tests := []struct {
		x                                FixpDBL
		sgl                              FixpSGL
		dbl                              FixpDBL
		addDiv2, subDiv2, add, sub       FixpDBL
		addDiv2BitExact, subDiv2BitExact FixpDBL
	}{
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1, 2, 2, 1, 1},
		{-1, -1, -1, -1, -1, -2, -2, -1, -1},
		{0x10000000, 0x4000, 0x40000000, 536870912, 0, 1073741824, 0, 536870912, 0},
		{0x70000000, MaxValSGL, MaxValDBL, -1342210049, 805339137, 1610547198, 1610678274, -1342210049, 805339137},
		{MinValDBL, MinValSGL, MinValDBL, -1073741824, 1073741824, MinValDBL, MinValDBL, -1073741824, 1073741824},
		{0x12345678, 12345, 0x12345678, 362951765, 247888027, 725903530, 495776054, 362951765, 247888027},
		{-0x1234567, -23456, -0x7654321, 25319615, -63497101, 50639230, -126994202, 25319615, -63497101},
	}
	for _, tt := range tests {
		if got := FixMAddDiv2SD(tt.x, tt.sgl, tt.dbl); got != tt.addDiv2 {
			t.Fatalf("FixMAddDiv2SD(%d,%d,%d) = %d, want %d", tt.x, tt.sgl, tt.dbl, got, tt.addDiv2)
		}
		if got := FixMAddDiv2DS(tt.x, tt.dbl, tt.sgl); got != tt.addDiv2 {
			t.Fatalf("FixMAddDiv2DS(%d,%d,%d) = %d, want %d", tt.x, tt.dbl, tt.sgl, got, tt.addDiv2)
		}
		if got := FixMSubDiv2SD(tt.x, tt.sgl, tt.dbl); got != tt.subDiv2 {
			t.Fatalf("FixMSubDiv2SD(%d,%d,%d) = %d, want %d", tt.x, tt.sgl, tt.dbl, got, tt.subDiv2)
		}
		if got := FixMSubDiv2DS(tt.x, tt.dbl, tt.sgl); got != tt.subDiv2 {
			t.Fatalf("FixMSubDiv2DS(%d,%d,%d) = %d, want %d", tt.x, tt.dbl, tt.sgl, got, tt.subDiv2)
		}
		if got := FixMAddSD(tt.x, tt.sgl, tt.dbl); got != tt.add {
			t.Fatalf("FixMAddSD(%d,%d,%d) = %d, want %d", tt.x, tt.sgl, tt.dbl, got, tt.add)
		}
		if got := FixMAddDS(tt.x, tt.dbl, tt.sgl); got != tt.add {
			t.Fatalf("FixMAddDS(%d,%d,%d) = %d, want %d", tt.x, tt.dbl, tt.sgl, got, tt.add)
		}
		if got := FixMSubSD(tt.x, tt.sgl, tt.dbl); got != tt.sub {
			t.Fatalf("FixMSubSD(%d,%d,%d) = %d, want %d", tt.x, tt.sgl, tt.dbl, got, tt.sub)
		}
		if got := FixMSubDS(tt.x, tt.dbl, tt.sgl); got != tt.sub {
			t.Fatalf("FixMSubDS(%d,%d,%d) = %d, want %d", tt.x, tt.dbl, tt.sgl, got, tt.sub)
		}
		if got := FixMAddDiv2BitExactSD(tt.x, tt.sgl, tt.dbl); got != tt.addDiv2BitExact {
			t.Fatalf("FixMAddDiv2BitExactSD(%d,%d,%d) = %d, want %d", tt.x, tt.sgl, tt.dbl, got, tt.addDiv2BitExact)
		}
		if got := FixMAddDiv2BitExactDS(tt.x, tt.dbl, tt.sgl); got != tt.addDiv2BitExact {
			t.Fatalf("FixMAddDiv2BitExactDS(%d,%d,%d) = %d, want %d", tt.x, tt.dbl, tt.sgl, got, tt.addDiv2BitExact)
		}
		if got := FixMSubDiv2BitExactSD(tt.x, tt.sgl, tt.dbl); got != tt.subDiv2BitExact {
			t.Fatalf("FixMSubDiv2BitExactSD(%d,%d,%d) = %d, want %d", tt.x, tt.sgl, tt.dbl, got, tt.subDiv2BitExact)
		}
		if got := FixMSubDiv2BitExactDS(tt.x, tt.dbl, tt.sgl); got != tt.subDiv2BitExact {
			t.Fatalf("FixMSubDiv2BitExactDS(%d,%d,%d) = %d, want %d", tt.x, tt.dbl, tt.sgl, got, tt.subDiv2BitExact)
		}
	}
}

func TestFixMAddSSVectors(t *testing.T) {
	tests := []struct {
		x                          FixpDBL
		a, b                       FixpSGL
		addDiv2, subDiv2, add, sub FixpDBL
	}{
		{0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 2, 0, 4, 0},
		{-1, -1, 1, -2, 0, -4, 0},
		{0x10000000, 0x4000, 0x4000, 536870912, 0, 1073741824, 0},
		{0x70000000, MaxValSGL, MaxValSGL, -1342242815, 805371903, 1610481666, 1610743806},
		{MinValDBL, MinValSGL, MinValSGL, -1073741824, 1073741824, MinValDBL, MinValDBL},
		{0x12345678, 12345, -23456, 15855576, 594984216, 31711152, 1189968432},
		{-0x1234567, MinValSGL, MaxValSGL, -1092797799, 1054620313, 2109371698, 2109240626},
	}
	for _, tt := range tests {
		if got := FixMAddDiv2SS(tt.x, tt.a, tt.b); got != tt.addDiv2 {
			t.Fatalf("FixMAddDiv2SS(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.addDiv2)
		}
		if got := FixMSubDiv2SS(tt.x, tt.a, tt.b); got != tt.subDiv2 {
			t.Fatalf("FixMSubDiv2SS(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.subDiv2)
		}
		if got := FixMAddSS(tt.x, tt.a, tt.b); got != tt.add {
			t.Fatalf("FixMAddSS(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.add)
		}
		if got := FixMSubSS(tt.x, tt.a, tt.b); got != tt.sub {
			t.Fatalf("FixMSubSS(%d,%d,%d) = %d, want %d", tt.x, tt.a, tt.b, got, tt.sub)
		}
	}
}

func TestFixPAddVectors(t *testing.T) {
	tests := []struct {
		x           FixpDBL
		dbl         FixpDBL
		sgl         FixpSGL
		dDiv2, dAdd FixpDBL
		sDiv2, sAdd FixpDBL
	}{
		{0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1, 2, 3},
		{-1, -1, -1, -1, -1, 0, 1},
		{0x10000000, 0x40000000, 0x4000, 536870912, 805306368, 536870912, 805306368},
		{0x70000000, MaxValDBL, MaxValSGL, -1342177281, -268435458, -1342242815, -268566526},
		{MinValDBL, MinValDBL, MinValSGL, -1073741824, 0, -1073741824, -1},
		{0x12345678, 0x12345678, 12345, 327138644, 348857392, 457818921, 610217946},
		{-0x1234567, -0x1234567, -23456, -19003905, -18919067, 531095193, 1081279129},
	}
	for _, tt := range tests {
		if got := FixPAddDiv2D(tt.x, tt.dbl); got != tt.dDiv2 {
			t.Fatalf("FixPAddDiv2D(%d,%d) = %d, want %d", tt.x, tt.dbl, got, tt.dDiv2)
		}
		if got := FixPAddD(tt.x, tt.dbl); got != tt.dAdd {
			t.Fatalf("FixPAddD(%d,%d) = %d, want %d", tt.x, tt.dbl, got, tt.dAdd)
		}
		if got := FixPAddDiv2S(tt.x, tt.sgl); got != tt.sDiv2 {
			t.Fatalf("FixPAddDiv2S(%d,%d) = %d, want %d", tt.x, tt.sgl, got, tt.sDiv2)
		}
		if got := FixPAddS(tt.x, tt.sgl); got != tt.sAdd {
			t.Fatalf("FixPAddS(%d,%d) = %d, want %d", tt.x, tt.sgl, got, tt.sAdd)
		}
	}
}

func TestFixMAddAllocs(t *testing.T) {
	var x FixpDBL = 0x12345678
	var y FixpDBL = -0x1234567
	var s FixpSGL = -23456
	allocs := testing.AllocsPerRun(1000, func() {
		x = FMultAddDiv2DD(x, y, x)
		y = FMultSubDS(y, x, s)
		x = FPow2AddS(x, s)
		fixpointSink = x ^ y
	})
	if allocs != 0 {
		t.Fatalf("fixed-point multiply-add helpers allocations = %v, want 0", allocs)
	}
}
