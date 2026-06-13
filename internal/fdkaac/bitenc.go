package fdkaac

const (
	globalGainOffset = 100
	icsReservedBit   = 0
	noiseOffset      = 90
	logNormPCM       = -15

	sectEscValLong  = 31
	sectEscValShort = 7
	codeBookBits    = 4
	sectBitsLong    = 5
	sectBitsShort   = 3

	maxTnsFilters = 2
	tnsMaxOrder   = 12

	acScalable = 0x000008
	acELD      = 0x000010
)

type TNSInfo struct {
	NumOfFilters [transFac]int
	CoefRes      [transFac]int
	Length       [transFac][maxTnsFilters]int
	Order        [transFac][maxTnsFilters]int
	Direction    [transFac][maxTnsFilters]int
	CoefCompress [transFac][maxTnsFilters]int
	Coef         [transFac][maxTnsFilters][tnsMaxOrder]int
}

func FDKaacEncEncodeSpectralData(sfbOffset []int, sectionData *SectionData, quantSpectrum []int16, bitstream *BitStream) int {
	checkEncodeSpectralDataInputs(sfbOffset, sectionData, quantSpectrum, bitstream)

	startBits := BitStreamValidBits(bitstream)
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.CodeBook != codeBookPNSNo {
			tmp := section.SfbStart + section.SfbCnt
			for sfb := section.SfbStart; sfb < tmp; sfb++ {
				FDKaacEncCodeValues(
					quantSpectrum[sfbOffset[sfb]:],
					sfbOffset[sfb+1]-sfbOffset[sfb],
					section.CodeBook,
					bitstream,
				)
			}
		}
	}
	return int(BitStreamValidBits(bitstream) - startBits)
}

func FDKaacEncEncodeGlobalGain(globalGain int, scalefac int, bitstream *BitStream, mdctScale int) int {
	if bitstream != nil {
		WriteBits(bitstream, uint32(globalGain-scalefac+globalGainOffset-4*(logNormPCM-mdctScale)), 8)
	}
	return 8
}

func FDKaacEncEncodeIcsInfo(blockType int, windowShape int, groupingMask int, maxSfbPerGroup int, bitstream *BitStream, syntaxFlags uint32) int {
	checkEncodeIcsInfoInputs(blockType, windowShape, groupingMask, maxSfbPerGroup)

	statBits := 0
	if blockType == ShortWindow {
		statBits = 8 + transFac - 1
	} else if syntaxFlags&acELD != 0 {
		statBits = 6
	} else if syntaxFlags&acScalable == 0 {
		statBits = 11
	} else {
		statBits = 10
	}

	if bitstream != nil {
		if syntaxFlags&acELD == 0 {
			WriteBits(bitstream, icsReservedBit, 1)
			WriteBits(bitstream, uint32(blockType), 2)
			shape := windowShape
			if shape == WindowShapeLOL {
				shape = WindowShapeKBD
			}
			WriteBits(bitstream, uint32(shape), 1)
		}

		switch blockType {
		case LongWindow, StartWindow, StopWindow:
			WriteBits(bitstream, uint32(maxSfbPerGroup), 6)
			if syntaxFlags&(acScalable|acELD) == 0 {
				WriteBits(bitstream, 0, 1)
			}
		case ShortWindow:
			WriteBits(bitstream, uint32(maxSfbPerGroup), 4)
			WriteBits(bitstream, uint32(groupingMask), transFac-1)
		}
	}
	return statBits
}

func FDKaacEncEncodeSectionData(sectionData *SectionData, bitstream *BitStream, useVCB11 bool) int {
	checkEncodeSectionDataInputs(sectionData)
	if bitstream == nil {
		return 0
	}

	sectEscapeVal := 0
	sectLenBits := uint32(0)
	switch sectionData.BlockType {
	case LongWindow, StartWindow, StopWindow:
		sectEscapeVal = sectEscValLong
		sectLenBits = sectBitsLong
	case ShortWindow:
		sectEscapeVal = sectEscValShort
		sectLenBits = sectBitsShort
	}

	startBits := BitStreamValidBits(bitstream)
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		WriteBits(bitstream, uint32(section.CodeBook), codeBookBits)

		sectLen := section.SfbCnt
		for sectLen >= sectEscapeVal {
			WriteBits(bitstream, uint32(sectEscapeVal), sectLenBits)
			sectLen -= sectEscapeVal
		}
		WriteBits(bitstream, uint32(sectLen), sectLenBits)
	}
	return int(BitStreamValidBits(bitstream) - startBits)
}

func FDKaacEncEncodeScaleFactorData(
	maxValueInSfb []uint32,
	sectionData *SectionData,
	scalefac []int,
	bitstream *BitStream,
	noiseNrg []int,
	isScale []int,
	globalGain int,
) int {
	checkEncodeScaleFactorDataInputs(maxValueInSfb, sectionData, scalefac, noiseNrg, isScale)
	if bitstream == nil {
		return 0
	}

	startBits := BitStreamValidBits(bitstream)
	lastValScf := scalefac[sectionData.FirstScf]
	lastValPns := globalGain - scalefac[sectionData.FirstScf] + globalGainOffset - 4*logNormPCM - noiseOffset
	noisePCMFlag := true
	lastValIs := 0

	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.CodeBook == codeBookZeroNo {
			continue
		}
		if section.CodeBook == codeBookISOutOfPhaseNo || section.CodeBook == codeBookISInPhaseNo {
			sfbStart := section.SfbStart
			tmp := sfbStart + section.SfbCnt
			for j := sfbStart; j < tmp; j++ {
				deltaIs := isScale[j] - lastValIs
				lastValIs = isScale[j]
				if FDKaacEncCodeScalefactorDelta(deltaIs, bitstream) != 0 {
					return 1
				}
			}
		} else if section.CodeBook == codeBookPNSNo {
			sfbStart := section.SfbStart
			tmp := sfbStart + section.SfbCnt
			for j := sfbStart; j < tmp; j++ {
				deltaPns := noiseNrg[j] - lastValPns
				lastValPns = noiseNrg[j]
				if noisePCMFlag {
					WriteBits(bitstream, uint32(deltaPns+(1<<(pnsPCMBits-1))), pnsPCMBits)
					noisePCMFlag = false
				} else if FDKaacEncCodeScalefactorDelta(deltaPns, bitstream) != 0 {
					return 1
				}
			}
		} else {
			tmp := section.SfbStart + section.SfbCnt
			for j := section.SfbStart; j < tmp; j++ {
				deltaScf := 0
				if maxValueInSfb[j] != 0 {
					deltaScf = -(scalefac[j] - lastValScf)
					lastValScf = scalefac[j]
				}
				if FDKaacEncCodeScalefactorDelta(deltaScf, bitstream) != 0 {
					return 1
				}
			}
		}
	}
	return int(BitStreamValidBits(bitstream) - startBits)
}

func FDKaacEncEncodeMSInfo(sfbCnt int, grpSfb int, maxSfb int, msDigest int, jsFlags []int, bitstream *BitStream) int {
	checkEncodeMSInfoInputs(sfbCnt, grpSfb, maxSfb, msDigest, jsFlags)

	msBits := 0
	switch msDigest {
	case MsMaskNone:
		if bitstream != nil {
			WriteBits(bitstream, MsMaskNone, 2)
		}
		msBits += 2
	case MsMaskAll:
		if bitstream != nil {
			WriteBits(bitstream, MsMaskAll, 2)
		}
		msBits += 2
	case MsMaskSome:
		if bitstream != nil {
			WriteBits(bitstream, MsMaskSome, 2)
		}
		msBits += 2
		for sfbOff := 0; sfbOff < sfbCnt; sfbOff += grpSfb {
			for sfb := 0; sfb < maxSfb; sfb++ {
				if bitstream != nil {
					if jsFlags[sfbOff+sfb]&1 != 0 {
						WriteBits(bitstream, 1, 1)
					} else {
						WriteBits(bitstream, 0, 1)
					}
				}
				msBits++
			}
		}
	}
	return msBits
}

func FDKaacEncEncodeTnsDataPresent(tnsInfo *TNSInfo, blockType int, bitstream *BitStream) int {
	checkEncodeTnsInfoInputs(tnsInfo, blockType)
	if bitstream != nil && tnsInfo != nil {
		if tnsDataPresent(tnsInfo, blockType) {
			WriteBits(bitstream, 1, 1)
		} else {
			WriteBits(bitstream, 0, 1)
		}
	}
	return 1
}

func FDKaacEncEncodeTnsData(tnsInfo *TNSInfo, blockType int, bitstream *BitStream) int {
	checkEncodeTnsInfoInputs(tnsInfo, blockType)
	if tnsInfo == nil || !tnsDataPresent(tnsInfo, blockType) {
		return 0
	}

	tnsBits := 0
	numOfWindows := tnsWindowCount(blockType)
	numFilterBits := uint32(2)
	lengthBits := uint32(6)
	orderBits := uint32(5)
	if blockType == ShortWindow {
		numFilterBits = 1
		lengthBits = 4
		orderBits = 3
	}

	for i := 0; i < numOfWindows; i++ {
		numFilters := tnsInfo.NumOfFilters[i]
		if bitstream != nil {
			WriteBits(bitstream, uint32(numFilters), numFilterBits)
		}
		tnsBits += int(numFilterBits)
		if numFilters != 0 {
			if bitstream != nil {
				if tnsInfo.CoefRes[i] == 4 {
					WriteBits(bitstream, 1, 1)
				} else {
					WriteBits(bitstream, 0, 1)
				}
			}
			tnsBits++
		}
		for j := 0; j < numFilters; j++ {
			if bitstream != nil {
				WriteBits(bitstream, uint32(tnsInfo.Length[i][j]), lengthBits)
				WriteBits(bitstream, uint32(tnsInfo.Order[i][j]), orderBits)
			}
			tnsBits += int(lengthBits + orderBits)
			if tnsInfo.Order[i][j] != 0 {
				coefBits := tnsCoefBits(tnsInfo, i, j)
				if bitstream != nil {
					WriteBits(bitstream, uint32(tnsInfo.Direction[i][j]), 1)
					WriteBits(bitstream, uint32(tnsInfo.CoefRes[i]-coefBits), 1)
					for k := 0; k < tnsInfo.Order[i][j]; k++ {
						WriteBits(bitstream, uint32(tnsInfo.Coef[i][j][k]&int(BitMask[coefBits])), uint32(coefBits))
					}
				}
				tnsBits += 2 + tnsInfo.Order[i][j]*coefBits
			}
		}
	}
	return tnsBits
}

func FDKaacEncEncodeGainControlData(bitstream *BitStream) int {
	if bitstream != nil {
		WriteBits(bitstream, 0, 1)
	}
	return 1
}

func FDKaacEncEncodePulseData(bitstream *BitStream) int {
	if bitstream != nil {
		WriteBits(bitstream, 0, 1)
	}
	return 1
}

func checkEncodeSpectralDataInputs(sfbOffset []int, sectionData *SectionData, quantSpectrum []int16, bitstream *BitStream) {
	if bitstream == nil {
		panic("fdkaac: nil spectral-data bitstream")
	}
	if sectionData == nil {
		panic("fdkaac: nil spectral-data section data")
	}
	if sectionData.NoOfSections < 0 || sectionData.NoOfSections > maxSections {
		panic("fdkaac: invalid spectral-data section count")
	}
	if len(sfbOffset) < sectionData.SfbCnt+1 {
		panic("fdkaac: short spectral-data offsets")
	}
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.SfbStart < 0 || section.SfbCnt < 0 || section.SfbStart+section.SfbCnt > sectionData.SfbCnt {
			panic("fdkaac: invalid spectral-data section")
		}
		if section.CodeBook == codeBookPNSNo {
			continue
		}
		for sfb := section.SfbStart; sfb < section.SfbStart+section.SfbCnt; sfb++ {
			if sfbOffset[sfb] < 0 || sfbOffset[sfb+1] < sfbOffset[sfb] || len(quantSpectrum) < sfbOffset[sfb+1] {
				panic("fdkaac: malformed spectral-data offsets")
			}
		}
	}
}

func checkEncodeIcsInfoInputs(blockType int, windowShape int, groupingMask int, maxSfbPerGroup int) {
	if blockType != LongWindow && blockType != StartWindow && blockType != ShortWindow && blockType != StopWindow {
		panic("fdkaac: invalid ICS block type")
	}
	if windowShape != WindowShapeSine && windowShape != WindowShapeKBD && windowShape != WindowShapeLOL {
		panic("fdkaac: invalid ICS window shape")
	}
	if maxSfbPerGroup < 0 || maxSfbPerGroup > maxGroupedSFB {
		panic("fdkaac: invalid ICS max SFB")
	}
	if groupingMask < 0 || groupingMask >= (1<<(transFac-1)) {
		panic("fdkaac: invalid ICS grouping mask")
	}
}

func checkEncodeSectionDataInputs(sectionData *SectionData) {
	if sectionData == nil {
		panic("fdkaac: nil section data")
	}
	if sectionData.BlockType != LongWindow && sectionData.BlockType != StartWindow && sectionData.BlockType != ShortWindow && sectionData.BlockType != StopWindow {
		panic("fdkaac: invalid section-data block type")
	}
	if sectionData.NoOfSections < 0 || sectionData.NoOfSections > maxSections {
		panic("fdkaac: invalid section-data count")
	}
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.CodeBook < codeBookZeroNo || section.CodeBook > codeBookISInPhaseNo {
			panic("fdkaac: invalid section-data codebook")
		}
		if section.SfbCnt < 0 {
			panic("fdkaac: invalid section-data length")
		}
	}
}

func checkEncodeMSInfoInputs(sfbCnt int, grpSfb int, maxSfb int, msDigest int, jsFlags []int) {
	if msDigest != MsMaskNone && msDigest != MsMaskSome && msDigest != MsMaskAll {
		panic("fdkaac: invalid MS digest")
	}
	if sfbCnt < 0 || sfbCnt > maxGroupedSFB {
		panic("fdkaac: invalid MS sfb count")
	}
	if msDigest != MsMaskSome {
		return
	}
	if grpSfb <= 0 || sfbCnt%grpSfb != 0 {
		panic("fdkaac: invalid MS group width")
	}
	if maxSfb < 0 || maxSfb > grpSfb {
		panic("fdkaac: invalid MS max sfb")
	}
	if len(jsFlags) < sfbCnt {
		panic("fdkaac: short MS mask")
	}
}

func checkEncodeTnsInfoInputs(tnsInfo *TNSInfo, blockType int) {
	if blockType != LongWindow && blockType != StartWindow && blockType != ShortWindow && blockType != StopWindow {
		panic("fdkaac: invalid TNS block type")
	}
	if tnsInfo == nil {
		return
	}

	numOfWindows := tnsWindowCount(blockType)
	lengthLimit := 1 << 6
	orderLimit := 1 << 5
	if blockType == ShortWindow {
		lengthLimit = 1 << 4
		orderLimit = 1 << 3
	}

	for i := 0; i < numOfWindows; i++ {
		numFilters := tnsInfo.NumOfFilters[i]
		if numFilters < 0 || numFilters > maxTnsFilters {
			panic("fdkaac: invalid TNS filter count")
		}
		if numFilters == 0 {
			continue
		}
		if tnsInfo.CoefRes[i] != 3 && tnsInfo.CoefRes[i] != 4 {
			panic("fdkaac: invalid TNS coefficient resolution")
		}
		for j := 0; j < numFilters; j++ {
			if tnsInfo.Length[i][j] < 0 || tnsInfo.Length[i][j] >= lengthLimit {
				panic("fdkaac: invalid TNS filter length")
			}
			if tnsInfo.Order[i][j] < 0 || tnsInfo.Order[i][j] > tnsMaxOrder || tnsInfo.Order[i][j] >= orderLimit {
				panic("fdkaac: invalid TNS filter order")
			}
			if tnsInfo.Direction[i][j] != 0 && tnsInfo.Direction[i][j] != 1 {
				panic("fdkaac: invalid TNS filter direction")
			}
			coefBits := tnsCoefBits(tnsInfo, i, j)
			coefMin := -(1 << (coefBits - 1))
			coefMax := (1 << (coefBits - 1)) - 1
			for k := 0; k < tnsInfo.Order[i][j]; k++ {
				if tnsInfo.Coef[i][j][k] < coefMin || tnsInfo.Coef[i][j][k] > coefMax {
					panic("fdkaac: invalid TNS coefficient")
				}
			}
		}
	}
}

func tnsWindowCount(blockType int) int {
	if blockType == ShortWindow {
		return transFac
	}
	return 1
}

func tnsDataPresent(tnsInfo *TNSInfo, blockType int) bool {
	numOfWindows := tnsWindowCount(blockType)
	for i := 0; i < numOfWindows; i++ {
		if tnsInfo.NumOfFilters[i] != 0 {
			return true
		}
	}
	return false
}

func tnsCoefBits(tnsInfo *TNSInfo, window int, filter int) int {
	if tnsInfo.CoefRes[window] == 4 {
		for k := 0; k < tnsInfo.Order[window][filter]; k++ {
			if tnsInfo.Coef[window][filter][k] > 3 || tnsInfo.Coef[window][filter][k] < -4 {
				return 4
			}
		}
		return 3
	}
	for k := 0; k < tnsInfo.Order[window][filter]; k++ {
		if tnsInfo.Coef[window][filter][k] > 1 || tnsInfo.Coef[window][filter][k] < -2 {
			return 3
		}
	}
	return 2
}

func checkEncodeScaleFactorDataInputs(maxValueInSfb []uint32, sectionData *SectionData, scalefac []int, noiseNrg []int, isScale []int) {
	if sectionData == nil {
		panic("fdkaac: nil scale-factor section data")
	}
	if sectionData.NoOfSections < 0 || sectionData.NoOfSections > maxSections {
		panic("fdkaac: invalid scale-factor section count")
	}
	if sectionData.FirstScf < 0 || len(scalefac) <= sectionData.FirstScf {
		panic("fdkaac: invalid first scale factor")
	}
	if len(maxValueInSfb) < sectionData.SfbCnt || len(scalefac) < sectionData.SfbCnt || len(noiseNrg) < sectionData.SfbCnt || len(isScale) < sectionData.SfbCnt {
		panic("fdkaac: short scale-factor data")
	}
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.SfbStart < 0 || section.SfbCnt < 0 || section.SfbStart+section.SfbCnt > sectionData.SfbCnt {
			panic("fdkaac: invalid scale-factor section")
		}
	}
}
