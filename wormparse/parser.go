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

		var pkgName string
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

		res.Types, err = ParseTypes(res.Info, pkgs[pkgName])
		if err != nil {
			return nil, err
		}

		parsedPackages[res.Info] = &res
		return &res, err
	}

	return parse(pkgPath, "", make(map[string]int))
}

func ParseTypes(pkgInfo PackageInfo, pkg *ast.Package) ([]Type, error) {
	var (
		parse  func(ast.Node) (Type, error)
		_parse func(ast.Node) (interface{}, error)
	)
	parse = func(node ast.Node) (res Type, err error) {
		switch v := node.(type) {
		case *ast.TypeSpec:
			res.Name = v.Name.Name
			res.Definition, err = _parse(v.Type)
			if err != nil {
				return
			}
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				return parse(spec)
			}
		default:
			fmt.Println("TROLOLO")
			spew.Dump(node)
			return Type{}, errors.New("No option")
		}

		return Type{}, nil
	}

	parseFuncSignature := func(node *ast.FuncType, f *Function) error {
		for _, param := range node.Params.List {
			var n string
			if len(param.Names) > 0 {
				n = param.Names[0].Name
			}
			v, err := _parse(param.Type)
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
			v, err := _parse(res.Type)
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

	parseBasicTypes := func(node *ast.Ident) (Type, error) {
		res := Type{
			Std: true,
		}
		switch v := node.Name; v {
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
			res.Name = v
		default:
			return res, fmt.Errorf("No std type matches: %s", spew.Sdump(node))
		}

		return res, nil
		// switch node.Obj.Type {

		// } {
		// }

	}

	_parse = func(node ast.Node) (interface{}, error) {
		switch v := node.(type) {
		case *ast.InterfaceType:
			var res Interface
			for _, field := range v.Methods.List {
				var f Function
				f.Name = field.Names[0].Name
				parseFuncSignature(field.Type.(*ast.FuncType), &f)
				res.Methods = append(res.Methods, f)
			}
			return res, nil
		default:
			ident, ok := v.(*ast.Ident)
			if ok {
				return parseBasicTypes(ident)
			}
			return Type{}, errors.New("No option")
		}
	}

	var res []Type
	for _, file := range pkg.Files {

		if file.Name.Name != "user" {
			continue
		}

		fmt.Println(file.Name.Name)
		spew.Dump(file)

		for _, decl := range file.Decls {
			fmt.Println("OLOLO")
			t, err := parse(decl)
			if err != nil {
				return nil, err
			}
			res = append(res, t)
		}
	}

	return res, nil
}
