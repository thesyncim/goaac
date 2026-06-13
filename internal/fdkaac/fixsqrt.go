package fdkaac

const (
	sqrtBits          = 7
	sqrtBitsMask      = 0x7f
	sqrtFractBitsMask = 0x007fffff
)

// invSqrtTab is the pinned FDK-AAC v2.0.3 inverse-square-root table.
// Source: third_party/fdk-aac/libFDK/src/FDK_tools_rom.cpp:invSqrtTab.
var invSqrtTab = [...]FixpDBL{
	0x5A827999, 0x5A287E03, 0x59CF8CBC, 0x5977A0AC, 0x5920B4DF, 0x58CAC480,
	0x5875CADE, 0x5821C364, 0x57CEA99D, 0x577C7930, 0x572B2DE0, 0x56DAC38E,
	0x568B3632, 0x563C81E0, 0x55EEA2C4, 0x55A19522, 0x55555555, 0x5509DFD0,
	0x54BF311A, 0x547545D0, 0x542C1AA4, 0x53E3AC5B, 0x539BF7CD, 0x5354F9E7,
	0x530EAFA5, 0x52C91618, 0x52842A5F, 0x523FE9AC, 0x51FC5140, 0x51B95E6B,
	0x51770E8F, 0x51355F1A, 0x50F44D89, 0x50B3D768, 0x5073FA50, 0x5034B3E7,
	0x4FF601E0, 0x4FB7E1FA, 0x4F7A5202, 0x4F3D4FCF, 0x4F00D944, 0x4EC4EC4F,
	0x4E8986EA, 0x4E4EA718, 0x4E144AE9, 0x4DDA7073, 0x4DA115DA, 0x4D683948,
	0x4D2FD8F4, 0x4CF7F31B, 0x4CC08605, 0x4C899000, 0x4C530F65, 0x4C1D0294,
	0x4BE767F5, 0x4BB23DF9, 0x4B7D8317, 0x4B4935CF, 0x4B1554A6, 0x4AE1DE2A,
	0x4AAED0F0, 0x4A7C2B93, 0x4A49ECB3, 0x4A1812FA, 0x49E69D16, 0x49B589BB,
	0x4984D7A4, 0x49548592, 0x49249249, 0x48F4FC97, 0x48C5C34B, 0x4896E53D,
	0x48686148, 0x483A364D, 0x480C6332, 0x47DEE6E1, 0x47B1C049, 0x4784EE60,
	0x4758701C, 0x472C447C, 0x47006A81, 0x46D4E130, 0x46A9A794, 0x467EBCBA,
	0x46541FB4, 0x4629CF98, 0x45FFCB80, 0x45D6128A, 0x45ACA3D5, 0x45837E88,
	0x455AA1CB, 0x45320CC8, 0x4509BEB0, 0x44E1B6B4, 0x44B9F40B, 0x449275ED,
	0x446B3B96, 0x44444444, 0x441D8F3B, 0x43F71BBF, 0x43D0E917, 0x43AAF68F,
	0x43854374, 0x435FCF15, 0x433A98C6, 0x43159FDC, 0x42F0E3AE, 0x42CC6398,
	0x42A81EF6, 0x42841527, 0x4260458E, 0x423CAF8D, 0x4219528B, 0x41F62DF2,
	0x41D3412A, 0x41B08BA2, 0x418E0CC8, 0x416BC40D, 0x4149B0E5, 0x4127D2C3,
	0x41062920, 0x40E4B374, 0x40C3713B, 0x40A261EF, 0x40818512, 0x4060DA22,
	0x404060A1, 0x40201814, 0x40000000, 0x3FE017EC,
}

func invSqrtNorm2(op FixpDBL, shift *int) FixpDBL {
	if shift == nil {
		panic("fdkaac: nil inverse square-root shift")
	}
	if op == 0 {
		*shift = 16
		return MaxValDBL
	}
	if op < 0 {
		panic("fdkaac: negative inverse square-root input")
	}

	val := op
	s := FixNormZD(val) - 1
	val <<= uint(s)
	s += 2

	index := int((val >> (DfractBits - 1 - (sqrtBits + 1))) & sqrtBitsMask)
	fract := FixpDBL((int32(val) & sqrtFractBitsMask) << (sqrtBits + 1))
	diff := invSqrtTab[index+1] - invSqrtTab[index]
	reg1 := invSqrtTab[index] + (FMultDiv2DD(diff, fract) << 1)

	if fract != 0 {
		fract = FMultDiv2DD(fract, FixpDBL(uint32(0x80000000)-uint32(fract))) << 1
		diff = diff - (invSqrtTab[index+2] - invSqrtTab[index+1])
		reg1 = FMultAddDiv2DD(reg1, fract, diff)
	}

	if s&1 != 0 {
		reg1 = FMultDiv2DD(reg1, 0x5A827999) << 2
	}
	*shift = s >> 1
	return reg1
}

func sqrtFixp(op FixpDBL) FixpDBL {
	if op < 0 {
		panic("fdkaac: negative square-root input")
	}
	tmpExp := 0
	tmpInv := invSqrtNorm2(op, &tmpExp)
	return FMultDiv2DD(op<<uint(tmpExp-1), tmpInv) << 2
}
