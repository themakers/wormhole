package types

import "errors"

var _ Type = Byte

const (
	Int Builtin = iota
	Int32
	Int64
	Uint
	Uint32
	Uint64
	Byte
	String
	Rune
	Bool
	Error
)

type Builtin uint

func (b Builtin) Hash() Sum {
	return b.hash(nil)
}

func (b Builtin) hash(_ map[Type]bool) Sum {
	return sum([]byte("BUILTIN"), []byte(b.String()))
}

func (b Builtin) String() string {
	return stringify(b)
}

func String2Builtin(s string) (res Builtin, err error) {
	switch s {
	case "int":
		res = Int
	case "int32":
		res = Int32
	case "int64":
		res = Int64
	case "uint":
		res = Uint
	case "uint32":
		res = Uint32
	case "uint64":
		res = Uint64
	case "byte":
		res = Byte
	case "string":
		res = String
	case "rune":
		res = Rune
	case "bool":
		res = Bool
	case "error":
		res = Error
	default:
		err = errors.New(
			"There's no builtin type that matches input string",
		)
	}
	return
}
