package types

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
	Error
)

type Builtin uint

func (b Builtin) Hash() string {
	return string(
		hash.Sum([]byte(b.String())),
	)
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
	case Error:
		return "error"
	default:
		panic("Invalid builtin type")
		return ""
	}
}

func String2Builtin(s string) Builtin {
	switch s {
	case "int":
		return Int
	case "int32":
		return Int32
	case "int64":
		return Int64
	case "uint":
		return Uint
	case "uint32":
		return Uint32
	case "uint64":
		return Uint64
	case "byte":
		return Byte
	case "string":
		return String
	case "rune":
		return Rune
	case "error":
		return Error
	default:
		panic("Invalid builtin type")
		return Byte
	}
}
