package fdkaac

const (
	invalidBitcount = fdkIntMax / 4

	codeBookZeroNo = 0
	codeBook1No    = 1
	codeBook2No    = 2
	codeBook3No    = 3
	codeBook4No    = 4
	codeBook5No    = 5
	codeBook6No    = 6
	codeBook7No    = 7
	codeBook8No    = 8
	codeBook9No    = 9
	codeBook10No   = 10
	codeBookEscNo  = 11

	codeBookEscLav = 16
	codeBookScfLav = 60
)

// fdkaacEncHuffLtabScf is the pinned FDK-AAC v2.0.3 scalefactor Huffman
// length table. Source: third_party/fdk-aac/libAACenc/src/aacEnc_rom.cpp.
var fdkaacEncHuffLtabScf = [121]uint8{
	0x12, 0x12, 0x12, 0x12, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13,
	0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x12, 0x13, 0x12,
	0x11, 0x11, 0x10, 0x11, 0x10, 0x10, 0x10, 0x10, 0x0f, 0x0f, 0x0e,
	0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0d, 0x0d, 0x0c, 0x0c, 0x0c, 0x0b,
	0x0c, 0x0b, 0x0a, 0x0a, 0x0a, 0x09, 0x09, 0x08, 0x08, 0x08, 0x07,
	0x06, 0x06, 0x05, 0x04, 0x03, 0x01, 0x04, 0x04, 0x05, 0x06, 0x06,
	0x07, 0x07, 0x08, 0x08, 0x09, 0x09, 0x0a, 0x0a, 0x0a, 0x0b, 0x0b,
	0x0b, 0x0b, 0x0c, 0x0c, 0x0d, 0x0d, 0x0d, 0x0e, 0x0e, 0x10, 0x0f,
	0x10, 0x0f, 0x12, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13,
	0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13,
	0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13, 0x13,
}

func FDKaacEncBitCountScalefactorDelta(delta int) int {
	index := delta + codeBookScfLav
	if index < 0 || index >= len(fdkaacEncHuffLtabScf) {
		panic("fdkaac: scalefactor delta out of range")
	}
	return int(fdkaacEncHuffLtabScf[index])
}

func FDKaacEncBitCount(values []int16, width int, maxVal int, bitCount []int) int {
	checkBitCountInputs(values, width, maxVal, bitCount)

	if maxVal == 0 {
		bitCount[0] = 0
	} else {
		bitCount[0] = invalidBitcount
	}

	switch minInt(maxVal, codeBookEscLav) {
	case 0, 1:
		fdkaacEncCount1234567891011(values, width, bitCount)
	case 2:
		fdkaacEncCount34567891011(values, width, bitCount)
	case 3, 4:
		fdkaacEncCount567891011(values, width, bitCount)
	case 5, 6, 7:
		fdkaacEncCount7891011(values, width, bitCount)
	case 8, 9, 10, 11, 12:
		fdkaacEncCount91011(values, width, bitCount)
	case 13, 14, 15:
		fdkaacEncCount11(values, width, bitCount)
	default:
		fdkaacEncCountEsc(values, width, bitCount)
	}
	return 0
}

func FDKaacEncCountValues(values []int16, width int, codeBook int) int {
	checkCountValuesInputs(values, width, codeBook)

	bitCnt := 0
	switch codeBook {
	case codeBookZeroNo:
	case codeBook1No:
		for i := 0; i < width; i += 4 {
			t0 := int(values[i+0])
			t1 := int(values[i+1])
			t2 := int(values[i+2])
			t3 := int(values[i+3])
			bitCnt += hiLtab(fdkaacEncHuffLtab12[t0+1][t1+1][t2+1][t3+1])
		}
	case codeBook2No:
		for i := 0; i < width; i += 4 {
			t0 := int(values[i+0])
			t1 := int(values[i+1])
			t2 := int(values[i+2])
			t3 := int(values[i+3])
			bitCnt += loLtab(fdkaacEncHuffLtab12[t0+1][t1+1][t2+1][t3+1])
		}
	case codeBook3No:
		for i := 0; i < width; i += 4 {
			t0 := absInt16(values[i+0])
			bitCnt += nonzeroBit(t0)
			t1 := absInt16(values[i+1])
			bitCnt += nonzeroBit(t1)
			t2 := absInt16(values[i+2])
			bitCnt += nonzeroBit(t2)
			t3 := absInt16(values[i+3])
			bitCnt += nonzeroBit(t3)
			bitCnt += hiLtab(fdkaacEncHuffLtab34[t0][t1][t2][t3])
		}
	case codeBook4No:
		for i := 0; i < width; i += 4 {
			t0 := absInt16(values[i+0])
			bitCnt += nonzeroBit(t0)
			t1 := absInt16(values[i+1])
			bitCnt += nonzeroBit(t1)
			t2 := absInt16(values[i+2])
			bitCnt += nonzeroBit(t2)
			t3 := absInt16(values[i+3])
			bitCnt += nonzeroBit(t3)
			bitCnt += loLtab(fdkaacEncHuffLtab34[t0][t1][t2][t3])
		}
	case codeBook5No:
		for i := 0; i < width; i += 4 {
			t0 := int(values[i+0])
			t1 := int(values[i+1])
			t2 := int(values[i+2])
			t3 := int(values[i+3])
			bitCnt += hiLtab(fdkaacEncHuffLtab56[t0+4][t1+4]) +
				hiLtab(fdkaacEncHuffLtab56[t2+4][t3+4])
		}
	case codeBook6No:
		for i := 0; i < width; i += 4 {
			t0 := int(values[i+0])
			t1 := int(values[i+1])
			t2 := int(values[i+2])
			t3 := int(values[i+3])
			bitCnt += loLtab(fdkaacEncHuffLtab56[t0+4][t1+4]) +
				loLtab(fdkaacEncHuffLtab56[t2+4][t3+4])
		}
	case codeBook7No:
		for i := 0; i < width; i += 4 {
			t0 := absInt16(values[i+0])
			bitCnt += nonzeroBit(t0)
			t1 := absInt16(values[i+1])
			bitCnt += nonzeroBit(t1)
			t2 := absInt16(values[i+2])
			bitCnt += nonzeroBit(t2)
			t3 := absInt16(values[i+3])
			bitCnt += nonzeroBit(t3)
			bitCnt += hiLtab(fdkaacEncHuffLtab78[t0][t1]) +
				hiLtab(fdkaacEncHuffLtab78[t2][t3])
		}
	case codeBook8No:
		for i := 0; i < width; i += 4 {
			t0 := absInt16(values[i+0])
			bitCnt += nonzeroBit(t0)
			t1 := absInt16(values[i+1])
			bitCnt += nonzeroBit(t1)
			t2 := absInt16(values[i+2])
			bitCnt += nonzeroBit(t2)
			t3 := absInt16(values[i+3])
			bitCnt += nonzeroBit(t3)
			bitCnt += loLtab(fdkaacEncHuffLtab78[t0][t1]) +
				loLtab(fdkaacEncHuffLtab78[t2][t3])
		}
	case codeBook9No:
		for i := 0; i < width; i += 4 {
			t0 := absInt16(values[i+0])
			bitCnt += nonzeroBit(t0)
			t1 := absInt16(values[i+1])
			bitCnt += nonzeroBit(t1)
			t2 := absInt16(values[i+2])
			bitCnt += nonzeroBit(t2)
			t3 := absInt16(values[i+3])
			bitCnt += nonzeroBit(t3)
			bitCnt += hiLtab(fdkaacEncHuffLtab910[t0][t1]) +
				hiLtab(fdkaacEncHuffLtab910[t2][t3])
		}
	case codeBook10No:
		for i := 0; i < width; i += 4 {
			t0 := absInt16(values[i+0])
			bitCnt += nonzeroBit(t0)
			t1 := absInt16(values[i+1])
			bitCnt += nonzeroBit(t1)
			t2 := absInt16(values[i+2])
			bitCnt += nonzeroBit(t2)
			t3 := absInt16(values[i+3])
			bitCnt += nonzeroBit(t3)
			bitCnt += loLtab(fdkaacEncHuffLtab910[t0][t1]) +
				loLtab(fdkaacEncHuffLtab910[t2][t3])
		}
	case codeBookEscNo:
		for i := 0; i < width; i += 2 {
			t0 := absInt16(values[i+0])
			bitCnt += nonzeroBit(t0)
			t1 := absInt16(values[i+1])
			bitCnt += nonzeroBit(t1)
			bitCnt += int(fdkaacEncHuffLtab11[minInt(t0, 16)][minInt(t1, 16)])
			if t0 >= 16 {
				bitCnt += 5
				for t0 >>= 1; t0 >= 16; t0 >>= 1 {
					bitCnt += 2
				}
			}
			if t1 >= 16 {
				bitCnt += 5
				for t1 >>= 1; t1 >= 16; t1 >>= 1 {
					bitCnt += 2
				}
			}
		}
	}
	return bitCnt
}

func fdkaacEncCount1234567891011(values []int16, width int, bitCount []int) {
	bc12 := 0
	bc34 := 0
	bc56 := 0
	bc78 := 0
	bc910 := 0
	bc11 := 0
	sc := 0

	for i := 0; i < width; i += 4 {
		t0 := int(values[i+0])
		t1 := int(values[i+1])
		t2 := int(values[i+2])
		t3 := int(values[i+3])

		bc12 += int(fdkaacEncHuffLtab12[t0+1][t1+1][t2+1][t3+1])
		bc56 += int(fdkaacEncHuffLtab56[t0+4][t1+4]) + int(fdkaacEncHuffLtab56[t2+4][t3+4])

		a0 := absInt(t0)
		sc += nonzeroBit(a0)
		a1 := absInt(t1)
		sc += nonzeroBit(a1)
		a2 := absInt(t2)
		sc += nonzeroBit(a2)
		a3 := absInt(t3)
		sc += nonzeroBit(a3)

		bc34 += int(fdkaacEncHuffLtab34[a0][a1][a2][a3])
		bc78 += int(fdkaacEncHuffLtab78[a0][a1]) + int(fdkaacEncHuffLtab78[a2][a3])
		bc910 += int(fdkaacEncHuffLtab910[a0][a1]) + int(fdkaacEncHuffLtab910[a2][a3])
		bc11 += int(fdkaacEncHuffLtab11[a0][a1]) + int(fdkaacEncHuffLtab11[a2][a3])
	}
	bitCount[1] = hiLtabInt(bc12)
	bitCount[2] = loLtabInt(bc12)
	bitCount[3] = hiLtabInt(bc34) + sc
	bitCount[4] = loLtabInt(bc34) + sc
	bitCount[5] = hiLtabInt(bc56)
	bitCount[6] = loLtabInt(bc56)
	bitCount[7] = hiLtabInt(bc78) + sc
	bitCount[8] = loLtabInt(bc78) + sc
	bitCount[9] = hiLtabInt(bc910) + sc
	bitCount[10] = loLtabInt(bc910) + sc
	bitCount[11] = bc11 + sc
}

func fdkaacEncCount34567891011(values []int16, width int, bitCount []int) {
	bc34 := 0
	bc56 := 0
	bc78 := 0
	bc910 := 0
	bc11 := 0
	sc := 0

	for i := 0; i < width; i += 4 {
		t0 := int(values[i+0])
		t1 := int(values[i+1])
		t2 := int(values[i+2])
		t3 := int(values[i+3])

		bc56 += int(fdkaacEncHuffLtab56[t0+4][t1+4]) + int(fdkaacEncHuffLtab56[t2+4][t3+4])

		a0 := absInt(t0)
		sc += nonzeroBit(a0)
		a1 := absInt(t1)
		sc += nonzeroBit(a1)
		a2 := absInt(t2)
		sc += nonzeroBit(a2)
		a3 := absInt(t3)
		sc += nonzeroBit(a3)

		bc34 += int(fdkaacEncHuffLtab34[a0][a1][a2][a3])
		bc78 += int(fdkaacEncHuffLtab78[a0][a1]) + int(fdkaacEncHuffLtab78[a2][a3])
		bc910 += int(fdkaacEncHuffLtab910[a0][a1]) + int(fdkaacEncHuffLtab910[a2][a3])
		bc11 += int(fdkaacEncHuffLtab11[a0][a1]) + int(fdkaacEncHuffLtab11[a2][a3])
	}
	bitCount[1] = invalidBitcount
	bitCount[2] = invalidBitcount
	bitCount[3] = hiLtabInt(bc34) + sc
	bitCount[4] = loLtabInt(bc34) + sc
	bitCount[5] = hiLtabInt(bc56)
	bitCount[6] = loLtabInt(bc56)
	bitCount[7] = hiLtabInt(bc78) + sc
	bitCount[8] = loLtabInt(bc78) + sc
	bitCount[9] = hiLtabInt(bc910) + sc
	bitCount[10] = loLtabInt(bc910) + sc
	bitCount[11] = bc11 + sc
}

func fdkaacEncCount567891011(values []int16, width int, bitCount []int) {
	bc56 := 0
	bc78 := 0
	bc910 := 0
	bc11 := 0
	sc := 0

	for i := 0; i < width; i += 4 {
		t0 := int(values[i+0])
		t1 := int(values[i+1])
		t2 := int(values[i+2])
		t3 := int(values[i+3])

		bc56 += int(fdkaacEncHuffLtab56[t0+4][t1+4]) + int(fdkaacEncHuffLtab56[t2+4][t3+4])

		a0 := absInt(t0)
		sc += nonzeroBit(a0)
		a1 := absInt(t1)
		sc += nonzeroBit(a1)
		a2 := absInt(t2)
		sc += nonzeroBit(a2)
		a3 := absInt(t3)
		sc += nonzeroBit(a3)

		bc78 += int(fdkaacEncHuffLtab78[a0][a1]) + int(fdkaacEncHuffLtab78[a2][a3])
		bc910 += int(fdkaacEncHuffLtab910[a0][a1]) + int(fdkaacEncHuffLtab910[a2][a3])
		bc11 += int(fdkaacEncHuffLtab11[a0][a1]) + int(fdkaacEncHuffLtab11[a2][a3])
	}
	bitCount[1] = invalidBitcount
	bitCount[2] = invalidBitcount
	bitCount[3] = invalidBitcount
	bitCount[4] = invalidBitcount
	bitCount[5] = hiLtabInt(bc56)
	bitCount[6] = loLtabInt(bc56)
	bitCount[7] = hiLtabInt(bc78) + sc
	bitCount[8] = loLtabInt(bc78) + sc
	bitCount[9] = hiLtabInt(bc910) + sc
	bitCount[10] = loLtabInt(bc910) + sc
	bitCount[11] = bc11 + sc
}

func fdkaacEncCount7891011(values []int16, width int, bitCount []int) {
	bc78 := 0
	bc910 := 0
	bc11 := 0
	sc := 0

	for i := 0; i < width; i += 4 {
		a0 := absInt16(values[i+0])
		sc += nonzeroBit(a0)
		a1 := absInt16(values[i+1])
		sc += nonzeroBit(a1)
		a2 := absInt16(values[i+2])
		sc += nonzeroBit(a2)
		a3 := absInt16(values[i+3])
		sc += nonzeroBit(a3)

		bc78 += int(fdkaacEncHuffLtab78[a0][a1]) + int(fdkaacEncHuffLtab78[a2][a3])
		bc910 += int(fdkaacEncHuffLtab910[a0][a1]) + int(fdkaacEncHuffLtab910[a2][a3])
		bc11 += int(fdkaacEncHuffLtab11[a0][a1]) + int(fdkaacEncHuffLtab11[a2][a3])
	}
	bitCount[1] = invalidBitcount
	bitCount[2] = invalidBitcount
	bitCount[3] = invalidBitcount
	bitCount[4] = invalidBitcount
	bitCount[5] = invalidBitcount
	bitCount[6] = invalidBitcount
	bitCount[7] = hiLtabInt(bc78) + sc
	bitCount[8] = loLtabInt(bc78) + sc
	bitCount[9] = hiLtabInt(bc910) + sc
	bitCount[10] = loLtabInt(bc910) + sc
	bitCount[11] = bc11 + sc
}

func fdkaacEncCount91011(values []int16, width int, bitCount []int) {
	bc910 := 0
	bc11 := 0
	sc := 0

	for i := 0; i < width; i += 4 {
		a0 := absInt16(values[i+0])
		sc += nonzeroBit(a0)
		a1 := absInt16(values[i+1])
		sc += nonzeroBit(a1)
		a2 := absInt16(values[i+2])
		sc += nonzeroBit(a2)
		a3 := absInt16(values[i+3])
		sc += nonzeroBit(a3)

		bc910 += int(fdkaacEncHuffLtab910[a0][a1]) + int(fdkaacEncHuffLtab910[a2][a3])
		bc11 += int(fdkaacEncHuffLtab11[a0][a1]) + int(fdkaacEncHuffLtab11[a2][a3])
	}
	bitCount[1] = invalidBitcount
	bitCount[2] = invalidBitcount
	bitCount[3] = invalidBitcount
	bitCount[4] = invalidBitcount
	bitCount[5] = invalidBitcount
	bitCount[6] = invalidBitcount
	bitCount[7] = invalidBitcount
	bitCount[8] = invalidBitcount
	bitCount[9] = hiLtabInt(bc910) + sc
	bitCount[10] = loLtabInt(bc910) + sc
	bitCount[11] = bc11 + sc
}

func fdkaacEncCount11(values []int16, width int, bitCount []int) {
	bc11 := 0
	sc := 0

	for i := 0; i < width; i += 4 {
		a0 := absInt16(values[i+0])
		sc += nonzeroBit(a0)
		a1 := absInt16(values[i+1])
		sc += nonzeroBit(a1)
		a2 := absInt16(values[i+2])
		sc += nonzeroBit(a2)
		a3 := absInt16(values[i+3])
		sc += nonzeroBit(a3)

		bc11 += int(fdkaacEncHuffLtab11[a0][a1]) + int(fdkaacEncHuffLtab11[a2][a3])
	}
	for i := 1; i < 11; i++ {
		bitCount[i] = invalidBitcount
	}
	bitCount[11] = bc11 + sc
}

func fdkaacEncCountEsc(values []int16, width int, bitCount []int) {
	bc11 := 0
	ec := 0
	sc := 0

	for i := 0; i < width; i += 2 {
		t0 := absInt16(values[i+0])
		t1 := absInt16(values[i+1])
		sc += nonzeroBit(t0) + nonzeroBit(t1)

		bc11 += int(fdkaacEncHuffLtab11[minInt(t0, 16)][minInt(t1, 16)])

		if t0 >= 16 {
			ec += 5
			for t0 >>= 1; t0 >= 16; t0 >>= 1 {
				ec += 2
			}
		}
		if t1 >= 16 {
			ec += 5
			for t1 >>= 1; t1 >= 16; t1 >>= 1 {
				ec += 2
			}
		}
	}

	for i := 0; i < 11; i++ {
		bitCount[i] = invalidBitcount
	}
	bitCount[11] = bc11 + sc + ec
}

func checkBitCountInputs(values []int16, width int, maxVal int, bitCount []int) {
	if width < 0 || width%4 != 0 || maxVal < 0 {
		panic("fdkaac: invalid spectral bit-count control")
	}
	if len(values) < width {
		panic("fdkaac: short spectral bit-count values")
	}
	if len(bitCount) < codeBookEscNo+1 {
		panic("fdkaac: short spectral bit-count output")
	}
}

func checkCountValuesInputs(values []int16, width int, codeBook int) {
	if width < 0 || codeBook < codeBookZeroNo || codeBook > codeBookEscNo {
		panic("fdkaac: invalid count-values control")
	}
	if codeBook == codeBookEscNo {
		if width%2 != 0 {
			panic("fdkaac: invalid escape count-values width")
		}
	} else if codeBook != codeBookZeroNo && width%4 != 0 {
		panic("fdkaac: invalid count-values width")
	}
	if len(values) < width {
		panic("fdkaac: short count-values spectrum")
	}
}

func hiLtab(x uint32) int {
	return int(x >> 16)
}

func loLtab(x uint32) int {
	return int(x & 0xffff)
}

func hiLtabInt(x int) int {
	return x >> 16
}

func loLtabInt(x int) int {
	return x & 0xffff
}

func nonzeroBit(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
