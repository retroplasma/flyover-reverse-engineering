package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/flyover-reverse-engineering/pkg/fly"
	"github.com/flyover-reverse-engineering/pkg/fly/c3m"
	"github.com/flyover-reverse-engineering/pkg/fly/exp"
	"github.com/flyover-reverse-engineering/pkg/mps"
	"github.com/flyover-reverse-engineering/pkg/mps/config"
	"github.com/flyover-reverse-engineering/pkg/mth"
	"github.com/flyover-reverse-engineering/pkg/oth"
	"github.com/flyover-reverse-engineering/pkg/web"
)

var l = log.New(os.Stderr, "", 0)

func main() {
	lat, lon := 34.007603, -118.499741
	exportDir := "./export"
	err := os.MkdirAll(exportDir, 0755)
	oth.CheckPanic(err)
	cache := mps.Cache{Enabled: true, Directory: "./cache"}
	err = cache.Init()
	oth.CheckPanic(err)
	config, err := config.FromJSONFile("./config.json")
	oth.CheckPanic(err)
	ctx, err := getContext(cache, config)
	oth.CheckPanic(err)
	ctx.exportOBJ(exportDir, lat, lon)
}

func (ctx context) exportOBJ(dir string, lat, lon float64) {
	authCtx := ctx.Context.AuthContext

	minDist, minPlace := math.Inf(1), fly.Trigger{}

	for _, v := range ctx.AltitudeManifest.Triggers {
		dist := math.Sqrt(math.Pow(lat-v.Lat, 2) + math.Pow(lon-v.Lon, 2))
		// radius can overlap. ignored for now
		if dist <= v.Radius && dist < minDist {
			minDist, minPlace = dist, v
		}
	}
	if minDist == math.Inf(1) {
		panic("minPlace not found")
	}
	l.Println(minPlace.Name, minPlace.Radius, minPlace.Lat, minPlace.Lon)

	p := minPlace
	z := 20
	x, y := mth.LatLonToTile(z, lat, lon)
	urlPrefix, err := ctx.Context.ResourceManifest.URLPrefixFromStyleID(mps.ResourceManifest_StyleConfig_C3M)
	oth.CheckPanic(err)

	xp := 0
	tryXY := 3
	// height is scanned because that info is probably in c3mm and its parser isn't done yet
	tryH := 2
	export := exp.New(dir, "exp_")

	c3m.DisableLogs()
	for d1 := -tryXY; d1 <= tryXY; d1++ {
		for d2 := -tryXY; d2 <= tryXY; d2++ {
			for i := 0; i < tryH; i++ {
				url := fmt.Sprintf("%s?style=15&v=%d&region=%d&x=%d&y=%d&z=%d&h=%d", urlPrefix, p.Version, p.Region, x+d1, y+d2, z, i)
				authURL, err := authCtx.AuthURL(url)
				oth.CheckPanic(err)
				l.Println("Downloading...")
				data, err := web.Get(authURL)
				oth.CheckPanic(err)
				l.Println("Decoding...")

				if len(data) == 0 {
					l.Println("Skipping")
					continue
				}
				tile, err := c3m.Parse(data)
				if err != nil {
					l.Println(err)
					continue
				}
				l.Println("h =", i)
				l.Println("Exporting...")
				export.Next(tile, fmt.Sprintf("%d", xp))
				xp++
			}
		}
	}
	l.Println(xp, "exported")
}

type context struct {
	Context          mps.Context
	AltitudeManifest fly.AltitudeManifest
}

func getContext(cache mps.Cache, config config.Config) (m context, err error) {
	ctx, err := mps.Init(cache, config)
	am, err := fly.GetAltitudeManifest(cache, ctx.ResourceManifest)
	if err != nil {
		return
	}
	m = context{ctx, am}
	return
}
