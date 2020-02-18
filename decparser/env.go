package decparser

import (
	"os"
	"path/filepath"
)

var (
	GOROOT string
	GOPATH string
	GOSRC  string
	GOSTD  string
)

var _ struct{} = env()

func env() struct{} {
	GOROOT = filepath.Clean(os.Getenv("GOROOT"))
	if GOROOT == "" {
		panic(ErrEnvVarNotSetGOROOT)
	}
	GOPATH = filepath.Clean(os.Getenv("GOPATH"))
	if GOPATH == "" {
		panic(ErrEnvVarNotSetGOPATH)
	}
	GOSRC = filepath.Join(GOPATH, "src")
	GOSTD = filepath.Join(GOROOT, "src")

	return struct{}{}
}
