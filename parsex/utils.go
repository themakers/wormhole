package parsex

import (
	"fmt"
	"github.com/themakers/wormhole/parsex/astwalker"
	"go/ast"
	"strings"
)

func testWalker() astwalker.VisitorFunc {
	return func(node *astwalker.Node) astwalker.VisitorFunc {
		fmt.Println(
			fmt.Sprintf("%5d", node.Depth()),
			strings.Repeat("•", node.Depth()*2),
			node.String(),
		)
		return nil
	}
}

func testWalk(f ast.Node) {
	ast.Walk(visitor(0), f)
}

type visitor int

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	idn := ""
	if n, ok := n.(*ast.Ident); ok {
		idn = n.Name
	}

	if idn != "" {
		idn = " - " + idn
	}

	fmt.Printf("%5d %s %T%s\n", v, strings.Repeat("•", int(v)*2), n, idn)
	return v + 1
}
