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
		Receiver  interface{}
		Signature Function
	}
)

type (
	Type struct {
		Builtin    bool
		From       string
		Name       string
		Definition interface{}
	}

	Struct map[string]StructField

	StructField struct {
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
