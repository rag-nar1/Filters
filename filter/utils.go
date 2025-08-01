package filter

import "bytes"

func SerializeUint(buf *bytes.Buffer, value uint64, size int) {
	byteData := make([]byte, size)
	for i := range size {
		byteData[i] = byte(value >> (i * 8))
	}
	buf.Write(byteData)
}

func DeserializeUint[T uint64 | uint32](buf *bytes.Buffer, size int) T {
	byteData := make([]byte, size)
	buf.Read(byteData)
	value := uint64(0)
	for i := range size {
		value |= uint64(byteData[i]) << (i * 8)
	}
	return T(value)
}

func NextPowerOfTwo(n uint32) uint32 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}