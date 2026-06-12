package fdkaac

const (
	LongWindow = iota
	StartWindow
	ShortWindow
	StopWindow
)

type MDCT struct {
	overlap               []FixpDBL
	prevWRS               []FixpSPK
	prevTL                int
	prevNR                int
	prevFR                int
	ovOffset              int
	prevAliasSymmetry     int
	prevPrevAliasSymmetry int
}

func MDCTInit(h *MDCT, overlap []FixpDBL) {
	if h == nil {
		panic("fdkaac: nil MDCT state")
	}
	*h = MDCT{overlap: overlap}
}

func MDCTBlock(h *MDCT, timeData []int16, noInSamples int, mdctData []FixpDBL, nSpec, tl int, pRightWindowPart []FixpSPK, fr int, pMdctDataE []int) int {
	if h == nil {
		panic("fdkaac: nil MDCT state")
	}
	if nSpec <= 0 || tl < 4 || fr <= 0 || fr > tl || noInSamples < tl {
		panic("fdkaac: MDCT parameters not supported")
	}
	if len(mdctData) < nSpec*tl || len(pMdctDataE) < nSpec || len(pRightWindowPart) < fr/2 {
		panic("fdkaac: MDCT buffers too small")
	}

	if h.prevFR == 0 {
		h.prevFR = fr
		h.prevWRS = pRightWindowPart
		h.prevTL = tl
	}
	if len(h.prevWRS) < h.prevFR/2 {
		panic("fdkaac: MDCT previous window state invalid")
	}

	nr := (tl - fr) >> 1
	timeOff := (noInSamples - tl) >> 1
	needTime := timeOff + (nSpec-1)*tl + 2*tl
	if timeOff < 0 || len(timeData) < needTime {
		panic("fdkaac: MDCT input buffer too small")
	}

	wrs := pRightWindowPart
	mdctOff := 0
	for n := 0; n < nSpec; n++ {
		mdctDataE := 1 + 1

		wls := h.prevWRS
		fl := h.prevFR
		nl := (tl - fl) >> 1
		if fl <= 0 || fl > tl || nl < 0 || len(wls) < fl/2 {
			panic("fdkaac: MDCT previous window state invalid")
		}

		td := timeData[timeOff:]
		md := mdctData[mdctOff:]

		for i := 0; i < nl; i++ {
			md[tl/2+i] = -FixpDBL(td[tl-i-1]) << (DfractBits - FractBits - 1)
		}

		for i := 0; i < fl/2; i++ {
			tmp0 := FMultDiv2SS(FixpSGL(td[i+nl]), wls[i].Im)
			md[tl/2+i+nl] = FMultSubDiv2SS(tmp0, FixpSGL(td[tl-nl-i-1]), wls[i].Re)
		}

		for i := 0; i < nr; i++ {
			md[tl/2-1-i] = -FixpDBL(td[tl+i]) << (DfractBits - FractBits - 1)
		}

		for i := 0; i < fr/2; i++ {
			tmp1 := FMultDiv2SS(FixpSGL(td[tl+nr+i]), wrs[i].Re)
			md[tl/2-nr-i-1] = -FMultAddDiv2SS(tmp1, FixpSGL(td[(tl*2)-nr-i-1]), wrs[i].Im)
		}

		DCTIV(md, tl, &mdctDataE)
		pMdctDataE[n] = mdctDataE

		timeOff += tl
		mdctOff += tl

		h.prevWRS = wrs
		h.prevFR = fr
		h.prevTL = tl
		h.prevNR = nr
	}

	return nSpec * tl
}

func FDKaacEncTransformReal(pTimeData []int16, mdctData []FixpDBL, blockType, windowShape int, prevWindowShape *int, mdctPers *MDCT, frameLength int, pMdctDataE *int, filterType int) int {
	_ = filterType

	var numSpec int
	var numMdctLines int
	if blockType == ShortWindow {
		numSpec = 8
		numMdctLines = frameLength >> 3
	} else {
		numSpec = 1
		numMdctLines = frameLength
	}

	offset := 0
	if windowShape == WindowShapeLOL {
		offset = (frameLength * 3) >> 2
	}

	var fr int
	switch blockType {
	case LongWindow, StopWindow:
		fr = frameLength - offset
	case StartWindow, ShortWindow:
		fr = frameLength >> 3
	default:
		return -1
	}

	var mdctDataE [8]int
	MDCTBlock(mdctPers, pTimeData, frameLength, mdctData, numSpec, numMdctLines, FDKGetWindowSlope(fr, windowShape), fr, mdctDataE[:])

	if blockType == ShortWindow {
		for i := 1; i < 8; i++ {
			if mdctDataE[i] != mdctDataE[0] {
				return -1
			}
		}
	}
	if prevWindowShape != nil {
		*prevWindowShape = windowShape
	}
	if pMdctDataE != nil {
		*pMdctDataE = mdctDataE[0]
	}

	return 0
}
