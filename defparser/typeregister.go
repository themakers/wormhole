package defparser

import (
	"fmt"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"github.com/themakers/wormhole/defparser/types"
)

type (
	typeRegister struct {
		pkg                  *types.Package
		usedNames            map[string]struct{}
		global               *global
		undefined            map[string]*undefined
		undefinedIdentifiers map[string]*types.Definition
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

func newTypeRegister() *typeRegister {
	return &typeRegister{
		global: &global{
			stdPkgs:        map[types.PackageInfo]*types.Package{},
			pkgs:           map[types.PackageInfo]*types.Package{},
			usedBuiltins:   map[string]types.Builtin{},
			implicit:       map[string]types.Type{},
			definitions:    map[string]*types.Definition{},
			methods:        map[string]*types.Method{},
			stdDefinitions: map[stdDefKey]*types.Definition{},
		},
		// methods: make(map[string]*types.Method),
		// importedDefinitions: make(map[impDef]*types.Definition),
	}
}

func (tr *typeRegister) getResult(rootPkg types.PackageInfo) (*Result, error) {
	root, ok := tr.global.pkgs[rootPkg]
	if !ok {
		return nil, fmt.Errorf(
			"Cannot find root package %s",
			spew.Sdump(rootPkg),
		)
	}

	res := &Result{
		Root:           root,
		Definitions:    make([]*types.Definition, len(tr.global.definitions)),
		STDDefinitions: make([]*types.Definition, len(tr.global.stdDefinitions)),
		Packages:       make([]*types.Package, len(tr.global.pkgs)),
		STDPackages:    make([]*types.Package, len(tr.global.stdPkgs)),
		Methods:        make([]*types.Method, len(tr.global.methods)),
		Types:          make([]types.Type, len(tr.global.implicit)),
	}

	{
		var i int
		for _, pkg := range tr.global.stdPkgs {
			res.STDPackages[i] = pkg
			i++
		}
	}
	{
		var i int
		for _, def := range tr.global.stdDefinitions {
			res.STDDefinitions[i] = def
			i++
		}
	}
	{
		var i int
		for _, pkg := range tr.global.pkgs {
			res.Packages[i] = pkg
			i++
		}
	}
	{
		var i int
		for _, def := range tr.global.definitions {
			res.Definitions[i] = def
			i++
		}
	}
	{
		var i int
		for _, impl := range tr.global.implicit {
			res.Types[i] = impl
			i++
		}
	}
	{
		var i int
		for _, meth := range tr.global.methods {
			res.Methods[i] = meth
			i++
		}
	}
	{
		var (
			i        int
			builtins = make([]types.Type, len(tr.global.usedBuiltins))
		)
		for _, b := range tr.global.usedBuiltins {
			builtins[i] = b
			i++
		}
		res.Types = append(builtins, res.Types...)
	}

	return res, nil
}

func (tr *typeRegister) newPackage(
	info types.PackageInfo,
	imports []types.Import,
) *typeRegister {
	pkg, ok := tr.global.pkgs[info]
	if !ok {
		importsMap := make(map[types.PackageInfo]types.Import)
		for _, imp := range imports {
			if imp.Package.Info.Std {
				if s, ok := tr.global.stdPkgs[imp.Package.Info]; ok {
					imp.Package = s
				} else {
					tr.global.stdPkgs[imp.Package.Info] = imp.Package
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
		tr.global.pkgs[info] = pkg
	} else {
		fmt.Printf(
			"WARNING: something went strange: %s",
			spew.Sdump(info),
		)
	}

	return &typeRegister{
		undefinedIdentifiers: make(map[string]*types.Definition),
		usedNames:            make(map[string]struct{}),
		pkg:                  pkg,
		global:               tr.global,
	}
}

func (tr *typeRegister) regBuiltin(b string) (types.Builtin, error) {
	t, err := types.String2Builtin(b)
	if err != nil {
		return types.Byte, err
	}
	tr.global.usedBuiltins[t.Hash()] = t
	return t, nil
}

func (tr *typeRegister) def(
	name string,
	declaration types.Type,
) (*types.Definition, error) {
	if def, ok := tr.undefinedIdentifiers[name]; ok {
		def.Declaration = declaration
		delete(tr.undefinedIdentifiers, name)
		return def, nil
	}

	if _, ok := tr.usedNames[name]; ok {
		return nil, fmt.Errorf(
			"Duplicated identifier: %s",
			name,
		)
	}
	tr.usedNames[name] = struct{}{}

	def := &types.Definition{
		Name:        name,
		Declaration: declaration,
		Exported:    isExported(name),
		Package:     tr.pkg,
	}

	if def, ok := tr.global.definitions[def.Hash()]; ok {
		return def, nil
	}
	tr.global.definitions[def.Hash()] = def

	tr.pkg.Definitions = append(tr.pkg.Definitions, def)
	tr.pkg.DefinitionsMap[def.Name] = def

	return def, nil
}

func (tr *typeRegister) defRef(name, from string) (*types.Definition, error) {
	if !isExported(name) {
		return nil, fmt.Errorf(
			"STD definition cannot be unexported: %s.%s",
			from,
			name,
		)
	}

	var pkg *types.Package
	if from != "" {
		{
			var pkgInfo *types.PackageInfo
			for _, imp := range tr.pkg.Imports {
				if imp.Alias == from {
					pkgInfo = &imp.Package.Info
				}
			}
			if pkgInfo == nil {
				for _, imp := range tr.pkg.Imports {
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
			pkg, ok = tr.global.pkgs[*pkgInfo]
			if !ok {
				pkg, ok = tr.global.stdPkgs[*pkgInfo]
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
			if s, ok := tr.global.stdDefinitions[stdDefKey]; ok {
				return s, nil
			}
			def := &types.Definition{
				Std:         true,
				Exported:    true,
				Package:     pkg,
				Name:        name,
				Declaration: types.Untyped,
			}
			tr.global.stdDefinitions[stdDefKey] = def
			return def, nil
		}
	} else {
		pkg = tr.pkg
	}

	if def, ok := pkg.DefinitionsMap[name]; ok {
		return def, nil
	} else if pkg == tr.pkg {
		if def, ok := tr.undefinedIdentifiers[name]; ok {
			return def, nil
		}
		var err error
		def, err = tr.def(name, types.Untyped)
		if err != nil {
			return nil, err
		}
		tr.undefinedIdentifiers[name] = def
		return def, nil
	}

	return nil, fmt.Errorf(
		"Cannot find definition for imported identifier: %s.%s",
		from,
		name,
	)
}

func (tr *typeRegister) meth(
	name string,
	receiver types.Type,
	signature *types.Function,
) *types.Method {

	m := &types.Method{
		Name:      name,
		Receiver:  receiver,
		Signature: signature,
	}

	if d, ok := tr.global.methods[m.Hash()]; ok {
		return d
	}
	tr.global.methods[m.Hash()] = m
	return m
}

func (tr *typeRegister) mkStructField(
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

func (tr *typeRegister) implChan(t types.Type) *types.Chan {
	c := &types.Chan{
		Type: t,
	}
	return tr.checkImplicit(c).(*types.Chan)
}

func (tr *typeRegister) implPtr(t types.Type) *types.Pointer {
	p := &types.Pointer{
		Type: t,
	}
	return tr.checkImplicit(p).(*types.Pointer)
}

func (tr *typeRegister) implStruct(fields []types.StructField) *types.Struct {
	fieldsMap := make(map[string]types.StructField)
	for _, field := range fields {
		fieldsMap[field.Name] = field
	}

	s := &types.Struct{
		Fields:    fields,
		FieldsMap: fieldsMap,
	}
	return tr.checkImplicit(s).(*types.Struct)
}

func (tr *typeRegister) implInter(methods []*types.Method) *types.Interface {
	i := &types.Interface{
		Methods: methods,
	}
	return tr.checkImplicit(i).(*types.Interface)
}

func (tr *typeRegister) implFunc(
	args []types.NameTypePair,
	results []types.NameTypePair,
) *types.Function {

	f := &types.Function{
		Args:    args,
		Results: results,
	}
	return tr.checkImplicit(f).(*types.Function)
}

func (tr *typeRegister) implMap(key, value types.Type) *types.Map {
	m := &types.Map{
		Key:   key,
		Value: value,
	}
	return tr.checkImplicit(m).(*types.Map)
}

func (tr *typeRegister) implSlice(t types.Type) *types.Slice {
	s := &types.Slice{
		Type: t,
	}
	return tr.checkImplicit(s).(*types.Slice)
}

func (tr *typeRegister) implArray(l int, t types.Type) *types.Array {
	a := &types.Array{
		Len:  l,
		Type: t,
	}
	return tr.checkImplicit(a).(*types.Array)
}

func (tr *typeRegister) checkImplicit(t types.Type) types.Type {
	if d, ok := tr.global.implicit[t.Hash()]; ok {
		return d
	}
	tr.global.implicit[t.Hash()] = t
	return t
}

func isExported(s string) bool {
	return unicode.IsLetter(rune(s[0])) &&
		unicode.IsUpper(rune(s[0]))
}
