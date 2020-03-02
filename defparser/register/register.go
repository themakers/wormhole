package register

import (
	"fmt"
	"sort"

	"github.com/themakers/wormhole/defparser/types"
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

func NewRegister(opts NewRegisterOpts) *Register {
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

	return &Register{
		ST: opts.State,
		Package: &types.Package{
			Info:           opts.Info,
			Imports:        imports,
			ImportsMap:     importsMap,
			DefinitionsMap: map[string]*types.Definition{},
			MethodsMap:     map[string]*types.Method{},
		},
	}
}

type State struct {
	TypesSet map[string]types.Type

	RootPackage types.PackageInfo
	PackagesMap map[string]*types.Package

	Result
}

func NewState(rootPackage types.PackageInfo) *State {
	return &State{
		TypesSet:    map[string]types.Type{},
		RootPackage: rootPackage,
		PackagesMap: map[string]*types.Package{},
	}
}

type Result struct {
	Root *types.Package

	Definitions    []*types.Definition
	STDDefinitions []*types.Definition

	Packages    []*types.Package
	STDPackages []*types.Package

	Methods []*types.Method
	Types   []types.Type
}

func (st *State) GetResult() (Result, error) {
	var ok bool
	st.Result.Root, ok = st.PackagesMap[st.RootPackage.PkgPath]
	if !ok {
		return Result{}, fmt.Errorf(
			"specified root package \"%s\" wasn't registered",
			st.RootPackage.PkgPath,
		)
	}

	for _, pkg := range st.PackagesMap {
		if pkg.Info.Std {
			st.Result.STDPackages = append(st.Result.STDPackages, pkg)
		} else {
			st.Result.Packages = append(st.Result.Packages, pkg)
		}

		for _, def := range pkg.Definitions {
			if def.Std {
				st.Result.STDDefinitions = append(
					st.Result.STDDefinitions,
					def,
				)
			} else {
				st.Result.Definitions = append(
					st.Result.Definitions,
					def,
				)
			}
		}
	}

	return st.Result
}
