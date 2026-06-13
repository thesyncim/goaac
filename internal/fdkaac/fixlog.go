package fdkaac

const (
	ldDataShift       = 6
	ldDataMinusOne    = MinValDBL
	ldDataStep1Over64 = FixpDBL(0x02000000)
	ldDataStep2Over64 = FixpDBL(0x04000000)
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

func CalcLdData(op FixpDBL) FixpDBL {
	return fLog2LD(op, 0)
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
