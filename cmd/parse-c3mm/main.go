package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

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

	file := os.Args[1]
	data, err := ioutil.ReadFile(file)
	oth.CheckPanic(err)
	l.Printf("File size: %d bytes\n", len(data))
	parseC3MM(data)
}

func parseC3MM(data []byte) {
	if "C3MM" != string(data[0:4]) {
		panic("invalid C3MM header")
	}
	version := bin.ReadInt16(data, 4)
	switch version {
	case 1:
		parseC3MMv1(data, 0 /* todo */)
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
	compressedSize := int(bin.ReadInt32(data, 19))
	l.Println("compressed size", compressedSize)
	uncompressedSize := int(bin.ReadInt32(data, 23))
	l.Println("uncompressed size", uncompressedSize)
	body := data[27:]
	if uncompressedSize != compressedSize {
		body = decompressLZMA(body, 0, len(body), uncompressedSize)
	}
	l.Println("body size", len(body))
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
