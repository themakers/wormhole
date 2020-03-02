package types

var _ Type = &Slice{}

type Slice struct {
	Type Type
}

func (s *Slice) Hash() Sum {
	return s.hash(map[Type]bool{})
}

func (s *Slice) hash(prev map[Type]bool) Sum {
	v := s.Type.hash(prev)
	return sum([]byte("SLICE"), v[:])
}

func (s *Slice) String() string {
	return stringify(s)
}
