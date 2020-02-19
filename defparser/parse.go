package defparser

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/themakers/wormhole/defparser/types"
)

func Parse(pkgPath string) (*Result, error) {
	var (
		do    func(pkgFullPath, pkgPath string, prev map[string]int) (*types.Package, error)
		index int
		tc    = newTypeChecker()
	)

	do = func(pkgFullPath, pkgPath string, prev map[string]int) (*types.Package, error) {
		var (
			pkgTC *typeChecker
			pkg   *types.Package
		)

		pkgFullPath = filepath.Clean(pkgFullPath)
		if !filepath.IsAbs(pkgFullPath) {
			return nil, ErrNotAbsoluteFilePath
		}

		pkgIndx, ok := prev[pkgFullPath]
		if ok {
			return nil, Loop(pkgFullPath)
		}

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
			func(info os.FileInfo) bool {
				if strings.HasSuffix(info.Name(), "_test.go") {
					fmt.Println(info.Name())
					return false
				}

				if strings.HasSuffix(info.Name(), ".gen.go") {
					fmt.Println(info.Name())
					return false
				}

				return true
			},
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
					fmtStr += fmt.Sprintf("\n%s", pkg)
					pkgName = pkg
					i++
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

			info := types.PackageInfo{
				PkgName:     pkgName,
				PkgPath:     pkgPath,
				PkgFullPath: pkgFullPath,
			}

			if pkg, ok = tc.global.pkgs[info]; ok {
				return pkg, nil
			}

			imports := make([]types.Import, len(imps))
			var i int
			for imp, alias := range imps {
				if _, err := os.Stat(path.Join(GOSRC, imp)); !os.IsNotExist(err) {
					impPath := path.Join(GOSRC, imp)
					pkg, err := do(impPath, imp, m)
					if err != nil {
						return nil, err
					}
					imports[i] = types.Import{
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

					imports[i] = types.Import{
						Alias: alias,
						Package: &types.Package{
							Info: types.PackageInfo{
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

			pkgTC = tc.newPackage(info, imports)
		}

		err = aggregateDefinitions(
			pkgTC,
			pkgs[pkgName],
		)
		return pkgTC.pkg, err
	}

	if _, err := do(pkgPath, "", make(map[string]int)); err != nil {
		return nil, err
	}

	return tc.getResult(), nil
}
