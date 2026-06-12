package fdkaac

type FixpSPK struct {
	Re FixpSGL
	Im FixpSGL
}

type FixpDPK struct {
	Re FixpDBL
	Im FixpDBL
}

func CplxMultDiv2DD(aRe, aIm, bRe, bIm FixpDBL) (FixpDBL, FixpDBL) {
	re := int64(aRe)*int64(bRe) - int64(aIm)*int64(bIm)
	im := int64(aRe)*int64(bIm) + int64(aIm)*int64(bRe)
	return FixpDBL(re >> 32), FixpDBL(im >> 32)
}

func CplxMultDiv2DS(aRe, aIm FixpDBL, bRe, bIm FixpSGL) (FixpDBL, FixpDBL) {
	return CplxMultDiv2DD(aRe, aIm, FXSGL2FXDBL(bRe), FXSGL2FXDBL(bIm))
}

func CplxMultDiv2SPK(aRe, aIm FixpDBL, w FixpSPK) (FixpDBL, FixpDBL) {
	return CplxMultDiv2DS(aRe, aIm, w.Re, w.Im)
}

func CplxMultDiv2SSPK(aRe, aIm FixpSGL, w FixpSPK) (FixpDBL, FixpDBL) {
	return CplxMultDiv2SS(aRe, aIm, w.Re, w.Im)
}

func CplxMultDiv2SSPKSGL(aRe, aIm FixpSGL, w FixpSPK) (FixpSGL, FixpSGL) {
	return CplxMultDiv2SSSGL(aRe, aIm, w.Re, w.Im)
}

func CplxMultDiv2DPK(aRe, aIm FixpDBL, w FixpDPK) (FixpDBL, FixpDBL) {
	return CplxMultDiv2DD(aRe, aIm, w.Re, w.Im)
}

func CplxMultDiv2SS(aRe, aIm, bRe, bIm FixpSGL) (FixpDBL, FixpDBL) {
	re := FMultDiv2SS(aRe, bRe) - FMultDiv2SS(aIm, bIm)
	im := FMultDiv2SS(aRe, bIm) + FMultDiv2SS(aIm, bRe)
	return re, im
}

func CplxMultDiv2SSSGL(aRe, aIm, bRe, bIm FixpSGL) (FixpSGL, FixpSGL) {
	re, im := CplxMultDiv2SS(aRe, aIm, bRe, bIm)
	return FXDBL2FXSGL(re), FXDBL2FXSGL(im)
}

func CplxMultSubDiv2DD(cRe, cIm, aRe, aIm, bRe, bIm FixpDBL) (FixpDBL, FixpDBL) {
	cRe -= FMultDiv2DD(aRe, bRe) - FMultDiv2DD(aIm, bIm)
	cIm -= FMultDiv2DD(aRe, bIm) + FMultDiv2DD(aIm, bRe)
	return cRe, cIm
}

func CplxMultSubDiv2DS(cRe, cIm, aRe, aIm FixpDBL, bRe, bIm FixpSGL) (FixpDBL, FixpDBL) {
	cRe -= FMultDiv2DS(aRe, bRe) - FMultDiv2DS(aIm, bIm)
	cIm -= FMultDiv2DS(aRe, bIm) + FMultDiv2DS(aIm, bRe)
	return cRe, cIm
}

func CplxMultDD(aRe, aIm, bRe, bIm FixpDBL) (FixpDBL, FixpDBL) {
	re := FMultDD(aRe, bRe) - FMultDD(aIm, bIm)
	im := FMultDD(aRe, bIm) + FMultDD(aIm, bRe)
	return re, im
}

func CplxMultDS(aRe, aIm FixpDBL, bRe, bIm FixpSGL) (FixpDBL, FixpDBL) {
	re := FMultDS(aRe, bRe) - FMultDS(aIm, bIm)
	im := FMultDS(aRe, bIm) + FMultDS(aIm, bRe)
	return re, im
}

func CplxMultSPK(aRe, aIm FixpDBL, w FixpSPK) (FixpDBL, FixpDBL) {
	return CplxMultDS(aRe, aIm, w.Re, w.Im)
}

func CplxMultDPK(aRe, aIm FixpDBL, w FixpDPK) (FixpDBL, FixpDBL) {
	return CplxMultDD(aRe, aIm, w.Re, w.Im)
}

func CplxMultSS(aRe, aIm, bRe, bIm FixpSGL) (FixpSGL, FixpSGL) {
	re := FMultSS(aRe, bRe) - FMultSS(aIm, bIm)
	im := FMultSS(aRe, bIm) + FMultSS(aIm, bRe)
	return FXDBL2FXSGL(re), FXDBL2FXSGL(im)
}
