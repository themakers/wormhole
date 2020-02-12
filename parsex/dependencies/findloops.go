package dependencies

type vertex struct {
	self    string
	onStack bool
	lowLink int
	index   int
	visited bool
	desc    []string
}

// Graph.FindLoops() uses Tarjan's algorithm
func (g Graph) FindLoops() [][]string {
	var (
		res   [][]string
		stack []*vertex
		index int
	)

	vertexes := make(map[string]*vertex, len(g))
	for v := range g {
		deps := g[v]
		desc := make([]string, len(deps))
		{
			var i int
			for dep := range deps {
				desc[i] = dep
				i++
			}
		}

		vertexes[v] = &vertex{
			self: v,
			desc: desc,
		}
	}

	var do func(*vertex)
	do = func(v *vertex) {
		{
			v.index = index
			v.lowLink = index
			v.visited = true
			stack = append(stack, v)
			v.onStack = true
			index++
		}

		for _, desc := range v.desc {
			w := vertexes[desc]
			if !w.visited {
				do(w)
				if v.lowLink > w.lowLink {
					v.lowLink = w.lowLink
				}
			} else if w.onStack {
				if v.lowLink > w.index {
					v.lowLink = w.index
				}
			}
		}

		if v.lowLink == v.index {
			var cycle []string

			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				w.onStack = false
				cycle = append(cycle, w.self)

				if v.self == w.self {
					break
				}
			}

			if len(cycle) > 1 {
				res = append(res, cycle)
			}
		}
	}

	for _, v := range vertexes {
		if !v.visited {
			do(v)
		}
	}

	return res
}
