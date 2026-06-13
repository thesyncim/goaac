package fdkaac

import "testing"

var preEchoSink FixpDBL
var preEchoIntSink int

func TestFDKaacEncInitPreEchoControlVectors(t *testing.T) {
	pcm := [...]FixpDBL{0x01000000, 0x02000000, 0x03000000, 0x04000000, 0x05000000, 0x06000000}
	var thresholdNm1 [len(pcm)]FixpDBL
	calcPreEcho := 0
	mdctScalenm1 := 0

	FDKaacEncInitPreEchoControl(thresholdNm1[:], &calcPreEcho, len(pcm), pcm[:], &mdctScalenm1)

	if thresholdNm1 != pcm {
		t.Fatalf("init threshold history = %v, want %v", thresholdNm1, pcm)
	}
	if calcPreEcho != 1 || mdctScalenm1 != 8 {
		t.Fatalf("init state calcPreEcho=%d mdctScalenm1=%d", calcPreEcho, mdctScalenm1)
	}
	if got, want := hashFixpDBL(thresholdNm1[:]), uint64(0xe2df0313db0ac5f0); got != want {
		t.Fatalf("init threshold history hash = %#016x, want %#016x", got, want)
	}
}

func TestFDKaacEncPreEchoControlVectors(t *testing.T) {
	t.Run("skip", func(t *testing.T) {
		thresholdNm1 := [...]FixpDBL{1, 2, 3, 4, 5, 6}
		threshold := [...]FixpDBL{0x01100000, 0x02200000, 0x03300000, 0x04400000, 0x05500000, 0x06600000}
		mdctScalenm1 := 8

		FDKaacEncPreEchoControl(thresholdNm1[:], 0, len(threshold), 2, 0x0148, threshold[:], 11, &mdctScalenm1)

		want := [...]FixpDBL{17825792, 35651584, 53477376, 71303168, 89128960, 106954752}
		if thresholdNm1 != want || threshold != want {
			t.Fatalf("skip thresholdNm1=%v threshold=%v, want %v", thresholdNm1, threshold, want)
		}
		if mdctScalenm1 != 11 {
			t.Fatalf("skip mdctScalenm1 = %d, want 11", mdctScalenm1)
		}
		if got, wantHash := hashFixpDBL(threshold[:]), uint64(0x43ba2bc26a21f480); got != wantHash {
			t.Fatalf("skip threshold hash = %#016x, want %#016x", got, wantHash)
		}
	})

	t.Run("current-scale-larger", func(t *testing.T) {
		thresholdNm1 := [...]FixpDBL{0x10000000, 0x08000000, 0x04000000, 0x02000000, 0x01000000, 0x00800000}
		threshold := [...]FixpDBL{0x20000000, 0x04000000, 0x01000000, 0x00200000, 0x00080000, 0x00010000}
		mdctScalenm1 := 8

		FDKaacEncPreEchoControl(thresholdNm1[:], 1, len(threshold), 2, 0x0148, threshold[:], 10, &mdctScalenm1)

		wantHistory := [...]FixpDBL{536870912, 67108864, 16777216, 2097152, 524288, 65536}
		wantThreshold := [...]FixpDBL{33554432, 16777216, 8388608, 2097152, 524288, 65536}
		if thresholdNm1 != wantHistory {
			t.Fatalf("larger-scale history = %v, want %v", thresholdNm1, wantHistory)
		}
		if threshold != wantThreshold {
			t.Fatalf("larger-scale threshold = %v, want %v", threshold, wantThreshold)
		}
		if mdctScalenm1 != 10 {
			t.Fatalf("larger-scale mdctScalenm1 = %d, want 10", mdctScalenm1)
		}
		if got, wantHash := hashFixpDBL(thresholdNm1[:]), uint64(0x5713f3573db45be7); got != wantHash {
			t.Fatalf("larger-scale history hash = %#016x, want %#016x", got, wantHash)
		}
		if got, wantHash := hashFixpDBL(threshold[:]), uint64(0xdf6f632e64536b0d); got != wantHash {
			t.Fatalf("larger-scale threshold hash = %#016x, want %#016x", got, wantHash)
		}
	})

	t.Run("previous-scale-larger", func(t *testing.T) {
		thresholdNm1 := [...]FixpDBL{0x00100000, 0x00200000, 0x00400000, 0x00800000, 0x01000000, 0x02000000}
		threshold := [...]FixpDBL{0x20000000, 0x10000000, 0x04000000, 0x02000000, 0x00800000, 0x00200000}
		mdctScalenm1 := 9

		FDKaacEncPreEchoControl(thresholdNm1[:], 1, len(threshold), 2, 0x0148, threshold[:], 6, &mdctScalenm1)

		wantHistory := [...]FixpDBL{536870912, 268435456, 67108864, 33554432, 8388608, 2097152}
		wantThreshold := [...]FixpDBL{134217728, 268435456, 67108864, 33554432, 8388608, 2097152}
		if thresholdNm1 != wantHistory {
			t.Fatalf("previous-scale history = %v, want %v", thresholdNm1, wantHistory)
		}
		if threshold != wantThreshold {
			t.Fatalf("previous-scale threshold = %v, want %v", threshold, wantThreshold)
		}
		if mdctScalenm1 != 6 {
			t.Fatalf("previous-scale mdctScalenm1 = %d, want 6", mdctScalenm1)
		}
		if got, wantHash := hashFixpDBL(thresholdNm1[:]), uint64(0x8d8b47f7b1576503); got != wantHash {
			t.Fatalf("previous-scale history hash = %#016x, want %#016x", got, wantHash)
		}
		if got, wantHash := hashFixpDBL(threshold[:]), uint64(0x08ee23952bb3227b); got != wantHash {
			t.Fatalf("previous-scale threshold hash = %#016x, want %#016x", got, wantHash)
		}
	})
}

func TestFDKaacEncPreEchoControlRejectsInvalid(t *testing.T) {
	var x [2]FixpDBL
	calcPreEcho := 0
	mdctScale := 0
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "init short history",
			fn: func() {
				FDKaacEncInitPreEchoControl(x[:1], &calcPreEcho, 2, x[:], &mdctScale)
			},
		},
		{
			name: "init nil state",
			fn: func() {
				FDKaacEncInitPreEchoControl(x[:], nil, 2, x[:], &mdctScale)
			},
		},
		{
			name: "negative bands",
			fn: func() {
				FDKaacEncPreEchoControl(x[:], 1, -1, 2, 0x0148, x[:], 0, &mdctScale)
			},
		},
		{
			name: "short current threshold",
			fn: func() {
				FDKaacEncPreEchoControl(x[:], 1, 2, 2, 0x0148, x[:1], 0, &mdctScale)
			},
		},
		{
			name: "nil scale",
			fn: func() {
				FDKaacEncPreEchoControl(x[:], 1, 2, 2, 0x0148, x[:], 0, nil)
			},
		},
		{
			name: "too-large current scale shift",
			fn: func() {
				prev := 0
				FDKaacEncPreEchoControl(x[:], 1, 2, 2, 0x0148, x[:], 16, &prev)
			},
		},
		{
			name: "too-large previous scale shift",
			fn: func() {
				prev := 16
				FDKaacEncPreEchoControl(x[:], 1, 2, 2, 0x0148, x[:], 0, &prev)
			},
		},
	}

	for _, tt := range tests {
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

func TestFDKaacEncPreEchoControlAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		pcm := [...]FixpDBL{0x01000000, 0x02000000, 0x03000000, 0x04000000, 0x05000000, 0x06000000}
		var thresholdNm1 [len(pcm)]FixpDBL
		calcPreEcho := 0
		mdctScalenm1 := 0
		FDKaacEncInitPreEchoControl(thresholdNm1[:], &calcPreEcho, len(pcm), pcm[:], &mdctScalenm1)

		threshold := [...]FixpDBL{0x20000000, 0x04000000, 0x01000000, 0x00200000, 0x00080000, 0x00010000}
		FDKaacEncPreEchoControl(thresholdNm1[:], calcPreEcho, len(threshold), 2, 0x0148, threshold[:], 10, &mdctScalenm1)
		preEchoSink = threshold[0] + thresholdNm1[0]
		preEchoIntSink = calcPreEcho + mdctScalenm1
	})
	if allocs != 0 {
		t.Fatalf("FDKaacEncPreEchoControl allocations = %v, want 0", allocs)
	}
}
