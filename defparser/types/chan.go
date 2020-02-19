package types

import "fmt"

var _ Type = &Chan{}

type Chan struct {
	Type Type
}

func (c *Chan) Hash() string {
	return hash(c.String())
}

const chanTmpl = "chan %s"

func (c *Chan) String() string {
	return fmt.Sprintf(chanTmpl, c.Type)
}
