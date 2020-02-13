package dependencies

import (
	"errors"
	"fmt"
	"strings"
)

type Tree map[Vertex]Tree

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

func (g Graph) TreeView() (Tree, error) {
	if loops := g.FindLoops(); len(loops) != 0 {
		return nil, errors.New(
			"Failed to get tree view of the graph. Graph contains loops.",
		)
	}

	res := make(Tree)
	for n := range g {
		if !g.isDependency(n) {
			res[n] = g.treeView(n)
		}
	}

	return res, nil
}

func (g Graph) treeView(n Vertex) Tree {
	var res Tree

	for dep, ok := range g[n] {
		if ok {
			if res == nil {
				res = make(Tree)
			}
			res[dep] = g.treeView(dep)
		}
	}

	return res
}
