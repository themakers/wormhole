package types

import "fmt"

var _ Type = &Map{}

type Map struct {
	Key   Type
	Value Type
}

func (m *Map) Hash() string {
	return m.hash(map[*Definition]bool{})
}

func (m *Map) hash(prev map[*Definition]bool) string {
	return sum(sum("MAP") + m.Key.hash(prev) + m.Value.hash(prev))
}

const mapTmpl = "map[%s]%s"

func (m *Map) String() string {
	return fmt.Sprintf(
		mapTmpl,
		m.Key,
		m.Value,
	)
}
