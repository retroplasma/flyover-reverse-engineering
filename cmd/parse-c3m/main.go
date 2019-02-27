package main

import (
	"flyover-reverse-engineering/pkg/bin"
	"flyover-reverse-engineering/pkg/dec/huffman"
	"flyover-reverse-engineering/pkg/dec/mesh"
	"flyover-reverse-engineering/pkg/mth"
	"flyover-reverse-engineering/pkg/oth"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	offset := 6

	pfx := l.Prefix()
	for processedItems := 0; processedItems < numberOfItems; processedItems++ {
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
	}
	l.SetPrefix(pfx)
	l.Println("All items processed")
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

	l.Printf("Scale: vtx?")

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
	*offset += 5
	numberOfItems := int(bin.ReadInt32(data, *offset+0))
	l.Printf("Number of meshes: %d \n", numberOfItems)
	//processedItems := 0
	*offset += 4

	pfx := l.Prefix()
	l.SetPrefix(l.Prefix() + "  ")

	var vtxCount int32
	for currentItem := 0; currentItem < numberOfItems; currentItem++ {
		l.SetPrefix(pfx)
		meshType := bin.ReadInt8(data, *offset+0)
		unknown_1_2 := int(bin.ReadInt8(data, *offset+1)) + int(bin.ReadInt8(data, *offset+2))<<8
		l.Printf("unknown_1_2: %d + %d<<8 = %d", bin.ReadInt8(data, *offset+1), bin.ReadInt8(data, *offset+2), unknown_1_2)
		l.Println()

		switch meshType {
		case 2:
			offset3 := *offset + 3

			l.Printf("- Mesh type: %d \n", meshType)
			l.SetPrefix(l.Prefix() + "    ")

			unknownA8 := bin.ReadInt8(data, offset3+0)
			l.Printf("unknown_a_8: %d \n", unknownA8)

			hpa := huffman.ReadParams(data, offset3+1)
			ebta := huffman.CreateTable(hpa)
			l.Printf("huffman_params_a: %v -> eb_table_a (%d)\n", hpa, ebta.Length())

			hpb := huffman.ReadParams(data, offset3+15)
			ebtb := huffman.CreateTable(hpb)
			l.Printf("huffman_params_b: %v -> eb_table_b (%d)\n", hpb, ebtb.Length())

			gUvCount := bin.ReadInt32(data, offset3+29+0)
			gFacesCount := bin.ReadInt32(data, offset3+29+4)
			groupCount := bin.ReadInt32(data, offset3+29+8)
			l.Printf("tex coords: %d (unknown_j_128_32_0)\n", gUvCount)
			l.Printf("faces: %d (unknown_j_128_32_1)\n", gFacesCount)
			l.Printf("groups: %d (unknown_j_128_32_2)\n", groupCount)
			dataOffset := int(bin.ReadInt32(data, offset3+29+12))
			l.Printf("dataOffset: %d (unknown_j_128_32_3) -> *\n", dataOffset)
			l.Printf("unknown_k_32: %d \n", bin.ReadUInt32(data, offset3+45))

			l.Println()

			if groupCount == 0 && unknownA8 == 6 {
				panic("??? 1")
			}
			if unknownA8 == 8 {
				panic("??? 2")
			}

			mesh.SetLogPrefix(l.Prefix() + "  ")
			l.Println("Decompressing")
			rmd := mesh.Decompress(data, dataOffset, ebta, ebtb)
			if rmd.UVCount != gUvCount || rmd.FacesCount != gFacesCount {
				panic("decompressed mesh counts != header counts")
			}

			l.Println("Vertices:", len(rmd.Vertices)/3, "| Faces:", len(rmd.Faces)/3, "| UV:", len(rmd.UV)/2)

			tmpBufFst := make([]int32, rmd.UVCount)
			for ctr := 0; ctr < 3*int(rmd.FacesCount); ctr++ {
				tmpBufFst[rmd.Res5[ctr]] = rmd.Faces[ctr]
			}
			tmpBufSnd := make([]int32, rmd.UVCount)

			preIdx := 0
			off := 0
			uvCount1 := int(rmd.UVCount)
			uvCount2 := int(rmd.UVCount)
			vertices := make([]vertex, uvCount2)

			for {
				tmpBufFstItm := tmpBufFst[off]
				uvCountMin1 := uvCount1 - 1
				if 0 != rmd.Res8[tmpBufFstItm] {
					uvCount1 = uvCountMin1
				} else {
					uvCountMin1 = preIdx
					preIdx++
				}
				tmpBufSnd[off] = int32(uvCountMin1)
				idx := uvCountMin1
				vertices[idx].x = rmd.Vertices[3*tmpBufFstItm+0]
				vertices[idx].y = rmd.Vertices[3*tmpBufFstItm+1]
				vertices[idx].z = rmd.Vertices[3*tmpBufFstItm+2]
				vertices[idx].u = rmd.UV[off*2+0]
				vertices[idx].v = rmd.UV[off*2+1]

				off++
				uvCount2--
				if uvCount2 == 0 {
					break
				}
			}

			for ctr := 0; ctr < 3*int(rmd.FacesCount); ctr++ {
				rmd.Res5[ctr] = tmpBufSnd[rmd.Res5[ctr]]
			}

			gm := make(map[int]int)
			for i := 0; i < len(rmd.Res6); i++ {
				gm[int(rmd.Res6[i])]++
			}
			l.Println("Groups:", gm)

			groups := make(map[int]group)
			for i := 0; i < len(rmd.Res6); i++ {
				e := int(rmd.Res6[i])
				if gm[e] > 0 {
					group := groups[e]
					if group.faces == nil {
						group.material = e
						group.faces = make([]face, gm[e])
						groups[e] = group
					}
					face := &group.faces[len(group.faces)-gm[e]]
					face.a, face.b, face.c = rmd.Res5[i*3], rmd.Res5[i*3+1], rmd.Res5[i*3+2]
					gm[e]--
				}
			}

			vtxCount += int32(len(vertices))

			*offset += unknown_1_2
		default:
			panic(fmt.Sprintf("Unsupported meshType %d", meshType))
		}
	}
}

type vertex struct {
	x, y, z, u, v float32
}

type group struct {
	material int
	faces    []face
}

type face struct {
	a, b, c int32
}
