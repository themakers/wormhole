package types

var _ Type = &Method{}

type Method struct {
	Name string

	// If Untyped, then it's a part of interface type definition
	Receiver *Definition

	Signature *Function
}

func (m *Method) Hash() Sum {
	return m.hash(map[Type]bool{})
}

func (m *Method) hash(prev map[Type]bool) Sum {
	return sum("METHOD", m.Receiver.hash(prev), m.Signature.hash(prev))
}

func (m *Method) String() string {
	return stringify(m)
}
