package fdkaac

const (
	ldDataShift       = 6
	ldDataMinusOne    = MinValDBL
	ldDataStep1Over64 = FixpDBL(0x02000000)
	ldDataStep2Over64 = FixpDBL(0x04000000)
	ldData31Over64    = FixpDBL(0x3e000000)
)

var ldCoeffDBL = [10]FixpDBL{
	MinValDBL,
	-1073741824,
	-715827883,
	-536870912,
	-429496730,
	-357913941,
	-306783378,
	-268435456,
	-238609294,
	-214748365,
}

var exp2TabLong = [32]uint32{
	0x40000000, 0x4166C34C, 0x42D561B4, 0x444C0740, 0x45CAE0F2, 0x47521CC6,
	0x48E1E9BA, 0x4A7A77D4, 0x4C1BF829, 0x4DC69CDD, 0x4F7A9930, 0x51382182,
	0x52FF6B55, 0x54D0AD5A, 0x56AC1F75, 0x5891FAC1, 0x5A82799A, 0x5C7DD7A4,
	0x5E8451D0, 0x60962665, 0x62B39509, 0x64DCDEC3, 0x6712460B, 0x69540EC9,
	0x6BA27E65, 0x6DFDDBCC, 0x70666F76, 0x72DC8374, 0x75606374, 0x77F25CCE,
	0x7A92BE8B, 0x7D41D96E,
}

var exp2WTabLong = [32]uint32{
	0x40000000, 0x400B1818, 0x4016321B, 0x40214E0C, 0x402C6BE9, 0x40378BB4,
	0x4042AD6D, 0x404DD113, 0x4058F6A8, 0x40641E2B, 0x406F479E, 0x407A7300,
	0x4085A051, 0x4090CF92, 0x409C00C4, 0x40A733E6, 0x40B268FA, 0x40BD9FFF,
	0x40C8D8F5, 0x40D413DD, 0x40DF50B8, 0x40EA8F86, 0x40F5D046, 0x410112FA,
	0x410C57A2, 0x41179E3D, 0x4122E6CD, 0x412E3152, 0x41397DCC, 0x4144CC3B,
	0x41501CA0, 0x415B6EFB,
}

var exp2XTabLong = [32]uint32{
	0x40000000, 0x400058B9, 0x4000B173, 0x40010A2D, 0x400162E8, 0x4001BBA3,
	0x4002145F, 0x40026D1B, 0x4002C5D8, 0x40031E95, 0x40037752, 0x4003D011,
	0x400428CF, 0x4004818E, 0x4004DA4E, 0x4005330E, 0x40058BCE, 0x4005E48F,
	0x40063D51, 0x40069613, 0x4006EED5, 0x40074798, 0x4007A05B, 0x4007F91F,
	0x400851E4, 0x4008AAA8, 0x4009036E, 0x40095C33, 0x4009B4FA, 0x400A0DC0,
	0x400A6688, 0x400ABF4F,
}

func CalcLdData(op FixpDBL) FixpDBL {
	return fLog2LD(op, 0)
}

func CalcInvLdData(x FixpDBL) FixpDBL {
	setZero := x >= -ldData31Over64
	setMax := x >= ldData31Over64 || x == 0

	frac := FixpSGL(int16(int32(x) & 0x3ff))
	index3 := uint32(x>>10) & 0x1f
	index2 := uint32(x>>15) & 0x1f
	index1 := uint32(x>>20) & 0x1f

	exp := 0
	if x > 0 {
		exp = 31 - int(x>>25)
	} else {
		exp = -int(x >> 25)
	}
	exp = minInt(31, exp)

	lookup1 := uint32(0)
	if setZero {
		lookup1 = exp2TabLong[index1]
	}
	lookup2 := exp2WTabLong[index2]
	lookup3 := exp2XTabLong[index3]
	lookup3f := lookup3 + uint32(FMultDiv2DS(FixpDBL(0x0016302F), frac))

	lookup12 := uint32(FMultDD(FixpDBL(lookup1), FixpDBL(lookup2)))
	lookup := uint32(FMultDD(FixpDBL(lookup12), FixpDBL(lookup3f)))
	retVal := FixpDBL(int32((lookup << 3) >> uint(exp)))

	if setMax {
		retVal = MaxValDBL
	}
	return retVal
}

func LdDataVector(srcVector []FixpDBL, destVector []FixpDBL, number int) {
	if number < 0 {
		panic("fdkaac: negative ld-data count")
	}
	if len(srcVector) < number || len(destVector) < number {
		panic("fdkaac: short ld-data vector")
	}
	for i := 0; i < number; i++ {
		destVector[i] = fLog2LD(srcVector[i], 0)
	}
}

func fLog2LD(xM FixpDBL, xE int) FixpDBL {
	if xM <= 0 {
		return ldDataMinusOne
	}
	resultM, resultE := fLog2MantExp(xM, xE)
	return ScaleValueDBL(resultM, resultE-ldDataShift)
}

func fLog2MantExp(xM FixpDBL, xE int) (FixpDBL, int) {
	if xM <= 0 {
		return ldDataMinusOne, DfractBits - 1
	}

	bNorm := FixNormZD(xM) - 1
	x2M := xM << uint(bNorm)
	xE -= bNorm
	x2M = -(x2M + ldDataMinusOne)

	resultM := FixpDBL(0)
	px2M := x2M
	for i := 0; i < len(ldCoeffDBL); i++ {
		resultM = FMultAddDiv2DD(resultM, ldCoeffDBL[i], px2M)
		px2M = FMultDD(px2M, x2M)
	}

	resultM = FMultAddDiv2DD(resultM, resultM, FixpDBL(0x71547653))

	if xE != 0 {
		enorm := DfractBits - FixNormD(FixpDBL(xE))
		resultM = (resultM >> uint(enorm-1)) + (FixpDBL(xE) << uint(DfractBits-1-enorm))
		return resultM, enorm
	}

	return resultM, 1
}
