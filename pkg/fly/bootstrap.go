package fly

import (
	"encoding/xml"
	"io/ioutil"
	"math"
	"os"
	"path"
	"regexp"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/mps"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/web"
)

type AltitudeManifest struct {
	XMLName  xml.Name  `xml:"manifest"`
	Triggers []Trigger `xml:"triggers>trigger"`
}

type Trigger struct {
	XMLName xml.Name `xml:"trigger"`
	Name    string   `xml:"name,attr"`
	LatRad  float64  `xml:"latitude,attr"`
	LonRad  float64  `xml:"longitude,attr"`
	Radius  float64  `xml:"radius,attr"`
	Region  int      `xml:"region,attr"`
	Version int      `xml:"version,attr"`

	Lat float64 // converted from LatRad
	Lon float64 // converted from LonRad
}

var regexAltitudeFile = regexp.MustCompile(`^altitude[a-zA-Z0-9-]*\.xml$`)

// GetAltitudeManifest finds altitude manifest reference in resource manifest
// and fetches its contents from cache or web and decodes it
func GetAltitudeManifest(cache mps.Cache, rm mps.ResourceManifest) (am AltitudeManifest, err error) {
	// find altitude file name
	altitudeFile, err := rm.CacheFileNameFromRegexp(regexAltitudeFile)
	if err != nil {
		return
	}

	// get altitude manifest
	rawAmCachePath := path.Join(cache.Directory, altitudeFile)
	var rawAm []byte
	if cache.Enabled {
		_, err = os.Stat(rawAmCachePath)
	}
	if !cache.Enabled || os.IsNotExist(err) {
		// from url
		if rawAm, err = web.Get(rm.CacheBaseUrl + "xml/" + altitudeFile); err != nil {
			return
		}
		if cache.Enabled {
			// to cache
			if err = ioutil.WriteFile(rawAmCachePath, rawAm, 0644); err != nil {
				return
			}
		}
	} else if err == nil {
		// from cache
		if rawAm, err = ioutil.ReadFile(rawAmCachePath); err != nil {
			return
		}
	} else {
		return
	}

	// decode altitude manifest
	am = AltitudeManifest{}
	if err = xml.Unmarshal(rawAm, &am); err != nil {
		return
	}

	// lat lon deg for convenience
	for i := range am.Triggers {
		t := &am.Triggers[i]
		t.Lat = t.LatRad / math.Pi * 180
		t.Lon = t.LonRad / math.Pi * 180
	}

	return
}
