package types

var _ Type = &Chan{}

type Chan struct {
	Type Type
}

func (c *Chan) Hash() string {
	return c.hash(map[*Definition]bool{})
}

func (c *Chan) hash(prev map[*Definition]bool) string {
	return sum(sum("CHAN") + c.Type.hash(prev))
}

func (c *Chan) String() string {
	return stringify(c)
}
