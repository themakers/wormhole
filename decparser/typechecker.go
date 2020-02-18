package decparser

import (
	"github.com/themakers/wormhole/decparser/types"
)

type globalScope struct {
	pkgs        map[types.PackageInfo]*types.Package
	implicit    map[string]types.Type
	definitions map[string]*types.Definition
	methods     map[string]*types.Method
	defs        map[string]*types.Definition
	// implementedMethods   map[string]*types.Method
	// implementedFunctions map[string]*types.Function // For the future
}
type typeChecker struct {
	pkg    *types.Package
	global *globalScope
}

func newTypeChecker() *typeChecker {
	return &typeChecker{
		global: &globalScope{
			pkgs:        make(map[types.PackageInfo]*types.Package),
			definitions: make(map[string]*types.Definition),
		},
		// methods: make(map[string]*types.Method),
		// importedDefinitions: make(map[impDef]*types.Definition),
	}
}

type impDef struct {
	From types.PackageInfo
	Name string
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
		pkg:    pkg,
		global: tc.global,
		// methods:     make(map[string]*types.Method),
		// importedDefinitions: make(map[impDef]*types.Definition),
	}, pkg
}

func (tc *typeChecker) def(name string, declaration types.Type) *types.Definition {
	def := &types.Definition{
		Name:        name,
		Declaration: declaration,
		Exported:    isExported(name),
		Package:     tc.pkg,
	}

	d, ok := tc.global.definitions[def.Hash()]
	if !ok {
		tc.global.definitions[def.Hash()] = def
		d = def
	}

	return d
}

func (tc *typeChecker) defRef(name, from string) (*types.Definition, error) {
	def := &types.Definition{
		Name:        name,
		Declaration: declaration,
		Exported:    isExported(name),
	}
	// for _, imp := range tc.pkg.Imports {
	// 	if imp.Alias == from || imp.Package.Info.PkgName == from {
	// 		def.Package = imp.Package
	// 	}
	// }

	if def.Package == nil {
		panic("TROLOLO OLOLO")
	}

	d, ok := tc.global.definitions[def.Hash()]
	if !ok {
		tc.global.definitions[def.Hash()] = def
		d = def
	}

	return d
}

func (tc *typeChecker) meth(name string, t types.Type, f *types.Function) *types.Method {
	m := &types.Method{
		Name:      name,
		Type:      t,
		Signature: f,
	}

	d, ok := tc.global.methods[m.Hash()]
	if !ok {
		tc.global.methods[m.Hash()] = m
		d = m
	}

	return d
}

func (tc *typeChecker) mkStructField(name, tag string, t types.Type) types.StructField {
	return types.StructField{
		Name:     name,
		Tag:      tag,
		Exported: isExported(name),
		Type:     t,
	}
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
	tc.global.implicit[s.Hash()] = s
	return s
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
	d, ok := tc.global.implicit[t.Hash()]
	if !ok {
		tc.global.implicit[t.Hash()] = d
		d = t
	}
	return d
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
