package types

var _ Type = &Method{}

type Method struct {
	Name string

	// If Untyped, then it's a part of interface type definition
	Receiver *Definition

	Signature *Function
}

func (m *Method) Hash() string {
	return m.hash(map[*Definition]bool{})
}

func (m *Method) hash(prev map[*Definition]bool) string {
	return sum(sum("METHOD") +
		sum(m.Name) +
		m.Receiver.hash(prev) +
		m.Signature.hash(prev),
	)
}

func (m *Method) String() string {
	return stringify(m)
}
