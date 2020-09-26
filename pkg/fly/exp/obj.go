package exp

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly/c3m"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/oth"
)

var transform = true

type Export interface {
	Next(c3m c3m.C3M, subPfx string)
	Close()
}

func New(dir string, fnPfx string) Export {
	objFile := create(path.Join(dir, fmt.Sprintf("%smodel.obj", fnPfx)))
	mtlFile := create(path.Join(dir, fmt.Sprintf("%smodel.mtl", fnPfx)))
	return &OBJExport{
		dir:           dir,
		fnPfx:         fnPfx,
		vtxCount:      0,
		objFile:       objFile,
		objFileWriter: bufio.NewWriter(objFile),
		mtlFile:       mtlFile,
		mtlFileWriter: bufio.NewWriter(mtlFile),
	}
}

func (e *OBJExport) Close() {
	oth.CheckPanic(e.objFileWriter.Flush())
	oth.CheckPanic(e.objFile.Close())
	oth.CheckPanic(e.mtlFileWriter.Flush())
	oth.CheckPanic(e.mtlFile.Close())
}

type OBJExport struct {
	dir           string
	fnPfx         string
	vtxCount      int
	objFile       *os.File
	objFileWriter *bufio.Writer
	mtlFile       *os.File
	mtlFileWriter *bufio.Writer
}

func (e *OBJExport) Next(c3m c3m.C3M, subPfx string) {
	dir, fnPfx := e.dir, e.fnPfx

	for i, material := range c3m.Materials {
		err := ioutil.WriteFile(path.Join(dir, fmt.Sprintf("%s%s_%d.jpg", fnPfx, subPfx, i)), material.JPEG, 0655)
		oth.CheckPanic(err)
		nxt := fmt.Sprintf(`
newmtl mtl_%s_%d
Kd 1.000 1.000 1.000
d 1.0
illum 0
map_Kd %s%s_%d.jpg
`, subPfx, i, fnPfx, subPfx, i)
		write(e.mtlFileWriter, nxt)
	}

	for i, mesh := range c3m.Meshes {

		write(e.objFileWriter, fmt.Sprintf("mtllib %smodel.mtl\n", fnPfx))
		write(e.objFileWriter, fmt.Sprintf("o test_%s_%d\n", subPfx, i))
		for _, vtx := range mesh.Vertices {
			x, y, z := float64(vtx.X), float64(vtx.Y), float64(vtx.Z)
			if transform {
				x, y, z =
					c3m.Header.Rotation[0]*x+c3m.Header.Rotation[1]*y+c3m.Header.Rotation[2]*z,
					c3m.Header.Rotation[3]*x+c3m.Header.Rotation[4]*y+c3m.Header.Rotation[5]*z,
					c3m.Header.Rotation[6]*x+c3m.Header.Rotation[7]*y+c3m.Header.Rotation[8]*z
				x += c3m.Header.Translation[0]
				y += c3m.Header.Translation[1]
				z += c3m.Header.Translation[2]
			}

			write(e.objFileWriter, fmt.Sprintln("v", x, y, z))
			write(e.objFileWriter, fmt.Sprintln("vt", vtx.U, vtx.V))
		}

		for i, group := range mesh.Groups {
			write(e.objFileWriter, fmt.Sprintf("g g_%s_%d\n", subPfx, i))
			write(e.objFileWriter, fmt.Sprintf("usemtl mtl_%s_%d\n", subPfx, i))
			for _, face := range group.Faces {
				a, b, c := int(face.A)+1+e.vtxCount, int(face.B)+1+e.vtxCount, int(face.C)+1+e.vtxCount
				write(e.objFileWriter, fmt.Sprintf("f %d/%d %d/%d %d/%d\n", a, a, b, b, c, c))
			}
		}
		e.vtxCount += len(mesh.Vertices)
	}
}

func create(fn string) *os.File {
	perm := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	file, err := os.OpenFile(fn, perm, 0655)
	oth.CheckPanic(err)
	return file
}

func write(f *bufio.Writer, txt string) {
	_, err := f.WriteString(txt)
	oth.CheckPanic(err)
}
