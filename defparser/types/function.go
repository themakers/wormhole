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
	return string(
		hash.Sum([]byte(f.String())),
	)
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
		strings.Join(args, ","),
		strings.Join(results, ","),
	)
}
