package fdkaac

import "testing"

var windowSink []FixpSPK
var windowPairSink FixpSPK

func TestAACLCWindowROM(t *testing.T) {
	tests := []struct {
		name    string
		table   []FixpSPK
		length  int
		hash    uint64
		samples map[int]FixpSPK
	}{
		{
			name:   "SineWindow1024",
			table:  SineWindow1024[:],
			length: 512,
			hash:   0xc9b16f1e391b8921,
			samples: map[int]FixpSPK{
				0: {Re: 32767, Im: 25}, 1: {Re: 32767, Im: 75}, 127: {Re: 32143, Im: 6368}, 511: {Re: 23188, Im: 23153},
			},
		},
		{
			name:   "KBDWindow1024",
			table:  KBDWindow1024[:],
			length: 512,
			hash:   0x15136cb129c6bf3b,
			samples: map[int]FixpSPK{
				0: {Re: 32767, Im: 10}, 1: {Re: 32767, Im: 14}, 127: {Re: 32743, Im: 1277}, 511: {Re: 23203, Im: 23138},
			},
		},
		{
			name:   "SineWindow960",
			table:  SineWindow960[:],
			length: 480,
			hash:   0x77b934efb4acd28c,
			samples: map[int]FixpSPK{
				0: {Re: 32767, Im: 27}, 1: {Re: 32767, Im: 80}, 127: {Re: 32058, Im: 6787}, 479: {Re: 23189, Im: 23152},
			},
		},
		{
			name:   "KBDWindow960",
			table:  KBDWindow960[:],
			length: 480,
			hash:   0x738d30349e4784d5,
			samples: map[int]FixpSPK{
				0: {Re: 32767, Im: 10}, 1: {Re: 32767, Im: 15}, 127: {Re: 32736, Im: 1458}, 479: {Re: 23205, Im: 23136},
			},
		},
		{
			name:   "SineWindow128",
			table:  SineWindow128[:],
			length: 64,
			hash:   0x509448fec0489c82,
			samples: map[int]FixpSPK{
				0: {Re: 32767, Im: 201}, 1: {Re: 32762, Im: 603}, 31: {Re: 30350, Im: 12354}, 63: {Re: 23312, Im: 23028},
			},
		},
		{
			name:   "KBDWindow128",
			table:  KBDWindow128[:],
			length: 64,
			hash:   0xab72f091bcd9afd9,
			samples: map[int]FixpSPK{
				0: {Re: 32767, Im: 1}, 1: {Re: 32767, Im: 4}, 31: {Re: 32593, Im: 3378}, 63: {Re: 23484, Im: 22853},
			},
		},
		{
			name:   "SineWindow120",
			table:  SineWindow120[:],
			length: 60,
			hash:   0x1e1192232394b2a4,
			samples: map[int]FixpSPK{
				0: {Re: 32767, Im: 214}, 1: {Re: 32762, Im: 643}, 31: {Re: 30022, Im: 13132}, 59: {Re: 23322, Im: 23018},
			},
		},
		{
			name:   "KBDWindow120",
			table:  KBDWindow120[:],
			length: 60,
			hash:   0xd940cb11bf156859,
			samples: map[int]FixpSPK{
				0: {Re: 32767, Im: 1}, 1: {Re: 32767, Im: 4}, 31: {Re: 32504, Im: 4149}, 59: {Re: 23505, Im: 22831},
			},
		},
	}
	for _, tt := range tests {
		if len(tt.table) != tt.length {
			t.Fatalf("%s length = %d, want %d", tt.name, len(tt.table), tt.length)
		}
		for i, want := range tt.samples {
			if got := tt.table[i]; got != want {
				t.Fatalf("%s[%d] = %+v, want %+v", tt.name, i, got, want)
			}
		}
		if got := hashFixpSPK(tt.table); got != tt.hash {
			t.Fatalf("%s hash = %#016x, want %#016x", tt.name, got, tt.hash)
		}
	}
}

func TestFDKGetAACLCWindowSlope(t *testing.T) {
	tests := []struct {
		length int
		shape  int
		want   []FixpSPK
	}{
		{length: 1024, shape: WindowShapeSine, want: SineWindow1024[:]},
		{length: 1024, shape: WindowShapeKBD, want: KBDWindow1024[:]},
		{length: 1024, shape: WindowShapeLOL, want: SineWindow1024[:]},
		{length: 1024, shape: 3, want: KBDWindow1024[:]},
		{length: 960, shape: WindowShapeSine, want: SineWindow960[:]},
		{length: 960, shape: WindowShapeKBD, want: KBDWindow960[:]},
		{length: 128, shape: WindowShapeSine, want: SineWindow128[:]},
		{length: 128, shape: WindowShapeKBD, want: KBDWindow128[:]},
		{length: 120, shape: WindowShapeSine, want: SineWindow120[:]},
		{length: 120, shape: WindowShapeKBD, want: KBDWindow120[:]},
	}
	for _, tt := range tests {
		got := FDKGetWindowSlope(tt.length, tt.shape)
		if !sameWindowTable(got, tt.want) {
			t.Fatalf("FDKGetWindowSlope(%d, %d) returned unexpected table", tt.length, tt.shape)
		}
	}
}

func TestFDKGetAACLCWindowSlopeRejectsUnsupported(t *testing.T) {
	for _, tt := range []struct {
		length int
		shape  int
	}{
		{length: 0, shape: WindowShapeSine},
		{length: -1, shape: WindowShapeSine},
		{length: 64, shape: WindowShapeSine},
		{length: 512, shape: WindowShapeKBD},
		{length: 768, shape: WindowShapeKBD},
		{length: 2048, shape: WindowShapeSine},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("FDKGetWindowSlope(%d, %d) did not panic", tt.length, tt.shape)
				}
			}()
			_ = FDKGetWindowSlope(tt.length, tt.shape)
		}()
	}
}

func TestFDKGetAACLCWindowSlopeAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		windowSink = FDKGetWindowSlope(1024, WindowShapeKBD)
		windowPairSink = windowSink[0]
	})
	if allocs != 0 {
		t.Fatalf("FDKGetWindowSlope allocations = %v, want 0", allocs)
	}
}

func sameWindowTable(a, b []FixpSPK) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	return &a[0] == &b[0]
}
