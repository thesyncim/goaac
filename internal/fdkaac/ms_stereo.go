package fdkaac

const (
	MsMaskNone = 0
	MsMaskSome = 1
	MsMaskAll  = 2
)

func FDKaacEncMsStereoProcessing(
	mdctSpectrumLeft []FixpDBL,
	mdctSpectrumRight []FixpDBL,
	sfbEnergyLeft []FixpDBL,
	sfbEnergyRight []FixpDBL,
	sfbEnergyMid []FixpDBL,
	sfbEnergySide []FixpDBL,
	sfbThresholdLeft []FixpDBL,
	sfbThresholdRight []FixpDBL,
	sfbSpreadEnLeft []FixpDBL,
	sfbSpreadEnRight []FixpDBL,
	sfbEnergyLeftLdData []FixpDBL,
	sfbEnergyRightLdData []FixpDBL,
	sfbEnergyMidLdData []FixpDBL,
	sfbEnergySideLdData []FixpDBL,
	sfbThresholdLeftLdData []FixpDBL,
	sfbThresholdRightLdData []FixpDBL,
	isBook []int,
	msDigest *int,
	msMask []int,
	allowMS int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	sfbOffset []int,
) {
	checkMsStereoInputs(
		mdctSpectrumLeft, mdctSpectrumRight, sfbEnergyLeft, sfbEnergyRight,
		sfbEnergyMid, sfbEnergySide, sfbThresholdLeft, sfbThresholdRight,
		sfbSpreadEnLeft, sfbSpreadEnRight, sfbEnergyLeftLdData,
		sfbEnergyRightLdData, sfbEnergyMidLdData, sfbEnergySideLdData,
		sfbThresholdLeftLdData, sfbThresholdRightLdData, isBook, msDigest,
		msMask, sfbCnt, sfbPerGroup, maxSfbPerGroup, sfbOffset,
	)

	msMaskTrueSomewhere := 0
	numMsMaskFalse := 0

	for sfb := 0; sfb < sfbCnt; sfb += sfbPerGroup {
		for sfboffs := 0; sfboffs < maxSfbPerGroup; sfboffs++ {
			idx := sfb + sfboffs
			if isBook == nil || isBook[idx] == 0 {
				minThresholdLdData := minFixpDBL(sfbThresholdLeftLdData[idx], sfbThresholdRightLdData[idx])

				tmp := maxFixpDBL(sfbEnergyLeftLdData[idx], sfbThresholdLeftLdData[idx])
				pnlrLdData := (sfbThresholdLeftLdData[idx] >> 1) - (tmp >> 1)
				pnlrLdData += sfbThresholdRightLdData[idx] >> 1
				tmp = maxFixpDBL(sfbEnergyRightLdData[idx], sfbThresholdRightLdData[idx])
				pnlrLdData -= tmp >> 1

				tmp = maxFixpDBL(sfbEnergyMidLdData[idx], minThresholdLdData)
				pnmsLdData := minThresholdLdData - (tmp >> 1)
				tmp = maxFixpDBL(sfbEnergySideLdData[idx], minThresholdLdData)
				pnmsLdData -= tmp >> 1

				useMS := 0
				if allowMS != 0 && pnmsLdData > pnlrLdData {
					useMS = 1
				}

				if useMS != 0 {
					msMask[idx] = 1
					msMaskTrueSomewhere = 1
					applyMsStereoBand(
						idx, sfbOffset, mdctSpectrumLeft, mdctSpectrumRight,
						sfbEnergyLeft, sfbEnergyRight, sfbEnergyMid, sfbEnergySide,
						sfbThresholdLeft, sfbThresholdRight, sfbSpreadEnLeft,
						sfbSpreadEnRight, sfbEnergyLeftLdData, sfbEnergyRightLdData,
						sfbEnergyMidLdData, sfbEnergySideLdData,
						sfbThresholdLeftLdData, sfbThresholdRightLdData,
					)
				} else {
					msMask[idx] = 0
					numMsMaskFalse++
				}
			} else {
				if msMask[idx] != 0 {
					msMaskTrueSomewhere = 1
				}
				numMsMaskFalse = 9
			}
		}
	}

	if msMaskTrueSomewhere == 1 {
		if numMsMaskFalse == 0 || (numMsMaskFalse < maxSfbPerGroup && numMsMaskFalse < 9) {
			*msDigest = MsMaskAll
			for sfb := 0; sfb < sfbCnt; sfb += sfbPerGroup {
				for sfboffs := 0; sfboffs < maxSfbPerGroup; sfboffs++ {
					idx := sfb + sfboffs
					if (isBook == nil || isBook[idx] == 0) && msMask[idx] == 0 {
						msMask[idx] = 1
						applyMsStereoBand(
							idx, sfbOffset, mdctSpectrumLeft, mdctSpectrumRight,
							sfbEnergyLeft, sfbEnergyRight, sfbEnergyMid, sfbEnergySide,
							sfbThresholdLeft, sfbThresholdRight, sfbSpreadEnLeft,
							sfbSpreadEnRight, sfbEnergyLeftLdData, sfbEnergyRightLdData,
							sfbEnergyMidLdData, sfbEnergySideLdData,
							sfbThresholdLeftLdData, sfbThresholdRightLdData,
						)
					}
				}
			}
		} else {
			*msDigest = MsMaskSome
		}
	} else {
		*msDigest = MsMaskNone
	}
}

func applyMsStereoBand(
	idx int,
	sfbOffset []int,
	mdctSpectrumLeft []FixpDBL,
	mdctSpectrumRight []FixpDBL,
	sfbEnergyLeft []FixpDBL,
	sfbEnergyRight []FixpDBL,
	sfbEnergyMid []FixpDBL,
	sfbEnergySide []FixpDBL,
	sfbThresholdLeft []FixpDBL,
	sfbThresholdRight []FixpDBL,
	sfbSpreadEnLeft []FixpDBL,
	sfbSpreadEnRight []FixpDBL,
	sfbEnergyLeftLdData []FixpDBL,
	sfbEnergyRightLdData []FixpDBL,
	sfbEnergyMidLdData []FixpDBL,
	sfbEnergySideLdData []FixpDBL,
	sfbThresholdLeftLdData []FixpDBL,
	sfbThresholdRightLdData []FixpDBL,
) {
	for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
		specL := mdctSpectrumLeft[j] >> 1
		specR := mdctSpectrumRight[j] >> 1
		mdctSpectrumLeft[j] = specL + specR
		mdctSpectrumRight[j] = specL - specR
	}

	minThreshold := minFixpDBL(sfbThresholdLeft[idx], sfbThresholdRight[idx])
	sfbThresholdLeft[idx] = minThreshold
	sfbThresholdRight[idx] = minThreshold

	minThresholdLdData := minFixpDBL(sfbThresholdLeftLdData[idx], sfbThresholdRightLdData[idx])
	sfbThresholdLeftLdData[idx] = minThresholdLdData
	sfbThresholdRightLdData[idx] = minThresholdLdData

	sfbEnergyLeft[idx] = sfbEnergyMid[idx]
	sfbEnergyRight[idx] = sfbEnergySide[idx]
	sfbEnergyLeftLdData[idx] = sfbEnergyMidLdData[idx]
	sfbEnergyRightLdData[idx] = sfbEnergySideLdData[idx]

	spreadEnergy := minFixpDBL(sfbSpreadEnLeft[idx], sfbSpreadEnRight[idx]) >> 1
	sfbSpreadEnLeft[idx] = spreadEnergy
	sfbSpreadEnRight[idx] = spreadEnergy
}

func checkMsStereoInputs(
	mdctSpectrumLeft []FixpDBL,
	mdctSpectrumRight []FixpDBL,
	sfbEnergyLeft []FixpDBL,
	sfbEnergyRight []FixpDBL,
	sfbEnergyMid []FixpDBL,
	sfbEnergySide []FixpDBL,
	sfbThresholdLeft []FixpDBL,
	sfbThresholdRight []FixpDBL,
	sfbSpreadEnLeft []FixpDBL,
	sfbSpreadEnRight []FixpDBL,
	sfbEnergyLeftLdData []FixpDBL,
	sfbEnergyRightLdData []FixpDBL,
	sfbEnergyMidLdData []FixpDBL,
	sfbEnergySideLdData []FixpDBL,
	sfbThresholdLeftLdData []FixpDBL,
	sfbThresholdRightLdData []FixpDBL,
	isBook []int,
	msDigest *int,
	msMask []int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	sfbOffset []int,
) {
	if msDigest == nil {
		panic("fdkaac: nil ms digest")
	}
	if sfbCnt <= 0 || sfbCnt > maxGroupedSFB || sfbPerGroup <= 0 || sfbCnt%sfbPerGroup != 0 {
		panic("fdkaac: invalid ms stereo band count")
	}
	if maxSfbPerGroup <= 0 || maxSfbPerGroup > sfbPerGroup {
		panic("fdkaac: invalid ms stereo group width")
	}
	checkGroupedSfbOffsets(
		sfbOffset,
		sfbCnt,
		sfbPerGroup,
		maxSfbPerGroup,
		false,
		"fdkaac: invalid ms stereo offset",
		"fdkaac: short ms stereo offsets",
		len(mdctSpectrumLeft),
		len(mdctSpectrumRight),
	)
	if isBook != nil && len(isBook) < sfbCnt {
		panic("fdkaac: short ms stereo is-book")
	}
	if len(msMask) < sfbCnt {
		panic("fdkaac: short ms stereo mask")
	}
	if len(sfbEnergyLeft) < sfbCnt || len(sfbEnergyRight) < sfbCnt || len(sfbEnergyMid) < sfbCnt || len(sfbEnergySide) < sfbCnt {
		panic("fdkaac: short ms stereo energy")
	}
	if len(sfbThresholdLeft) < sfbCnt || len(sfbThresholdRight) < sfbCnt {
		panic("fdkaac: short ms stereo threshold")
	}
	if len(sfbSpreadEnLeft) < sfbCnt || len(sfbSpreadEnRight) < sfbCnt {
		panic("fdkaac: short ms stereo spread energy")
	}
	if len(sfbEnergyLeftLdData) < sfbCnt || len(sfbEnergyRightLdData) < sfbCnt || len(sfbEnergyMidLdData) < sfbCnt || len(sfbEnergySideLdData) < sfbCnt {
		panic("fdkaac: short ms stereo energy ld-data")
	}
	if len(sfbThresholdLeftLdData) < sfbCnt || len(sfbThresholdRightLdData) < sfbCnt {
		panic("fdkaac: short ms stereo threshold ld-data")
	}
}

func minFixpDBL(a, b FixpDBL) FixpDBL {
	if a < b {
		return a
	}
	return b
}

func maxFixpDBL(a, b FixpDBL) FixpDBL {
	if a > b {
		return a
	}
	return b
}
