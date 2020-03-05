package types

var _ Type = &Slice{}

type Slice struct {
	Type Type
}

func (s *Slice) Hash() Sum {
	return s.hash(map[Type]bool{})
}

func (s *Slice) hash(prev map[Type]bool) Sum {
	return sum(
		"SLICE",
		s.Type.hash(prev),
	)
}

func (s *Slice) String() string {
	return stringify(s)
}
