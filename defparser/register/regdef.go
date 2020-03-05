package register

import (
	"fmt"
	"sort"
	"unicode"

	"github.com/themakers/wormhole/defparser/types"
)

func (r *Register) regDef(
	name string,
	decl types.Type,
	pkg *types.Package,
) error {
	if r.UsedNames[name] {
		return fmt.Errorf("name was used: %s", name)
	}

	d := &types.Definition{
		Name:        name,
		Declaration: decl,
		Package:     pkg,
		MethodsMap:  map[string]*types.Method{},
		Exported:    isExported(name),
		Std:         pkg.Info.Std,
	}

	pkg.Definitions = append(pkg.Definitions, d)
	sort.Slice(pkg.Definitions, func(i, j int) bool {
		return pkg.Definitions[i].Name < pkg.Definitions[j].Name
	})
	pkg.DefinitionsMap[name] = d
	r.UsedNames[name] = true

	return nil
}

func isExported(s string) bool {
	return unicode.IsLetter(rune(s[0])) &&
		unicode.IsUpper(rune(s[0]))
}
