package register

import (
	"go/token"
)

type (
	Error interface {
		Pos() Pos
		error
	}

	Pos struct {
		From token.Position
		To   token.Position
	}
)

var _ Error = ErrBadSyntax{}

type ErrBadSyntax struct {
	from token.Position
	to   token.Position
}

func (e ErrBadSyntax) Error() string {
	return "bad syntax in type declaration"
}

func (e ErrBadSyntax) Pos() Pos {
	return Pos{e.from, e.to}
}
