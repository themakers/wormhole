package types

import (
	"fmt"
	"strings"
)

var _ Type = &Struct{}

type Struct struct {
	Fields    []StructField
	FieldsMap map[string]StructField
}

func (s *Struct) Hash() string {
	return string(
		hash.Sum([]byte(s.String())),
	)
}

const structTmpl = "struct{%s}"

func (s *Struct) String() string {
	fields := make([]string, len(s.Fields))

	for i, field := range s.Fields {
		fields[i] = field.Name + ":" + field.Type.String()
	}

	return fmt.Sprintf(
		structTmpl,
		strings.Join(fields, ","),
	)
}

type StructField struct {
	Name     string
	Tag      string
	Exported bool
	Type     Type
}
