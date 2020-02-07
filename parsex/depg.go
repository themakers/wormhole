package parsex

// type DepTree map[string]DepTree

// type DepGraph map[string]map[string]bool

// func NewDepGraph() DepGraph {
// 	return make(DepGraph)
// }

// ///////

// type DependencyGraph struct {
// 	g map[string]map[string]bool
// }

// func NewDependencyGraph() *DependencyGraph {
// 	return &DependencyGraph{
// 		g: make(map[string]map[string]bool),
// 	}
// }

// func (dg *DependencyGraph) SetDependency(src, dst string) {
// 	deps := dg.g[src]
// 	if deps == nil {
// 		deps = make(map[string]bool)
// 		dg.g[src] = deps
// 	}
// 	deps[dst] = true
// }

// func (dg *DependencyGraph) FindLoops() [][2]string {
// 	var (
// 		startNodes []string
// 		res        = make([][2]string, 0)
// 		graph      = dg.Copy()
// 	)

// 	for n := range graph.g {
// 		if !graph.isDependency(n) {
// 			startNodes = append(startNodes, n)
// 		}
// 	}

// 	for len(startNodes) > 0 {
// 		n := startNodes[0]
// 		startNodes = startNodes[1:]

// 		for d, ok := range graph.g[n] {
// 			if ok {
// 				graph.g[n][d] = false
// 				if !graph.isDependency(d) {
// 					startNodes = append(startNodes, n)
// 				}
// 			}
// 		}
// 	}

// 	for src, deps := range graph.g {
// 		for dst, ok := range deps {
// 			if ok {
// 				res = append(res, [2]string{src, dst})
// 			}
// 		}
// 	}

// 	return res
// }

// func (dg *DependencyGraph) Copy() *DependencyGraph {
// 	res := &DependencyGraph{
// 		g: make(map[string]map[string]bool),
// 	}

// 	for k, v := range dg.g {
// 		d := make(map[string]bool)
// 		for k, v := range v {
// 			d[k] = v
// 		}
// 		dg.g[k] = d
// 	}

// 	return res
// }

// func (dg *DependencyGraph) isDependency(n string) bool {
// 	for _, deps := range dg.g {
// 		if deps[n] {
// 			return true
// 		}
// 	}
// 	return false
// }
