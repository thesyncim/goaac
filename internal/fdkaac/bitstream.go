package fdkaac

import "fmt"

const (
	maxBufSizeBytes uint32 = 0x10000000
	cacheBits       uint32 = 32
)

var BitMask = [...]uint32{
	0x0, 0x1, 0x3, 0x7, 0xf, 0x1f,
	0x3f, 0x7f, 0xff, 0x1ff, 0x3ff, 0x7ff,
	0xfff, 0x1fff, 0x3fff, 0x7fff, 0xffff, 0x1ffff,
	0x3ffff, 0x7ffff, 0xfffff, 0x1fffff, 0x3fffff, 0x7fffff,
	0xffffff, 0x1ffffff, 0x3ffffff, 0x7ffffff, 0xfffffff, 0x1fffffff,
	0x3fffffff, 0x7fffffff, 0xffffffff,
}

type BitBuffer struct {
	ValidBits   uint32
	ReadOffset  uint32
	WriteOffset uint32
	BitNdx      uint32

	Buffer  []byte
	BufBits uint32
}

func InitBitBuffer(h *BitBuffer, buffer []byte, validBits uint32) error {
	if h == nil {
		return fmt.Errorf("fdkaac: nil bit buffer")
	}
	bufSize := uint32(len(buffer))
	if bufSize == 0 || bufSize > maxBufSizeBytes || bufSize&(bufSize-1) != 0 {
		return fmt.Errorf("fdkaac: bit buffer size %d is not a supported power of two", bufSize)
	}
	bufBits := bufSize << 3
	if validBits > bufBits {
		return fmt.Errorf("fdkaac: valid bits %d exceed buffer bits %d", validBits, bufBits)
	}
	*h = BitBuffer{
		ValidBits: validBits,
		Buffer:    buffer,
		BufBits:   bufBits,
	}
	return nil
}

func ResetBitBuffer(h *BitBuffer) {
	h.ValidBits = 0
	h.ReadOffset = 0
	h.WriteOffset = 0
	h.BitNdx = 0
}

func Put(h *BitBuffer, value uint32, numberOfBits uint32) {
	if numberOfBits == 0 {
		return
	}
	checkBitCount(numberOfBits)

	byteOffset0 := h.BitNdx >> 3
	bitOffset := h.BitNdx & 0x07
	h.BitNdx = (h.BitNdx + numberOfBits) & (h.BufBits - 1)
	h.ValidBits += numberOfBits

	byteMask := h.bufSize() - 1
	byteOffset1 := (byteOffset0 + 1) & byteMask
	byteOffset2 := (byteOffset0 + 2) & byteMask
	byteOffset3 := (byteOffset0 + 3) & byteMask

	tmp := (value << (32 - numberOfBits)) >> bitOffset
	mask := ^((BitMask[numberOfBits] << (32 - numberOfBits)) >> bitOffset)

	cache := (uint32(h.Buffer[byteOffset0]) << 24) |
		(uint32(h.Buffer[byteOffset1]) << 16) |
		(uint32(h.Buffer[byteOffset2]) << 8) |
		uint32(h.Buffer[byteOffset3])
	cache = (cache & mask) | tmp

	h.Buffer[byteOffset0] = byte(cache >> 24)
	h.Buffer[byteOffset1] = byte(cache >> 16)
	h.Buffer[byteOffset2] = byte(cache >> 8)
	h.Buffer[byteOffset3] = byte(cache)

	if bitOffset+numberOfBits > 32 {
		byteOffset4 := (byteOffset0 + 4) & byteMask
		bits := (bitOffset + numberOfBits) & 7
		cache = uint32(h.Buffer[byteOffset4]) & ^(BitMask[bits] << (8 - bits))
		cache |= value << (8 - bits)
		h.Buffer[byteOffset4] = byte(cache)
	}
}

func PushBack(h *BitBuffer, numberOfBits uint32, config BitStreamConfig) {
	if config == BSReader {
		h.ValidBits += numberOfBits
	} else {
		h.ValidBits -= numberOfBits
	}
	h.BitNdx = uint32(int32(h.BitNdx)-int32(numberOfBits)) & (h.BufBits - 1)
}

func PushForward(h *BitBuffer, numberOfBits uint32, config BitStreamConfig) {
	if config == BSReader {
		h.ValidBits -= numberOfBits
	} else {
		h.ValidBits += numberOfBits
	}
	h.BitNdx = uint32(int32(h.BitNdx)+int32(numberOfBits)) & (h.BufBits - 1)
}

func ValidBits(h *BitBuffer) uint32 {
	return h.ValidBits
}

func FreeBits(h *BitBuffer) int32 {
	return int32(h.BufBits - h.ValidBits)
}

func Fetch(h *BitBuffer, out []byte) int {
	bToWrite := h.ValidBits >> 3
	noOfBytes := minU32(bToWrite, uint32(len(out)))
	total := uint32(0)

	for noOfBytes > 0 {
		bToWrite = h.bufSize() - h.WriteOffset
		bToWrite = minU32(bToWrite, noOfBytes)

		start := int(h.WriteOffset)
		end := int(h.WriteOffset + bToWrite)
		copy(out[total:total+bToWrite], h.Buffer[start:end])

		h.ValidBits -= bToWrite << 3
		total += bToWrite
		h.WriteOffset = (h.WriteOffset + bToWrite) & (h.bufSize() - 1)
		noOfBytes -= bToWrite
	}

	return int(total)
}

func (h *BitBuffer) bufSize() uint32 {
	return uint32(len(h.Buffer))
}

type BitStreamConfig uint8

const (
	BSReader BitStreamConfig = iota
	BSWriter
)

type BitStream struct {
	CacheWord   uint32
	BitsInCache uint32
	BitBuf      BitBuffer
	ConfigCache BitStreamConfig
}

func InitBitStream(h *BitStream, buffer []byte, validBits uint32, config BitStreamConfig) error {
	if h == nil {
		return fmt.Errorf("fdkaac: nil bitstream")
	}
	if err := InitBitBuffer(&h.BitBuf, buffer, validBits); err != nil {
		return err
	}
	h.CacheWord = 0
	h.BitsInCache = 0
	h.ConfigCache = config
	return nil
}

func ResetBitStream(h *BitStream, config BitStreamConfig) {
	ResetBitBuffer(&h.BitBuf)
	h.CacheWord = 0
	h.BitsInCache = 0
	h.ConfigCache = config
}

func WriteBits(h *BitStream, value uint32, numberOfBits uint32) uint8 {
	checkBitCount(numberOfBits)
	if h == nil {
		return uint8(numberOfBits)
	}

	validMask := BitMask[numberOfBits]
	if h.BitsInCache+numberOfBits < cacheBits {
		h.BitsInCache += numberOfBits
		h.CacheWord = (h.CacheWord << numberOfBits) | (value & validMask)
		return uint8(numberOfBits)
	}

	missingBits := cacheBits - h.BitsInCache
	remainingBits := numberOfBits - missingBits
	value &= validMask

	cacheWord := uint32(0)
	if missingBits != 32 {
		cacheWord = h.CacheWord << missingBits
	}
	cacheWord |= value >> remainingBits
	Put(&h.BitBuf, cacheWord, 32)

	h.CacheWord = value
	h.BitsInCache = remainingBits
	return uint8(numberOfBits)
}

func WriteEscapedValue(h *BitStream, value, nBits1, nBits2, nBits3 uint32) uint8 {
	nbits := uint8(0)
	tmp := (uint32(1) << nBits1) - 1
	if value < tmp {
		nbits += WriteBits(h, value, nBits1)
		return nbits
	}

	nbits += WriteBits(h, tmp, nBits1)
	value -= tmp
	tmp = (uint32(1) << nBits2) - 1

	if value < tmp {
		nbits += WriteBits(h, value, nBits2)
		return nbits
	}

	nbits += WriteBits(h, tmp, nBits2)
	value -= tmp
	nbits += WriteBits(h, value, nBits3)
	return nbits
}

func SyncCache(h *BitStream) {
	if h.ConfigCache == BSReader {
		PushBack(&h.BitBuf, h.BitsInCache, h.ConfigCache)
	} else if h.BitsInCache != 0 {
		Put(&h.BitBuf, h.CacheWord, h.BitsInCache)
	}
	h.BitsInCache = 0
	h.CacheWord = 0
}

func ByteAlign(h *BitStream, alignmentAnchor uint32) {
	SyncCache(h)
	if h.ConfigCache == BSReader {
		delta := uint32((8 - ((int32(alignmentAnchor) - int32(ValidBits(&h.BitBuf))) & 0x07)) & 0x07)
		PushForward(&h.BitBuf, delta, h.ConfigCache)
		return
	}
	Put(&h.BitBuf, 0, (8-((ValidBits(&h.BitBuf)-alignmentAnchor)&0x07))&0x07)
}

func BitStreamValidBits(h *BitStream) uint32 {
	SyncCache(h)
	return ValidBits(&h.BitBuf)
}

func BitStreamFreeBits(h *BitStream) int32 {
	return FreeBits(&h.BitBuf)
}

func FetchBuffer(h *BitStream, out []byte) int {
	SyncCache(h)
	return Fetch(&h.BitBuf, out)
}

func checkBitCount(n uint32) {
	if n > 32 {
		panic(fmt.Sprintf("fdkaac: bit count %d exceeds 32", n))
	}
}

func minU32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
