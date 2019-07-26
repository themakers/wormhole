package astwalker

import (
	"fmt"
	"go/ast"
	"reflect"
)

////////////////////////////////////////////////////////////////
//// ast.Visitor wrapper
////////////////////////////////////////////////////////////////

var _ ast.Visitor = StdVisitorFunc(nil)

type StdVisitorFunc func(node ast.Node) ast.Visitor

func (vfn StdVisitorFunc) Visit(node ast.Node) ast.Visitor {
	if vfn == nil {
		return vfn
	}
	return vfn(node)
}

////////////////////////////////////////////////////////////////
//// VisitorFunc
////////////////////////////////////////////////////////////////

type VisitorFunc func(node *Node) VisitorFunc

////////////////////////////////////////////////////////////////
//// WalkAST
////////////////////////////////////////////////////////////////

func walker(node *Node, visitor VisitorFunc) ast.Visitor {
	return StdVisitorFunc(func(an ast.Node) ast.Visitor {
		if an == nil {
			return nil
		}

		node := node
		if node == nil {
			node = NewRoot(an)
		} else {
			node = node.push(an)
		}

		if visitor != nil {
			if vfn := visitor(node); vfn != nil {
				return walker(node, vfn)
			} else {
				return walker(node, visitor)
			}
		} else {
			return nil
		}
	})
}

func WalkAST(an ast.Node, root *Node, visitor VisitorFunc) *Node {
	if root == nil {
		root = NewRoot(nil)
	}
	ast.Walk(walker(root, visitor), an)
	return root
}

////////////////////////////////////////////////////////////////
//// Node
////////////////////////////////////////////////////////////////

type Node struct {
	root     *Node
	parent   *Node
	children []*Node

	depth int

	an ast.Node
}

func NewRoot(an ast.Node) *Node {
	nc := &Node{
		parent:   nil,
		children: []*Node{},
		depth:    0,
		an:       an,
	}
	nc.root = nc
	return nc
}

func (n *Node) Ok() bool { return n.an != nil }

func (n *Node) Node() ast.Node { return n.an }

func (n *Node) Root() *Node { return n.root }

func (n *Node) Parent() *Node { return n.parent }

func (n *Node) Depth() int { return n.depth }

func (n *Node) Children() []*Node { return n.children }

func (n *Node) String() string {
	text := ""

	switch {
	case n.IsGenDecl():
		text = " "
	case n.IsIdent():
		text = n.Ident().Name
	case n.IsFuncType():
	case n.IsField():
	case n.IsFieldList():
	case n.IsInterfaceType():
	}

	if text != "" {
		text = " - " + text
	}

	return fmt.Sprint(reflect.TypeOf(n.an), text)
}

func (node *Node) Top(n int) *Node {
	for i := 0; i < n && node != nil; i++ {
		node = node.Parent()
	}
	return node
}

func (n *Node) push(an ast.Node) *Node {
	cn := &Node{
		root:     n.root,
		parent:   n,
		children: []*Node{},
		depth:    n.depth + 1,
		an:       an,
	}

	n.children = append(n.children, cn)

	return cn
}

////////////////////////////////////////////////////////////////
func (n Node) GenDecl() *ast.GenDecl {
	if n, ok := n.an.(*ast.GenDecl); ok {
		return n
	} else {
		return nil
	}
}

func (n Node) IsGenDecl() bool {
	return n.GenDecl() != nil
}

////////////////////////////////////////////////////////////////
func (n Node) Ident() *ast.Ident {
	if n, ok := n.an.(*ast.Ident); ok {
		return n
	} else {
		return nil
	}
}

func (n Node) IsIdent() bool {
	return n.Ident() != nil
}

////////////////////////////////////////////////////////////////
func (n Node) InterfaceType() *ast.InterfaceType {
	if n, ok := n.an.(*ast.InterfaceType); ok {
		return n
	} else {
		return nil
	}
}

func (n Node) IsInterfaceType() bool {
	return n.InterfaceType() != nil
}

////////////////////////////////////////////////////////////////
func (n Node) FuncType() *ast.FuncType {
	if n, ok := n.an.(*ast.FuncType); ok {
		return n
	} else {
		return nil
	}
}

func (n Node) IsFuncType() bool {
	return n.FuncType() != nil
}

////////////////////////////////////////////////////////////////
func (n Node) FieldList() *ast.FieldList {
	if n, ok := n.an.(*ast.FieldList); ok {
		return n
	} else {
		return nil
	}
}

func (n Node) IsFieldList() bool {
	return n.FieldList() != nil
}

////////////////////////////////////////////////////////////////
func (n Node) Field() *ast.Field {
	if n, ok := n.an.(*ast.Field); ok {
		return n
	} else {
		return nil
	}
}

func (n Node) IsField() bool {
	return n.Field() != nil
}

////////////////////////////////////////////////////////////////
func (n Node) TypeSpec() *ast.TypeSpec {
	if n, ok := n.an.(*ast.TypeSpec); ok {
		return n
	} else {
		return nil
	}
}

func (n Node) IsTypeSpec() bool {
	return n.TypeSpec() != nil
}

////////////////////////////////////////////////////////////////
func (n Node) File() *ast.File {
	if n, ok := n.an.(*ast.File); ok {
		return n
	} else {
		return nil
	}
}

func (n Node) IsFile() bool {
	return n.File() != nil
}
