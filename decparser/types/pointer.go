package types

import "fmt"

type Pointer struct {
	Type Type
}

func (p *Pointer) Hash() string {
	return string(
		hash.Sum([]byte(p.String())),
	)
}

const pointerTmpl = "ptr*%s*"

func (p *Pointer) String() string {
	return fmt.Sprintf(
		pointerTmpl,
		p.Type,
	)
}
