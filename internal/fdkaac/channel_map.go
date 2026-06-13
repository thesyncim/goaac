package fdkaac

var channelMapDefault = [...][24]int{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	{0, 1},
	{0, 1},
	{2, 0, 1},
	{2, 0, 1, 3},
	{2, 0, 1, 3, 4},
	{2, 0, 1, 4, 5, 3},
	{2, 6, 7, 0, 1, 4, 5, 3},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	{2, 0, 1, 4, 5, 6, 3},
	{2, 0, 1, 6, 7, 4, 5, 3},
	{2, 6, 7, 0, 1, 10, 11, 4, 5, 8, 3, 9, 14, 12, 13, 18, 19, 15, 16, 17, 20, 21, 22, 23},
	{2, 0, 1, 4, 5, 3, 6, 7},
}

var channelMapDefaultNumChannels = [...]int{
	24, 2, 2, 3, 4, 5, 6, 8, 24, 24, 24, 7, 8, 24, 8,
}

const (
	relBits04  FixpDBL = 0x33333340
	relBits06  FixpDBL = 0x4ccccd00
	relBits03  FixpDBL = 0x26666680
	relBits026 FixpDBL = 0x2147ae00
	relBits037 FixpDBL = 0x2f5c2900
	relBits024 FixpDBL = 0x1eb851e0
	relBits035 FixpDBL = 0x2cccccc0
	relBits006 FixpDBL = 0x07ae1478
	relBits02  FixpDBL = 0x199999a0
	relBits027 FixpDBL = 0x23333340
	relBits005 FixpDBL = 0x06666668
	relBits018 FixpDBL = 0x170a3d80
	relBits004 FixpDBL = 0x051eb850
	relBits055 FixpDBL = 0x46666680
	invInt5    FixpDBL = 0x19999999
)

func FDKaacEncDetermineEncoderMode(mode *ChannelMode, nChannels int) int {
	if mode == nil {
		panic("fdkaac: nil encoder channel mode")
	}
	encMode := ModeInvalid
	if *mode == ModeUnknown {
		for _, cfg := range channelModeConfig {
			if cfg.NChannels == nChannels {
				encMode = cfg.EncMode
				break
			}
		}
		*mode = encMode
	} else if cfg, ok := FDKaacEncGetChannelModeConfiguration(*mode); ok && cfg.NChannels == nChannels {
		encMode = *mode
	}

	if encMode == ModeInvalid {
		return AACEncUnsupportedChannelConf
	}
	return AACEncOK
}

func FDKaacEncInitChannelMapping(mode ChannelMode, channelOrder ChannelOrder, cm *ChannelMapping) int {
	if cm == nil {
		panic("fdkaac: nil channel mapping")
	}
	*cm = ChannelMapping{}

	if cfg, ok := FDKaacEncGetChannelModeConfiguration(mode); ok {
		cm.EncMode = cfg.EncMode
		cm.NChannels = cfg.NChannels
		cm.NChannelsEff = cfg.NChannelsEff
		cm.NElements = cfg.NElements
	}

	mapIdx := fdkaacEncChannelMapIndex(mode)
	var count int
	var instanceCount [idEnd + 1]int

	switch mode {
	case Mode1:
		return fdkaacEncInitElement(&cm.ElInfo[0], idSCE, &count, channelOrder, mapIdx, &instanceCount, MaxValDBL)
	case Mode2:
		return fdkaacEncInitElement(&cm.ElInfo[0], idCPE, &count, channelOrder, mapIdx, &instanceCount, MaxValDBL)
	case Mode1_2:
		if err := fdkaacEncInitElement(&cm.ElInfo[0], idSCE, &count, channelOrder, mapIdx, &instanceCount, relBits04); err != AACEncOK {
			return err
		}
		return fdkaacEncInitElement(&cm.ElInfo[1], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits06)
	case Mode1_2_1:
		if err := fdkaacEncInitElement(&cm.ElInfo[0], idSCE, &count, channelOrder, mapIdx, &instanceCount, relBits03); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[1], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits04); err != AACEncOK {
			return err
		}
		return fdkaacEncInitElement(&cm.ElInfo[2], idSCE, &count, channelOrder, mapIdx, &instanceCount, relBits03)
	case Mode1_2_2:
		if err := fdkaacEncInitElement(&cm.ElInfo[0], idSCE, &count, channelOrder, mapIdx, &instanceCount, relBits026); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[1], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits037); err != AACEncOK {
			return err
		}
		return fdkaacEncInitElement(&cm.ElInfo[2], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits037)
	case Mode1_2_2_1:
		if err := fdkaacEncInitElement(&cm.ElInfo[0], idSCE, &count, channelOrder, mapIdx, &instanceCount, relBits024); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[1], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits035); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[2], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits035); err != AACEncOK {
			return err
		}
		return fdkaacEncInitElement(&cm.ElInfo[3], idLFE, &count, channelOrder, mapIdx, &instanceCount, relBits006)
	case Mode6_1:
		if err := fdkaacEncInitElement(&cm.ElInfo[0], idSCE, &count, channelOrder, mapIdx, &instanceCount, relBits02); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[1], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits027); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[2], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits027); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[3], idSCE, &count, channelOrder, mapIdx, &instanceCount, relBits02); err != AACEncOK {
			return err
		}
		return fdkaacEncInitElement(&cm.ElInfo[4], idLFE, &count, channelOrder, mapIdx, &instanceCount, relBits005)
	case Mode1_2_2_2_1, Mode7_1Back, Mode7_1TopFront, Mode7_1RearSurround, Mode7_1FrontCenter:
		if err := fdkaacEncInitElement(&cm.ElInfo[0], idSCE, &count, channelOrder, mapIdx, &instanceCount, relBits018); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[1], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits026); err != AACEncOK {
			return err
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[2], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits026); err != AACEncOK {
			return err
		}
		if mode != Mode7_1TopFront {
			if err := fdkaacEncInitElement(&cm.ElInfo[3], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits026); err != AACEncOK {
				return err
			}
			return fdkaacEncInitElement(&cm.ElInfo[4], idLFE, &count, channelOrder, mapIdx, &instanceCount, relBits004)
		}
		if err := fdkaacEncInitElement(&cm.ElInfo[3], idLFE, &count, channelOrder, mapIdx, &instanceCount, relBits004); err != AACEncOK {
			return err
		}
		return fdkaacEncInitElement(&cm.ElInfo[4], idCPE, &count, channelOrder, mapIdx, &instanceCount, relBits026)
	default:
		return AACEncUnsupportedChannelConf
	}
}

func fdkaacEncInitElement(
	elInfo *ElementInfo,
	elType int,
	count *int,
	channelOrder ChannelOrder,
	mapIdx int,
	instanceCount *[idEnd + 1]int,
	relativeBits FixpDBL,
) int {
	if elInfo == nil || count == nil || instanceCount == nil {
		panic("fdkaac: nil channel mapping element")
	}

	counter := *count
	elInfo.ElType = elType
	elInfo.RelativeBits = relativeBits

	switch elType {
	case idSCE, idLFE:
		elInfo.NChannelsInEl = 1
		elInfo.ChannelIndex[0] = fdkaacEncChannelMapValue(channelOrder, counter, mapIdx)
		counter++
		elInfo.InstanceTag = instanceCount[elType]
		instanceCount[elType]++
	case idCPE:
		elInfo.NChannelsInEl = 2
		elInfo.ChannelIndex[0] = fdkaacEncChannelMapValue(channelOrder, counter, mapIdx)
		counter++
		elInfo.ChannelIndex[1] = fdkaacEncChannelMapValue(channelOrder, counter, mapIdx)
		counter++
		elInfo.InstanceTag = instanceCount[elType]
		instanceCount[elType]++
	case idDSE:
		elInfo.NChannelsInEl = 0
		elInfo.ChannelIndex[0] = 0
		elInfo.ChannelIndex[1] = 0
		elInfo.InstanceTag = instanceCount[elType]
		instanceCount[elType]++
	default:
		return AACEncInvalidElementInfoType
	}

	*count = counter
	return AACEncOK
}

func fdkaacEncChannelMapIndex(mode ChannelMode) int {
	switch mode {
	case Mode7_1RearSurround:
		return int(Mode7_1Back)
	case Mode7_1FrontCenter:
		return int(Mode1_2_2_2_1)
	default:
		if int(mode) > int(Mode7_1TopFront) {
			return 0
		}
		return int(mode)
	}
}

func fdkaacEncChannelMapValue(channelOrder ChannelOrder, chIdx int, mapIdx int) int {
	if channelOrder == ChannelOrderMPEG {
		return chIdx
	}
	if mapIdx >= 0 && mapIdx < len(channelMapDefault) && chIdx >= 0 && chIdx < channelMapDefaultNumChannels[mapIdx] {
		return channelMapDefault[mapIdx][chIdx]
	}
	return chIdx
}

func FDKaacEncInitElementBits(elementBits []*ElementBits, cm *ChannelMapping, bitrateTot int, averageBitsTot int, maxChannelBits int) int {
	if cm == nil {
		panic("fdkaac: nil element-bits channel mapping")
	}
	if bitrateTot < 0 || averageBitsTot < 0 || maxChannelBits < 0 {
		panic("fdkaac: negative element-bits control")
	}

	nElements := fdkaacEncInitElementBitsCount(cm.EncMode)
	if nElements < 0 {
		return AACEncUnsupportedChannelConf
	}
	fdkaacEncCheckElementBits(elementBits, nElements)

	scBrTot := CountLeadingBits(FixpDBL(bitrateTot))

	switch cm.EncMode {
	case Mode1:
		elementBits[0].ChBitrateEl = bitrateTot
		elementBits[0].MaxBitsEl = maxChannelBits
		elementBits[0].RelativeBitsEl = cm.ElInfo[0].RelativeBits
	case Mode2:
		elementBits[0].ChBitrateEl = bitrateTot >> 1
		elementBits[0].MaxBitsEl = 2 * maxChannelBits
		elementBits[0].RelativeBitsEl = cm.ElInfo[0].RelativeBits
	case Mode1_2:
		elementBits[0].RelativeBitsEl = cm.ElInfo[0].RelativeBits
		elementBits[1].RelativeBitsEl = cm.ElInfo[1].RelativeBits
		sceRate := elementBits[0].RelativeBitsEl
		cpeRate := elementBits[1].RelativeBitsEl
		elementBits[0].ChBitrateEl = fdkaacEncInitElementBitrate(sceRate, bitrateTot, scBrTot, scBrTot)
		elementBits[1].ChBitrateEl = fdkaacEncInitElementBitrate(cpeRate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[0].MaxBitsEl = maxChannelBits
		elementBits[1].MaxBitsEl = 2 * maxChannelBits
	case Mode1_2_1:
		elementBits[0].RelativeBitsEl = cm.ElInfo[0].RelativeBits
		elementBits[1].RelativeBitsEl = cm.ElInfo[1].RelativeBits
		elementBits[2].RelativeBitsEl = cm.ElInfo[2].RelativeBits
		sce1Rate := elementBits[0].RelativeBitsEl
		cpeRate := elementBits[1].RelativeBitsEl
		sce2Rate := elementBits[2].RelativeBitsEl
		elementBits[0].ChBitrateEl = fdkaacEncInitElementBitrate(sce1Rate, bitrateTot, scBrTot, scBrTot)
		elementBits[1].ChBitrateEl = fdkaacEncInitElementBitrate(cpeRate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[2].ChBitrateEl = fdkaacEncInitElementBitrate(sce2Rate, bitrateTot, scBrTot, scBrTot)
		elementBits[0].MaxBitsEl = maxChannelBits
		elementBits[1].MaxBitsEl = 2 * maxChannelBits
		elementBits[2].MaxBitsEl = maxChannelBits
	case Mode1_2_2:
		elementBits[0].RelativeBitsEl = cm.ElInfo[0].RelativeBits
		elementBits[1].RelativeBitsEl = cm.ElInfo[1].RelativeBits
		elementBits[2].RelativeBitsEl = cm.ElInfo[2].RelativeBits
		sceRate := elementBits[0].RelativeBitsEl
		cpe1Rate := elementBits[1].RelativeBitsEl
		cpe2Rate := elementBits[2].RelativeBitsEl
		elementBits[0].ChBitrateEl = fdkaacEncInitElementBitrate(sceRate, bitrateTot, scBrTot, scBrTot)
		elementBits[1].ChBitrateEl = fdkaacEncInitElementBitrate(cpe1Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[2].ChBitrateEl = fdkaacEncInitElementBitrate(cpe2Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[0].MaxBitsEl = maxChannelBits
		elementBits[1].MaxBitsEl = 2 * maxChannelBits
		elementBits[2].MaxBitsEl = 2 * maxChannelBits
	case Mode1_2_2_1:
		elementBits[0].RelativeBitsEl = cm.ElInfo[0].RelativeBits
		elementBits[1].RelativeBitsEl = cm.ElInfo[1].RelativeBits
		elementBits[2].RelativeBitsEl = cm.ElInfo[2].RelativeBits
		elementBits[3].RelativeBitsEl = cm.ElInfo[3].RelativeBits
		sceRate := elementBits[0].RelativeBitsEl
		cpe1Rate := elementBits[1].RelativeBitsEl
		cpe2Rate := elementBits[2].RelativeBitsEl
		lfeRate := elementBits[3].RelativeBitsEl
		maxLfeBits, maxBitsEl := fdkaacEncInitLFEElementBits(lfeRate, maxChannelBits, averageBitsTot, 5, true)
		elementBits[0].ChBitrateEl = fdkaacEncInitElementBitrate(sceRate, bitrateTot, scBrTot, scBrTot)
		elementBits[1].ChBitrateEl = fdkaacEncInitElementBitrate(cpe1Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[2].ChBitrateEl = fdkaacEncInitElementBitrate(cpe2Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[3].ChBitrateEl = fdkaacEncInitElementBitrate(lfeRate, bitrateTot, scBrTot, scBrTot)
		elementBits[0].MaxBitsEl = maxBitsEl
		elementBits[1].MaxBitsEl = 2 * maxBitsEl
		elementBits[2].MaxBitsEl = 2 * maxBitsEl
		elementBits[3].MaxBitsEl = maxLfeBits
	case Mode6_1:
		elementBits[0].RelativeBitsEl = cm.ElInfo[0].RelativeBits
		elementBits[1].RelativeBitsEl = cm.ElInfo[1].RelativeBits
		elementBits[2].RelativeBitsEl = cm.ElInfo[2].RelativeBits
		elementBits[3].RelativeBitsEl = cm.ElInfo[3].RelativeBits
		elementBits[4].RelativeBitsEl = cm.ElInfo[4].RelativeBits
		sceRate := elementBits[0].RelativeBitsEl
		cpe1Rate := elementBits[1].RelativeBitsEl
		cpe2Rate := elementBits[2].RelativeBitsEl
		sce2Rate := elementBits[3].RelativeBitsEl
		lfeRate := elementBits[4].RelativeBitsEl
		maxLfeBits, maxBitsEl := fdkaacEncInitLFEElementBits(lfeRate, maxChannelBits, averageBitsTot, 6, false)
		elementBits[0].ChBitrateEl = fdkaacEncInitElementBitrate(sceRate, bitrateTot, scBrTot, scBrTot)
		elementBits[1].ChBitrateEl = fdkaacEncInitElementBitrate(cpe1Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[2].ChBitrateEl = fdkaacEncInitElementBitrate(cpe2Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[3].ChBitrateEl = fdkaacEncInitElementBitrate(sce2Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[4].ChBitrateEl = fdkaacEncInitElementBitrate(lfeRate, bitrateTot, scBrTot, scBrTot)
		elementBits[0].MaxBitsEl = maxBitsEl
		elementBits[1].MaxBitsEl = 2 * maxBitsEl
		elementBits[2].MaxBitsEl = 2 * maxBitsEl
		elementBits[3].MaxBitsEl = maxBitsEl
		elementBits[4].MaxBitsEl = maxLfeBits
	case Mode1_2_2_2_1, Mode7_1Back, Mode7_1TopFront, Mode7_1RearSurround, Mode7_1FrontCenter:
		cpe3Idx := 3
		lfeIdx := 4
		if cm.EncMode == Mode7_1TopFront {
			cpe3Idx = 4
			lfeIdx = 3
		}
		elementBits[0].RelativeBitsEl = cm.ElInfo[0].RelativeBits
		elementBits[1].RelativeBitsEl = cm.ElInfo[1].RelativeBits
		elementBits[2].RelativeBitsEl = cm.ElInfo[2].RelativeBits
		elementBits[cpe3Idx].RelativeBitsEl = cm.ElInfo[cpe3Idx].RelativeBits
		elementBits[lfeIdx].RelativeBitsEl = cm.ElInfo[lfeIdx].RelativeBits
		sceRate := elementBits[0].RelativeBitsEl
		cpe1Rate := elementBits[1].RelativeBitsEl
		cpe2Rate := elementBits[2].RelativeBitsEl
		cpe3Rate := elementBits[cpe3Idx].RelativeBitsEl
		lfeRate := elementBits[lfeIdx].RelativeBitsEl
		maxLfeBits, maxBitsEl := fdkaacEncInitLFEElementBits(lfeRate, maxChannelBits, averageBitsTot, 7, false)
		elementBits[0].ChBitrateEl = fdkaacEncInitElementBitrate(sceRate, bitrateTot, scBrTot, scBrTot)
		elementBits[1].ChBitrateEl = fdkaacEncInitElementBitrate(cpe1Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[2].ChBitrateEl = fdkaacEncInitElementBitrate(cpe2Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[cpe3Idx].ChBitrateEl = fdkaacEncInitElementBitrate(cpe3Rate, bitrateTot, scBrTot, scBrTot+1)
		elementBits[lfeIdx].ChBitrateEl = fdkaacEncInitElementBitrate(lfeRate, bitrateTot, scBrTot, scBrTot)
		elementBits[0].MaxBitsEl = maxBitsEl
		elementBits[1].MaxBitsEl = 2 * maxBitsEl
		elementBits[2].MaxBitsEl = 2 * maxBitsEl
		elementBits[cpe3Idx].MaxBitsEl = 2 * maxBitsEl
		elementBits[lfeIdx].MaxBitsEl = maxLfeBits
	}

	return AACEncOK
}

func fdkaacEncInitElementBitsCount(mode ChannelMode) int {
	switch mode {
	case Mode1, Mode2:
		return 1
	case Mode1_2:
		return 2
	case Mode1_2_1, Mode1_2_2:
		return 3
	case Mode1_2_2_1:
		return 4
	case Mode6_1, Mode1_2_2_2_1, Mode7_1Back, Mode7_1TopFront, Mode7_1RearSurround, Mode7_1FrontCenter:
		return 5
	default:
		return -1
	}
}

func fdkaacEncCheckElementBits(elementBits []*ElementBits, nElements int) {
	if len(elementBits) < nElements {
		panic("fdkaac: too few element-bits entries")
	}
	for i := 0; i < nElements; i++ {
		if elementBits[i] == nil {
			panic("fdkaac: nil element-bits entry")
		}
	}
}

func fdkaacEncInitElementBitrate(rate FixpDBL, bitrateTot int, scale int, postShift int) int {
	return int(FMultDD(rate, fdkaacEncShiftedFixpDBL(bitrateTot, scale)) >> uint(postShift))
}

func fdkaacEncInitLFEElementBits(lfeRate FixpDBL, maxChannelBits int, averageBitsTot int, effectiveChannels int, reciprocal bool) (int, int) {
	sc := CountLeadingBits(FixpDBL(maxInt(maxChannelBits, averageBitsTot)))
	maxLfeBitsA := int(FMultDD(lfeRate, fdkaacEncShiftedFixpDBL(maxChannelBits, sc))>>uint(sc)) << 1
	maxLfeBitsB := int((FMultDD(relBits055, FMultDD(lfeRate, fdkaacEncShiftedFixpDBL(averageBitsTot, sc))) << 1) >> uint(sc))
	maxLfeBits := maxInt(maxLfeBitsA, maxLfeBitsB)

	maxBitsTot := maxChannelBits * effectiveChannels
	maxBitsEl := maxBitsTot - maxLfeBits
	if reciprocal {
		sc = CountLeadingBits(FixpDBL(maxBitsEl))
		maxBitsEl = int(FMultDD(fdkaacEncShiftedFixpDBL(maxBitsEl, sc), invInt5) >> uint(sc))
	} else {
		maxBitsEl /= effectiveChannels
	}
	return maxLfeBits, maxBitsEl
}

func fdkaacEncShiftedFixpDBL(value int, scale int) FixpDBL {
	if value < 0 || scale < 0 {
		panic("fdkaac: invalid fixed-point shift")
	}
	shifted := value << uint(scale)
	if shifted > int(MaxValDBL) {
		panic("fdkaac: fixed-point shift overflow")
	}
	return FixpDBL(shifted)
}
