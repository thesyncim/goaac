package fdkaac

const (
	isCorrThresh                 = FixpDBL(0x7999999a)
	isTotalErrorThresh           = FixpDBL(0x051eb852)
	isLocalErrorThresh           = FixpDBL(0x0147ae14)
	isDirectionDeviationThresh   = FixpDBL(0x40000000)
	isRegionMinLoudness          = FixpDBL(0x0ccccccd)
	isMinSFBs                    = 6
	isLeftRightRatioThresh       = FixpDBL(0x5999999a)
	realScaleSF                  = 1
	overallLoudnessSF            = 6
	maxSfbPerGroupSF             = 6
	mdctSpecSF                   = 6
	isDirectionDeviationThreshSF = 2
	isRealScaleMax               = FixpDBL(0x3c000000)
	isRealScaleRound             = FixpDBL(0x00400000)
)

var invCount = [80]FixpDBL{
	0x00000000, 0x7fffffff, 0x40000000, 0x2aaaaaab, 0x20000000, 0x1999999a,
	0x15555555, 0x12492492, 0x10000000, 0x0e38e38e, 0x0ccccccd, 0x0ba2e8ba,
	0x0aaaaaab, 0x09d89d8a, 0x09249249, 0x08888889, 0x08000000, 0x07878788,
	0x071c71c7, 0x06bca1af, 0x06666666, 0x06186186, 0x05d1745d, 0x0590b216,
	0x05555555, 0x051eb852, 0x04ec4ec5, 0x04bda12f, 0x04924925, 0x0469ee58,
	0x04444444, 0x04210842, 0x04000000, 0x03e0f83e, 0x03c3c3c4, 0x03a83a84,
	0x038e38e4, 0x03759f23, 0x035e50d8, 0x03483483, 0x03333333, 0x031f3832,
	0x030c30c3, 0x02fa0be8, 0x02e8ba2f, 0x02d82d83, 0x02c8590b, 0x02b93105,
	0x02aaaaab, 0x029cbc15, 0x028f5c29, 0x02828283, 0x02762762, 0x026a439f,
	0x025ed098, 0x0253c825, 0x02492492, 0x023ee090, 0x0234f72c, 0x022b63cc,
	0x02222222, 0x02192e2a, 0x02108421, 0x02082082, 0x02000000, 0x01f81f82,
	0x01f07c1f, 0x01e9131b, 0x01e1e1e2, 0x01dae607, 0x01d41d42, 0x01cd8569,
	0x01c71c72, 0x01c0e070, 0x01bacf91, 0x01b4e81b, 0x01af286c, 0x01a98ef6,
	0x01a41a42, 0x019ec8e9,
}

type intensityParameters struct {
	CorrThresh               FixpDBL
	TotalErrorThresh         FixpDBL
	LocalErrorThresh         FixpDBL
	DirectionDeviationThresh FixpDBL
	IsRegionMinLoudness      FixpDBL
	MinISSFBs                int
	LeftRightRatioThreshold  FixpDBL
}

func FDKaacEncIntensityStereoProcessing(
	sfbEnergyLeft []FixpDBL,
	sfbEnergyRight []FixpDBL,
	mdctSpectrumLeft []FixpDBL,
	mdctSpectrumRight []FixpDBL,
	sfbThresholdLeft []FixpDBL,
	sfbThresholdRight []FixpDBL,
	sfbThresholdLdDataRight []FixpDBL,
	sfbSpreadEnLeft []FixpDBL,
	sfbSpreadEnRight []FixpDBL,
	sfbEnergyLdDataLeft []FixpDBL,
	sfbEnergyLdDataRight []FixpDBL,
	msDigest *int,
	msMask []int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	sfbOffset []int,
	allowIS int,
	isBook []int,
	isScale []int,
	pnsData []*PNSData,
) {
	checkIntensityStereoInputs(
		sfbEnergyLeft, sfbEnergyRight, mdctSpectrumLeft, mdctSpectrumRight,
		sfbThresholdLeft, sfbThresholdRight, sfbThresholdLdDataRight,
		sfbSpreadEnLeft, sfbSpreadEnRight, sfbEnergyLdDataLeft,
		sfbEnergyLdDataRight, msDigest, msMask, sfbCnt, sfbPerGroup,
		maxSfbPerGroup, sfbOffset, isBook, isScale, pnsData,
	)

	clear(isBook[:sfbCnt])
	clear(isScale[:sfbCnt])
	if allowIS == 0 {
		return
	}

	var hrrErr [maxGroupedSFB]FixpDBL
	var normSfbLoudness [maxGroupedSFB]FixpDBL
	var realIsScale [maxGroupedSFB]FixpDBL
	var isMask [maxGroupedSFB]int
	isParams := intensityParameters{
		CorrThresh:               isCorrThresh,
		TotalErrorThresh:         isTotalErrorThresh,
		LocalErrorThresh:         isLocalErrorThresh,
		DirectionDeviationThresh: isDirectionDeviationThresh,
		IsRegionMinLoudness:      isRegionMinLoudness,
		MinISSFBs:                isMinSFBs,
		LeftRightRatioThreshold:  isLeftRightRatioThresh,
	}

	fdkaacEncPrepareIntensityDecision(
		sfbEnergyLeft, sfbEnergyRight, sfbEnergyLdDataLeft, sfbEnergyLdDataRight,
		mdctSpectrumLeft, mdctSpectrumRight, &isParams, hrrErr[:], isMask[:],
		realIsScale[:], normSfbLoudness[:], sfbCnt, sfbPerGroup,
		maxSfbPerGroup, sfbOffset,
	)
	fdkaacEncFinalizeIntensityDecision(
		hrrErr[:], isMask[:], realIsScale[:], normSfbLoudness[:], &isParams,
		sfbCnt, sfbPerGroup, maxSfbPerGroup,
	)

	for sfb := 0; sfb < sfbCnt; sfb += sfbPerGroup {
		for sfboffs := 0; sfboffs < maxSfbPerGroup; sfboffs++ {
			idx := sfb + sfboffs
			mdctSpecShift := mdctSpecSF

			msMask[idx] = 0
			if isMask[idx] == 0 {
				continue
			}
			if sfbEnergyLeft[idx] < sfbThresholdLeft[idx] &&
				FMultDD(0x55555555, sfbEnergyRight[idx]) > sfbThresholdRight[idx] {
				continue
			}
			if len(pnsData) > 0 && pnsData[0] != nil {
				pnsData[0].PNSFlag[idx] = 0
				if len(pnsData) > 1 && pnsData[1] != nil {
					pnsData[1].PNSFlag[idx] = 0
				}
			}

			width := sfbOffset[idx+1] - sfbOffset[idx]
			if width > 1<<mdctSpecShift {
				mdctSpecShift++
			}

			invN := getInvInt(width >> 1)
			sL := calcIntensitySfbMaxScale(mdctSpectrumLeft, sfbOffset[idx], sfbOffset[idx+1])
			sR := calcIntensitySfbMaxScale(mdctSpectrumRight, sfbOffset[idx], sfbOffset[idx+1])

			lr := FixpDBL(0)
			for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
				lr += FMultDiv2DD(FMultDiv2DD(mdctSpectrumLeft[j]<<uint(sL), mdctSpectrumRight[j]<<uint(sR)), invN)
			}
			lr <<= 1

			if lr < 0 {
				s0 := minInt(sL, sR)
				ed := FixpDBL(0)
				for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
					d := ((mdctSpectrumLeft[j] << uint(s0)) >> 1) - ((mdctSpectrumRight[j] << uint(s0)) >> 1)
					ed += FMultDiv2DD(d, d) >> uint(mdctSpecShift-1)
				}
				msMask[idx] = 1
				tmp, s1 := fDivNormExp(sfbEnergyLeft[idx], ed)
				s2 := s1 + (2 * s0) - 2 - mdctSpecShift
				if s2&1 != 0 {
					tmp >>= 1
					s2++
				}
				s2 = (s2 >> 1) + 1
				s2 = minInt(maxInt(s2, -(DfractBits-1)), DfractBits-1)
				scale := sqrtFixp(tmp)
				if s2 < 0 {
					shift := uint(-s2)
					for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
						mdctSpectrumLeft[j] = (FMultDiv2DD(mdctSpectrumLeft[j], scale) - FMultDiv2DD(mdctSpectrumRight[j], scale)) >> shift
						mdctSpectrumRight[j] = 0
					}
				} else {
					shift := uint(s2)
					for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
						mdctSpectrumLeft[j] = (FMultDiv2DD(mdctSpectrumLeft[j], scale) - FMultDiv2DD(mdctSpectrumRight[j], scale)) << shift
						mdctSpectrumRight[j] = 0
					}
				}
			} else {
				s0 := minInt(sL, sR)
				es := FixpDBL(0)
				for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
					sum := ((mdctSpectrumLeft[j] << uint(s0)) >> 1) + ((mdctSpectrumRight[j] << uint(s0)) >> 1)
					es += FMultDiv2DD(sum, sum) >> uint(mdctSpecShift-1)
				}
				msMask[idx] = 0
				tmp, s1 := fDivNormExp(sfbEnergyLeft[idx], es)
				s2 := s1 + (2 * s0) - 2 - mdctSpecShift
				if s2&1 != 0 {
					tmp >>= 1
					s2++
				}
				s2 = (s2 >> 1) + 1
				s2 = minInt(maxInt(s2, -(DfractBits-1)), DfractBits-1)
				scale := sqrtFixp(tmp)
				if s2 < 0 {
					shift := uint(-s2)
					for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
						mdctSpectrumLeft[j] = (FMultDiv2DD(mdctSpectrumLeft[j], scale) + FMultDiv2DD(mdctSpectrumRight[j], scale)) >> shift
						mdctSpectrumRight[j] = 0
					}
				} else {
					shift := uint(s2)
					for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
						mdctSpectrumLeft[j] = (FMultDiv2DD(mdctSpectrumLeft[j], scale) + FMultDiv2DD(mdctSpectrumRight[j], scale)) << shift
						mdctSpectrumRight[j] = 0
					}
				}
			}

			isBook[idx] = codeBookISInPhaseNo
			if realIsScale[idx] < 0 {
				isScale[idx] = int(((realIsScale[idx] >> 1) - isRealScaleRound) >> (DfractBits - 1 - realScaleSF - ldDataShift - 1))
				isScale[idx]++
			} else {
				isScale[idx] = int(((realIsScale[idx] >> 1) + isRealScaleRound) >> (DfractBits - 1 - realScaleSF - ldDataShift - 1))
			}

			sfbEnergyRight[idx] = 0
			sfbEnergyLdDataRight[idx] = ldDataMinusOne
			sfbThresholdRight[idx] = 0
			sfbThresholdLdDataRight[idx] = psyPostMinThresholdLdData
			sfbSpreadEnRight[idx] = 0

			*msDigest = MsMaskSome
		}
	}
}

func fdkaacEncPrepareIntensityDecision(
	sfbEnergyLeft []FixpDBL,
	sfbEnergyRight []FixpDBL,
	sfbEnergyLdDataLeft []FixpDBL,
	sfbEnergyLdDataRight []FixpDBL,
	mdctSpectrumLeft []FixpDBL,
	mdctSpectrumRight []FixpDBL,
	isParams *intensityParameters,
	hrrErr []FixpDBL,
	isMask []int,
	realScale []FixpDBL,
	normSfbLoudness []FixpDBL,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	sfbOffset []int,
) {
	var overallLoudness [maxNoOfGroups]FixpDBL

	grpCounter := 0
	for sfboffs := 0; sfboffs < sfbCnt; sfboffs += sfbPerGroup {
		overallLoudness[grpCounter] = 0
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			idx := sfb + sfboffs
			isValue := sfbEnergyLdDataLeft[idx] - sfbEnergyLdDataRight[idx]
			realScale[idx] = minFixpDBL(isRealScaleMax, maxFixpDBL(-isRealScaleMax, isValue))

			sL := maxInt(0, CountLeadingBits(sfbEnergyLeft[idx])-1)
			sR := maxInt(0, CountLeadingBits(sfbEnergyRight[idx])-1)
			shift := (minInt(sL, sR) >> 2) << 2
			loudness := ((sfbEnergyLeft[idx] << uint(shift)) >> 1) + ((sfbEnergyRight[idx] << uint(shift)) >> 1)
			normSfbLoudness[idx] = sqrtFixp(sqrtFixp(loudness)) >> uint(shift>>2)
			overallLoudness[grpCounter] += normSfbLoudness[idx] >> overallLoudnessSF

			if sfbEnergyLeft[idx] >= FMultDD(isParams.LeftRightRatioThreshold, sfbEnergyRight[idx]) &&
				FMultDD(isParams.LeftRightRatioThreshold, sfbEnergyLeft[idx]) <= sfbEnergyRight[idx] {
				hrrErr[idx] = 0x10000000
			}
		}
		grpCounter++
	}

	grpCounter = 0
	for sfboffs := 0; sfboffs < sfbCnt; sfboffs += sfbPerGroup {
		invOverallLoudness := FixpDBL(0)
		invOverallLoudnessSF := 0
		if overallLoudness[grpCounter] != 0 {
			invOverallLoudness, invOverallLoudnessSF = fDivNormExp(MaxValDBL, overallLoudness[grpCounter])
			invOverallLoudnessSF = invOverallLoudnessSF - overallLoudnessSF + 1
		}
		invOverallLoudnessSF = minInt(maxInt(invOverallLoudnessSF, -(DfractBits-1)), DfractBits-1)

		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			idx := sfb + sfboffs
			tmp := FMultDiv2DD((normSfbLoudness[idx]>>overallLoudnessSF)<<overallLoudnessSF, invOverallLoudness)
			normSfbLoudness[idx] = ScaleValueDBL(tmp, invOverallLoudnessSF)
			channelCorr := FixpDBL(0)
			invN := getInvInt((sfbOffset[idx+1] - sfbOffset[idx]) >> 1)
			if invN > 0 {
				ml := FixpDBL(0)
				mr := FixpDBL(0)
				prodLR := FixpDBL(0)
				squareL := FixpDBL(0)
				squareR := FixpDBL(0)

				sL := calcIntensitySfbMaxScale(mdctSpectrumLeft, sfbOffset[idx], sfbOffset[idx+1])
				sR := calcIntensitySfbMaxScale(mdctSpectrumRight, sfbOffset[idx], sfbOffset[idx+1])
				shift := minInt(sL, sR)

				for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
					ml += FMultDiv2DD(mdctSpectrumLeft[j]<<uint(shift), invN)
					mr += FMultDiv2DD(mdctSpectrumRight[j]<<uint(shift), invN)
				}
				ml = FMultDiv2DD(ml, invN)
				mr = FMultDiv2DD(mr, invN)

				for j := sfbOffset[idx]; j < sfbOffset[idx+1]; j++ {
					tmpL := FMultDiv2DD(mdctSpectrumLeft[j]<<uint(shift), invN) - ml
					tmpR := FMultDiv2DD(mdctSpectrumRight[j]<<uint(shift), invN) - mr
					prodLR += FMultDiv2DD(tmpL, tmpR)
					squareL += FixPow2Div2D(tmpL)
					squareR += FixPow2Div2D(tmpR)
				}
				prodLR <<= 1
				squareL <<= 1
				squareR <<= 1

				if squareL > 0 && squareR > 0 {
					sL = maxInt(0, CountLeadingBits(squareL)-1)
					sR = maxInt(0, CountLeadingBits(squareR)-1)
					scaleShift := ((sL + sR) >> 1) << 1
					sL = minInt(sL, scaleShift)
					sR = scaleShift - sL
					tmp = FMultDD(squareL<<uint(sL), squareR<<uint(sR))
					tmp = sqrtFixp(tmp)

					var channelCorrSF int
					if prodLR < 0 {
						channelCorr, channelCorrSF = fDivNormExp(-prodLR, tmp)
						channelCorr = -channelCorr
					} else {
						channelCorr, channelCorrSF = fDivNormExp(prodLR, tmp)
					}
					channelCorrSF = minInt(maxInt(channelCorrSF+((sL+sR)>>1), -(DfractBits-1)), DfractBits-1)
					if channelCorrSF < 0 {
						channelCorr >>= uint(-channelCorrSF)
					} else if fixpAbsDBL(channelCorr) > MaxValDBL>>uint(channelCorrSF) {
						if channelCorr < 0 {
							channelCorr = -MaxValDBL
						} else {
							channelCorr = MaxValDBL
						}
					} else {
						channelCorr <<= uint(channelCorrSF)
					}
				}
			}

			if hrrErr[idx] == 0x10000000 {
				continue
			}
			hrrErr[idx] = FMultDiv2DD(0x20000000-(channelCorr>>2), normSfbLoudness[idx])
			if fixpAbsDBL(channelCorr) >= isParams.CorrThresh {
				isMask[idx] = 1
			}
		}
		grpCounter++
	}
}

func fdkaacEncFinalizeIntensityDecision(
	hrrErr []FixpDBL,
	isMask []int,
	realIsScale []FixpDBL,
	normSfbLoudness []FixpDBL,
	isParams *intensityParameters,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
) {
	isScaleLast := FixpDBL(0)
	isStartValueFound := 0

	for sfboffs := 0; sfboffs < sfbCnt; sfboffs += sfbPerGroup {
		startIsSfb := 0
		inIsBlock := 0
		currentIsSfbCount := 0
		overallHrrError := FixpDBL(0)
		isRegionLoudness := FixpDBL(0)

		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			idx := sfboffs + sfb
			if isMask[idx] == 1 {
				if currentIsSfbCount == 0 {
					startIsSfb = idx
				}
				if isStartValueFound == 0 {
					isScaleLast = realIsScale[idx]
					isStartValueFound = 1
				}
				inIsBlock = 1
				currentIsSfbCount++
				overallHrrError += hrrErr[idx] >> (maxSfbPerGroupSF - 3)
				isRegionLoudness += normSfbLoudness[idx] >> maxSfbPerGroupSF
			} else if inIsBlock != 0 {
				overallHrrError += hrrErr[idx] >> (maxSfbPerGroupSF - 3)
				isRegionLoudness += normSfbLoudness[idx] >> maxSfbPerGroupSF
				if hrrErr[idx] < (isParams.LocalErrorThresh>>3) &&
					overallHrrError < (isParams.TotalErrorThresh>>maxSfbPerGroupSF) {
					currentIsSfbCount++
					isMask[idx] = 1
				} else {
					inIsBlock = 0
				}
			}

			if inIsBlock != 0 {
				if fixpAbsDBL(isScaleLast-realIsScale[idx]) <
					(isParams.DirectionDeviationThresh >> (realScaleSF + ldDataShift - isDirectionDeviationThreshSF)) {
					isScaleLast = realIsScale[idx]
				} else {
					isMask[idx] = 0
					inIsBlock = 0
					currentIsSfbCount--
				}
			}

			if currentIsSfbCount > 0 && (inIsBlock == 0 || sfb == maxSfbPerGroup-1) {
				if currentIsSfbCount < isParams.MinISSFBs ||
					isRegionLoudness < (isParams.IsRegionMinLoudness>>maxSfbPerGroupSF) {
					for j := startIsSfb; j <= idx; j++ {
						isMask[j] = 0
					}
					isScaleLast = 0
					isStartValueFound = 0
					for j := 0; j < startIsSfb; j++ {
						if isMask[j] != 0 {
							isScaleLast = realIsScale[j]
							isStartValueFound = 1
						}
					}
				}
				currentIsSfbCount = 0
				overallHrrError = 0
				isRegionLoudness = 0
			}
		}
	}
}

func calcIntensitySfbMaxScale(mdctSpectrum []FixpDBL, l1 int, l2 int) int {
	maxSpec := FixpDBL(0)
	for i := l1; i < l2; i++ {
		maxSpec = maxFixpDBL(maxSpec, fixpAbsDBL(mdctSpectrum[i]))
	}
	if maxSpec == 0 {
		return DfractBits - 2
	}
	return CountLeadingBits(maxSpec) - 1
}

func getInvInt(value int) FixpDBL {
	return invCount[minInt(maxInt(value, 0), len(invCount)-1)]
}

func checkIntensityStereoInputs(
	sfbEnergyLeft []FixpDBL,
	sfbEnergyRight []FixpDBL,
	mdctSpectrumLeft []FixpDBL,
	mdctSpectrumRight []FixpDBL,
	sfbThresholdLeft []FixpDBL,
	sfbThresholdRight []FixpDBL,
	sfbThresholdLdDataRight []FixpDBL,
	sfbSpreadEnLeft []FixpDBL,
	sfbSpreadEnRight []FixpDBL,
	sfbEnergyLdDataLeft []FixpDBL,
	sfbEnergyLdDataRight []FixpDBL,
	msDigest *int,
	msMask []int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	sfbOffset []int,
	isBook []int,
	isScale []int,
	pnsData []*PNSData,
) {
	if msDigest == nil {
		panic("fdkaac: nil intensity stereo digest")
	}
	if sfbCnt <= 0 || sfbCnt > maxGroupedSFB || sfbPerGroup <= 0 || sfbCnt%sfbPerGroup != 0 {
		panic("fdkaac: invalid intensity stereo band count")
	}
	if maxSfbPerGroup <= 0 || maxSfbPerGroup > sfbPerGroup {
		panic("fdkaac: invalid intensity stereo group width")
	}
	checkGroupedSfbOffsets(
		sfbOffset,
		sfbCnt,
		sfbPerGroup,
		maxSfbPerGroup,
		false,
		"fdkaac: invalid intensity stereo offset",
		"fdkaac: short intensity stereo offsets",
		len(mdctSpectrumLeft),
		len(mdctSpectrumRight),
	)
	if len(sfbEnergyLeft) < sfbCnt || len(sfbEnergyRight) < sfbCnt ||
		len(sfbThresholdLeft) < sfbCnt || len(sfbThresholdRight) < sfbCnt ||
		len(sfbThresholdLdDataRight) < sfbCnt || len(sfbSpreadEnLeft) < sfbCnt ||
		len(sfbSpreadEnRight) < sfbCnt || len(sfbEnergyLdDataLeft) < sfbCnt ||
		len(sfbEnergyLdDataRight) < sfbCnt || len(msMask) < sfbCnt ||
		len(isBook) < sfbCnt || len(isScale) < sfbCnt {
		panic("fdkaac: short intensity stereo vector")
	}
	if len(pnsData) == 1 {
		panic("fdkaac: short intensity stereo pns pair")
	}
}
