package main

import (
	"flyover-reverse-engineering/pkg/fly"
	"flyover-reverse-engineering/pkg/oth"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var l = log.New(os.Stderr, "", 0)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [c3m_file]\n", os.Args[0])
		os.Exit(1)
	}

	file := os.Args[1]
	data, err := ioutil.ReadFile(file)
	oth.CheckPanic(err)
	l.Printf("File size: %d bytes\n", len(data))
	c3m := fly.ParseC3M(data)
	_ = c3m
}
