package fdkaac

const (
	ldDataShift       = 6
	ldDataMinusOne    = MinValDBL
	ldDataStep1Over64 = FixpDBL(0x02000000)
	ldDataStep2Over64 = FixpDBL(0x04000000)
	ldData31Over64    = FixpDBL(0x3e000000)
	pow2Precision     = 8
	halfDBL           = FixpDBL(0x40000000)
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

var pow2Coeff = [pow2Precision]FixpDBL{
	0x58b90bfc,
	0x1ebfbe00,
	0x071ac236,
	0x013b2ab7,
	0x002bb100,
	0x00050c24,
	0x00007ff3,
	0x00000b16,
}

var ldIntCoeff = [193]FixpDBL{
	MinValDBL + 1, 0x00000000, 0x02000000, 0x032b8034, 0x04000000, 0x04a4d3c2,
	0x052b8034, 0x059d5da0, 0x06000000, 0x06570069, 0x06a4d3c2, 0x06eb3a9f,
	0x072b8034, 0x0766a009, 0x079d5da0, 0x07d053f7, 0x08000000, 0x082cc7ee,
	0x08570069, 0x087ef05b, 0x08a4d3c2, 0x08c8ddd4, 0x08eb3a9f, 0x090c1050,
	0x092b8034, 0x0949a785, 0x0966a009, 0x0982809d, 0x099d5da0, 0x09b74949,
	0x09d053f7, 0x09e88c6b, 0x0a000000, 0x0a16bad3, 0x0a2cc7ee, 0x0a423162,
	0x0a570069, 0x0a6b3d79, 0x0a7ef05b, 0x0a92203d, 0x0aa4d3c2, 0x0ab7110e,
	0x0ac8ddd4, 0x0ada3f60, 0x0aeb3a9f, 0x0afbd42b, 0x0b0c1050, 0x0b1bf312,
	0x0b2b8034, 0x0b3abb40, 0x0b49a785, 0x0b584822, 0x0b66a009, 0x0b74b1fd,
	0x0b82809d, 0x0b900e61, 0x0b9d5da0, 0x0baa708f, 0x0bb74949, 0x0bc3e9ca,
	0x0bd053f7, 0x0bdc899b, 0x0be88c6b, 0x0bf45e09, 0x0c000000, 0x0c0b73cb,
	0x0c16bad3, 0x0c21d671, 0x0c2cc7ee, 0x0c379085, 0x0c423162, 0x0c4caba8,
	0x0c570069, 0x0c6130af, 0x0c6b3d79, 0x0c7527b9, 0x0c7ef05b, 0x0c88983f,
	0x0c92203d, 0x0c9b8926, 0x0ca4d3c2, 0x0cae00d2, 0x0cb7110e, 0x0cc0052b,
	0x0cc8ddd4, 0x0cd19bb0, 0x0cda3f60, 0x0ce2c97d, 0x0ceb3a9f, 0x0cf39355,
	0x0cfbd42b, 0x0d03fda9, 0x0d0c1050, 0x0d140ca0, 0x0d1bf312, 0x0d23c41d,
	0x0d2b8034, 0x0d3327c7, 0x0d3abb40, 0x0d423b08, 0x0d49a785, 0x0d510118,
	0x0d584822, 0x0d5f7cff, 0x0d66a009, 0x0d6db197, 0x0d74b1fd, 0x0d7ba190,
	0x0d82809d, 0x0d894f75, 0x0d900e61, 0x0d96bdad, 0x0d9d5da0, 0x0da3ee7f,
	0x0daa708f, 0x0db0e412, 0x0db74949, 0x0dbda072, 0x0dc3e9ca, 0x0dca258e,
	0x0dd053f7, 0x0dd6753e, 0x0ddc899b, 0x0de29143, 0x0de88c6b, 0x0dee7b47,
	0x0df45e09, 0x0dfa34e1, 0x0e000000, 0x0e05bf94, 0x0e0b73cb, 0x0e111cd2,
	0x0e16bad3, 0x0e1c4dfb, 0x0e21d671, 0x0e275460, 0x0e2cc7ee, 0x0e323143,
	0x0e379085, 0x0e3ce5d8, 0x0e423162, 0x0e477346, 0x0e4caba8, 0x0e51daa8,
	0x0e570069, 0x0e5c1d0b, 0x0e6130af, 0x0e663b74, 0x0e6b3d79, 0x0e7036db,
	0x0e7527b9, 0x0e7a1030, 0x0e7ef05b, 0x0e83c857, 0x0e88983f, 0x0e8d602e,
	0x0e92203d, 0x0e96d888, 0x0e9b8926, 0x0ea03232, 0x0ea4d3c2, 0x0ea96df0,
	0x0eae00d2, 0x0eb28c7f, 0x0eb7110e, 0x0ebb8e96, 0x0ec0052b, 0x0ec474e4,
	0x0ec8ddd4, 0x0ecd4012, 0x0ed19bb0, 0x0ed5f0c4, 0x0eda3f60, 0x0ede8797,
	0x0ee2c97d, 0x0ee70525, 0x0eeb3a9f, 0x0eef69ff, 0x0ef39355, 0x0ef7b6b4,
	0x0efbd42b, 0x0effebcd, 0x0f03fda9, 0x0f0809cf, 0x0f0c1050, 0x0f10113b,
	0x0f140ca0, 0x0f18028d, 0x0f1bf312, 0x0f1fde3d, 0x0f23c41d, 0x0f27a4c0,
	0x0f2b8034,
}

func CalcLdData(op FixpDBL) FixpDBL {
	return fLog2LD(op, 0)
}

func CalcLdInt(i int) FixpDBL {
	if i > 0 && i < len(ldIntCoeff) {
		return ldIntCoeff[i]
	}
	return 0
}

func schurDiv(num FixpDBL, denum FixpDBL, count int) FixpDBL {
	if num < 0 || denum <= 0 || num > denum || count <= 0 || count > DfractBits {
		panic("fdkaac: invalid Schur division input")
	}
	lNum := int32(num) >> 1
	lDenum := int32(denum) >> 1
	div := int32(0)
	k := count
	if lNum != 0 {
		for {
			k--
			if k == 0 {
				break
			}
			div <<= 1
			lNum <<= 1
			if lNum >= lDenum {
				lNum -= lDenum
				div++
			}
		}
	}
	return FixpDBL(div << uint(DfractBits-count))
}

func fDivNorm(num FixpDBL, denom FixpDBL) FixpDBL {
	if num < 0 || denom <= 0 || num > denom {
		panic("fdkaac: invalid normalized division input")
	}
	res, exp := fDivNormExp(num, denom)
	if res == FixpDBL(1<<(DfractBits-2)) && exp == 1 {
		return MaxValDBL
	}
	return ScaleValueDBL(res, exp)
}

func fDivNormExp(num FixpDBL, denom FixpDBL) (FixpDBL, int) {
	if num < 0 || denom <= 0 {
		panic("fdkaac: invalid normalized division input")
	}
	if num == 0 {
		return 0, 0
	}

	normNum := CountLeadingBits(num)
	num <<= uint(normNum)
	num >>= 1
	exp := -normNum + 1

	normDen := CountLeadingBits(denom)
	denom <<= uint(normDen)
	exp += normDen

	return schurDiv(num, denom, FractBits), exp
}

func fDivNormSignedExp(num FixpDBL, denom FixpDBL) (FixpDBL, int) {
	sign := (num >= 0) != (denom >= 0)
	if num == 0 {
		return 0, 0
	}
	if denom == 0 {
		return MaxValDBL, 14
	}

	normNum := CountLeadingBits(num)
	num <<= uint(normNum)
	num >>= 2
	num = fixpAbsDBL(num)
	exp := -normNum + 1

	normDen := CountLeadingBits(denom)
	denom <<= uint(normDen)
	denom >>= 1
	denom = fixpAbsDBL(denom)
	exp += normDen

	div := schurDiv(num, denom, FractBits)
	if sign {
		div = -div
	}
	return div, exp
}

func f2Pow(expM FixpDBL, expE int) (FixpDBL, int) {
	if expE >= DfractBits || expE <= -DfractBits {
		panic("fdkaac: invalid pow2 exponent")
	}

	intPart := 0
	var fracPart FixpDBL
	if expE > 0 {
		expBits := DfractBits - 1 - expE
		if expBits < 0 {
			panic("fdkaac: invalid pow2 exponent")
		}
		intPart = int(expM >> uint(expBits))
		fracPart = expM - FixpDBL(intPart<<uint(expBits))
		fracPart <<= uint(expE)
	} else {
		fracPart = expM >> uint(-expE)
	}

	if fracPart > halfDBL {
		intPart++
		fracPart += MinValDBL
	}
	if fracPart < -halfDBL {
		intPart--
		fracPart = -(MinValDBL - fracPart)
	}

	resultE := intPart + 1
	p := fracPart
	resultM := halfDBL
	for i := 0; i < len(pow2Coeff); i++ {
		resultM = FMultAddDiv2DD(resultM, pow2Coeff[i], p)
		p = FMultDD(p, fracPart)
	}
	return resultM, resultE
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
