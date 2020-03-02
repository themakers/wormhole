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
	r := m.Receiver.hash(prev)
	s := m.Signature.hash(prev)
	return sum(
		[]byte("METHOD"),
		r[:],
		s[:],
	)
}

func (m *Method) String() string {
	return stringify(m)
}
