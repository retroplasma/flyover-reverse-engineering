package c3mm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"sort"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/bin"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly/c3mm/internal"
)

var l = log.New(os.Stderr, "", 0)

func DisableLogs() {
	l.SetFlags(0)
	l.SetOutput(ioutil.Discard)
	//internal.DisableLogs()
}

func Parse(data []byte, partIfv1 int) (result C3MM, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintln(e, string(debug.Stack())))
		}
	}()

	result = parseC3MM(data, partIfv1)
	return
}

func parseC3MM(data []byte, partIfv1 int) C3MM {
	if "C3MM" != string(data[0:4]) {
		panic("invalid C3MM header")
	}
	version := bin.ReadInt16(data, 4)
	switch version {
	case 1:
		return parseC3MMv1(data, partIfv1)
	case 2:
		panic("C3MM v2 not implemented")
	default:
		panic("Unknown C3MM version")
	}
}

func parseC3MMv1(data []byte, part int) (c3mm C3MM) {
	if "C3MM\x01\x00" != string(data[0:6]) {
		panic("invalid C3MM v1 header")
	}

	c3mm.Header.Unkn6 = int(bin.ReadInt32(data, 6))
	c3mm.Header.FileType = data[10]
	c3mm.Header.Mult1 = bin.ReadFloat32(data, 11)
	c3mm.Header.Mult2 = bin.ReadFloat32(data, 15)
	c3mm.Header.CompressedSize = int(bin.ReadInt32(data, 19))
	c3mm.Header.UncompressedSize = int(bin.ReadInt32(data, 23))

	body := data[27:]
	if c3mm.Header.UncompressedSize != c3mm.Header.CompressedSize {
		var err error
		body, err = internal.DecompressLZMA(body, 0, len(body), c3mm.Header.UncompressedSize)
		if err != nil {
			panic(err)
		}
	}

	offset := 0

	if part == 0 {
		c3mm.FileIndex = &FileIndex{}
		c3mm.RootIndex = &RootIndex{}

		// file index
		{
			if 2 != body[offset] {
				panic("t must be 2")
			}
			size := int(bin.ReadInt32(body, offset+1))
			seg := body[offset+1+4 : offset+size]

			for i := 0; i < len(seg); i += 4 {
				cur := int(bin.ReadInt32(seg, i))
				c3mm.FileIndex.Entries = append(c3mm.FileIndex.Entries, cur)
			}

			offset += int(size)
		}

		// root index
		{
			if 0 != body[offset] {
				panic("t must be 0")
			}
			size := int(bin.ReadInt32(body, offset+1))
			seg := body[offset+1+4 : offset+size]

			for i := 0; i < len(seg); i += 17 {
				root := Root{
					Tile: Tile{
						Z: int(bin.ReadInt32(seg, i)),
						Y: int(bin.ReadInt32(seg, i+4)), // inverted
						X: int(bin.ReadInt32(seg, i+4+4)),
						H: 0,
					},
					Offset:        int(bin.ReadInt32(seg, i+4+4+4)),
					StructureType: int(seg[i+4+4+4+4]),
				}
				if !(root.StructureType < 2) {
					panic("structure type not < 2")
				}
				if root.StructureType != 1 {
					panic("not implemented: structure type != 1")
				}
				c3mm.RootIndex.Entries = append(c3mm.RootIndex.Entries, root)
			}

			// sort
			sort.Slice(c3mm.RootIndex.Entries, func(i, j int) bool {
				entries := c3mm.RootIndex.Entries
				return entries[i].Tile.Less(entries[j].Tile)
			})

			c3mm.RootIndex.SmallestZ = c3mm.RootIndex.Entries[0].Tile.Z
			for _, root := range c3mm.RootIndex.Entries {
				if root.Tile.Z < c3mm.RootIndex.SmallestZ {
					c3mm.RootIndex.SmallestZ = root.Tile.Z
				}
			}

			offset += int(size)
		}

		// skipping object tree
		{
			if 3 == body[offset] {
				skip := int(bin.ReadInt32(body, offset+1))
				offset += skip
			}
		}

		// data section
		{
			if 1 != body[offset] {
				panic("t must be 1")
			}
			_ = int(bin.ReadInt32(body, offset+1)) // -> end of all parts?
			offset += 5
		}
	}

	c3mm.DataSection.Raw = body[offset:]
	return
}

type C3MM struct {
	Header      Header
	FileIndex   *FileIndex
	RootIndex   *RootIndex
	DataSection DataSection
}

type Header struct {
	Unkn6            int
	FileType         uint8
	Mult1            float32
	Mult2            float32
	CompressedSize   int
	UncompressedSize int
}

type FileIndex struct {
	Entries []int
}

type RootIndex struct {
	SmallestZ int
	Entries   []Root
}

type Root struct {
	Tile          Tile
	Offset        int
	StructureType int
}

type DataSection struct {
	Raw []byte
}

type Octant struct {
	Bits         int16
	AltitudeHigh float32
	AltitudeLow  float32
	Next         int
}

type Tile struct {
	Z int
	Y int
	X int
	H int
}

func (t Tile) Less(b Tile) bool {
	a := t
	if a.Z < b.Z {
		return true
	}
	if a.Z > b.Z {
		return false
	}
	if a.Y < b.Y {
		return true
	}
	if a.Y > b.Y {
		return false
	}
	if a.X < b.X {
		return true
	}
	if a.X > b.X {
		return false
	}
	return a.H < b.H
}
