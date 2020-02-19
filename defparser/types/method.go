package types

import "fmt"

var _ Type = &Method{}

type Method struct {
	Name      string
	Type      Type
	Signature *Function
}

func (m *Method) Hash() string {
	return hash(m.String())
}

const methodTmpl = "%s."

func (m *Method) String() string {
	return fmt.Sprintf(
		methodTmpl,
		m.Signature.String(),
	)
}
