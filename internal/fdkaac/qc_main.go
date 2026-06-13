package fdkaac

func FDKaacEncQCMainPrepare(
	elInfo *ElementInfo,
	adjThrStateElement *ATSElement,
	psyOutElement *PsyOutElement,
	qcOutElement *QCOutElement,
	aot int,
	syntaxFlags uint32,
	epConfig int8,
) int {
	checkQCMainPrepareInputs(elInfo, adjThrStateElement, psyOutElement, qcOutElement)

	nChannels := elInfo.NChannelsInEl
	psyOutChannel := psyOutElement.PsyOutChannel[:nChannels]
	qcOutChannel := qcOutElement.QCOutChannel[:nChannels]

	FDKaacEncCalcFormFactor(qcOutChannel, psyOutChannel, nChannels)
	FDKaacEncPECalculation(&qcOutElement.PEData, psyOutChannel, qcOutChannel, &psyOutElement.ToolsInfo, adjThrStateElement, nChannels)

	bitDemand, errCode := FDKaacEncChannelElementWrite(nil, elInfo, nil, psyOutElement, psyOutChannel, syntaxFlags, aot, epConfig, 0)
	qcOutElement.StaticBitsUsed = bitDemand
	return errCode
}

func checkQCMainPrepareInputs(
	elInfo *ElementInfo,
	adjThrStateElement *ATSElement,
	psyOutElement *PsyOutElement,
	qcOutElement *QCOutElement,
) {
	if elInfo == nil {
		panic("fdkaac: nil QC prepare element info")
	}
	if adjThrStateElement == nil {
		panic("fdkaac: nil QC prepare threshold state")
	}
	if psyOutElement == nil {
		panic("fdkaac: nil QC prepare psy element")
	}
	if qcOutElement == nil {
		panic("fdkaac: nil QC prepare output element")
	}
	if elInfo.ElType != idSCE && elInfo.ElType != idCPE && elInfo.ElType != idLFE {
		panic("fdkaac: invalid QC prepare element type")
	}
	if elInfo.NChannelsInEl != channelElementCount(elInfo.ElType) {
		panic("fdkaac: invalid QC prepare channel count")
	}
	for ch := 0; ch < elInfo.NChannelsInEl; ch++ {
		if psyOutElement.PsyOutChannel[ch] == nil {
			panic("fdkaac: nil QC prepare psy channel")
		}
		if qcOutElement.QCOutChannel[ch] == nil {
			panic("fdkaac: nil QC prepare output channel")
		}
	}
}
