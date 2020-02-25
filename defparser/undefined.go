package defparser

import (
	"fmt"

	"github.com/themakers/wormhole/defparser/types"
)

type undefined struct {
	parents map[string]types.Type
	name    string
	isPtr   bool
}

func (u *undefined) Hash() string {
	return u.hash(nil)
}

func (u *undefined) hash(_ map[*types.Definition]bool) string {
	return u.String()
}

func (u *undefined) String() string {
	return fmt.Sprintf(
		"<undefined>%s",
		u.name,
	)
}

func (u *undefined) define() error {
	def, ok := u.pkg.DefinitionsMap[u.name]
	if !ok {
		return fmt.Errorf(
			"\"%s\" is undefined",
			u.name,
		)
	}

	for _, parent := range u.parents {
		switch p := parent.(type) {
		case *types.Definition:
			p.Declaration = def

		case *types.Function:
			for i, arg := range p.Args {
				if arg.Type == u {
					p.Args[i].Type = def
				}
			}
			for i, result := range p.Results {
				if result.Type == u {
					p.Results[i].Type = def
				}
			}

		case *types.Method:
			p.Receiver = def

		case *types.Struct:
			field, ok := p.FieldsMap[u.name]
			if !ok {
				panic(fmt.Errorf(
					"\"%s\"undefined and \"%s\" fields are desynchronized",
					u.name,
					p,
				))
			}

			// var t types.Type
			// switch d := def.Declaration.(type) {
			// case *types.Struct:

			// case *types.Interface:

			// }

			if u.isPtr {
				field.Type = &types.Pointer{
					Type: def,
				}
			} else {
				field.Type = def
			}
			p.FieldsMap[u.name] = field
		}
	}

	return nil
}
