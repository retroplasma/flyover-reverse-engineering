package bin

import (
	"encoding/binary"
	"math"
)

/*
 * binary writers
 */

// WriteUInt64 writes a uint64 value to data at offset
func WriteUInt64(data []byte, offset int, value uint64) {
	binary.LittleEndian.PutUint64(data[offset:], value)
}

// WriteInt64 writes an int64 value to data at offset
func WriteInt64(data []byte, offset int, value int64) {
	binary.LittleEndian.PutUint64(data[offset:], uint64(value))
}

// WriteInt32 writes an int32 value to data at offset
func WriteInt32(data []byte, offset int, value int32) {
	binary.LittleEndian.PutUint32(data[offset:], uint32(value))
}

// WriteUInt32 writes a uint32 value to data at offset
func WriteUInt32(data []byte, offset int, value uint32) {
	binary.LittleEndian.PutUint32(data[offset:], value)
}

// WriteInt16 writes an int16 value to data at offset
func WriteInt16(data []byte, offset int, value int16) {
	binary.LittleEndian.PutUint16(data[offset:], uint16(value))
}

// WriteUInt16 writes a uint16 value to data at offset
func WriteUInt16(data []byte, offset int, value uint16) {
	binary.LittleEndian.PutUint16(data[offset:], value)
}

/*
 * binary readers
 */

// ReadFloat32 reads a float32 value from data at offset
func ReadFloat32(data []byte, offset int) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(data[offset:]))
}

// ReadFloat64 reads a float64 value from data at offset
func ReadFloat64(data []byte, offset int) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
}

// ReadUInt8 reads a uint8 value from data at offset
func ReadUInt8(data []byte, offset int) uint8 {
	return uint8(data[offset])
}

// ReadInt8 reads an int8 value from data at offset
func ReadInt8(data []byte, offset int) int8 {
	return int8(data[offset])
}

// ReadInt16 reads an int16 value from data at offset
func ReadInt16(data []byte, offset int) int16 {
	return int16(binary.LittleEndian.Uint16(data[offset:]))
}

// ReadUInt16 reads a uint16 value from data at offset
func ReadUInt16(data []byte, offset int) uint16 {
	return binary.LittleEndian.Uint16(data[offset:])
}

// ReadInt32 reads an int32 value from data at offset
func ReadInt32(data []byte, offset int) int32 {
	return int32(binary.LittleEndian.Uint32(data[offset:]))
}

// ReadUInt32 reads a uint32 value from data at offset
func ReadUInt32(data []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(data[offset:])
}

// ReadInt32BE reads a big endian int32 value from data at offset
func ReadInt32BE(data []byte, offset int) int32 {
	return int32(binary.BigEndian.Uint32(data[offset:]))
}

// ReadUInt32BE reads a big endian uint32 value from data at offset
func ReadUInt32BE(data []byte, offset int) uint32 {
	return binary.BigEndian.Uint32(data[offset:])
}

// ReadInt64 reads an int64 value from data at offset
func ReadInt64(data []byte, offset int) int64 {
	return int64(binary.LittleEndian.Uint64(data[offset:]))
}

// ReadUInt64 reads a uint64 value from data at offset
func ReadUInt64(data []byte, offset int) uint64 {
	return binary.LittleEndian.Uint64(data[offset:])
}

// ByteSwapUInt64 returns a uint64 with its bytes swapped
func ByteSwapUInt64(value uint64) uint64 {
	return value>>56&0xff | value>>48&0xff<<8 | value>>40&0xff<<16 | value>>32&0xff<<24 |
		value>>24&0xff<<32 | value>>16&0xff<<40 | value>>8&0xff<<48 | value&0xff<<56
}

// ByteSwapUInt32 returns a uint32 with its bytes swapped
func ByteSwapUInt32(value uint32) uint32 {
	return value>>24&0xff | value>>16&0xff<<8 | value>>8&0xff<<16 | value&0xff<<24
}
