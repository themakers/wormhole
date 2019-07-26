package parsex

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
	Arg       string
	Ret       string
}
