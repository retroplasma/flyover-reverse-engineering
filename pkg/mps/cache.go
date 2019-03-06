package mps

import "os"

type Cache struct {
	Enabled   bool
	Directory string
}

func (cache Cache) Init() error {
	return os.MkdirAll(cache.Directory, 0755)
}
