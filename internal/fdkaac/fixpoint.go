package fdkaac

import "math/bits"

const (
	FractBits  = 16
	DfractBits = 32
	AccuBits   = 40
)

type FixpSGL int16
type FixpDBL int32

const (
	MaxValSGL FixpSGL = 1<<15 - 1
	MinValSGL FixpSGL = -1 << 15
	MaxValDBL FixpDBL = 1<<31 - 1
	MinValDBL FixpDBL = -1 << 31
)

func FXSGL2FXDBL(val FixpSGL) FixpDBL {
	return FixpDBL(int32(val) << (DfractBits - FractBits))
}

func FXDBL2FXSGL(val FixpDBL) FixpSGL {
	return FixpSGL(val >> (DfractBits - FractBits))
}

func FixMulDiv2DD(a, b FixpDBL) FixpDBL {
	return FixpDBL((int64(a) * int64(b)) >> 32)
}

func FixMulDiv2BitExactDD(a, b FixpDBL) FixpDBL {
	return FixMulDiv2DD(a, b)
}

func FixMulDD(a, b FixpDBL) FixpDBL {
	return FixpDBL((int64(a) * int64(b)) >> 31)
}

func FixMulBitExactDD(a, b FixpDBL) FixpDBL {
	return FixpDBL(FixMulDiv2BitExactDD(a, b) << 1)
}

func FixMulDiv2SS(a, b FixpSGL) FixpDBL {
	return FixpDBL(int32(a) * int32(b))
}

func FixMulSS(a, b FixpSGL) FixpDBL {
	return FixpDBL(FixMulDiv2SS(a, b) << 1)
}

func FixMulDiv2SD(a FixpSGL, b FixpDBL) FixpDBL {
	return FixMulDiv2DD(FXSGL2FXDBL(a), b)
}

func FixMulDiv2DS(a FixpDBL, b FixpSGL) FixpDBL {
	return FixMulDiv2SD(b, a)
}

func FixMulDiv2BitExactSD(a FixpSGL, b FixpDBL) FixpDBL {
	return FixMulDiv2SD(a, b)
}

func FixMulDiv2BitExactDS(a FixpDBL, b FixpSGL) FixpDBL {
	return FixMulDiv2DS(a, b)
}

func FixMulSD(a FixpSGL, b FixpDBL) FixpDBL {
	return FixpDBL(FixMulDiv2SD(a, b) << 1)
}

func FixMulDS(a FixpDBL, b FixpSGL) FixpDBL {
	return FixMulSD(b, a)
}

func FixMulBitExactSD(a FixpSGL, b FixpDBL) FixpDBL {
	return FixpDBL(FixMulDiv2BitExactSD(a, b) << 1)
}

func FixMulBitExactDS(a FixpDBL, b FixpSGL) FixpDBL {
	return FixMulBitExactSD(b, a)
}

func FixPow2Div2D(a FixpDBL) FixpDBL {
	return FixMulDiv2DD(a, a)
}

func FixPow2D(a FixpDBL) FixpDBL {
	return FixpDBL(FixPow2Div2D(a) << 1)
}

func FixPow2Div2S(a FixpSGL) FixpDBL {
	return FixMulDiv2SS(a, a)
}

func FixPow2S(a FixpSGL) FixpDBL {
	result := FixMulSS(a, a)
	return result ^ (result >> 31)
}

func FMultDD(a, b FixpDBL) FixpDBL {
	return FixMulDD(a, b)
}

func FMultSS(a, b FixpSGL) FixpDBL {
	return FixMulSS(a, b)
}

func FMultSD(a FixpSGL, b FixpDBL) FixpDBL {
	return FixMulSD(a, b)
}

func FMultDS(a FixpDBL, b FixpSGL) FixpDBL {
	return FixMulDS(a, b)
}

func FMultDiv2DD(a, b FixpDBL) FixpDBL {
	return FixMulDiv2DD(a, b)
}

func FMultDiv2SS(a, b FixpSGL) FixpDBL {
	return FixMulDiv2SS(a, b)
}

func FMultDiv2SD(a FixpSGL, b FixpDBL) FixpDBL {
	return FixMulDiv2SD(a, b)
}

func FMultDiv2DS(a FixpDBL, b FixpSGL) FixpDBL {
	return FixMulDiv2DS(a, b)
}

func FMultDiv2BitExactDD(a, b FixpDBL) FixpDBL {
	return FixMulDiv2BitExactDD(a, b)
}

func FMultDiv2BitExactSD(a FixpSGL, b FixpDBL) FixpDBL {
	return FixMulDiv2BitExactSD(a, b)
}

func FMultDiv2BitExactDS(a FixpDBL, b FixpSGL) FixpDBL {
	return FixMulDiv2BitExactDS(a, b)
}

func FMultBitExactDD(a, b FixpDBL) FixpDBL {
	return FixMulBitExactDD(a, b)
}

func FMultBitExactSD(a FixpSGL, b FixpDBL) FixpDBL {
	return FixMulBitExactSD(a, b)
}

func FMultBitExactDS(a FixpDBL, b FixpSGL) FixpDBL {
	return FixMulBitExactDS(a, b)
}

func FixNormZD(val FixpDBL) int {
	return bits.LeadingZeros32(uint32(val))
}

func FixNormD(val FixpDBL) int {
	if val == 0 {
		return 0
	}
	if val < 0 {
		val = ^val
	}
	return FixNormZD(val) - 1
}

func CountLeadingBits(val FixpDBL) int {
	return FixNormD(val)
}

func FAddSaturateSGL(a, b FixpSGL) FixpSGL {
	sum := int32(a) + int32(b)
	if sum > int32(MaxValSGL) {
		return MaxValSGL
	}
	if sum < int32(MinValSGL) {
		return MinValSGL
	}
	return FixpSGL(sum)
}

func FAddSaturateDBL(a, b FixpDBL) FixpDBL {
	sum := (a >> 1) + (b >> 1)
	max := MaxValDBL >> 1
	min := MinValDBL >> 1
	if sum > max {
		return max << 1
	}
	if sum < min {
		return min << 1
	}
	return sum << 1
}

func FixNormZS(val FixpSGL) int {
	shifted := int32(val) << (DfractBits - FractBits)
	if shifted == 0 {
		return FractBits
	}
	return bits.LeadingZeros32(uint32(shifted))
}

func FixNormS(val FixpSGL) int {
	shifted := FixpDBL(int32(val) << (DfractBits - FractBits))
	if shifted == 0 {
		return 0
	}
	if shifted < 0 {
		shifted = ^shifted
	}
	return FixNormZD(shifted) - 1
}

func ScaleValueDBL(value FixpDBL, scalefactor int) FixpDBL {
	if scalefactor > 0 {
		return value << uint(scalefactor)
	}
	return value >> uint(-scalefactor)
}

func ScaleValueSaturateDBL(value FixpDBL, scalefactor int) FixpDBL {
	headroom := FixNormZD(value ^ (value >> 31))
	minPlusOne := MinValDBL + 1
	if scalefactor >= 0 {
		if headroom <= scalefactor {
			if value > 0 {
				return MaxValDBL
			}
			return minPlusOne
		}
		scaled := value << uint(scalefactor)
		if scaled < minPlusOne {
			return minPlusOne
		}
		return scaled
	}

	scalefactor = -scalefactor
	if DfractBits-headroom <= scalefactor {
		return 0
	}
	scaled := value >> uint(scalefactor)
	if scaled < minPlusOne {
		return minPlusOne
	}
	return scaled
}

func SaturateRightShift(src FixpDBL, scale, dBits int) FixpDBL {
	max := saturateMax(dBits)
	if ((src ^ (src >> 31)) >> uint(scale)) > max {
		return (src >> 31) ^ max
	}
	return src >> uint(scale)
}

func SaturateLeftShift(src FixpDBL, scale, dBits int) FixpDBL {
	max := saturateMax(dBits)
	if (src ^ (src >> 31)) > (max >> uint(scale)) {
		return (src >> 31) ^ max
	}
	return src << uint(scale)
}

func SaturateShift(src FixpDBL, scale, dBits int) FixpDBL {
	if scale < 0 {
		return SaturateLeftShift(src, -scale, dBits)
	}
	return SaturateRightShift(src, scale, dBits)
}

func SaturateRightShiftAlt(src FixpDBL, scale, dBits int) FixpDBL {
	max := saturateMax(dBits)
	min := ^(max - 1)
	shifted := src >> uint(scale)
	if shifted > max {
		return max
	}
	if shifted < min {
		return min
	}
	return shifted
}

func SaturateLeftShiftAlt(src FixpDBL, scale, dBits int) FixpDBL {
	max := saturateMax(dBits)
	if src > (max >> uint(scale)) {
		return max
	}
	if src <= ^(max >> uint(scale)) {
		return ^(max - 1)
	}
	return src << uint(scale)
}

func saturateMax(dBits int) FixpDBL {
	return FixpDBL(int32((uint32(1) << uint(dBits-1)) - 1))
}
