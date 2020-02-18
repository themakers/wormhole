package decparser

type Type interface {
	Hash() string
	String() string
}

type (
	Package struct {
		Info        PackageInfo
		Imports     []Import
		ImportsMap  map[PackageInfo]*Definition
		Definitions []*Definition
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
	return ""
}

func (p *Package) String() string {
	return ""
}

type Definition struct {
	Name        string
	Declaration Type
	Exported    bool
}

func (d *Definition) Hash() string {
	return ""
}

func (d *Definition) String() string {
	return ""
}

type (
	Struct struct {
		Fields    []*StructField
		FieldsMap map[string]*StructField
	}

	StructField struct {
		Name     string
		Tag      string
		Exported bool
		Type     Type
	}
)
