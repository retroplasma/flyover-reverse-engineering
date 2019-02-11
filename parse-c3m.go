package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
)

var l = log.New(os.Stderr, "", 0)

func main() {

	/*cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()*/

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [c3m_file]\n", os.Args[0])
		os.Exit(1)
	}

	file := os.Args[1]
	data, err := ioutil.ReadFile(file)
	checkPanic(err)

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
			l.Printf("- Mesh at 0x%x (%d)", offset, offset)
			l.SetPrefix(l.Prefix() + "  ")
			panic("Type 2 not implemented")
		case 3:
			l.Printf("Scene Graph? / Animation? at 0x%x (%d)", offset, offset)
			l.SetPrefix(l.Prefix() + "  ")
			panic("Type 3 not implemented")
		default:
			panic("Invalid item type")
		}

		if processedItems >= numberOfItems {
			l.Println("All items processed")
			return
		}
		processedItems++
	}
}

func parseHeader(data []byte, offset *int) {

	qx := readDouble(data, *offset+9)
	qy := readDouble(data, *offset+17)
	qz := readDouble(data, *offset+25)
	qw := readDouble(data, *offset+33)
	l.Printf("Rotation Quaternion XYZW:     [% 16f,% 16f,% 15f,% f ]\n", qx, qy, qz, qw)

	x := readDouble(data, *offset+41)
	y := readDouble(data, *offset+49)
	z := readDouble(data, *offset+57)
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
	panic("Type 1 not implemented")
}

func readDouble(bytes []byte, offset int) float64 {
	bits := binary.LittleEndian.Uint64(bytes[offset : offset+8])
	float := math.Float64frombits(bits)
	return float
}

func checkPanic(e error) {
	if e != nil {
		panic(e)
	}
}
