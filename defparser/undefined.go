package defparser

import (
	"fmt"

	"github.com/themakers/wormhole/defparser/types"
)

type undefined struct {
	parents []interface{}
	*types.Definition
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
		u.Name,
	)
}

func (u *undefined) define(d *types.Definition) {
	for _, parent := range u.parents {
		switch p := parent.(type) {
		case *types.StructField:
			p.Type = d

		case *types.Chan:
			p.Type = d

		case *types.Array:
			p.Type = d

		case *types.Slice:
			p.Type = d

		case *types.Definition:
			p.Declaration = d

		case *types.Pointer:
			p.Type = d

		case *types.Map:
			if p.Key == u {
				p.Key = d
			} else if p.Value == u {
				p.Value = d
			} else {
				panic("WTF?")
			}

		case *types.NameTypePair:
			p.Type = d

		default:
			panic("Unknown parent")
		}
	}
}
