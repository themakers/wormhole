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
	k := m.Key.hash(prev)
	v := m.Value.hash(prev)
	return sum([]byte("MAP"), k[:], v[:])
}

func (m *Map) String() string {
	return stringify(m)
}
