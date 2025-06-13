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

// Type-safe helper functions for Rec.Val access

// AsFloat32 safely extracts float32 value from Rec
func (r *Rec) AsFloat32() (float32, bool) {
	if val, ok := r.Val.(float32); ok {
		return val, true
	}
	return 0, false
}

// AsFloat64 safely extracts float64 value from Rec
func (r *Rec) AsFloat64() (float64, bool) {
	if val, ok := r.Val.(float64); ok {
		return val, true
	}
	return 0, false
}

// AsString safely extracts string value from Rec
func (r *Rec) AsString() (string, bool) {
	if val, ok := r.Val.(string); ok {
		return val, true
	}
	return "", false
}

// AsFloat32Array safely extracts []float32 value from Rec
func (r *Rec) AsFloat32Array() ([]float32, bool) {
	if val, ok := r.Val.([]float32); ok {
		return val, true
	}
	return nil, false
}

// AsFloat64Array safely extracts []float64 value from Rec
func (r *Rec) AsFloat64Array() ([]float64, bool) {
	if val, ok := r.Val.([]float64); ok {
		return val, true
	}
	return nil, false
}

// GetNumericValue attempts to convert any numeric type to float64
func (r *Rec) GetNumericValue() (float64, bool) {
	switch v := r.Val.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int8:
		return float64(v), true
	case uint8:
		return float64(v), true
	case int16:
		return float64(v), true
	case uint16:
		return float64(v), true
	case int32:
		return float64(v), true
	case uint32:
		return float64(v), true
	default:
		return 0, false
	}
}
