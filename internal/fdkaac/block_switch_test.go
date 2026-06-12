package fdkaac

import "testing"

var blockSwitchSink int
var blockSwitchStateSink uint64

func TestBlockSwitchingInit(t *testing.T) {
	var lc BlockSwitchingControl
	FDKaacEncInitBlockSwitching(&lc, false)
	if lc.NBlockSwitchWindows != 8 || lc.AllowShortFrames != 1 || lc.AllowLookAhead != 1 {
		t.Fatalf("LC init flags = windows:%d short:%d lookahead:%d", lc.NBlockSwitchWindows, lc.AllowShortFrames, lc.AllowLookAhead)
	}
	if lc.LastWindowSequence != LongWindow || lc.WindowShape != WindowShapeKBD || lc.NoOfGroups != maxNoOfGroups {
		t.Fatalf("LC init state = seq:%d shape:%d groups:%d", lc.LastWindowSequence, lc.WindowShape, lc.NoOfGroups)
	}
	if got, want := hashBlockSwitchState(&lc), uint64(0x2c537714312f6d30); got != want {
		t.Fatalf("LC init state hash = %#016x, want %#016x", got, want)
	}

	var ld BlockSwitchingControl
	FDKaacEncInitBlockSwitching(&ld, true)
	if ld.NBlockSwitchWindows != 4 || ld.AllowShortFrames != 0 || ld.AllowLookAhead != 0 {
		t.Fatalf("LD init flags = windows:%d short:%d lookahead:%d", ld.NBlockSwitchWindows, ld.AllowShortFrames, ld.AllowLookAhead)
	}
	if ld.LastWindowSequence != LongWindow || ld.WindowShape != WindowShapeSine || ld.NoOfGroups != maxNoOfGroups {
		t.Fatalf("LD init state = seq:%d shape:%d groups:%d", ld.LastWindowSequence, ld.WindowShape, ld.NoOfGroups)
	}
	if got, want := hashBlockSwitchState(&ld), uint64(0xaf25a2a7558c6911); got != want {
		t.Fatalf("LD init state hash = %#016x, want %#016x", got, want)
	}
}

func TestBlockSwitchingLCMonoVectors(t *testing.T) {
	var c BlockSwitchingControl
	FDKaacEncInitBlockSwitching(&c, false)

	tests := []struct {
		name            string
		fill            func([]int16)
		seq             int
		shape           int
		attack          int
		lastAttack      int
		attackIndex     int
		lastAttackIndex int
		groups          int
		groupLen        [maxNoOfGroups]int
		maxWindowNrg    FixpDBL
		accWindowNrg    FixpDBL
		iirStates       [blockSwitchingIIRLen]FixpDBL
		nrgHash         uint64
		nrgFHash        uint64
		stateHash       uint64
	}{
		{name: "silence", fill: fillBlockSwitchSilence, seq: LongWindow, shape: WindowShapeKBD, groupLen: [maxNoOfGroups]int{1, 0, 0, 0}, groups: 1, nrgHash: 0x0c8210784d8af5a5, nrgFHash: 0x0c8210784d8af5a5, stateHash: 0x69ae8f00cc61c244},
		{name: "pulse6", fill: fillBlockSwitchPulse6, seq: StartWindow, shape: WindowShapeSine, attack: 1, attackIndex: 6, groupLen: [maxNoOfGroups]int{1, 0, 0, 0}, groups: 1, accWindowNrg: 539037188, iirStates: [blockSwitchingIIRLen]FixpDBL{0, 4}, nrgHash: 0x5bd5922df50dac73, nrgFHash: 0xddaafd43f8f4bb11, stateHash: 0xdd69d1c0fb98f32e},
		{name: "post-pulse silence 1", fill: fillBlockSwitchSilence, seq: ShortWindow, shape: WindowShapeSine, lastAttack: 1, attackIndex: 6, lastAttackIndex: 6, groupLen: [maxNoOfGroups]int{3, 3, 1, 1}, groups: 4, maxWindowNrg: 1800000000, accWindowNrg: 31106984, iirStates: [blockSwitchingIIRLen]FixpDBL{0, 4}, nrgHash: 0x0c8210784d8af5a5, nrgFHash: 0x0c8210784d8af5a5, stateHash: 0xd876bfe2ead5d8c4},
		{name: "post-pulse silence 2", fill: fillBlockSwitchSilence, seq: StopWindow, shape: WindowShapeKBD, attackIndex: 6, lastAttackIndex: 6, groupLen: [maxNoOfGroups]int{1, 0, 0, 0}, groups: 1, accWindowNrg: 1793504, iirStates: [blockSwitchingIIRLen]FixpDBL{0, 4}, nrgHash: 0x0c8210784d8af5a5, nrgFHash: 0x0c8210784d8af5a5, stateHash: 0xee3c4e106c9200a3},
		{name: "pulse7", fill: fillBlockSwitchPulse7, seq: StartWindow, shape: WindowShapeSine, attack: 1, attackIndex: 7, lastAttackIndex: 6, groupLen: [maxNoOfGroups]int{1, 0, 0, 0}, groups: 1, accWindowNrg: 103404, iirStates: [blockSwitchingIIRLen]FixpDBL{-983040000, -983099622}, nrgHash: 0x079b6b74e50830b3, nrgFHash: 0xd57ac348f0ec412b, stateHash: 0xf7b2d7464a0863c0},
		{name: "decay", fill: fillBlockSwitchDecay, seq: ShortWindow, shape: WindowShapeSine, lastAttack: 1, attackIndex: 7, lastAttackIndex: 7, groupLen: [maxNoOfGroups]int{3, 3, 1, 1}, groups: 4, maxWindowNrg: 1800000000, accWindowNrg: 99688506, iirStates: [blockSwitchingIIRLen]FixpDBL{-360448, -422154}, nrgHash: 0xe1ce8816e107921f, nrgFHash: 0x4e627348f6cd5581, stateHash: 0x29f9050daff5e6dd},
	}

	var frame [1024]int16
	for _, tt := range tests {
		tt.fill(frame[:])
		if rc := FDKaacEncBlockSwitching(&c, 1024, false, frame[:]); rc != 0 {
			t.Fatalf("%s block switching rc = %d, want 0", tt.name, rc)
		}
		if rc := FDKaacEncSyncBlockSwitching(&c, nil, 1, true); rc != 0 {
			t.Fatalf("%s sync rc = %d, want 0", tt.name, rc)
		}
		assertBlockSwitchState(t, tt.name, &c, blockSwitchWant{
			seq: tt.seq, shape: tt.shape, attack: tt.attack, lastAttack: tt.lastAttack,
			attackIndex: tt.attackIndex, lastAttackIndex: tt.lastAttackIndex,
			groups: tt.groups, groupLen: tt.groupLen, maxWindowNrg: tt.maxWindowNrg,
			accWindowNrg: tt.accWindowNrg, iirStates: tt.iirStates, nrgHash: tt.nrgHash,
			nrgFHash: tt.nrgFHash, stateHash: tt.stateHash,
		})
	}
}

func TestBlockSwitchingLFEAndLowDelay(t *testing.T) {
	var frame [1024]int16
	fillBlockSwitchPulse6(frame[:])

	var lfe BlockSwitchingControl
	FDKaacEncInitBlockSwitching(&lfe, false)
	if rc := FDKaacEncBlockSwitching(&lfe, 1024, true, frame[:]); rc != 0 {
		t.Fatalf("LFE rc = %d, want 0", rc)
	}
	assertBlockSwitchState(t, "LFE", &lfe, blockSwitchWant{
		seq: LongWindow, shape: WindowShapeSine, groups: 1, groupLen: [maxNoOfGroups]int{1, 0, 0, 0},
		nrgHash: 0x0c8210784d8af5a5, nrgFHash: 0x0c8210784d8af5a5, stateHash: 0x3d266c4acebac485,
	})

	var ld BlockSwitchingControl
	FDKaacEncInitBlockSwitching(&ld, true)
	if rc := FDKaacEncBlockSwitching(&ld, 1024, false, frame[:]); rc != 0 {
		t.Fatalf("LD rc = %d, want 0", rc)
	}
	if rc := FDKaacEncSyncBlockSwitching(&ld, nil, 1, true); rc != 0 {
		t.Fatalf("LD sync rc = %d, want 0", rc)
	}
	assertBlockSwitchState(t, "LD", &ld, blockSwitchWant{
		seq: LongWindow, shape: WindowShapeLOL, attack: 1, attackIndex: 3, groups: 4,
		iirStates: [blockSwitchingIIRLen]FixpDBL{0, 4}, nrgHash: 0x646e4280d3e8e9b3,
		nrgFHash: 0x5aae428df58470cc, stateHash: 0x39c7cdc07a075616,
	})
}

func TestBlockSwitchingStereoSync(t *testing.T) {
	var left, right BlockSwitchingControl
	FDKaacEncInitBlockSwitching(&left, false)
	FDKaacEncInitBlockSwitching(&right, false)
	var frame [1024]int16

	fillBlockSwitchPulse6(frame[:])
	if rc := FDKaacEncBlockSwitching(&left, 1024, false, frame[:]); rc != 0 {
		t.Fatalf("left block rc = %d, want 0", rc)
	}
	fillBlockSwitchSilence(frame[:])
	if rc := FDKaacEncBlockSwitching(&right, 1024, false, frame[:]); rc != 0 {
		t.Fatalf("right block rc = %d, want 0", rc)
	}
	if rc := FDKaacEncSyncBlockSwitching(&left, &right, 2, true); rc != 0 {
		t.Fatalf("sync rc = %d, want 0", rc)
	}
	assertBlockSwitchState(t, "left sync", &left, blockSwitchWant{
		seq: StartWindow, shape: WindowShapeSine, attack: 1, attackIndex: 6, groups: 1,
		groupLen: [maxNoOfGroups]int{1, 0, 0, 0}, accWindowNrg: 539037188,
		iirStates: [blockSwitchingIIRLen]FixpDBL{0, 4}, nrgHash: 0x5bd5922df50dac73,
		nrgFHash: 0xddaafd43f8f4bb11, stateHash: 0xdd69d1c0fb98f32e,
	})
	assertBlockSwitchState(t, "right sync", &right, blockSwitchWant{
		seq: StartWindow, shape: WindowShapeSine, groups: 1, groupLen: [maxNoOfGroups]int{1, 0, 0, 0},
		nrgHash: 0x0c8210784d8af5a5, nrgFHash: 0x0c8210784d8af5a5, stateHash: 0xb2eb2ea25385d7b4,
	})
}

func TestBlockSwitchingStereoShortGroupingSync(t *testing.T) {
	var left, right BlockSwitchingControl
	FDKaacEncInitBlockSwitching(&left, false)
	FDKaacEncInitBlockSwitching(&right, false)
	left.LastWindowSequence = ShortWindow
	left.WindowShape = WindowShapeSine
	left.NoOfGroups = maxNoOfGroups
	left.GroupLen = [maxNoOfGroups]int{1, 3, 3, 1}
	left.MaxWindowNrg = 1000
	right.LastWindowSequence = ShortWindow
	right.WindowShape = WindowShapeSine
	right.NoOfGroups = maxNoOfGroups
	right.GroupLen = [maxNoOfGroups]int{3, 1, 1, 3}
	right.MaxWindowNrg = 2000

	if rc := FDKaacEncSyncBlockSwitching(&left, &right, 2, true); rc != 0 {
		t.Fatalf("manual sync rc = %d, want 0", rc)
	}
	if left.GroupLen != right.GroupLen || left.GroupLen != ([maxNoOfGroups]int{3, 1, 1, 3}) {
		t.Fatalf("synced grouping left=%v right=%v, want right winner", left.GroupLen, right.GroupLen)
	}
	if got, want := hashBlockSwitchState(&left), uint64(0x86476f19131735d6); got != want {
		t.Fatalf("left manual sync hash = %#016x, want %#016x", got, want)
	}
	if got, want := hashBlockSwitchState(&right), uint64(0xbaf8b0c63a4879ca); got != want {
		t.Fatalf("right manual sync hash = %#016x, want %#016x", got, want)
	}
}

func TestBlockSwitchingRejectsUnsupported(t *testing.T) {
	var c BlockSwitchingControl
	var frame [1024]int16
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{name: "bad window count", fn: func() {
			FDKaacEncBlockSwitching(&c, 1024, false, frame[:])
		}},
		{name: "short input", fn: func() {
			FDKaacEncInitBlockSwitching(&c, false)
			FDKaacEncBlockSwitching(&c, 1024, false, frame[:1023])
		}},
		{name: "nil right common stereo", fn: func() {
			FDKaacEncInitBlockSwitching(&c, false)
			FDKaacEncSyncBlockSwitching(&c, nil, 2, true)
		}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", tt.name)
				}
			}()
			tt.fn()
		}()
	}
}

func TestBlockSwitchingAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var c BlockSwitchingControl
		var frame [1024]int16
		FDKaacEncInitBlockSwitching(&c, false)
		fillBlockSwitchPulse6(frame[:])
		FDKaacEncBlockSwitching(&c, 1024, false, frame[:])
		FDKaacEncSyncBlockSwitching(&c, nil, 1, true)
		blockSwitchSink = c.LastWindowSequence + c.WindowShape + c.AttackIndex
		blockSwitchStateSink = hashBlockSwitchState(&c)
	})
	if allocs != 0 {
		t.Fatalf("block switching allocations = %v, want 0", allocs)
	}
}

type blockSwitchWant struct {
	seq             int
	shape           int
	attack          int
	lastAttack      int
	attackIndex     int
	lastAttackIndex int
	groups          int
	groupLen        [maxNoOfGroups]int
	maxWindowNrg    FixpDBL
	accWindowNrg    FixpDBL
	iirStates       [blockSwitchingIIRLen]FixpDBL
	nrgHash         uint64
	nrgFHash        uint64
	stateHash       uint64
}

func assertBlockSwitchState(t *testing.T, name string, got *BlockSwitchingControl, want blockSwitchWant) {
	t.Helper()
	if got.LastWindowSequence != want.seq || got.WindowShape != want.shape {
		t.Fatalf("%s seq/shape = %d/%d, want %d/%d", name, got.LastWindowSequence, got.WindowShape, want.seq, want.shape)
	}
	if got.Attack != want.attack || got.LastAttack != want.lastAttack || got.AttackIndex != want.attackIndex || got.LastAttackIndex != want.lastAttackIndex {
		t.Fatalf("%s attack state = attack:%d last:%d index:%d lastIndex:%d", name, got.Attack, got.LastAttack, got.AttackIndex, got.LastAttackIndex)
	}
	if got.NoOfGroups != want.groups || got.GroupLen != want.groupLen {
		t.Fatalf("%s grouping = %d %v, want %d %v", name, got.NoOfGroups, got.GroupLen, want.groups, want.groupLen)
	}
	if got.MaxWindowNrg != want.maxWindowNrg || got.AccWindowNrg != want.accWindowNrg || got.IIRStates != want.iirStates {
		t.Fatalf("%s energy state = max:%d acc:%d iir:%v", name, got.MaxWindowNrg, got.AccWindowNrg, got.IIRStates)
	}
	if h := hashFixpDBL(got.WindowNrg[1][:]); h != want.nrgHash {
		t.Fatalf("%s windowNrg hash = %#016x, want %#016x", name, h, want.nrgHash)
	}
	if h := hashFixpDBL(got.WindowNrgF[1][:]); h != want.nrgFHash {
		t.Fatalf("%s windowNrgF hash = %#016x, want %#016x", name, h, want.nrgFHash)
	}
	if h := hashBlockSwitchState(got); h != want.stateHash {
		t.Fatalf("%s state hash = %#016x, want %#016x", name, h, want.stateHash)
	}
}

func fillBlockSwitchSilence(x []int16) {
	for i := range x {
		x[i] = 0
	}
}

func fillBlockSwitchPulse6(x []int16) {
	fillBlockSwitchSilence(x)
	fillBlockSwitchPulseWindow(x, 6)
}

func fillBlockSwitchPulse7(x []int16) {
	fillBlockSwitchSilence(x)
	fillBlockSwitchPulseWindow(x, 7)
}

func fillBlockSwitchPulseWindow(x []int16, w int) {
	for i := w * 128; i < (w+1)*128 && i < len(x); i++ {
		if i&1 != 0 {
			x[i] = -30000
		} else {
			x[i] = 30000
		}
	}
}

func fillBlockSwitchDecay(x []int16) {
	for i := range x {
		v := int16(((len(x) - i) * 12000) / len(x))
		if i&1 != 0 {
			x[i] = -v
		} else {
			x[i] = v
		}
	}
}

func hashBlockSwitchState(c *BlockSwitchingControl) uint64 {
	h := uint64(14695981039346656037)
	add := func(v int32) {
		u := uint32(v)
		h = fnv64AddByte(h, byte(u))
		h = fnv64AddByte(h, byte(u>>8))
		h = fnv64AddByte(h, byte(u>>16))
		h = fnv64AddByte(h, byte(u>>24))
	}
	add(int32(c.LastWindowSequence))
	add(int32(c.WindowShape))
	add(int32(c.Attack))
	add(int32(c.LastAttack))
	add(int32(c.AttackIndex))
	add(int32(c.LastAttackIndex))
	add(int32(c.NoOfGroups))
	add(int32(c.MaxWindowNrg))
	add(int32(c.AccWindowNrg))
	add(int32(c.IIRStates[0]))
	add(int32(c.IIRStates[1]))
	for _, v := range c.GroupLen {
		add(int32(v))
	}
	for _, v := range c.WindowNrg[1] {
		add(int32(v))
	}
	for _, v := range c.WindowNrgF[1] {
		add(int32(v))
	}
	return h
}
