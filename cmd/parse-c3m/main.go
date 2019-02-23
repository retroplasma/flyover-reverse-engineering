package main

import (
	"flyover-reverse-engineering/pkg/bin"
	"flyover-reverse-engineering/pkg/dec/huffman"
	"flyover-reverse-engineering/pkg/mth"
	"flyover-reverse-engineering/pkg/oth"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var l = log.New(os.Stderr, "", 0)
var out = false

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [c3m_file]\n", os.Args[0])
		os.Exit(1)
	}

	if len(os.Args) == 3 && os.Args[2] == "out" {
		out = true
	}

	file := os.Args[1]
	data, err := ioutil.ReadFile(file)
	oth.CheckPanic(err)

	l.Printf("%v\n\n", time.Now())

	l.Printf("File size: %d bytes\n", len(data))
	if len(data) < 4 || data[0] != 'C' || data[1] != '3' || data[2] != 'M' {
		panic("Invalid C3M header")
	}
	switch data[4] {
	case 0x03:
		l.Println("C3M v3")
		parseC3Mv3(data)
	case 0x02:
		l.Println("C3M v2")
		panic("Parser not implemented")
	default:
		l.Println("C3M v1")
		panic("Parser not implemented")
	}
}

func parseC3Mv3(data []byte) {
	size := len(data)
	if size < 134 || data[0] != 'C' || data[1] != '3' || data[2] != 'M' || data[3] != 0x03 {
		panic("Invalid C3M v3 data")
	}

	l.SetPrefix(l.Prefix() + "  ")

	numberOfItems := int(data[5])
	l.Printf("Number of items: %d\n", numberOfItems)
	processedItems := 0
	offset := 6

	pfx := l.Prefix()
	for {
		l.SetPrefix(pfx)
		switch data[offset] {
		case 0:
			l.Printf("Header at 0x%x (%d)", offset, offset)
			l.SetPrefix(l.Prefix() + "  ")
			parseHeader(data, &offset)
		case 1:
			l.Printf("Material at 0x%x (%d)", offset, offset)
			l.SetPrefix(l.Prefix() + "  ")
			parseMaterial(data, &offset)
		case 2:
			l.Printf("Mesh at 0x%x (%d)", offset, offset)
			l.SetPrefix(l.Prefix() + "  ")
			parseMesh(data, &offset)
		case 3:
			l.Printf("Scene Graph? / Animation? at 0x%x (%d)", offset, offset)
			l.SetPrefix(l.Prefix() + "  ")
			l.Printf("Not implemented, can't skip yet")
			os.Exit(1)
		default:
			panic("Invalid item type")
		}

		if processedItems+1 >= numberOfItems {
			l.Println("All items processed")
			return
		}
		processedItems++
	}
}

func parseHeader(data []byte, offset *int) {

	qx := bin.ReadFloat64(data, *offset+9)
	qy := bin.ReadFloat64(data, *offset+17)
	qz := bin.ReadFloat64(data, *offset+25)
	qw := bin.ReadFloat64(data, *offset+33)
	l.Printf("Rotation Quaternion XYZW:     [ %f, %f, %f, %f ]\n", qx, qy, qz, qw)

	x := bin.ReadFloat64(data, *offset+41)
	y := bin.ReadFloat64(data, *offset+49)
	z := bin.ReadFloat64(data, *offset+57)
	l.Printf("Translation ECEF XYZ:         [ %f, %f, %f ]\n", x, y, z)

	m := mth.QuaternionToMatrix(qx, qy, qz, qw)
	l.Printf("=> Transformation Matrix 4x4: [% f,% f,% f,% f,\n", m[0], m[1], m[2], x)
	l.Printf("                               % f,% f,% f,% f,\n", m[3], m[4], m[5], y)
	l.Printf("                               % f,% f,% f,% f,\n", m[6], m[7], m[8], z)
	l.Printf("                               % f,% f,% f,% f ]\n", 0.0, 0.0, 0.0, 1.0)

	l.Printf("Scale: ?")

	*offset += 113
}

func parseMaterial(data []byte, offset *int) {
	*offset += 5
	numberOfItems := int(bin.ReadInt32(data, *offset+0))
	l.Printf("Number of materials: %d \n", numberOfItems)
	processedItems := 0
	*offset += 4

	pfx := l.Prefix()
	l.SetPrefix(l.Prefix() + "  ")

	for {
		l.SetPrefix(pfx)
		materialType := data[*offset]

		switch materialType {
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10:
			l.Printf("- Material type: %d", materialType)
			l.SetPrefix(l.Prefix() + "    ")

			textureFormat := data[*offset+3]
			textureOffset := bin.ReadInt32(data, *offset+4)
			textureLength := bin.ReadInt32(data, *offset+8)
			textureLength2 := bin.ReadInt32(data, *offset+12)

			l.Printf("texOff: %d, texLen1: %d, texLen2: %d", textureOffset, textureLength, textureLength2)
			switch textureFormat {
			case 0:
				l.Printf("Format: JPEG")

				if out {
					fn := fmt.Sprintf("/tmp/jpg/%d.jpg", processedItems)
					err := ioutil.WriteFile(fn, data[textureOffset:textureOffset+textureLength2], 0655)
					oth.CheckPanic(err)
					l.Printf("Exported: %s", fn)
				}

				*offset += 16
			default:
				panic(fmt.Sprintf("Unsupported textureFormat %d", textureFormat))
			}
		default:
			panic(fmt.Sprintf("materialType %d not implemented", materialType))
		}

		l.SetPrefix(pfx)
		if processedItems+1 >= numberOfItems {
			l.Println("All materials processed")
			return
		}
		processedItems++
	}
}

func parseMesh(data []byte, offset *int) {
	//l.Printf("Not implemented, can't skip yet")
	*offset += 5
	numberOfItems := int(bin.ReadInt32(data, *offset+0))
	l.Printf("Number of meshes: %d \n", numberOfItems)
	//processedItems := 0
	*offset += 4

	pfx := l.Prefix()
	l.SetPrefix(l.Prefix() + "  ")

	for {
		l.SetPrefix(pfx)
		meshType := bin.ReadInt8(data, *offset+0)
		l.Printf("unknown_1_2: %d + %d<<8 = %d", bin.ReadInt8(data, *offset+1), bin.ReadInt8(data, *offset+2), int(bin.ReadInt8(data, *offset+1))+int(bin.ReadInt8(data, *offset+2))<<8)
		l.Println()

		switch meshType {
		case 2:
			offset3 := *offset + 3

			l.Printf("- Mesh type: %d \n", meshType)
			l.SetPrefix(l.Prefix() + "    ")

			unknownA8 := bin.ReadInt8(data, offset3+0)
			l.Printf("unknown_a_8: %d \n", unknownA8)

			hpa := huffman.ReadParams(data, offset3+1)
			l.Printf("huffman_params_a: %v\n", hpa)
			ebta := huffman.CreateTable(hpa)
			l.Printf("-> eb_table_a (%d)", ebta.Length())

			hpb := huffman.ReadParams(data, offset3+15)
			l.Printf("huffman_params_b: %v\n", hpb)
			ebtb := huffman.CreateTable(hpb)
			l.Printf("-> eb_table_b (%d)", ebtb.Length())

			l.Printf("unknown_j_128_32_0: %d \n", bin.ReadInt32(data, offset3+29+0))
			l.Printf("unknown_j_128_32_1: %d \n", bin.ReadInt32(data, offset3+29+4))
			l.Printf("  multiplied by 3 sometimes (and then by 32 for buffer)")
			l.Printf("unknown_j_128_32_2: %d \n", bin.ReadInt32(data, offset3+29+8))
			l.Printf("  multiplied by 4 sometimes")
			dataOffset := int(bin.ReadInt32(data, offset3+29+12))
			l.Printf("dataOffset: %d (unknown_j_128_32_3) -> *\n", dataOffset)
			l.Printf("unknown_k_32: %d \n", bin.ReadUInt32(data, offset3+45))

			l.Println()

			if bin.ReadInt32(data, offset3+29+8) == 0 && unknownA8 == 6 {
				panic("??? 1")
			}
			if unknownA8 == 8 {
				panic("??? 2")
			}

			bufs := read10MeshBufs(data, dataOffset, ebta, ebtb)

			l.Printf("buf 0:")
			b0 := bufs[0]
			i32_0 := bin.ReadInt32(b0, 0)
			f64_0 := bin.ReadFloat64(b0, 4)
			f64_1 := bin.ReadFloat64(b0, 12)
			f64_2 := bin.ReadFloat64(b0, 20)
			f32_0 := bin.ReadFloat32(b0, 28)
			f32_1 := bin.ReadFloat32(b0, 32)
			f32_2 := bin.ReadFloat32(b0, 36)
			i8_0 := bin.ReadUInt8(b0, 40)
			res3 := bin.ReadInt32(b0, 41)
			i32_1 := bin.ReadInt32(b0, 45)
			i32_2 := bin.ReadInt32(b0, 49)
			i8_1 := bin.ReadUInt8(b0, 53)
			i32_3 := bin.ReadInt32(b0, 54)
			i32_4 := bin.ReadUInt32(b0, 58)

			l.Printf("  i32_0:    %d", i32_0)
			l.Printf("  f64_0-2:  %f %f %f", f64_0, f64_1, f64_2)
			l.Printf("  f32_0-2:  %f %f %f", f32_0, f32_1, f32_2)
			l.Printf("  i8_0:     %d", i8_0)
			l.Printf("  res3:     %d", res3)
			l.Printf("  i32_1:    %d", i32_1)
			l.Printf("  i32_2:    %d", i32_2)
			l.Printf("  i8_1:     %d", i8_1)
			l.Printf("  i32_3:    %d", i32_3)
			l.Printf("  i32_4:    %d", i32_4)

			if i32_0 < 0 || 0 == i8_0 || (i32_1|i32_2) < 0 || 0 == i8_1 || (i32_4&0x80000000) != 0 {
				panic("incorrect values in buf 0")
			}

			l.Printf("buf 5:")
			b5 := bufs[5]
			res9 := bin.ReadInt32(b5, 0)
			l.Printf("  res9: %d", res9)
			if res9 < 0 {
				panic("incorrect values in buf 5 #1")
			}
			i32_0min32 := i32_0 - 32
			fst := i32_0min32
			snd := 32
			buf_res9vmul3mul4_a := make([]int32, res9*3)
			for i := 0; i < len(buf_res9vmul3mul4_a); i++ {
				buf_res9vmul3mul4_a[i] = -1 // 0xffffffff
			}
			if i32_0min32 >= 128 {
				for {
					val_in_data5_a := bin.ReadInt32(b5, snd/8)
					val_in_data5_b := bin.ReadInt32(b5, snd/8+4)
					if val_in_data5_a >= 0 {
						buf_res9vmul3mul4_a[val_in_data5_a] = val_in_data5_b
					}
					if val_in_data5_b >= 0 {
						buf_res9vmul3mul4_a[val_in_data5_b] = val_in_data5_a
					}

					fst -= 64
					snd += 64

					if !(fst > 127) {
						break
					}
				}
			}
			l.Printf("  buf_res9vmul3mul4_a len: %d", len(buf_res9vmul3mul4_a))

			res1 := bin.ReadInt32(b5, snd/8)
			l.Printf("  res1: %d", res1)
			b5unkn32 := bin.ReadInt32(b5, snd/8+4)
			l.Printf("  b5unkn32: %d", b5unkn32)

			if (res1 | b5unkn32) < 0 {
				panic("incorrect values in buf 5 #2")
			}

			l.Println("buf 2:")
			_, bufCLERS := decodeCLERS(bufs[2], res9, b5unkn32, buf_res9vmul3mul4_a)
			l.Printf("  CLERS: %s", oth.AbbrStr(fmt.Sprintf("%s", bufCLERS), 48))

			// can't skip yet
			os.Exit(0)

		default:
			panic(fmt.Sprintf("Unsupported meshType %d", meshType))
		}

		break
	}

	// can't skip yet
	os.Exit(0)
}

func decodeCLERS(b2 []byte, res9 int32, b5unkn32 int32, buf_res9vmul3mul4_a []int32) ([]int, []byte) {

	bufMeta := make([]int, res9)
	bufCLERS := make([]byte, res9*3)
	writeBufOff := 0
	if b5unkn32 == 0 {
		writeBufOff = 0
		if res9 > 0 {
			writeBufOff = 1
			bufCLERS[0] = 'P'
		}
	}

	if writeBufOff >= int(res9) {
		panic("not implemented: no decoding of data2")
	}
	var input uint64
	rs := 0
	bmcTmp := 0
	updown := 0
	var code uint64
	readBufOff := 0
	bufMetaCtr := bmcTmp

BIG_LOOP:
	for {
		triCtr := 3 * writeBufOff
		wboTmp := writeBufOff
		othCtr := 0
		readShift := rs
		for {
			if readShift <= 0 {
				input |= uint64(bin.ReadUInt32BE(b2, readBufOff)) << uint(32-readShift)
				readShift += 32
				readBufOff += 4
			}
			rs = readShift - 1
			outVal := 'C'
			tmp := input & 0x8000000000000000
			flag := 0
			if tmp != 0 {
				flag = 1
			}
			input *= 2
			if 0 != flag {
				if readShift <= 2 {
					input |= uint64(bin.ReadUInt32BE(b2, readBufOff)) << uint(33-readShift)
					readBufOff += 4
					rs = readShift + 31
				}
				code = input >> 62
				rs -= 2
				input *= 4
				if 0 == uint32(code) {
					break
				}
				if uint32(code) == 3 {
					writeBufOff += othCtr + 1
					bufCLERS[wboTmp+othCtr] = 'E'
					if updown > 0 {
						updown--
						if writeBufOff < int(res9) {
							continue BIG_LOOP
						}
						break BIG_LOOP
					}
					bmcTmp = bufMetaCtr + 1
					if writeBufOff < int(res9) {
						if bmcTmp >= int(b5unkn32) {
							bufCLERS[wboTmp+1+othCtr] = 'P'
							writeBufOff = wboTmp + othCtr + 2
						} else {
							bufMeta[bufMetaCtr+1] = writeBufOff
						}
					}
					if writeBufOff >= int(res9) {
						bufMetaCtr++
						break BIG_LOOP
					}
					bufMetaCtr = bmcTmp
					continue BIG_LOOP
				}
				outVal = 'L'
				if uint32(code) == 1 {
					outVal = 'R'
				}
			}
			bufCLERS[writeBufOff+othCtr] = byte(outVal)
			othCtr++
			triCtr += 3
			readShift = rs
			if othCtr+writeBufOff >= int(res9) {
				writeBufOff += othCtr
				break BIG_LOOP
			}
		}
		bufCLERS[writeBufOff+othCtr] = 'S'

		idx := triCtr + 2 - align3(triCtr+2) + align3(triCtr)

		if buf_res9vmul3mul4_a[idx] == -1 {
			updown++
		}
		writeBufOff += othCtr + 1

		if writeBufOff < int(res9) {
			continue
		}
		break
	}

	return bufMeta, bufCLERS
}

func align3(input int) int {
	return 3 * (input / 3)
}

func read10MeshBufs(data []byte, dataOffset int, ebta huffman.Table, ebtb huffman.Table) (bufs [10][]byte) {
	l.Println("* buf  type  len1   len2   data")
	off := 120
	for i := 0; i < 10; i++ {
		len1 := int(bin.ReadUInt32(data, dataOffset+12*i))
		len2 := int(bin.ReadUInt32(data, dataOffset+12*i+4))
		val := bin.ReadUInt32(data, dataOffset+12*i+8)

		outBuf := make([]byte, len1+3)
		switch val {
		case 0:
			buf := data[dataOffset+off : dataOffset+off+int(len2)]
			l.Printf("  %d    0     %-5d  %-5d  %s", i, len1, len2, oth.AbbrHexStr(buf, 32))
			copy(outBuf, buf)
		case 3:
			buf := data[dataOffset+off : dataOffset+off+int(len2)]
			l.Printf("  %d    3     %-5d  %-5d  %s", i, len1, len2, oth.AbbrHexStr(buf, 32))
			hp, s := ebta, "a"
			if i == 7 {
				hp, s = ebtb, "b"
			}
			huffman.DecodeUsingTable(buf, len1, len2, hp, &outBuf)
			l.Printf("  -> decoded (eb_table_%s): %s", s, oth.AbbrHexStr(outBuf, 32))
		case 1:
			panic("Unsupported type: 1")
		default:
			panic(fmt.Sprintf("Unknown type: %d", val))
		}
		bufs[i] = outBuf

		off += len2
	}
	return
}
