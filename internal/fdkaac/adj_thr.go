package fdkaac

const (
	chaosMeasureMin      FixpDBL = 0x18000000
	chaosMeasurePatchUse FixpDBL = 0x64000000
	chaosMeasurePatch34  FixpDBL = 0x68000000
	chaosMeasurePatch12  FixpDBL = 0x66000000
	chaosMeasureShort    FixpDBL = 0x60000000
	adjThrTrue                   = 1
	adjThrFalse                  = 0
)

const (
	AvoidHoleNone     uint8 = 0
	AvoidHoleInactive uint8 = 1
	AvoidHoleActive   uint8 = 2
)

const (
	qBitFac  = 24
	qAvgBits = 17

	constPartHeadroom = 4

	vbrScaleGroupEnergy = 8
	vbrScaleFormFac     = 4
	vbrScaleNRGs        = 8
	vbrScaleNLines      = 16
	vbrScaleNRGsSqrt4   = 2
	vbrScaleNLinesP34   = 12
	vbrWinTypeScale     = 3

	bitSaveSlopeLong   FixpDBL = 0x3bbbbbba
	bitSpendSlopeLong  FixpDBL = 0x55555554
	bitSaveSlopeShort  FixpDBL = 0x2e8ba2e9
	bitSpendSlopeShort         = MaxValDBL

	bitresFillBias07 FixpDBL = 0x59999980

	peAdjustMinFacHi FixpDBL = 0x26666680
	peAdjustMaxFacHi         = MaxValDBL
	peAdjustMinFacLo FixpDBL = 0x11eb8520
	peAdjustMaxFacLo FixpDBL = 0x08f5c290
	peAdjustMinDiff  FixpDBL = 0x15555560

	peCorrection12Over2   FixpDBL = 0x4ccccd00
	peCorrection065       FixpDBL = 0x53333300
	peCorrection11Over2   FixpDBL = 0x46666680
	peCorrectionHalf      FixpDBL = 0x40000000
	peCorrection085Over2F FixpDBL = 0x36666680
	peCorrection09Over2   FixpDBL = 0x39999980
	peCorrection115Over2  FixpDBL = 0x49999980
	peCorrection085       FixpDBL = 0x6ccccd00
	peCorrection015       FixpDBL = 0x13333340
	peCorrection07        FixpDBL = 0x59999980
	peCorrection03        FixpDBL = 0x26666680
	peCorrection085Over2D FixpDBL = 0x36666666

	lowBitresCorrectionAmp FixpDBL = 0x00a3d70a
	lowBitresMaxDiff       FixpDBL = 0x20000000
	lowBitresCorrectionMin FixpDBL = 0x30000000

	minSnrLimitLD64               FixpDBL = -0x00a4d3c2
	minSnrAdaptLog10              FixpDBL = 0x268826c0
	minSnrAdaptAvgLdShift         FixpDBL = 0x0c000000
	minSnrAdaptDefaultMaxRed              = FixpDBL(0x00800000)
	minSnrAdaptDefaultStartRatio          = FixpDBL(0x06a4d3c0)
	minSnrAdaptDefaultRedRatioFac         = FixpDBL(-0x30000000)
	minSnrAdaptDefaultRedOffs             = FixpDBL(0x02c00000)

	snrLdMin1 FixpDBL = -0x0352f221
	snrLdMin2 FixpDBL = 0x0351e1a2
	snrLdFac  FixpDBL = -0x00a4d3c2
	snrLdMin3 FixpDBL = -0x02000000
	snrLdMin4 FixpDBL = 0x02000000
	snrLdMin5 FixpDBL = -0x04000000

	avoidHoleShortSpreadFac FixpDBL = 0x50a3d700
	avoidHoleMsSpreadFac    FixpDBL = 0x73333300
	avoidHoleNegHalf        FixpDBL = -0x40000000

	minThresholdRatio29DB FixpDBL = 0x134469eb
	vbrChaosHalf          FixpDBL = 0x40000000
	vbrChaosAvgFac0       FixpDBL = 0x20000000
	vbrChaosAvgFac1       FixpDBL = 0x60000000
	vbrChaos02            FixpDBL = 0x1999999a
	vbrChaos01            FixpDBL = 0x0ccccccd
	vbrChaos07Over12      FixpDBL = 0x4aaaaaab
	vbrMinLdThresh        FixpDBL = -0x42000000
	vbrLimitThrReduced    FixpDBL = 0x00008000
	vbr282Over4           FixpDBL = 0x5a3d70a4

	correctThreshLdShift FixpDBL = 0x0c000000
	correctThreshMaxLd   FixpDBL = 0x28000000

	reduceMinSnrMaxNLines = int(MaxValDBL>>uint(peConstPartShift-1)) / 3

	allowMoreHolesNumEnergyLevels = 8
	allowMoreHolesThrOffset       = FixpDBL(0x02000000)
	allowMoreHolesMsMargin        = FixpDBL(-44356546)
)

var adjThrInvInt = [8]FixpDBL{
	MaxValDBL, MaxValDBL, 0x40000000, 0x2aaaaaaa,
	0x20000000, 0x19999999, 0x15555555, 0x12492492,
}

var adjThrInvSqrt4 = [8]FixpDBL{
	MaxValDBL, MaxValDBL, 0x6ba27e65, 0x61424bb5,
	0x5a827999, 0x55994845, 0x51c8e33c, 0x4eb160d1,
}

var allowMoreHolesEnergyLevel = [allowMoreHolesNumEnergyLevels]FixpDBL{
	0x08888890, 0x1999999a, 0x2aaaaab9, 0x3bbbbbc3,
	0x4cccccf8, 0x5dddde02, 0x6eeeeef6, 0,
}

type BitresMode int

const (
	BitresModeFull BitresMode = iota
	BitresModeReduced
	BitresModeDisabled
)

type BRESParam struct {
	ClipSaveLow   FixpDBL
	ClipSaveHigh  FixpDBL
	MinBitSave    FixpDBL
	MaxBitSave    FixpDBL
	ClipSpendLow  FixpDBL
	ClipSpendHigh FixpDBL
	MinBitSpend   FixpDBL
	MaxBitSpend   FixpDBL
}

type AHParam struct {
	ModifyMinSnr int
	StartSfbL    int
	StartSfbS    int
}

type MinSNRAdaptParam struct {
	MaxRed      FixpDBL
	StartRatio  FixpDBL
	MaxRatio    FixpDBL
	RedRatioFac FixpDBL
	RedOffs     FixpDBL
}

type ATSElement struct {
	PeMin               int
	PeMax               int
	PeOffset            int
	Bits2PeFactorM      FixpDBL
	Bits2PeFactorE      int
	AHParam             AHParam
	MinSNRAdaptParam    MinSNRAdaptParam
	PeLast              int
	DynBitsLast         int
	PeCorrectionFactorM FixpDBL
	PeCorrectionFactorE int
	VBRQualFactor       FixpDBL
	ChaosMeasureOld     FixpDBL
	ChaosMeasureEnFac   [2]FixpDBL
	LastEnFacPatch      [2]int
}

type AdjThrState struct {
	BRESParamLong       BRESParam
	BRESParamShort      BRESParam
	BitDistributionMode int
	MaxIter2ndGuess     int
}

type CorrectThresholdScratch struct {
	SfbPEFactorsLdData    [2][maxGroupedSFB]FixpDBL
	SfbNActiveLinesLdData [2][maxGroupedSFB]FixpDBL
}

func FDKaacEncInitBitresState(state *AdjThrState) {
	if state == nil {
		panic("fdkaac: nil bit reservoir state")
	}
	state.BRESParamLong = BRESParam{
		ClipSaveLow:   0x1999999a,
		ClipSaveHigh:  0x7999999a,
		MinBitSave:    -0x06666666,
		MaxBitSave:    0x26666666,
		ClipSpendLow:  0x1999999a,
		ClipSpendHigh: 0x7999999a,
		MinBitSpend:   -0x0ccccccd,
		MaxBitSpend:   0x33333333,
	}
	state.BRESParamShort = BRESParam{
		ClipSaveLow:   0x199999a0,
		ClipSaveHigh:  0x5fffffff,
		MinBitSave:    0,
		MaxBitSave:    0x199999a0,
		ClipSpendLow:  0x199999a0,
		ClipSpendHigh: 0x5fffffff,
		MinBitSpend:   -0x06666668,
		MaxBitSpend:   0x40000000,
	}
	state.BitDistributionMode = 0
	state.MaxIter2ndGuess = 1
}

func FDKaacEncInitMinSnrAdaptParam(param *MinSNRAdaptParam) {
	if param == nil {
		panic("fdkaac: nil min-SNR adaptation parameter")
	}
	param.MaxRed = minSnrAdaptDefaultMaxRed
	param.StartRatio = minSnrAdaptDefaultStartRatio
	param.MaxRatio = 0
	param.RedRatioFac = minSnrAdaptDefaultRedRatioFac
	param.RedOffs = minSnrAdaptDefaultRedOffs
}

func FDKaacEncCalcThresholdExp(
	thrExp *[2][maxGroupedSFB]FixpDBL,
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	nChannels int,
) {
	checkThresholdAdjustmentInputs(qcOutChannel, psyOutChannel, nChannels)
	if thrExp == nil {
		panic("fdkaac: nil threshold exponent scratch")
	}

	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		qcCh := qcOutChannel[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				thrExp[ch][idx] = CalcInvLdData(qcCh.SfbThresholdLdData[idx] >> 2)
			}
		}
	}
}

func FDKaacEncAdaptMinSnr(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	msaParam *MinSNRAdaptParam,
	nChannels int,
) {
	checkThresholdAdjustmentInputs(qcOutChannel, psyOutChannel, nChannels)
	if msaParam == nil {
		panic("fdkaac: nil min-SNR adaptation parameter")
	}

	msaParamRedRatioFac := FMultDD(msaParam.RedRatioFac, minSnrAdaptLog10)

	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		qcCh := qcOutChannel[ch]

		nSfb := 0
		accu := FixpDBL(0)
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			nSfb += psyCh.MaxSfbPerGroup
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				accu += psyCh.SfbEnergy[sfbGrp+sfb] >> 6
			}
		}

		avgEnLD64 := MinValDBL
		if accu != 0 && nSfb != 0 {
			avgEnLD64 = CalcLdData(accu) + minSnrAdaptAvgLdShift - CalcLdInt(nSfb)
		}

		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				sfbEnergyLdData := qcCh.SfbEnergyLdData[idx]
				sfbMinSnrLdData := qcCh.SfbMinSnrLdData[idx]
				dbRatio := avgEnLD64 - sfbEnergyLdData
				update := msaParam.StartRatio < dbRatio
				minSnrRed := msaParam.RedOffs + FMultDD(msaParamRedRatioFac, dbRatio)
				minSnrRed = maxFixpDBL(minSnrRed, msaParam.MaxRed)
				minSnrRed = FMultDD(sfbMinSnrLdData, minSnrRed) << 6
				minSnrRed = minFixpDBL(minSnrLimitLD64, minSnrRed)
				if update {
					qcCh.SfbMinSnrLdData[idx] = minSnrRed
				}
			}
		}
	}
}

func FDKaacEncInitAvoidHoleFlag(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	ahFlag *[2][maxGroupedSFB]uint8,
	toolsInfo *ToolsInfo,
	nChannels int,
	ahParam *AHParam,
) {
	checkAvoidHoleInputs(qcOutChannel, psyOutChannel, ahFlag, toolsInfo, nChannels, ahParam)

	for ch := 0; ch < nChannels; ch++ {
		qcCh := qcOutChannel[ch]
		psyCh := psyOutChannel[ch]
		if psyCh.LastWindowSequence != ShortWindow {
			for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
				for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
					qcCh.SfbSpreadEnergy[sfbGrp+sfb] >>= 1
				}
			}
		} else {
			for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
				for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
					idx := sfbGrp + sfb
					qcCh.SfbSpreadEnergy[idx] = FMultDD(avoidHoleShortSpreadFac, qcCh.SfbSpreadEnergy[idx])
				}
			}
		}
	}

	if ahParam.ModifyMinSnr != 0 {
		for ch := 0; ch < nChannels; ch++ {
			qcCh := qcOutChannel[ch]
			psyCh := psyOutChannel[ch]
			for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
				for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
					idx := sfbGrp + sfb
					sfbEnM1 := qcCh.SfbEnergy[idx]
					if sfb > 0 {
						sfbEnM1 = qcCh.SfbEnergy[idx-1]
					}
					sfbEnP1 := qcCh.SfbEnergy[idx]
					if sfb < psyCh.MaxSfbPerGroup-1 {
						sfbEnP1 = qcCh.SfbEnergy[idx+1]
					}

					avgEn := (sfbEnM1 >> 1) + (sfbEnP1 >> 1)
					avgEnLdData := CalcLdData(avgEn)
					sfbEn := qcCh.SfbEnergy[idx]
					sfbEnLdData := qcCh.SfbEnergyLdData[idx]

					if sfbEn > avgEn {
						tmpMinSnrLdData := FixpDBL(0)
						if psyCh.LastWindowSequence == LongWindow {
							tmpMinSnrLdData = snrLdFac + maxFixpDBL(avgEnLdData-sfbEnLdData, snrLdMin1-snrLdFac)
						} else {
							tmpMinSnrLdData = snrLdFac + maxFixpDBL(avgEnLdData-sfbEnLdData, snrLdMin3-snrLdFac)
						}
						qcCh.SfbMinSnrLdData[idx] = minFixpDBL(qcCh.SfbMinSnrLdData[idx], tmpMinSnrLdData)
					}

					if sfbEnLdData+snrLdMin4 < avgEnLdData && sfbEn > 0 {
						tmpMinSnrLdData := avgEnLdData - sfbEnLdData - snrLdMin4 + qcCh.SfbMinSnrLdData[idx]
						tmpMinSnrLdData = minFixpDBL(snrLdFac, tmpMinSnrLdData)
						qcCh.SfbMinSnrLdData[idx] = minFixpDBL(tmpMinSnrLdData, qcCh.SfbMinSnrLdData[idx]+snrLdMin2)
					}
				}
			}
		}
	}

	if nChannels == 2 {
		qcOutChanM := qcOutChannel[0]
		qcOutChanS := qcOutChannel[1]
		psyOutChanM := psyOutChannel[0]
		for sfbGrp := 0; sfbGrp < psyOutChanM.SfbCnt; sfbGrp += psyOutChanM.SfbPerGroup {
			for sfb := 0; sfb < psyOutChanM.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				if toolsInfo.MsMask[idx] == 0 {
					continue
				}

				maxSfbEnLd := maxFixpDBL(qcOutChanM.SfbEnergyLdData[idx], qcOutChanS.SfbEnergyLdData[idx])
				maxThrLd := FixpDBL(0)
				if ((snrLdMin5 >> 1) + (maxSfbEnLd >> 1) + (qcOutChanM.SfbMinSnrLdData[idx] >> 1)) <= avoidHoleNegHalf {
					maxThrLd = MinValDBL
				} else {
					maxThrLd = snrLdMin5 + maxSfbEnLd + qcOutChanM.SfbMinSnrLdData[idx]
				}

				sfbMinSnrTmpLd := FixpDBL(0)
				if qcOutChanM.SfbEnergy[idx] > 0 {
					sfbMinSnrTmpLd = maxThrLd - qcOutChanM.SfbEnergyLdData[idx]
				}
				qcOutChanM.SfbMinSnrLdData[idx] = maxFixpDBL(qcOutChanM.SfbMinSnrLdData[idx], sfbMinSnrTmpLd)
				if qcOutChanM.SfbMinSnrLdData[idx] <= 0 {
					qcOutChanM.SfbMinSnrLdData[idx] = minFixpDBL(qcOutChanM.SfbMinSnrLdData[idx], snrLdFac)
				}

				sfbMinSnrTmpLd = 0
				if qcOutChanS.SfbEnergy[idx] > 0 {
					sfbMinSnrTmpLd = maxThrLd - qcOutChanS.SfbEnergyLdData[idx]
				}
				qcOutChanS.SfbMinSnrLdData[idx] = maxFixpDBL(qcOutChanS.SfbMinSnrLdData[idx], sfbMinSnrTmpLd)
				if qcOutChanS.SfbMinSnrLdData[idx] <= 0 {
					qcOutChanS.SfbMinSnrLdData[idx] = minFixpDBL(qcOutChanS.SfbMinSnrLdData[idx], snrLdFac)
				}

				if qcOutChanM.SfbEnergy[idx] > qcOutChanM.SfbSpreadEnergy[idx] {
					qcOutChanS.SfbSpreadEnergy[idx] = FMultDD(qcOutChanS.SfbEnergy[idx], avoidHoleMsSpreadFac)
				}
				if qcOutChanS.SfbEnergy[idx] > qcOutChanS.SfbSpreadEnergy[idx] {
					qcOutChanM.SfbSpreadEnergy[idx] = FMultDD(qcOutChanM.SfbEnergy[idx], avoidHoleMsSpreadFac)
				}
			}
		}
	}

	for ch := 0; ch < nChannels; ch++ {
		qcCh := qcOutChannel[ch]
		psyCh := psyOutChannel[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				if qcCh.SfbSpreadEnergy[idx] > qcCh.SfbEnergy[idx] || qcCh.SfbMinSnrLdData[idx] > 0 {
					ahFlag[ch][idx] = AvoidHoleNone
				} else {
					ahFlag[ch][idx] = AvoidHoleInactive
				}
			}
		}
	}
}

func FDKaacEncCalcPENoAH(
	peData *PEData,
	ahFlag *[2][maxGroupedSFB]uint8,
	psyOutChannel []*PsyOutChannel,
	nChannels int,
) (int, int, int) {
	checkPENoAHInputs(peData, ahFlag, psyOutChannel, nChannels)

	peTmp := peData.Offset
	constPartTmp := 0
	nActiveLinesTmp := 0
	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		peChanData := &peData.PEChannelData[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				if ahFlag[ch][idx] < AvoidHoleActive {
					peTmp += int(peChanData.SfbPe[idx])
					constPartTmp += int(peChanData.SfbConstPart[idx] >> constPartHeadroom)
					nActiveLinesTmp += int(peChanData.SfbNActiveLines[idx])
				}
			}
		}
	}

	pe := peTmp >> peConstPartShift
	constPart := constPartTmp >> (peConstPartShift - constPartHeadroom)
	return pe, constPart, nActiveLinesTmp
}

func FDKaacEncReduceThresholdsCBR(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	ahFlag *[2][maxGroupedSFB]uint8,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	nChannels int,
	redValM FixpDBL,
	redValE int,
) {
	checkThresholdReductionInputs(qcOutChannel, psyOutChannel, ahFlag, thrExp, nChannels, redValM, redValE)
	if redValM == 0 {
		return
	}

	for ch := 0; ch < nChannels; ch++ {
		qcCh := qcOutChannel[ch]
		psyCh := psyOutChannel[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				sfbEnLdData := qcCh.SfbWeightedEnergyLdData[idx]
				sfbThrLdData := qcCh.SfbThresholdLdData[idx]
				sfbThrExp := thrExp[ch][idx]
				if sfbEnLdData > sfbThrLdData && ahFlag[ch][idx] != AvoidHoleActive {
					minScale := minInt(CountLeadingBits(sfbThrExp), CountLeadingBits(redValM)-redValE) - 1

					sfbThrReducedLdData := CalcLdData(fixpAbsDBL(
						ScaleValueDBL(sfbThrExp, minScale)+
							ScaleValueDBL(redValM, redValE+minScale),
					)) - FixpDBL(minScale<<(DfractBits-1-ldDataShift))
					sfbThrReducedLdData <<= 2

					if sfbThrReducedLdData > qcCh.SfbMinSnrLdData[idx]+sfbEnLdData && ahFlag[ch][idx] != AvoidHoleNone {
						if qcCh.SfbMinSnrLdData[idx] > MinValDBL-sfbEnLdData {
							sfbThrReducedLdData = maxFixpDBL(qcCh.SfbMinSnrLdData[idx]+sfbEnLdData, sfbThrLdData)
						} else {
							sfbThrReducedLdData = sfbThrLdData
						}
						ahFlag[ch][idx] = AvoidHoleActive
					}

					if sfbEnLdData+MaxValDBL > minThresholdRatio29DB {
						sfbThrReducedLdData = maxFixpDBL(sfbThrReducedLdData, sfbEnLdData-minThresholdRatio29DB)
					}

					qcCh.SfbThresholdLdData[idx] = sfbThrReducedLdData
				}
			}
		}
	}
}

func FDKaacEncCalcChaosMeasure(psyOutChannel *PsyOutChannel, sfbFormFactorLdData []FixpDBL) FixpDBL {
	checkChaosMeasureInputs(psyOutChannel, sfbFormFactorLdData)

	frameNLines := 0
	frameFormFactor := FixpDBL(0)
	frameEnergy := FixpDBL(0)
	for sfbGrp := 0; sfbGrp < psyOutChannel.SfbCnt; sfbGrp += psyOutChannel.SfbPerGroup {
		for sfb := 0; sfb < psyOutChannel.MaxSfbPerGroup; sfb++ {
			idx := sfbGrp + sfb
			if psyOutChannel.SfbEnergyLdData[idx] > psyOutChannel.SfbThresholdLdData[idx] {
				frameFormFactor += CalcInvLdData(sfbFormFactorLdData[idx]) >> vbrScaleFormFac
				frameNLines += psyOutChannel.SfbOffsets[idx+1] - psyOutChannel.SfbOffsets[idx]
				frameEnergy += psyOutChannel.SfbEnergy[idx] >> vbrScaleNRGs
			}
		}
	}

	if frameNLines <= 0 {
		return MaxValDBL
	}

	scaleOffset := FixpDBL(-((-vbrScaleFormFac + vbrScaleNRGsSqrt4 - formFacShift + vbrScaleNLinesP34) << (DfractBits - 1 - ldDataShift)))
	nLinesLd := CalcLdData(FixpDBL(frameNLines << (DfractBits - 1 - vbrScaleNLines)))
	chaosMeasureLd := (((CalcLdData(frameFormFactor) >> 1) -
		(CalcLdData(frameEnergy) >> (2 + 1))) -
		(FMultDiv2DD(vbrChaosAvgFac1, nLinesLd) - (scaleOffset >> 1)))
	return CalcInvLdData(chaosMeasureLd << 1)
}

func FDKaacEncReduceThresholdsVBR(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	ahFlag *[2][maxGroupedSFB]uint8,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	nChannels int,
	vbrQualFactor FixpDBL,
	chaosMeasureOld *FixpDBL,
) {
	checkThresholdReductionVBRInputs(qcOutChannel, psyOutChannel, ahFlag, thrExp, nChannels, vbrQualFactor, chaosMeasureOld)

	var chGroupEnergy [transFac][2]FixpDBL
	var redVal [transFac]FixpDBL
	frameEnergy := FixpDBL(0)
	chaosMeasure := FixpDBL(0)

	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		chEnergy := FixpDBL(0)
		groupCnt := 0
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			chGroupEnergy[groupCnt][ch] = 0
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				chGroupEnergy[groupCnt][ch] += psyCh.SfbEnergy[sfbGrp+sfb] >> vbrScaleGroupEnergy
			}
			chEnergy += chGroupEnergy[groupCnt][ch]
			groupCnt++
		}
		frameEnergy += chEnergy

		chChaosMeasure := vbrChaosHalf
		if psyOutChannel[0].LastWindowSequence != ShortWindow {
			chChaosMeasure = FDKaacEncCalcChaosMeasure(psyCh, qcOutChannel[ch].SfbFormFactorLdData[:])
		}
		chaosMeasure += FMultDD(chChaosMeasure, chEnergy)
	}

	if frameEnergy > chaosMeasure {
		scale := FixNormZD(frameEnergy) - 1
		chaosMeasure = schurDiv(chaosMeasure<<uint(scale), frameEnergy<<uint(scale), FractBits)
	} else {
		chaosMeasure = MaxValDBL
	}

	chaosMeasureAvg := FMultDD(vbrChaosAvgFac0, chaosMeasure) + FMultDD(vbrChaosAvgFac1, *chaosMeasureOld)
	chaosMeasure = minFixpDBL(chaosMeasure, chaosMeasureAvg)
	*chaosMeasureOld = chaosMeasure

	chaosMeasure = (vbrChaos02 >> 2) + FMultDD(vbrChaos07Over12, chaosMeasure-vbrChaos02)
	chaosMeasure = minFixpDBL(MaxValDBL>>2, maxFixpDBL(vbrChaos01>>2, chaosMeasure)) << 2

	if psyOutChannel[0].LastWindowSequence == ShortWindow {
		groupCnt := 0
		for sfbGrp := 0; sfbGrp < psyOutChannel[0].SfbCnt; sfbGrp += psyOutChannel[0].SfbPerGroup {
			groupEnergy := FixpDBL(0)
			for ch := 0; ch < nChannels; ch++ {
				groupEnergy += chGroupEnergy[groupCnt][ch]
			}

			groupLen := psyOutChannel[0].GroupLen[groupCnt]
			groupEnergy = FMultDD(groupEnergy, adjThrInvInt[groupLen])
			groupEnergy = minFixpDBL(groupEnergy, frameEnergy>>vbrWinTypeScale)
			groupEnergy >>= 2

			redVal[groupCnt] = FMultDD(
				FMultDD(vbrQualFactor, chaosMeasure),
				CalcInvLdData(CalcLdData(groupEnergy)>>2),
			) << uint((2+(2*vbrWinTypeScale)+vbrScaleGroupEnergy)>>2)
			groupCnt++
		}
	} else {
		redVal[0] = FMultDD(
			FMultDD(vbrQualFactor, chaosMeasure),
			CalcInvLdData(CalcLdData(frameEnergy)>>2),
		) << uint(vbrScaleGroupEnergy>>2)
	}

	for ch := 0; ch < nChannels; ch++ {
		qcCh := qcOutChannel[ch]
		psyCh := psyOutChannel[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				sfbEnLdData := qcCh.SfbWeightedEnergyLdData[idx]
				sfbThrLdData := qcCh.SfbThresholdLdData[idx]
				sfbThrExp := thrExp[ch][idx]
				if sfbThrLdData >= vbrMinLdThresh && sfbEnLdData > sfbThrLdData && ahFlag[ch][idx] != AvoidHoleActive {
					var sfbThrReducedLdData FixpDBL
					if psyCh.LastWindowSequence == ShortWindow {
						groupNumber := sfb / psyCh.SfbPerGroup
						groupLen := psyCh.GroupLen[groupNumber]
						sfbThrExp = FMultDD(
							sfbThrExp,
							FMultDD(vbr282Over4, adjThrInvSqrt4[groupLen]),
						) << 2

						if sfbThrExp <= vbrLimitThrReduced-redVal[groupNumber] {
							sfbThrReducedLdData = MinValDBL
						} else if redVal[groupNumber] >= MaxValDBL-sfbThrExp {
							sfbThrReducedLdData = 0
						} else {
							sfbThrReducedLdData = CalcLdData(sfbThrExp+redVal[groupNumber]) << 2
						}
						sfbThrReducedLdData += CalcLdInt(groupLen) - FixpDBL(6<<(DfractBits-1-ldDataShift))
					} else {
						if redVal[0] >= MaxValDBL-sfbThrExp {
							sfbThrReducedLdData = 0
						} else {
							sfbThrReducedLdData = CalcLdData(sfbThrExp+redVal[0]) << 2
						}
					}

					if sfbThrReducedLdData-sfbEnLdData > qcCh.SfbMinSnrLdData[idx] && ahFlag[ch][idx] != AvoidHoleNone {
						if qcCh.SfbMinSnrLdData[idx] > MinValDBL-sfbEnLdData {
							sfbThrReducedLdData = maxFixpDBL(qcCh.SfbMinSnrLdData[idx]+sfbEnLdData, sfbThrLdData)
						} else {
							sfbThrReducedLdData = sfbThrLdData
						}
						ahFlag[ch][idx] = AvoidHoleActive
					}

					if sfbThrReducedLdData < avoidHoleNegHalf {
						sfbThrReducedLdData = MinValDBL
					}
					sfbThrReducedLdData = maxFixpDBL(vbrMinLdThresh, sfbThrReducedLdData)
					qcCh.SfbThresholdLdData[idx] = sfbThrReducedLdData
				}
			}
		}
	}
}

func FDKaacEncCorrectThresholds(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	peData *PEData,
	ahFlag *[2][maxGroupedSFB]uint8,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	scratch *CorrectThresholdScratch,
	nChannels int,
	redValM FixpDBL,
	redValE int,
	deltaPe int,
) {
	checkCorrectThresholdInputs(qcOutChannel, psyOutChannel, peData, ahFlag, thrExp, scratch, nChannels, redValM, redValE)

	normFactorInt := 0
	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		peChanData := &peData.PEChannelData[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				if peChanData.SfbNActiveLines[idx] < 0 {
					panic("fdkaac: invalid correct-threshold active lines")
				}
				nActiveLines := int(peChanData.SfbNActiveLines[idx])
				if nActiveLines == 0 {
					scratch.SfbNActiveLinesLdData[ch][idx] = MinValDBL
				} else {
					scratch.SfbNActiveLinesLdData[ch][idx] = CalcLdInt(nActiveLines)
				}

				if (ahFlag[ch][idx] < AvoidHoleActive || deltaPe > 0) && nActiveLines != 0 {
					if thrExp[ch][idx] > -redValM {
						minScale := minInt(CountLeadingBits(thrExp[ch][idx]), CountLeadingBits(redValM)-redValE) - 1
						sumLd := CalcLdData(
							ScaleValueDBL(thrExp[ch][idx], minScale)+
								ScaleValueDBL(redValM, redValE+minScale),
						) - FixpDBL(minScale<<(DfractBits-1-ldDataShift))

						if sumLd < 0 {
							scratch.SfbPEFactorsLdData[ch][idx] = scratch.SfbNActiveLinesLdData[ch][idx] - sumLd
						} else if scratch.SfbNActiveLinesLdData[ch][idx] > MinValDBL+sumLd {
							scratch.SfbPEFactorsLdData[ch][idx] = scratch.SfbNActiveLinesLdData[ch][idx] - sumLd
						} else {
							scratch.SfbPEFactorsLdData[ch][idx] = scratch.SfbNActiveLinesLdData[ch][idx]
						}

						normFactorInt += int(CalcInvLdData(scratch.SfbPEFactorsLdData[ch][idx]))
					} else {
						scratch.SfbPEFactorsLdData[ch][idx] = MaxValDBL
					}
				} else {
					scratch.SfbPEFactorsLdData[ch][idx] = MinValDBL
				}
			}
		}
	}

	normFactorLdData := FixpDBL(0)
	if deltaPe != 0 {
		if normFactorInt <= 0 {
			panic("fdkaac: invalid correct-threshold norm factor")
		}
		normFactorLdData = CalcLdData(FixpDBL(absInt(deltaPe))) - CalcLdData(FixpDBL(normFactorInt))
	}

	for ch := 0; ch < nChannels; ch++ {
		qcCh := qcOutChannel[ch]
		psyCh := psyOutChannel[ch]
		peChanData := &peData.PEChannelData[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				if peChanData.SfbNActiveLines[idx] <= 0 {
					continue
				}

				thrFactorLdData := FixpDBL(0)
				if scratch.SfbPEFactorsLdData[ch][idx] != MinValDBL && deltaPe != 0 {
					tmp := CalcInvLdData(
						scratch.SfbPEFactorsLdData[ch][idx] +
							normFactorLdData -
							scratch.SfbNActiveLinesLdData[ch][idx] -
							correctThreshLdShift,
					)
					if deltaPe >= 0 {
						tmp = -tmp
					}
					thrFactorLdData = minFixpDBL(tmp, correctThreshMaxLd)
				}

				sfbThrLdData := qcCh.SfbThresholdLdData[idx]
				sfbEnLdData := qcCh.SfbWeightedEnergyLdData[idx]
				sfbThrReducedLdData := sfbThrLdData + thrFactorLdData
				if thrFactorLdData < 0 && sfbThrLdData <= MinValDBL-thrFactorLdData {
					sfbThrReducedLdData = MinValDBL
				}

				if sfbThrReducedLdData-sfbEnLdData > qcCh.SfbMinSnrLdData[idx] &&
					ahFlag[ch][idx] == AvoidHoleInactive {
					if sfbEnLdData > sfbThrLdData-qcCh.SfbMinSnrLdData[idx] {
						sfbThrReducedLdData = qcCh.SfbMinSnrLdData[idx] + sfbEnLdData
					} else {
						sfbThrReducedLdData = sfbThrLdData
					}
					ahFlag[ch][idx] = AvoidHoleActive
				}

				qcCh.SfbThresholdLdData[idx] = sfbThrReducedLdData
			}
		}
	}
}

func FDKaacEncReduceMinSnr(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	peData *PEData,
	ahFlag *[2][maxGroupedSFB]uint8,
	nChannels int,
	desiredPe int,
	redPeGlobal *int,
) {
	checkReduceMinSnrInputs(qcOutChannel, psyOutChannel, peData, ahFlag, nChannels, desiredPe, redPeGlobal)

	newGlobalPe := *redPeGlobal
	if newGlobalPe <= desiredPe {
		return
	}

	globalMaxSfb := 0
	for ch := 0; ch < nChannels; ch++ {
		globalMaxSfb = maxInt(globalMaxSfb, psyOutChannel[ch].MaxSfbPerGroup)
	}

	for newGlobalPe > desiredPe {
		globalMaxSfb--
		if globalMaxSfb < 0 {
			break
		}

		for ch := 0; ch < nChannels; ch++ {
			qcCh := qcOutChannel[ch]
			psyCh := psyOutChannel[ch]
			peChanData := &peData.PEChannelData[ch]

			if globalMaxSfb < psyCh.MaxSfbPerGroup {
				deltaPe := FixpDBL(0)

				for sfb := globalMaxSfb; sfb < psyCh.SfbCnt; sfb += psyCh.SfbPerGroup {
					if peChanData.SfbPe[sfb] < 0 ||
						peChanData.SfbNLines[sfb] < 0 ||
						peChanData.SfbNLines[sfb] > reduceMinSnrMaxNLines {
						panic("fdkaac: invalid reduce-min-SNR PE band")
					}

					if ahFlag[ch][sfb] != AvoidHoleNone &&
						qcCh.SfbMinSnrLdData[sfb] < snrLdFac &&
						qcCh.SfbWeightedEnergyLdData[sfb] > qcCh.SfbThresholdLdData[sfb]-snrLdFac {
						qcCh.SfbMinSnrLdData[sfb] = snrLdFac
						qcCh.SfbThresholdLdData[sfb] = qcCh.SfbWeightedEnergyLdData[sfb] + snrLdFac

						deltaPe -= peChanData.SfbPe[sfb]
						peChanData.SfbPe[sfb] = FixpDBL((3 * peChanData.SfbNLines[sfb]) << (peConstPartShift - 1))
						deltaPe += peChanData.SfbPe[sfb]
					}
				}

				deltaPeInt := int(deltaPe >> peConstPartShift)
				peData.Pe += FixpDBL(deltaPeInt)
				peChanData.Pe += FixpDBL(deltaPeInt)
				newGlobalPe += deltaPeInt
			}

			if newGlobalPe <= desiredPe {
				*redPeGlobal = newGlobalPe
				return
			}
		}
	}

	*redPeGlobal = newGlobalPe
}

func FDKaacEncAllowMoreHoles(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	peData *PEData,
	toolsInfo *ToolsInfo,
	adjThrStateElement *ATSElement,
	ahFlag *[2][maxGroupedSFB]uint8,
	nChannels int,
	desiredPe int,
	currentPe int,
) {
	checkAllowMoreHolesInputs(qcOutChannel, psyOutChannel, peData, toolsInfo, adjThrStateElement, ahFlag, nChannels, desiredPe, currentPe)

	actPe := currentPe
	if actPe <= desiredPe {
		return
	}

	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		peChanData := &peData.PEChannelData[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := psyCh.MaxSfbPerGroup; sfb < psyCh.SfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				if peChanData.SfbPe[idx] < 0 {
					panic("fdkaac: invalid allow-more-holes PE band")
				}
				peChanData.SfbPe[idx] = 0
			}
		}
	}

	if nChannels == 2 && psyOutChannel[0].LastWindowSequence == psyOutChannel[1].LastWindowSequence {
		for sfb := psyOutChannel[0].MaxSfbPerGroup - 1; sfb >= 0; sfb-- {
			for sfbGrp := 0; sfbGrp < psyOutChannel[0].SfbCnt; sfbGrp += psyOutChannel[0].SfbPerGroup {
				idx := sfbGrp + sfb
				if toolsInfo.MsMask[idx] == 0 {
					continue
				}

				energyLdL := qcOutChannel[0].SfbWeightedEnergyLdData[idx]
				energyLdR := qcOutChannel[1].SfbWeightedEnergyLdData[idx]

				if ahFlag[1][idx] != AvoidHoleNone &&
					((allowMoreHolesMsMargin>>1)+(qcOutChannel[0].SfbMinSnrLdData[idx]>>1)) > ((energyLdR>>1)-(energyLdL>>1)) {
					checkAllowMoreHolesPEBand(peData.PEChannelData[1].SfbPe[idx])
					ahFlag[1][idx] = AvoidHoleNone
					qcOutChannel[1].SfbThresholdLdData[idx] = allowMoreHolesThrOffset + energyLdR
					actPe -= int(peData.PEChannelData[1].SfbPe[idx] >> peConstPartShift)
				} else if ahFlag[0][idx] != AvoidHoleNone &&
					((allowMoreHolesMsMargin>>1)+(qcOutChannel[1].SfbMinSnrLdData[idx]>>1)) > ((energyLdL>>1)-(energyLdR>>1)) {
					checkAllowMoreHolesPEBand(peData.PEChannelData[0].SfbPe[idx])
					ahFlag[0][idx] = AvoidHoleNone
					qcOutChannel[0].SfbThresholdLdData[idx] = allowMoreHolesThrOffset + energyLdL
					actPe -= int(peData.PEChannelData[0].SfbPe[idx] >> peConstPartShift)
				}
			}
			if actPe <= desiredPe {
				return
			}
		}
	}

	if actPe <= desiredPe {
		return
	}

	maxSfbSlots := 0
	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			maxSfbSlots += psyCh.MaxSfbPerGroup
		}
	}
	avgEnE := DfractBits - FixNormZD(FixpDBL(maxInt(0, maxSfbSlots-1)))

	var startSfb [2]int
	var sfbCnt [2]int
	var sfbPerGroup [2]int
	var maxSfbPerGroup [2]int
	maxSfb := 0
	minSfb := maxGroupedSFB
	avgEn := FixpDBL(0)
	minEnLD64 := FixpDBL(0)
	ahCnt := 0

	for ch := 0; ch < nChannels; ch++ {
		qcCh := qcOutChannel[ch]
		psyCh := psyOutChannel[ch]
		maxSfbPerGroup[ch] = psyCh.MaxSfbPerGroup
		sfbCnt[ch] = psyCh.SfbCnt
		sfbPerGroup[ch] = psyCh.SfbPerGroup
		maxSfb = maxInt(maxSfb, psyCh.MaxSfbPerGroup)
		if psyCh.LastWindowSequence != ShortWindow {
			startSfb[ch] = adjThrStateElement.AHParam.StartSfbL
		} else {
			startSfb[ch] = adjThrStateElement.AHParam.StartSfbS
		}
		minSfb = minInt(minSfb, startSfb[ch])

		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := startSfb[ch]; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				if ahFlag[ch][idx] != AvoidHoleNone &&
					qcCh.SfbWeightedEnergyLdData[idx] > qcCh.SfbThresholdLdData[idx] {
					minEnLD64 = minFixpDBL(minEnLD64, qcCh.SfbEnergyLdData[idx])
					avgEn += qcCh.SfbEnergy[idx] >> uint(avgEnE)
					ahCnt++
				}
			}
		}
	}

	avgEnLD64 := FixpDBL(0)
	if avgEn != 0 && ahCnt != 0 {
		avgEnLD64 = CalcLdData(avgEn) +
			FixpDBL(avgEnE<<(DfractBits-1-ldDataShift)) -
			CalcLdInt(ahCnt)
	}

	var enLD64 [allowMoreHolesNumEnergyLevels]FixpDBL
	for i := 0; i < allowMoreHolesNumEnergyLevels-1; i++ {
		enLD64[i] = minEnLD64 + FMultDD(avgEnLD64-minEnLD64, allowMoreHolesEnergyLevel[i])
	}
	enLD64[allowMoreHolesNumEnergyLevels-1] = minEnLD64 + (avgEnLD64 - minEnLD64)

	done := false
	enIdx := 0
	sfb := maxSfb - 1
	for !done {
		for ch := 0; ch < nChannels; ch++ {
			qcCh := qcOutChannel[ch]
			if sfb >= startSfb[ch] && sfb < maxSfbPerGroup[ch] {
				for sfbGrp := 0; sfbGrp < sfbCnt[ch]; sfbGrp += sfbPerGroup[ch] {
					idx := sfbGrp + sfb
					if ahFlag[ch][idx] != AvoidHoleNone &&
						qcCh.SfbEnergyLdData[idx] < enLD64[enIdx] {
						checkAllowMoreHolesPEBand(peData.PEChannelData[ch].SfbPe[idx])
						ahFlag[ch][idx] = AvoidHoleNone
						qcCh.SfbThresholdLdData[idx] = allowMoreHolesThrOffset + qcCh.SfbWeightedEnergyLdData[idx]
						actPe -= int(peData.PEChannelData[ch].SfbPe[idx] >> peConstPartShift)
					}
					if actPe <= desiredPe {
						return
					}
				}
			}
		}

		sfb--
		if sfb < minSfb {
			sfb = maxSfb
			enIdx++
			if enIdx >= allowMoreHolesNumEnergyLevels {
				done = true
			}
		}
	}
}

func FDKaacEncResetAHFlags(
	ahFlag *[2][maxGroupedSFB]uint8,
	psyOutChannel []*PsyOutChannel,
	nChannels int,
) {
	checkResetAHFlagsInputs(ahFlag, psyOutChannel, nChannels)

	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				if ahFlag[ch][idx] == AvoidHoleActive {
					ahFlag[ch][idx] = AvoidHoleInactive
				}
			}
		}
	}
}

func FDKaacEncBitresCalcBitFac(
	bitresBits int,
	maxBitresBits int,
	pe int,
	lastWindowSequence int,
	avgBits int,
	maxBitFac FixpDBL,
	adjThr *AdjThrState,
	adjThrChan *ATSElement,
) (FixpDBL, int) {
	checkBitresCalcInputs(bitresBits, maxBitresBits, pe, lastWindowSequence, avgBits, maxBitFac, adjThr, adjThrChan)

	bresParam := &adjThr.BRESParamLong
	bitsaveSlope := bitSaveSlopeLong
	bitspendSlope := bitSpendSlopeLong
	if lastWindowSequence == ShortWindow {
		bresParam = &adjThr.BRESParamShort
		bitsaveSlope = bitSaveSlopeShort
		bitspendSlope = bitSpendSlopeShort
	}

	fillLevelFix := MaxValDBL
	if bitresBits < maxBitresBits {
		fillLevelFix = fDivNorm(FixpDBL(bitresBits), FixpDBL(maxBitresBits))
	}

	pex := maxInt(pe, adjThrChan.PeMin)
	pex = minInt(pex, adjThrChan.PeMax)

	bitSave := fdkaacEncCalcBitSave(
		fillLevelFix,
		bresParam.ClipSaveLow,
		bresParam.ClipSaveHigh,
		bresParam.MinBitSave,
		bresParam.MaxBitSave,
		bitsaveSlope,
	)
	bitSpend := fdkaacEncCalcBitSpend(
		fillLevelFix,
		bresParam.ClipSpendLow,
		bresParam.ClipSpendHigh,
		bresParam.MinBitSpend,
		bresParam.MaxBitSpend,
		bitspendSlope,
	)

	slope := schurDiv(FixpDBL(pex-adjThrChan.PeMin), FixpDBL(adjThrChan.PeMax-adjThrChan.PeMin), 31)

	bitresFac := (MaxValDBL >> 1) - (bitSave >> 1)
	bitresFacE := 1
	bitresFac = FMultAddDiv2DD(bitresFac, slope, bitSpend+bitSave)

	fillLevel, fillLevelE := fDivNormExp(FixpDBL(bitresBits), FixpDBL(avgBits))
	if fillLevelE < 0 {
		fillLevel = ScaleValueDBL(fillLevel, fillLevelE)
		fillLevelE = 0
	}
	fillLevel >>= 1
	fillLevelE++
	fillLevel += ScaleValueDBL(bitresFillBias07, -fillLevelE)
	if ScaleValueDBL(bitresFac, -fillLevelE+1) > fillLevel {
		bitresFac = fillLevel
		bitresFacE = fillLevelE
	}

	if ScaleValueDBL(bitresFac, bitresFacE-(DfractBits-1-qBitFac)) > maxBitFac {
		bitresFac = maxBitFac
		bitresFacE = DfractBits - 1 - qBitFac
	}

	fdkaacEncAdjustPEMinMax(pe, &adjThrChan.PeMin, &adjThrChan.PeMax)
	return bitresFac, bitresFacE
}

func FDKaacEncDistributeBits(
	adjThrState *AdjThrState,
	adjThrStateElement *ATSElement,
	psyOutChannel []*PsyOutChannel,
	peData *PEData,
	nChannels int,
	commonWindow int,
	grantedDynBits int,
	bitresBits int,
	maxBitresBits int,
	maxBitFac FixpDBL,
	bitResMode BitresMode,
) (int, int) {
	checkDistributeBitsInputs(adjThrState, adjThrStateElement, psyOutChannel, peData, nChannels, grantedDynBits, bitresBits, maxBitresBits, maxBitFac, bitResMode)

	noRedPE := int(peData.Pe)
	curWindowSequence := LongWindow
	if nChannels == 2 {
		if psyOutChannel[0].LastWindowSequence == ShortWindow || psyOutChannel[1].LastWindowSequence == ShortWindow {
			curWindowSequence = ShortWindow
		}
	} else {
		curWindowSequence = psyOutChannel[0].LastWindowSequence
	}
	_ = commonWindow

	grantedPE := 0
	if grantedDynBits >= 1 {
		if bitResMode != BitresModeFull {
			grantedPE = fdkaacEncBits2PE2(grantedDynBits, adjThrStateElement.Bits2PeFactorM, adjThrStateElement.Bits2PeFactorE)
		} else {
			bitFactor, bitFactorE := FDKaacEncBitresCalcBitFac(
				bitresBits,
				maxBitresBits,
				noRedPE,
				curWindowSequence,
				grantedDynBits,
				maxBitFac,
				adjThrState,
				adjThrStateElement,
			)
			grantedPE = fdkaacEncBits2PE2(
				grantedDynBits,
				FMultDD(bitFactor, adjThrStateElement.Bits2PeFactorM),
				adjThrStateElement.Bits2PeFactorE+bitFactorE,
			)
		}
	}

	switch bitResMode {
	case BitresModeDisabled, BitresModeReduced:
		adjThrStateElement.PeCorrectionFactorM, adjThrStateElement.PeCorrectionFactorE =
			fdkaacEncCalcPECorrectionLowBitres(
				adjThrStateElement.PeCorrectionFactorM,
				adjThrStateElement.PeLast,
				adjThrStateElement.DynBitsLast,
				bitresBits,
				nChannels,
				adjThrStateElement.Bits2PeFactorM,
				adjThrStateElement.Bits2PeFactorE,
			)
	case BitresModeFull:
		fallthrough
	default:
		adjThrStateElement.PeCorrectionFactorM, adjThrStateElement.PeCorrectionFactorE =
			fdkaacEncCalcPECorrection(
				adjThrStateElement.PeCorrectionFactorM,
				minInt(grantedPE, noRedPE),
				adjThrStateElement.PeLast,
				adjThrStateElement.DynBitsLast,
				adjThrStateElement.Bits2PeFactorM,
				adjThrStateElement.Bits2PeFactorE,
			)
	}

	if grantedPE > int(MaxValDBL)>>qAvgBits {
		panic("fdkaac: PE grant exceeds fixed-point range")
	}
	grantedPECorr := int(FMultDD(
		FixpDBL(grantedPE<<qAvgBits),
		adjThrStateElement.PeCorrectionFactorM,
	) >> uint(qAvgBits-adjThrStateElement.PeCorrectionFactorE))

	adjThrStateElement.PeLast = grantedPE
	adjThrStateElement.DynBitsLast = -1
	return grantedPE, grantedPECorr
}

func fdkaacEncBits2PE2(bits int, factorM FixpDBL, factorE int) int {
	if bits < 0 || factorM < 0 || factorE < 0 || factorE > qAvgBits {
		panic("fdkaac: invalid bits-to-PE input")
	}
	if bits == 0 || factorM == 0 {
		return 0
	}
	if bits > int(MaxValDBL)>>qAvgBits {
		panic("fdkaac: bits-to-PE input exceeds fixed-point range")
	}
	return int(FMultDD(factorM, FixpDBL(bits<<qAvgBits)) >> uint(qAvgBits-factorE))
}

func fdkaacEncCalcBitSave(
	fillLevel FixpDBL,
	clipLow FixpDBL,
	clipHigh FixpDBL,
	minBitSave FixpDBL,
	maxBitSave FixpDBL,
	bitsaveSlope FixpDBL,
) FixpDBL {
	fillLevel = maxFixpDBL(fillLevel, clipLow)
	fillLevel = minFixpDBL(fillLevel, clipHigh)
	_ = minBitSave
	return maxBitSave - FMultDD(fillLevel-clipLow, bitsaveSlope)
}

func fdkaacEncCalcBitSpend(
	fillLevel FixpDBL,
	clipLow FixpDBL,
	clipHigh FixpDBL,
	minBitSpend FixpDBL,
	maxBitSpend FixpDBL,
	bitspendSlope FixpDBL,
) FixpDBL {
	fillLevel = maxFixpDBL(fillLevel, clipLow)
	fillLevel = minFixpDBL(fillLevel, clipHigh)
	_ = maxBitSpend
	return minBitSpend + FMultDD(fillLevel-clipLow, bitspendSlope)
}

func fdkaacEncAdjustPEMinMax(currPE int, peMin *int, peMax *int) {
	if peMin == nil || peMax == nil {
		panic("fdkaac: nil PE bounds")
	}
	if currPE < 0 || *peMin < 0 || *peMax <= *peMin {
		panic("fdkaac: invalid PE bounds")
	}

	minDiff := FMultI(peAdjustMinDiff, currPE)
	if currPE > *peMax {
		diff := currPE - *peMax
		*peMin += FMultI(peAdjustMinFacHi, diff)
		*peMax += FMultI(peAdjustMaxFacHi, diff)
	} else if currPE < *peMin {
		diff := *peMin - currPE
		*peMin -= FMultI(peAdjustMinFacLo, diff)
		*peMax -= FMultI(peAdjustMaxFacLo, diff)
	} else {
		*peMin += FMultI(peAdjustMinFacHi, currPE-*peMin)
		*peMax -= FMultI(peAdjustMaxFacLo, *peMax-currPE)
	}

	if (*peMax - *peMin) < minDiff {
		peMaxFix := *peMax
		peMinFix := *peMin
		partLo := FixpDBL(maxInt(0, currPE-peMinFix))
		partHi := FixpDBL(maxInt(0, peMaxFix-currPE))
		denom := partLo + partHi
		if denom <= 0 {
			panic("fdkaac: collapsed PE bounds")
		}

		peMaxFix = currPE + FMultI(fDivNorm(partHi, denom), minDiff)
		peMinFix = currPE - FMultI(fDivNorm(partLo, denom), minDiff)
		peMinFix = maxInt(0, peMinFix)

		*peMax = peMaxFix
		*peMin = peMinFix
	}
}

func fdkaacEncCalcPECorrection(
	correctionFacM FixpDBL,
	peAct int,
	peLast int,
	bitsLast int,
	bits2PeFactorM FixpDBL,
	bits2PeFactorE int,
) (FixpDBL, int) {
	if bitsLast > 0 &&
		2*peAct < 3*peLast &&
		10*peAct >= 7*peLast &&
		fdkaacEncBits2PE2(bitsLast, FMultDD(peCorrection12Over2, bits2PeFactorM), bits2PeFactorE+1) > peLast &&
		fdkaacEncBits2PE2(bitsLast, FMultDD(peCorrection065, bits2PeFactorM), bits2PeFactorE) < peLast {
		corrFac := correctionFacM
		denum := FixpDBL(fdkaacEncBits2PE2(bitsLast, bits2PeFactorM, bits2PeFactorE))
		if peLast <= 0 || denum <= 0 {
			return peCorrectionHalf, 1
		}

		newFac, scaling := fDivNormExp(FixpDBL(peLast), denum)
		if FixpDBL(peLast) <= denum {
			newFac = maxFixpDBL(
				ScaleValueDBL(
					minFixpDBL(
						FMultDD(peCorrection11Over2, newFac),
						ScaleValueDBL(peCorrectionHalf, -scaling),
					),
					scaling,
				),
				peCorrection085Over2F,
			)
		} else {
			newFac = maxFixpDBL(
				minFixpDBL(
					ScaleValueDBL(FMultDD(peCorrection09Over2, newFac), scaling),
					peCorrection115Over2,
				),
				peCorrectionHalf,
			)
		}

		if (newFac > peCorrectionHalf && corrFac < peCorrectionHalf) ||
			(newFac < peCorrectionHalf && corrFac > peCorrectionHalf) {
			corrFac = peCorrectionHalf
		}

		if (corrFac < peCorrectionHalf && newFac < corrFac) ||
			(corrFac > peCorrectionHalf && newFac > corrFac) {
			corrFac = FMultDD(peCorrection085, corrFac) + FMultDD(peCorrection015, newFac)
		} else {
			corrFac = FMultDD(peCorrection07, corrFac) + FMultDD(peCorrection03, newFac)
		}

		corrFac = maxFixpDBL(minFixpDBL(corrFac, peCorrection115Over2), peCorrection085Over2D)
		return corrFac, 1
	}

	return peCorrectionHalf, 1
}

func fdkaacEncCalcPECorrectionLowBitres(
	correctionFacM FixpDBL,
	peLast int,
	bitsLast int,
	bitresLevel int,
	nChannels int,
	bits2PeFactorM FixpDBL,
	bits2PeFactorE int,
) (FixpDBL, int) {
	if bitsLast > 0 {
		bitsBalLast := peLast - fdkaacEncBits2PE2(bitsLast, bits2PeFactorM, bits2PeFactorE)
		headroomBits := 0
		if bitresLevel < 50*nChannels {
			headroomBits = 100 * nChannels
		}
		headroom := fdkaacEncBits2PE2(headroomBits, bits2PeFactorM, bits2PeFactorE)
		denominator := FixpDBL(fdkaacEncBits2PE2(bitresLevel, bits2PeFactorM, bits2PeFactorE) + headroom)
		if denominator <= 0 {
			panic("fdkaac: invalid low-reservoir PE correction denominator")
		}

		balMinusHeadroom := bitsBalLast - headroom
		diff := FixpDBL(0)
		scaling := 0
		if balMinusHeadroom >= 0 {
			var div FixpDBL
			div, scaling = fDivNormExp(FixpDBL(balMinusHeadroom), denominator)
			diff = FMultDD(lowBitresCorrectionAmp, div)
		} else {
			var div FixpDBL
			div, scaling = fDivNormExp(FixpDBL(-balMinusHeadroom), denominator)
			diff = -FMultDD(lowBitresCorrectionAmp, div)
		}

		scaling--
		if scaling <= 0 {
			diff = maxFixpDBL(
				minFixpDBL(diff>>uint(-scaling), lowBitresMaxDiff>>1),
				(-lowBitresMaxDiff)>>1,
			)
		} else {
			diff = maxFixpDBL(
				minFixpDBL(diff, lowBitresMaxDiff>>uint(1+scaling)),
				(-lowBitresMaxDiff)>>uint(1+scaling),
			) << uint(scaling)
		}

		return maxFixpDBL(
			minFixpDBL(correctionFacM+diff, peCorrectionHalf),
			lowBitresCorrectionMin,
		), 1
	}

	return lowBitresCorrectionMin, 1
}

func checkThresholdAdjustmentInputs(qcOutChannel []*QCOutChannel, psyOutChannel []*PsyOutChannel, nChannels int) {
	if nChannels <= 0 || nChannels > 2 {
		panic("fdkaac: invalid threshold adjustment channel count")
	}
	if len(qcOutChannel) < nChannels || len(psyOutChannel) < nChannels {
		panic("fdkaac: short threshold adjustment channel inputs")
	}
	for ch := 0; ch < nChannels; ch++ {
		if qcOutChannel[ch] == nil {
			panic("fdkaac: nil threshold adjustment qc output")
		}
		if psyOutChannel[ch] == nil {
			panic("fdkaac: nil threshold adjustment psy output")
		}
		checkPEChannelShape(psyOutChannel[ch])
	}
}

func checkAvoidHoleInputs(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	ahFlag *[2][maxGroupedSFB]uint8,
	toolsInfo *ToolsInfo,
	nChannels int,
	ahParam *AHParam,
) {
	checkThresholdAdjustmentInputs(qcOutChannel, psyOutChannel, nChannels)
	if ahFlag == nil {
		panic("fdkaac: nil avoid-hole flag scratch")
	}
	if toolsInfo == nil {
		panic("fdkaac: nil avoid-hole tools info")
	}
	if ahParam == nil {
		panic("fdkaac: nil avoid-hole parameter")
	}
	if nChannels == 2 {
		left := psyOutChannel[0]
		right := psyOutChannel[1]
		if left.SfbCnt != right.SfbCnt || left.SfbPerGroup != right.SfbPerGroup || left.MaxSfbPerGroup != right.MaxSfbPerGroup {
			panic("fdkaac: mismatched avoid-hole stereo bands")
		}
	}
}

func checkPENoAHInputs(peData *PEData, ahFlag *[2][maxGroupedSFB]uint8, psyOutChannel []*PsyOutChannel, nChannels int) {
	if peData == nil {
		panic("fdkaac: nil no-AH PE data")
	}
	if ahFlag == nil {
		panic("fdkaac: nil no-AH flag scratch")
	}
	if nChannels <= 0 || nChannels > 2 {
		panic("fdkaac: invalid no-AH channel count")
	}
	if len(psyOutChannel) < nChannels {
		panic("fdkaac: short no-AH channel inputs")
	}
	for ch := 0; ch < nChannels; ch++ {
		if psyOutChannel[ch] == nil {
			panic("fdkaac: nil no-AH psy output")
		}
		checkPEChannelShape(psyOutChannel[ch])
	}
}

func checkThresholdReductionInputs(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	ahFlag *[2][maxGroupedSFB]uint8,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	nChannels int,
	redValM FixpDBL,
	redValE int,
) {
	checkThresholdAdjustmentInputs(qcOutChannel, psyOutChannel, nChannels)
	if ahFlag == nil {
		panic("fdkaac: nil threshold reduction avoid-hole flags")
	}
	if thrExp == nil {
		panic("fdkaac: nil threshold reduction exponent scratch")
	}
	if redValM < 0 || redValE < -DfractBits || redValE > DfractBits {
		panic("fdkaac: invalid threshold reduction value")
	}
}

func checkChaosMeasureInputs(psyOutChannel *PsyOutChannel, sfbFormFactorLdData []FixpDBL) {
	if psyOutChannel == nil {
		panic("fdkaac: nil chaos-measure psy output")
	}
	checkPEChannelShape(psyOutChannel)
	if len(sfbFormFactorLdData) < psyOutChannel.SfbCnt {
		panic("fdkaac: short chaos-measure form-factor data")
	}
}

func checkThresholdReductionVBRInputs(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	ahFlag *[2][maxGroupedSFB]uint8,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	nChannels int,
	vbrQualFactor FixpDBL,
	chaosMeasureOld *FixpDBL,
) {
	checkThresholdAdjustmentInputs(qcOutChannel, psyOutChannel, nChannels)
	if ahFlag == nil {
		panic("fdkaac: nil VBR threshold reduction avoid-hole flags")
	}
	if thrExp == nil {
		panic("fdkaac: nil VBR threshold reduction exponent scratch")
	}
	if chaosMeasureOld == nil {
		panic("fdkaac: nil VBR threshold reduction chaos history")
	}
	if vbrQualFactor < 0 {
		panic("fdkaac: invalid VBR quality factor")
	}
	if psyOutChannel[0].LastWindowSequence == ShortWindow {
		groupCnt := 0
		for sfbGrp := 0; sfbGrp < psyOutChannel[0].SfbCnt; sfbGrp += psyOutChannel[0].SfbPerGroup {
			if groupCnt >= transFac {
				panic("fdkaac: too many VBR short groups")
			}
			groupLen := psyOutChannel[0].GroupLen[groupCnt]
			if groupLen < 0 || groupLen >= len(adjThrInvInt) || groupLen >= len(adjThrInvSqrt4) {
				panic("fdkaac: invalid VBR short group length")
			}
			groupCnt++
		}
	}
}

func checkCorrectThresholdInputs(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	peData *PEData,
	ahFlag *[2][maxGroupedSFB]uint8,
	thrExp *[2][maxGroupedSFB]FixpDBL,
	scratch *CorrectThresholdScratch,
	nChannels int,
	redValM FixpDBL,
	redValE int,
) {
	checkThresholdAdjustmentInputs(qcOutChannel, psyOutChannel, nChannels)
	if peData == nil {
		panic("fdkaac: nil correct-threshold PE data")
	}
	if ahFlag == nil {
		panic("fdkaac: nil correct-threshold avoid-hole flags")
	}
	if thrExp == nil {
		panic("fdkaac: nil correct-threshold exponent scratch")
	}
	if scratch == nil {
		panic("fdkaac: nil correct-threshold scratch")
	}
	if redValM < 0 || redValE < -DfractBits || redValE > DfractBits {
		panic("fdkaac: invalid correct-threshold reduction value")
	}
}

func checkReduceMinSnrInputs(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	peData *PEData,
	ahFlag *[2][maxGroupedSFB]uint8,
	nChannels int,
	desiredPe int,
	redPeGlobal *int,
) {
	checkThresholdAdjustmentInputs(qcOutChannel, psyOutChannel, nChannels)
	if peData == nil {
		panic("fdkaac: nil reduce-min-SNR PE data")
	}
	if ahFlag == nil {
		panic("fdkaac: nil reduce-min-SNR avoid-hole flags")
	}
	if redPeGlobal == nil {
		panic("fdkaac: nil reduce-min-SNR PE total")
	}
	if desiredPe < 0 || *redPeGlobal < 0 {
		panic("fdkaac: invalid reduce-min-SNR PE target")
	}
}

func checkAllowMoreHolesInputs(
	qcOutChannel []*QCOutChannel,
	psyOutChannel []*PsyOutChannel,
	peData *PEData,
	toolsInfo *ToolsInfo,
	adjThrStateElement *ATSElement,
	ahFlag *[2][maxGroupedSFB]uint8,
	nChannels int,
	desiredPe int,
	currentPe int,
) {
	checkThresholdAdjustmentInputs(qcOutChannel, psyOutChannel, nChannels)
	if peData == nil {
		panic("fdkaac: nil allow-more-holes PE data")
	}
	if toolsInfo == nil {
		panic("fdkaac: nil allow-more-holes tools info")
	}
	if adjThrStateElement == nil {
		panic("fdkaac: nil allow-more-holes state")
	}
	if ahFlag == nil {
		panic("fdkaac: nil allow-more-holes avoid-hole flags")
	}
	if desiredPe < 0 || currentPe < 0 {
		panic("fdkaac: invalid allow-more-holes PE target")
	}
	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		startSfb := adjThrStateElement.AHParam.StartSfbL
		if psyCh.LastWindowSequence == ShortWindow {
			startSfb = adjThrStateElement.AHParam.StartSfbS
		}
		if startSfb < 0 || startSfb > psyCh.SfbPerGroup {
			panic("fdkaac: invalid allow-more-holes start band")
		}
	}
	if nChannels == 2 {
		left := psyOutChannel[0]
		right := psyOutChannel[1]
		if left.SfbCnt != right.SfbCnt || left.SfbPerGroup != right.SfbPerGroup || left.MaxSfbPerGroup != right.MaxSfbPerGroup {
			panic("fdkaac: mismatched allow-more-holes stereo bands")
		}
	}
}

func checkAllowMoreHolesPEBand(pe FixpDBL) {
	if pe < 0 {
		panic("fdkaac: invalid allow-more-holes PE band")
	}
}

func checkResetAHFlagsInputs(
	ahFlag *[2][maxGroupedSFB]uint8,
	psyOutChannel []*PsyOutChannel,
	nChannels int,
) {
	if ahFlag == nil {
		panic("fdkaac: nil AH reset flags")
	}
	if nChannels <= 0 || nChannels > 2 {
		panic("fdkaac: invalid AH reset channel count")
	}
	if len(psyOutChannel) < nChannels {
		panic("fdkaac: short AH reset psy inputs")
	}
	for ch := 0; ch < nChannels; ch++ {
		if psyOutChannel[ch] == nil {
			panic("fdkaac: nil AH reset psy output")
		}
		checkPEChannelShape(psyOutChannel[ch])
	}
}

func checkBitresCalcInputs(
	bitresBits int,
	maxBitresBits int,
	pe int,
	lastWindowSequence int,
	avgBits int,
	maxBitFac FixpDBL,
	adjThr *AdjThrState,
	adjThrChan *ATSElement,
) {
	if adjThr == nil {
		panic("fdkaac: nil bit reservoir state")
	}
	if adjThrChan == nil {
		panic("fdkaac: nil bit reservoir element")
	}
	if bitresBits < 0 || maxBitresBits <= 0 || avgBits <= 0 {
		panic("fdkaac: invalid bit reservoir levels")
	}
	if pe < 0 || maxBitFac < 0 {
		panic("fdkaac: invalid bit reservoir PE input")
	}
	if !validPEWindowSequence(lastWindowSequence) {
		panic("fdkaac: invalid bit reservoir block type")
	}
	if adjThrChan.PeMin < 0 || adjThrChan.PeMax <= adjThrChan.PeMin {
		panic("fdkaac: invalid bit reservoir PE bounds")
	}
	checkBRESParam(adjThr.BRESParamLong)
	checkBRESParam(adjThr.BRESParamShort)
}

func checkDistributeBitsInputs(
	adjThrState *AdjThrState,
	adjThrStateElement *ATSElement,
	psyOutChannel []*PsyOutChannel,
	peData *PEData,
	nChannels int,
	grantedDynBits int,
	bitresBits int,
	maxBitresBits int,
	maxBitFac FixpDBL,
	bitResMode BitresMode,
) {
	if adjThrState == nil {
		panic("fdkaac: nil bit distribution state")
	}
	if adjThrStateElement == nil {
		panic("fdkaac: nil bit distribution element")
	}
	if peData == nil {
		panic("fdkaac: nil bit distribution PE data")
	}
	if nChannels <= 0 || nChannels > 2 {
		panic("fdkaac: invalid bit distribution channel count")
	}
	if len(psyOutChannel) < nChannels {
		panic("fdkaac: short bit distribution channel inputs")
	}
	for ch := 0; ch < nChannels; ch++ {
		if psyOutChannel[ch] == nil {
			panic("fdkaac: nil bit distribution psy output")
		}
		if !validPEWindowSequence(psyOutChannel[ch].LastWindowSequence) {
			panic("fdkaac: invalid bit distribution block type")
		}
	}
	if peData.Pe < 0 || grantedDynBits < 0 || bitresBits < 0 || maxBitresBits <= 0 || maxBitFac < 0 {
		panic("fdkaac: invalid bit distribution level")
	}
	if bitResMode != BitresModeFull && bitResMode != BitresModeReduced && bitResMode != BitresModeDisabled {
		panic("fdkaac: invalid bit reservoir mode")
	}
	if adjThrStateElement.Bits2PeFactorM <= 0 || adjThrStateElement.Bits2PeFactorE < 0 || adjThrStateElement.Bits2PeFactorE > qAvgBits {
		panic("fdkaac: invalid bits-to-PE factor")
	}
	if adjThrStateElement.PeCorrectionFactorM < 0 || adjThrStateElement.PeCorrectionFactorE < 0 || adjThrStateElement.PeCorrectionFactorE > qAvgBits {
		panic("fdkaac: invalid PE correction factor")
	}
	if bitResMode == BitresModeFull && (adjThrStateElement.PeMin < 0 || adjThrStateElement.PeMax <= adjThrStateElement.PeMin) {
		panic("fdkaac: invalid bit distribution PE bounds")
	}
	checkBRESParam(adjThrState.BRESParamLong)
	checkBRESParam(adjThrState.BRESParamShort)
}

func checkBRESParam(param BRESParam) {
	if param.ClipSaveHigh <= param.ClipSaveLow || param.ClipSpendHigh <= param.ClipSpendLow {
		panic("fdkaac: invalid bit reservoir parameter clips")
	}
	if param.MaxBitSave < param.MinBitSave || param.MaxBitSpend < param.MinBitSpend {
		panic("fdkaac: invalid bit reservoir parameter limits")
	}
}

func validPEWindowSequence(windowSequence int) bool {
	return windowSequence == LongWindow ||
		windowSequence == StartWindow ||
		windowSequence == ShortWindow ||
		windowSequence == StopWindow
}

func FDKaacEncPECalculation(
	peData *PEData,
	psyOutChannel []*PsyOutChannel,
	qcOutChannel []*QCOutChannel,
	toolsInfo *ToolsInfo,
	adjThrStateElement *ATSElement,
	nChannels int,
) {
	checkPECalculationInputs(peData, psyOutChannel, qcOutChannel, toolsInfo, adjThrStateElement, nChannels)

	fdkaacEncPreparePE(peData, psyOutChannel, qcOutChannel, nChannels, adjThrStateElement.PeOffset)
	fdkaacEncCalcWeighting(peData, psyOutChannel, qcOutChannel, toolsInfo, adjThrStateElement, nChannels, 1)

	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		qcCh := qcOutChannel[ch]
		for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
			for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
				idx := sfbGrp + sfb
				qcCh.SfbWeightedEnergyLdData[idx] = qcCh.SfbEnergyLdData[idx] - qcCh.SfbEnFacLd[idx]
				qcCh.SfbThresholdLdData[idx] -= qcCh.SfbEnFacLd[idx]
			}
		}
	}

	fdkaacEncCalcPE(psyOutChannel, qcOutChannel, peData, nChannels)
}

func fdkaacEncPreparePE(
	peData *PEData,
	psyOutChannel []*PsyOutChannel,
	qcOutChannel []*QCOutChannel,
	nChannels int,
	peOffset int,
) {
	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		FDKaacEncPrepareSfbPe(
			&peData.PEChannelData[ch],
			psyCh.SfbEnergyLdData[:],
			psyCh.SfbThresholdLdData[:],
			qcOutChannel[ch].SfbFormFactorLdData[:],
			psyCh.SfbOffsets[:],
			psyCh.SfbCnt,
			psyCh.SfbPerGroup,
			psyCh.MaxSfbPerGroup,
		)
	}
	peData.Offset = peOffset
}

func fdkaacEncCalcWeighting(
	peData *PEData,
	psyOutChannel []*PsyOutChannel,
	qcOutChannel []*QCOutChannel,
	toolsInfo *ToolsInfo,
	adjThrStateElement *ATSElement,
	nChannels int,
	usePatchTool int,
) {
	noShortWindowInFrame := adjThrTrue
	exePatchM := adjThrFalse

	for ch := 0; ch < nChannels; ch++ {
		if psyOutChannel[ch].LastWindowSequence == ShortWindow {
			noShortWindowInFrame = adjThrFalse
		}
		clear(qcOutChannel[ch].SfbEnFacLd[:])
	}

	if usePatchTool == 0 {
		return
	}

	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]

		if noShortWindowInFrame != 0 {
			nrgSum14 := FixpDBL(0)
			nrgSum12 := FixpDBL(0)
			nrgSum34 := FixpDBL(0)
			nrgTotal := FixpDBL(0)
			nLinesSum := 0

			for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
				for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
					idx := sfbGrp + sfb
					nrgFac12 := CalcInvLdData(psyCh.SfbEnergyLdData[idx] >> 1)
					nrgFac14 := CalcInvLdData(psyCh.SfbEnergyLdData[idx] >> 2)

					nLinesSum += peData.PEChannelData[ch].SfbNLines[idx]
					nrgTotal += psyCh.SfbEnergy[idx] >> 6
					nrgSum12 += nrgFac12 >> 6
					nrgSum14 += nrgFac14 >> 6
					nrgSum34 += FMultDD(nrgFac14, nrgFac12) >> 6
				}
			}

			nrgTotalLd := CalcLdData(nrgTotal)
			nrgFacLd14 := CalcLdData(nrgSum14) - nrgTotalLd
			nrgFacLd12 := CalcLdData(nrgSum12) - nrgTotalLd
			nrgFacLd34 := CalcLdData(nrgSum34) - nrgTotalLd

			chaos := fDivNorm(FixpDBL(nLinesSum), FixpDBL(psyCh.SfbOffsets[psyCh.SfbCnt]))
			if chaos < chaosMeasureMin {
				chaos = chaosMeasureMin
			}
			adjThrStateElement.ChaosMeasureEnFac[ch] = chaos

			usePatch := adjThrFalse
			if chaos > chaosMeasurePatchUse {
				usePatch = adjThrTrue
			}
			exePatch := adjThrFalse
			if usePatch != 0 && adjThrStateElement.LastEnFacPatch[ch] != 0 {
				exePatch = adjThrTrue
			}

			for sfbGrp := 0; sfbGrp < psyCh.SfbCnt; sfbGrp += psyCh.SfbPerGroup {
				for sfb := 0; sfb < psyCh.MaxSfbPerGroup; sfb++ {
					idx := sfbGrp + sfb
					sfbExePatch := exePatch
					if ch == 1 && toolsInfo.MsMask[idx] != 0 {
						sfbExePatch = exePatchM
					}

					if sfbExePatch != 0 && psyCh.SfbEnergy[idx] > 0 {
						if chaos > chaosMeasurePatch34 {
							qcOutChannel[ch].SfbEnFacLd[idx] = (nrgFacLd14 + psyCh.SfbEnergyLdData[idx] + (psyCh.SfbEnergyLdData[idx] >> 1)) >> 1
						} else if chaos > chaosMeasurePatch12 {
							qcOutChannel[ch].SfbEnFacLd[idx] = (nrgFacLd12 + psyCh.SfbEnergyLdData[idx]) >> 1
						} else {
							qcOutChannel[ch].SfbEnFacLd[idx] = (nrgFacLd34 + (psyCh.SfbEnergyLdData[idx] >> 1)) >> 1
						}
						if qcOutChannel[ch].SfbEnFacLd[idx] > 0 {
							qcOutChannel[ch].SfbEnFacLd[idx] = 0
						}
					}
				}
			}

			adjThrStateElement.LastEnFacPatch[ch] = usePatch
			exePatchM = exePatch
			continue
		}

		adjThrStateElement.ChaosMeasureEnFac[ch] = chaosMeasureShort
		adjThrStateElement.LastEnFacPatch[ch] = adjThrTrue
	}
}

func fdkaacEncCalcPE(
	psyOutChannel []*PsyOutChannel,
	qcOutChannel []*QCOutChannel,
	peData *PEData,
	nChannels int,
) {
	peData.Pe = FixpDBL(peData.Offset)
	peData.ConstPart = 0
	peData.NActiveLines = 0

	for ch := 0; ch < nChannels; ch++ {
		psyCh := psyOutChannel[ch]
		peChanData := &peData.PEChannelData[ch]
		FDKaacEncCalcSfbPe(
			peChanData,
			qcOutChannel[ch].SfbWeightedEnergyLdData[:],
			qcOutChannel[ch].SfbThresholdLdData[:],
			psyCh.SfbCnt,
			psyCh.SfbPerGroup,
			psyCh.MaxSfbPerGroup,
			psyCh.IsBook[:],
			psyCh.IsScale[:],
		)

		peData.Pe += peChanData.Pe
		peData.ConstPart += peChanData.ConstPart
		peData.NActiveLines += peChanData.NActiveLines
	}
}

func checkPECalculationInputs(
	peData *PEData,
	psyOutChannel []*PsyOutChannel,
	qcOutChannel []*QCOutChannel,
	toolsInfo *ToolsInfo,
	adjThrStateElement *ATSElement,
	nChannels int,
) {
	if peData == nil {
		panic("fdkaac: nil PE data")
	}
	if toolsInfo == nil {
		panic("fdkaac: nil PE tools info")
	}
	if adjThrStateElement == nil {
		panic("fdkaac: nil PE adjustment state")
	}
	if nChannels <= 0 || nChannels > 2 {
		panic("fdkaac: invalid PE channel count")
	}
	if len(psyOutChannel) < nChannels || len(qcOutChannel) < nChannels {
		panic("fdkaac: short PE channel inputs")
	}
	for ch := 0; ch < nChannels; ch++ {
		if psyOutChannel[ch] == nil {
			panic("fdkaac: nil PE psy output")
		}
		if qcOutChannel[ch] == nil {
			panic("fdkaac: nil PE qc output")
		}
		checkPEChannelShape(psyOutChannel[ch])
	}
}

func checkPEChannelShape(psyCh *PsyOutChannel) {
	if psyCh.SfbCnt <= 0 || psyCh.SfbCnt > maxGroupedSFB || psyCh.SfbPerGroup <= 0 || psyCh.SfbCnt%psyCh.SfbPerGroup != 0 {
		panic("fdkaac: invalid PE channel band count")
	}
	if psyCh.MaxSfbPerGroup <= 0 || psyCh.MaxSfbPerGroup > psyCh.SfbPerGroup {
		panic("fdkaac: invalid PE channel group width")
	}
	if psyCh.LastWindowSequence != LongWindow && psyCh.LastWindowSequence != StartWindow && psyCh.LastWindowSequence != ShortWindow && psyCh.LastWindowSequence != StopWindow {
		panic("fdkaac: invalid PE channel block type")
	}
	if psyCh.SfbOffsets[0] < 0 {
		panic("fdkaac: invalid PE channel offset")
	}
	for sfb := 0; sfb < psyCh.SfbCnt; sfb++ {
		if psyCh.SfbOffsets[sfb+1] < psyCh.SfbOffsets[sfb] {
			panic("fdkaac: invalid PE channel offset")
		}
	}
	if psyCh.SfbOffsets[psyCh.SfbCnt] <= 0 {
		panic("fdkaac: empty PE channel spectrum")
	}
}
