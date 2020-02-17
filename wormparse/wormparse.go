package wormparse

import (
	"os"
	"path/filepath"

	"github.com/themakers/wormhole/wormparse/datatypes"
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
		Std        bool
		From       string
		Name       string
		DataType   datatypes.DataType
		Definition interface{}
	}

	Struct map[string]struct {
		Tag  string
		Type Type
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
		Name string // empty string means that function is anonymous

		Args []NameTypePair

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
