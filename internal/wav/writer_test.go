package wav

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestWriteS16(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteS16(&buf, []int16{-1, 0, 1, 2}, 44100, 2); err != nil {
		t.Fatal(err)
	}
	out := buf.Bytes()
	if string(out[:4]) != "RIFF" || string(out[8:12]) != "WAVE" {
		t.Fatalf("bad RIFF/WAVE header: % x", out[:12])
	}
	if binary.LittleEndian.Uint32(out[40:]) != 8 {
		t.Fatalf("data bytes = %d, want 8", binary.LittleEndian.Uint32(out[40:]))
	}
	if len(out) != 52 {
		t.Fatalf("len = %d, want 52", len(out))
	}
}
