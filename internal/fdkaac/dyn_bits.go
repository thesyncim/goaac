package fdkaac

const (
	maxSections = maxGroupedSFB

	codeBookResNo          = 12
	codeBookPNSNo          = 13
	codeBookISOutOfPhaseNo = 14
	codeBookISInPhaseNo    = 15

	codeBookEscNdx = codeBookEscNo

	pnsPCMBits = 9
	noNoisePNS = fdkIntMin

	acErVCB11 = 0x000001
)

var fdkaacEncSideInfoTabLong = [...]int{
	9, 9, 9, 9, 9, 9, 9, 9, 9,
	9, 9, 9, 9, 9, 9, 9, 9, 9,
	9, 9, 9, 9, 9, 9, 9, 9, 9,
	9, 9, 9, 9, 14, 14, 14, 14, 14,
	14, 14, 14, 14, 14, 14, 14, 14, 14,
	14, 14, 14, 14, 14, 14, 14,
}

var fdkaacEncSideInfoTabShort = [...]int{
	7, 7, 7, 7, 7, 7, 7, 10,
	10, 10, 10, 10, 10, 10, 13, 13,
}

type SectionInfo struct {
	CodeBook    int
	SfbStart    int
	SfbCnt      int
	SectionBits int
}

type SectionData struct {
	BlockType      int
	NoOfGroups     int
	SfbCnt         int
	MaxSfbPerGroup int
	SfbPerGroup    int
	NoOfSections   int
	Huffsection    [maxSections]SectionInfo
	SideInfoBits   int
	HuffmanBits    int
	ScalefacBits   int
	NoiseNrgBits   int
	FirstScf       int
}

type BitCounterState struct {
	BitLookUp       [maxGroupedSFB][codeBookEscNdx + 1]int
	MergeGainLookUp [maxGroupedSFB]int
}

func FDKaacEncDynBitCount(
	state *BitCounterState,
	quantSpectrum []int16,
	maxValueInSfb []uint32,
	scalefac []int,
	blockType int,
	sfbCnt int,
	maxSfbPerGroup int,
	sfbPerGroup int,
	sfbOffset []int,
	sectionData *SectionData,
	noiseNrg []int,
	isBook []int,
	isScale []int,
	syntaxFlags uint32,
) int {
	checkDynBitCountInputs(
		state, quantSpectrum, maxValueInSfb, scalefac, blockType, sfbCnt,
		maxSfbPerGroup, sfbPerGroup, sfbOffset, sectionData, noiseNrg, isBook,
		isScale,
	)

	sectionData.BlockType = blockType
	sectionData.SfbCnt = sfbCnt
	sectionData.SfbPerGroup = sfbPerGroup
	sectionData.NoOfGroups = sfbCnt / sfbPerGroup
	sectionData.MaxSfbPerGroup = maxSfbPerGroup

	FDKaacEncNoiselessCounter(
		sectionData, state.MergeGainLookUp[:], state.BitLookUp[:],
		quantSpectrum, maxValueInSfb, sfbOffset, blockType, noiseNrg, isBook,
		(syntaxFlags&acErVCB11) != 0,
	)
	FDKaacEncScfCount(scalefac, maxValueInSfb, sectionData, isScale)
	FDKaacEncNoiseCount(sectionData, noiseNrg)

	return sectionData.HuffmanBits + sectionData.SideInfoBits + sectionData.ScalefacBits + sectionData.NoiseNrgBits
}

func FDKaacEncNoiselessCounter(
	sectionData *SectionData,
	mergeGainLookUp []int,
	bitLookUp [][codeBookEscNdx + 1]int,
	quantSpectrum []int16,
	maxValueInSfb []uint32,
	sfbOffset []int,
	blockType int,
	noiseNrg []int,
	isBook []int,
	useVCB11 bool,
) {
	sideInfoTab := fdkaacEncSideInfoTabLong[:]
	if blockType == ShortWindow {
		sideInfoTab = fdkaacEncSideInfoTabShort[:]
	}

	sectionData.NoOfSections = 0
	sectionData.HuffmanBits = 0
	sectionData.SideInfoBits = 0

	if sectionData.MaxSfbPerGroup == 0 {
		return
	}

	for grpNdx := 0; grpNdx < sectionData.SfbCnt; grpNdx += sectionData.SfbPerGroup {
		huffsection := sectionData.Huffsection[sectionData.NoOfSections:]
		FDKaacEncBuildBitLookUp(
			quantSpectrum, sectionData.MaxSfbPerGroup, sfbOffset[grpNdx:],
			maxValueInSfb[grpNdx:], bitLookUp, huffsection,
		)
		fdkaacEncGmStage0(huffsection, bitLookUp, sectionData.MaxSfbPerGroup, noiseNrg[grpNdx:], isBook[grpNdx:])
		fdkaacEncGmStage1(huffsection, bitLookUp, sectionData.MaxSfbPerGroup, sideInfoTab, useVCB11)
		fdkaacEncGmStage2(huffsection, mergeGainLookUp, bitLookUp, sectionData.MaxSfbPerGroup, sideInfoTab, useVCB11)

		for i := 0; i < sectionData.MaxSfbPerGroup; i += huffsection[i].SfbCnt {
			if isNoiseOrIntensityBook(huffsection[i].CodeBook) {
				huffsection[i].SectionBits = 0
			} else {
				fdkaacEncFindBestBook(bitLookUp[i], &huffsection[i].CodeBook, useVCB11)
			}

			huffsection[i].SfbStart += grpNdx
			sideBits := fdkaacEncGetSideInfoBits(&huffsection[i], sideInfoTab, useVCB11)
			sectionData.SideInfoBits += sideBits
			if !isNoiseOrIntensityBook(huffsection[i].CodeBook) {
				sectionData.HuffmanBits += huffsection[i].SectionBits - sideBits
			}
			sectionData.Huffsection[sectionData.NoOfSections] = huffsection[i]
			sectionData.NoOfSections++
		}
	}
}

func FDKaacEncBuildBitLookUp(
	quantSpectrum []int16,
	maxSfb int,
	sfbOffset []int,
	sfbMax []uint32,
	bitLookUp [][codeBookEscNdx + 1]int,
	huffsection []SectionInfo,
) {
	for i := 0; i < maxSfb; i++ {
		huffsection[i].SfbCnt = 1
		huffsection[i].SfbStart = i
		huffsection[i].SectionBits = invalidBitcount
		huffsection[i].CodeBook = -1
		sfbWidth := sfbOffset[i+1] - sfbOffset[i]
		FDKaacEncBitCount(quantSpectrum[sfbOffset[i]:], sfbWidth, int(sfbMax[i]), bitLookUp[i][:])
	}
}

func FDKaacEncScfCount(scalefacGain []int, maxValueInSfb []uint32, sectionData *SectionData, isScale []int) {
	sectionData.ScalefacBits = 0
	if scalefacGain == nil {
		return
	}

	lastValScf := 0
	deltaScf := 0
	found := 0
	scfSkipCounter := 0
	lastValIs := 0

	sectionData.FirstScf = 0
	for i := 0; i < sectionData.NoOfSections; i++ {
		if sectionData.Huffsection[i].CodeBook != codeBookZeroNo {
			sectionData.FirstScf = sectionData.Huffsection[i].SfbStart
			lastValScf = scalefacGain[sectionData.FirstScf]
			break
		}
	}

	for i := 0; i < sectionData.NoOfSections; i++ {
		section := &sectionData.Huffsection[i]
		if section.CodeBook == codeBookISOutOfPhaseNo || section.CodeBook == codeBookISInPhaseNo {
			for j := section.SfbStart; j < section.SfbStart+section.SfbCnt; j++ {
				deltaIs := isScale[j] - lastValIs
				lastValIs = isScale[j]
				sectionData.ScalefacBits += FDKaacEncBitCountScalefactorDelta(deltaIs)
			}
		} else if section.CodeBook != codeBookZeroNo && section.CodeBook != codeBookPNSNo {
			tmp := section.SfbStart + section.SfbCnt
			for j := section.SfbStart; j < tmp; j++ {
				if maxValueInSfb[j] == 0 {
					found = 0
					if scfSkipCounter == 0 {
						if j == tmp-1 {
							found = 0
						} else {
							for k := j + 1; k < tmp; k++ {
								if maxValueInSfb[k] != 0 {
									found = 1
									if absInt(scalefacGain[k]-lastValScf) <= codeBookScfLav {
										deltaScf = 0
									} else {
										deltaScf = lastValScf - scalefacGain[j]
										lastValScf = scalefacGain[j]
										scfSkipCounter = 0
									}
									break
								}
								scfSkipCounter++
							}
						}

						for m := i + 1; m < sectionData.NoOfSections && found == 0; m++ {
							next := &sectionData.Huffsection[m]
							if next.CodeBook != codeBookZeroNo && next.CodeBook != codeBookPNSNo {
								end := next.SfbStart + next.SfbCnt
								for n := next.SfbStart; n < end; n++ {
									if maxValueInSfb[n] != 0 {
										found = 1
										if absInt(scalefacGain[n]-lastValScf) <= codeBookScfLav {
											deltaScf = 0
										} else {
											deltaScf = lastValScf - scalefacGain[j]
											lastValScf = scalefacGain[j]
											scfSkipCounter = 0
										}
										break
									}
									scfSkipCounter++
								}
							}
						}

						if found == 0 {
							deltaScf = 0
							scfSkipCounter = 0
						}
					} else {
						deltaScf = 0
						scfSkipCounter--
					}
				} else {
					deltaScf = lastValScf - scalefacGain[j]
					lastValScf = scalefacGain[j]
				}
				sectionData.ScalefacBits += FDKaacEncBitCountScalefactorDelta(deltaScf)
			}
		}
	}
}

func FDKaacEncNoiseCount(sectionData *SectionData, noiseNrg []int) {
	noisePCMFlag := true
	lastValPns := 0
	sectionData.NoiseNrgBits = 0

	for i := 0; i < sectionData.NoOfSections; i++ {
		section := &sectionData.Huffsection[i]
		if section.CodeBook == codeBookPNSNo {
			sfbEnd := section.SfbStart + section.SfbCnt
			for j := section.SfbStart; j < sfbEnd; j++ {
				if noisePCMFlag {
					sectionData.NoiseNrgBits += pnsPCMBits
					lastValPns = noiseNrg[j]
					noisePCMFlag = false
				} else {
					deltaPns := noiseNrg[j] - lastValPns
					lastValPns = noiseNrg[j]
					sectionData.NoiseNrgBits += FDKaacEncBitCountScalefactorDelta(deltaPns)
				}
			}
		}
	}
}

func fdkaacEncGetSideInfoBits(huffsection *SectionInfo, sideInfoTab []int, useHCR bool) int {
	if useHCR && (huffsection.CodeBook == codeBookEscNo || huffsection.CodeBook >= 16) {
		return 5
	}
	return sideInfoTab[huffsection.SfbCnt]
}

func fdkaacEncFindBestBook(bc [codeBookEscNdx + 1]int, book *int, useVCB11 bool) int {
	minBits := invalidBitcount
	for j := 0; j <= codeBookEscNdx; j++ {
		if bc[j] < minBits {
			minBits = bc[j]
			*book = j
		}
	}
	return minBits
}

func fdkaacEncFindMinMergeBits(bc1 [codeBookEscNdx + 1]int, bc2 [codeBookEscNdx + 1]int, useVCB11 bool) int {
	minBits := invalidBitcount
	for j := 0; j <= codeBookEscNdx; j++ {
		minBits = minInt(minBits, bc1[j]+bc2[j])
	}
	return minBits
}

func fdkaacEncMergeBitLookUp(bc1 *[codeBookEscNdx + 1]int, bc2 [codeBookEscNdx + 1]int) {
	for j := 0; j <= codeBookEscNdx; j++ {
		bc1[j] = minInt(bc1[j]+bc2[j], invalidBitcount)
	}
}

func fdkaacEncFindMaxMerge(mergeGainLookUp []int, huffsection []SectionInfo, maxSfb int, maxNdx *int) int {
	maxMergeGain := 0
	lastMaxNdx := 0
	for i := 0; i+huffsection[i].SfbCnt < maxSfb; i += huffsection[i].SfbCnt {
		if mergeGainLookUp[i] > maxMergeGain {
			maxMergeGain = mergeGainLookUp[i]
			lastMaxNdx = i
		}
	}
	*maxNdx = lastMaxNdx
	return maxMergeGain
}

func fdkaacEncCalcMergeGain(
	huffsection []SectionInfo,
	bitLookUp [][codeBookEscNdx + 1]int,
	sideInfoTab []int,
	ndx1 int,
	ndx2 int,
	useVCB11 bool,
) int {
	mergeBits := sideInfoTab[huffsection[ndx1].SfbCnt+huffsection[ndx2].SfbCnt] +
		fdkaacEncFindMinMergeBits(bitLookUp[ndx1], bitLookUp[ndx2], useVCB11)
	splitBits := huffsection[ndx1].SectionBits + huffsection[ndx2].SectionBits
	mergeGain := splitBits - mergeBits
	if isNoiseOrIntensityBook(huffsection[ndx1].CodeBook) || isNoiseOrIntensityBook(huffsection[ndx2].CodeBook) {
		mergeGain = -1
	}
	return mergeGain
}

func fdkaacEncGmStage0(
	huffsection []SectionInfo,
	bitLookUp [][codeBookEscNdx + 1]int,
	maxSfb int,
	noiseNrg []int,
	isBook []int,
) {
	for i := 0; i < maxSfb; i++ {
		if huffsection[i].SectionBits == invalidBitcount {
			if noiseNrg[i] != noNoisePNS {
				huffsection[i].CodeBook = codeBookPNSNo
				huffsection[i].SectionBits = 0
			} else if isBook[i] != 0 {
				huffsection[i].CodeBook = isBook[i]
				huffsection[i].SectionBits = 0
			} else {
				huffsection[i].SectionBits = fdkaacEncFindBestBook(bitLookUp[i], &huffsection[i].CodeBook, false)
			}
		}
	}
}

func fdkaacEncGmStage1(
	huffsection []SectionInfo,
	bitLookUp [][codeBookEscNdx + 1]int,
	maxSfb int,
	sideInfoTab []int,
	useVCB11 bool,
) {
	mergeStart := 0
	for {
		mergeEnd := mergeStart + 1
		for ; mergeEnd < maxSfb; mergeEnd++ {
			if huffsection[mergeStart].CodeBook != huffsection[mergeEnd].CodeBook {
				break
			}
			huffsection[mergeStart].SfbCnt++
			huffsection[mergeStart].SectionBits += huffsection[mergeEnd].SectionBits
			fdkaacEncMergeBitLookUp(&bitLookUp[mergeStart], bitLookUp[mergeEnd])
		}
		huffsection[mergeStart].SectionBits += fdkaacEncGetSideInfoBits(&huffsection[mergeStart], sideInfoTab, useVCB11)
		huffsection[mergeEnd-1].SfbStart = huffsection[mergeStart].SfbStart
		mergeStart = mergeEnd
		if mergeStart >= maxSfb {
			break
		}
	}
}

func fdkaacEncGmStage2(
	huffsection []SectionInfo,
	mergeGainLookUp []int,
	bitLookUp [][codeBookEscNdx + 1]int,
	maxSfb int,
	sideInfoTab []int,
	useVCB11 bool,
) {
	for i := 0; i < maxSfb; i++ {
		mergeGainLookUp[i] = 0
	}
	for i := 0; i+huffsection[i].SfbCnt < maxSfb; i += huffsection[i].SfbCnt {
		mergeGainLookUp[i] = fdkaacEncCalcMergeGain(huffsection, bitLookUp, sideInfoTab, i, i+huffsection[i].SfbCnt, useVCB11)
	}

	for {
		maxNdx := 0
		maxMergeGain := fdkaacEncFindMaxMerge(mergeGainLookUp, huffsection, maxSfb, &maxNdx)
		if maxMergeGain <= 0 {
			break
		}

		maxNdxNext := maxNdx + huffsection[maxNdx].SfbCnt
		huffsection[maxNdx].SfbCnt += huffsection[maxNdxNext].SfbCnt
		huffsection[maxNdx].SectionBits += huffsection[maxNdxNext].SectionBits - maxMergeGain
		fdkaacEncMergeBitLookUp(&bitLookUp[maxNdx], bitLookUp[maxNdxNext])

		if maxNdx != 0 {
			maxNdxLast := huffsection[maxNdx-1].SfbStart
			mergeGainLookUp[maxNdxLast] = fdkaacEncCalcMergeGain(huffsection, bitLookUp, sideInfoTab, maxNdxLast, maxNdx, useVCB11)
		}
		maxNdxNext = maxNdx + huffsection[maxNdx].SfbCnt
		huffsection[maxNdxNext-1].SfbStart = huffsection[maxNdx].SfbStart

		if maxNdxNext < maxSfb {
			mergeGainLookUp[maxNdx] = fdkaacEncCalcMergeGain(huffsection, bitLookUp, sideInfoTab, maxNdx, maxNdxNext, useVCB11)
		}
	}
}

func isNoiseOrIntensityBook(codeBook int) bool {
	return codeBook == codeBookPNSNo || codeBook == codeBookISOutOfPhaseNo || codeBook == codeBookISInPhaseNo
}

func checkDynBitCountInputs(
	state *BitCounterState,
	quantSpectrum []int16,
	maxValueInSfb []uint32,
	scalefac []int,
	blockType int,
	sfbCnt int,
	maxSfbPerGroup int,
	sfbPerGroup int,
	sfbOffset []int,
	sectionData *SectionData,
	noiseNrg []int,
	isBook []int,
	isScale []int,
) {
	if state == nil {
		panic("fdkaac: nil bit-counter state")
	}
	if sectionData == nil {
		panic("fdkaac: nil section data")
	}
	if sfbCnt < 0 || sfbCnt > maxGroupedSFB || sfbPerGroup <= 0 || sfbCnt%sfbPerGroup != 0 {
		panic("fdkaac: invalid dynamic bit-count SFB shape")
	}
	if maxSfbPerGroup < 0 || maxSfbPerGroup > sfbPerGroup {
		panic("fdkaac: invalid dynamic bit-count max SFB")
	}
	sideInfoTab := fdkaacEncSideInfoTabLong[:]
	if blockType == ShortWindow {
		sideInfoTab = fdkaacEncSideInfoTabShort[:]
	}
	if maxSfbPerGroup >= len(sideInfoTab) {
		panic("fdkaac: dynamic bit-count side-info overflow")
	}
	if len(maxValueInSfb) < sfbCnt || len(noiseNrg) < sfbCnt || len(isBook) < sfbCnt || len(isScale) < sfbCnt {
		panic("fdkaac: short dynamic bit-count band state")
	}
	if scalefac != nil && len(scalefac) < sfbCnt {
		panic("fdkaac: short dynamic bit-count scalefactors")
	}
	if len(sfbOffset) < sfbCnt+1 {
		panic("fdkaac: short dynamic bit-count offsets")
	}
	if maxSfbPerGroup == 0 {
		return
	}
	maxLine := 0
	for grpNdx := 0; grpNdx < sfbCnt; grpNdx += sfbPerGroup {
		for i := 0; i < maxSfbPerGroup; i++ {
			start := sfbOffset[grpNdx+i]
			stop := sfbOffset[grpNdx+i+1]
			if start < 0 || stop < start {
				panic("fdkaac: malformed dynamic bit-count offsets")
			}
			if stop > maxLine {
				maxLine = stop
			}
		}
	}
	if len(quantSpectrum) < maxLine {
		panic("fdkaac: short dynamic bit-count spectrum")
	}
}
