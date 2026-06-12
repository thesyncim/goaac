package fdkaac

import (
	"slices"
	"testing"
)

var scaleSink int

func TestScaleValuesVectors(t *testing.T) {
	baseD := []FixpDBL{0, 1, -1, 0x12345678, -0x1234567, 0x40000000, MaxValDBL, MinValDBL, 0x01000000}
	baseS := []FixpSGL{0, 1, -1, 12345, -23456, 0x4000, MaxValSGL, MinValSGL, 256}
	tests := []struct {
		scale int
		dbl   []FixpDBL
		sgl   []FixpSGL
		ss    []FixpSGL
		pcm   []FixpSGL
	}{
		{
			scale: 0,
			dbl:   []FixpDBL{0, 1, -1, 305419896, -19088743, 1073741824, 2147483647, -2147483648, 16777216},
			sgl:   []FixpSGL{0, 1, -1, 12345, -23456, 16384, 32767, -32768, 256},
			ss:    []FixpSGL{0, 1, -1, 12345, -23456, 16384, 32767, -32768, 256},
			pcm:   []FixpSGL{0, 0, -1, 4660, -292, 16384, 32767, -32768, 256},
		},
		{
			scale: 1,
			dbl:   []FixpDBL{0, 2, -2, 610839792, -38177486, -2147483648, -2, 0, 33554432},
			sgl:   []FixpSGL{0, 2, -2, 24690, 18624, -32768, -2, 0, 512},
			ss:    []FixpSGL{0, 2, -2, 24690, 18624, -32768, -2, 0, 512},
			pcm:   []FixpSGL{0, 0, -1, 9320, -583, -32768, -1, 0, 512},
		},
		{
			scale: -4,
			dbl:   []FixpDBL{0, 0, -1, 19088743, -1193047, 67108864, 134217727, -134217728, 1048576},
			sgl:   []FixpSGL{0, 0, -1, 771, -1466, 1024, 2047, -2048, 16},
			ss:    []FixpSGL{0, 0, -1, 771, -1466, 1024, 2047, -2048, 16},
			pcm:   []FixpSGL{0, 0, -1, 291, -19, 1024, 2047, -2048, 16},
		},
		{
			scale: 31,
			dbl:   []FixpDBL{0, -2147483648, -2147483648, 0, -2147483648, 0, -2147483648, 0, 0},
			sgl:   []FixpSGL{0, -32768, -32768, -32768, 0, 0, -32768, 0, 0},
			ss:    []FixpSGL{0, 0, 0, 0, 0, 0, 0, 0, 0},
			pcm:   []FixpSGL{0, -32768, -32768, 0, -32768, 0, -32768, 0, 0},
		},
		{
			scale: -31,
			dbl:   []FixpDBL{0, 0, -1, 0, -1, 0, 0, -1, 0},
			sgl:   []FixpSGL{0, 0, -1, 0, -1, 0, 0, -1, 0},
			ss:    []FixpSGL{0, 0, -1, 0, -1, 0, 0, -1, 0},
			pcm:   []FixpSGL{0, 0, -1, 0, -1, 0, 0, -1, 0},
		},
	}

	for _, tt := range tests {
		dbl := slices.Clone(baseD)
		ScaleValuesDBL(dbl, tt.scale)
		if !slices.Equal(dbl, tt.dbl) {
			t.Fatalf("ScaleValuesDBL scale %d = %v, want %v", tt.scale, dbl, tt.dbl)
		}

		sgl := slices.Clone(baseS)
		ScaleValuesSGL(sgl, tt.scale)
		if !slices.Equal(sgl, tt.sgl) {
			t.Fatalf("ScaleValuesSGL scale %d = %v, want %v", tt.scale, sgl, tt.sgl)
		}

		dstD := make([]FixpDBL, len(baseD))
		ScaleValuesDBLToDBL(dstD, baseD, tt.scale)
		if !slices.Equal(dstD, tt.dbl) {
			t.Fatalf("ScaleValuesDBLToDBL scale %d = %v, want %v", tt.scale, dstD, tt.dbl)
		}

		dstS := make([]FixpSGL, len(baseS))
		ScaleValuesSGLToSGL(dstS, baseS, tt.scale)
		if !slices.Equal(dstS, tt.ss) {
			t.Fatalf("ScaleValuesSGLToSGL scale %d = %v, want %v", tt.scale, dstS, tt.ss)
		}

		pcm := make([]FixpSGL, len(baseD))
		ScaleValuesPCMFromDBL(pcm, baseD, tt.scale)
		if !slices.Equal(pcm, tt.pcm) {
			t.Fatalf("ScaleValuesPCMFromDBL scale %d = %v, want %v", tt.scale, pcm, tt.pcm)
		}
	}
}

func TestScaleValuesSaturateVectors(t *testing.T) {
	baseD := []FixpDBL{0, 1, -1, 0x12345678, -0x1234567, 0x40000000, MaxValDBL, MinValDBL, 0x01000000}
	baseS := []FixpSGL{0, 1, -1, 12345, -23456, 0x4000, MaxValSGL, MinValSGL, 256}
	tests := []struct {
		scale int
		dbl   []FixpDBL
		sgl   []FixpSGL
		sd    []FixpSGL
	}{
		{
			scale: 0,
			dbl:   []FixpDBL{0, 1, -1, 305419896, -19088743, 1073741824, 2147483647, -2147483648, 16777216},
			sgl:   []FixpSGL{0, 1, -1, 12345, -23456, 16384, 32767, -32768, 256},
			sd:    []FixpSGL{0, 0, 0, 4660, -291, 16384, 32767, -32768, 256},
		},
		{
			scale: 1,
			dbl:   []FixpDBL{0, 2, -2, 610839792, -38177486, 2147483647, 2147483647, -2147483647, 33554432},
			sgl:   []FixpSGL{0, 2, -2, 24690, -32768, 32767, 32767, -32768, 512},
			sd:    []FixpSGL{0, 0, 0, 9321, -583, 32767, 32767, -32768, 512},
		},
		{
			scale: -4,
			dbl:   []FixpDBL{0, 0, 0, 19088743, -1193047, 67108864, 134217727, -134217728, 1048576},
			sgl:   []FixpSGL{0, 0, -1, 771, -1466, 1024, 2047, -2048, 16},
			sd:    []FixpSGL{0, 0, 0, 291, -18, 1024, 2048, -2048, 16},
		},
		{
			scale: 31,
			dbl:   []FixpDBL{0, 2147483647, -2147483647, 2147483647, -2147483647, 2147483647, 2147483647, -2147483647, 2147483647},
			sgl:   []FixpSGL{0, 32767, -32768, 32767, -32768, 32767, 32767, -32768, 32767},
			sd:    []FixpSGL{0, 32767, -32768, 32767, -32768, 32767, 32767, -32768, 32767},
		},
	}

	for _, tt := range tests {
		dbl := slices.Clone(baseD)
		ScaleValuesSaturateDBL(dbl, tt.scale)
		if !slices.Equal(dbl, tt.dbl) {
			t.Fatalf("ScaleValuesSaturateDBL scale %d = %v, want %v", tt.scale, dbl, tt.dbl)
		}

		dstD := make([]FixpDBL, len(baseD))
		ScaleValuesSaturateDBLToDBL(dstD, baseD, tt.scale)
		if !slices.Equal(dstD, tt.dbl) {
			t.Fatalf("ScaleValuesSaturateDBLToDBL scale %d = %v, want %v", tt.scale, dstD, tt.dbl)
		}

		sgl := slices.Clone(baseS)
		ScaleValuesSaturateSGL(sgl, tt.scale)
		if !slices.Equal(sgl, tt.sgl) {
			t.Fatalf("ScaleValuesSaturateSGL scale %d = %v, want %v", tt.scale, sgl, tt.sgl)
		}

		dstS := make([]FixpSGL, len(baseS))
		ScaleValuesSaturateSGLToSGL(dstS, baseS, tt.scale)
		if !slices.Equal(dstS, tt.sgl) {
			t.Fatalf("ScaleValuesSaturateSGLToSGL scale %d = %v, want %v", tt.scale, dstS, tt.sgl)
		}

		fromD := make([]FixpSGL, len(baseD))
		ScaleValuesSaturateSGLFromDBL(fromD, baseD, tt.scale)
		if !slices.Equal(fromD, tt.sd) {
			t.Fatalf("ScaleValuesSaturateSGLFromDBL scale %d = %v, want %v", tt.scale, fromD, tt.sd)
		}
	}
}

func TestScaleValuesWithFactorVectors(t *testing.T) {
	baseD := []FixpDBL{0, 1, -1, 0x12345678, -0x1234567, 0x40000000, MaxValDBL, MinValDBL, 0x01000000}
	tests := []struct {
		factor FixpDBL
		scale  int
		want   []FixpDBL
	}{
		{factor: 0x40000000, scale: -2, want: []FixpDBL{0, 0, -1, 38177487, -2386093, 134217728, 268435455, -268435456, 2097152}},
		{factor: 0x40000000, scale: 1, want: []FixpDBL{0, 0, -4, 305419896, -19088744, 1073741824, 2147483644, -2147483648, 16777216}},
		{factor: MinValDBL, scale: 0, want: []FixpDBL{0, -2, 0, -305419896, 19088742, -1073741824, -2147483648, -2147483648, -16777216}},
		{factor: 0x12345678, scale: 3, want: []FixpDBL{0, 0, -16, 347499968, -21718752, 1221679584, -1851608144, 1851608128, 19088736}},
	}
	for _, tt := range tests {
		got := slices.Clone(baseD)
		ScaleValuesWithFactorDBL(got, tt.factor, tt.scale)
		if !slices.Equal(got, tt.want) {
			t.Fatalf("ScaleValuesWithFactorDBL factor %d scale %d = %v, want %v", tt.factor, tt.scale, got, tt.want)
		}
	}
}

func TestGetScalefactorVectors(t *testing.T) {
	baseD := []FixpDBL{0, 1, -1, 0x12345678, -0x1234567, 0x40000000, MaxValDBL, MinValDBL, 0x01000000}
	baseS := []FixpSGL{0, 1, -1, 12345, -23456, 0x4000, MaxValSGL, MinValSGL, 256}
	if got := GetScalefactorDBL(baseD); got != 0 {
		t.Fatalf("GetScalefactorDBL = %d, want 0", got)
	}
	if got := GetScalefactorSGL(baseS); got != 0 {
		t.Fatalf("GetScalefactorSGL = %d, want 0", got)
	}
	if got := GetScalefactorShort(baseS); got != 0 {
		t.Fatalf("GetScalefactorShort = %d, want 0", got)
	}
	pcmStride := []FixpSGL{1, 99, -2, 99, MaxValSGL, 99, MinValSGL, 99, 12345}
	if got := GetScalefactorPCM(pcmStride, 2); got != 0 {
		t.Fatalf("GetScalefactorPCM = %d, want 0", got)
	}
	if got := GetScalefactorDBL([]FixpDBL{0, 1, -1, 2, -2, 0x10000, -0x10000}); got != 14 {
		t.Fatalf("GetScalefactorDBL small = %d, want 14", got)
	}
	if got := GetScalefactorSGL([]FixpSGL{0, 1, -1, 2, -2, 256, -256}); got != 6 {
		t.Fatalf("GetScalefactorSGL small = %d, want 6", got)
	}
}

func TestScaleAllocs(t *testing.T) {
	dbl := []FixpDBL{0, 1, -1, 0x12345678, -0x1234567, 0x40000000, MaxValDBL, MinValDBL}
	sgl := []FixpSGL{0, 1, -1, 12345, -23456, 0x4000, MaxValSGL, MinValSGL}
	dstD := make([]FixpDBL, len(dbl))
	dstS := make([]FixpSGL, len(sgl))
	allocs := testing.AllocsPerRun(1000, func() {
		ScaleValuesDBLToDBL(dstD, dbl, 3)
		ScaleValuesSaturateSGLFromDBL(dstS, dstD, -4)
		ScaleValuesWithFactorDBL(dstD, 0x40000000, 1)
		scaleSink = GetScalefactorSGL(dstS)
	})
	if allocs != 0 {
		t.Fatalf("scale helpers allocations = %v, want 0", allocs)
	}
}
