package internal

import (
	"bufio"
	"bytes"
	"io"
	"log"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/bin"
	"github.com/ulikunitz/xz/lzma"
)

func DecompressLZMA(data []byte, offset, compressedSize, decompressedSize int) []byte {
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
