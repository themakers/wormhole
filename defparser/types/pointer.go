package types

var _ Type = &Pointer{}

type Pointer struct {
	Type Type
}

func (p *Pointer) Hash() string {
	return p.hash(map[*Definition]bool{})
}

func (p *Pointer) hash(prev map[*Definition]bool) string {
	return sum(sum("POINTER") + p.Type.hash(prev))
}

func (p *Pointer) String() string {
	return stringify(p)
}
