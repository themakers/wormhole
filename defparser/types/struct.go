package types

var (
	_ Type     = &Struct{}
	_ Selector = &Struct{}
)

// Struct represent Go's structs.
type Struct struct {
	// Structure fields in the right order.
	Fields []*StructField

	// Fast access alternative to Struct.Fields with field names as keys.
	FieldsMap map[string]*StructField

	// All fields that considered embedded.
	// Struct.Fields includes Struct.Embedded.
	Embedded []Selector
}

func (s *Struct) Select(name string) (Type, error) {
	if field, ok := s.FieldsMap[name]; ok {
		return field.Type, nil
	}

	var res Type
	for _, s := range s.Embedded {
		t, err := s.Select(name)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return nil, ErrAmbigiousSelector{
				Sel: name,
			}
		}
		res = t
	}

	return res, nil
}

func (s *Struct) Hash() string {
	return s.hash(map[Type]bool{})
}

func (s *Struct) hash(prev map[Type]bool) string {
	res := sum("STRUCT")
	for _, field := range s.Fields {
		res += sum(field.Name) + field.Type.hash(prev) + sum(field.Tag)
	}
	return sum(res)
}

func (s *Struct) String() string {
	return stringify(s)
}

type (
	StructField struct {
		Name     string
		Tag      string
		Exported bool
		Embedded bool
		Type     Type
	}
)
