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
