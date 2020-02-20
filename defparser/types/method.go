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
	return hash(m.String())
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
