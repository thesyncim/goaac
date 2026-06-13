package fdkaac

const pcmQuantThrScale = 16

func FDKaacEncInitPreEchoControl(pbThresholdNm1 []FixpDBL, calcPreEcho *int, numPb int, sfbPcmQuantThreshold []FixpDBL, mdctScalenm1 *int) {
	checkPreEchoVectors(pbThresholdNm1, sfbPcmQuantThreshold, numPb)
	if calcPreEcho == nil || mdctScalenm1 == nil {
		panic("fdkaac: nil pre-echo state")
	}

	*mdctScalenm1 = pcmQuantThrScale >> 1
	copy(pbThresholdNm1[:numPb], sfbPcmQuantThreshold[:numPb])
	*calcPreEcho = 1
}

func FDKaacEncPreEchoControl(pbThresholdNm1 []FixpDBL, calcPreEcho int, numPb int, maxAllowedIncreaseFactor int, minRemainingThresholdFactor FixpSGL, pbThreshold []FixpDBL, mdctScale int, mdctScalenm1 *int) {
	checkPreEchoVectors(pbThresholdNm1, pbThreshold, numPb)
	if mdctScalenm1 == nil {
		panic("fdkaac: nil pre-echo scale")
	}

	if calcPreEcho == 0 {
		copy(pbThresholdNm1[:numPb], pbThreshold[:numPb])
		*mdctScalenm1 = mdctScale
		return
	}

	if mdctScale > *mdctScalenm1 {
		scaling := 2 * (mdctScale - *mdctScalenm1)
		checkPreEchoShift(scaling)
		shift := uint(scaling)
		for i := 0; i < numPb; i++ {
			tmpThreshold1 := FixpDBL(int32(maxAllowedIncreaseFactor) * int32(pbThresholdNm1[i]>>shift))
			tmpThreshold2 := FMultSD(minRemainingThresholdFactor, pbThreshold[i])
			tmp := pbThreshold[i]
			pbThresholdNm1[i] = tmp
			if tmpThreshold1 < tmp {
				tmp = tmpThreshold1
			}
			if tmpThreshold2 > tmp {
				tmp = tmpThreshold2
			}
			pbThreshold[i] = tmp
		}
	} else {
		scaling := 2 * (*mdctScalenm1 - mdctScale)
		checkPreEchoShift(scaling + 1)
		shift := uint(scaling + 1)
		factor := int32(maxAllowedIncreaseFactor >> 1)
		for i := 0; i < numPb; i++ {
			tmpThreshold1 := FixpDBL(factor * int32(pbThresholdNm1[i]))
			tmpThreshold2 := FMultSD(minRemainingThresholdFactor, pbThreshold[i])
			pbThresholdNm1[i] = pbThreshold[i]
			if (pbThreshold[i] >> shift) > tmpThreshold1 {
				pbThreshold[i] = tmpThreshold1 << shift
			}
			if tmpThreshold2 > pbThreshold[i] {
				pbThreshold[i] = tmpThreshold2
			}
		}
	}

	*mdctScalenm1 = mdctScale
}

func checkPreEchoVectors(a []FixpDBL, b []FixpDBL, numPb int) {
	if numPb < 0 {
		panic("fdkaac: negative pre-echo band count")
	}
	if len(a) < numPb || len(b) < numPb {
		panic("fdkaac: short pre-echo vector")
	}
}

func checkPreEchoShift(shift int) {
	if shift < 0 || shift > DfractBits-1 {
		panic("fdkaac: invalid pre-echo shift")
	}
}
