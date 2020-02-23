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

func (b Builtin) Hash() string {
	return b.hash(map[*Definition]bool{})
}

func (b Builtin) hash(_ map[*Definition]bool) string {
	return sum(sum("BUILTIN") + sum(b.String()))
}

func (b Builtin) String() string {
	switch b {
	case Int:
		return "int"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Byte:
		return "byte"
	case String:
		return "string"
	case Rune:
		return "rune"
	case Bool:
		return "bool"
	case Error:
		return "error"
	default:
		panic("Invalid builtin type")
	}
	return ""
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
