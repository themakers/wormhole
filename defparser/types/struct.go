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
	Embedded map[string]Selector

	// Non-ambigious fields and methods, that are available
	// through embedded fields
	EmbeddedComponents EmbeddedComponents
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

func (s *Struct) Hash() Sum {
	return s.hash(map[Type]bool{})
}

func (str *Struct) hash(prev map[Type]bool) Sum {
	s := make([][]byte, 1+4*len(str.Fields))
	s[0] = []byte("STRUCT")

	i := 1
	for _, field := range str.Fields {
		s += sum(field.Name) + field.Type.hash(prev) + sum(field.Tag)
		if field.Embedded {
			s += "EMBEDDED"
		}
	}
	return sum(s)
}

func (s *Struct) String() string {
	return stringify(s)
}

type (
	StructField struct {
		Name         string
		Tag          string
		Exported     bool
		Embedded     bool
		Type         Type
		ParentStruct *Struct
	}

	EmbeddedComponents struct {
		Fields    []*StructField
		FieldsMap map[string]*StructField

		Methods    []*Method
		MethodsMap map[string]*Method
	}
)
