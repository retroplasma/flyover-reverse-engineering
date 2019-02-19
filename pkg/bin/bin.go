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
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, value)
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
	data[offset+2] = buf[2]
	data[offset+3] = buf[3]
	data[offset+4] = buf[4]
	data[offset+5] = buf[5]
	data[offset+6] = buf[6]
	data[offset+7] = buf[7]
}

// WriteInt64 writes an int64 value to data at offset
func WriteInt64(data []byte, offset int, value int64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
	data[offset+2] = buf[2]
	data[offset+3] = buf[3]
	data[offset+4] = buf[4]
	data[offset+5] = buf[5]
	data[offset+6] = buf[6]
	data[offset+7] = buf[7]
}

// WriteInt32 writes an int32 value to data at offset
func WriteInt32(data []byte, offset int, value int32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(value))
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
	data[offset+2] = buf[2]
	data[offset+3] = buf[3]
}

// WriteUInt32 writes a uint32 value to data at offset
func WriteUInt32(data []byte, offset int, value uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, value)
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
	data[offset+2] = buf[2]
	data[offset+3] = buf[3]
}

// WriteInt16 writes an int16 value to data at offset
func WriteInt16(data []byte, offset int, value int16) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(value))
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
}

// WriteUInt16 writes a uint16 value to data at offset
func WriteUInt16(data []byte, offset int, value uint16) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, value)
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
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
	buf1 := make([]byte, 8)
	buf2 := make([]byte, 8)
	WriteUInt64(buf1, 0, value)
	buf2[0] = buf1[7]
	buf2[1] = buf1[6]
	buf2[2] = buf1[5]
	buf2[3] = buf1[4]
	buf2[4] = buf1[3]
	buf2[5] = buf1[2]
	buf2[6] = buf1[1]
	buf2[7] = buf1[0]
	return ReadUInt64(buf2, 0)
}

// ByteSwapUInt32 returns a uint32 with its bytes swapped
func ByteSwapUInt32(value uint32) uint32 {
	buf1 := make([]byte, 4)
	buf2 := make([]byte, 4)
	WriteUInt32(buf1, 0, value)
	buf2[0] = buf1[3]
	buf2[1] = buf1[2]
	buf2[2] = buf1[1]
	buf2[3] = buf1[0]
	return ReadUInt32(buf2, 0)
}
