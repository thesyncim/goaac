package fdkaac

type PsyOutChannel struct {
	SfbCnt             int
	SfbPerGroup        int
	MaxSfbPerGroup     int
	LastWindowSequence int
	WindowShape        int
	GroupingMask       int
	SfbOffsets         [maxGroupedSFB + 1]int
	MdctScale          int
	GroupLen           [maxNoOfGroups]int
	TNSInfo            TNSInfo
	NoiseNrg           [maxGroupedSFB]int
	IsScale            [maxGroupedSFB]int
	MdctSpectrum       []FixpDBL
	SfbEnergy          [maxGroupedSFB]FixpDBL
	SfbSpreadEnergy    [maxGroupedSFB]FixpDBL
}

func FDKaacEncBuildPsyOutChannel(
	out *PsyOutChannel,
	sfbEnergy *SFBEnergy,
	sfbSpreadEnergy *SFBEnergy,
	blockSwitchingControl *BlockSwitchingControl,
	maxSfbPerGroup int,
	mdctScale int,
	isShortWindow bool,
	longSfbActive int,
	shortSfbCnt int,
) {
	checkBuildPsyOutInputs(out, sfbEnergy, sfbSpreadEnergy, blockSwitchingControl, maxSfbPerGroup, isShortWindow, longSfbActive, shortSfbCnt)

	out.MaxSfbPerGroup = maxSfbPerGroup
	out.MdctScale = mdctScale

	if !isShortWindow {
		out.SfbCnt = longSfbActive
		out.SfbPerGroup = longSfbActive
		out.LastWindowSequence = blockSwitchingControl.LastWindowSequence
		out.WindowShape = blockSwitchingControl.WindowShape
	} else {
		sfbCnt := blockSwitchingControl.NoOfGroups * shortSfbCnt
		out.SfbCnt = sfbCnt
		out.SfbPerGroup = shortSfbCnt
		out.LastWindowSequence = ShortWindow
		out.WindowShape = WindowShapeSine
	}

	mask := 0
	for grp := 0; grp < blockSwitchingControl.NoOfGroups; grp++ {
		mask <<= 1
		for j := 1; j < blockSwitchingControl.GroupLen[grp]; j++ {
			mask = (mask << 1) | 1
		}
	}
	out.GroupingMask = mask

	out.GroupLen = blockSwitchingControl.GroupLen
	copy(out.SfbEnergy[:], sfbEnergy.Long[:])
	copy(out.SfbSpreadEnergy[:], sfbSpreadEnergy.Long[:])
}

func checkBuildPsyOutInputs(
	out *PsyOutChannel,
	sfbEnergy *SFBEnergy,
	sfbSpreadEnergy *SFBEnergy,
	blockSwitchingControl *BlockSwitchingControl,
	maxSfbPerGroup int,
	isShortWindow bool,
	longSfbActive int,
	shortSfbCnt int,
) {
	if out == nil {
		panic("fdkaac: nil psy output")
	}
	if sfbEnergy == nil || sfbSpreadEnergy == nil {
		panic("fdkaac: nil psy output energy")
	}
	if blockSwitchingControl == nil {
		panic("fdkaac: nil psy output block switching")
	}
	if blockSwitchingControl.NoOfGroups <= 0 || blockSwitchingControl.NoOfGroups > maxNoOfGroups {
		panic("fdkaac: invalid psy output group count")
	}
	groupSum := 0
	for i := 0; i < blockSwitchingControl.NoOfGroups; i++ {
		if blockSwitchingControl.GroupLen[i] <= 0 {
			panic("fdkaac: invalid psy output group length")
		}
		groupSum += blockSwitchingControl.GroupLen[i]
	}
	if isShortWindow {
		if blockSwitchingControl.LastWindowSequence != ShortWindow {
			panic("fdkaac: invalid short psy output sequence")
		}
		if groupSum != transFac {
			panic("fdkaac: invalid short psy output grouping")
		}
		if shortSfbCnt <= 0 || shortSfbCnt > maxSFBShort {
			panic("fdkaac: invalid short psy output sfb count")
		}
		if blockSwitchingControl.NoOfGroups*shortSfbCnt > maxGroupedSFB {
			panic("fdkaac: short psy output too large")
		}
		if maxSfbPerGroup <= 0 || maxSfbPerGroup > shortSfbCnt {
			panic("fdkaac: invalid short psy output max sfb")
		}
		return
	}
	if blockSwitchingControl.LastWindowSequence == ShortWindow {
		panic("fdkaac: invalid long psy output sequence")
	}
	if blockSwitchingControl.NoOfGroups != 1 || blockSwitchingControl.GroupLen[0] != 1 {
		panic("fdkaac: invalid long psy output grouping")
	}
	if longSfbActive <= 0 || longSfbActive > maxGroupedSFB {
		panic("fdkaac: invalid long psy output sfb count")
	}
	if maxSfbPerGroup <= 0 || maxSfbPerGroup > longSfbActive {
		panic("fdkaac: invalid long psy output max sfb")
	}
}
