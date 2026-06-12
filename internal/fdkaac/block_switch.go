package fdkaac

const (
	LowOVWindow = 4
	WrongWindow = 5

	blockSwitchWindows     = 8
	blockSwitchingIIRLen   = 2
	blockSwitchEnergyShift = 7
	maxNoOfGroups          = 4
)

const (
	blockSwitchHiPassCoeff0         FixpSGL = -16695
	blockSwitchHiPassCoeff1         FixpSGL = 24733
	blockSwitchAccWindowNrgFac      FixpDBL = 644245094
	blockSwitchOneMinusAccNrgFac    FixpSGL = 22938
	blockSwitchInvAttackRatio       FixpSGL = 3277
	blockSwitchMinAttackNrg         FixpDBL = 15625
	blockSwitchSpreadAttackRatioDBL FixpDBL = 10 << (DfractBits - 1 - 4)
)

var blockType2WindowShape = [2][5]int{
	{WindowShapeSine, WindowShapeKBD, WrongWindow, WindowShapeSine, WindowShapeKBD},
	{WindowShapeKBD, WindowShapeSine, WindowShapeSine, WindowShapeKBD, WrongWindow},
}

var suggestedGroupingTable = [8][maxNoOfGroups]int{
	{1, 3, 3, 1},
	{1, 1, 3, 3},
	{2, 1, 3, 2},
	{3, 1, 3, 1},
	{3, 1, 1, 3},
	{3, 2, 1, 2},
	{3, 3, 1, 1},
	{3, 3, 1, 1},
}

var chgWndSq = [2][6]int{
	{LongWindow, StopWindow, WrongWindow, LongWindow, StopWindow, WrongWindow},
	{StartWindow, LowOVWindow, WrongWindow, StartWindow, LowOVWindow, WrongWindow},
}

var chgWndSqLkAhd = [2][2][6]int{
	{
		{LongWindow, ShortWindow, StopWindow, LongWindow, WrongWindow, WrongWindow},
		{StartWindow, ShortWindow, ShortWindow, StartWindow, WrongWindow, WrongWindow},
	},
	{
		{LongWindow, ShortWindow, ShortWindow, LongWindow, WrongWindow, WrongWindow},
		{StartWindow, ShortWindow, ShortWindow, StartWindow, WrongWindow, WrongWindow},
	},
}

var synchronizedBlockTypeTable = [5][5]int{
	{LongWindow, StartWindow, ShortWindow, StopWindow, LowOVWindow},
	{StartWindow, StartWindow, ShortWindow, ShortWindow, LowOVWindow},
	{ShortWindow, ShortWindow, ShortWindow, ShortWindow, WrongWindow},
	{StopWindow, ShortWindow, ShortWindow, StopWindow, LowOVWindow},
	{LowOVWindow, LowOVWindow, WrongWindow, LowOVWindow, LowOVWindow},
}

type BlockSwitchingControl struct {
	LastWindowSequence  int
	WindowShape         int
	LastWindowShape     int
	NBlockSwitchWindows int
	Attack              int
	LastAttack          int
	AttackIndex         int
	LastAttackIndex     int
	AllowShortFrames    int
	AllowLookAhead      int
	NoOfGroups          int
	GroupLen            [maxNoOfGroups]int
	MaxWindowNrg        FixpDBL
	WindowNrg           [2][blockSwitchWindows]FixpDBL
	WindowNrgF          [2][blockSwitchWindows]FixpDBL
	AccWindowNrg        FixpDBL
	IIRStates           [blockSwitchingIIRLen]FixpDBL
}

func FDKaacEncInitBlockSwitching(c *BlockSwitchingControl, isLowDelay bool) {
	if c == nil {
		panic("fdkaac: nil block switching control")
	}
	*c = BlockSwitchingControl{}
	if isLowDelay {
		c.NBlockSwitchWindows = 4
		c.AllowShortFrames = 0
		c.AllowLookAhead = 0
	} else {
		c.NBlockSwitchWindows = 8
		c.AllowShortFrames = 1
		c.AllowLookAhead = 1
	}
	c.NoOfGroups = maxNoOfGroups
	c.LastWindowSequence = LongWindow
	c.WindowShape = blockType2WindowShape[c.AllowShortFrames][c.LastWindowSequence]
}

func FDKaacEncBlockSwitching(c *BlockSwitchingControl, granuleLength int, isLFE bool, pTimeSignal []int16) int {
	if c == nil {
		panic("fdkaac: nil block switching control")
	}
	nBlockSwitchWindows := c.NBlockSwitchWindows
	if nBlockSwitchWindows != 4 && nBlockSwitchWindows != 8 {
		panic("fdkaac: block switching window count not supported")
	}

	if isLFE {
		c.LastWindowSequence = LongWindow
		c.WindowShape = WindowShapeSine
		c.NoOfGroups = 1
		c.GroupLen[0] = 1
		return 0
	}

	c.LastAttack = c.Attack
	c.LastAttackIndex = c.AttackIndex
	c.WindowNrg[0] = c.WindowNrg[1]
	c.WindowNrgF[0] = c.WindowNrgF[1]

	if c.AllowShortFrames != 0 {
		c.GroupLen = [maxNoOfGroups]int{}
		c.NoOfGroups = maxNoOfGroups
		c.GroupLen = suggestedGroupingTable[c.LastAttackIndex]
		if c.Attack != 0 {
			c.MaxWindowNrg = fdkaacEncGetWindowEnergy(c.WindowNrg[0][:], c.LastAttackIndex)
		} else {
			c.MaxWindowNrg = 0
		}
	}

	windowLen := granuleLength >> 3
	if nBlockSwitchWindows == 4 {
		windowLen = granuleLength >> 2
	}
	fdkaacEncCalcWindowEnergy(c, windowLen, pTimeSignal)

	c.Attack = 0
	enMax := FixpDBL(0)
	enM1 := c.WindowNrgF[0][nBlockSwitchWindows-1]

	for i := 0; i < nBlockSwitchWindows; i++ {
		tmp := FMultDiv2SD(blockSwitchOneMinusAccNrgFac, c.AccWindowNrg)
		c.AccWindowNrg = FMultAddDD(tmp, blockSwitchAccWindowNrgFac, enM1)

		if FMultDS(c.WindowNrgF[1][i], blockSwitchInvAttackRatio) > c.AccWindowNrg {
			c.Attack = 1
			c.AttackIndex = i
		}
		enM1 = c.WindowNrgF[1][i]
		if enM1 > enMax {
			enMax = enM1
		}
	}

	if enMax < blockSwitchMinAttackNrg {
		c.Attack = 0
	}

	if c.Attack == 0 && c.LastAttack != 0 {
		if ((c.WindowNrgF[0][nBlockSwitchWindows-1] >> 4) > FMultDD(blockSwitchSpreadAttackRatioDBL, c.WindowNrgF[1][1])) &&
			c.LastAttackIndex == nBlockSwitchWindows-1 {
			c.Attack = 1
			c.AttackIndex = 0
		}
	}

	if c.AllowLookAhead != 0 {
		c.LastWindowSequence = chgWndSqLkAhd[c.LastAttack][c.Attack][c.LastWindowSequence]
	} else {
		c.LastWindowSequence = chgWndSq[c.Attack][c.LastWindowSequence]
	}
	c.WindowShape = blockType2WindowShape[c.AllowShortFrames][c.LastWindowSequence]
	return 0
}

func fdkaacEncGetWindowEnergy(in []FixpDBL, blSwWndIdx int) FixpDBL {
	return in[blSwWndIdx]
}

func fdkaacEncCalcWindowEnergy(c *BlockSwitchingControl, windowLen int, pTimeSignal []int16) {
	if windowLen <= 0 || len(pTimeSignal) < windowLen*c.NBlockSwitchWindows {
		panic("fdkaac: block switching input buffer too small")
	}

	tempIIRState0 := c.IIRStates[0]
	tempIIRState1 := c.IIRStates[1]
	off := 0
	for w := 0; w < c.NBlockSwitchWindows; w++ {
		var tempWindowNrg uint32
		var tempWindowNrgF uint32

		for i := 0; i < windowLen; i++ {
			tempUnfiltered := FixpDBL(pTimeSignal[off]) << (DfractBits - FractBits - 1)
			off++

			t1 := FMultDiv2SD(blockSwitchHiPassCoeff1, tempUnfiltered-tempIIRState0)
			t2 := FMultDiv2SD(blockSwitchHiPassCoeff0, tempIIRState1)
			tempIIRState0 = tempUnfiltered
			tempIIRState1 = (t1 - t2) << 1

			tempWindowNrg += uint32(FixPow2Div2D(tempIIRState0) >> (blockSwitchEnergyShift - 1 - 2))
			tempWindowNrgF += uint32(FixPow2Div2D(tempIIRState1) >> (blockSwitchEnergyShift - 1 - 2))
		}
		c.WindowNrg[1][w] = blockSwitchEnergyToFixpDBL(tempWindowNrg)
		c.WindowNrgF[1][w] = blockSwitchEnergyToFixpDBL(tempWindowNrgF)
	}
	c.IIRStates[0] = tempIIRState0
	c.IIRStates[1] = tempIIRState1
}

func blockSwitchEnergyToFixpDBL(v uint32) FixpDBL {
	if v > uint32(MaxValDBL) {
		return MaxValDBL
	}
	return FixpDBL(v)
}

func FDKaacEncSyncBlockSwitching(left, right *BlockSwitchingControl, nChannels int, commonWindow bool) int {
	if left == nil {
		panic("fdkaac: nil left block switching control")
	}
	patchType := LongWindow
	if nChannels == 2 && commonWindow {
		if right == nil {
			panic("fdkaac: nil right block switching control")
		}
		patchType = synchronizedBlockTypeTable[patchType][left.LastWindowSequence]
		patchType = synchronizedBlockTypeTable[patchType][right.LastWindowSequence]
		if patchType == WrongWindow {
			return -1
		}
		left.LastWindowSequence = patchType
		right.LastWindowSequence = patchType
		left.WindowShape = blockType2WindowShape[left.AllowShortFrames][left.LastWindowSequence]
		right.WindowShape = blockType2WindowShape[left.AllowShortFrames][right.LastWindowSequence]
	}

	if left.AllowShortFrames != 0 {
		if nChannels == 2 {
			if right == nil {
				panic("fdkaac: nil right block switching control")
			}
			if commonWindow {
				windowSequenceLeftOld := left.LastWindowSequence
				windowSequenceRightOld := right.LastWindowSequence

				if patchType != ShortWindow {
					left.NoOfGroups = 1
					right.NoOfGroups = 1
					left.GroupLen[0] = 1
					right.GroupLen[0] = 1
					for i := 1; i < maxNoOfGroups; i++ {
						left.GroupLen[i] = 0
						right.GroupLen[i] = 0
					}
				} else {
					if windowSequenceLeftOld == ShortWindow && windowSequenceRightOld == ShortWindow {
						if left.MaxWindowNrg > right.MaxWindowNrg {
							right.NoOfGroups = left.NoOfGroups
							right.GroupLen = left.GroupLen
						} else {
							left.NoOfGroups = right.NoOfGroups
							left.GroupLen = right.GroupLen
						}
					} else if windowSequenceLeftOld == ShortWindow && windowSequenceRightOld != ShortWindow {
						right.NoOfGroups = left.NoOfGroups
						right.GroupLen = left.GroupLen
					} else if windowSequenceRightOld == ShortWindow && windowSequenceLeftOld != ShortWindow {
						left.NoOfGroups = right.NoOfGroups
						left.GroupLen = right.GroupLen
					} else {
						left.NoOfGroups = 2
						right.NoOfGroups = 2
						left.GroupLen[0] = 4
						right.GroupLen[0] = 4
						left.GroupLen[1] = 4
						right.GroupLen[1] = 4
					}
				}
			} else {
				if left.LastWindowSequence != ShortWindow {
					left.NoOfGroups = 1
					left.GroupLen[0] = 1
					for i := 1; i < maxNoOfGroups; i++ {
						left.GroupLen[i] = 0
					}
				}
				if right.LastWindowSequence != ShortWindow {
					right.NoOfGroups = 1
					right.GroupLen[0] = 1
					for i := 1; i < maxNoOfGroups; i++ {
						right.GroupLen[i] = 0
					}
				}
			}
		} else {
			if left.LastWindowSequence != ShortWindow {
				left.NoOfGroups = 1
				left.GroupLen[0] = 1
				for i := 1; i < maxNoOfGroups; i++ {
					left.GroupLen[i] = 0
				}
			}
		}
	}

	if left.AllowShortFrames == 0 {
		if left.LastWindowSequence != LongWindow && left.LastWindowSequence != StopWindow {
			left.LastWindowSequence = LongWindow
			left.WindowShape = WindowShapeLOL
		}
	}
	if nChannels == 2 {
		if right == nil {
			panic("fdkaac: nil right block switching control")
		}
		if right.AllowShortFrames == 0 {
			if right.LastWindowSequence != LongWindow && right.LastWindowSequence != StopWindow {
				right.LastWindowSequence = LongWindow
				right.WindowShape = WindowShapeLOL
			}
		}
	}

	return 0
}
