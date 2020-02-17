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
		Info      PackageInfo
		Imports   []Import
		Types     []Type
		Methods   []Method
		Functions []Function
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

	Method struct {
		Name      string
		Reciever  interface{}
		Signature Function
	}
)

type (
	Type struct {
		Basic      bool
		From       string
		Name       string
		Definition interface{}
	}

	Struct map[string]TagTypePair

	TagTypePair struct {
		Tag  string
		Type interface{}
	}

	Slice struct {
		Type interface{}
	}

	Array struct {
		Len  int
		Type interface{}
	}

	Map struct {
		Key   interface{}
		Value interface{}
	}

	Pointer struct {
		Type interface{}
	}

	Function struct {
		Name   string // empty string means that function is anonymous
		Args   []NameTypePair
		Return []NameTypePair
	}

	NameTypePair struct {
		Name string
		Type interface{}
	}
	Interface struct {
		Methods []Function
	}

	Chan struct {
		Type interface{}
	}
)
