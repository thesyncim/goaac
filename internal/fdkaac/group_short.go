package fdkaac

const (
	transFac      = 8
	maxSFBShort   = 15
	maxGroupedSFB = 60
)

type SFBThreshold struct {
	Long  [maxGroupedSFB]FixpDBL
	Short [transFac][maxSFBShort]FixpDBL
}

type SFBEnergy struct {
	Long  [maxGroupedSFB]FixpDBL
	Short [transFac][maxSFBShort]FixpDBL
}

type GroupShortScratch struct {
	Spectrum [1024]FixpDBL
}

func FDKaacEncGroupShortData(
	mdctSpectrum []FixpDBL,
	sfbThreshold *SFBThreshold,
	sfbEnergy *SFBEnergy,
	sfbEnergyMS *SFBEnergy,
	sfbSpreadEnergy *SFBEnergy,
	sfbCnt int,
	sfbActive int,
	sfbOffset []int,
	sfbMinSnrLdData []FixpDBL,
	groupedSfbOffset []int,
	maxSfbPerGroup *int,
	groupedSfbMinSnrLdData []FixpDBL,
	noOfGroups int,
	groupLen []int,
	granuleLength int,
	scratch *GroupShortScratch,
) {
	checkGroupShortInputs(
		mdctSpectrum, sfbThreshold, sfbEnergy, sfbEnergyMS, sfbSpreadEnergy,
		sfbCnt, sfbActive, sfbOffset, sfbMinSnrLdData, groupedSfbOffset,
		maxSfbPerGroup, groupedSfbMinSnrLdData, noOfGroups, groupLen,
		granuleLength, scratch,
	)

	granuleLengthShort := granuleLength / transFac

	highestSfb := 0
	for wnd := 0; wnd < transFac; wnd++ {
		sfb := sfbActive - 1
		for ; sfb >= highestSfb; sfb-- {
			line := sfbOffset[sfb+1] - 1
			for ; line >= sfbOffset[sfb]; line-- {
				if mdctSpectrum[wnd*granuleLengthShort+line] != 0 {
					break
				}
			}
			if line >= sfbOffset[sfb] {
				break
			}
		}
		if sfb > highestSfb {
			highestSfb = sfb
		}
	}
	if highestSfb < 0 {
		highestSfb = 0
	}
	*maxSfbPerGroup = highestSfb + 1

	i := 0
	offset := 0
	for grp := 0; grp < noOfGroups; grp++ {
		sfb := 0
		for ; sfb < sfbActive+1; sfb++ {
			groupedSfbOffset[i] = offset + sfbOffset[sfb]*groupLen[grp]
			i++
		}
		i += sfbCnt - sfb
		offset += groupLen[grp] * granuleLengthShort
	}
	groupedSfbOffset[i] = granuleLength

	i = 0
	for grp := 0; grp < noOfGroups; grp++ {
		sfb := 0
		for ; sfb < sfbActive; sfb++ {
			groupedSfbMinSnrLdData[i] = sfbMinSnrLdData[sfb]
			i++
		}
		i += sfbCnt - sfb
	}

	groupShortEnergy(sfbThreshold.Long[:], &sfbThreshold.Short, sfbCnt, sfbActive, noOfGroups, groupLen)
	groupShortEnergy(sfbEnergy.Long[:], &sfbEnergy.Short, sfbCnt, sfbActive, noOfGroups, groupLen)
	groupShortEnergy(sfbEnergyMS.Long[:], &sfbEnergyMS.Short, sfbCnt, sfbActive, noOfGroups, groupLen)
	groupShortEnergy(sfbSpreadEnergy.Long[:], &sfbSpreadEnergy.Short, sfbCnt, sfbActive, noOfGroups, groupLen)

	tmpSpectrum := scratch.Spectrum[:granuleLength]
	wnd := 0
	i = 0
	for grp := 0; grp < noOfGroups; grp++ {
		sfb := 0
		for ; sfb < sfbActive; sfb++ {
			width := sfbOffset[sfb+1] - sfbOffset[sfb]
			mdctBase := wnd*granuleLengthShort + sfbOffset[sfb]
			for j := 0; j < groupLen[grp]; j++ {
				p := mdctBase + j*granuleLengthShort
				for line := 0; line < width; line++ {
					tmpSpectrum[i] = mdctSpectrum[p+line]
					i++
				}
			}
		}
		i += groupLen[grp] * (sfbOffset[sfbCnt] - sfbOffset[sfb])
		wnd += groupLen[grp]
	}

	copy(mdctSpectrum[:granuleLength], tmpSpectrum)
}

func groupShortEnergy(dst []FixpDBL, src *[transFac][maxSFBShort]FixpDBL, sfbCnt int, sfbActive int, noOfGroups int, groupLen []int) {
	wnd := 0
	i := 0
	for grp := 0; grp < noOfGroups; grp++ {
		sfb := 0
		for ; sfb < sfbActive; sfb++ {
			energy := src[wnd][sfb]
			for j := 1; j < groupLen[grp]; j++ {
				energy = nrgAddSaturate(energy, src[wnd+j][sfb])
			}
			dst[i] = energy
			i++
		}
		i += sfbCnt - sfb
		wnd += groupLen[grp]
	}
}

func nrgAddSaturate(a, b FixpDBL) FixpDBL {
	if int64(a) >= int64(MaxValDBL)-int64(b) {
		return MaxValDBL
	}
	return a + b
}

func checkGroupShortInputs(
	mdctSpectrum []FixpDBL,
	sfbThreshold *SFBThreshold,
	sfbEnergy *SFBEnergy,
	sfbEnergyMS *SFBEnergy,
	sfbSpreadEnergy *SFBEnergy,
	sfbCnt int,
	sfbActive int,
	sfbOffset []int,
	sfbMinSnrLdData []FixpDBL,
	groupedSfbOffset []int,
	maxSfbPerGroup *int,
	groupedSfbMinSnrLdData []FixpDBL,
	noOfGroups int,
	groupLen []int,
	granuleLength int,
	scratch *GroupShortScratch,
) {
	if sfbThreshold == nil || sfbEnergy == nil || sfbEnergyMS == nil || sfbSpreadEnergy == nil {
		panic("fdkaac: nil group-short energy")
	}
	if maxSfbPerGroup == nil || scratch == nil {
		panic("fdkaac: nil group-short output")
	}
	if sfbCnt <= 0 || sfbCnt > maxSFBShort {
		panic("fdkaac: invalid short sfb count")
	}
	if sfbActive != sfbCnt {
		panic("fdkaac: inactive short grouping bands unsupported")
	}
	if len(sfbOffset) < sfbCnt+1 {
		panic("fdkaac: short sfb offsets")
	}
	if granuleLength <= 0 || granuleLength%transFac != 0 || granuleLength > len(mdctSpectrum) || granuleLength > len(scratch.Spectrum) {
		panic("fdkaac: invalid grouped granule length")
	}
	granuleLengthShort := granuleLength / transFac
	prev := sfbOffset[0]
	if prev != 0 {
		panic("fdkaac: invalid short sfb offset")
	}
	for i := 0; i < sfbCnt; i++ {
		next := sfbOffset[i+1]
		if next < prev || next > granuleLengthShort {
			panic("fdkaac: invalid short sfb offset")
		}
		prev = next
	}
	if sfbOffset[sfbCnt] != granuleLengthShort {
		panic("fdkaac: incomplete short sfb offsets")
	}
	if len(sfbMinSnrLdData) < sfbCnt {
		panic("fdkaac: short min-snr vector")
	}
	if noOfGroups <= 0 || noOfGroups > maxNoOfGroups || len(groupLen) < noOfGroups {
		panic("fdkaac: invalid short group count")
	}
	sumGroups := 0
	for i := 0; i < noOfGroups; i++ {
		if groupLen[i] <= 0 {
			panic("fdkaac: invalid short group length")
		}
		sumGroups += groupLen[i]
	}
	if sumGroups != transFac {
		panic("fdkaac: invalid short group length")
	}
	groupedBands := noOfGroups*sfbCnt + 1
	if len(groupedSfbOffset) < groupedBands {
		panic("fdkaac: short grouped sfb offset")
	}
	if len(groupedSfbMinSnrLdData) < groupedBands-1 {
		panic("fdkaac: short grouped min-snr vector")
	}
}
