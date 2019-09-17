package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/bin"
	"github.com/ulikunitz/xz/lzma"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/oth"
)

var l = log.New(os.Stderr, "", 0)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [c3mm_file]\n", os.Args[0])
		os.Exit(1)
	}

	partIfv1 := 0
	if len(os.Args) == 3 {
		var err error
		if partIfv1, err = strconv.Atoi(os.Args[2]); err != nil {
			panic(err)
		}
	}
	l.Println(partIfv1)

	file := os.Args[1]
	data, err := ioutil.ReadFile(file)
	oth.CheckPanic(err)
	l.Printf("File size: %d bytes\n", len(data))
	parseC3MM(data, partIfv1)
}

func parseC3MM(data []byte, partIfv1 int) {
	if "C3MM" != string(data[0:4]) {
		panic("invalid C3MM header")
	}
	version := bin.ReadInt16(data, 4)
	switch version {
	case 1:
		parseC3MMv1(data, partIfv1)
	case 2:
		parseC3MMv2(data)
	default:
		panic("Unknown C3MM version")
	}
}

func parseC3MMv1(data []byte, part int) {
	if "C3MM\x01\x00" != string(data[0:6]) {
		panic("invalid C3MM v1 header")
	}
	l.Println("unkn6", bin.ReadInt32(data, 6))
	l.Println("file type", data[10])
	mult1, mult2 := bin.ReadFloat32(data, 11), bin.ReadFloat32(data, 15)
	l.Println("mult1", mult2)
	l.Println("mult2", mult2)
	compressedSize := int(bin.ReadInt32(data, 19))
	l.Println("compressed size", compressedSize)
	uncompressedSize := int(bin.ReadInt32(data, 23))
	l.Println("uncompressed size", uncompressedSize)
	body := data[27:]
	if uncompressedSize != compressedSize {
		body = decompressLZMA(body, 0, len(body), uncompressedSize)
	}
	l.Println("body size", len(body))

	if part == 0 {
		l.Print("part 0:")
		offset := 0

		// parts body concat offsets?
		{
			if 2 != body[offset] {
				panic("t must be 2")
			}
			size := int(bin.ReadInt32(body, offset+1))
			seg := body[offset+1+4 : offset+size]

			l.Println(2, size, len(seg))
			for i := 0; i < len(seg); i += 4 {
				cur := int(bin.ReadInt32(seg, i))
				l.Println("  ", i/4, cur, (cur+1)%3 == 0, (cur-5)%9 == 0)
			}

			lastPartNumber := len(seg)/4 - 1
			l.Println("last part number:", lastPartNumber)

			offset += int(size)
		}

		// part tree info or something
		// one less than total parts
		{
			if 0 != body[offset] {
				panic("t must be 0")
			}
			size := int(bin.ReadInt32(body, offset+1))
			seg := body[offset+1+4 : offset+size]
			l.Println(0, size, len(seg))

			for i := 0; i < len(seg); i += 17 {
				l.Println("  ", i/17, "->", bin.ReadInt32(seg, i),
					bin.ReadInt32(seg, i+4),
					bin.ReadInt32(seg, i+4+4),
					bin.ReadInt32(seg, i+4+4+4),
					seg[0+4+4+4+4], "structure type (< 2)")
			}

			offset += int(size)
		}

		// part data
		{
			if 1 != body[offset] {
				panic("t must be 1")
			}
			sizeUnknPartContinuation := int(bin.ReadInt32(body, offset+1))
			actualSize := len(body) - (offset + 1 + 4)
			seg := body[offset+1+4 : offset+1+4+actualSize]
			l.Println(1, sizeUnknPartContinuation, len(seg), len(seg)%9 == 0)
			parsePartSegment(seg, mult1, mult2)
		}

	} else {
		parsePartSegment(body, mult1, mult2)
	}
}

func parsePartSegment(seg []byte, mult1, mult2 float32) {
	for i := 0; i < len(seg); i += 9 {
		s := seg[i : i+9]
		input0, input2, input3, input5 := bin.ReadInt16(s, 0), bin.ReadUInt8(s, 2), bin.ReadInt16(s, 3), bin.ReadInt32(s, 5)
		m1 := float32(input3) * mult1
		m2 := m1 + float32(input2)*mult2
		_ = m2
		l.Println("  ", hex.EncodeToString(s), input0, input2, input3, input5)
	}
}

func parseC3MMv2(data []byte) {
	if "C3MM\x02\x00" != string(data[0:6]) {
		panic("invalid C3MM v2 header")
	}
	count := int(bin.ReadInt32(data, 8))
	l.Println("count:", count, "| unkn6:", bin.ReadInt16(data, 6))
	offset := 12
	for i := 0; i < count; i++ {
		t := bin.ReadInt8(data, offset)
		totalSize := int(bin.ReadInt32(data, offset+1))
		isCompressed := 1 == bin.ReadInt8(data, offset+5)
		bodySize := int(bin.ReadInt32(data, offset+6))
		l.Printf("type: %d | total: %-5d | compressed: %-5v | body: %d", t, totalSize, isCompressed, bodySize)
		body := data[offset+10 : offset+10+totalSize-10]
		if isCompressed {
			body = decompressLZMA(data, offset+10, totalSize-10, bodySize)
		}
		switch t {
		case 4:
			type metaLevel struct {
				val1, val2 byte
			}
			ml := make([]metaLevel, int(body[0]))
			for i := range ml {
				ml[i].val1, ml[i].val2 = body[1+2*i+0], body[1+2*i+1]
			}
			l.Println("  meta levels", ml)
		default:
			l.Println("  parser not implemented")
		}
		offset += totalSize
	}
}

func decompressLZMA(data []byte, offset, compressedSize, decompressedSize int) []byte {
	d := make([]byte, compressedSize)
	copy(d, data[offset:offset+compressedSize])
	bin.WriteInt64(d, 5, int64(decompressedSize))
	r, err := lzma.NewReader(io.Reader(bytes.NewReader(d)))
	if err != nil {
		log.Fatalf("reader error: %v", err)
	}
	var out bytes.Buffer
	if _, err = io.Copy(bufio.NewWriter(&out), r); err != nil {
		log.Fatalf("copy error: %v", err)
	}
	return out.Bytes()
}
