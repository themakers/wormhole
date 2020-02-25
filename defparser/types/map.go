package types

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

func (m *Map) String() string {
	return stringify(m)
}
