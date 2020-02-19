package types

import "fmt"

var _ Type = &Slice{}

type Slice struct {
	Type Type
}

func (s *Slice) Hash() string {
	return string(
		hash.Sum([]byte(s.String())),
	)
}

const sliceTmpl = "[]%s"

func (s *Slice) String() string {
	return fmt.Sprintf(
		sliceTmpl,
		s.Type,
	)
}
