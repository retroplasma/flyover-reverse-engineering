package exp

import (
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
}

func New(dir string, fnPfx string) Export {
	return &OBJExport{dir, fnPfx, 0, true}
}

type OBJExport struct {
	dir      string
	fnPfx    string
	vtxCount int
	first    bool
}

func (e *OBJExport) Next(c3m c3m.C3M, subPfx string) {
	dir, fnPfx, first := e.dir, e.fnPfx, e.first
	e.first = false

	mtllib := ""
	for i, material := range c3m.Materials {
		err := ioutil.WriteFile(fmt.Sprintf(path.Join(dir, "%s%s_%d.jpg"), fnPfx, subPfx, i), material.JPEG, 0655)
		oth.CheckPanic(err)
		mtllib += fmt.Sprintf(`
newmtl mtl_%s_%d
Ka 1.000000 1.000000 1.000000
Kd 1.000000 1.000000 1.000000
Ks 0.000000 0.000000 0.000000
Tr 1.000000
illum 1
Ns 0.000000
map_Kd %s%s_%d.jpg
	`, subPfx, i, fnPfx, subPfx, i)
	}
	err := write(fmt.Sprintf(path.Join(dir, "%smodel.mtl"), fnPfx), mtllib, first)
	oth.CheckPanic(err)

	obj := ""
	for i, mesh := range c3m.Meshes {

		obj += fmt.Sprintf("mtllib %smodel.mtl\n", fnPfx)
		obj += fmt.Sprintf("o test_%s_%d\n", subPfx, i)
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

			obj += fmt.Sprintln("v", x, y, z)
			obj += fmt.Sprintln("vt", vtx.U, vtx.V)
		}

		for i, group := range mesh.Groups {
			obj += fmt.Sprintf("g g_%s_%d\n", subPfx, i)
			obj += fmt.Sprintf("usemtl mtl_%s_%d\n", subPfx, i)
			for _, face := range group.Faces {
				a, b, c := int(face.A)+1+e.vtxCount, int(face.B)+1+e.vtxCount, int(face.C)+1+e.vtxCount
				obj += fmt.Sprintf("f %d/%d %d/%d %d/%d\n", a, a, b, b, c, c)
			}
		}

		e.vtxCount += len(mesh.Vertices)
	}

	err = write(fmt.Sprintf(path.Join(dir, "%smodel.obj"), fnPfx), obj, first)
	oth.CheckPanic(err)
}

func write(fn string, txt string, first bool) (err error) {
	perm := os.O_CREATE | os.O_WRONLY
	if !first {
		perm |= os.O_APPEND
	} else {
		perm |= os.O_TRUNC
	}
	f, err := os.OpenFile(fn, perm, 0655)
	if err != nil {
		return
	}
	defer f.Close()
	if _, err = f.WriteString(txt); err != nil {
		return
	}
	return
}
