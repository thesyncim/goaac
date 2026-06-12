package fdkaac

import (
	"bytes"
	"testing"
)

func TestBitMaskTable(t *testing.T) {
	for i := uint32(0); i < 32; i++ {
		want := (uint32(1) << i) - 1
		if BitMask[i] != want {
			t.Fatalf("BitMask[%d] = %#x, want %#x", i, BitMask[i], want)
		}
	}
	if BitMask[32] != 0xffffffff {
		t.Fatalf("BitMask[32] = %#x, want 0xffffffff", BitMask[32])
	}
}

func TestBitStreamWriteBitsKnownVectors(t *testing.T) {
	// Expected bytes were checked against FDK-AAC v2.0.3 FDKwriteBits,
	// FDKwriteEscapedValue, FDKbyteAlign, and FDKfetchBuffer.
	tests := []struct {
		name  string
		write func(*BitStream)
		want  []byte
		bits  uint32
	}{
		{
			name: "single byte",
			write: func(bs *BitStream) {
				WriteBits(bs, 0x5, 3)
				WriteBits(bs, 0x13, 5)
			},
			want: []byte{0xb3},
			bits: 8,
		},
		{
			name: "mask value",
			write: func(bs *BitStream) {
				WriteBits(bs, 0xff, 3)
				ByteAlign(bs, 0)
			},
			want: []byte{0xe0},
			bits: 8,
		},
		{
			name: "cross cache then align",
			write: func(bs *BitStream) {
				WriteBits(bs, 0x12345678, 32)
				WriteBits(bs, 0xabc, 12)
				ByteAlign(bs, 0)
			},
			want: []byte{0x12, 0x34, 0x56, 0x78, 0xab, 0xc0},
			bits: 48,
		},
		{
			name: "escaped value",
			write: func(bs *BitStream) {
				WriteEscapedValue(bs, 42, 3, 4, 8)
				ByteAlign(bs, 0)
			},
			want: []byte{0xfe, 0x28},
			bits: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var storage [16]byte
			var bs BitStream
			if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
				t.Fatal(err)
			}
			tt.write(&bs)
			if got := BitStreamValidBits(&bs); got != tt.bits {
				t.Fatalf("valid bits = %d, want %d", got, tt.bits)
			}
			var out [16]byte
			n := FetchBuffer(&bs, out[:])
			if !bytes.Equal(out[:n], tt.want) {
				t.Fatalf("bytes = % x, want % x", out[:n], tt.want)
			}
		})
	}
}

func TestBitStreamPartialByteFetchAndAlignment(t *testing.T) {
	var storage [8]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}
	WriteBits(&bs, 0xa, 4)
	if got := BitStreamValidBits(&bs); got != 4 {
		t.Fatalf("valid bits = %d, want 4", got)
	}
	var out [8]byte
	if n := FetchBuffer(&bs, out[:]); n != 0 {
		t.Fatalf("partial-byte fetch wrote %d bytes, want 0", n)
	}
	if got := BitStreamValidBits(&bs); got != 4 {
		t.Fatalf("valid bits after partial fetch = %d, want 4", got)
	}
	ByteAlign(&bs, 0)
	if n := FetchBuffer(&bs, out[:]); n != 1 || out[0] != 0xa0 {
		t.Fatalf("aligned fetch = n:%d bytes:% x, want one byte a0", n, out[:n])
	}
}

func TestBitStreamFetchWrapsRingBuffer(t *testing.T) {
	var storage [8]byte
	var bs BitStream
	if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
		t.Fatal(err)
	}
	WriteBits(&bs, 0x01020304, 32)
	WriteBits(&bs, 0x050607, 24)

	var out [8]byte
	n := FetchBuffer(&bs, out[:])
	if n != 7 || !bytes.Equal(out[:n], []byte{1, 2, 3, 4, 5, 6, 7}) {
		t.Fatalf("first fetch = n:%d bytes:% x, want 01 02 03 04 05 06 07", n, out[:n])
	}

	WriteBits(&bs, 0xa1b2, 16)
	n = FetchBuffer(&bs, out[:])
	if n != 2 || !bytes.Equal(out[:n], []byte{0xa1, 0xb2}) {
		t.Fatalf("wrapped fetch = n:%d bytes:% x, want a1 b2", n, out[:n])
	}
}

func TestBitStreamRejectsInvalidBuffer(t *testing.T) {
	var bs BitStream
	if err := InitBitStream(&bs, make([]byte, 3), 0, BSWriter); err == nil {
		t.Fatal("InitBitStream with non-power-of-two buffer succeeded")
	}
	if err := InitBitStream(&bs, make([]byte, 8), 65, BSWriter); err == nil {
		t.Fatal("InitBitStream with too many valid bits succeeded")
	}
}

func TestBitStreamWriterAllocs(t *testing.T) {
	var storage [16]byte
	var out [16]byte
	var bs BitStream
	allocs := testing.AllocsPerRun(1000, func() {
		clear(storage[:])
		clear(out[:])
		if err := InitBitStream(&bs, storage[:], 0, BSWriter); err != nil {
			t.Fatal(err)
		}
		WriteBits(&bs, 0x12345678, 32)
		WriteBits(&bs, 0xabc, 12)
		ByteAlign(&bs, 0)
		if n := FetchBuffer(&bs, out[:]); n != 6 {
			t.Fatal("unexpected fetch length")
		}
	})
	if allocs != 0 {
		t.Fatalf("bitstream writer allocations = %v, want 0", allocs)
	}
}
