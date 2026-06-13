package fdkaac

const (
	qAtanInp = 25
	qAtanOut = 30
	atiSF    = (DfractBits - 1) - qAtanInp
)

func fixpAtan(x FixpDBL) FixpDBL {
	sign := false
	if x < 0 {
		sign = true
		if x == MinValDBL {
			x = MaxValDBL
		} else {
			x = -x
		}
	}

	const (
		atanOneOver64    FixpDBL = 0x02000000
		atanOneP28Over64 FixpDBL = 0x028f5c29
		atanP281         FixpDBL = 0x00013000
		atanOneP571      FixpDBL = 0x6487ef00
		atanCoeffA1      FixpDBL = 0x7fe39ad6
		atanCoeffA2      FixpDBL = -688340004
		atanCoeffA3      FixpDBL = 0x128ec947
		atanCoeffA4      FixpDBL = -82150838
		atanPiBy4Q30     FixpDBL = 0x3243f69a
	)

	var result FixpDBL
	if x < atanOneOver64 {
		x <<= atiSF
		x2 := FixPow2D(x)
		temp := FMultAddDiv2DD(atanCoeffA3>>1, x2, atanCoeffA4)
		temp = FMultAddDiv2DD(atanCoeffA2>>2, x2, temp)
		temp = FMultAddDiv2DD(atanCoeffA1>>3, x2, temp)
		result = FMultDD(x, temp<<2)
	} else if x < atanOneP28Over64 {
		deltaFix := (x - atanOneOver64) << 5
		result = atanPiBy4Q30 + (deltaFix >> 1) - FixPow2Div2D(deltaFix)
	} else {
		temp := FixPow2Div2D(x) + atanP281
		div, resE := fDivNormExp(x, temp)
		result = ScaleValueDBL(div, (qAtanOut-qAtanInp+18-DfractBits+1)+resE)
		result = atanOneP571 - result
	}

	if sign {
		result = -result
	}
	return result
}
