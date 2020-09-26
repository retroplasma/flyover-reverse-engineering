package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly/c3m"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly/c3mm"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly/exp"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/mps"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/mps/config"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/mth"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/oth"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/web"
)

var l = log.New(os.Stderr, "", 0)

func printUsage(msg string) {
	if msg == "" {
		l.Println("Error:", msg)
	}
	l.Println("Usage", os.Args[0], "[lat] [lon] [zoom] [tryXY] [tryH]")
	l.Println()
	l.Println("  Name    Description       Example")
	l.Println("  --------------------------------------")
	ex := []string{"34.007603", "-118.499741", "20", "3", "40"}
	l.Println("  lat     Latitude         ", ex[0])
	l.Println("  lon     Longitude        ", ex[1])
	l.Println("  zoom    Zoom (~ 13-20)   ", ex[2])
	l.Println("  tryXY   Area scan        ", ex[3])
	l.Println("  tryH    Altitude scan    ", ex[4])
	l.Println("Example:", os.Args[0], ex[0], ex[1], ex[2], ex[3], ex[4])
	os.Exit(1)
}

func main() {

	var err error
	if len(os.Args) == 0 {
		printUsage("")
	}
	if len(os.Args) != 6 {
		printUsage("Invalid argument number")
	}
	lat, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		printUsage("Invalid lat")
	}
	lon, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		printUsage("Invalite lon")
	}
	zoom, err := strconv.ParseInt(os.Args[3], 10, 32)
	if err != nil {
		printUsage("Invalid zoom")
	}
	tryXY, err := strconv.ParseInt(os.Args[4], 10, 32)
	if err != nil {
		printUsage("Invalid tryXY")
	}
	tryH, err := strconv.ParseInt(os.Args[5], 10, 32)
	if err != nil {
		printUsage("Invalid tryH")
	}

	cache := mps.Cache{Enabled: true, Directory: "./cache"}
	err = cache.Init()
	oth.CheckPanic(err)
	config, err := config.FromJSONFile("./config.json")
	oth.CheckPanic(err)
	ctx, err := getContext(cache, config)
	oth.CheckPanic(err)

	z := int(zoom)
	x, y := mth.LatLonToTileTMS(z, lat, lon)

	p, err := ctx.findPlace(lat, lon)
	oth.CheckPanic(err)
	l.Println(p.Name, p.Radius, p.Lat, p.Lon)

	err = os.MkdirAll(fmt.Sprintf("./cache/c3mm/%d_%d", p.Region, p.Version), 0755)
	oth.CheckPanic(err)

	exportDir := fmt.Sprintf("./downloaded_files/obj/%f-%f-%d-%d-%d", lat, lon, zoom, tryXY, tryH)
	err = os.MkdirAll(exportDir, 0755)
	oth.CheckPanic(err)

	c3mm0, err := ctx.getC3mm(p, 0)
	oth.CheckPanic(err)
	smallestZ := 999
	for i := 0; i < len(c3mm0.RootIndex.Entries); i++ {
		e := c3mm0.RootIndex.Entries[i]
		if e.Z < smallestZ {
			smallestZ = e.Z
		}
	}

	c3mPfx, err := ctx.Context.ResourceManifest.URLPrefixFromStyleID(mps.ResourceManifest_StyleConfig_C3M)
	oth.CheckPanic(err)

	xp := 0
	export := exp.New(exportDir, "exp_")
	defer export.Close()

	c3m.DisableLogs()

	for d1 := -tryXY; d1 <= tryXY; d1++ {
		for d2 := -tryXY; d2 <= tryXY; d2++ {
			for h := 0; h < int(tryH); h++ {
				xn := x + int(d1)
				yn := y + int(d2)
				hn := h
				ok, err := check(ctx, p, smallestZ, z, yn, xn, hn)
				if err != nil {
					panic(err)
				}
				if !ok {
					continue
				}
				yn = mth.TileCountPerAxis(z) - 1 - yn
				url := fmt.Sprintf("%s?style=15&v=%d&region=%d&x=%d&y=%d&z=%d&h=%d", c3mPfx, p.Version, p.Region, xn, yn, z, hn)

				data, err := ctx.get(url)
				if err != nil {
					panic(err)
				}
				tile, err := c3m.Parse(data)
				if err != nil {
					panic(err)
				}
				l.Println("Exporting", d1, d2, "h =", h)
				export.Next(tile, fmt.Sprintf("%d", xp))
				xp++
			}
		}
	}
	l.Println(xp, "exported")
}

func check(ctx context, p fly.Trigger, smallestZ, z, y, x, h int) (bool, error) {
	if z < smallestZ {
		return false, errors.New("z too small")
	}

	c3mm0, err := ctx.getC3mm(p, 0)
	if err != nil {
		return false, err
	}

	list := make([]tile, 0)
	for z, y, x, h := z, y, x, h; z >= smallestZ; z, y, x, h = z-1, y/2, x/2, h/2 {
		list = append(list, tile{z, y, x, h})
	}

	root, listIdx := c3mm.Root{}, len(list)-1
	for ; listIdx >= 0; listIdx-- {
		t := list[listIdx]
		z, y, x, h := t.z, t.y, t.x, t.h
		n := len(c3mm0.RootIndex.Entries)
		idx := sort.Search(n, func(i int) bool {
			root := c3mm0.RootIndex.Entries[i]
			return !(root.Z < z || root.Z <= z && (root.Y < y || root.Y <= y && (root.X < x || root.X <= x && root.H < h)))
		})
		if idx == n {
			continue
		}
		root = c3mm0.RootIndex.Entries[idx]
		if !(z >= root.Z && (z > root.Z || y >= root.Y && (y > root.Y || x >= root.X && (x > root.X || h >= root.H)))) {
			continue
		}
		break
	}
	if listIdx < 0 {
		return false, nil
	}

	octantShift := root.Shift
	partNum := len(c3mm0.FileIndex.Entries) - 1
	for i := 0; i < len(c3mm0.FileIndex.Entries)-1; i++ {
		if octantShift < c3mm0.FileIndex.Entries[i+1] {
			partNum = i
			break
		}
	}
	c3mm1, err := ctx.getC3mm(p, partNum)
	if err != nil {
		return false, err
	}
	octant := c3mm1.GetOctant(&octantShift, c3mm0.FileIndex.Entries[partNum])

	le := list[listIdx]
	if le.z == z && le.y == y && le.x == x && le.h == h {
		return true, nil
	}
	if listIdx == 0 {
		return false, nil
	}

	for ; octant.Next > 0; listIdx-- {
		octantShift = octant.Next
		le := list[listIdx]
		zn, yn, xn, hn := 1+le.z, 2*le.y, 2*le.x, 2*le.h
		prepre := list[listIdx-1]
		bits := octant.Bits
		bittestPassed := false
		for o := 0; o < 8; o++ {
			if (bits>>(o*2))&1 != 1 {
				continue
			}

			partNum := len(c3mm0.FileIndex.Entries) - 1
			for i := 0; i < len(c3mm0.FileIndex.Entries)-1; i++ {
				if octantShift < c3mm0.FileIndex.Entries[i+1] {
					partNum = i
					break
				}
			}
			c3mm1, err := ctx.getC3mm(p, partNum)
			if err != nil {
				return false, err
			}
			octant = c3mm1.GetOctant(&octantShift, c3mm0.FileIndex.Entries[partNum])

			yn2 := yn | (o>>1)&1
			xn2 := xn | (o>>0)&1
			hn2 := hn | (o>>2)&1

			if zn == prepre.z && yn2 == prepre.y && xn2 == prepre.x && hn2 == prepre.h {
				bittestPassed = true
				break
			}
		}
		if !bittestPassed {
			return false, nil
		}
		if z == prepre.z && y == prepre.y && x == prepre.x && h == prepre.h {
			return true, nil
		}
	}
	return false, nil
}

type tile struct {
	z, y, x, h int
}

func (ctx *context) getC3mm(p fly.Trigger, part int) (c3mm.C3MM, error) {
	if ctx.C3mms == nil || ctx.C3mms[part] == nil {
		data, err := ctx._getC3mm(p, part)
		if err != nil {
			return c3mm.C3MM{}, err
		}
		if part == 0 {
			ctx.C3mms = make([]*c3mm.C3MM, len(data.FileIndex.Entries))
		}
		ctx.C3mms[part] = &data
		return data, nil
	}
	return *ctx.C3mms[part], nil
}

func (ctx *context) _getC3mm(p fly.Trigger, part int) (c3mm.C3MM, error) {
	fn := fmt.Sprintf("./cache/c3mm/%d_%d/%d", p.Region, p.Version, part)
	file, err := ioutil.ReadFile(fn)
	if err != nil && !os.IsNotExist(err) {
		return c3mm.C3MM{}, err
	}
	if err == nil {
		return c3mm.Parse(file, part)
	}
	if ctx.URLPrefixC3mm == nil {
		urlPfx, err := ctx.Context.ResourceManifest.URLPrefixFromStyleID(mps.ResourceManifest_StyleConfig_C3MM_1)
		if err != nil {
			return c3mm.C3MM{}, err
		}
		ctx.URLPrefixC3mm = &urlPfx
	}
	data, err := ctx.get(fmt.Sprintf("%s?style=14&v=%d&part=%d&region=%d", *ctx.URLPrefixC3mm, p.Version, part, p.Region))
	if err != nil {
		return c3mm.C3MM{}, err
	}
	if ioutil.WriteFile(fn+".tmp", data, 0655) != nil {
		return c3mm.C3MM{}, err
	}
	if os.Rename(fn+".tmp", fn) != nil {
		return c3mm.C3MM{}, err
	}
	return c3mm.Parse(data, part)
}

func (ctx context) findPlace(lat, lon float64) (fly.Trigger, error) {
	// radius non spherical yet
	minDist, minPlace := math.Inf(1), fly.Trigger{}

	for _, v := range ctx.AltitudeManifest.Triggers {
		dist := math.Sqrt(math.Pow(lat-v.Lat, 2) + math.Pow(lon-v.Lon, 2))
		// radius can overlap. ignored for now
		if dist <= v.Radius && dist < minDist {
			minDist, minPlace = dist, v
		}
	}
	if minDist == math.Inf(1) {
		return fly.Trigger{}, errors.New("minPlace not found")
	}
	return minPlace, nil
}

func (ctx context) get(url string) ([]byte, error) {
	authURL, err := ctx.Context.AuthContext.AuthURL(url)
	if err != nil {
		return nil, err
	}
	return get(authURL)
}

func get(url string) (data []byte, err error) {
	jpgErr := errors.New("received jpeg")
	data, err = web.GetWithCheck(url, func(res *http.Response) (err error) {
		// fail early if there's a jpeg, which is sometimes sent if there's no c3mm
		if res.Header.Get("content-type") == "image/jpeg" {
			err = jpgErr
		}
		return
	})
	return
}

type context struct {
	Context          mps.Context
	AltitudeManifest fly.AltitudeManifest
	URLPrefixC3mm    *string
	C3mms            []*c3mm.C3MM
}

func getContext(cache mps.Cache, config config.Config) (m context, err error) {
	ctx, err := mps.Init(cache, config)
	if err != nil {
		return
	}
	am, err := fly.GetAltitudeManifest(cache, ctx.ResourceManifest)
	if err != nil {
		return
	}
	m = context{ctx, am, nil, nil}
	return
}
