package dependency

import "errors"

type Tree map[string]Tree

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

func (g Graph) SetDependency(src, dst string) {
	deps := g[src]
	if deps == nil {
		deps = make(map[string]bool)
		g[src] = deps
	}
	deps[dst] = true
}

func (g Graph) FindLoops() [][2]string {
	res := make([][2]string, 0)
	g = g.Copy()
	g.sort()

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
		g[k] = d
	}
	return res
}

// WARNING: destructive
func (g Graph) sort() []string {
	var startNodes, res []string

	for n := range g {
		if !g.isDependency(n) {
			startNodes = append(startNodes, n)
		}
	}

	for len(startNodes) > 0 {
		n := startNodes[0]
		startNodes = startNodes[1:]
		res = append(res, n)

		for d, ok := range g[n] {
			if ok {
				g[n][d] = false
				if !g.isDependency(d) {
					startNodes = append(startNodes, n)
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
