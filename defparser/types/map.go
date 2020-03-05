package types

var _ Type = &Map{}

type Map struct {
	Key   Type
	Value Type
}

func (m *Map) Hash() Sum {
	return m.hash(map[Type]bool{})
}

func (m *Map) hash(prev map[Type]bool) Sum {
	return sum("MAP", m.Key.hash(prev), m.Value.hash(prev))
}

func (m *Map) String() string {
	return stringify(m)
}
