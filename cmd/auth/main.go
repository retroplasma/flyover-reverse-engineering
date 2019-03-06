package main

import (
	"fmt"
	"os"

	"github.com/flyover-reverse-engineering/pkg/mps/config"

	"github.com/flyover-reverse-engineering/pkg/mps"
	"github.com/flyover-reverse-engineering/pkg/mps/bootstrap"
)

var cache = mps.Cache{Enabled: true, Directory: "./cache"}
var configFile = "./config.json"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [url]\n", os.Args[0])
		os.Exit(1)
	}
	config, err := config.FromJSONFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config file: %v\n", err)
		os.Exit(1)
	}
	if config.ResourceManifestURL == "" || config.TokenP1 == "" {
		fmt.Fprintln(os.Stderr, "please set values in config.json")
		os.Exit(1)
	}
	if err = cache.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing cache: %v\n", err)
		os.Exit(1)
	}
	session, err := bootstrap.GetSession(cache)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting session: %v\n", err)
		os.Exit(1)
	}
	rm, err := bootstrap.GetResourceManifest(cache, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting resource manifest: %v\n", err)
		os.Exit(1)
	}
	authCtx := mps.AuthContext{Session: session, ResourceManifest: rm, TokenP1: mps.TokenP1(config.TokenP1)}
	res, err := authCtx.AuthURL(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error authenticating url: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s", res)
}
