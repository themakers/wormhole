package dependencies

import (
	"fmt"
	"strings"
)

type Tree map[string]Tree

func (t Tree) String() string {
	return strings.Join(stringify(t, 0), "\n")
}

func stringify(t Tree, i int) []string {
	var res []string

	for pkg, deps := range t {
		res = append(res, fmt.Sprintf(
			"%s %s",
			strings.Repeat("-- ", i),
			pkg,
		))

		if deps != nil {
			res = append(res, stringify(deps, i+1)...)
		}
	}

	return res
}
