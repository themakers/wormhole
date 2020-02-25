package types

var _ Type = &Slice{}

type Slice struct {
	Type Type
}

func (s *Slice) Hash() string {
	return s.hash(map[*Definition]bool{})
}

func (s *Slice) hash(prev map[*Definition]bool) string {
	return sum(sum("SLICE") + s.Type.hash(prev))
}

func (s *Slice) String() string {
	return stringify(s)
}
