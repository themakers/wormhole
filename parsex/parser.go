package parsex

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/themakers/wormhole/parsex/astwalker"
	"github.com/themakers/wormhole/parsex/dependency"
)

// TODO Implement post loading of unresolved types. i.e. when dir name differ from pkg name, or pkg imported as .

var (
	GOROOT = filepath.Clean(os.Getenv("GOROOT"))
	GOPATH = filepath.Clean(os.Getenv("GOPATH"))
	GOSRC  = filepath.Join(GOPATH, "src")
	GOSTD  = filepath.Join(GOROOT, "src")
)

func PWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func Parse(pkgPath string) (*Parsed, error) {
	p, err := NewParserX(pkgPath)
	if err != nil {
		return nil, err
	}

	if err := p.buildDepGraph(pkgPath); err != nil {
		panic(err)
		return nil, err
	}

	loops := p.depGraph.FindLoops()
	fmt.Println(len(loops))
	fmt.Println(loops)
	fmt.Println("\n\n\n\n\n\n\ndep graph was built")

	// {
	// 	g := p.depGraph.Copy()
	// 	spew.Dump(map[string]map[string]bool(g))
	// 	res := g.Sort()
	// 	spew.Dump(res)
	// 	spew.Dump(map[string]map[string]bool(g))
	// }
	// fmt.Println(p.depGraph.TreeView())

	panic("FINISH")
	return nil, nil
}

// func Parse(files ...string) (*Parsed, error) {
// 	t := &Parsed{
// 		InterfacesMap: map[string]*Interface{},
// 	}

// 	fset := token.NewFileSet()

// 	for _, file := range files {
// 		file = filepath.Clean(file)
// 		if !filepath.IsAbs(file) {
// 			return nil, ErrNotAbsoluteFilePath
// 		}

// 		log.Println("parsing", file)

// 		file, err := parser.ParseFile(fset, file, nil, 0)
// 		if err != nil {
// 			return nil, err
// 		}

// 		for _, importSpec := range file.Imports {
// 			path := importSpec.Path.Value
// 			log.Printf("IMPORT %s", path)

// 		}

// 		astwalker.WalkAST(file, nil, func(node *astwalker.Node) astwalker.VisitorFunc {
// 			shift := strings.Repeat("--*", node.Depth())
// 			log.Printf("%s #%d :: @%d :: %s\n", shift, node.Node().Pos(), node.Depth(), node.String())

// 			return nil
// 		})
// 	}

// 	panic("FINISH")

// 	return nil, nil

// 	for _, file := range files {
// 		file := filepath.Clean(file)

// 		if !filepath.IsAbs(file) {
// 			panic("Only absolute paths are allowed. " + file)
// 		}

// 		log.Println("parsing", file)

// 		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments) // |parser.Trace
// 		if err != nil {
// 			return nil, err
// 		}

// 		//fileStr := (func() string {
// 		//	data, err := ioutil.ReadFile(file)
// 		//	if err != nil {
// 		//		panic(err)
// 		//	}
// 		//	return string(data)
// 		//})()

// 		//pkgPath := strings.TrimPrefix(strings.TrimSuffix(strings.TrimSuffix(file, filepath.Base(file)), "/"), GOSRC+"/")
// 		//pkgName := f.Message.Message

// 		fmt.Println(strings.Repeat("•", 64))
// 		astwalker.WalkAST(f, nil, testWalker()).Root().Children()[0].Children()
// 		fmt.Println(strings.Repeat("•", 64))
// 		astwalker.WalkAST(f, nil, InterfaceMethodWalker(func(pkg, ifcName, method string, arg, res []Param) {
// 			t.Pkg = pkg

// 			ifc := t.InterfacesMap[ifcName]
// 			if ifc == nil {
// 				ifc = &Interface{
// 					Name:       ifcName,
// 					MethodsMap: map[string]*Method{},
// 				}
// 				t.Interfaces = append(t.Interfaces, ifc)
// 				t.InterfacesMap[ifcName] = ifc
// 			}

// 			meth := &Method{
// 				Interface: ifcName,
// 				Name:      method,
// 				Args:      arg,
// 				Rets:      res,
// 			}

// 			ifc.Methods = append(ifc.Methods, meth)
// 			ifc.MethodsMap[meth.Name] = meth
// 		}))
// 	}

// 	return t, nil
// }

type parserx struct {
	pkgPath string
	// fset     *token.FileSet
	depGraph dependency.Graph
}

// func (p *parser) getDepGraph(pkgPath string) (dependency.Graph, error) {
// 	return nil, nil
// }

func NewParserX(pkgPath string) (*parserx, error) {
	pkgPath = filepath.Clean(pkgPath)
	if !filepath.IsAbs(pkgPath) {
		return nil, fmt.Errorf(
			"Provided file path isn't absolute: %s",
			pkgPath,
		)
	}

	return &parserx{
		pkgPath: pkgPath,
		// fset:     token.NewFileSet(),
		depGraph: dependency.NewGraph(),
	}, nil
}

func (p *parserx) buildDepGraph(pkgPath string) error {
	fmt.Println("NEW ITERATION", pkgPath)
	defer fmt.Println("END ITERATION")

	if !p.depGraph.AddNode(pkgPath) {
		fmt.Println("WAS PARSED")
		return nil
	}

	pkgs, err := parser.ParseDir(
		token.NewFileSet(),
		pkgPath,
		nil,
		parser.ImportsOnly,
	)
	fmt.Println("PARSED FILES: ", pkgs)
	if err != nil {
		return err
	}

	var pkgName string
	{
		var (
			fmtStr string
			i      int
		)
		for pkg := range pkgs {
			if !strings.HasSuffix(pkg, "_test") {
				fmtStr += fmt.Sprintf(" %s", pkg)
				pkgName = pkg
				i++
			}
		}
		if i == 0 {
			fmt.Println("ERR 1")
			return fmt.Errorf(
				"No Go packages were defined in specified directory",
			)
		} else if i > 1 {
			fmt.Println("ERR 2")
			return fmt.Errorf("" +
				"More than 1 package were defined in specified directory:" +
				fmtStr,
			)
		}
	}

	imps := make(map[string]struct{})
	for id, file := range pkgs[pkgName].Files {
		fmt.Println("FILE: ", id, " ", len(file.Imports))
		for _, imp := range file.Imports {
			s := imp.Path.Value
			imps[s[1:len(s)-1]] = struct{}{}
		}
	}

	for imp := range imps {
		fmt.Printf("AAA %s: %s\n", pkgName, imp)
		var impPath string
		if _, err := os.Stat(path.Join(GOSRC, imp)); !os.IsNotExist(err) {
			impPath = path.Join(GOSRC, imp)
			if err := p.buildDepGraph(impPath); err != nil {
				return err
			}
		} else if _, err := os.Stat(path.Join(GOSTD, imp)); !os.IsNotExist(err) {
			impPath = path.Join(GOSTD, imp)
		} else {
			return fmt.Errorf("Package weren't found: %s", imp)
		}

		p.depGraph.SetDependency(pkgPath, impPath)
		if loops := p.depGraph.FindLoops(); len(loops) > 0 {
			return fmt.Errorf("Found loops: %v", loops)
		}
	}

	return nil
}

func (p *parserx) parse(pkgPath string) (*ast.File, error) {
	// parser.ParseDir(p.fset, pkgPath, nil, parser.ParseComments)
	return nil, nil
}

func typeName(e ast.Expr) string {
	if sel, ok := e.(*ast.SelectorExpr); ok {
		tx := fmt.Sprint(sel.X)
		tsel := fmt.Sprint(sel.Sel)
		if tx != "" {
			return fmt.Sprintf("%s.%s", tx, tsel)
		} else {
			return tsel
		}
	} else if idn, ok := e.(*ast.Ident); ok {
		return idn.Name
	} else if sexp, ok := e.(*ast.StarExpr); ok {
		return fmt.Sprintf("*%s", typeName(sexp.X))
	} else {
		panic(fmt.Sprintf("unknown element: %#v", e))
	}
}

func InterfaceMethodWalker(h func(pkg, ifc, method string, args, ress []Param)) astwalker.VisitorFunc {
	pkg := ""
	return func(node *astwalker.Node) astwalker.VisitorFunc {

		//> Should be interface method
		if node.Top(0).IsFuncType() &&
			len(node.Top(1).Children()) == 2 && node.Top(1).Children()[0].IsIdent() &&
			node.Top(3).IsInterfaceType() &&
			node.Top(4).IsTypeSpec() &&
			node.Top(5).IsGenDecl() &&
			len(node.Top(4).Children()) == 2 && node.Top(4).Children()[0].IsIdent() {

			ifcName := node.Top(4).Children()[0].Ident().Name
			methName := node.Top(1).Children()[0].Ident().Name

			ft := node.FuncType()

			var i = 0
			getParams := func(fields []*ast.Field) (params []Param) {
				for _, p := range fields {
					var par Param
					if len(p.Names) > 0 {
						for _, n := range p.Names {
							i++
							if n.Name == "" {
								n.Name = fmt.Sprintf("__%d", i)
							}
							par.Names = append(par.Names, n.Name)
						}
					} else {
						i++
						par.Names = append(par.Names, fmt.Sprintf("__%d", i))
					}
					par.Type = typeName(p.Type)
					params = append(params, par)
				}

				return params
			}

			var args []Param

			var ress []Param

			if len(ft.Params.List) > 1 {
				args = getParams(ft.Params.List[1:])
			}

			if len(ft.Results.List) > 0 {
				ress = getParams(ft.Results.List)
			}

			h(pkg, ifcName, methName, args, ress)
		} else if node.Top(0).IsIdent() &&
			node.Top(1).IsFile() {
			pkg = node.Top(0).Ident().Name
		}
		return nil
	}
}
