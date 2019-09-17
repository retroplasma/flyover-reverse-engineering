package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly/c3m"
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
	ex := []string{"34.007603", "-118.499741", "20", "3", "2"}
	l.Println("  lat     Latitude         ", ex[0])
	l.Println("  lon     Longitude        ", ex[1])
	l.Println("  zoom    Zoom (~ 13-20)   ", ex[2])
	l.Println("  tryXY   Horizontal scan  ", ex[3])
	l.Println("  tryH    Vertical scan    ", ex[4])
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

	exportDir := fmt.Sprintf("./downloaded_files/obj/%f-%f-%d-%d-%d", lat, lon, zoom, tryXY, tryH)
	err = os.MkdirAll(exportDir, 0755)
	oth.CheckPanic(err)
	cache := mps.Cache{Enabled: true, Directory: "./cache"}
	err = cache.Init()
	oth.CheckPanic(err)
	config, err := config.FromJSONFile("./config.json")
	oth.CheckPanic(err)
	ctx, err := getContext(cache, config)
	oth.CheckPanic(err)
	ctx.exportOBJ(exportDir, lat, lon, int(tryXY), int(tryH), int(zoom))
}

func (ctx context) exportOBJ(dir string, lat, lon float64, tryXY, tryH int, z int) {
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
	x, y := mth.LatLonToTile(z, lat, lon)
	urlPrefix, err := ctx.Context.ResourceManifest.URLPrefixFromStyleID(mps.ResourceManifest_StyleConfig_C3M)
	oth.CheckPanic(err)

	xp := 0
	export := exp.New(dir, "exp_")

	c3m.DisableLogs()
	numSkippedDataTmp := 0
	l.Println("Searching...")

	for d1 := -tryXY; d1 <= tryXY; d1++ {
		for d2 := -tryXY; d2 <= tryXY; d2++ {
			// height is scanned because that info is probably in c3mm and its parser isn't done yet
			for i := 0; i < tryH; i++ {
				url := fmt.Sprintf("%s?style=15&v=%d&region=%d&x=%d&y=%d&z=%d&h=%d", urlPrefix, p.Version, p.Region, x+d1, y+d2, z, i)
				authURL, err := authCtx.AuthURL(url)
				oth.CheckPanic(err)
				//l.Println("Downloading...")
				data, err := web.Get(authURL)
				oth.CheckPanic(err)
				//l.Println("Decoding...")

				if len(data) == 0 {
					//l.Println("Skipping")
					numSkippedDataTmp++
					if numSkippedDataTmp == tryH+5 {
						l.Println("Searching...")
						numSkippedDataTmp = 0
					}
					continue
				}
				tile, err := c3m.Parse(data)
				if err != nil {
					l.Println(err)
					continue
				}
				l.Println("Exporting", d1, d2, "h =", i)
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
	if err != nil {
		return
	}
	am, err := fly.GetAltitudeManifest(cache, ctx.ResourceManifest)
	if err != nil {
		return
	}
	m = context{ctx, am}
	return
}
