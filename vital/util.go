package vital

import (
	"encoding/binary"
	"math"
)

func bytesToFloat32(b []byte) float32 {
	bits := binary.LittleEndian.Uint32(b)
	return math.Float32frombits(bits)
}

func bytesToFloat64(b []byte) float64 {
	bits := binary.LittleEndian.Uint64(b)
	return math.Float64frombits(bits)
}

func unpackStr(b []byte, pos int) (string, int) {
	if pos+4 > len(b) {
		return "", 0
	}
	strlen := int(binary.LittleEndian.Uint32(b[pos : pos+4]))
	pos += 4
	if pos+strlen > len(b) {
		return "", 4
	}
	val := string(b[pos : pos+strlen])
	return val, 4 + strlen
}
