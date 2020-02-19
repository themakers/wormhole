package defparser

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
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
		stdDefinitions map[stdDefKey]*types.Definition
		// implementedMethods   map[string]*types.Method
		// implementedFunctions map[string]*types.Function // For the future
	}

	stdDefKey struct {
		name    string
		pkgInfo types.PackageInfo
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
			stdDefinitions: make(map[stdDefKey]*types.Definition),
		},
		// methods: make(map[string]*types.Method),
		// importedDefinitions: make(map[impDef]*types.Definition),
	}
}

func (tc *typeChecker) getResult() *Result {
	res := &Result{
		Definitions:    make([]*types.Definition, len(tc.global.definitions)),
		STDDefinitions: make([]*types.Definition, len(tc.global.stdDefinitions)),
		Packages:       make([]*types.Package, len(tc.global.pkgs)),
		STDPackages:    make([]*types.Package, len(tc.global.stdPkgs)),
		Methods:        make([]*types.Method, len(tc.global.methods)),
		Types:          make([]types.Type, len(tc.global.implicit)),
	}

	{
		var i int
		for _, pkg := range tc.global.stdPkgs {
			res.STDPackages[i] = pkg
			i++
		}
	}
	{
		var i int
		for _, def := range tc.global.stdDefinitions {
			res.STDDefinitions[i] = def
			i++
		}
	}
	{
		var i int
		for _, pkg := range tc.global.pkgs {
			res.Packages[i] = pkg
			i++
		}
	}
	{
		var i int
		for _, def := range tc.global.definitions {
			res.Definitions[i] = def
			i++
		}
	}
	{
		var i int
		for _, impl := range tc.global.implicit {
			res.Types[i] = impl
			i++
		}
	}
	{
		var i int
		for _, meth := range tc.global.methods {
			res.Methods[i] = meth
			i++
		}
	}

	return res
}

func (tc *typeChecker) newPackage(
	info types.PackageInfo,
	imports []types.Import,
) *typeChecker {
	pkg, ok := tc.global.pkgs[info]
	if !ok {
		importsMap := make(map[types.PackageInfo]types.Import)
		for _, imp := range imports {
			if imp.Package.Info.Std {
				if s, ok := tc.global.stdPkgs[imp.Package.Info]; ok {
					imp.Package = s
				} else {
					tc.global.stdPkgs[imp.Package.Info] = imp.Package
				}
			}
			importsMap[imp.Package.Info] = imp
		}

		pkg = &types.Package{
			Info:           info,
			Imports:        imports,
			ImportsMap:     importsMap,
			DefinitionsMap: make(map[string]*types.Definition),
			MethodsMap:     make(map[string]*types.Method),
		}
		tc.global.pkgs[info] = pkg
	} else {
		fmt.Printf(
			"WARNING: something went strange: %s",
			spew.Sdump(info),
		)
	}

	return &typeChecker{
		usedNames: make(map[string]struct{}),
		pkg:       pkg,
		global:    tc.global,
	}
}

func (tc *typeChecker) regBuiltin(b string) (types.Builtin, error) {
	t, err := types.String2Builtin(b)
	if err != nil {
		return types.Byte, err
	}
	tc.global.usedBuiltins[t.Hash()] = t
	return t, nil
}

func (tc *typeChecker) def(name string, declaration types.Type) error {
	if _, ok := tc.usedNames[name]; ok {
		return fmt.Errorf(
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

	if _, ok := tc.global.definitions[def.Hash()]; ok {
		return nil
	}
	tc.global.definitions[def.Hash()] = def

	tc.pkg.Definitions = append(tc.pkg.Definitions, def)
	tc.pkg.DefinitionsMap[def.Name] = def

	return nil
}

func (tc *typeChecker) defRef(name, from string) (*types.Definition, error) {
	if !isExported(name) {
		return nil, fmt.Errorf(
			"STD definition cannot be unexported: %s.%s",
			from,
			name,
		)
	}

	var pkg *types.Package
	{
		var pkgInfo *types.PackageInfo
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

			if pkgInfo == nil {
				return nil, fmt.Errorf(""+
					"There's no package that fits "+
					"imported identifier: %s.%s",
					from,
					name,
				)
			}
		}

		var ok bool
		pkg, ok = tc.global.pkgs[*pkgInfo]
		if !ok {
			pkg, ok = tc.global.stdPkgs[*pkgInfo]
			if !ok {
				panic(fmt.Errorf(""+
					"TypeChecker: Imports and STD package buffer "+
					"are desinchronized: %s",
					spew.Sdump(pkgInfo),
				))
			}
		}
	}

	if pkg.Info.Std {
		stdDefKey := stdDefKey{
			name:    name,
			pkgInfo: pkg.Info,
		}
		if s, ok := tc.global.stdDefinitions[stdDefKey]; ok {
			return s, nil
		}
		def := &types.Definition{
			Std:      true,
			Exported: true,
			Package:  pkg,
			Name:     name,
		}
		tc.global.stdDefinitions[stdDefKey] = def
		return def, nil
	}

	if def, ok := pkg.DefinitionsMap[name]; ok {
		return def, nil
	}

	return nil, fmt.Errorf(
		"Cannot definition for imported identifier: %s.%s",
		from,
		name,
	)
}

func (tc *typeChecker) meth(
	name string,
	t types.Type,
	f *types.Function,
) *types.Method {

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

func (tc *typeChecker) mkStructField(
	name,
	tag string,
	t types.Type,
) types.StructField {

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

func (tc *typeChecker) implFunc(
	args []types.NameTypePair,
	results []types.NameTypePair,
) *types.Function {

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
	// FIXME
	// unicode.IsLetter()
	// unicode.IsUpper()
	return true
}
