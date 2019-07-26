package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/themakers/wormhole/parsex"
	"path/filepath"
)

func main()  {
	path, err := filepath.Abs("./pkg1/pkg1.go")
	if err != nil {
		panic(err)
	}

	p, err := parsex.Parse(path)
	if err != nil {
		panic(err)
	}

	println(spew.Sdump(p))
}
