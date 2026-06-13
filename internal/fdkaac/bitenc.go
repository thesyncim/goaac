package fdkaac

const (
	globalGainOffset = 100
	icsReservedBit   = 0
	noiseOffset      = 90
	logNormPCM       = -15

	sectEscValLong  = 31
	sectEscValShort = 7
	codeBookBits    = 4
	sectBitsLong    = 5
	sectBitsShort   = 3

	maxTnsFilters = 2
	tnsMaxOrder   = 12

	elIDBits = 3

	extTypeBits          = 4
	extDataElVersionBits = 4
	extFillNibbleBits    = 4

	fillElCountBits    = 4
	fillElEscCountBits = 8
	maxFillDataBytes   = 269

	dataByteAlignFlag     = 0
	elInstanceTagBits     = 4
	dataByteAlignFlagBits = 1
	dataLenCountBits      = 8
	dataLenEscCountBits   = 8
	maxDSEDataBytes       = 510
	maxElementExtensions  = 1
	maxGlobalExtensions   = 4

	acScalable = 0x000008
	acELD      = 0x000010
	acER       = 0x000040
	acERVCB11  = 0x000001
	acERRVLC   = 0x000002
	acERHCR    = 0x000004

	aotAACLC = 2
	aotSBR   = 5
	aotPS    = 29

	idSCE = 0
	idCPE = 1
	idLFE = 3
	idDSE = 4
	idFIL = 6
	idEnd = 7

	ExtFIL          = 0x00
	ExtFillData     = 0x01
	ExtDataElement  = 0x02
	ExtLDSACData    = 0x09
	ExtDynamicRange = 0x0b
	ExtSBRData      = 0x0d
	ExtSBRDataCRC   = 0x0e

	AACEncOK                      = 0x0000
	AACEncUnknown                 = 0x0002
	AACEncInvalidHandle           = 0x2020
	AACEncInvalidFrameLength      = 0x2080
	AACEncUnsupportedBitrate      = 0x3020
	AACEncUnsupportedBitrateMode  = 0x3028
	AACEncUnsupportedAncBitrate   = 0x3040
	AACEncUnsupportedERFormat     = 0x30a0
	AACEncUnsupportedChannelConf  = 0x30e0
	AACEncUnsupportedSamplingRate = 0x3100
	AACEncNoMemory                = 0x3120
	AACEncQuantError              = 0x4020
	AACEncWrittenBitsError        = 0x4040
	AACEncBitresTooLow            = 0x40a0
	AACEncBitresTooHigh           = 0x40a1
	AACEncInvalidChannelBitrate   = 0x4100
	AACEncUnsupportedAOT          = 0x3000
	AACEncInvalidElementInfoType  = 0x4120
	AACEncWriteScalError          = 0x41e0
	AACEncWriteSecError           = 0x4200
	AACEncWriteSpecError          = 0x4220
)

type rbdID uint8

const (
	rbdElementInstanceTag rbdID = iota
	rbdCommonWindow
	rbdGlobalGain
	rbdICSInfo
	rbdMaxSfb
	rbdMS
	rbdLTPDataPresent
	rbdLTPData
	rbdSectionData
	rbdScaleFactorData
	rbdPulse
	rbdTNSDataPresent
	rbdTNSData
	rbdGainControlDataPresent
	rbdGainControlData
	rbdEsc1HCR
	rbdEsc2RVLC
	rbdSpectralData
	rbdScaleFactorDataUSAC
	rbdCoreMode
	rbdCommonTW
	rbdLPDChannelStream
	rbdTWData
	rbdNoise
	rbdACSpectralData
	rbdFACData
	rbdTNSActive
	rbdTNSDataPresentUSAC
	rbdCommonMaxSfb
	rbdCoupledElements
	rbdGainElementLists
	rbdADTSCRCStartReg1
	rbdADTSCRCStartReg2
	rbdADTSCRCEndReg1
	rbdADTSCRCEndReg2
	rbdDRMCRCStartReg
	rbdDRMCRCEndReg
	rbdNextChannel
	rbdNextChannelLoop
	rbdLinkSequence
	rbdEndOfSequence
)

type bitstreamElementList struct {
	ID   []rbdID
	Next [2]*bitstreamElementList
}

type ElementInfo struct {
	ElType        int
	InstanceTag   int
	NChannelsInEl int
	ChannelIndex  [2]int
	RelativeBits  FixpDBL
}

type ToolsInfo struct {
	MsDigest int
	MsMask   [maxGroupedSFB]int
}

type PsyOutElement struct {
	CommonWindow  int
	ToolsInfo     ToolsInfo
	PsyOutChannel [2]*PsyOutChannel
}

type TNSInfo struct {
	NumOfFilters [transFac]int
	CoefRes      [transFac]int
	Length       [transFac][maxTnsFilters]int
	Order        [transFac][maxTnsFilters]int
	Direction    [transFac][maxTnsFilters]int
	CoefCompress [transFac][maxTnsFilters]int
	Coef         [transFac][maxTnsFilters][tnsMaxOrder]int
}

type QCOutExtension struct {
	Type        int
	PayloadBits int
	Payload     []byte
}

type WriteBitstreamResult struct {
	FrameBits         int
	ChannelElements   int
	ElementExtensions int
	GlobalExtensions  int
}

var (
	elAACSCE = [...]rbdID{
		rbdADTSCRCStartReg1, rbdElementInstanceTag, rbdGlobalGain, rbdICSInfo,
		rbdSectionData, rbdScaleFactorData, rbdPulse, rbdTNSDataPresent,
		rbdTNSData, rbdGainControlDataPresent, rbdSpectralData,
		rbdADTSCRCEndReg1, rbdEndOfSequence,
	}
	elAACCPE = [...]rbdID{
		rbdADTSCRCStartReg1, rbdElementInstanceTag, rbdCommonWindow, rbdLinkSequence,
	}
	elAACCPE0 = [...]rbdID{
		rbdGlobalGain, rbdICSInfo, rbdSectionData, rbdScaleFactorData, rbdPulse,
		rbdTNSDataPresent, rbdTNSData, rbdGainControlDataPresent, rbdSpectralData,
		rbdNextChannel,
		rbdADTSCRCStartReg2, rbdGlobalGain, rbdICSInfo, rbdSectionData,
		rbdScaleFactorData, rbdPulse, rbdTNSDataPresent, rbdTNSData,
		rbdGainControlDataPresent, rbdSpectralData, rbdADTSCRCEndReg1,
		rbdADTSCRCEndReg2, rbdEndOfSequence,
	}
	elAACCPE1 = [...]rbdID{
		rbdICSInfo, rbdMS,
		rbdGlobalGain, rbdSectionData, rbdScaleFactorData, rbdPulse,
		rbdTNSDataPresent, rbdTNSData, rbdGainControlDataPresent, rbdSpectralData,
		rbdNextChannel,
		rbdADTSCRCStartReg2, rbdGlobalGain, rbdSectionData, rbdScaleFactorData,
		rbdPulse, rbdTNSDataPresent, rbdTNSData, rbdGainControlDataPresent,
		rbdSpectralData, rbdADTSCRCEndReg1, rbdADTSCRCEndReg2,
		rbdEndOfSequence,
	}
	nodeAACSCE  = bitstreamElementList{ID: elAACSCE[:]}
	nodeAACCPE0 = bitstreamElementList{ID: elAACCPE0[:]}
	nodeAACCPE1 = bitstreamElementList{ID: elAACCPE1[:]}
	nodeAACCPE  = bitstreamElementList{ID: elAACCPE[:], Next: [2]*bitstreamElementList{&nodeAACCPE0, &nodeAACCPE1}}
)

func FDKaacEncEncodeSpectralData(sfbOffset []int, sectionData *SectionData, quantSpectrum []int16, bitstream *BitStream) int {
	checkEncodeSpectralDataInputs(sfbOffset, sectionData, quantSpectrum, bitstream)

	startBits := BitStreamValidBits(bitstream)
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.CodeBook != codeBookPNSNo {
			tmp := section.SfbStart + section.SfbCnt
			for sfb := section.SfbStart; sfb < tmp; sfb++ {
				FDKaacEncCodeValues(
					quantSpectrum[sfbOffset[sfb]:],
					sfbOffset[sfb+1]-sfbOffset[sfb],
					section.CodeBook,
					bitstream,
				)
			}
		}
	}
	return int(BitStreamValidBits(bitstream) - startBits)
}

func FDKaacEncEncodeGlobalGain(globalGain int, scalefac int, bitstream *BitStream, mdctScale int) int {
	if bitstream != nil {
		WriteBits(bitstream, uint32(globalGain-scalefac+globalGainOffset-4*(logNormPCM-mdctScale)), 8)
	}
	return 8
}

func FDKaacEncEncodeIcsInfo(blockType int, windowShape int, groupingMask int, maxSfbPerGroup int, bitstream *BitStream, syntaxFlags uint32) int {
	checkEncodeIcsInfoInputs(blockType, windowShape, groupingMask, maxSfbPerGroup)

	statBits := 0
	if blockType == ShortWindow {
		statBits = 8 + transFac - 1
	} else if syntaxFlags&acELD != 0 {
		statBits = 6
	} else if syntaxFlags&acScalable == 0 {
		statBits = 11
	} else {
		statBits = 10
	}

	if bitstream != nil {
		if syntaxFlags&acELD == 0 {
			WriteBits(bitstream, icsReservedBit, 1)
			WriteBits(bitstream, uint32(blockType), 2)
			shape := windowShape
			if shape == WindowShapeLOL {
				shape = WindowShapeKBD
			}
			WriteBits(bitstream, uint32(shape), 1)
		}

		switch blockType {
		case LongWindow, StartWindow, StopWindow:
			WriteBits(bitstream, uint32(maxSfbPerGroup), 6)
			if syntaxFlags&(acScalable|acELD) == 0 {
				WriteBits(bitstream, 0, 1)
			}
		case ShortWindow:
			WriteBits(bitstream, uint32(maxSfbPerGroup), 4)
			WriteBits(bitstream, uint32(groupingMask), transFac-1)
		}
	}
	return statBits
}

func FDKaacEncEncodeSectionData(sectionData *SectionData, bitstream *BitStream, useVCB11 bool) int {
	checkEncodeSectionDataInputs(sectionData)
	if bitstream == nil {
		return 0
	}

	sectEscapeVal := 0
	sectLenBits := uint32(0)
	switch sectionData.BlockType {
	case LongWindow, StartWindow, StopWindow:
		sectEscapeVal = sectEscValLong
		sectLenBits = sectBitsLong
	case ShortWindow:
		sectEscapeVal = sectEscValShort
		sectLenBits = sectBitsShort
	}

	startBits := BitStreamValidBits(bitstream)
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		WriteBits(bitstream, uint32(section.CodeBook), codeBookBits)

		sectLen := section.SfbCnt
		for sectLen >= sectEscapeVal {
			WriteBits(bitstream, uint32(sectEscapeVal), sectLenBits)
			sectLen -= sectEscapeVal
		}
		WriteBits(bitstream, uint32(sectLen), sectLenBits)
	}
	return int(BitStreamValidBits(bitstream) - startBits)
}

func FDKaacEncEncodeScaleFactorData(
	maxValueInSfb []uint32,
	sectionData *SectionData,
	scalefac []int,
	bitstream *BitStream,
	noiseNrg []int,
	isScale []int,
	globalGain int,
) int {
	checkEncodeScaleFactorDataInputs(maxValueInSfb, sectionData, scalefac, noiseNrg, isScale)
	if bitstream == nil {
		return 0
	}

	startBits := BitStreamValidBits(bitstream)
	lastValScf := scalefac[sectionData.FirstScf]
	lastValPns := globalGain - scalefac[sectionData.FirstScf] + globalGainOffset - 4*logNormPCM - noiseOffset
	noisePCMFlag := true
	lastValIs := 0

	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.CodeBook == codeBookZeroNo {
			continue
		}
		if section.CodeBook == codeBookISOutOfPhaseNo || section.CodeBook == codeBookISInPhaseNo {
			sfbStart := section.SfbStart
			tmp := sfbStart + section.SfbCnt
			for j := sfbStart; j < tmp; j++ {
				deltaIs := isScale[j] - lastValIs
				lastValIs = isScale[j]
				if FDKaacEncCodeScalefactorDelta(deltaIs, bitstream) != 0 {
					return 1
				}
			}
		} else if section.CodeBook == codeBookPNSNo {
			sfbStart := section.SfbStart
			tmp := sfbStart + section.SfbCnt
			for j := sfbStart; j < tmp; j++ {
				deltaPns := noiseNrg[j] - lastValPns
				lastValPns = noiseNrg[j]
				if noisePCMFlag {
					WriteBits(bitstream, uint32(deltaPns+(1<<(pnsPCMBits-1))), pnsPCMBits)
					noisePCMFlag = false
				} else if FDKaacEncCodeScalefactorDelta(deltaPns, bitstream) != 0 {
					return 1
				}
			}
		} else {
			tmp := section.SfbStart + section.SfbCnt
			for j := section.SfbStart; j < tmp; j++ {
				deltaScf := 0
				if maxValueInSfb[j] != 0 {
					deltaScf = -(scalefac[j] - lastValScf)
					lastValScf = scalefac[j]
				}
				if FDKaacEncCodeScalefactorDelta(deltaScf, bitstream) != 0 {
					return 1
				}
			}
		}
	}
	return int(BitStreamValidBits(bitstream) - startBits)
}

func FDKaacEncEncodeMSInfo(sfbCnt int, grpSfb int, maxSfb int, msDigest int, jsFlags []int, bitstream *BitStream) int {
	checkEncodeMSInfoInputs(sfbCnt, grpSfb, maxSfb, msDigest, jsFlags)

	msBits := 0
	switch msDigest {
	case MsMaskNone:
		if bitstream != nil {
			WriteBits(bitstream, MsMaskNone, 2)
		}
		msBits += 2
	case MsMaskAll:
		if bitstream != nil {
			WriteBits(bitstream, MsMaskAll, 2)
		}
		msBits += 2
	case MsMaskSome:
		if bitstream != nil {
			WriteBits(bitstream, MsMaskSome, 2)
		}
		msBits += 2
		for sfbOff := 0; sfbOff < sfbCnt; sfbOff += grpSfb {
			for sfb := 0; sfb < maxSfb; sfb++ {
				if bitstream != nil {
					if jsFlags[sfbOff+sfb]&1 != 0 {
						WriteBits(bitstream, 1, 1)
					} else {
						WriteBits(bitstream, 0, 1)
					}
				}
				msBits++
			}
		}
	}
	return msBits
}

func FDKaacEncEncodeTnsDataPresent(tnsInfo *TNSInfo, blockType int, bitstream *BitStream) int {
	checkEncodeTnsInfoInputs(tnsInfo, blockType)
	if bitstream != nil && tnsInfo != nil {
		if tnsDataPresent(tnsInfo, blockType) {
			WriteBits(bitstream, 1, 1)
		} else {
			WriteBits(bitstream, 0, 1)
		}
	}
	return 1
}

func FDKaacEncEncodeTnsData(tnsInfo *TNSInfo, blockType int, bitstream *BitStream) int {
	checkEncodeTnsInfoInputs(tnsInfo, blockType)
	if tnsInfo == nil || !tnsDataPresent(tnsInfo, blockType) {
		return 0
	}

	tnsBits := 0
	numOfWindows := tnsWindowCount(blockType)
	numFilterBits := uint32(2)
	lengthBits := uint32(6)
	orderBits := uint32(5)
	if blockType == ShortWindow {
		numFilterBits = 1
		lengthBits = 4
		orderBits = 3
	}

	for i := 0; i < numOfWindows; i++ {
		numFilters := tnsInfo.NumOfFilters[i]
		if bitstream != nil {
			WriteBits(bitstream, uint32(numFilters), numFilterBits)
		}
		tnsBits += int(numFilterBits)
		if numFilters != 0 {
			if bitstream != nil {
				if tnsInfo.CoefRes[i] == 4 {
					WriteBits(bitstream, 1, 1)
				} else {
					WriteBits(bitstream, 0, 1)
				}
			}
			tnsBits++
		}
		for j := 0; j < numFilters; j++ {
			if bitstream != nil {
				WriteBits(bitstream, uint32(tnsInfo.Length[i][j]), lengthBits)
				WriteBits(bitstream, uint32(tnsInfo.Order[i][j]), orderBits)
			}
			tnsBits += int(lengthBits + orderBits)
			if tnsInfo.Order[i][j] != 0 {
				coefBits := tnsCoefBits(tnsInfo, i, j)
				if bitstream != nil {
					WriteBits(bitstream, uint32(tnsInfo.Direction[i][j]), 1)
					WriteBits(bitstream, uint32(tnsInfo.CoefRes[i]-coefBits), 1)
					for k := 0; k < tnsInfo.Order[i][j]; k++ {
						WriteBits(bitstream, uint32(tnsInfo.Coef[i][j][k]&int(BitMask[coefBits])), uint32(coefBits))
					}
				}
				tnsBits += 2 + tnsInfo.Order[i][j]*coefBits
			}
		}
	}
	return tnsBits
}

func FDKaacEncEncodeGainControlData(bitstream *BitStream) int {
	if bitstream != nil {
		WriteBits(bitstream, 0, 1)
	}
	return 1
}

func FDKaacEncEncodePulseData(bitstream *BitStream) int {
	if bitstream != nil {
		WriteBits(bitstream, 0, 1)
	}
	return 1
}

func FDKaacEncWriteExtensionPayload(bitstream *BitStream, extPayloadType int, extPayloadData []byte, extPayloadBits int) int {
	checkWriteExtensionPayloadInputs(extPayloadType, extPayloadData, extPayloadBits)

	extBitsUsed := 0
	if extPayloadBits < extTypeBits {
		return 0
	}

	fillByte := byte(0x00)
	if bitstream != nil {
		WriteBits(bitstream, uint32(extPayloadType), extTypeBits)
	}
	extBitsUsed += extTypeBits

	switch extPayloadType {
	case ExtLDSACData:
		if bitstream != nil {
			WriteBits(bitstream, uint32(extPayloadData[0]), 4)
		}
		extBitsUsed += 4
		writeExtensionPayloadBytes(bitstream, extPayloadData[1:], extPayloadBits)
		extBitsUsed += extPayloadBits
	case ExtDynamicRange, ExtSBRData, ExtSBRDataCRC:
		writeExtensionPayloadBytes(bitstream, extPayloadData, extPayloadBits)
		extBitsUsed += extPayloadBits
	case ExtDataElement:
		dataElementLength := (extPayloadBits + 7) >> 3
		cnt := dataElementLength
		loopCounter := 1
		for dataElementLength >= 255 {
			loopCounter++
			dataElementLength -= 255
		}
		if bitstream != nil {
			WriteBits(bitstream, 0x00, extDataElVersionBits)
			for i := 1; i < loopCounter; i++ {
				WriteBits(bitstream, 255, 8)
			}
			WriteBits(bitstream, uint32(dataElementLength), 8)
			for i := 0; i < cnt; i++ {
				WriteBits(bitstream, uint32(extPayloadData[i]), 8)
			}
		}
		extBitsUsed += extDataElVersionBits + loopCounter*8 + cnt*8
	case ExtFillData:
		fillByte = 0xa5
		fallthrough
	case ExtFIL:
		if bitstream != nil {
			writeBits := extPayloadBits
			WriteBits(bitstream, 0x00, extFillNibbleBits)
			writeBits -= 8
			for writeBits >= 8 {
				WriteBits(bitstream, uint32(fillByte), 8)
				writeBits -= 8
			}
		}
		extBitsUsed += extFillNibbleBits + (extPayloadBits &^ 0x7) - 8
	default:
		if bitstream != nil {
			writeBits := extPayloadBits
			WriteBits(bitstream, 0x00, extFillNibbleBits)
			writeBits -= 8
			for writeBits >= 8 {
				WriteBits(bitstream, uint32(fillByte), 8)
				writeBits -= 8
			}
		}
		extBitsUsed += extFillNibbleBits + (extPayloadBits &^ 0x7) - 8
	}

	return extBitsUsed
}

func FDKaacEncWriteDataStreamElement(bitstream *BitStream, elementInstanceTag int, dataPayloadBytes int, dataBuffer []byte, alignAnchor uint32) int {
	checkWriteDataStreamElementInputs(elementInstanceTag, dataPayloadBytes, dataBuffer)

	dseBitsUsed := 0
	offset := 0
	for dataPayloadBytes > 0 {
		escCount := -1
		cnt := minInt(maxDSEDataBytes, dataPayloadBytes)

		dseBitsUsed += elIDBits + elInstanceTagBits + dataByteAlignFlagBits + dataLenCountBits
		if cnt >= 255 {
			escCount = cnt - 255
			dseBitsUsed += dataLenEscCountBits
		}
		dataPayloadBytes -= cnt
		dseBitsUsed += cnt * 8

		if bitstream != nil {
			WriteBits(bitstream, idDSE, elIDBits)
			WriteBits(bitstream, uint32(elementInstanceTag), elInstanceTagBits)
			WriteBits(bitstream, dataByteAlignFlag, dataByteAlignFlagBits)
			if escCount >= 0 {
				WriteBits(bitstream, 255, dataLenCountBits)
				WriteBits(bitstream, uint32(escCount), dataLenEscCountBits)
			} else {
				WriteBits(bitstream, uint32(cnt), dataLenCountBits)
			}
			for i := 0; i < cnt; i++ {
				WriteBits(bitstream, uint32(dataBuffer[offset+i]), 8)
			}
		}
		offset += cnt
	}
	return dseBitsUsed
}

func FDKaacEncWriteExtensionData(bitstream *BitStream, extension *QCOutExtension, elInstanceTag int, alignAnchor uint32, syntaxFlags uint32, aot int, epConfig int8) int {
	checkWriteExtensionDataInputs(extension, elInstanceTag)

	payloadBits := extension.PayloadBits
	extBitsUsed := 0

	if syntaxFlags&(acScalable|acER) != 0 {
		if syntaxFlags&acELD != 0 && (extension.Type == ExtSBRData || extension.Type == ExtSBRDataCRC) {
			writeExtensionPayloadBytes(bitstream, extension.Payload, payloadBits)
			return payloadBits
		}
		return FDKaacEncWriteExtensionPayload(bitstream, extension.Type, extension.Payload, payloadBits)
	}

	if extension.Type == ExtDataElement {
		return FDKaacEncWriteDataStreamElement(bitstream, elInstanceTag, extension.PayloadBits>>3, extension.Payload, alignAnchor)
	}

	for payloadBits >= elIDBits+fillElCountBits {
		escCount := -1
		alignBits := 7

		if extension.Type == ExtFillData || extension.Type == ExtFIL {
			payloadBits -= elIDBits + fillElCountBits
			if payloadBits >= 15*8 {
				payloadBits -= fillElEscCountBits
				escCount = 0
			}
			alignBits = 0
		}

		cnt := minInt(maxFillDataBytes, (payloadBits+alignBits)>>3)
		if cnt >= 15 {
			escCount = cnt - 15 + 1
		}

		if bitstream != nil {
			WriteBits(bitstream, idFIL, elIDBits)
			if escCount >= 0 {
				WriteBits(bitstream, 15, fillElCountBits)
				WriteBits(bitstream, uint32(escCount), fillElEscCountBits)
			} else {
				WriteBits(bitstream, uint32(cnt), fillElCountBits)
			}
		}

		extBitsUsed += elIDBits + fillElCountBits
		if escCount >= 0 {
			extBitsUsed += fillElEscCountBits
		}

		cntBits := minInt(cnt*8, payloadBits)
		extBitsUsed += FDKaacEncWriteExtensionPayload(bitstream, extension.Type, extension.Payload, cntBits)
		payloadBits -= cntBits
	}
	return extBitsUsed
}

func writeExtensionPayloadBytes(bitstream *BitStream, payload []byte, payloadBits int) {
	if bitstream == nil {
		return
	}
	writeBits := payloadBits
	i := 0
	for writeBits >= 8 {
		WriteBits(bitstream, uint32(payload[i]), 8)
		i++
		writeBits -= 8
	}
	if writeBits > 0 {
		WriteBits(bitstream, uint32(payload[i]>>(8-writeBits)), uint32(writeBits))
	}
}

func FDKaacEncGetBitstreamElementList(aot int, epConfig int8, nChannels int, layer int, elFlags uint32) *bitstreamElementList {
	if aot != aotAACLC && aot != aotSBR && aot != aotPS {
		return nil
	}
	if epConfig != -1 {
		return nil
	}
	if nChannels == 1 {
		return &nodeAACSCE
	}
	if nChannels == 2 {
		return &nodeAACCPE
	}
	return nil
}

func FDKaacEncChannelElementWrite(
	bitstream *BitStream,
	elInfo *ElementInfo,
	qcOutChannel []*QCOutChannel,
	psyOutElement *PsyOutElement,
	psyOutChannel []*PsyOutChannel,
	syntaxFlags uint32,
	aot int,
	epConfig int8,
	minCnt byte,
) (int, int) {
	checkChannelElementWriteInputs(bitstream, elInfo, qcOutChannel, psyOutElement, psyOutChannel, syntaxFlags, aot, epConfig, minCnt)

	bitDemand := 0
	numberOfChannels := 2
	if elInfo.ElType == idSCE || elInfo.ElType == idLFE {
		numberOfChannels = 1
	} else if elInfo.ElType != idCPE {
		return 0, AACEncInvalidElementInfoType
	}

	list := FDKaacEncGetBitstreamElementList(aot, epConfig, numberOfChannels, 0, 0)
	if list == nil {
		return 0, AACEncUnsupportedAOT
	}

	if syntaxFlags&(acScalable|acER) == 0 {
		if bitstream != nil {
			WriteBits(bitstream, uint32(elInfo.ElType), elIDBits)
		}
		bitDemand += elIDBits
	}

	ch := 0
	decisionBit := 0
	for i := 0; ; i++ {
		if i >= len(list.ID) {
			return bitDemand, AACEncUnknown
		}
		item := list.ID[i]
		if item == rbdEndOfSequence {
			break
		}

		psyCh := psyOutChannel[ch]
		var qcCh *QCOutChannel
		var sectionData *SectionData
		var tnsInfo *TNSInfo
		chGlobalGain := 0
		chBlockType := 0
		chMaxSfbPerGrp := 0
		chSfbPerGrp := 0
		chSfbCnt := 0
		chFirstScf := 0
		if minCnt == 0 {
			if qcOutChannel != nil {
				qcCh = qcOutChannel[ch]
				sectionData = &qcCh.SectionData
				chGlobalGain = qcCh.GlobalGain
				chBlockType = sectionData.BlockType
				chMaxSfbPerGrp = sectionData.MaxSfbPerGroup
				chSfbPerGrp = sectionData.SfbPerGroup
				chSfbCnt = sectionData.SfbCnt
				chFirstScf = qcCh.Scf[sectionData.FirstScf]
			} else {
				chSfbCnt = psyCh.SfbCnt
				chSfbPerGrp = psyCh.SfbPerGroup
				chMaxSfbPerGrp = psyCh.MaxSfbPerGroup
			}
			tnsInfo = &psyCh.TNSInfo
		}
		if qcOutChannel == nil {
			chBlockType = psyCh.LastWindowSequence
		}

		switch item {
		case rbdElementInstanceTag:
			if bitstream != nil {
				WriteBits(bitstream, uint32(elInfo.InstanceTag), elInstanceTagBits)
			}
			bitDemand += elInstanceTagBits
		case rbdCommonWindow:
			decisionBit = psyOutElement.CommonWindow
			if bitstream != nil {
				WriteBits(bitstream, uint32(psyOutElement.CommonWindow), 1)
			}
			bitDemand++
		case rbdICSInfo:
			bitDemand += FDKaacEncEncodeIcsInfo(chBlockType, psyCh.WindowShape, psyCh.GroupingMask, chMaxSfbPerGrp, bitstream, syntaxFlags)
		case rbdLTPDataPresent:
			if bitstream != nil {
				WriteBits(bitstream, 0, 1)
			}
			bitDemand++
		case rbdLTPData:
		case rbdMS:
			msDigest := MsMaskNone
			if minCnt == 0 {
				msDigest = psyOutElement.ToolsInfo.MsDigest
			}
			bitDemand += FDKaacEncEncodeMSInfo(chSfbCnt, chSfbPerGrp, chMaxSfbPerGrp, msDigest, psyOutElement.ToolsInfo.MsMask[:], bitstream)
		case rbdGlobalGain:
			bitDemand += FDKaacEncEncodeGlobalGain(chGlobalGain, chFirstScf, bitstream, psyCh.MdctScale)
		case rbdSectionData:
			if sectionData != nil {
				siBits := FDKaacEncEncodeSectionData(sectionData, bitstream, syntaxFlags&acERRVLC != 0)
				if bitstream != nil && siBits != sectionData.SideInfoBits {
					return bitDemand + siBits, AACEncWriteSecError
				}
				bitDemand += siBits
			}
		case rbdScaleFactorData:
			if sectionData != nil {
				sfDataBits := FDKaacEncEncodeScaleFactorData(qcCh.MaxValueInSfb[:], sectionData, qcCh.Scf[:], bitstream, psyCh.NoiseNrg[:], psyCh.IsScale[:], chGlobalGain)
				if bitstream != nil && sfDataBits != sectionData.ScalefacBits+sectionData.NoiseNrgBits {
					return bitDemand + sfDataBits, AACEncWriteScalError
				}
				bitDemand += sfDataBits
			}
		case rbdEsc2RVLC:
			if syntaxFlags&acERRVLC != 0 {
				return bitDemand, AACEncUnsupportedAOT
			}
		case rbdPulse:
			bitDemand += FDKaacEncEncodePulseData(bitstream)
		case rbdTNSDataPresent:
			bitDemand += FDKaacEncEncodeTnsDataPresent(tnsInfo, chBlockType, bitstream)
		case rbdTNSData:
			bitDemand += FDKaacEncEncodeTnsData(tnsInfo, chBlockType, bitstream)
		case rbdGainControlDataPresent:
			bitDemand += FDKaacEncEncodeGainControlData(bitstream)
		case rbdGainControlData:
		case rbdEsc1HCR:
			if syntaxFlags&acERHCR != 0 {
				return bitDemand, AACEncUnknown
			}
		case rbdSpectralData:
			if bitstream != nil {
				spectralBits := FDKaacEncEncodeSpectralData(psyCh.SfbOffsets[:], sectionData, qcCh.QuantSpec[:], bitstream)
				if spectralBits != sectionData.HuffmanBits {
					return bitDemand + spectralBits, AACEncWriteSpecError
				}
				bitDemand += spectralBits
			}
		case rbdADTSCRCStartReg1, rbdADTSCRCStartReg2, rbdADTSCRCEndReg1, rbdADTSCRCEndReg2, rbdDRMCRCStartReg, rbdDRMCRCEndReg:
		case rbdNextChannel:
			ch = (ch + 1) % numberOfChannels
		case rbdLinkSequence:
			if decisionBit != 0 {
				decisionBit = 1
			}
			if list.Next[decisionBit] == nil {
				return bitDemand, AACEncUnknown
			}
			list = list.Next[decisionBit]
			i = -1
		default:
			return bitDemand, AACEncUnknown
		}
	}

	return bitDemand, AACEncOK
}

func FDKaacEncWriteBitstream(
	bitstream *BitStream,
	cm *ChannelMapping,
	qcOut *QCOut,
	qcElement []*QCOutElement,
	psyOutElement []*PsyOutElement,
	qcKernel *QCKernel,
	aot int,
	syntaxFlags uint32,
	epConfig int8,
) (WriteBitstreamResult, int) {
	checkWriteBitstreamInputs(bitstream, cm, qcOut, qcElement, psyOutElement, qcKernel, aot, syntaxFlags, epConfig)

	result := WriteBitstreamResult{}
	alignAnchor := BitStreamValidBits(bitstream)
	bitMarkUp := int(alignAnchor)
	frameBits := bitMarkUp

	for i := 0; i < cm.NElements; i++ {
		elInfo := cm.ElInfo[i]
		elementUsedBits := 0

		switch elInfo.ElType {
		case idSCE, idCPE, idLFE:
			nChannels := elInfo.NChannelsInEl
			qcChannels := qcElement[i].QCOutChannel[:nChannels]
			psyChannels := psyOutElement[i].PsyOutChannel[:nChannels]
			if _, errCode := FDKaacEncChannelElementWrite(
				bitstream,
				&elInfo,
				qcChannels,
				psyOutElement[i],
				psyChannels,
				syntaxFlags,
				aot,
				epConfig,
				0,
			); errCode != AACEncOK {
				return result, errCode
			}

			if syntaxFlags&acER == 0 {
				for n := 0; n < qcElement[i].NExtensions; n++ {
					FDKaacEncWriteExtensionData(bitstream, &qcElement[i].Extension[n], 0, alignAnchor, syntaxFlags, aot, epConfig)
					result.ElementExtensions++
				}
			}
		default:
			return result, AACEncInvalidElementInfoType
		}

		if elInfo.ElType != idDSE {
			elementUsedBits -= bitMarkUp
			bitMarkUp = int(BitStreamValidBits(bitstream))
			elementUsedBits += bitMarkUp
			frameBits += elementUsedBits
			result.ChannelElements++
		}
	}

	n := qcOut.NExtensions
	qcOut.Extension[n] = QCOutExtension{Type: ExtFillData, PayloadBits: qcOut.TotFillBits}
	qcOut.NExtensions++

	for n = 0; n < qcOut.NExtensions && n < maxGlobalExtensions; n++ {
		FDKaacEncWriteExtensionData(bitstream, &qcOut.Extension[n], 0, alignAnchor, syntaxFlags, aot, epConfig)
		result.GlobalExtensions++
	}

	if syntaxFlags&(acScalable|acER) == 0 {
		WriteBits(bitstream, idEnd, elIDBits)
	}

	if ((BitStreamValidBits(bitstream) - alignAnchor + uint32(qcOut.AlignBits)) & 0x7) != 0 {
		return result, AACEncWrittenBitsError
	}
	FDKaacEncByteAlignment(bitstream, qcOut.AlignBits)

	frameBits -= bitMarkUp
	frameBits += int(BitStreamValidBits(bitstream))
	result.FrameBits = frameBits
	if frameBits != qcOut.TotalBits+qcKernel.GlobHdrBits {
		return result, AACEncWrittenBitsError
	}

	return result, AACEncOK
}

func FDKaacEncByteAlignment(bitstream *BitStream, alignBits int) {
	if alignBits > 0 {
		WriteBits(bitstream, 0, uint32(alignBits))
	}
}

func checkEncodeSpectralDataInputs(sfbOffset []int, sectionData *SectionData, quantSpectrum []int16, bitstream *BitStream) {
	if bitstream == nil {
		panic("fdkaac: nil spectral-data bitstream")
	}
	if sectionData == nil {
		panic("fdkaac: nil spectral-data section data")
	}
	if sectionData.NoOfSections < 0 || sectionData.NoOfSections > maxSections {
		panic("fdkaac: invalid spectral-data section count")
	}
	if len(sfbOffset) < sectionData.SfbCnt+1 {
		panic("fdkaac: short spectral-data offsets")
	}
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.SfbStart < 0 || section.SfbCnt < 0 || section.SfbStart+section.SfbCnt > sectionData.SfbCnt {
			panic("fdkaac: invalid spectral-data section")
		}
		if section.CodeBook == codeBookPNSNo {
			continue
		}
		for sfb := section.SfbStart; sfb < section.SfbStart+section.SfbCnt; sfb++ {
			if sfbOffset[sfb] < 0 || sfbOffset[sfb+1] < sfbOffset[sfb] || len(quantSpectrum) < sfbOffset[sfb+1] {
				panic("fdkaac: malformed spectral-data offsets")
			}
		}
	}
}

func checkEncodeIcsInfoInputs(blockType int, windowShape int, groupingMask int, maxSfbPerGroup int) {
	if blockType != LongWindow && blockType != StartWindow && blockType != ShortWindow && blockType != StopWindow {
		panic("fdkaac: invalid ICS block type")
	}
	if windowShape != WindowShapeSine && windowShape != WindowShapeKBD && windowShape != WindowShapeLOL {
		panic("fdkaac: invalid ICS window shape")
	}
	if maxSfbPerGroup < 0 || maxSfbPerGroup > maxGroupedSFB {
		panic("fdkaac: invalid ICS max SFB")
	}
	if groupingMask < 0 || groupingMask >= (1<<(transFac-1)) {
		panic("fdkaac: invalid ICS grouping mask")
	}
}

func checkEncodeSectionDataInputs(sectionData *SectionData) {
	if sectionData == nil {
		panic("fdkaac: nil section data")
	}
	if sectionData.BlockType != LongWindow && sectionData.BlockType != StartWindow && sectionData.BlockType != ShortWindow && sectionData.BlockType != StopWindow {
		panic("fdkaac: invalid section-data block type")
	}
	if sectionData.NoOfSections < 0 || sectionData.NoOfSections > maxSections {
		panic("fdkaac: invalid section-data count")
	}
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.CodeBook < codeBookZeroNo || section.CodeBook > codeBookISInPhaseNo {
			panic("fdkaac: invalid section-data codebook")
		}
		if section.SfbCnt < 0 {
			panic("fdkaac: invalid section-data length")
		}
	}
}

func checkEncodeMSInfoInputs(sfbCnt int, grpSfb int, maxSfb int, msDigest int, jsFlags []int) {
	if msDigest != MsMaskNone && msDigest != MsMaskSome && msDigest != MsMaskAll {
		panic("fdkaac: invalid MS digest")
	}
	if sfbCnt < 0 || sfbCnt > maxGroupedSFB {
		panic("fdkaac: invalid MS sfb count")
	}
	if msDigest != MsMaskSome {
		return
	}
	if grpSfb <= 0 || sfbCnt%grpSfb != 0 {
		panic("fdkaac: invalid MS group width")
	}
	if maxSfb < 0 || maxSfb > grpSfb {
		panic("fdkaac: invalid MS max sfb")
	}
	if len(jsFlags) < sfbCnt {
		panic("fdkaac: short MS mask")
	}
}

func checkEncodeTnsInfoInputs(tnsInfo *TNSInfo, blockType int) {
	if blockType != LongWindow && blockType != StartWindow && blockType != ShortWindow && blockType != StopWindow {
		panic("fdkaac: invalid TNS block type")
	}
	if tnsInfo == nil {
		return
	}

	numOfWindows := tnsWindowCount(blockType)
	lengthLimit := 1 << 6
	orderLimit := 1 << 5
	if blockType == ShortWindow {
		lengthLimit = 1 << 4
		orderLimit = 1 << 3
	}

	for i := 0; i < numOfWindows; i++ {
		numFilters := tnsInfo.NumOfFilters[i]
		if numFilters < 0 || numFilters > maxTnsFilters {
			panic("fdkaac: invalid TNS filter count")
		}
		if numFilters == 0 {
			continue
		}
		if tnsInfo.CoefRes[i] != 3 && tnsInfo.CoefRes[i] != 4 {
			panic("fdkaac: invalid TNS coefficient resolution")
		}
		for j := 0; j < numFilters; j++ {
			if tnsInfo.Length[i][j] < 0 || tnsInfo.Length[i][j] >= lengthLimit {
				panic("fdkaac: invalid TNS filter length")
			}
			if tnsInfo.Order[i][j] < 0 || tnsInfo.Order[i][j] > tnsMaxOrder || tnsInfo.Order[i][j] >= orderLimit {
				panic("fdkaac: invalid TNS filter order")
			}
			if tnsInfo.Direction[i][j] != 0 && tnsInfo.Direction[i][j] != 1 {
				panic("fdkaac: invalid TNS filter direction")
			}
			coefBits := tnsCoefBits(tnsInfo, i, j)
			coefMin := -(1 << (coefBits - 1))
			coefMax := (1 << (coefBits - 1)) - 1
			for k := 0; k < tnsInfo.Order[i][j]; k++ {
				if tnsInfo.Coef[i][j][k] < coefMin || tnsInfo.Coef[i][j][k] > coefMax {
					panic("fdkaac: invalid TNS coefficient")
				}
			}
		}
	}
}

func tnsWindowCount(blockType int) int {
	if blockType == ShortWindow {
		return transFac
	}
	return 1
}

func tnsDataPresent(tnsInfo *TNSInfo, blockType int) bool {
	numOfWindows := tnsWindowCount(blockType)
	for i := 0; i < numOfWindows; i++ {
		if tnsInfo.NumOfFilters[i] != 0 {
			return true
		}
	}
	return false
}

func tnsCoefBits(tnsInfo *TNSInfo, window int, filter int) int {
	if tnsInfo.CoefRes[window] == 4 {
		for k := 0; k < tnsInfo.Order[window][filter]; k++ {
			if tnsInfo.Coef[window][filter][k] > 3 || tnsInfo.Coef[window][filter][k] < -4 {
				return 4
			}
		}
		return 3
	}
	for k := 0; k < tnsInfo.Order[window][filter]; k++ {
		if tnsInfo.Coef[window][filter][k] > 1 || tnsInfo.Coef[window][filter][k] < -2 {
			return 3
		}
	}
	return 2
}

func checkWriteExtensionPayloadInputs(extPayloadType int, extPayloadData []byte, extPayloadBits int) {
	if extPayloadType < 0 || extPayloadType > 0xf {
		panic("fdkaac: invalid extension payload type")
	}
	if extPayloadBits < 0 {
		panic("fdkaac: invalid extension payload length")
	}
	if extPayloadBits < extTypeBits {
		return
	}

	needBytes := 0
	switch extPayloadType {
	case ExtLDSACData:
		needBytes = 1 + ((extPayloadBits + 7) >> 3)
	case ExtDynamicRange, ExtSBRData, ExtSBRDataCRC:
		needBytes = (extPayloadBits + 7) >> 3
	case ExtDataElement:
		needBytes = (extPayloadBits + 7) >> 3
	}
	if len(extPayloadData) < needBytes {
		panic("fdkaac: short extension payload")
	}
}

func checkWriteDataStreamElementInputs(elementInstanceTag int, dataPayloadBytes int, dataBuffer []byte) {
	if elementInstanceTag < 0 || elementInstanceTag >= (1<<elInstanceTagBits) {
		panic("fdkaac: invalid DSE instance tag")
	}
	if dataPayloadBytes < 0 {
		panic("fdkaac: invalid DSE payload length")
	}
	if len(dataBuffer) < dataPayloadBytes {
		panic("fdkaac: short DSE payload")
	}
}

func checkWriteExtensionDataInputs(extension *QCOutExtension, elInstanceTag int) {
	if extension == nil {
		panic("fdkaac: nil extension payload")
	}
	checkWriteExtensionPayloadInputs(extension.Type, extension.Payload, extension.PayloadBits)
	if extension.Type == ExtDataElement {
		checkWriteDataStreamElementInputs(elInstanceTag, extension.PayloadBits>>3, extension.Payload)
	}
}

func checkChannelElementWriteInputs(
	bitstream *BitStream,
	elInfo *ElementInfo,
	qcOutChannel []*QCOutChannel,
	psyOutElement *PsyOutElement,
	psyOutChannel []*PsyOutChannel,
	syntaxFlags uint32,
	aot int,
	epConfig int8,
	minCnt byte,
) {
	if elInfo == nil {
		panic("fdkaac: nil element info")
	}
	if elInfo.InstanceTag < 0 || elInfo.InstanceTag >= (1<<elInstanceTagBits) {
		panic("fdkaac: invalid element instance tag")
	}
	if elInfo.ElType != idSCE && elInfo.ElType != idCPE && elInfo.ElType != idLFE {
		panic("fdkaac: invalid channel element type")
	}
	if psyOutElement == nil {
		panic("fdkaac: nil psy element")
	}
	if psyOutElement.CommonWindow != 0 && psyOutElement.CommonWindow != 1 {
		panic("fdkaac: invalid common-window flag")
	}
	if minCnt > 1 {
		panic("fdkaac: invalid minimum channel-element count")
	}
	if minCnt != 0 && bitstream != nil {
		panic("fdkaac: minimum channel-element count cannot write bits")
	}
	if syntaxFlags&(acScalable|acER) != 0 {
		panic("fdkaac: ER/scalable channel-element writer is not ported")
	}
	if FDKaacEncGetBitstreamElementList(aot, epConfig, channelElementCount(elInfo.ElType), 0, 0) == nil {
		panic("fdkaac: unsupported channel-element sequence")
	}

	nChannels := channelElementCount(elInfo.ElType)
	if len(psyOutChannel) < nChannels {
		panic("fdkaac: short channel-element inputs")
	}
	if qcOutChannel == nil {
		if bitstream != nil {
			panic("fdkaac: nil channel-element qc output")
		}
	} else if len(qcOutChannel) < nChannels {
		panic("fdkaac: short channel-element inputs")
	}
	for ch := 0; ch < nChannels; ch++ {
		if psyOutChannel[ch] == nil {
			panic("fdkaac: nil channel-element psy output")
		}
		if minCnt == 0 || qcOutChannel != nil {
			if psyOutChannel[ch].SfbCnt <= 0 || psyOutChannel[ch].SfbCnt > maxGroupedSFB {
				panic("fdkaac: invalid channel-element sfb count")
			}
			if psyOutChannel[ch].SfbPerGroup <= 0 || psyOutChannel[ch].SfbCnt%psyOutChannel[ch].SfbPerGroup != 0 {
				panic("fdkaac: invalid channel-element sfb group")
			}
			if psyOutChannel[ch].MaxSfbPerGroup < 0 || psyOutChannel[ch].MaxSfbPerGroup > psyOutChannel[ch].SfbPerGroup {
				panic("fdkaac: invalid channel-element max sfb")
			}
		}
		if qcOutChannel == nil {
			if !validPEWindowSequence(psyOutChannel[ch].LastWindowSequence) {
				panic("fdkaac: invalid channel-element block type")
			}
			continue
		}
		if qcOutChannel[ch] == nil {
			panic("fdkaac: nil channel-element qc output")
		}
		sectionData := &qcOutChannel[ch].SectionData
		if sectionData.FirstScf < 0 || sectionData.FirstScf >= len(qcOutChannel[ch].Scf) {
			panic("fdkaac: invalid channel-element first scalefactor")
		}
		if sectionData.SfbCnt <= 0 || sectionData.SfbCnt > maxGroupedSFB {
			panic("fdkaac: invalid channel-element sfb count")
		}
		if sectionData.SfbPerGroup <= 0 || sectionData.SfbCnt%sectionData.SfbPerGroup != 0 {
			panic("fdkaac: invalid channel-element sfb group")
		}
		if sectionData.MaxSfbPerGroup < 0 || sectionData.MaxSfbPerGroup > sectionData.SfbPerGroup {
			panic("fdkaac: invalid channel-element max sfb")
		}
		if psyOutChannel[ch].SfbOffsets[0] < 0 {
			panic("fdkaac: invalid channel-element offsets")
		}
		for sfb := 0; sfb < sectionData.SfbCnt; sfb++ {
			if psyOutChannel[ch].SfbOffsets[sfb+1] < psyOutChannel[ch].SfbOffsets[sfb] {
				panic("fdkaac: invalid channel-element offsets")
			}
		}
		if psyOutChannel[ch].SfbOffsets[sectionData.SfbCnt] > len(qcOutChannel[ch].QuantSpec) {
			panic("fdkaac: short channel-element quant spectrum")
		}
	}
}

func checkWriteBitstreamInputs(
	bitstream *BitStream,
	cm *ChannelMapping,
	qcOut *QCOut,
	qcElement []*QCOutElement,
	psyOutElement []*PsyOutElement,
	qcKernel *QCKernel,
	aot int,
	syntaxFlags uint32,
	epConfig int8,
) {
	if bitstream == nil {
		panic("fdkaac: nil write-bitstream bitstream")
	}
	if cm == nil {
		panic("fdkaac: nil write-bitstream channel mapping")
	}
	if qcOut == nil {
		panic("fdkaac: nil write-bitstream QC output")
	}
	if qcKernel == nil {
		panic("fdkaac: nil write-bitstream kernel")
	}
	if syntaxFlags&(acScalable|acER) != 0 {
		panic("fdkaac: ER/scalable write-bitstream path is not ported")
	}
	if FDKaacEncGetBitstreamElementList(aot, epConfig, 1, 0, 0) == nil {
		panic("fdkaac: unsupported write-bitstream sequence")
	}
	if cm.NElements < 0 || cm.NElements > maxChannelElements ||
		len(qcElement) < cm.NElements || len(psyOutElement) < cm.NElements {
		panic("fdkaac: invalid write-bitstream element count")
	}
	if qcOut.NExtensions < 0 || qcOut.NExtensions >= maxGlobalExtensions {
		panic("fdkaac: invalid write-bitstream global extension count")
	}
	if qcOut.TotFillBits < 0 || qcOut.AlignBits < 0 || qcOut.AlignBits > 7 ||
		qcOut.TotalBits < 0 || qcKernel.GlobHdrBits < 0 {
		panic("fdkaac: invalid write-bitstream bit accounting")
	}
	for n := 0; n < qcOut.NExtensions; n++ {
		checkWriteExtensionDataInputs(&qcOut.Extension[n], 0)
	}
	for i := 0; i < cm.NElements; i++ {
		elInfo := cm.ElInfo[i]
		switch elInfo.ElType {
		case idSCE, idCPE, idLFE:
		default:
			continue
		}
		if elInfo.NChannelsInEl != channelElementCount(elInfo.ElType) {
			panic("fdkaac: invalid write-bitstream channel count")
		}
		if qcElement[i] == nil {
			panic("fdkaac: nil write-bitstream QC element")
		}
		if psyOutElement[i] == nil {
			panic("fdkaac: nil write-bitstream psy element")
		}
		if qcElement[i].NExtensions < 0 || qcElement[i].NExtensions > maxElementExtensions {
			panic("fdkaac: invalid write-bitstream element extension count")
		}
		for n := 0; n < qcElement[i].NExtensions; n++ {
			checkWriteExtensionDataInputs(&qcElement[i].Extension[n], 0)
		}
		for ch := 0; ch < elInfo.NChannelsInEl; ch++ {
			if qcElement[i].QCOutChannel[ch] == nil {
				panic("fdkaac: nil write-bitstream QC channel")
			}
			if psyOutElement[i].PsyOutChannel[ch] == nil {
				panic("fdkaac: nil write-bitstream psy channel")
			}
		}
	}
}

func channelElementCount(elType int) int {
	if elType == idSCE || elType == idLFE {
		return 1
	}
	return 2
}

func checkEncodeScaleFactorDataInputs(maxValueInSfb []uint32, sectionData *SectionData, scalefac []int, noiseNrg []int, isScale []int) {
	if sectionData == nil {
		panic("fdkaac: nil scale-factor section data")
	}
	if sectionData.NoOfSections < 0 || sectionData.NoOfSections > maxSections {
		panic("fdkaac: invalid scale-factor section count")
	}
	if sectionData.FirstScf < 0 || len(scalefac) <= sectionData.FirstScf {
		panic("fdkaac: invalid first scale factor")
	}
	if len(maxValueInSfb) < sectionData.SfbCnt || len(scalefac) < sectionData.SfbCnt || len(noiseNrg) < sectionData.SfbCnt || len(isScale) < sectionData.SfbCnt {
		panic("fdkaac: short scale-factor data")
	}
	for i := 0; i < sectionData.NoOfSections; i++ {
		section := sectionData.Huffsection[i]
		if section.SfbStart < 0 || section.SfbCnt < 0 || section.SfbStart+section.SfbCnt > sectionData.SfbCnt {
			panic("fdkaac: invalid scale-factor section")
		}
	}
}
