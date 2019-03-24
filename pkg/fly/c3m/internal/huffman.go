package internal

import (
	"bytes"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/bin"
)

func (table HuffmanTable) Decode(data []byte, len1 int, len2 int, writeBuf *[]byte) {
	readBuf := make([]byte, len2+3)
	copy(readBuf, data)
	if len1 < 2 {
		return
	}

	len2mul8 := 8 * len2
	len1div2 := len1 / 2
	tblFst := table.data[0]
	readShift1 := 0
	var input1 uint64
	readBufOff := 0
	writeBufOff := 0
	for {
		if readShift1 <= 0 {
			input1 |= uint64(bin.ReadUInt32BE(readBuf, readBufOff)) << uint(32-readShift1)
			readShift1 += 32
			readBufOff += 4
		}
		negToggle := input1 >> 63
		readShift2 := readShift1 - 1
		input2 := 2 * input1
		shiftTest := len2mul8 - (8*readBufOff - (readShift1 - 1))
		var tblFstIdx uint64
		if shiftTest > 15 {
			if readShift1 <= 16 {
				input2 |= uint64(bin.ReadUInt32BE(readBuf, readBufOff)) << uint(33-readShift1)
				readBufOff += 4
				readShift2 = readShift1 + 31
			}
			tblFstIdx = input2 >> 48
		} else {
			if readShift1 <= int(shiftTest) {
				l.Println("not visited. check values when visited #1")
				input2 |= uint64(bin.ReadUInt32BE(readBuf, readBufOff)) << uint(33-readShift1)
				readBufOff += 4
				readShift2 = readShift1 + 31
			}
			tblFstIdx = uint64(input2 >> uint(64-uint8(shiftTest)) << uint(16-shiftTest))
		}
		tblFstVal := int(bin.ReadInt8(tblFst, 8*int(tblFstIdx)+4))
		if tblFstVal <= 0 {
			tblFstValNeg := -tblFstVal
			tblIdx := bin.ReadInt32(tblFst, 8*int(tblFstIdx))
			if readShift2 <= 15 {
				input2 |= uint64(bin.ReadUInt32BE(readBuf, readBufOff)) << uint(32-readShift2)
				readShift2 += 32
				readBufOff += 4
			}
			readShift3 := readShift2 - 16
			input3 := input2 << 16
			if readShift2-16 < tblFstValNeg {
				input3 |= uint64(bin.ReadUInt32BE(readBuf, readBufOff)) << uint(48-readShift2)
				readBufOff += 4
				readShift3 = readShift2 + 16
			}
			tblOthIdx := uint(input3 >> uint(64-tblFstValNeg))
			tblOth := table.data[tblIdx]
			tblOthValNeg := int(tblOth[8*tblOthIdx+4]) - 16
			if readShift3 < tblOthValNeg {
				input3 |= uint64(bin.ReadUInt32BE(readBuf, readBufOff)) << uint(32-readShift3)
				readShift3 += 32
				readBufOff += 4
			}
			readShift1 = readShift3 - tblOthValNeg
			input1 = input3 << uint(tblOthValNeg)
			outVal := -bin.ReadInt32(tblOth, 8*int(tblOthIdx))
			if int32(negToggle) == 0 {
				outVal = bin.ReadInt32(tblOth, 8*int(tblOthIdx))
			}
			bin.WriteInt16(*writeBuf, writeBufOff, int16(outVal))
		} else {
			if readShift2 < tblFstVal {
				l.Println("not visited. check values when visited #3")
				input2 |= uint64(bin.ReadUInt32BE(readBuf, readBufOff)) << uint(32-readShift2)
				readShift2 += 32
				readBufOff += 4
			}
			input1 = input2 << uint(tblFstVal)
			outVal := -bin.ReadInt32(tblFst, 8*int(tblFstIdx))
			if int32(negToggle) == 0 {
				outVal = bin.ReadInt32(tblFst, 8*int(tblFstIdx))
			}
			bin.WriteInt16(*writeBuf, writeBufOff, int16(outVal))
			readShift1 = readShift2 - tblFstVal
		}
		writeBufOff += 2
		len1div2--

		// end
		if len1div2 == 0 {
			break
		}
	}

	// todo check unvisited branches
}

func dbgComp(a, b [][]byte) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b == nil || a == nil && b != nil || len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if x != nil && y == nil || x == nil && y != nil || !bytes.Equal(x, y) {
			return false
		}
	}
	return true
}

func (hp HuffmanParams) CreateTable() HuffmanTable {

	type tree struct {
		index     int16
		unknown4  int32
		unknown8  int32
		child1    *tree
		child2    *tree
		unknown40 int8
	}
	type tree2 struct {
		xUnknown8  int32
		xUnknown40 int8
	}

	bufs := make([]*tree, hp.p3)
	if hp.p3 == 0 {
		panic("not implemented")
	}

	for i, hp1 := 0, hp.p1; i < hp.p3; i, hp1 = i+1, hp1+hp.p0 {
		var buf tree
		buf.index = int16(i)
		buf.unknown4 = int32(0xFFFFFFFF / (hp.p2 + hp1*i))
		bufs[i] = &buf
	}

	for hp3 := hp.p3; hp3 > 1; hp3-- {
		b1 := bufs[hp3-1]
		b2 := bufs[hp3-2]

		var buf1 tree
		buf1.index = int16(-1)
		buf1.unknown4 = int32(b1.unknown4 + b2.unknown4)
		buf1.child1 = b1
		buf1.child2 = b2

		for _hp3 := hp3; ; {
			b := bufs[_hp3-2]
			if buf1.unknown4 <= b.unknown4 {
				bufs[_hp3-1] = &buf1
				break
			}
			bufs[_hp3-1] = b
			_hp3--
			if _hp3-2 == -1 {
				bufs[0] = &buf1
				break
			}
		}
	}

	buf3 := make([]*tree, 20)
	buf2 := make([]*tree2, hp.p3)
	buf3[0] = bufs[0]

	for counter := 1; counter != 0; {
		tree := buf3[counter-1]
		if tree.index < 0 {
			buf3[counter-1] = tree.child1
			buf3[counter-0] = tree.child2
			counter++
		} else {
			var t2 tree2
			t2.xUnknown8 = tree.unknown8
			t2.xUnknown40 = tree.unknown40
			buf2[tree.index] = &t2
			counter--
		}
		if tree.child1 != nil {
			tree.child1.unknown8 = 2 * tree.unknown8
			tree.child1.unknown40 = tree.unknown40 + 1
		}
		if tree.child2 != nil {
			tree.child2.unknown8 = 2*tree.unknown8 + 1
			tree.child2.unknown40 = tree.unknown40 + 1
		}
	}

	buf4 := make([]int8, 0x10001)
	buf5 := make([]int16, 0x20000/2)
	var counter int16 = 1

	for buf2idx, hp3 := 0, hp.p3; hp3 != 0; buf2idx, hp3 = buf2idx+1, hp3-1 {
		b := buf2[buf2idx]
		xu40 := b.xUnknown40
		if xu40 >= 17 {
			xu40m16 := xu40 - 16
			idx5 := b.xUnknown8 >> uint(xu40m16)
			idx4 := buf5[idx5]
			if idx4 == 0 {
				buf5[idx5] = counter
				idx4 = counter
				counter++
			}
			if xu40m16 > buf4[idx4] {
				buf4[idx4] = xu40m16
			}
		}
	}

	buf4[0] = 16
	buf6 := make([][]byte, counter)
	for i, j := 16, 0; j < int(counter); i, j = int(buf4[j+1]), j+1 {
		count := 1 << uint(i)
		buf6SubBuf := make([]byte, 8*count)
		buf6[j] = buf6SubBuf
		for ctr := 0; ctr < count; ctr++ {
			bin.WriteUInt16(buf6SubBuf, 8*ctr, 0xFFFF)
		}
	}

	for buf2idx := 0; buf2idx < hp.p3; buf2idx++ {
		b2xu4 := buf2[buf2idx].xUnknown40
		if b2xu4 > 16 {
			b2xu8mod := buf2[buf2idx].xUnknown8 >> uint(b2xu4-16)
			b5val := buf5[b2xu8mod]
			bin.WriteInt32(buf6[0], 8*int(b2xu8mod), int32(b5val))
			bin.WriteInt32(buf6[0], 8*int(b2xu8mod)+4, int32(-buf4[b5val]))
			b4val := buf4[b5val]
			b2xu4m16 := b2xu4 - 16
			lob := 0xff & buf2[buf2idx].xUnknown8
			subbufPtr := int8(int8(lob&((1<<uint(b2xu4m16))-1)) << uint(b4val-b2xu4m16))
			bin.WriteInt32(buf6[b5val], 8*int(subbufPtr), int32(buf2idx))
			bin.WriteInt32(buf6[b5val], 8*int(subbufPtr)+4, int32(buf2[buf2idx].xUnknown40))
		} else {
			subbufPtr := buf2[buf2idx].xUnknown8 << uint(16-b2xu4)
			bin.WriteInt32(buf6[0], 8*int(subbufPtr), int32(buf2idx))
			bin.WriteInt32(buf6[0], 8*int(subbufPtr)+4, int32(b2xu4))
		}
	}

	for i, j := 16, 0; j < int(counter); i, j = int(buf4[j+1]), j+1 {
		if i != 0 {
			buf6SubBuf := buf6[j]
			buf6SubBufVal := bin.ReadUInt64(buf6SubBuf, 0)
			count := 1 << uint(i)
			for ctr := 1; ctr < count; ctr++ {
				if bin.ReadUInt16(buf6SubBuf[ctr*8:], 0) == 0xFFFF {
					bin.WriteUInt64(buf6SubBuf, ctr*8, buf6SubBufVal)
				} else {
					buf6SubBufVal = bin.ReadUInt64(buf6SubBuf[ctr*8:], 0)
				}
			}
		}
	}

	return HuffmanTable{buf6}
}

func ReadHuffmanParams(data []byte, offset int) HuffmanParams {
	return HuffmanParams{
		int(bin.ReadInt32(data, offset+0)),
		int(bin.ReadInt32(data, offset+4)),
		int(bin.ReadInt32(data, offset+8)),
		int(bin.ReadInt16(data, offset+12))}
}

type HuffmanTable struct {
	data [][]byte
}

func (hp HuffmanTable) Length() int {
	return len(hp.data)
}

type HuffmanParams struct {
	p0 int
	p1 int
	p2 int
	p3 int
}
