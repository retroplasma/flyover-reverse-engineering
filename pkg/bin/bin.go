package bin

import (
	"encoding/binary"
	"math"
)

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

func WriteInt32(data []byte, offset int, value int32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(value))
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
	data[offset+2] = buf[2]
	data[offset+3] = buf[3]
}

func WriteUInt32(data []byte, offset int, value uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, value)
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
	data[offset+2] = buf[2]
	data[offset+3] = buf[3]
}

func WriteInt16(data []byte, offset int, value int16) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(value))
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
}

func WriteUInt16(data []byte, offset int, value uint16) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, value)
	data[offset+0] = buf[0]
	data[offset+1] = buf[1]
}

/*
 * binary readers
 */

func ReadFloat32(data []byte, offset int) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(data[offset:]))
}

func ReadFloat64(data []byte, offset int) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
}

func ReadUInt8(data []byte, offset int) uint8 {
	return uint8(data[offset])
}

func ReadInt8(data []byte, offset int) int8 {
	return int8(data[offset])
}

func ReadInt16(data []byte, offset int) int16 {
	return int16(binary.LittleEndian.Uint16(data[offset:]))
}

func ReadUInt16(data []byte, offset int) uint16 {
	return binary.LittleEndian.Uint16(data[offset:])
}

func ReadInt32(data []byte, offset int) int32 {
	return int32(binary.LittleEndian.Uint32(data[offset:]))
}

func ReadUInt32(data []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(data[offset:])
}

func ReadInt64(data []byte, offset int) int64 {
	return int64(binary.LittleEndian.Uint64(data[offset:]))
}

func ReadUInt64(data []byte, offset int) uint64 {
	return binary.LittleEndian.Uint64(data[offset:])
}

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
