package fdkaac

func fPow(baseM FixpDBL, baseE int, expM FixpDBL, expE int) (FixpDBL, int) {
	if baseM <= 0 {
		return 0, 0
	}

	baseLog2, baseLog2E := fLog2MantExp(baseM, baseE)
	leadingBits := CountLeadingBits(fixpAbsDBL(expM))
	expM <<= uint(leadingBits)
	expE -= leadingBits

	ansLog2 := FMultDD(baseLog2, expM)
	ansLog2E := expE + baseLog2E
	return f2Pow(ansLog2, ansLog2E)
}
