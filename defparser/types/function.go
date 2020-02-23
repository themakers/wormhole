package types

import (
	"fmt"
	"strings"
)

var _ Type = &Function{}

type (
	Function struct {
		Args    []NameTypePair
		Results []NameTypePair
	}

	NameTypePair struct {
		Name string
		Type Type
	}
)

func (f *Function) Hash() string {
	return f.hash(map[*Definition]bool{})
}

func (f *Function) hash(prev map[*Definition]bool) string {
	s := sum("FUNC")
	for _, arg := range f.Args {
		s += arg.Type.hash(prev)
	}
	for _, result := range f.Results {
		s += result.Type.hash(prev)
	}
	return sum(s)
}

const funcTmpl = "func(%s)(%s)"

func (f *Function) String() string {
	args := make([]string, len(f.Args))
	results := make([]string, len(f.Results))

	for i, arg := range f.Args {
		args[i] = arg.Type.String()
	}

	for i, result := range f.Results {
		results[i] = result.Type.String()
	}

	return fmt.Sprintf(
		funcTmpl,
		strings.Join(args, ","),
		strings.Join(results, ","),
	)
}
