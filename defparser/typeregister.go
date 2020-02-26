package defparser

import (
	"fmt"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"github.com/themakers/wormhole/defparser/types"
)

type (
	typeRegister struct {
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
			DefinitionsMap: map[string]*types.Definition{},
			MethodsMap:     map[string]*types.Method{},
		}
		tr.global.pkgs[info] = pkg
	} else {
		fmt.Printf(
			"WARNING: something went strange: %s",
			spew.Sdump(info),
		)
	}

	return &typeRegister{
		usedNames: map[string]struct{}{},
		pkg:       pkg,
		global:    tr.global,
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

func (tr *typeRegister) define(
	name string,
) (*types.Definition, error) {
	if _, ok := tr.usedNames[name]; ok {
		return nil, fmt.Errorf(
			"Duplicated identifier: %s",
			name,
		)
	}
	tr.usedNames[name] = struct{}{}

	def := &types.Definition{
		Name:       name,
		Exported:   isExported(name),
		Package:    tr.pkg,
		Methods:    []*types.Method{},
		MethodsMap: map[string]*types.Method{},
	}

	if def, ok := tr.global.definitions[def.Hash()]; ok {
		return def, nil
	}
	tr.global.definitions[def.Hash()] = def

	tr.pkg.Definitions = append(tr.pkg.Definitions, def)
	tr.pkg.DefinitionsMap[def.Name] = def

	return def, nil
}

func (tr *typeRegister) definitionRef(
	name,
	from string,
) (*types.Definition, error) {

	var pkg *types.Package
	if from != "" {
		if !isExported(name) {
			return nil, fmt.Errorf(
				"definition from another package cannot be unexported: %s.%s",
				from,
				name,
			)
		}

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
	}

	return nil, fmt.Errorf(
		"Cannot find definition for imported identifier: %s.%s",
		from,
		name,
	)
}

func (tr *typeRegister) method(
	name string,
	receiver *types.Definition,
	signature *types.Function,
) error {
	if receiver.Package == tr.pkg {
		return fmt.Errorf(
			"can't define method \"%s\" for type \"%s\" defined another package \"%s\"",
			name,
			receiver.Name,
			receiver.Package.Info.PkgPath,
		)
	}

	if _, ok := receiver.MethodsMap[name]; ok {
		return fmt.Errorf(
			"method \"%s\" were defined on type \"%s\"before in package %s",
			name,
			receiver.Name,
			receiver.Package.Info.PkgPath,
		)
	}

	m := &types.Method{
		Name:      name,
		Receiver:  receiver,
		Signature: signature,
	}

	receiver.Methods = append(receiver.Methods, m)
	receiver.MethodsMap[name] = m
	tr.pkg.Methods = append(tr.pkg.Methods, m)
	tr.pkg.MethodsMap[name] = m
	tr.global.methods[m.Hash()] = m

	return nil
}

func (tr *typeRegister) implMethod(
	name string,
	signature *types.Function,
) (*types.Method, error) {
	return &types.Method{
		Name:      name,
		Signature: signature,
	}, nil
}

func (tr *typeRegister) mkStructField(
	name,
	tag string,
	t types.Type,
) *types.StructField {
	field := &types.StructField{
		Name:     name,
		Tag:      tag,
		Exported: isExported(name),
		Type:     t,
	}

	return field
}

func (tr *typeRegister) mkNameTypePair(
	name string,
	t types.Type,
) *types.NameTypePair {
	pair := &types.NameTypePair{
		Name: name,
		Type: t,
	}

	return pair
}

func (tr *typeRegister) implChan(t types.Type) *types.Chan {
	return tr.checkImplicit(&types.Chan{
		Type: t,
	}).(*types.Chan)
}

func (tr *typeRegister) implPtr(t types.Type) *types.Pointer {
	return tr.checkImplicit(&types.Pointer{
		Type: t,
	}).(*types.Pointer)
}

func (tr *typeRegister) implStruct(fields []*types.StructField) *types.Struct {
	fieldsMap := make(map[string]*types.StructField)
	for _, field := range fields {
		fieldsMap[field.Name] = field
	}
	return tr.checkImplicit(&types.Struct{
		Fields:    fields,
		FieldsMap: fieldsMap,
	}).(*types.Struct)
}

func (tr *typeRegister) implInter(methods []*types.Method) *types.Interface {
	return tr.checkImplicit(&types.Interface{
		Methods: methods,
	}).(*types.Interface)
}

func (tr *typeRegister) implFunc(
	args []*types.NameTypePair,
	results []*types.NameTypePair,
) *types.Function {

	return tr.checkImplicit(&types.Function{
		Args:    args,
		Results: results,
	}).(*types.Function)
}

func (tr *typeRegister) implMap(key, value types.Type) *types.Map {
	return tr.checkImplicit(&types.Map{
		Key:   key,
		Value: value,
	}).(*types.Map)
}

func (tr *typeRegister) implSlice(t types.Type) *types.Slice {
	return tr.checkImplicit(&types.Slice{
		Type: t,
	}).(*types.Slice)
}

func (tr *typeRegister) implArray(l int, t types.Type) *types.Array {
	return tr.checkImplicit(&types.Array{
		Len:  l,
		Type: t,
	}).(*types.Array)
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
