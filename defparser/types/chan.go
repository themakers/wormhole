package types

var _ Type = &Chan{}

type Chan struct {
	Type Type
}

func (c *Chan) Hash() Sum {
	return c.hash(map[Type]bool{})
}

func (c *Chan) hash(prev map[Type]bool) Sum {
	t := c.Type.hash(prev)
	return sum([]byte("CHAN"), t[:])
}

func (c *Chan) String() string {
	return stringify(c)
}
