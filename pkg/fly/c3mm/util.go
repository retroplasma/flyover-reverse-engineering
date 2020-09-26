package c3mm

import "github.com/retroplasma/flyover-reverse-engineering/pkg/bin"

// GetPartNumber finds the part number for octant data access
func (fi FileIndex) GetPartNumber(octantOffset int) int {
	partNum := len(fi.Entries) - 1
	for i := 0; i < len(fi.Entries)-1; i++ {
		if octantOffset < fi.Entries[i+1] {
			partNum = i
			break
		}
	}
	return partNum
}

// GetOctant retrieves an octant by the given offsets and increases octantOffset
func (c C3MM) GetOctant(octantOffset *int, partOffset int) Octant {
	offset := *octantOffset - partOffset
	*octantOffset += 9
	if offset%9 != 0 {
		panic("offset%9 != 0")
	}
	s := c.DataSection.Raw[offset:]
	if len(s)%9 != 0 {
		panic("len(s)%9 != 0")
	}
	if len(s) == 0 {
		panic("not sure if this can happen") // if yes, make octant optional
	}
	valueA := bin.ReadInt16(s, 0)
	valueB := bin.ReadUInt8(s, 2)
	valueC := bin.ReadInt16(s, 3)
	valueD := bin.ReadInt32(s, 5)
	valueCm := float32(valueC) * c.Header.Mult1
	valueBm := (float32(valueB) * c.Header.Mult2) + valueCm
	return Octant{Bits: valueA, AltitudeHigh: valueBm, AltitudeLow: valueCm, Next: int(valueD)}
}

// ZoomedOut returns a tile with decreased Z and halved XYH
func (t Tile) ZoomedOut() Tile {
	return Tile{Z: t.Z - 1, Y: t.Y / 2, X: t.X / 2, H: t.H / 2}
}

// ZoomedInCandidates returns a function which returns sub tiles by octant number o (0-7)
func (t Tile) ZoomedInCandidates() func(o int) Tile {
	tn := Tile{Z: t.Z + 1, Y: t.Y * 2, X: t.X * 2, H: t.H * 2}
	return func(o int) Tile {
		return Tile{
			Z: tn.Z,
			Y: tn.Y | (o>>1)&1,
			X: tn.X | (o>>0)&1,
			H: tn.H | (o>>2)&1,
		}
	}
}
