package fdkaac

const (
	peConstPartShift = FractBits
	c1LdData         = FixpDBL(0x06000000)
	c2LdData         = FixpDBL(0x02a4d3c3)
	c3LdData         = FixpDBL(0x4799051f)
)

type PEChannelData struct {
	SfbNLines       [maxGroupedSFB]int
	SfbPe           [maxGroupedSFB]FixpDBL
	SfbConstPart    [maxGroupedSFB]FixpDBL
	SfbNActiveLines [maxGroupedSFB]FixpDBL
	Pe              FixpDBL
	ConstPart       FixpDBL
	NActiveLines    FixpDBL
}

type PEData struct {
	PEChannelData [2]PEChannelData
	Pe            FixpDBL
	ConstPart     FixpDBL
	NActiveLines  FixpDBL
	Offset        int
}

func FDKaacEncPrepareSfbPe(
	peChanData *PEChannelData,
	sfbEnergyLdData []FixpDBL,
	sfbThresholdLdData []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbOffset []int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
) {
	checkPrepareSfbPeInputs(
		peChanData, sfbEnergyLdData, sfbThresholdLdData, sfbFormFactorLdData,
		sfbOffset, sfbCnt, sfbPerGroup, maxSfbPerGroup,
	)

	for sfbGrp := 0; sfbGrp < sfbCnt; sfbGrp += sfbPerGroup {
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			idx := sfbGrp + sfb
			if sfbEnergyLdData[idx] > sfbThresholdLdData[idx] {
				sfbWidth := sfbOffset[idx+1] - sfbOffset[idx]
				avgFormFactorLdData := (((-sfbEnergyLdData[idx] >> 1) + (CalcLdInt(sfbWidth) >> 1)) >> 1)
				nLines := int(CalcInvLdData(sfbFormFactorLdData[idx] + formFactorLdScale + avgFormFactorLdData))
				peChanData.SfbNLines[idx] = minInt(sfbWidth, nLines)
			} else {
				peChanData.SfbNLines[idx] = 0
			}
		}
	}
}

func FDKaacEncCalcSfbPe(
	peChanData *PEChannelData,
	sfbEnergyLdData []FixpDBL,
	sfbThresholdLdData []FixpDBL,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	isBook []int,
	isScale []int,
) {
	checkCalcSfbPeInputs(peChanData, sfbEnergyLdData, sfbThresholdLdData, sfbCnt, sfbPerGroup, maxSfbPerGroup, isBook, isScale)

	lastValIs := 0
	pe := FixpDBL(0)
	constPart := FixpDBL(0)
	nActiveLines := FixpDBL(0)

	for sfbGrp := 0; sfbGrp < sfbCnt; sfbGrp += sfbPerGroup {
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			thisSfb := sfbGrp + sfb
			tmpPe := FixpDBL(0)
			tmpConstPart := FixpDBL(0)
			tmpNActiveLines := FixpDBL(0)

			if sfbEnergyLdData[thisSfb] > sfbThresholdLdData[thisSfb] {
				logDataRatio := sfbEnergyLdData[thisSfb] - sfbThresholdLdData[thisSfb]
				nLines := peChanData.SfbNLines[thisSfb]
				factor := FixpDBL(nLines << (ldDataShift + peConstPartShift + 1))
				if logDataRatio >= c1LdData {
					tmpPe = FMultDiv2DD(logDataRatio, factor)
					tmpConstPart = FMultDiv2DD(sfbEnergyLdData[thisSfb], factor)
				} else {
					tmpPe = FMultDiv2DD(c2LdData+FMultDD(c3LdData, logDataRatio), factor)
					tmpConstPart = FMultDiv2DD(c2LdData+FMultDD(c3LdData, sfbEnergyLdData[thisSfb]), factor)
					nLines = FMultI(c3LdData, nLines)
				}
				tmpNActiveLines = FixpDBL(nLines)
			} else if isBook[thisSfb] != 0 {
				delta := isScale[thisSfb] - lastValIs
				lastValIs = isScale[thisSfb]
				peChanData.SfbPe[thisSfb] = FixpDBL(FDKaacEncBitCountScalefactorDelta(delta) << peConstPartShift)
				peChanData.SfbConstPart[thisSfb] = 0
				peChanData.SfbNActiveLines[thisSfb] = 0
			}

			peChanData.SfbPe[thisSfb] = tmpPe
			peChanData.SfbConstPart[thisSfb] = tmpConstPart
			peChanData.SfbNActiveLines[thisSfb] = tmpNActiveLines

			pe += tmpPe
			constPart += tmpConstPart
			nActiveLines += tmpNActiveLines
		}
	}

	peChanData.Pe = pe >> peConstPartShift
	peChanData.ConstPart = constPart >> peConstPartShift
	peChanData.NActiveLines = nActiveLines
}

func checkPrepareSfbPeInputs(
	peChanData *PEChannelData,
	sfbEnergyLdData []FixpDBL,
	sfbThresholdLdData []FixpDBL,
	sfbFormFactorLdData []FixpDBL,
	sfbOffset []int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
) {
	checkLinePeBands(peChanData, sfbCnt, sfbPerGroup, maxSfbPerGroup)
	if len(sfbEnergyLdData) < sfbCnt || len(sfbThresholdLdData) < sfbCnt || len(sfbFormFactorLdData) < sfbCnt {
		panic("fdkaac: short PE prepare data")
	}
	checkGroupedSfbOffsets(
		sfbOffset,
		sfbCnt,
		sfbPerGroup,
		maxSfbPerGroup,
		false,
		"fdkaac: invalid PE prepare offset",
		"fdkaac: short PE prepare offsets",
	)
	for sfbGrp := 0; sfbGrp < sfbCnt; sfbGrp += sfbPerGroup {
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			idx := sfbGrp + sfb
			if sfbEnergyLdData[idx] > sfbThresholdLdData[idx] && sfbOffset[idx+1] == sfbOffset[idx] {
				panic("fdkaac: empty PE prepare active band")
			}
		}
	}
}

func checkCalcSfbPeInputs(
	peChanData *PEChannelData,
	sfbEnergyLdData []FixpDBL,
	sfbThresholdLdData []FixpDBL,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	isBook []int,
	isScale []int,
) {
	checkLinePeBands(peChanData, sfbCnt, sfbPerGroup, maxSfbPerGroup)
	if len(sfbEnergyLdData) < sfbCnt || len(sfbThresholdLdData) < sfbCnt || len(isBook) < sfbCnt || len(isScale) < sfbCnt {
		panic("fdkaac: short PE data")
	}
}

func checkLinePeBands(peChanData *PEChannelData, sfbCnt int, sfbPerGroup int, maxSfbPerGroup int) {
	if peChanData == nil {
		panic("fdkaac: nil PE channel data")
	}
	if sfbCnt <= 0 || sfbCnt > maxGroupedSFB || sfbPerGroup <= 0 || sfbCnt%sfbPerGroup != 0 {
		panic("fdkaac: invalid PE band count")
	}
	if maxSfbPerGroup <= 0 || maxSfbPerGroup > sfbPerGroup {
		panic("fdkaac: invalid PE group width")
	}
}
