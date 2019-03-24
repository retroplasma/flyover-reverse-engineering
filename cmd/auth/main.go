package main

import (
	"fmt"
	"os"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/mps/config"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/mps"
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
	if !config.IsValid() {
		fmt.Fprintln(os.Stderr, "please set values in config.json")
		os.Exit(1)
	}
	if err = cache.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing cache: %v\n", err)
		os.Exit(1)
	}
	ctx, err := mps.Init(cache, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting context: %v\n", err)
		os.Exit(1)
	}
	res, err := ctx.AuthContext.AuthURL(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error authenticating url: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s", res)
}
