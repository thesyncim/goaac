package fdkaac

func FDKaacEncSpreadingMax(pbCnt int, maskLowFactor []FixpDBL, maskHighFactor []FixpDBL, pbSpreadEnergy []FixpDBL) {
	checkSpreadingInputs(pbCnt, maskLowFactor, maskHighFactor, pbSpreadEnergy)

	delay := pbSpreadEnergy[0]
	for i := 1; i < pbCnt; i++ {
		spread := FMultDD(maskHighFactor[i], delay)
		if spread > pbSpreadEnergy[i] {
			delay = spread
		} else {
			delay = pbSpreadEnergy[i]
		}
		pbSpreadEnergy[i] = delay
	}

	delay = pbSpreadEnergy[pbCnt-1]
	for i := pbCnt - 2; i >= 0; i-- {
		spread := FMultDD(maskLowFactor[i], delay)
		if spread > pbSpreadEnergy[i] {
			delay = spread
		} else {
			delay = pbSpreadEnergy[i]
		}
		pbSpreadEnergy[i] = delay
	}
}

func checkSpreadingInputs(pbCnt int, maskLowFactor []FixpDBL, maskHighFactor []FixpDBL, pbSpreadEnergy []FixpDBL) {
	if pbCnt <= 0 {
		panic("fdkaac: empty spreading band count")
	}
	if len(maskLowFactor) < pbCnt || len(maskHighFactor) < pbCnt || len(pbSpreadEnergy) < pbCnt {
		panic("fdkaac: short spreading vector")
	}
}
