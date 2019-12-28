package parsex

import (
	"fmt"
	"strings"
)

type Parsed struct {
	Pkg           string
	Interfaces    []*Interface
	InterfacesMap map[string]*Interface
}

type Interface struct {
	Name       string
	Methods    []*Method
	MethodsMap map[string]*Method
}

type Method struct {
	Interface string
	Name      string
	Args      []Param
	Rets      []Param
}

func (m Method) ArgsList() string {
	var args []string

	for _, a := range m.Args {
		for _, a := range a.Names {
			args = append(args, fmt.Sprintf("%s", a))
		}
	}

	return strings.Join(args, ", ")
}

func (m Method) Ins() (ins []string) {

	for _, a := range m.Args {
		for _, a := range a.Names {
			ins = append(ins, a)
		}
	}

	return
}

func (m Method) RetsList() string {
	var rets []string

	for _, a := range m.Rets {
		for _, a := range a.Names {
			rets = append(rets, fmt.Sprintf("%s", a))
		}
	}

	return strings.Join(rets, ", ")
}

func (m Method) Outs() (outs []string) {

	for _, a := range m.Rets {
		for _, a := range a.Names {
			outs = append(outs, a)
		}
	}

	return
}

func (m Method) String() string {
	var (
		args []string
		rets []string
	)
	for _, a := range m.Args {
		args = append(args, a.String())
	}
	for _, r := range m.Rets {
		rets = append(rets, r.String())
	}
	return fmt.Sprintf("%s(ctx context.Context, %s) (%s)", m.Name, strings.Join(args, ", "), strings.Join(rets, ", "))
}

type Param struct {
	Names []string
	Type  string
}

func (p Param) String() string {
	return fmt.Sprintf("%s %s", strings.Join(p.Names, ", "), p.Type)
}
