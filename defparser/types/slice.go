package types

import "fmt"

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

const sliceTmpl = "[]%s"

func (s *Slice) String() string {
	return fmt.Sprintf(
		sliceTmpl,
		s.Type,
	)
}
