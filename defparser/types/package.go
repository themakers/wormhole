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
	return p.hash(nil)
}

func (p *Package) hash(_ map[*Definition]bool) string {
	if p.Info.Std {
		return sum(sum("PACKAGE") + sum("STD") + sum(p.Info.PkgPath))
	}

	s := sum("PACKAGE") + sum(p.Info.PkgPath)

	for _, imp := range p.Imports {
		s += imp.Package.Hash()
	}

	for _, def := range p.Definitions {
		s += def.Hash()
	}

	for _, meth := range p.Methods {
		s += meth.Hash()
	}

	return sum(s)
}

func (p *Package) String() string {
	return stringify(p)
}
