package fdkaac

const thrShiftBits = 4

type PsyThresholdConfig struct {
	ClipEnergy                  FixpDBL
	SfbPcmQuantThreshold        [maxGroupedSFB]FixpDBL
	MaxAllowedIncreaseFactor    int
	MinRemainingThresholdFactor FixpSGL
	SfbMaskLowFactor            [maxGroupedSFB]FixpDBL
	SfbMaskHighFactor           [maxGroupedSFB]FixpDBL
	SfbMaskLowFactorSprEn       [maxGroupedSFB]FixpDBL
	SfbMaskHighFactorSprEn      [maxGroupedSFB]FixpDBL
}

type PsyThresholdState struct {
	SfbThresholdNm1 [maxGroupedSFB]FixpDBL
	CalcPreEcho     int
	MdctScalenm1    int
}

func FDKaacEncAdvanceThresholds(
	sfbThreshold []FixpDBL,
	sfbEnergy []FixpDBL,
	sfbSpreadEnergy []FixpDBL,
	sfbActive int,
	nWindows int,
	maxSfb int,
	mdctScale int,
	isShortWindow bool,
	isLFE bool,
	lastWindowSequence int,
	conf *PsyThresholdConfig,
	state *PsyThresholdState,
) {
	checkAdvanceThresholdInputs(sfbThreshold, sfbEnergy, sfbSpreadEnergy, sfbActive, nWindows, maxSfb, isShortWindow, isLFE, lastWindowSequence, conf, state)

	energyShift := mdctScale * 2
	clipNrgShift := energyShift - thrShiftBits
	headroom := 0
	if isShortWindow {
		headroom = 6
	}

	var clipEnergy FixpDBL
	switch {
	case clipNrgShift >= 0:
		clipEnergy = conf.ClipEnergy >> uint(clipNrgShift)
	case clipNrgShift >= -headroom:
		clipEnergy = conf.ClipEnergy << uint(-clipNrgShift)
	default:
		clipEnergy = MaxValDBL
	}

	for w := 0; w < nWindows; w++ {
		base := w * maxSfb
		threshold := sfbThreshold[base:]

		for i := 0; i < sfbActive; i++ {
			if threshold[i] > clipEnergy {
				threshold[i] = clipEnergy
			}
		}

		FDKaacEncSpreadingMax(sfbActive, conf.SfbMaskLowFactor[:], conf.SfbMaskHighFactor[:], threshold)

		energyShift += pcmQuantThrScale
		if energyShift >= 0 {
			energyShift = minInt(DfractBits-1, energyShift)
			shift := uint(energyShift)
			for i := 0; i < sfbActive; i++ {
				threshold[i] >>= thrShiftBits
				pcmThreshold := conf.SfbPcmQuantThreshold[i] >> shift
				if pcmThreshold > threshold[i] {
					threshold[i] = pcmThreshold
				}
			}
		} else {
			energyShift = minInt(DfractBits-1, -energyShift)
			shift := uint(energyShift)
			for i := 0; i < sfbActive; i++ {
				threshold[i] >>= thrShiftBits
				pcmThreshold := conf.SfbPcmQuantThreshold[i] << shift
				if pcmThreshold > threshold[i] {
					threshold[i] = pcmThreshold
				}
			}
		}

		if !isLFE {
			if lastWindowSequence == StopWindow {
				for i := 0; i < sfbActive; i++ {
					state.SfbThresholdNm1[i] = MaxValDBL
				}
				state.MdctScalenm1 = 0
				state.CalcPreEcho = 0
			}

			FDKaacEncPreEchoControl(
				state.SfbThresholdNm1[:],
				state.CalcPreEcho,
				sfbActive,
				conf.MaxAllowedIncreaseFactor,
				conf.MinRemainingThresholdFactor,
				threshold,
				mdctScale,
				&state.MdctScalenm1,
			)
			state.CalcPreEcho = 1

			if lastWindowSequence == StartWindow {
				for i := 0; i < sfbActive; i++ {
					state.SfbThresholdNm1[i] = MaxValDBL
				}
				state.MdctScalenm1 = 0
				state.CalcPreEcho = 0
			}
		}

		spreadEnergy := sfbSpreadEnergy[base:]
		copy(spreadEnergy[:sfbActive], sfbEnergy[base:base+sfbActive])
		FDKaacEncSpreadingMax(sfbActive, conf.SfbMaskLowFactorSprEn[:], conf.SfbMaskHighFactorSprEn[:], spreadEnergy)
	}
}

func checkAdvanceThresholdInputs(
	sfbThreshold []FixpDBL,
	sfbEnergy []FixpDBL,
	sfbSpreadEnergy []FixpDBL,
	sfbActive int,
	nWindows int,
	maxSfb int,
	isShortWindow bool,
	isLFE bool,
	lastWindowSequence int,
	conf *PsyThresholdConfig,
	state *PsyThresholdState,
) {
	if conf == nil {
		panic("fdkaac: nil threshold config")
	}
	if !isLFE && state == nil {
		panic("fdkaac: nil threshold state")
	}
	if sfbActive <= 0 || sfbActive > maxGroupedSFB {
		panic("fdkaac: invalid threshold band count")
	}
	if maxSfb < sfbActive || maxSfb > maxGroupedSFB {
		panic("fdkaac: invalid threshold stride")
	}
	if isShortWindow {
		if nWindows <= 0 || nWindows > transFac {
			panic("fdkaac: invalid short threshold window count")
		}
	} else if nWindows != 1 {
		panic("fdkaac: invalid long threshold window count")
	}
	if lastWindowSequence != LongWindow && lastWindowSequence != StartWindow && lastWindowSequence != ShortWindow && lastWindowSequence != StopWindow && lastWindowSequence != LowOVWindow {
		panic("fdkaac: invalid threshold window sequence")
	}
	need := (nWindows-1)*maxSfb + sfbActive
	if len(sfbThreshold) < need || len(sfbEnergy) < need || len(sfbSpreadEnergy) < need {
		panic("fdkaac: short threshold buffer")
	}
}
