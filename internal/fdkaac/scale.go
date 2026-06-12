package fdkaac

func ScaleValuesSGL(vector []FixpSGL, scalefactor int) {
	if scalefactor == 0 {
		return
	}
	if scalefactor > 0 {
		scalefactor = minInt(scalefactor, FractBits-1)
		for i := range vector {
			vector[i] = FixpSGL(int32(vector[i]) << uint(scalefactor))
		}
		return
	}
	negScalefactor := minInt(-scalefactor, FractBits-1)
	for i := range vector {
		vector[i] = FixpSGL(int32(vector[i]) >> uint(negScalefactor))
	}
}

func ScaleValuesDBL(vector []FixpDBL, scalefactor int) {
	if scalefactor == 0 {
		return
	}
	if scalefactor > 0 {
		scalefactor = minInt(scalefactor, DfractBits-1)
		for i := range vector {
			vector[i] <<= uint(scalefactor)
		}
		return
	}
	negScalefactor := minInt(-scalefactor, DfractBits-1)
	for i := range vector {
		vector[i] >>= uint(negScalefactor)
	}
}

func ScaleValuesDBLToDBL(dst, src []FixpDBL, scalefactor int) {
	n := minInt(len(dst), len(src))
	if scalefactor == 0 {
		copy(dst[:n], src[:n])
		return
	}
	if scalefactor > 0 {
		scalefactor = minInt(scalefactor, DfractBits-1)
		for i := 0; i < n; i++ {
			dst[i] = src[i] << uint(scalefactor)
		}
		return
	}
	negScalefactor := minInt(-scalefactor, DfractBits-1)
	for i := 0; i < n; i++ {
		dst[i] = src[i] >> uint(negScalefactor)
	}
}

func ScaleValuesPCMFromDBL(dst []FixpSGL, src []FixpDBL, scalefactor int) {
	n := minInt(len(dst), len(src))
	scalefactor -= DfractBits - FractBits
	if scalefactor > 0 {
		scalefactor = minInt(scalefactor, DfractBits-1)
		for i := 0; i < n; i++ {
			dst[i] = FixpSGL(src[i] << uint(scalefactor))
		}
		return
	}
	negScalefactor := minInt(-scalefactor, DfractBits-1)
	for i := 0; i < n; i++ {
		dst[i] = FixpSGL(src[i] >> uint(negScalefactor))
	}
}

func ScaleValuesSGLToSGL(dst, src []FixpSGL, scalefactor int) {
	n := minInt(len(dst), len(src))
	if scalefactor == 0 {
		copy(dst[:n], src[:n])
		return
	}
	if scalefactor > 0 {
		scalefactor = minInt(scalefactor, DfractBits-1)
		for i := 0; i < n; i++ {
			dst[i] = FixpSGL(int32(src[i]) << uint(scalefactor))
		}
		return
	}
	negScalefactor := minInt(-scalefactor, DfractBits-1)
	for i := 0; i < n; i++ {
		dst[i] = FixpSGL(int32(src[i]) >> uint(negScalefactor))
	}
}

func ScaleValuesWithFactorDBL(vector []FixpDBL, factor FixpDBL, scalefactor int) {
	shift := minInt(scalefactor+1, DfractBits-1)
	if shift >= 0 {
		for i := range vector {
			vector[i] = FMultDiv2DD(vector[i], factor) << uint(shift)
		}
		return
	}
	shift = -shift
	for i := range vector {
		vector[i] = FMultDiv2DD(vector[i], factor) >> uint(shift)
	}
}

func ScaleValuesSaturateDBL(vector []FixpDBL, scalefactor int) {
	if scalefactor == 0 {
		return
	}
	scalefactor = clampScaleDBL(scalefactor)
	for i := range vector {
		vector[i] = ScaleValueSaturateDBL(vector[i], scalefactor)
	}
}

func ScaleValuesSaturateDBLToDBL(dst, src []FixpDBL, scalefactor int) {
	n := minInt(len(dst), len(src))
	if scalefactor == 0 {
		copy(dst[:n], src[:n])
		return
	}
	scalefactor = clampScaleDBL(scalefactor)
	for i := 0; i < n; i++ {
		dst[i] = ScaleValueSaturateDBL(src[i], scalefactor)
	}
}

func ScaleValuesSaturateSGLFromDBL(dst []FixpSGL, src []FixpDBL, scalefactor int) {
	n := minInt(len(dst), len(src))
	scalefactor = clampScaleDBL(scalefactor)
	for i := 0; i < n; i++ {
		dst[i] = FXDBL2FXSGL(FAddSaturateDBL(ScaleValueSaturateDBL(src[i], scalefactor), 0x8000))
	}
}

func ScaleValuesSaturateSGL(vector []FixpSGL, scalefactor int) {
	if scalefactor == 0 {
		return
	}
	scalefactor = clampScaleDBL(scalefactor)
	for i := range vector {
		vector[i] = FXDBL2FXSGL(ScaleValueSaturateDBL(FXSGL2FXDBL(vector[i]), scalefactor))
	}
}

func ScaleValuesSaturateSGLToSGL(dst, src []FixpSGL, scalefactor int) {
	n := minInt(len(dst), len(src))
	if scalefactor == 0 {
		copy(dst[:n], src[:n])
		return
	}
	scalefactor = clampScaleDBL(scalefactor)
	for i := 0; i < n; i++ {
		dst[i] = FXDBL2FXSGL(ScaleValueSaturateDBL(FXSGL2FXDBL(src[i]), scalefactor))
	}
}

func GetScalefactorShort(vector []FixpSGL) int {
	var maxVal int32
	for _, value := range vector {
		temp := int32(value)
		maxVal = int32(int16(maxVal | (temp ^ (temp >> 15))))
	}
	return maxInt(0, FixNormZD(FixpDBL(maxVal))-1-(DfractBits-FractBits))
}

func GetScalefactorPCM(vector []FixpSGL, stride int) int {
	if stride <= 0 {
		return 0
	}
	var maxVal int32
	for i := 0; i < len(vector); i += stride {
		temp := int32(vector[i])
		maxVal = int32(int16(maxVal | (temp ^ (temp >> 15))))
	}
	return maxInt(0, FixNormZD(FixpDBL(maxVal))-1-(DfractBits-FractBits))
}

func GetScalefactorDBL(vector []FixpDBL) int {
	var maxVal FixpDBL
	for _, value := range vector {
		maxVal |= value ^ (value >> (DfractBits - 1))
	}
	return maxInt(0, FixNormZD(maxVal)-1)
}

func GetScalefactorSGL(vector []FixpSGL) int {
	var maxVal int32
	for _, value := range vector {
		temp := int32(value)
		maxVal = int32(int16(maxVal | (temp ^ (temp >> (FractBits - 1)))))
	}
	return maxInt(0, FixNormZS(FixpSGL(maxVal))-1)
}

func clampScaleDBL(scalefactor int) int {
	if scalefactor > DfractBits-1 {
		return DfractBits - 1
	}
	if scalefactor < -(DfractBits - 1) {
		return -(DfractBits - 1)
	}
	return scalefactor
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
