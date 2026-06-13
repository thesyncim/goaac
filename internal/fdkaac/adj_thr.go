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
