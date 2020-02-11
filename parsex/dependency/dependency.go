package dependency

import (
	"errors"
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

type Graph map[string]map[string]bool

func NewGraph() Graph {
	return make(Graph)
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

func (g Graph) treeView(n string) Tree {
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

func (g Graph) AddNode(n string) bool {
	if g[n] == nil {
		g[n] = make(map[string]bool)
		return true
	}
	return false
}

func (g Graph) SetDependency(src, dst string) {
	deps := g[src]
	if deps == nil {
		deps = make(map[string]bool)
		g[src] = deps
	}
	deps[dst] = true
}

func (g Graph) FindLoops() Loops {
	res := make([][2]string, 0)
	g = g.Copy()
	g.Sort()

	for src, deps := range g {
		for dst, ok := range deps {
			if ok {
				res = append(res, [2]string{src, dst})
			}
		}
	}

	return res
}

func (g Graph) Copy() Graph {
	res := make(map[string]map[string]bool)
	for k, v := range g {
		d := make(map[string]bool)
		for k, v := range v {
			d[k] = v
		}
		res[k] = d
	}
	return res
}

// WARNING: destructive
func (g Graph) Sort() []string {
	var startNodes, res []string

	for n := range g {
		if !g.isDependency(n) {
			startNodes = append(startNodes, n)
		}
	}

	fmt.Println("START_NODES")
	fmt.Println(startNodes)

	for len(startNodes) > 0 {
		n := startNodes[0]
		startNodes = startNodes[1:]
		res = append(res, n)

		for d, ok := range g[n] {
			if ok {
				g[n][d] = false
				if !g.isDependency(d) {
					startNodes = append(startNodes, d)
				}
			}
		}
	}

	return res
}

func (g Graph) isDependency(n string) bool {
	for _, deps := range g {
		if deps[n] {
			return true
		}
	}
	return false
}

type Loops [][2]string

func (l Loops) String() string {
	res := make([]string, len(l))
	for i, loop := range l {
		res[i] = fmt.Sprintf("%s <-> %s", loop[0], loop[1])
	}
	return strings.Join(res, "\n")
}
