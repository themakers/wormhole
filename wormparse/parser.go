package wormparse

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

func Parse(pkgPath string) (*Package, error) {
	var (
		index          int
		parsedPackages = make(map[PackageInfo]*Package)
		parse          func(pkgFullPath, pkgPath string, prev map[string]int) (*Package, error)
	)

	parse = func(pkgFullPath, pkgPath string, prev map[string]int) (*Package, error) {
		pkgFullPath = filepath.Clean(pkgFullPath)
		if !filepath.IsAbs(pkgFullPath) {
			return nil, ErrNotAbsoluteFilePath
		}

		pkgIndx, ok := prev[pkgFullPath]
		if ok {
			return nil, Loop{
				index: pkgIndx,
			}
		}

		var res Package
		m := make(map[string]int)
		{
			pkgIndx = index
			index++
			for k, v := range prev {
				m[k] = v
			}
			m[pkgFullPath] = pkgIndx
		}

		pkgs, err := parser.ParseDir(
			token.NewFileSet(),
			pkgFullPath,
			nil,
			0,
		)
		if err != nil {
			return nil, err
		}

		var (
			pkgName string
			stdLibs = make(map[string]struct{})
		)
		{
			{
				var (
					fmtStr string
					i      int
				)
				for pkg := range pkgs {
					if !strings.HasSuffix(pkg, "_test") {
						fmtStr += fmt.Sprintf("\n%s", pkg)
						pkgName = pkg
						i++
					}
				}
				if i == 0 {
					return nil, PackagingError(fmt.Errorf(""+
						"No Go packages were defined "+
						" in specified directory: %s",
						pkgFullPath,
					))
				} else if i > 1 {
					return nil, PackagingError(fmt.Errorf("" +
						"More than 1 package were defined:" +
						fmtStr +
						"in specified directory: %s" +
						pkgFullPath,
					))
				}
			}

			imps := make(map[string]string)
			for _, file := range pkgs[pkgName].Files {
				for _, imp := range file.Imports {
					s := imp.Path.Value
					s = s[1 : len(s)-1]
					if imp.Name != nil {
						imps[s] = imp.Name.Name
					} else {
						imps[s] = ""
					}
				}
			}

			res.Info = PackageInfo{
				PkgName:     pkgName,
				PkgPath:     pkgPath,
				PkgFullPath: pkgFullPath,
			}
			{
				res, ok := parsedPackages[res.Info]
				if ok {
					return res, nil
				}
			}

			res.Imports = make([]Import, len(imps))
			var i int
			for imp, alias := range imps {
				if _, err := os.Stat(path.Join(GOSRC, imp)); !os.IsNotExist(err) {
					impPath := path.Join(GOSRC, imp)
					pkg, err := parse(impPath, imp, m)
					if err != nil {
						return nil, err
					}
					res.Imports[i] = Import{
						Alias:   alias,
						Package: pkg,
					}
				} else if _, err := os.Stat(path.Join(GOSTD, imp)); !os.IsNotExist(err) {
					stdLibs[imp] = struct{}{}
					impPath := path.Join(GOSTD, imp)
					var name string
					{
						s := strings.Split(imp, "/")
						name = s[len(s)-1]
					}

					res.Imports[i] = Import{
						Alias: alias,
						Package: &Package{
							Info: PackageInfo{
								PkgName:     name,
								PkgPath:     imp,
								PkgFullPath: impPath,
								Std:         true,
							},
						},
					}
				} else {
					return nil, PackagingError(fmt.Errorf(
						"Package weren't found: %s",
						imp,
					))
				}

				i++
			}
		}

		res.Types, res.Methods, err = _parse(
			stdLibs,
			res.Info,
			pkgs[pkgName],
		)
		if err != nil {
			return nil, err
		}

		parsedPackages[res.Info] = &res
		return &res, err
	}

	return parse(pkgPath, "", make(map[string]int))
}

func _parse(stdLibs map[string]struct{}, pkgInfo PackageInfo, pkg *ast.Package) ([]Type, []Method, error) {
	var (
		parseTypeDefinition   func(*ast.TypeSpec) (Type, error)
		parseMethodDefinition func(ast.Node) (Method, error)
		parseTypeDeclaration  func(ast.Node) (interface{}, error)
	)

	matchBasicType := func(t string) (Type, error) {
		res := Type{
			Basic: true,
		}
		switch t {
		case "bool":
			fallthrough
		case "int":
			fallthrough
		case "int64":
			fallthrough
		case "int32":
			fallthrough
		case "uint":
			fallthrough
		case "uint64":
			fallthrough
		case "uint32":
			fallthrough
		case "string":
			fallthrough
		case "rune":
			fallthrough
		case "byte":
			fallthrough
		case "error":
			res.Name = t
		default:
			return res, fmt.Errorf("No std type matches: %s", t)
		}

		return res, nil
	}

	parseFuncSignature := func(node *ast.FuncType, f *Function) error {
		for _, param := range node.Params.List {
			var n string
			if len(param.Names) > 0 {
				n = param.Names[0].Name
			}
			v, err := parseTypeDeclaration(param.Type)
			if err != nil {
				return err
			}
			f.Args = append(f.Args, NameTypePair{
				Name: n,
				Type: v.(Type),
			})
		}

		for _, res := range node.Results.List {
			var n string
			if len(res.Names) > 0 {
				n = res.Names[0].Name
			}
			v, err := parseTypeDeclaration(res.Type)
			if err != nil {
				return err
			}
			f.Return = append(f.Return, NameTypePair{
				Name: n,
				Type: v.(Type),
			})
		}

		return nil
	}

	isIgnorable := func(node ast.Node) bool {
		switch node.(type) {
		case *ast.ImportSpec:
		case *ast.CommentGroup:
		case *ast.Comment:
		default:
			return false
		}
		return true
	}

	parseTypeDefinition = func(d *ast.TypeSpec) (res Type, err error) {
		res.Name = d.Name.Name

		switch n := d.Type.(type) {
		case *ast.InterfaceType:
			var err error
			res.Definition, err = parseTypeDeclaration(n)
			if err != nil {
				return Type{}, err
			}

		case *ast.StructType:
			var err error
			res.Definition, err = parseTypeDeclaration(n)
			if err != nil {
				return Type{}, err
			}
		}

		return res, nil
	}

	parseTypeDeclaration = func(node ast.Node) (res interface{}, err error) {
		switch n := node.(type) {

		case *ast.StructType:
			s := make(Struct)
			for _, field := range n.Fields.List {
				t, err := parseTypeDeclaration(field.Type)
				if err != nil {
					return Type{}, err
				}

				var tag string
				if field.Tag != nil {
					tag = field.Tag.Value
				}

				s[field.Names[0].Name] = TagTypePair{
					Tag:  tag,
					Type: t,
				}
			}
			return s, err

		case *ast.InterfaceType:
			var i Interface
			for _, field := range n.Methods.List {
				var f Function
				f.Name = field.Names[0].Name
				if err := parseFuncSignature(
					field.Type.(*ast.FuncType),
					&f,
				); err != nil {
					return Type{}, err
				}
				i.Methods = append(i.Methods, f)
			}
			return i, nil

		case *ast.Field:
			ident, ok := n.Type.(*ast.Ident)
			if ok {
				t, err := matchBasicType(ident.Name)
				if err == nil {
					return t, nil
				}
			}
			panic("What's next?")
			return nil, nil

		case *ast.Ident:
			t, err := matchBasicType(n.Name)
			if err == nil {
				return t, nil
			}
			panic("What's next?")
			return nil, nil

		case *ast.SelectorExpr:
			return Type{
				From: n.X.(*ast.Ident).Name,
				Name: n.Sel.Name,
			}, nil

		default:
			return nil, fmt.Errorf(
				"No match for type declaration: %s",
				spew.Sdump(n),
			)
		}
	}

	parseMethodDefinition = func(node ast.Node) (res Method, err error) {
		return Method{}, nil
	}
	parseMethodDefinition(nil)

	var (
		types   []Type
		methods []Method
		parse   func(node ast.Node) error
	)

	parse = func(node ast.Node) error {
		fmt.Printf("MOKOKOKO: %s", spew.Sdump(node))
		switch n := node.(type) {
		case *ast.GenDecl:
			for _, spec := range n.Specs {
				if err := parse(spec); err != nil {
					return err
				}
			}
		case *ast.TypeSpec:
			t, err := parseTypeDefinition(n)
			if err != nil {
				return err
			}
			types = append(types, t)
		default:
			if isIgnorable(node) {
				return nil
			}

			return errors.New("No matches")
		}

		return nil
	}

	for _, file := range pkg.Files {

		if file.Name.Name != "user" {
			continue
		}

		spew.Dump(file)

		for _, decl := range file.Decls {
			if err := parse(decl); err != nil {
				return nil, nil, err
			}
		}
	}

	return types, methods, nil

	// parse = func(node ast.Node) (res Type, err error) {
	// 	switch v := node.(type) {
	// 	case *ast.TypeSpec:
	// 		res.Name = v.Name.Name
	// 		res.Definition, err = _parse(v.Type)
	// 		return

	// 	case *ast.ImportSpec:
	// 		return res, pass
	// 	case *ast.CommentGroup:
	// 		return res, pass
	// 	case *ast.Comment:
	// 		return res, pass
	// 	case *ast.GenDecl:
	// 		return res, pass
	// 	default:
	// 		fmt.Println("TROLOLO")
	// 		spew.Dump(node)
	// 		return Type{}, errors.New("No option")
	// 	}

	// 	return Type{}, nil
	// }

	// var res []Type
	// for _, file := range pkg.Files {

	// 	if file.Name.Name != "user" {
	// 		continue
	// 	}

	// 	fmt.Println(file.Name.Name)
	// 	spew.Dump(file)

	// 	for _, decl := range file.Decls {
	// 		switch v := decl.(type) {
	// 		case *ast.GenDecl:
	// 			for _, spec := range v.Specs {
	// 				t, err := parse(spec)
	// 				if err == pass {
	// 					continue
	// 				}
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 				res = append(res, t)
	// 			}
	// 		default:
	// 			return nil, errors.New("Something strange happened")
	// 		}

	// 		// fmt.Println("OLOLO", len(file.Decls))
	// 		t, err := parse(decl)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		res = append(res, t)
	// 	}
	// }
}
