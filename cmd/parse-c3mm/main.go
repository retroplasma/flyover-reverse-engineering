package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/retroplasma/flyover-reverse-engineering/pkg/fly/c3mm"
	"github.com/retroplasma/flyover-reverse-engineering/pkg/oth"
)

var l = log.New(os.Stderr, "", 0)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [c3mm_file]\n", os.Args[0])
		os.Exit(1)
	}

	partIfv1 := 0
	if len(os.Args) == 3 {
		var err error
		if partIfv1, err = strconv.Atoi(os.Args[2]); err != nil {
			panic(err)
		}
	}

	file := os.Args[1]
	data, err := ioutil.ReadFile(file)
	oth.CheckPanic(err)
	l.Printf("File size: %d bytes\n", len(data))
	c3mm, err := c3mm.Parse(data, partIfv1)
	oth.CheckPanic(err)
	_ = c3mm
}
