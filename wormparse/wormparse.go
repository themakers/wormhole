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
		Info    PackageInfo
		Imports []Import
		Types   []Type
		Methods []Method
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
		Type       Type
		UsePointer bool
		Signature  Function
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
		Type Type
	}

	Array struct {
		Len  int
		Type Type
	}

	Map struct {
		From Type
		To   Type
	}

	Pointer struct {
		Type Type
	}

	Function struct {
		Name   string // empty string means that function is anonymous
		Args   []NameTypePair
		Return []NameTypePair
	}

	NameTypePair struct {
		Name string
		Type Type
	}
	Interface struct {
		Methods []Function
	}
)
