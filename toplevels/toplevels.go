package toplevels

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TODO Implement post loading of unresolved types. i.e. when dir name differ from pkg name, or pkg imported as .

var (
	GOROOT = filepath.Clean(os.Getenv("GOROOT"))
	GOPATH = filepath.Clean(os.Getenv("GOPATH"))
	GOSRC  = filepath.Join(GOPATH, "src")
)

func PWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

type Type struct {
	Prefix  string
	Package string
	Name    string

	Path string
}

func (t Type) ID() string {
	return fmt.Sprintf("%s#%s", t.Path, t.Name)
}

func (t Type) String() string {
	if t.Package == "" {
		return fmt.Sprint(t.Prefix, t.Name)
	} else {
		return fmt.Sprint(t.Prefix, t.Package, ".", t.Name)
	}
}

type Interface struct {
	Type      Type
	Name      string
	Doc       []string
	Functions map[string]*Function
}

type Function struct {
	Doc  []string
	Name string
	Args []*FunctionParam
	Rets []*FunctionParam
}

type FunctionParam struct {
	Name string
	Type string
}

type Parser struct {
	Toplevels map[string]interface{}

	Interfaces map[string]*Interface
}

func NewParser() *Parser {
	p := &Parser{
		Toplevels:  make(map[string]interface{}),
		Interfaces: make(map[string]*Interface),
	}
	return p
}

func (p *Parser) Parse(files ...string) {
	fset := token.NewFileSet()

	for _, file := range files {
		file := filepath.Clean(file)

		if !filepath.IsAbs(file) {
			panic("Only absolute paths are allowed. " + file)
		}

		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments) // |parser.Trace
		if err != nil {
			panic(err)
			// return
		}

		fileStr := (func() string {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				panic(err)
			}
			return string(data)
		})()

		pkgPath := strings.TrimPrefix(strings.TrimSuffix(strings.TrimSuffix(file, filepath.Base(file)), "/"), GOSRC+"/")
		pkgName := f.Name.Name

		for name, obj := range f.Scope.Objects {
			switch obj.Kind {
			case ast.Typ:
				log.Println(name, obj)
				typeSpec := obj.Decl.(*ast.TypeSpec)

				if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {

					_interface := &Interface{
						Type: Type{
							Name:    name,
							Path:    pkgPath,
							Package: pkgName,
						},
						Name:      name,
						Functions: make(map[string]*Function),
					}

					for _, field := range interfaceType.Methods.List {

						_function := &Function{}

						for _, name := range field.Names {
							_function.Name = name.Name
						}

						funcType := field.Type.(*ast.FuncType)

						parseParams := func(params *ast.FieldList) (out []*FunctionParam) {
							if params != nil && len(params.List) > 0 {
								for _, param := range params.List {
									_arg := &FunctionParam{}

									// var parseType func(ast.Expr)
									// parseType = func(param ast.Expr) {
									// 	switch typ := param.(type) {
									// 	case *ast.Ident:
									// 		_arg.Type.Name = typ.Name
									// 	case *ast.SelectorExpr:
									// 		_arg.Type.Name = typ.Sel.Name
									// 		_arg.Type.Package = typ.X.(*ast.Ident).Name
									// 	case *ast.ArrayType:
									// 		_arg.Type.Prefix += "[]"
									// 		parseType(typ.Elt)
									// 	case *ast.StarExpr:
									// 		_arg.Type.Prefix += "*"
									// 		parseType(typ.X)
									// 	default:
									// 		typ.Pos
									// 		// panic(fmt.Sprintf("Unknown type: %#v", typ))
									// 	}
									// }

									// parseType(param.Type)

									_arg.Type = fileStr[param.Pos() : param.Pos()+param.End()]

									if len(param.Names) > 0 {
										for _, name := range param.Names {
											arg := *_arg
											arg.Name = name.Name
											out = append(out, &arg)
										}
									} else {
										arg := *_arg
										out = append(out, &arg)
									}
								}
							}
							return
						}

						_function.Args = parseParams(funcType.Params)
						_function.Rets = parseParams(funcType.Results)

						_interface.Functions[_function.Name] = _function
					}

					p.Interfaces[_interface.Type.ID()] = _interface
					p.Toplevels[_interface.Type.ID()] = _interface
				}
			}
		}
	}
}
