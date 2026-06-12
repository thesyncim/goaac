package fdkaac

import "testing"

var mdctSink FixpDBL
var mdctScaleSink int
var mdctReturnSink int

func TestMDCTBlockAACLCVectors(t *testing.T) {
	t.Run("long", func(t *testing.T) {
		var h MDCT
		MDCTInit(&h, nil)
		timeData := make([]int16, 2048)
		fillMDCTInput(timeData, 1)
		mdctData := make([]FixpDBL, 1024)
		mdctDataE := [1]int{}

		processed := MDCTBlock(&h, timeData, 1024, mdctData, 1, 1024, FDKGetWindowSlope(1024, WindowShapeSine), 1024, mdctDataE[:])
		if processed != 1024 {
			t.Fatalf("processed = %d, want 1024", processed)
		}
		if mdctDataE[0] != 12 {
			t.Fatalf("scale = %d, want 12", mdctDataE[0])
		}
		if got, want := hashFixpDBL(mdctData), uint64(0x55e81fd2f12fb66d); got != want {
			t.Fatalf("long MDCT hash = %#016x, want %#016x", got, want)
		}
		if h.prevFR != 1024 || h.prevTL != 1024 || h.prevNR != 0 || !sameWindowTable(h.prevWRS, SineWindow1024[:]) {
			t.Fatalf("unexpected long MDCT state: prevFR=%d prevTL=%d prevNR=%d", h.prevFR, h.prevTL, h.prevNR)
		}
	})

	t.Run("short", func(t *testing.T) {
		var h MDCT
		MDCTInit(&h, nil)
		timeData := make([]int16, 2048)
		fillMDCTInput(timeData, 2)
		mdctData := make([]FixpDBL, 1024)
		mdctDataE := [8]int{}

		processed := MDCTBlock(&h, timeData, 1024, mdctData, 8, 128, FDKGetWindowSlope(128, WindowShapeSine), 128, mdctDataE[:])
		if processed != 1024 {
			t.Fatalf("processed = %d, want 1024", processed)
		}
		for i, got := range mdctDataE {
			if got != 9 {
				t.Fatalf("short scale[%d] = %d, want 9", i, got)
			}
		}
		if got, want := hashFixpDBL(mdctData), uint64(0xb96748afff7f60cb); got != want {
			t.Fatalf("short MDCT hash = %#016x, want %#016x", got, want)
		}
		if h.prevFR != 128 || h.prevTL != 128 || h.prevNR != 0 || !sameWindowTable(h.prevWRS, SineWindow128[:]) {
			t.Fatalf("unexpected short MDCT state: prevFR=%d prevTL=%d prevNR=%d", h.prevFR, h.prevTL, h.prevNR)
		}
	})
}

func TestFDKaacEncTransformRealAACLCTransitions(t *testing.T) {
	var h MDCT
	MDCTInit(&h, nil)
	prevWindowShape := WindowShapeSine
	mdctDataE := 0

	tests := []struct {
		name      string
		blockType int
		shape     int
		seed      int
		hash      uint64
		scale     int
		prevFR    int
		prevTL    int
		prevNR    int
		prevWRS   []FixpSPK
	}{
		{name: "long-kbd", blockType: LongWindow, shape: WindowShapeKBD, seed: 3, hash: 0x5c848161cdc33a5c, scale: 12, prevFR: 1024, prevTL: 1024, prevNR: 0, prevWRS: KBDWindow1024[:]},
		{name: "start-sine", blockType: StartWindow, shape: WindowShapeSine, seed: 4, hash: 0x5d3dc6a72902e52e, scale: 12, prevFR: 128, prevTL: 1024, prevNR: 448, prevWRS: SineWindow128[:]},
		{name: "short-kbd", blockType: ShortWindow, shape: WindowShapeKBD, seed: 5, hash: 0xd3d8aeeb60cb7758, scale: 9, prevFR: 128, prevTL: 128, prevNR: 0, prevWRS: KBDWindow128[:]},
		{name: "stop-sine", blockType: StopWindow, shape: WindowShapeSine, seed: 6, hash: 0x5c45355f92fe721c, scale: 12, prevFR: 1024, prevTL: 1024, prevNR: 0, prevWRS: SineWindow1024[:]},
	}

	for _, tt := range tests {
		timeData := make([]int16, 2048)
		fillMDCTInput(timeData, tt.seed)
		mdctData := make([]FixpDBL, 1024)

		rc := FDKaacEncTransformReal(timeData, mdctData, tt.blockType, tt.shape, &prevWindowShape, &h, 1024, &mdctDataE, 0)
		if rc != 0 {
			t.Fatalf("%s rc = %d, want 0", tt.name, rc)
		}
		if prevWindowShape != tt.shape {
			t.Fatalf("%s prev window shape = %d, want %d", tt.name, prevWindowShape, tt.shape)
		}
		if mdctDataE != tt.scale {
			t.Fatalf("%s scale = %d, want %d", tt.name, mdctDataE, tt.scale)
		}
		if got := hashFixpDBL(mdctData); got != tt.hash {
			t.Fatalf("%s hash = %#016x, want %#016x", tt.name, got, tt.hash)
		}
		if h.prevFR != tt.prevFR || h.prevTL != tt.prevTL || h.prevNR != tt.prevNR || !sameWindowTable(h.prevWRS, tt.prevWRS) {
			t.Fatalf("%s state prevFR=%d prevTL=%d prevNR=%d", tt.name, h.prevFR, h.prevTL, h.prevNR)
		}
	}
}

func TestFDKaacEncTransformRealRejectsUnsupported(t *testing.T) {
	var h MDCT
	MDCTInit(&h, nil)
	var timeData [2048]int16
	var mdctData [1024]FixpDBL
	prevWindowShape := WindowShapeSine
	mdctDataE := 99

	if rc := FDKaacEncTransformReal(timeData[:], mdctData[:], 99, WindowShapeSine, &prevWindowShape, &h, 1024, &mdctDataE, 0); rc != -1 {
		t.Fatalf("bad block type rc = %d, want -1", rc)
	}
	if mdctDataE != 99 {
		t.Fatalf("bad block type changed scale = %d", mdctDataE)
	}

	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{
			name: "mixed-radix long",
			fn: func() {
				var localH MDCT
				MDCTInit(&localH, nil)
				var localTime [1920]int16
				var localOut [960]FixpDBL
				scale := 0
				FDKaacEncTransformReal(localTime[:], localOut[:], LongWindow, WindowShapeSine, nil, &localH, 960, &scale, 0)
			},
		},
		{
			name: "short input",
			fn: func() {
				var localH MDCT
				MDCTInit(&localH, nil)
				MDCTBlock(&localH, timeData[:255], 128, mdctData[:128], 1, 128, FDKGetWindowSlope(128, WindowShapeSine), 128, []int{0})
			},
		},
		{
			name: "short window table",
			fn: func() {
				var localH MDCT
				MDCTInit(&localH, nil)
				MDCTBlock(&localH, timeData[:256], 128, mdctData[:128], 1, 128, FDKGetWindowSlope(128, WindowShapeSine)[:63], 128, []int{0})
			},
		},
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

func TestFDKaacEncTransformRealAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var h MDCT
		MDCTInit(&h, nil)
		var timeData [2048]int16
		var mdctData [1024]FixpDBL
		fillMDCTInput(timeData[:], 7)
		prevWindowShape := WindowShapeSine
		mdctDataE := 0
		rc := FDKaacEncTransformReal(timeData[:], mdctData[:], LongWindow, WindowShapeSine, &prevWindowShape, &h, 1024, &mdctDataE, 0)
		mdctSink = mdctData[0]
		mdctScaleSink = mdctDataE + prevWindowShape
		mdctReturnSink = rc
	})
	if allocs != 0 {
		t.Fatalf("FDKaacEncTransformReal allocations = %v, want 0", allocs)
	}
}

func fillMDCTInput(x []int16, seed int) {
	for i := range x {
		x[i] = int16((((i*73 + seed*37) % 101) - 50) * 256)
	}
}
