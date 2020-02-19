package types

import "fmt"

var _ Type = &Map{}

type Map struct {
	Key   Type
	Value Type
}

func (m *Map) Hash() string {
	return string(
		hash.Sum([]byte(m.String())),
	)
}

const mapTmpl = "map[%s]%s"

func (m *Map) String() string {
	return fmt.Sprintf(
		mapTmpl,
		m.Key,
		m.Value,
	)
}
