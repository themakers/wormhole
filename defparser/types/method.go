package types

import "fmt"

var _ Type = &Method{}

type Method struct {
	Name string

	// If Untyped, then it's a part of interface type definition
	Receiver Type

	Signature *Function
}

func (m *Method) Hash() string {
	return m.hash(map[*Definition]bool{})
}

func (m *Method) hash(prev map[*Definition]bool) string {
	return sum(sum("METHOD") + sum(m.Name) + m.Signature.hash(prev))
}

const methodTmpl = "(%s)%s-%s"

func (m *Method) String() string {
	return fmt.Sprintf(
		methodTmpl,
		m.Receiver,
		m.Name,
		m.Signature,
	)
}
