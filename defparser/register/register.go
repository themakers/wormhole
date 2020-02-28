package register

import (
	"sort"

	"github.com/themakers/wormhole/defparser/types"
)

type (
	State map[string]types.Type

	Result struct {
		Root *types.Package

		Definitions    []*types.Definition
		STDDefinitions []*types.Definition

		Packages    []*types.Package
		STDPackages []*types.Package

		Methods []*types.Method
		Types   []types.Type
	}
)

type Register struct {
	ST State
	*types.Package
}

type NewRegisterOpts struct {
	State   State
	Info    types.PackageInfo
	Imports []types.Import
}

func New(opts NewRegisterOpts) *Register {
	var (
		imports    = make([]types.Import, len(opts.Imports))
		importsMap = map[string]types.Import{}
	)

	for i, imp := range opts.Imports {
		imports[i] = imp
	}
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Package.Info.PkgPath <
			imports[j].Package.Info.PkgPath
	})

	for _, imp := range imports {
		if imp.Alias != "" {
			importsMap[imp.Alias] = imp
		} else {
			importsMap[imp.Package.Info.PkgName] = imp
		}
	}

	pkg := &types.Package{
		Info:           opts.Info,
		Imports:        imports,
		ImportsMap:     importsMap,
		DefinitionsMap: map[string]*types.Definition{},
		MethodsMap:     map[string]*types.Method{},
	}

	return &Register{
		ST:      opts.State,
		Package: pkg,
	}
}
