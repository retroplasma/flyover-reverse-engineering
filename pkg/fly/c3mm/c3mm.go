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
		body = internal.DecompressLZMA(body, 0, len(body), c3mm.Header.UncompressedSize)
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
					Z:             int(bin.ReadInt32(seg, i)),
					Y:             int(bin.ReadInt32(seg, i+4)), // inverted
					X:             int(bin.ReadInt32(seg, i+4+4)),
					H:             0,
					Shift:         int(bin.ReadInt32(seg, i+4+4+4)),
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
				return rootLess(entries[i], entries[j])
			})

			offset += int(size)
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

func (c C3MM) GetOctant(rootShift *int, partShift int) Octant {
	offset := *rootShift - partShift
	*rootShift += 9
	if offset%9 != 0 {
		panic("offset%9 != 0")
	}
	s := c.DataSection.Raw[offset:]
	if len(s)%9 != 0 {
		panic("len(s)%9 != 0")
	}
	if len(s) == 0 {
		panic("not sure if this can happen")
		// if yes, make octant optional
	}
	//l.Println(c.Header.Mult1, c.Header.Mult2)
	valueA := bin.ReadInt16(s, 0)
	valueB := bin.ReadUInt8(s, 2)
	valueC := bin.ReadInt16(s, 3)
	valueD := bin.ReadInt32(s, 5)
	valueCm := float32(valueC) * c.Header.Mult1
	valueBm := (float32(valueB) * c.Header.Mult2) + valueCm
	return Octant{Bits: valueA, Unkn1: valueBm, Unkn2: valueCm, Next: int(valueD)}
}

type Octant struct {
	Bits  int16
	Unkn1 float32
	Unkn2 float32
	Next  int
}

type FileIndex struct {
	Entries []int
}

type RootIndex struct {
	Entries []Root
}

type DataSection struct {
	Raw []byte
}

type Root struct {
	Z             int
	Y             int
	X             int
	H             int
	Shift         int
	StructureType int
}

type Header struct {
	Unkn6            int
	FileType         uint8
	Mult1            float32
	Mult2            float32
	CompressedSize   int
	UncompressedSize int
}

func rootLess(a Root, b Root) bool {
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
