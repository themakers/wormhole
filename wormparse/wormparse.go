package wormparse

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	GOROOT = filepath.Clean(os.Getenv("GOROOT"))
	GOPATH = filepath.Clean(os.Getenv("GOPATH"))
	GOSRC  = filepath.Join(GOPATH, "src")
	GOSTD  = filepath.Join(GOROOT, "src")
)

type Hashable interface {
	Hash() string
}

func hash1(input string) string {
	return ""
}

type ParseFunc func(pkgPath string) (Result, *Package)

type Result struct {
	Types map[string]*Type

	Packages []*Package
}

func (res *Result) SortedTypes() []*Type {
	// TODO Return types slice sorted by hash
	return nil
}

type (
	Package struct {
		Doc string

		Info      PackageInfo
		Imports   []Import
		Types     []Type
		Methods   []*Method
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

var _ Hashable = new(Type)

type Type struct {
	Package *Package

	Builtin bool

	Name       string
	Definition interface{}

	Methods []*Method
}

func (t *Type) Hash() string {
}

var _ Hashable = new(Map)

type Map struct {
	Key   Hashable
	Value Hashable
}

func (m Map) Hash() string {
	return hash1(fmt.Sprintf("map:6:%s:%s", m.Key.Hash(), m.Value.Hash()))
}

type Function struct {
	Name   string // empty string means that function is anonymous
	Type   *Type
	Args   []NameTypePair
	Return []NameTypePair
}

func (f *Function) Hash() string {
	// Don't take name into account
	return hash1(fmt.Sprintf("function:%s:%s", m.Key.Hash(), m.Value.Hash()))
}

type TypeDefinition struct {
	Package *Package
	Name    string
	Type    *Type
}

func (td *TypeDefinition) Hash() string {
	// Don't take name into account
	return hash1(fmt.Sprintf("function:%s:%s:%s", td.Name, td.Package.Info.PkgName, td.Type.Hash()))
}

type (
	Struct struct {
		Fields    []*StructField
		FieldsMap map[string]*StructField
	}

	StructField struct {
		Name string
		Tag  string
		Type interface{}

		Exported bool
	}

	Slice struct {
		Type interface{}
	}

	Array struct {
		Len  int
		Type interface{}
	}

	Pointer struct {
		Type interface{}
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
