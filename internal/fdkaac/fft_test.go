package fdkaac

import "testing"

var fftSink FixpDBL
var fftScaleSink int
var romSink FixpSPK

func TestFFT2Vectors(t *testing.T) {
	tests := []struct {
		in, out [4]FixpDBL
	}{
		{
			in:  [4]FixpDBL{0, 0, 0, 0},
			out: [4]FixpDBL{0, 0, 0, 0},
		},
		{
			in:  [4]FixpDBL{0x40000000, 0, 0, 0},
			out: [4]FixpDBL{536870912, 0, 536870912, 0},
		},
		{
			in:  [4]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000},
			out: [4]FixpDBL{167772160, -50331648, 100663296, -83886080},
		},
		{
			in:  [4]FixpDBL{0x3fffffff, -0x40000000, -0x20000000, 0x10000000},
			out: [4]FixpDBL{268435455, -402653184, 805306367, -671088640},
		},
	}
	for _, tt := range tests {
		got := tt.in
		FFT2(got[:])
		if got != tt.out {
			t.Fatalf("FFT2(%v) = %v, want %v", tt.in, got, tt.out)
		}
	}
}

func TestFFT4Vectors(t *testing.T) {
	// Expected values were checked against pinned FDK-AAC v2.0.3 on arm64.
	tests := []struct {
		in, out [8]FixpDBL
	}{
		{
			in:  [8]FixpDBL{0, 0, 0, 0, 0, 0, 0, 0},
			out: [8]FixpDBL{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			in:  [8]FixpDBL{0x40000000, 0, 0, 0, 0, 0, 0, 0},
			out: [8]FixpDBL{536870912, 0, 536870912, 0, 536870912, 0, 536870912, 0},
		},
		{
			in:  [8]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000},
			out: [8]FixpDBL{0, 0, 301989888, -201326592, 0, 0, 234881024, -67108864},
		},
		{
			in:  [8]FixpDBL{0x3fffffff, -0x40000000, -0x20000000, 0x10000000, 0x08000000, -0x04000000, -0x02000000, 0x01000000},
			out: [8]FixpDBL{318767103, -427819008, 595591167, -251658240, 889192447, -713031680, 343932927, -754974720},
		},
	}
	for _, tt := range tests {
		got := tt.in
		FFT4(got[:])
		if got != tt.out {
			t.Fatalf("FFT4(%v) = %v, want %v", tt.in, got, tt.out)
		}
	}
}

func TestFFT8Vectors(t *testing.T) {
	tests := []struct {
		in, out [16]FixpDBL
	}{
		{
			in:  [16]FixpDBL{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			out: [16]FixpDBL{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			in:  [16]FixpDBL{0x40000000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			out: [16]FixpDBL{268435456, 0, 268435456, 0, 268435456, 0, 268435456, 0, 268435456, 0, 268435456, 0, 268435456, 0, 268435456, 0},
		},
		{
			in:  [16]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000, 0x08000000, 0x04000000, -0x08000000, 0x01000000, 0x02000000, -0x01000000, -0x02000000, 0x00800000},
			out: [16]FixpDBL{0, 18874368, 108375168, 5888384, 178257920, -54525952, -50163584, -165866880, 83886080, 6291456, 34231168, 44443264, 140509184, -37748736, 41774976, -85791360},
		},
		{
			in:  [16]FixpDBL{0x3fffffff, -0x40000000, -0x20000000, 0x10000000, 0x08000000, -0x04000000, -0x02000000, 0x01000000, 0x01000000, -0x00800000, -0x00400000, 0x00200000, 0x00100000, -0x00080000, -0x00040000, 0x00020000},
			out: [16]FixpDBL{162725887, -215580672, 209341337, -155443682, 302219263, -126812160, 419189025, -177136806, 450166783, -359301120, 285848677, -443817502, 175407103, -380436480, 142585565, -288955226},
		},
	}
	for _, tt := range tests {
		got := tt.in
		FFT8(got[:])
		if got != tt.out {
			t.Fatalf("FFT8(%v) = %v, want %v", tt.in, got, tt.out)
		}
	}
}

func TestScrambleVectors(t *testing.T) {
	tests := []struct {
		length int
		in     []FixpDBL
		out    []FixpDBL
	}{
		{
			length: 4,
			in:     []FixpDBL{10, 11, 20, 21, 30, 31, 40, 41},
			out:    []FixpDBL{10, 11, 30, 31, 20, 21, 40, 41},
		},
		{
			length: 8,
			in:     []FixpDBL{10, 11, 20, 21, 30, 31, 40, 41, 50, 51, 60, 61, 70, 71, 80, 81},
			out:    []FixpDBL{10, 11, 50, 51, 30, 31, 70, 71, 20, 21, 60, 61, 40, 41, 80, 81},
		},
		{
			length: 16,
			in: []FixpDBL{
				10, 11, 20, 21, 30, 31, 40, 41, 50, 51, 60, 61, 70, 71, 80, 81,
				90, 91, 100, 101, 110, 111, 120, 121, 130, 131, 140, 141, 150, 151, 160, 161,
			},
			out: []FixpDBL{
				10, 11, 90, 91, 50, 51, 130, 131, 30, 31, 110, 111, 70, 71, 150, 151,
				20, 21, 100, 101, 60, 61, 140, 141, 40, 41, 120, 121, 80, 81, 160, 161,
			},
		},
	}
	for _, tt := range tests {
		got := append([]FixpDBL(nil), tt.in...)
		Scramble(got, tt.length)
		if !equalFixpDBL(got, tt.out) {
			t.Fatalf("Scramble length %d = %v, want %v", tt.length, got, tt.out)
		}
	}
}

func TestDITFFTVectors(t *testing.T) {
	// Expected values are source-derived from the pinned FDK-AAC v2.0.3 arm64
	// SINETABLE_16BIT radix-2 implementation.
	trig := []FixpSPK{
		{32767, 0}, {32610, 3212}, {32138, 6393}, {31357, 9512},
		{30274, 12540}, {28899, 15447}, {27246, 18205}, {25330, 20788},
		{23170, 23170}, {20788, 25330}, {18205, 27246}, {15447, 28899},
		{12540, 30274}, {9512, 31357}, {6393, 32138}, {3212, 32610},
	}
	tests := []struct {
		ldn int
		in  []FixpDBL
		out []FixpDBL
	}{
		{
			ldn: 3,
			in:  []FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000, 0x08000000, 0x04000000, -0x08000000, 0x01000000, 0x02000000, -0x01000000, -0x02000000, 0x00800000},
			out: []FixpDBL{0, 18874368, 108375168, 5888384, 178257920, -54525952, -50163584, -165866880, 83886080, 6291456, 34231168, 44443264, 140509184, -37748736, 41774976, -85791360},
		},
		{
			ldn: 3,
			in:  []FixpDBL{0x3fffffff, -0x20000000, -0x10000000, 0x08000000, 0x04000000, -0x02000000, -0x01000000, 0x00800000, 0x00400000, -0x00200000, -0x00100000, 0x00080000, 0x00040000, -0x00020000, -0x00010000, 0x00008000},
			out: []FixpDBL{214745087, -107372544, 239828901, -78027833, 284221439, -63160320, 345165831, -88917097, 357908479, -178954240, 278233177, -222782407, 221061119, -189480960, 206319607, -145046423},
		},
		{
			ldn: 4,
			in: []FixpDBL{
				0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000,
				0x08000000, 0x04000000, -0x08000000, 0x01000000, 0x02000000, -0x01000000, -0x02000000, 0x00800000,
				0x06000000, -0x03000000, 0x01800000, 0x00c00000, -0x06000000, 0x03000000, -0x01800000, -0x00c00000,
				0x03000000, 0x01800000, -0x03000000, 0x00600000, 0x00c00000, -0x00600000, -0x00c00000, 0x00300000,
			},
			out: []FixpDBL{
				0, 12976128, 25334859, 14184762, 74507928, 4048264, 40593232, 578149,
				122552320, -37486592, 41772517, -55114260, -34487464, -114033480, -16670468, 8582513,
				57671680, 4325376, 6704341, -13495002, 23533928, 30554744, 37204048, 8689851,
				96600064, -25952256, 31045883, -29461580, 28720296, -58981560, 1787748, -17850513,
			},
		},
		{
			ldn: 4,
			in: []FixpDBL{
				0x3fffffff, -0x20000000, -0x10000000, 0x08000000, 0x04000000, -0x02000000, -0x01000000, 0x00800000,
				0x00400000, -0x00200000, -0x00100000, 0x00080000, 0x00040000, -0x00020000, -0x00010000, 0x00008000,
				-0x08000000, 0x04000000, 0x02000000, -0x01000000, 0x00800000, 0x00400000, -0x00200000, -0x00100000,
				0x00080000, 0x00040000, -0x00020000, -0x00010000, 0x00008000, 0x00004000, -0x00002000, -0x00001000,
			},
			out: []FixpDBL{
				95628287, -46975488, 125375735, -50383154, 105186513, -35880035, 146878994, -36589913,
				122372607, -28126208, 178696162, -39485808, 150486536, -36466168, 203719291, -74758495,
				159380479, -78292480, 179664913, -118596896, 121203501, -99902365, 138662078, -117261341,
				94739967, -82404352, 119063314, -95359038, 90526198, -61714952, 115899059, -71545133,
			},
		},
	}
	for _, tt := range tests {
		got := append([]FixpDBL(nil), tt.in...)
		DITFFT(got, tt.ldn, trig, len(trig))
		if !equalFixpDBL(got, tt.out) {
			t.Fatalf("DITFFT ldn %d = %v, want %v", tt.ldn, got, tt.out)
		}
	}
}

func TestSineTable512ROM(t *testing.T) {
	if len(SineTable512) != 257 {
		t.Fatalf("len(SineTable512) = %d, want 257", len(SineTable512))
	}
	if sineTable512Size != 512 {
		t.Fatalf("sineTable512Size = %d, want 512", sineTable512Size)
	}
	samples := map[int]FixpSPK{
		0:   {Re: 32767, Im: 0},
		1:   {Re: 32767, Im: 101},
		2:   {Re: 32767, Im: 201},
		4:   {Re: 32766, Im: 402},
		8:   {Re: 32758, Im: 804},
		64:  {Re: 32138, Im: 6393},
		128: {Re: 30274, Im: 12540},
		256: {Re: 23170, Im: 23170},
	}
	for i, want := range samples {
		if got := SineTable512[i]; got != want {
			t.Fatalf("SineTable512[%d] = %+v, want %+v", i, got, want)
		}
	}
	const wantHash uint64 = 0x97e5ffd5a49695b4
	if got := hashFixpSPK(SineTable512[:]); got != wantHash {
		t.Fatalf("SineTable512 hash = %#016x, want %#016x", got, wantHash)
	}
}

func TestSineTable1024ROM(t *testing.T) {
	if len(SineTable1024) != 513 {
		t.Fatalf("len(SineTable1024) = %d, want 513", len(SineTable1024))
	}
	if sineTable1024Size != 1024 {
		t.Fatalf("sineTable1024Size = %d, want 1024", sineTable1024Size)
	}
	samples := map[int]FixpSPK{
		0:   {Re: 32767, Im: 0},
		1:   {Re: 32767, Im: 50},
		2:   {Re: 32767, Im: 101},
		4:   {Re: 32767, Im: 201},
		8:   {Re: 32766, Im: 402},
		64:  {Re: 32610, Im: 3212},
		128: {Re: 32138, Im: 6393},
		256: {Re: 30274, Im: 12540},
		512: {Re: 23170, Im: 23170},
	}
	for i, want := range samples {
		if got := SineTable1024[i]; got != want {
			t.Fatalf("SineTable1024[%d] = %+v, want %+v", i, got, want)
		}
	}
	for i, want := range SineTable512 {
		if got := SineTable1024[i*2]; got != want {
			t.Fatalf("SineTable1024[%d] = %+v, want SineTable512[%d] %+v", i*2, got, i, want)
		}
	}
	const wantHash uint64 = 0xe78a87bde9241ac1
	if got := hashFixpSPK(SineTable1024[:]); got != wantHash {
		t.Fatalf("SineTable1024 hash = %#016x, want %#016x", got, wantHash)
	}
}

func TestSineTableROMAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		a := SineTable1024[128]
		b := SineTable512[64]
		romSink = FixpSPK{Re: a.Re + b.Re, Im: a.Im + b.Im}
	})
	if allocs != 0 {
		t.Fatalf("sine table ROM access allocations = %v, want 0", allocs)
	}
}

func TestDITFFT512WithFDKROM(t *testing.T) {
	tests := []struct {
		ldn  int
		hash uint64
	}{
		{ldn: 6, hash: 0x24da1b06d4d1acd5},
		{ldn: 7, hash: 0x6e30afc5327a54af},
		{ldn: 8, hash: 0xf072fffa1018bb22},
		{ldn: 9, hash: 0xec99e0d7675b0da8},
	}
	for _, tt := range tests {
		x := make([]FixpDBL, 2<<uint(tt.ldn))
		fillDITFFT512Input(x)
		DITFFT512(x, tt.ldn)
		if got := hashFixpDBL(x); got != tt.hash {
			t.Fatalf("DITFFT512 ldn %d hash = %#016x, want %#016x", tt.ldn, got, tt.hash)
		}
	}
}

func TestDITFFT512RejectsUnsupportedLengths(t *testing.T) {
	for _, ldn := range []int{3, 5, 10} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("DITFFT512 ldn %d did not panic", ldn)
				}
			}()
			var x [128]FixpDBL
			DITFFT512(x[:], ldn)
		}()
	}
}

func TestFFTDispatcher(t *testing.T) {
	t.Run("short kernels", func(t *testing.T) {
		x2 := [4]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000}
		scale := 11
		FFT(2, x2[:], &scale)
		if want := [4]FixpDBL{167772160, -50331648, 100663296, -83886080}; x2 != want {
			t.Fatalf("FFT length 2 = %v, want %v", x2, want)
		}
		if scale != 11+fftScaleFactor2 {
			t.Fatalf("FFT length 2 scale = %d, want %d", scale, 11+fftScaleFactor2)
		}

		x4 := [8]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000}
		scale = 11
		FFT(4, x4[:], &scale)
		if want := [8]FixpDBL{0, 0, 301989888, -201326592, 0, 0, 234881024, -67108864}; x4 != want {
			t.Fatalf("FFT length 4 = %v, want %v", x4, want)
		}
		if scale != 11+fftScaleFactor4 {
			t.Fatalf("FFT length 4 scale = %d, want %d", scale, 11+fftScaleFactor4)
		}

		x8 := [16]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000, 0x08000000, 0x04000000, -0x08000000, 0x01000000, 0x02000000, -0x01000000, -0x02000000, 0x00800000}
		scale = 11
		FFT(8, x8[:], &scale)
		if want := [16]FixpDBL{0, 18874368, 108375168, 5888384, 178257920, -54525952, -50163584, -165866880, 83886080, 6291456, 34231168, 44443264, 140509184, -37748736, 41774976, -85791360}; x8 != want {
			t.Fatalf("FFT length 8 = %v, want %v", x8, want)
		}
		if scale != 11+fftScaleFactor8 {
			t.Fatalf("FFT length 8 scale = %d, want %d", scale, 11+fftScaleFactor8)
		}
	})

	t.Run("radix lengths", func(t *testing.T) {
		tests := []struct {
			length     int
			scaleDelta int
			hash       uint64
		}{
			{length: 64, scaleDelta: fftScaleFactor64, hash: 0x24da1b06d4d1acd5},
			{length: 128, scaleDelta: fftScaleFactor128, hash: 0x6e30afc5327a54af},
			{length: 256, scaleDelta: fftScaleFactor256, hash: 0xf072fffa1018bb22},
			{length: 512, scaleDelta: fftScaleFactor512, hash: 0xec99e0d7675b0da8},
		}
		for _, tt := range tests {
			x := make([]FixpDBL, 2*tt.length)
			fillDITFFT512Input(x)
			scale := 11
			FFT(tt.length, x, &scale)
			if got := hashFixpDBL(x); got != tt.hash {
				t.Fatalf("FFT length %d hash = %#016x, want %#016x", tt.length, got, tt.hash)
			}
			if scale != 11+tt.scaleDelta {
				t.Fatalf("FFT length %d scale = %d, want %d", tt.length, scale, 11+tt.scaleDelta)
			}
		}
	})
}

func TestFFTRejectsUnsupportedLengths(t *testing.T) {
	for _, length := range []int{0, 3, 5, 16, 32, 120, 1024} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("FFT length %d did not panic", length)
				}
			}()
			var x [2048]FixpDBL
			scale := 0
			FFT(length, x[:], &scale)
		}()
	}
}

func TestFFTShortAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		x2 := [4]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000}
		x4 := [8]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000}
		x8 := [16]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000, 0x08000000, 0x04000000, -0x08000000, 0x01000000, 0x02000000, -0x01000000, -0x02000000, 0x00800000}
		FFT2(x2[:])
		FFT4(x4[:])
		FFT8(x8[:])
		fftSink = x2[0] ^ x4[0] ^ x8[0]
	})
	if allocs != 0 {
		t.Fatalf("short FFT helpers allocations = %v, want 0", allocs)
	}
}

func TestDITFFTAllocs(t *testing.T) {
	trig := []FixpSPK{
		{32767, 0}, {32610, 3212}, {32138, 6393}, {31357, 9512},
		{30274, 12540}, {28899, 15447}, {27246, 18205}, {25330, 20788},
		{23170, 23170}, {20788, 25330}, {18205, 27246}, {15447, 28899},
		{12540, 30274}, {9512, 31357}, {6393, 32138}, {3212, 32610},
	}
	allocs := testing.AllocsPerRun(1000, func() {
		x := [16]FixpDBL{0x10000000, -0x08000000, 0x04000000, 0x02000000, -0x10000000, 0x08000000, -0x04000000, -0x02000000, 0x08000000, 0x04000000, -0x08000000, 0x01000000, 0x02000000, -0x01000000, -0x02000000, 0x00800000}
		DITFFT(x[:], 3, trig, len(trig))
		fftSink = x[0]
	})
	if allocs != 0 {
		t.Fatalf("DITFFT allocations = %v, want 0", allocs)
	}
}

func TestFFTDispatcherAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var x [128]FixpDBL
		fillDITFFT512Input(x[:])
		scale := 3
		FFT(64, x[:], &scale)
		fftSink = x[0]
		fftScaleSink = scale
	})
	if allocs != 0 {
		t.Fatalf("FFT dispatcher allocations = %v, want 0", allocs)
	}
}

func TestDITFFT512Allocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		var x [128]FixpDBL
		fillDITFFT512Input(x[:])
		DITFFT512(x[:], 6)
		fftSink = x[0]
	})
	if allocs != 0 {
		t.Fatalf("DITFFT512 allocations = %v, want 0", allocs)
	}
}

func equalFixpDBL(a, b []FixpDBL) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func fillDITFFT512Input(x []FixpDBL) {
	for i := 0; i < len(x)/2; i++ {
		x[2*i] = FixpDBL(((i % 17) - 8) * 0x0102030)
		x[2*i+1] = FixpDBL(((i % 19) - 9) * 0x00f0e0d)
	}
}

func hashFixpSPK(x []FixpSPK) uint64 {
	h := uint64(14695981039346656037)
	for _, v := range x {
		u := uint16(v.Re)
		h = fnv64AddByte(h, byte(u))
		h = fnv64AddByte(h, byte(u>>8))
		u = uint16(v.Im)
		h = fnv64AddByte(h, byte(u))
		h = fnv64AddByte(h, byte(u>>8))
	}
	return h
}

func hashFixpDBL(x []FixpDBL) uint64 {
	h := uint64(14695981039346656037)
	for _, v := range x {
		u := uint32(v)
		h = fnv64AddByte(h, byte(u))
		h = fnv64AddByte(h, byte(u>>8))
		h = fnv64AddByte(h, byte(u>>16))
		h = fnv64AddByte(h, byte(u>>24))
	}
	return h
}

func fnv64AddByte(h uint64, b byte) uint64 {
	h ^= uint64(b)
	h *= 1099511628211
	return h
}
