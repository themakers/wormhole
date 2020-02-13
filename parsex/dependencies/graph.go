package dependencies

import (
	"errors"
	"fmt"
)

// Graph nodes shouldn't be empty strings
type Graph map[Vertex]map[Vertex]bool

type Vertex struct {
	PkgName  string
	Path     string
	FullPath string
	Alias    string
}

func NewGraph() Graph {
	return make(Graph)
}

func (g Graph) AddNode(n Vertex) bool {
	if g[n] == nil {
		g[n] = make(map[Vertex]bool)
		return true
	}
	return false
}

func (g Graph) SetDependency(src, dst Vertex) {
	{
		var empt Vertex
		if dst == empt || src == empt {
			panic(errors.New("Graph nodes shouldn't be empty strings"))
		}
	}

	deps := g[src]
	if deps == nil {
		deps = make(map[Vertex]bool)
		g[src] = deps
	}
	if g[dst] == nil {
		g[dst] = make(map[Vertex]bool)
	}
	deps[dst] = true
}

func (g Graph) Copy() Graph {
	res := make(map[Vertex]map[Vertex]bool)
	for k, v := range g {
		d := make(map[Vertex]bool)
		for k, v := range v {
			d[k] = v
		}
		res[k] = d
	}
	return res
}

// WARNING: destructive
func (g Graph) Sort() []Vertex {
	var startNodes, res []Vertex

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

func (g Graph) isDependency(n Vertex) bool {
	for _, deps := range g {
		if deps[n] {
			return true
		}
	}
	return false
}
