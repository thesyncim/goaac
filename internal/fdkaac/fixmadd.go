package fdkaac

func FixMAddDiv2DD(x, a, b FixpDBL) FixpDBL {
	return x + FMultDiv2DD(a, b)
}

func FixMAddDiv2SD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMAddDiv2DD(x, FXSGL2FXDBL(a), b)
}

func FixMAddDiv2DS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMAddDiv2DD(x, a, FXSGL2FXDBL(b))
}

func FixMAddDiv2SS(x FixpDBL, a, b FixpSGL) FixpDBL {
	return x + FMultDiv2SS(a, b)
}

func FixMSubDiv2DD(x, a, b FixpDBL) FixpDBL {
	return x - FMultDiv2DD(a, b)
}

func FixMSubDiv2SD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMSubDiv2DD(x, FXSGL2FXDBL(a), b)
}

func FixMSubDiv2DS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMSubDiv2DD(x, a, FXSGL2FXDBL(b))
}

func FixMSubDiv2SS(x FixpDBL, a, b FixpSGL) FixpDBL {
	return x - FMultDiv2SS(a, b)
}

func FixMAddDiv2BitExactDD(x, a, b FixpDBL) FixpDBL {
	return x + FMultDiv2BitExactDD(a, b)
}

func FixMAddDiv2BitExactSD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return x + FMultDiv2BitExactSD(a, b)
}

func FixMAddDiv2BitExactDS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return x + FMultDiv2BitExactDS(a, b)
}

func FixMSubDiv2BitExactDD(x, a, b FixpDBL) FixpDBL {
	return x - FMultDiv2BitExactDD(a, b)
}

func FixMSubDiv2BitExactSD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return x - FMultDiv2BitExactSD(a, b)
}

func FixMSubDiv2BitExactDS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return x - FMultDiv2BitExactDS(a, b)
}

func FixMAddDD(x, a, b FixpDBL) FixpDBL {
	return FixMAddDiv2DD(x, a, b) << 1
}

func FixMAddSD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMAddDD(x, FXSGL2FXDBL(a), b)
}

func FixMAddDS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMAddDD(x, a, FXSGL2FXDBL(b))
}

func FixMAddSS(x FixpDBL, a, b FixpSGL) FixpDBL {
	return (x + FMultDiv2SS(a, b)) << 1
}

func FixMSubDD(x, a, b FixpDBL) FixpDBL {
	return FixMSubDiv2DD(x, a, b) << 1
}

func FixMSubSD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMSubDD(x, FXSGL2FXDBL(a), b)
}

func FixMSubDS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMSubDD(x, a, FXSGL2FXDBL(b))
}

func FixMSubSS(x FixpDBL, a, b FixpSGL) FixpDBL {
	return (x - FMultDiv2SS(a, b)) << 1
}

func FixPAddDiv2D(x, a FixpDBL) FixpDBL {
	return FixMAddDiv2DD(x, a, a)
}

func FixPAddD(x, a FixpDBL) FixpDBL {
	return x + FixPow2D(a)
}

func FixPAddDiv2S(x FixpDBL, a FixpSGL) FixpDBL {
	return FixMAddDiv2SS(x, a, a)
}

func FixPAddS(x FixpDBL, a FixpSGL) FixpDBL {
	return x + FixPow2S(a)
}

func FMultAddDiv2DD(x, a, b FixpDBL) FixpDBL {
	return FixMAddDiv2DD(x, a, b)
}

func FMultAddDiv2SD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMAddDiv2SD(x, a, b)
}

func FMultAddDiv2DS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMAddDiv2DS(x, a, b)
}

func FMultAddDiv2SS(x FixpDBL, a, b FixpSGL) FixpDBL {
	return FixMAddDiv2SS(x, a, b)
}

func FMultSubDiv2DD(x, a, b FixpDBL) FixpDBL {
	return FixMSubDiv2DD(x, a, b)
}

func FMultSubDiv2SD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMSubDiv2SD(x, a, b)
}

func FMultSubDiv2DS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMSubDiv2DS(x, a, b)
}

func FMultSubDiv2SS(x FixpDBL, a, b FixpSGL) FixpDBL {
	return FixMSubDiv2SS(x, a, b)
}

func FMultAddDD(x, a, b FixpDBL) FixpDBL {
	return FixMAddDD(x, a, b)
}

func FMultAddSD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMAddSD(x, a, b)
}

func FMultAddDS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMAddDS(x, a, b)
}

func FMultAddSS(x FixpDBL, a, b FixpSGL) FixpDBL {
	return FixMAddSS(x, a, b)
}

func FMultSubDD(x, a, b FixpDBL) FixpDBL {
	return FixMSubDD(x, a, b)
}

func FMultSubSD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMSubSD(x, a, b)
}

func FMultSubDS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMSubDS(x, a, b)
}

func FMultSubSS(x FixpDBL, a, b FixpSGL) FixpDBL {
	return FixMSubSS(x, a, b)
}

func FMultAddDiv2BitExactDD(x, a, b FixpDBL) FixpDBL {
	return FixMAddDiv2BitExactDD(x, a, b)
}

func FMultAddDiv2BitExactSD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMAddDiv2BitExactSD(x, a, b)
}

func FMultAddDiv2BitExactDS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMAddDiv2BitExactDS(x, a, b)
}

func FMultSubDiv2BitExactDD(x, a, b FixpDBL) FixpDBL {
	return FixMSubDiv2BitExactDD(x, a, b)
}

func FMultSubDiv2BitExactSD(x FixpDBL, a FixpSGL, b FixpDBL) FixpDBL {
	return FixMSubDiv2BitExactSD(x, a, b)
}

func FMultSubDiv2BitExactDS(x, a FixpDBL, b FixpSGL) FixpDBL {
	return FixMSubDiv2BitExactDS(x, a, b)
}

func FPow2AddDiv2D(x, a FixpDBL) FixpDBL {
	return FixPAddDiv2D(x, a)
}

func FPow2AddD(x, a FixpDBL) FixpDBL {
	return FixPAddD(x, a)
}

func FPow2AddDiv2S(x FixpDBL, a FixpSGL) FixpDBL {
	return FixPAddDiv2S(x, a)
}

func FPow2AddS(x FixpDBL, a FixpSGL) FixpDBL {
	return FixPAddS(x, a)
}
