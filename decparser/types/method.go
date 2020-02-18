package types

var _ Type = &Method{}

type Method struct {
	Type      Type
	Signature *Function
}

func (m *Method) Hash() string {
	return string(
		hash.Sum([]byte(m.String())),
	)
}

func (m *Method) String() string {
	return m.Signature.String()
}
