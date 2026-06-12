package wav

import (
	"encoding/binary"
	"fmt"
	"io"
)

func WriteS16(w io.Writer, samples []int16, sampleRate, channels int) error {
	if sampleRate <= 0 {
		return fmt.Errorf("wav: invalid sample rate %d", sampleRate)
	}
	if channels <= 0 {
		return fmt.Errorf("wav: invalid channel count %d", channels)
	}
	dataBytes := len(samples) * 2
	if dataBytes > int(^uint32(0)-44) {
		return fmt.Errorf("wav: file too large")
	}
	byteRate := sampleRate * channels * 2
	blockAlign := channels * 2
	header := make([]byte, 44)
	copy(header[0:], "RIFF")
	binary.LittleEndian.PutUint32(header[4:], uint32(36+dataBytes))
	copy(header[8:], "WAVE")
	copy(header[12:], "fmt ")
	binary.LittleEndian.PutUint32(header[16:], 16)
	binary.LittleEndian.PutUint16(header[20:], 1)
	binary.LittleEndian.PutUint16(header[22:], uint16(channels))
	binary.LittleEndian.PutUint32(header[24:], uint32(sampleRate))
	binary.LittleEndian.PutUint32(header[28:], uint32(byteRate))
	binary.LittleEndian.PutUint16(header[32:], uint16(blockAlign))
	binary.LittleEndian.PutUint16(header[34:], 16)
	copy(header[36:], "data")
	binary.LittleEndian.PutUint32(header[40:], uint32(dataBytes))
	if _, err := w.Write(header); err != nil {
		return err
	}
	var buf [4096]byte
	n := 0
	for _, s := range samples {
		if n+2 > len(buf) {
			if _, err := w.Write(buf[:n]); err != nil {
				return err
			}
			n = 0
		}
		binary.LittleEndian.PutUint16(buf[n:], uint16(s))
		n += 2
	}
	if n > 0 {
		_, err := w.Write(buf[:n])
		return err
	}
	return nil
}
