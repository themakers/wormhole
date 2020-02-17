package datatypes

const (
	Int64 DataType = iota
	Int32
	Int
	Uint
	Uint32
	Uint64
	Byte
	Rune
	String
	Map
	Slice
	Array
	Interface
	Type
	Struct
	Pointer
	Function
)

type DataType uint
