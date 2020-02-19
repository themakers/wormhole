package defparser

import (
	"fmt"

	"github.com/themakers/wormhole/defparser/types"
)

type (
	typeChecker struct {
		pkg       *types.Package
		usedNames map[string]struct{}
		global    *global
	}

	global struct {
		stdPkgs        map[types.PackageInfo]*types.Package
		pkgs           map[types.PackageInfo]*types.Package
		usedBuiltins   map[string]types.Builtin
		implicit       map[string]types.Type
		definitions    map[string]*types.Definition
		methods        map[string]*types.Method
		stdDefinitions map[string]*types.Definition
		// implementedMethods   map[string]*types.Method
		// implementedFunctions map[string]*types.Function // For the future
	}
)

func newTypeChecker() *typeChecker {
	return &typeChecker{
		global: &global{
			stdPkgs:        make(map[types.PackageInfo]*types.Package),
			pkgs:           make(map[types.PackageInfo]*types.Package),
			usedBuiltins:   make(map[string]types.Builtin),
			implicit:       make(map[string]types.Type),
			definitions:    make(map[string]*types.Definition),
			methods:        make(map[string]*types.Method),
			stdDefinitions: make(map[string]*types.Definition),
		},
		// methods: make(map[string]*types.Method),
		// importedDefinitions: make(map[impDef]*types.Definition),
	}
}

// func (tc *typeChecker) getRes

func (tc *typeChecker) newPackage(info types.PackageInfo, imports []types.Import) (*typeChecker, *types.Package) {
	importsMap := make(map[types.PackageInfo]types.Import)
	for _, imp := range imports {
		importsMap[imp.Package.Info] = imp
	}

	pkg := &types.Package{
		Info:           info,
		Imports:        imports,
		ImportsMap:     importsMap,
		DefinitionsMap: make(map[string]*types.Definition),
		MethodsMap:     make(map[string]*types.Method),
	}

	if _, ok := tc.global.pkgs[info]; ok {
		panic("WTF?")
	}
	tc.global.pkgs[info] = pkg

	return &typeChecker{
		usedNames: make(map[string]struct{}),
		pkg:       pkg,
		global:    tc.global,
		// methods:     make(map[string]*types.Method),
		// importedDefinitions: make(map[impDef]*types.Definition),
	}, pkg
}

func (tc *typeChecker) regSTDPkg(info types.PackageInfo) *types.Package {
	pkg := &types.Package{
		Info:           info,
		DefinitionsMap: make(map[string]*types.Definition),
	}
	if s, ok := tc.global.pkgs[info]; ok {
		return s
	}
	tc.global.stdPkgs[info] = pkg
	return pkg
}

func (tc *typeChecker) regBuiltin(b string) (types.Builtin, error) {
	t, err := types.String2Builtin(b)
	if err != nil {
		return types.Byte, err
	}
	tc.global.usedBuiltins[t.Hash()] = t
	return t, nil
}

func (tc *typeChecker) def(name string, declaration types.Type) (*types.Definition, error) {
	if _, ok := tc.usedNames[name]; ok {
		return nil, fmt.Errorf(
			"Duplicated identifier: %s",
			name,
		)
	}
	tc.usedNames[name] = struct{}{}

	def := &types.Definition{
		Name:        name,
		Declaration: declaration,
		Exported:    isExported(name),
		Package:     tc.pkg,
	}

	if d, ok := tc.global.definitions[def.Hash()]; ok {
		return d, nil
	}
	tc.global.definitions[def.Hash()] = def

	tc.pkg.Definitions = append(tc.pkg.Definitions, def)
	tc.pkg.DefinitionsMap[def.Name] = def

	return def, nil
}

func (tc *typeChecker) defRef(name, from string) (*types.Definition, error) {
	if !isExported(name) {
		return nil, fmt.Errorf(
			"STD definition cannot be unexported: %s.%s",
			from,
			name,
		)
	}

	var (
		pkgInfo *types.PackageInfo
	)

	for _, imp := range tc.pkg.Imports {
		if imp.Alias == from {
			pkgInfo = &imp.Package.Info
		}
	}
	if pkgInfo == nil {
		for _, imp := range tc.pkg.Imports {
			if imp.Package.Info.PkgName == from {
				pkgInfo = &imp.Package.Info
			}
		}
	}
	if pkgInfo != nil {
		if pkgInfo.Std {
			var (
				def = &types.Definition{
					Name:     name,
					Std:      true,
					Exported: true,
				}
				ok bool
			)
			def.Package, ok = tc.global.stdPkgs[*pkgInfo]
			if !ok {
				panic("RARARARA LATER")
			}
		} else {
			pkg, ok := tc.global.pkgs[*pkgInfo]
			if !ok {
				panic("OLOLOLO LATER")
			}
			def, ok := pkg.DefinitionsMap[name]
			if !ok {
				panic("TROLOLOLO LATER")
			}
			return def, nil
		}
	}

	return nil, fmt.Errorf(
		"Cannot identify imported definition: %s.%s",
		from,
		name,
	)
}

func (tc *typeChecker) meth(name string, t types.Type, f *types.Function) *types.Method {
	m := &types.Method{
		Name:      name,
		Type:      t,
		Signature: f,
	}

	if d, ok := tc.global.methods[m.Hash()]; ok {
		return d
	}
	tc.global.methods[m.Hash()] = m
	return m
}

func (tc *typeChecker) mkStructField(name, tag string, t types.Type) types.StructField {
	return types.StructField{
		Name:     name,
		Tag:      tag,
		Exported: isExported(name),
		Type:     t,
	}
}

func (tc *typeChecker) implChan(t types.Type) *types.Chan {
	c := &types.Chan{
		Type: t,
	}
	return tc.checkImplicit(c).(*types.Chan)
}

func (tc *typeChecker) implPtr(t types.Type) *types.Pointer {
	p := &types.Pointer{
		Type: t,
	}
	return tc.checkImplicit(p).(*types.Pointer)
}

func (tc *typeChecker) implStruct(fields []types.StructField) *types.Struct {
	fieldsMap := make(map[string]types.StructField)
	for _, field := range fields {
		fieldsMap[field.Name] = field
	}

	s := &types.Struct{
		Fields:    fields,
		FieldsMap: fieldsMap,
	}
	return tc.checkImplicit(s).(*types.Struct)
}

func (tc *typeChecker) implInter(methods []*types.Method) *types.Interface {
	i := &types.Interface{
		Methods: methods,
	}
	return tc.checkImplicit(i).(*types.Interface)
}

func (tc *typeChecker) implFunc(args []types.NameTypePair, results []types.NameTypePair) *types.Function {
	f := &types.Function{
		Args:    args,
		Results: results,
	}
	return tc.checkImplicit(f).(*types.Function)
}

func (tc *typeChecker) implMap(key, value types.Type) *types.Map {
	m := &types.Map{
		Key:   key,
		Value: value,
	}
	return tc.checkImplicit(m).(*types.Map)
}

func (tc *typeChecker) implSlice(t types.Type) *types.Slice {
	s := &types.Slice{
		Type: t,
	}
	return tc.checkImplicit(s).(*types.Slice)
}

func (tc *typeChecker) implArray(l int, t types.Type) *types.Array {
	a := &types.Array{
		Len:  l,
		Type: t,
	}
	return tc.checkImplicit(a).(*types.Array)
}

func (tc *typeChecker) checkImplicit(t types.Type) types.Type {
	if d, ok := tc.global.implicit[t.Hash()]; ok {
		return d
	}
	tc.global.implicit[t.Hash()] = t
	return t
}

func isExported(s string) bool {
	const (
		A = rune(65)
		Z = rune(90)
	)
	if l := rune(s[0]); l < A && l > Z {
		return false
	}
	return true
}
