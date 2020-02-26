package types

var _ Type = &Chan{}

type Chan struct {
	Type Type
}

func (c *Chan) Hash() string {
	return c.hash(map[Type]bool{})
}

func (c *Chan) hash(prev map[Type]bool) string {
	return sum(sum("CHAN") + c.Type.hash(prev))
}

func (c *Chan) String() string {
	return stringify(c)
}
