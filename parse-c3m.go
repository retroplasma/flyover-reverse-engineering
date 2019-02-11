package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
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
	checkPanic(err)

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
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
			l.Printf("Not implemented, can't skip yet")
			os.Exit(1)
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

	qx := readFloat64(data, *offset+9)
	qy := readFloat64(data, *offset+17)
	qz := readFloat64(data, *offset+25)
	qw := readFloat64(data, *offset+33)
	l.Printf("Rotation Quaternion XYZW:     [% 16f,% 16f,% 15f,% f ]\n", qx, qy, qz, qw)

	x := readFloat64(data, *offset+41)
	y := readFloat64(data, *offset+49)
	z := readFloat64(data, *offset+57)
	l.Printf("Translation ECEF XYZ:         [% 16f,% 16f,% 15f ]\n", x, y, z)

	m0 := 1 - 2*qy*qy - 2*qz*qz
	m1 := 2*qx*qy + 2*qw*qz
	m2 := 2*qx*qz - 2*qw*qy
	m3 := 2*qx*qy - 2*qw*qz
	m4 := 1 - 2*qx*qx - 2*qz*qz
	m5 := 2*qy*qz + 2*qw*qx
	m6 := 2*qx*qz + 2*qw*qy
	m7 := 2*qy*qz - 2*qw*qx
	m8 := 1 - 2*qx*qx - 2*qy*qy
	l.Printf("=> Transformation Matrix 4x4: [% 16f,% 16f,% 15f,% f,\n", m0, m1, m2, 0.0)
	l.Printf("                               % 16f,% 16f,% 15f,% f,\n", m3, m4, m5, 0.0)
	l.Printf("                               % 16f,% 16f,% 15f,% f,\n", m6, m7, m8, 0.0)
	l.Printf("                               % 16f,% 16f,% 15f,% f ]\n", x, y, z, 1.0)

	*offset += 113
}

func parseMaterial(data []byte, offset *int) {
	*offset += 5
	numberOfItems := int(readInt32(data, *offset+0))
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
			l.Printf("- Material type %d", materialType)
			l.SetPrefix(l.Prefix() + "  ")

			textureFormat := data[*offset+3]
			textureOffset := readInt32(data, *offset+4)
			textureLength := readInt32(data, *offset+8)
			textureLength2 := readInt32(data, *offset+12)

			l.SetPrefix(l.Prefix() + "  ")
			l.Printf("texOff: %d, texLen1: %d, texLen2: %d", textureOffset, textureLength, textureLength2)
			switch textureFormat {
			case 0:
				l.Printf("Format: JPEG")

				if out {
					fn := fmt.Sprintf("/tmp/jpg/%d.jpg", processedItems)
					err := ioutil.WriteFile(fn, data[textureOffset:textureOffset+textureLength2], 0655)
					checkPanic(err)
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

/*
 * binary readers
 */

func readFloat32(data []byte, offset int) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(data[offset:]))
}

func readFloat64(data []byte, offset int) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
}

func readInt32(data []byte, offset int) int32 {
	return int32(binary.LittleEndian.Uint32(data[offset:]))
}

func readUInt32(data []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(data[offset:])
}

func readInt64(data []byte, offset int) int64 {
	return int64(binary.LittleEndian.Uint64(data[offset:]))
}

func readUInt64(data []byte, offset int) uint64 {
	return binary.LittleEndian.Uint64(data[offset:])
}

/*
 * other
 */

func checkPanic(e error) {
	if e != nil {
		panic(e)
	}
}
