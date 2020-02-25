package types

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
	return p.hash(map[*Definition]bool{})
}

func (p *Package) hash(_ map[*Definition]bool) string {
	return sum(sum("PACKAGE") + sum(p.Info.PkgPath))
}

func (p *Package) String() string {
	return stringify(p)
}
