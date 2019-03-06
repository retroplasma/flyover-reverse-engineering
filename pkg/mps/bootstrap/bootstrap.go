package bootstrap

import (
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/flyover-reverse-engineering/pkg/mps"
	"github.com/flyover-reverse-engineering/pkg/mps/auth"
	"github.com/flyover-reverse-engineering/pkg/mps/config"
	"github.com/flyover-reverse-engineering/pkg/mps/pro"
	"github.com/flyover-reverse-engineering/pkg/web"
	"github.com/golang/protobuf/proto"
)

// GetSession creates a new session or fetches it from cache
func GetSession(cache mps.Cache) (s mps.Session, err error) {
	rawSidCachePath := path.Join(cache.Directory, "session.txt")
	var rawSid []byte
	var info os.FileInfo
	if cache.Enabled {
		info, err = os.Stat(rawSidCachePath)
	}
	if !cache.Enabled || os.IsNotExist(err) || err == nil && time.Now().Sub(info.ModTime()).Hours() > 24 {
		// from generator
		rawSid = []byte(auth.GenRandStr(40, "0123456789"))
		if cache.Enabled {
			// to cache
			if err = ioutil.WriteFile(rawSidCachePath, rawSid, 0644); err != nil {
				return
			}
		}
	} else if err == nil {
		// from cache
		if rawSid, err = ioutil.ReadFile(rawSidCachePath); err != nil {
			return
		}
	} else {
		return
	}
	s = mps.Session{ID: string(rawSid)}
	return
}

// GetResourceManifest fetches resource manifest from cache or web and decodes it
func GetResourceManifest(cache mps.Cache, config config.Config) (rm pro.ResourceManifest, err error) {
	rawRmCachePath := path.Join(cache.Directory, "ResourceManifest.pbd")
	var rawRm []byte
	var info os.FileInfo
	if cache.Enabled {
		info, err = os.Stat(rawRmCachePath)
	}
	if !cache.Enabled || os.IsNotExist(err) || err == nil && time.Now().Sub(info.ModTime()).Hours() > 1 {
		// from url
		if rawRm, err = web.Get(config.ResourceManifestURL); err != nil {
			return
		}
		if cache.Enabled {
			// to cache
			if err = ioutil.WriteFile(rawRmCachePath, rawRm, 0644); err != nil {
				return
			}
		}
	} else if err == nil {
		// from cache
		if rawRm, err = ioutil.ReadFile(rawRmCachePath); err != nil {
			return
		}
	} else {
		return
	}

	// decode resource manifest
	rm = pro.ResourceManifest{}
	if err = proto.Unmarshal(rawRm, &rm); err != nil {
		return
	}

	return
}
