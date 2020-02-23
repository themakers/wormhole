package types

import "fmt"

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

const pointerTmpl = "ptr*%s*"

func (p *Pointer) String() string {
	return fmt.Sprintf(
		pointerTmpl,
		p.Type,
	)
}
