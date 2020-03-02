package types

type (
	Type interface {
		Hash() Sum
		hash(prev map[Type]bool) Sum
	}

	Selector interface {
		Select(string) (Type, error)
		Type
	}
)
