package fdkaac

func dctGetTables(length int) ([]FixpSPK, []FixpSPK, int) {
	if length < 4 {
		panic("fdkaac: DCT length not supported")
	}

	ld2Length := DfractBits - 1 - FixNormZD(FixpDBL(length)) - 1
	if length>>uint(ld2Length-1) != 0x4 {
		panic("fdkaac: DCT length not supported")
	}

	sinStep := 1 << uint(10-ld2Length)
	return FDKGetWindowSlope(length, WindowShapeSine), SineTable1024[:], sinStep
}

func DCTIV(pDat []FixpDBL, length int, pDatE *int) {
	if length < 4 || len(pDat) < length {
		panic("fdkaac: DCT-IV length not supported")
	}

	twiddle, sinTwiddle, sinStep := dctGetTables(length)
	m := length >> 1

	p0 := 0
	p1 := length - 2
	i := 0
	for ; i < m-1; i += 2 {
		accu1 := pDat[p1+1]
		accu2 := pDat[p0]
		accu3 := pDat[p0+1]
		accu4 := pDat[p1]

		accu1, accu2 = CplxMultDiv2SPK(accu1, accu2, twiddle[i])
		accu3, accu4 = CplxMultDiv2SPK(accu4, accu3, twiddle[i+1])

		pDat[p0] = accu2 >> 1
		pDat[p0+1] = accu1 >> 1
		pDat[p1] = accu4 >> 1
		pDat[p1+1] = -(accu3 >> 1)

		p0 += 2
		p1 -= 2
	}
	if m&1 != 0 {
		accu1 := pDat[p1+1]
		accu2 := pDat[p0]

		accu1, accu2 = CplxMultDiv2SPK(accu1, accu2, twiddle[i])

		pDat[p0] = accu2 >> 1
		pDat[p0+1] = accu1 >> 1
	}

	FFT(m, pDat[:length], pDatE)

	p0 = 0
	p1 = length - 2
	accu1 := pDat[p1]
	accu2 := pDat[p1+1]

	pDat[p1+1] = -pDat[p0+1]

	for idx, i := sinStep, 1; i < (m+1)>>1; i, idx = i+1, idx+sinStep {
		twd := sinTwiddle[idx]
		accu3, accu4 := CplxMultSPK(accu1, accu2, twd)
		pDat[p0+1] = accu3
		pDat[p1] = accu4

		p0 += 2
		p1 -= 2

		accu3, accu4 = CplxMultSPK(pDat[p0+1], pDat[p0], twd)

		accu1 = pDat[p1]
		accu2 = pDat[p1+1]

		pDat[p1+1] = -accu3
		pDat[p0] = accu4
	}

	if m&1 == 0 {
		wPiFourth := STC(0x5a82799a)
		accu1 = FMultDS(accu1, wPiFourth)
		accu2 = FMultDS(accu2, wPiFourth)

		pDat[p1] = accu1 + accu2
		pDat[p0+1] = accu1 - accu2
	}

	*pDatE += 2
}
