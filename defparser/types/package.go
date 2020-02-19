package types

import "fmt"

var _ Type = &Package{}

type (
	Package struct {
		Info           PackageInfo
		Imports        []Import
		ImportsMap     map[PackageInfo]Import
		Definitions    []*Definition
		DefinitionsMap map[string]*Definition
		Methods        []*Method
		MethodsMap     map[string]*Method
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
)

func (p *Package) Hash() string {
	return hash(p.String())
}

const packageTmpl = "<%s>"

func (p *Package) String() string {
	return fmt.Sprintf(
		packageTmpl,
		p.Info.PkgPath,
	)
}
