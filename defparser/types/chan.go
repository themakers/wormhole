package types

var _ Type = &Chan{}

type Chan struct {
	Type Type
}

func (c *Chan) Hash() Sum {
	return c.hash(map[Type]bool{})
}

func (c *Chan) hash(prev map[Type]bool) Sum {
	return sum("CHAN", c.hash(prev))
}

func (c *Chan) String() string {
	return stringify(c)
}
