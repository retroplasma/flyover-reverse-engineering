package main

import (
	"fmt"
	"os"

	"github.com/flyover-reverse-engineering/pkg/mps/auth"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "usage: %s [url] [session_id] [token_1] [token_2]\n", os.Args[0])
		os.Exit(1)
	}
	res, err := auth.AuthURL(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s", res)
}
