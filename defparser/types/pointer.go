package types

var _ Type = &Pointer{}

type Pointer struct {
	Type Type
}

func (p *Pointer) Hash() Sum {
	return p.hash(map[Type]bool{})
}

func (p *Pointer) hash(prev map[Type]bool) Sum {
	v := p.Type.hash(prev)
	return sum([]byte("POINTER"), v[:])
}

func (p *Pointer) String() string {
	return stringify(p)
}
