package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/bin"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/oth"
)

var l = log.New(os.Stderr, "", 0)

func SetLogPrefix(s string) {
	l.SetPrefix(s)
}

func DisableLogs() {
	l.SetFlags(0)
	l.SetOutput(ioutil.Discard)
}

type RawMeshData struct {
	Vertices      []float32
	VerticesCount int32
	UV            []float32
	UVCount       int32
	Faces         []int32
	Res5          []int32
	Res6          []int32
	Res7          []int32
	Res8          []int32
	FacesCount    int32
}

func Decompress(data []byte, dataOffset int, ebta HuffmanTable, ebtb HuffmanTable) RawMeshData {
	bufs := read10MeshBufs(data, dataOffset, ebta, ebtb)

	l.Printf("buf 0:")
	b0 := bufs[0]
	i32_0 := bin.ReadInt32(b0, 0)
	f64_0 := bin.ReadFloat64(b0, 4)
	f64_1 := bin.ReadFloat64(b0, 12)
	f64_2 := bin.ReadFloat64(b0, 20)
	f32_0 := bin.ReadFloat32(b0, 28)
	f32_1 := bin.ReadFloat32(b0, 32)
	f32_2 := bin.ReadFloat32(b0, 36)
	i8_0 := bin.ReadUInt8(b0, 40)
	res3 := bin.ReadInt32(b0, 41)
	i32_1 := bin.ReadInt32(b0, 45)
	i32_2 := bin.ReadInt32(b0, 49)
	i8_1 := bin.ReadUInt8(b0, 53)
	i32_3 := bin.ReadInt32(b0, 54)
	i32_4 := bin.ReadUInt32(b0, 58)

	l.Printf("  i32_0:    %d", i32_0)
	l.Printf("  f64_0-2:  %f %f %f", f64_0, f64_1, f64_2)
	l.Printf("  f32_0-2:  %f %f %f", f32_0, f32_1, f32_2)
	l.Printf("  i8_0:     %d", i8_0)
	l.Printf("  res3:     %d", res3)
	l.Printf("  i32_1:    %d", i32_1)
	l.Printf("  i32_2:    %d", i32_2)
	l.Printf("  i8_1:     %d", i8_1)
	l.Printf("  i32_3:    %d", i32_3)
	l.Printf("  i32_4:    %d", i32_4)

	if i32_0 < 0 || 0 == i8_0 || (i32_1|i32_2) < 0 || 0 == i8_1 || (i32_4&0x80000000) != 0 {
		panic("incorrect values in buf 0")
	}

	l.Printf("buf 5:")
	b5 := bufs[5]
	res9 := bin.ReadInt32(b5, 0)
	l.Printf("  res9: %d", res9)
	if res9 < 0 {
		panic("incorrect values in buf 5 #1")
	}
	i32_0min32 := i32_0 - 32
	fst := i32_0min32
	snd := 32
	buf_res9vmul3mul4_a := make([]int32, res9*3)
	for i := 0; i < len(buf_res9vmul3mul4_a); i++ {
		buf_res9vmul3mul4_a[i] = -1
	}

	buf_res9vmul3mul4_b := make([]int32, res9*3)
	for i := 0; i < len(buf_res9vmul3mul4_b); i++ {
		buf_res9vmul3mul4_b[i] = -1
	}

	if i32_0min32 >= 128 {
		for {
			val_in_data5_a := bin.ReadInt32(b5, snd/8)
			val_in_data5_b := bin.ReadInt32(b5, snd/8+4)
			if val_in_data5_a >= 0 {
				buf_res9vmul3mul4_a[val_in_data5_a] = val_in_data5_b
			}
			if val_in_data5_b >= 0 {
				buf_res9vmul3mul4_a[val_in_data5_b] = val_in_data5_a
			}

			fst -= 64
			snd += 64

			if !(fst > 127) {
				break
			}
		}
	}
	l.Printf("  buf_res9vmul3mul4_a len: %d", len(buf_res9vmul3mul4_a))

	res1 := bin.ReadInt32(b5, snd/8)
	l.Printf("  res1: %d", res1)
	b5unkn32 := bin.ReadInt32(b5, snd/8+4)
	l.Printf("  b5unkn32: %d", b5unkn32)

	if (res1 | b5unkn32) < 0 {
		panic("incorrect values in buf 5 #2")
	}

	l.Println("buf 2:")
	bufMetaCtr, bufMeta, writeBufOff, bufCLERS := decodeCLERS(bufs[2], res9, b5unkn32, buf_res9vmul3mul4_a)
	l.Printf("  CLERS: %s", oth.AbbrStr(fmt.Sprintf("%s", bufCLERS), 48))
	processCLERS(bufMeta, bufMetaCtr, bufCLERS, writeBufOff, b5unkn32, res1, buf_res9vmul3mul4_a, buf_res9vmul3mul4_b)
	res4 := buf_res9vmul3mul4_b
	l.Printf("res4: %p", res4)

	res7 := mkRes7(res9, bufs[3], i32_1, buf_res9vmul3mul4_a)
	for i := 0; i < len(res7); i++ {
		if res7[i] != 0 {
			l.Println("res7: has non zero!")
			break
		}
	}
	l.Printf("res7: %p", res7)

	res6 := make([]int32, res9)
	f60 := 0
	decompressList(res6, int(i32_2), bufs[4], &f60, int(i8_1))
	if i32_2 == 1 && res9 >= 2 {
		for i := range res6[1:] {
			res6[i+1] = res6[0]
		}
	}
	l.Printf("res6: %p", res6)

	d9dec := make([]int32, res1)
	f70p1 := 0
	decompressList(d9dec, int(i32_3), bufs[9], &f70p1, 1)
	bufoth_a := make([]int32, res1)
	bufoth_b := make([]int32, res1)
	for i := 0; i < 3*int(res9); i++ {
		if buf_res9vmul3mul4_a[i] == -1 {
			bufoth_b[res4[align3(i)+i+2-align3(i+2)]] = 1
			bufoth_b[res4[align3(i)+i+1-align3(i+1)]] = 1
		}
	}
	for i, read := 0, 0; i < int(res1); i++ {
		if bufoth_b[i] != 0 {
			bufoth_a[i] = d9dec[read]
			read++
		}
	}
	d9dec = bufoth_a
	res8 := d9dec
	l.Printf("res8: %p", res8)

	d8dec := make([]int32, res1)
	f70 := 0
	decompressList(d8dec, int(res1), bufs[8], &f70, 1)
	d1dec := make([]int32, i32_4)
	f50p1 := 0
	decompressList(d1dec, int(i32_4), bufs[1], &f50p1, 1)

	data_6_uv := make([]int16, len(bufs[6])/2)
	for i := range data_6_uv {
		data_6_uv[i] = bin.ReadInt16(bufs[6], i*2)
	}
	data_7_vtx := make([]int16, len(bufs[7])/2)
	for i := range data_7_vtx {
		data_7_vtx[i] = bin.ReadInt16(bufs[7], i*2)
	}

	uv_unpacked := make([]int16, res3*2)
	bf_res9mul12_a := make([]int32, res9*3)
	res5 := bf_res9mul12_a
	bf_res1mul4_a := make([]int32, res1)
	bf_res1mul4_b := make([]int32, res1)
	for i := range bf_res1mul4_b {
		bf_res1mul4_b[i] = -1
	}
	bf_res9mul4_a := make([]int32, res9)
	bf_res9mul12_b := make([]int32, res9*3)
	bf_res9mul12_c := make([]int32, res9*3)
	vtx_unpacked := make([]int16, res1*3)
	bf_res9mul4_b_res6t := make([]int32, res9)

	res3 = 0
	ctrA, ctrC, ctrB, ctrD, ctd := 0, 0, 0, 0, 0

BIG_LOOP:
	for {
		// a
		if ctrC < int(res9) {
			for {
				if 0 == bf_res9mul4_a[ctrC] {
					break
				}
				ctrC++
				if !(ctrC < int(res9)) {
					break
				}
			}
		}
		if ctrC == int(res9) {
			break
		}
		// b
		bf_res9mul12_b[ctrA] = 3 * int32(ctrC)
		ctrA++
		e1 := res4[align3(3*ctrC)+3*ctrC+2-align3(3*ctrC+2)]
		e2 := res4[3*ctrC]
		e3 := res4[align3(3*ctrC)+3*ctrC+1-align3(3*ctrC+1)]
		vtx_unpacked[e1*3+0] = data_7_vtx[e1*3+0] //e1
		vtx_unpacked[e1*3+1] = data_7_vtx[e1*3+1]
		vtx_unpacked[e1*3+2] = data_7_vtx[e1*3+2]
		vtx_unpacked[e2*3+0] = data_7_vtx[e2*3+0] //e2
		vtx_unpacked[e2*3+1] = data_7_vtx[e2*3+1]
		vtx_unpacked[e2*3+2] = data_7_vtx[e2*3+2]
		vtx_unpacked[e3*3+0] = data_7_vtx[e3*3+0] //e3
		vtx_unpacked[e3*3+1] = data_7_vtx[e3*3+1]
		vtx_unpacked[e3*3+2] = data_7_vtx[e3*3+2]
		bf_res1mul4_a[e1], bf_res1mul4_a[e2], bf_res1mul4_a[e3] = 1, 1, 1
		unpackUv(data_6_uv, uv_unpacked, &res3, res4,
			bf_res9mul12_a, bf_res1mul4_b, 3*ctrC)

		// c
		bf_res9mul4_a[ctrC] = 1
		bf_res9mul4_b_res6t[ctrC] = res6[ctrB]
		ctrB++
		if ctd|ctrA != 0 {
			ctrBplus1 := ctrB
			for {
				ctrBplus1plus0or1 := -1
				if ctrA != 0 {
					ctrBplus1plus0or1 = ctrBplus1
				} else {
					cntdwn := ctd - 1
					var v, x int32
					for {
						v = bf_res9mul12_c[cntdwn]
						x = bf_res9mul4_a[v/3]
						ctd--
						if ctd == 0 {
							break
						}
						cntdwn--
						if x == 0 {
							break
						}
					}
					if x != 0 {
						ctrA = 0
						ctrB = ctrBplus1
						continue BIG_LOOP
					}
					for i := range bf_res1mul4_b {
						bf_res1mul4_b[i] = -1
					}
					if 0 == bf_res1mul4_a[res4[v]] {
						unpackVtx(data_7_vtx, vtx_unpacked, buf_res9vmul3mul4_a, res4, bf_res1mul4_a, v)
					}
					unpackUv(data_6_uv, uv_unpacked, &res3, res4, bf_res9mul12_a, bf_res1mul4_b, int(v))
					ctrA = 1
					bf_res9mul4_a[v/3] = 1
					ctrBplus1plus0or1 = ctrBplus1 + 1
					bf_res9mul4_b_res6t[v/3] = res6[ctrBplus1]
					bf_res9mul12_b[0] = v
				}
				ctrBplus1 = ctrBplus1plus0or1
				ctrAmin1 := ctrA - 1
				cond := bf_res9mul12_b[ctrA-1]
				r6idx := ctrBplus1plus0or1 - 1
				ii := int(bf_res9mul12_b[ctrA-1])

				for {
					aVal := int(buf_res9vmul3mul4_a[ii])
					if aVal >= 0 {
						if 0 == bf_res9mul4_a[aVal/3] {
							idx1 := align3(ii) + ii + 2 - align3(ii+2)
							idx2 := align3(ii) + ii + 1 - align3(ii+1)
							other := true
							if d8dec[res4[idx1]] != 0 && d8dec[res4[idx2]] != 0 {
								ctrD++
								if d1dec[ctrD-1] != 0 {
									bf_res9mul12_c[ctd] = int32(aVal)
									ctd++
									other = false
								}
							}
							if other {
								tmp1 := res4[aVal]
								if 0 == bf_res1mul4_a[tmp1] {
									unpackVtx(data_7_vtx, vtx_unpacked,
										buf_res9vmul3mul4_a, res4,
										bf_res1mul4_a, int32(aVal))
								}
								r3tmp := bf_res1mul4_b[tmp1]
								if r3tmp == -1 {
									aValNxt := int(buf_res9vmul3mul4_a[aVal])
									nidx1 := bf_res9mul12_a[align3(aValNxt)+aValNxt+1-align3(aValNxt+1)]
									nidx2 := bf_res9mul12_a[align3(aValNxt)+aValNxt+2-align3(aValNxt+2)]
									nidx3 := bf_res9mul12_a[aValNxt]

									nu := uv_unpacked[2*nidx1+0] + uv_unpacked[2*nidx2+0] - uv_unpacked[2*nidx3+0]
									nv := uv_unpacked[2*nidx1+1] + uv_unpacked[2*nidx2+1] - uv_unpacked[2*nidx3+1]
									r3tmp = res3
									uv_unpacked[2*r3tmp] = nu - data_6_uv[2*r3tmp]
									uv_unpacked[2*r3tmp+1] = nv - data_6_uv[2*r3tmp+1]
									res3++
									bf_res1mul4_b[res4[aVal]] = r3tmp
								}
								bf_res9mul12_a[aVal] = r3tmp
								bf_res9mul12_a[align3(aVal)+aVal+2-align3(aVal+2)] = bf_res9mul12_a[idx2]
								bf_res9mul12_a[align3(aVal)+aVal+1-align3(aVal+1)] = bf_res9mul12_a[idx1]
								bf_res9mul4_a[aVal/3] = 1
								bf_res9mul4_b_res6t[aVal/3] = res6[r6idx]
								bf_res9mul12_b[ctrAmin1] = int32(aVal)
								ctrAmin1++
							}

						}
					}
					ii = align3(ii) + ii + 1 - align3(ii+1)
					if !(ii != int(cond)) {
						break
					}
				}
				ctrA = ctrAmin1

				if 0 == (ctd | ctrAmin1) {
					ctrB = ctrBplus1
					continue BIG_LOOP
				}
			}

		}
	}
	l.Printf("res5: %p", res5)

	vtxBuff := make([]float32, res1*3)
	for i := 0; i < int(res1)*3; i += 3 {
		vtxBuff[i+0] = float32(float64(vtx_unpacked[i+0])*f64_0 + float64(f32_0))
		vtxBuff[i+1] = float32(float64(vtx_unpacked[i+1])*f64_1 + float64(f32_1))
		vtxBuff[i+2] = float32(float64(vtx_unpacked[i+2])*f64_2 + float64(f32_2))
	}
	res0 := vtxBuff
	l.Printf("res0: %p", res0)

	uvBuff := make([]float32, res3*2)
	scale := 1.0 / float64(((uint(1) << uint(i8_0)) - 1))
	for i := 0; i < int(res3)*2; i += 2 {
		uvBuff[i+0] = float32(float64(uv_unpacked[i+0]) * scale)
		uvBuff[i+1] = float32(float64(uv_unpacked[i+1]) * scale)
	}

	res2 := uvBuff
	l.Printf("res2: %p", res2)

	res6 = bf_res9mul4_b_res6t
	l.Printf("rewrote res6: %p", res6)

	rmd := RawMeshData{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}

	if len(rmd.Vertices)/3 != int(rmd.VerticesCount) || len(rmd.Faces)/3 != int(rmd.FacesCount) || len(rmd.UV)/2 != int(rmd.UVCount) {
		panic("decompression error")
	}

	return rmd
}

func unpackVtx(data_7_vtx []int16, vtx_unpacked []int16, buf_res9vmul3mul4_a []int32,
	res4 []int32, bf_res1mul4_a []int32, aVal int32) {

	idx3 := int(buf_res9vmul3mul4_a[aVal])
	vtx_h := 3 * res4[align3(idx3)+idx3+1-align3(idx3+1)]
	vtx_i := 3 * res4[align3(idx3)+idx3+2-align3(idx3+2)]
	vtx_j := 3 * res4[idx3]
	k := res4[aVal]
	vtx_k := 3 * k
	vtx_unpacked[vtx_k+0] = vtx_unpacked[vtx_h+0] + vtx_unpacked[vtx_i+0] - vtx_unpacked[vtx_j+0] - data_7_vtx[vtx_k+0]
	vtx_unpacked[vtx_k+1] = vtx_unpacked[vtx_h+1] + vtx_unpacked[vtx_i+1] - vtx_unpacked[vtx_j+1] - data_7_vtx[vtx_k+1]
	vtx_unpacked[vtx_k+2] = vtx_unpacked[vtx_h+2] + vtx_unpacked[vtx_i+2] - vtx_unpacked[vtx_j+2] - data_7_vtx[vtx_k+2]
	bf_res1mul4_a[k] = 1
}

func unpackUv(data_6_uv []int16, uv_unpacked []int16, res3 *int32, res4 []int32,
	bf_res9mul12_a []int32, bf_res1mul4_b []int32, ctrCMul3 int) {
	for _, idx := range []int{
		align3(ctrCMul3) + ctrCMul3 + 2 - align3(ctrCMul3+2),
		ctrCMul3,
		align3(ctrCMul3) + ctrCMul3 + 1 - align3(ctrCMul3+1),
	} {
		uv_unpacked[2**res3+0] = data_6_uv[2**res3+0]
		uv_unpacked[2**res3+1] = data_6_uv[2**res3+1]
		bf_res9mul12_a[idx] = *res3
		bf_res1mul4_b[res4[idx]] = bf_res9mul12_a[idx]
		*res3++
	}
}

func mkRes7(res9 int32, buf3 []byte, i32_1 int32, buf_res9vmul3mul4_a []int32) (res7 []int32) {
	res7 = make([]int32, res9)

	field58plus1 := 0
	decompressList(res7, int(i32_1), buf3, &field58plus1, 1)
	if res9 <= 0 {
		panic("res9 <= 0 not implemented")
	}

	bufres94 := make([]int32, res9)

	bufres94ptr := 1
	res9vmin1 := res9 - 1
	var bufres94tmpVal int32
	triOff := 2
	ctr := 0
	ctr2 := 0
	ctr3 := 0
	var bufres94res int32
	for {
		if bufres94tmpVal == 0 {
			ctr3 = 3 * ctr
			other := false
			if buf_res9vmul3mul4_a[triOff-2] == -1 {
				other = true
			} else if buf_res9vmul3mul4_a[triOff-1] == -1 {
				ctr3++
				other = true
			} else {
				bufres94res = 0
				ctr3 = triOff
				if buf_res9vmul3mul4_a[triOff] == -1 {
					other = true
				}
			}
			if other {
				res7idx := ctr2
				ctr2++
				bufres94res = res7[res7idx]
				if 0 != bufres94res {
					l.Println("branch not visited yet")
					inner := buf_res9vmul3mul4_a[align3(ctr3)+ctr3+2-align3(ctr3+2)]
					bufres94[inner/3] = bufres94res
				} else {
					bufres94res = 0
				}
			}
			bufres94[bufres94ptr-1] = bufres94res
		}
		if 0 == res9vmin1 {
			return
		}
		ctr++
		bufres94tmpVal = bufres94[bufres94ptr]
		bufres94ptr++
		res9vmin1--
		triOff += 3
	}
}

func decompressList(outBuf []int32, length int, inBuf []byte, outNum *int, sh int) {
	readShift := 0
	inBufOff := 0
	outBufOff := 0
	var result uint64
	if length > 0 {
		var input uint64
		for {
			if readShift < sh {
				input |= uint64(bin.ReadUInt32BE(inBuf, inBufOff)) << uint(32-readShift)
				readShift += 32
				inBufOff += 4
			}
			result = input >> uint(64-sh)
			readShift -= sh
			input <<= uint(sh)
			outBuf[outBufOff] = int32(result)
			outBufOff++
			length--
			if length == 0 {
				break
			}
		}
	}
	*outNum += 8*inBufOff - readShift
	//l.Println("decompressList result", result)
	_ = result
}

func processCLERS(bufMeta []int, bufMetaCtr int, bufCLERS []byte, writeBufOff int, _b5unkn32 int32, res1 int32, buf_res9vmul3mul4_a []int32, buf_res9vmul3mul4_b []int32) {
	b5unkn32 := int(_b5unkn32)
	if bufMetaCtr <= 0 {
		panic("not implemented: bufMetaCtr <= 0")
	}
	if writeBufOff <= 0 {
		panic("not implemeted: writeBufOff <= 0")
	}
	res1vaga := 0
	res1v_min1ag := res1 - 1
	tmpBuf := make([]int32, 3*writeBufOff)
	writeBufOff--

	for {
		res9vmul3agb := b5unkn32
		bufMetaCtr--
		tmp1 := -1
		for {
			if !(bufMetaCtr >= b5unkn32 || bufMetaCtr >= 0 && writeBufOff >= bufMeta[bufMetaCtr]) {
				break
			}
			clersVal := bufCLERS[writeBufOff]

			switch clersVal {
			case 'C':
				writeBufOffMul3 := 3 * writeBufOff
				if tmp1 >= 0 {
					buf_res9vmul3mul4_a[tmp1] = int32(3*writeBufOff + 1)
				}
				if writeBufOffMul3 >= -1 {
					buf_res9vmul3mul4_a[3*writeBufOff+1] = int32(tmp1)
				}
				tmp2 := res1v_min1ag
				res1v_min1ag--
				closeStar(buf_res9vmul3mul4_a, buf_res9vmul3mul4_b, writeBufOffMul3+2, tmp2)
				b5unkn32 = res9vmul3agb
			case 'L':
				if tmp1 >= 0 {
					buf_res9vmul3mul4_a[tmp1] = int32(3*writeBufOff + 1)
				}
				if 3*writeBufOff >= -1 {
					buf_res9vmul3mul4_a[3*writeBufOff+1] = int32(tmp1)
				}
			case 'E':
				if tmp1 > 0 {
					tmpBuf[res1vaga] = int32(tmp1)
					res1vaga++
				}
			case 'R':
				tmp3 := 3*writeBufOff + 2
				if tmp1 >= 0 {
					buf_res9vmul3mul4_a[tmp1] = int32(tmp3)
				}
				if tmp3 >= 0 {
					buf_res9vmul3mul4_a[3*writeBufOff+2] = int32(tmp1)
				}
			case 'S':
				writeBufOffMul3_2 := 3 * writeBufOff
				if tmp1 >= 0 {
					buf_res9vmul3mul4_a[tmp1] = int32(3*writeBufOff + 1)
				}
				if writeBufOffMul3_2 >= -1 {
					buf_res9vmul3mul4_a[3*writeBufOff+1] = int32(tmp1)
				}
				tmp4 := writeBufOffMul3_2 + 2
				tmp5 := buf_res9vmul3mul4_a[3*writeBufOff+2]
				if tmp5 == -1 {
					tmp6 := tmpBuf[res1vaga-1]
					if tmp4 >= 0 {
						buf_res9vmul3mul4_a[3*writeBufOff+2] = tmp6
					}
					res1vaga--
					if tmp6 >= 0 {
						buf_res9vmul3mul4_a[tmp6] = int32(tmp4)
					}
				} else if tmp5 <= -2 {
					tmp7 := -tmp5
					if tmp4 >= 0 {
						buf_res9vmul3mul4_a[3*writeBufOff+2] = tmp7
					}
					buf_res9vmul3mul4_a[tmp7] = int32(tmp4)
					tmp8 := 0
					for {
						tmp8 = align3(tmp4) + tmp4 + 1 - align3(tmp4+1)
						tmp4 = int(buf_res9vmul3mul4_a[tmp8])
						if !(tmp4 >= 0) {
							break
						}
					}
					readBoundary(buf_res9vmul3mul4_a, buf_res9vmul3mul4_b, tmp8, &res1v_min1ag)
				}
			case 'P':
				writeBufOffMul3_1 := 3 * writeBufOff
				if tmp1 >= 0 {
					buf_res9vmul3mul4_a[tmp1] = int32(writeBufOffMul3_1)
				}
				if writeBufOff >= 0 {
					buf_res9vmul3mul4_a[3*writeBufOff] = int32(tmp1)
				}
				closeStar(buf_res9vmul3mul4_a, buf_res9vmul3mul4_b, writeBufOffMul3_1+1, res1v_min1ag-2)
				closeStar(buf_res9vmul3mul4_a, buf_res9vmul3mul4_b, writeBufOffMul3_1+2, res1v_min1ag-1)
				closeStar(buf_res9vmul3mul4_a, buf_res9vmul3mul4_b, writeBufOffMul3_1, res1v_min1ag)
				res1v_min1ag -= 3
				bufMetaCtr--
				b5unkn32 = res9vmul3agb
			default:
				//l.Println("skipping symbol", clersVal)
				panic(fmt.Sprintf("unknown symbol %b", clersVal))
			}
			tmp1 = 3 * writeBufOff
			writeBufOff--
		}

		tmp9 := 0
		if b5unkn32 != 0 {
			readBoundary(buf_res9vmul3mul4_a, buf_res9vmul3mul4_b, 3*bufMeta[bufMetaCtr]+1, &res1v_min1ag)
			tmp9 = b5unkn32 - 1
		}
		b5unkn32 = tmp9

		if bufMetaCtr <= 0 {
			break
		}
	}
	if res1v_min1ag != -1 {
		panic("res1v_min1ag not -1")
	}
}

func closeStar(buf_res9vmul3mul4_a []int32, buf_res9vmul3mul4_b []int32, writeBufOffMul3plus2 int, t2 int32) {
	tmp1 := align3(writeBufOffMul3plus2)
	tmp2 := tmp1 + writeBufOffMul3plus2 + 2 - align3(writeBufOffMul3plus2+2)
	tmp3 := tmp2
	tmp4 := buf_res9vmul3mul4_a[tmp2]
	tmp5 := writeBufOffMul3plus2
	tmp6 := 0
	result := 0
outer:
	for {
		for {
			tmp6 = tmp5 + 1
			if tmp4 < 0 {
				break
			}
			result = tmp1 + tmp6 - align3(tmp6)
			buf_res9vmul3mul4_b[result] = t2
			if tmp4 == int32(writeBufOffMul3plus2) {
				break outer
			}
			tmp5 = int(buf_res9vmul3mul4_a[tmp3])
			tmp1 = align3(tmp5)
			tmp2 = tmp1 + tmp5 + 2 - align3(tmp5+2)
			tmp3 = tmp2
			tmp4 = buf_res9vmul3mul4_a[tmp2]
		}
		result = tmp1 + tmp6 - align3(tmp6)
		buf_res9vmul3mul4_b[result] = t2
		break
	}
	if tmp2 >= 0 {
		buf_res9vmul3mul4_a[tmp3] = int32(writeBufOffMul3plus2)
	}
	if writeBufOffMul3plus2 >= 0 {
		buf_res9vmul3mul4_a[writeBufOffMul3plus2] = int32(tmp2)
	}
	//l.Println("closeStar result", result)
	_ = result
}

func readBoundary(buf_res9vmul3mul4_a []int32, buf_res9vmul3mul4_b []int32, someIdx int, outNum *int32) {
	tmp1 := align3(someIdx) + someIdx + 1 - align3(someIdx+1)
	result := 0
	for {
		result = align3(tmp1) + tmp1 + 1 - align3(tmp1+1)
		tmp1 = int(buf_res9vmul3mul4_a[result])
		if !(tmp1 >= 0) {
			break
		}
	}
	tmp2 := *outNum
	for {
		tmp3 := align3(result)
		buf_res9vmul3mul4_b[tmp3+result+1-align3(result+1)] = tmp2
		result = tmp3 + result + 2 - align3(result+2)
		tmp4 := int(buf_res9vmul3mul4_a[result])
		var i int32
		for i = *outNum; tmp4 >= 0; i = *outNum {
			tmp5 := align3(tmp4)
			buf_res9vmul3mul4_b[tmp5+tmp4+1-align3(tmp4+1)] = i
			result = tmp5 + tmp4 + 2 - align3(tmp4+2)
			tmp4 = int(buf_res9vmul3mul4_a[result])
		}
		tmp2 = i - 1
		*outNum = tmp2

		if !(buf_res9vmul3mul4_b[align3(result)+result+1-align3(result+1)] == -1) {
			break
		}
	}
	//l.Println("readBoundary result", result)
	_ = result
}

func decodeCLERS(b2 []byte, res9 int32, b5unkn32 int32, buf_res9vmul3mul4_a []int32) (int, []int, int, []byte) {

	bufMeta := make([]int, res9)
	bufCLERS := make([]byte, res9*3)
	writeBufOff := 0
	if b5unkn32 == 0 {
		writeBufOff = 0
		if res9 > 0 {
			writeBufOff = 1
			bufCLERS[0] = 'P'
		}
	}

	if writeBufOff >= int(res9) {
		panic("not implemented: no decoding of data2")
	}
	var input uint64
	rs := 0
	bmcTmp := 0
	updown := 0
	var code uint64
	readBufOff := 0
	bufMetaCtr := bmcTmp

BIG_LOOP:
	for {
		triCtr := 3 * writeBufOff
		wboTmp := writeBufOff
		othCtr := 0
		readShift := rs
		for {
			if readShift <= 0 {
				input |= uint64(bin.ReadUInt32BE(b2, readBufOff)) << uint(32-readShift)
				readShift += 32
				readBufOff += 4
			}
			rs = readShift - 1
			outVal := 'C'
			tmp := input & 0x8000000000000000
			flag := 0
			if tmp != 0 {
				flag = 1
			}
			input *= 2
			if 0 != flag {
				if readShift <= 2 {
					input |= uint64(bin.ReadUInt32BE(b2, readBufOff)) << uint(33-readShift)
					readBufOff += 4
					rs = readShift + 31
				}
				code = input >> 62
				rs -= 2
				input *= 4
				if 0 == uint32(code) {
					break
				}
				if uint32(code) == 3 {
					writeBufOff += othCtr + 1
					bufCLERS[wboTmp+othCtr] = 'E'
					if updown > 0 {
						updown--
						if writeBufOff < int(res9) {
							continue BIG_LOOP
						}
						break BIG_LOOP
					}
					bmcTmp = bufMetaCtr + 1
					if writeBufOff < int(res9) {
						if bmcTmp >= int(b5unkn32) {
							bufCLERS[wboTmp+1+othCtr] = 'P'
							writeBufOff = wboTmp + othCtr + 2
						} else {
							bufMeta[bufMetaCtr+1] = writeBufOff
						}
					}
					if writeBufOff >= int(res9) {
						bufMetaCtr++
						break BIG_LOOP
					}
					bufMetaCtr = bmcTmp
					continue BIG_LOOP
				}
				outVal = 'L'
				if uint32(code) == 1 {
					outVal = 'R'
				}
			}
			bufCLERS[writeBufOff+othCtr] = byte(outVal)
			othCtr++
			triCtr += 3
			readShift = rs
			if othCtr+writeBufOff >= int(res9) {
				writeBufOff += othCtr
				break BIG_LOOP
			}
		}
		bufCLERS[writeBufOff+othCtr] = 'S'

		idx := triCtr + 2 - align3(triCtr+2) + align3(triCtr)

		if buf_res9vmul3mul4_a[idx] == -1 {
			updown++
		}
		writeBufOff += othCtr + 1

		if writeBufOff < int(res9) {
			continue
		}
		break
	}

	return bufMetaCtr, bufMeta, writeBufOff, bufCLERS
}

func align3(input int) int {
	return 3 * (input / 3)
}

func read10MeshBufs(data []byte, dataOffset int, ebta HuffmanTable, ebtb HuffmanTable) (bufs [10][]byte) {
	l.Println("* buf  type  len1   len2   data                                 desc")
	off := 120
	for i := 0; i < 10; i++ {
		len1 := int(bin.ReadUInt32(data, dataOffset+12*i))
		len2 := int(bin.ReadUInt32(data, dataOffset+12*i+4))
		val := bin.ReadUInt8(data, dataOffset+12*i+8)

		desc := "?"
		switch i {
		case 0:
			desc = "header"
		case 2:
			desc = "eb clers"
		case 5:
			desc = "eb other"
		case 6:
			desc = "uv"
		case 7:
			desc = "vtx"
		}

		outBuf := make([]byte, len1+3)
		switch val {
		case 0:
			buf := data[dataOffset+off : dataOffset+off+int(len2)]
			l.Printf("  %d    0     %-5d  %-5d  %-35s  %s", i, len1, len2, oth.AbbrHexStr(buf, 32), desc)
			copy(outBuf, buf)
		case 3:
			buf := data[dataOffset+off : dataOffset+off+int(len2)]
			l.Printf("  %d    3     %-5d  %-5d  %-35s  %s", i, len1, len2, oth.AbbrHexStr(buf, 32), desc)
			hp, s := ebta, "a"
			if i == 7 {
				hp, s = ebtb, "b"
			}
			hp.Decode(buf, len1, len2, &outBuf)
			l.Printf("  -> decoded (eb_table_%s): %s", s, oth.AbbrHexStr(outBuf, 32))
		case 1:
			panic("Unsupported type: 1")
		default:
			panic(fmt.Sprintf("Unknown type: %d", val))
		}
		bufs[i] = outBuf

		off += len2
	}
	return
}
