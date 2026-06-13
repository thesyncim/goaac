package fdkaac

type TNSSubblockInfo struct {
	TnsActive      [maxTnsFilters]int
	PredictionGain [maxTnsFilters]int
}

type TNSDataShort struct {
	SubBlockInfo   [transFac]TNSSubblockInfo
	RatioMultTable [transFac][maxSFBShort]FixpDBL
}

type TNSDataLong struct {
	SubBlockInfo   TNSSubblockInfo
	RatioMultTable [maxSFBLong]FixpDBL
}

type TNSDataRaw struct {
	Long  TNSDataLong
	Short TNSDataShort
}

type TNSData struct {
	NumOfSubblocks  int
	DataRaw         TNSDataRaw
	TnsMaxScaleSpec int
	FiltersMerged   int
}

func FDKaacEncTnsSync(
	tnsDataDest *TNSData,
	tnsDataSrc *TNSData,
	tnsInfoDest *TNSInfo,
	tnsInfoSrc *TNSInfo,
	blockTypeDest int,
	blockTypeSrc int,
	tC *TNSConfig,
) {
	checkTnsSyncInputs(tnsDataDest, tnsDataSrc, tnsInfoDest, tnsInfoSrc, blockTypeDest, blockTypeSrc, tC)

	if (blockTypeSrc == ShortWindow && blockTypeDest != ShortWindow) ||
		(blockTypeDest == ShortWindow && blockTypeSrc != ShortWindow) {
		return
	}

	nWindows := 1
	if blockTypeDest == ShortWindow {
		nWindows = transFac
	}

	for w := 0; w < nWindows; w++ {
		sbInfoSrc := tnsSubblockInfo(tnsDataSrc, blockTypeSrc, w)
		sbInfoDest := tnsSubblockInfo(tnsDataDest, blockTypeDest, w)
		doSync := 1
		absDiffSum := 0

		if sbInfoDest.TnsActive[tnsHiFilt] == 0 && sbInfoSrc.TnsActive[tnsHiFilt] == 0 {
			continue
		}

		for i := 0; i < tC.MaxOrder; i++ {
			absDiff := absInt(tnsInfoDest.Coef[w][tnsHiFilt][i] - tnsInfoSrc.Coef[w][tnsHiFilt][i])
			absDiffSum += absDiff
			if absDiff > 1 || absDiffSum > 2 {
				doSync = 0
				break
			}
		}

		if doSync == 0 {
			continue
		}

		if sbInfoSrc.TnsActive[tnsHiFilt] != 0 {
			if sbInfoDest.TnsActive[tnsHiFilt] == 0 ||
				tnsInfoDest.NumOfFilters[w] > tnsInfoSrc.NumOfFilters[w] {
				sbInfoDest.TnsActive[tnsHiFilt] = 1
				tnsInfoDest.NumOfFilters[w] = 1
			}
			tnsDataDest.FiltersMerged = tnsDataSrc.FiltersMerged
			tnsInfoDest.Order[w][tnsHiFilt] = tnsInfoSrc.Order[w][tnsHiFilt]
			tnsInfoDest.Length[w][tnsHiFilt] = tnsInfoSrc.Length[w][tnsHiFilt]
			tnsInfoDest.Direction[w][tnsHiFilt] = tnsInfoSrc.Direction[w][tnsHiFilt]
			tnsInfoDest.CoefCompress[w][tnsHiFilt] = tnsInfoSrc.CoefCompress[w][tnsHiFilt]
			for i := 0; i < tC.MaxOrder; i++ {
				tnsInfoDest.Coef[w][tnsHiFilt][i] = tnsInfoSrc.Coef[w][tnsHiFilt][i]
			}
		} else {
			sbInfoDest.TnsActive[tnsHiFilt] = 0
			tnsInfoDest.NumOfFilters[w] = 0
		}
	}
}

func tnsSubblockInfo(data *TNSData, blockType int, window int) *TNSSubblockInfo {
	if blockType == ShortWindow {
		return &data.DataRaw.Short.SubBlockInfo[window]
	}
	return &data.DataRaw.Long.SubBlockInfo
}

func checkTnsSyncInputs(
	tnsDataDest *TNSData,
	tnsDataSrc *TNSData,
	tnsInfoDest *TNSInfo,
	tnsInfoSrc *TNSInfo,
	blockTypeDest int,
	blockTypeSrc int,
	tC *TNSConfig,
) {
	if tnsDataDest == nil {
		panic("fdkaac: nil destination tns data")
	}
	if tnsDataSrc == nil {
		panic("fdkaac: nil source tns data")
	}
	if tnsInfoDest == nil {
		panic("fdkaac: nil destination tns info")
	}
	if tnsInfoSrc == nil {
		panic("fdkaac: nil source tns info")
	}
	if !validSFBWindowSequence(blockTypeDest) || !validSFBWindowSequence(blockTypeSrc) {
		panic("fdkaac: invalid tns sync block type")
	}
	if tC == nil {
		panic("fdkaac: nil tns sync configuration")
	}
	if tC.MaxOrder < 0 || tC.MaxOrder > tnsMaxOrder {
		panic("fdkaac: invalid tns sync order")
	}
}
