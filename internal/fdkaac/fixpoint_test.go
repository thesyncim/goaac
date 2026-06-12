package fdkaac

import "testing"

var fixpointSink FixpDBL

func TestFixpointConstantsAndConversions(t *testing.T) {
	if FractBits != 16 || DfractBits != 32 || AccuBits != 40 {
		t.Fatalf("fixed-point bit widths = %d/%d/%d, want 16/32/40", FractBits, DfractBits, AccuBits)
	}
	if MaxValSGL != 32767 || MinValSGL != -32768 || MaxValDBL != 2147483647 || MinValDBL != -2147483648 {
		t.Fatalf("fixed-point limits = %d/%d/%d/%d", MaxValSGL, MinValSGL, MaxValDBL, MinValDBL)
	}

	tests := []struct {
		name string
		sgl  FixpSGL
		dbl  FixpDBL
	}{
		{name: "zero", sgl: 0, dbl: 0},
		{name: "one", sgl: 1, dbl: 0x00010000},
		{name: "minus one", sgl: -1, dbl: -0x00010000},
		{name: "max", sgl: MaxValSGL, dbl: 0x7fff0000},
		{name: "min", sgl: MinValSGL, dbl: MinValDBL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FXSGL2FXDBL(tt.sgl); got != tt.dbl {
				t.Fatalf("FXSGL2FXDBL(%d) = %d, want %d", tt.sgl, got, tt.dbl)
			}
			if got := FXDBL2FXSGL(tt.dbl); got != tt.sgl {
				t.Fatalf("FXDBL2FXSGL(%d) = %d, want %d", tt.dbl, got, tt.sgl)
			}
		})
	}
}

func TestFixpointConstConversionVectors(t *testing.T) {
	tests := []struct {
		in   uint32
		want FixpSGL
	}{
		{0x00000000, 0},
		{0x00008000, 1},
		{0x0000ffff, 1},
		{0x40000000, 0x4000},
		{0x5a82799a, 0x5a82},
		{0x7fffffff, MaxValSGL},
		{0x80000000, MinValSGL},
		{0xc0000000, -0x4000},
	}
	for _, tt := range tests {
		if got := STC(tt.in); got != tt.want {
			t.Fatalf("STC(0x%08x) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestFixMulDDVectors(t *testing.T) {
	// Expected values were checked against pinned FDK-AAC v2.0.3 on arm64.
	tests := []struct {
		a, b                 FixpDBL
		div2, mul            FixpDBL
		div2Bit, mulBitExact FixpDBL
	}{
		{0, 0, 0, 0, 0, 0},
		{1, 1, 0, 0, 0, 0},
		{-1, 1, -1, -1, -1, -2},
		{0x40000000, 0x40000000, 268435456, 536870912, 268435456, 536870912},
		{MaxValDBL, MaxValDBL, 1073741823, 2147483646, 1073741823, 2147483646},
		{MinValDBL, MinValDBL, 1073741824, MinValDBL, 1073741824, MinValDBL},
		{0x12345678, -0x01234567, -1357422, -2714844, -1357422, -2714844},
		{-0x1234567, 0x76543210, -8823242, -17646483, -8823242, -17646484},
		{MaxValDBL, MinValDBL, -1073741824, -2147483647, -1073741824, MinValDBL},
		{MinValDBL, MaxValDBL, -1073741824, -2147483647, -1073741824, MinValDBL},
	}
	for _, tt := range tests {
		if got := FixMulDiv2DD(tt.a, tt.b); got != tt.div2 {
			t.Fatalf("FixMulDiv2DD(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.div2)
		}
		if got := FixMulDD(tt.a, tt.b); got != tt.mul {
			t.Fatalf("FixMulDD(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.mul)
		}
		if got := FixMulDiv2BitExactDD(tt.a, tt.b); got != tt.div2Bit {
			t.Fatalf("FixMulDiv2BitExactDD(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.div2Bit)
		}
		if got := FixMulBitExactDD(tt.a, tt.b); got != tt.mulBitExact {
			t.Fatalf("FixMulBitExactDD(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.mulBitExact)
		}
	}
}

func TestFixMulSSVectors(t *testing.T) {
	tests := []struct {
		a, b      FixpSGL
		div2, mul FixpDBL
		pow2Div2  FixpDBL
		pow2      FixpDBL
	}{
		{0, 0, 0, 0, 0, 0},
		{1, 1, 1, 2, 1, 2},
		{-1, 1, -1, -2, 1, 2},
		{0x4000, 0x4000, 268435456, 536870912, 268435456, 536870912},
		{MaxValSGL, MaxValSGL, 1073676289, 2147352578, 1073676289, 2147352578},
		{MinValSGL, MinValSGL, 1073741824, MinValDBL, 1073741824, MaxValDBL},
		{12345, -23456, -289564320, -579128640, 152399025, 304798050},
		{MinValSGL, MaxValSGL, -1073709056, -2147418112, 1073741824, MaxValDBL},
	}
	for _, tt := range tests {
		if got := FixMulDiv2SS(tt.a, tt.b); got != tt.div2 {
			t.Fatalf("FixMulDiv2SS(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.div2)
		}
		if got := FixMulSS(tt.a, tt.b); got != tt.mul {
			t.Fatalf("FixMulSS(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.mul)
		}
		if got := FixPow2Div2S(tt.a); got != tt.pow2Div2 {
			t.Fatalf("FixPow2Div2S(%d) = %d, want %d", tt.a, got, tt.pow2Div2)
		}
		if got := FixPow2S(tt.a); got != tt.pow2 {
			t.Fatalf("FixPow2S(%d) = %d, want %d", tt.a, got, tt.pow2)
		}
	}
}

func TestFixMulMixedVectors(t *testing.T) {
	tests := []struct {
		sgl                  FixpSGL
		dbl                  FixpDBL
		div2, mul            FixpDBL
		div2Bit, mulBitExact FixpDBL
	}{
		{0, 0, 0, 0, 0, 0},
		{1, 1, 0, 0, 0, 0},
		{-1, -1, 0, 0, 0, 0},
		{0x4000, 0x40000000, 268435456, 536870912, 268435456, 536870912},
		{MaxValSGL, MaxValDBL, 1073709055, 2147418110, 1073709055, 2147418110},
		{MinValSGL, MinValDBL, 1073741824, MinValDBL, 1073741824, MinValDBL},
		{12345, 0x12345678, 57531869, 115063738, 57531869, 115063738},
		{-23456, -0x7654321, 44408358, 88816716, 44408358, 88816716},
	}
	for _, tt := range tests {
		if got := FixMulDiv2SD(tt.sgl, tt.dbl); got != tt.div2 {
			t.Fatalf("FixMulDiv2SD(%d,%d) = %d, want %d", tt.sgl, tt.dbl, got, tt.div2)
		}
		if got := FixMulDiv2DS(tt.dbl, tt.sgl); got != tt.div2 {
			t.Fatalf("FixMulDiv2DS(%d,%d) = %d, want %d", tt.dbl, tt.sgl, got, tt.div2)
		}
		if got := FixMulSD(tt.sgl, tt.dbl); got != tt.mul {
			t.Fatalf("FixMulSD(%d,%d) = %d, want %d", tt.sgl, tt.dbl, got, tt.mul)
		}
		if got := FixMulDS(tt.dbl, tt.sgl); got != tt.mul {
			t.Fatalf("FixMulDS(%d,%d) = %d, want %d", tt.dbl, tt.sgl, got, tt.mul)
		}
		if got := FixMulDiv2BitExactSD(tt.sgl, tt.dbl); got != tt.div2Bit {
			t.Fatalf("FixMulDiv2BitExactSD(%d,%d) = %d, want %d", tt.sgl, tt.dbl, got, tt.div2Bit)
		}
		if got := FixMulDiv2BitExactDS(tt.dbl, tt.sgl); got != tt.div2Bit {
			t.Fatalf("FixMulDiv2BitExactDS(%d,%d) = %d, want %d", tt.dbl, tt.sgl, got, tt.div2Bit)
		}
		if got := FixMulBitExactSD(tt.sgl, tt.dbl); got != tt.mulBitExact {
			t.Fatalf("FixMulBitExactSD(%d,%d) = %d, want %d", tt.sgl, tt.dbl, got, tt.mulBitExact)
		}
		if got := FixMulBitExactDS(tt.dbl, tt.sgl); got != tt.mulBitExact {
			t.Fatalf("FixMulBitExactDS(%d,%d) = %d, want %d", tt.dbl, tt.sgl, got, tt.mulBitExact)
		}
	}
}

func TestFixPow2DVectors(t *testing.T) {
	tests := []struct {
		in        FixpDBL
		div2, pow FixpDBL
	}{
		{0, 0, 0},
		{1, 0, 0},
		{-1, 0, 0},
		{0x40000000, 268435456, 536870912},
		{MaxValDBL, 1073741823, 2147483646},
		{MinValDBL, 1073741824, MinValDBL},
		{0x12345678, 21718748, 43437496},
		{-0x1234567, 84838, 169676},
	}
	for _, tt := range tests {
		if got := FixPow2Div2D(tt.in); got != tt.div2 {
			t.Fatalf("FixPow2Div2D(%d) = %d, want %d", tt.in, got, tt.div2)
		}
		if got := FixPow2D(tt.in); got != tt.pow {
			t.Fatalf("FixPow2D(%d) = %d, want %d", tt.in, got, tt.pow)
		}
	}
}

func TestFixNormVectors(t *testing.T) {
	dbl := []struct {
		in                   FixpDBL
		normz, norm, leading int
	}{
		{0, 32, 0, 0},
		{1, 31, 30, 30},
		{-1, 0, 31, 31},
		{2, 30, 29, 29},
		{-2, 0, 30, 30},
		{0x40000000, 1, 0, 0},
		{MaxValDBL, 1, 0, 0},
		{MinValDBL, 0, 0, 0},
		{0x12345678, 3, 2, 2},
		{-0x1234567, 0, 6, 6},
	}
	for _, tt := range dbl {
		if got := FixNormZD(tt.in); got != tt.normz {
			t.Fatalf("FixNormZD(%d) = %d, want %d", tt.in, got, tt.normz)
		}
		if got := FixNormD(tt.in); got != tt.norm {
			t.Fatalf("FixNormD(%d) = %d, want %d", tt.in, got, tt.norm)
		}
		if got := CountLeadingBits(tt.in); got != tt.leading {
			t.Fatalf("CountLeadingBits(%d) = %d, want %d", tt.in, got, tt.leading)
		}
	}

	sgl := []struct {
		in          FixpSGL
		normz, norm int
	}{
		{0, 16, 0},
		{1, 15, 14},
		{-1, 0, 15},
		{2, 14, 13},
		{-2, 0, 14},
		{0x4000, 1, 0},
		{MaxValSGL, 1, 0},
		{MinValSGL, 0, 0},
		{12345, 2, 1},
		{-23456, 0, 0},
	}
	for _, tt := range sgl {
		if got := FixNormZS(tt.in); got != tt.normz {
			t.Fatalf("FixNormZS(%d) = %d, want %d", tt.in, got, tt.normz)
		}
		if got := FixNormS(tt.in); got != tt.norm {
			t.Fatalf("FixNormS(%d) = %d, want %d", tt.in, got, tt.norm)
		}
	}
}

func TestScaleValueVectors(t *testing.T) {
	tests := []struct {
		value          FixpDBL
		scale          int
		raw, saturated FixpDBL
	}{
		{0, 31, 0, 0},
		{1, 30, 1073741824, 1073741824},
		{-1, 30, -1073741824, -1073741824},
		{0x40000000, 1, MinValDBL, MaxValDBL},
		{-0x40000000, 1, MinValDBL, -2147483647},
		{MaxValDBL, 1, -2, MaxValDBL},
		{MinValDBL, 1, 0, -2147483647},
		{0x12345678, 3, -1851608128, MaxValDBL},
		{-0x1234567, 4, -305419888, -305419888},
		{0x12345678, -3, 38177487, 38177487},
		{-0x1234567, -4, -1193047, -1193047},
		{1, -31, 0, 0},
		{-1, -31, -1, 0},
		{MinValDBL, -1, -1073741824, -1073741824},
	}
	for _, tt := range tests {
		if got := ScaleValueDBL(tt.value, tt.scale); got != tt.raw {
			t.Fatalf("ScaleValueDBL(%d,%d) = %d, want %d", tt.value, tt.scale, got, tt.raw)
		}
		if got := ScaleValueSaturateDBL(tt.value, tt.scale); got != tt.saturated {
			t.Fatalf("ScaleValueSaturateDBL(%d,%d) = %d, want %d", tt.value, tt.scale, got, tt.saturated)
		}
	}
}

func TestSaturateShiftVectors(t *testing.T) {
	tests := []struct {
		src                FixpDBL
		scale, bits        int
		right, left, shift FixpDBL
		rightAlt, leftAlt  FixpDBL
	}{
		{0, 1, 32, 0, 0, 0, 0, 0},
		{1, 1, 32, 0, 2, 2, 0, 2},
		{-1, 1, 32, -1, -2, -2, -1, -2},
		{0x40000000, 1, 32, 536870912, MaxValDBL, MaxValDBL, 536870912, MaxValDBL},
		{-0x40000000, 1, 32, -536870912, MinValDBL, MinValDBL, -536870912, -2147483647},
		{0x40000001, 1, 32, 536870912, MaxValDBL, MaxValDBL, 536870912, MaxValDBL},
		{-0x40000001, 1, 32, -536870913, MinValDBL, MinValDBL, -536870913, -2147483647},
		{MaxValDBL, 16, 16, 32767, 32767, 32767, 32767, 32767},
		{MinValDBL, 16, 16, -32768, -32768, -32768, -32767, -32767},
		{0x00100000, 4, 16, 32767, 32767, 32767, 32767, 32767},
		{-0x00100000, 4, 16, -32768, -32768, -32768, -32767, -32767},
		{0x00010000, 1, 16, 32767, 32767, 32767, 32767, 32767},
		{-0x00010000, 1, 16, -32768, -32768, -32768, -32767, -32767},
	}
	for _, tt := range tests {
		if got := SaturateRightShift(tt.src, tt.scale, tt.bits); got != tt.right {
			t.Fatalf("SaturateRightShift(%d,%d,%d) = %d, want %d", tt.src, tt.scale, tt.bits, got, tt.right)
		}
		if got := SaturateLeftShift(tt.src, tt.scale, tt.bits); got != tt.left {
			t.Fatalf("SaturateLeftShift(%d,%d,%d) = %d, want %d", tt.src, tt.scale, tt.bits, got, tt.left)
		}
		if got := SaturateShift(tt.src, -tt.scale, tt.bits); got != tt.shift {
			t.Fatalf("SaturateShift(%d,%d,%d) = %d, want %d", tt.src, -tt.scale, tt.bits, got, tt.shift)
		}
		if got := SaturateRightShiftAlt(tt.src, tt.scale, tt.bits); got != tt.rightAlt {
			t.Fatalf("SaturateRightShiftAlt(%d,%d,%d) = %d, want %d", tt.src, tt.scale, tt.bits, got, tt.rightAlt)
		}
		if got := SaturateLeftShiftAlt(tt.src, tt.scale, tt.bits); got != tt.leftAlt {
			t.Fatalf("SaturateLeftShiftAlt(%d,%d,%d) = %d, want %d", tt.src, tt.scale, tt.bits, got, tt.leftAlt)
		}
	}
}

func TestFixpointAllocs(t *testing.T) {
	var x FixpDBL = 0x12345678
	var y FixpDBL = -0x1234567
	var s FixpSGL = 12345
	allocs := testing.AllocsPerRun(1000, func() {
		x = FMultDD(x, y)
		y = ScaleValueSaturateDBL(y, -3)
		x = SaturateLeftShift(FMultDS(x, s), 1, DfractBits)
		fixpointSink = x ^ y
	})
	if allocs != 0 {
		t.Fatalf("fixed-point helpers allocations = %v, want 0", allocs)
	}
}
