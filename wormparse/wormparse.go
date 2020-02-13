package wormparse

import (
	"os"
	"path/filepath"
)

var (
	GOROOT = filepath.Clean(os.Getenv("GOROOT"))
	GOPATH = filepath.Clean(os.Getenv("GOPATH"))
	GOSRC  = filepath.Join(GOPATH, "src")
	GOSTD  = filepath.Join(GOROOT, "src")
)

type (
	Package struct {
		Info       PackageInfo
		Imports    []Import
		Types      []Type
		Interfaces []Interface
	}

	PackageInfo struct {
		PkgName     string
		PkgPath     string
		PkgFullPath string
		Std         bool
	}

	Import struct {
		Package *Package
		Alias   string
	}

	Type struct{}

	Interface struct{}
)
