package wire_io

import "io"

type Reader func(r io.Reader) ValueReader

type Writer func(w io.Writer) ValueWriter

type Handler interface {
	NewReader(r io.Reader) (int, ValueReader, func(), error)
	NewWriter(sz int, w io.Writer, wf func(ValueWriter) error) error
}

////////////////////////////////////////////////////////////////
//// Writer
////

type ValueWriter interface {
	WriteNil() error
	WriteString(string) error
	WriteInt(int) error
	WriteFloat(float64) error
	WriteBoolean(bool) error
	WriteBinary([]byte) error
	WriteArray(int, func(ValueWriter) error) error
	WriteMap(int, func(MapEntryWriter) error) error
}

type MapEntryWriter func(key string, val func(ValueWriter) error) error

////////////////////////////////////////////////////////////////
//// Reader
////

type ArrayReader func() (sz int, r ValueReader, err error)
type MapReader func() (sz int, r ValueReader, err error)

type ValueReader func() (interface{}, error)

//type Type int

//const (
//	TypeAny = iota + 1
//	TypeString
//	TypeInt
//	TypeFloat
//	TypeBoolean
//	TypeBinary
//	TypeArray
//	TypeMap
//	// TypeFunc
//)

////// V2
//
//type ArrayReader2 func(func(sz int, r ValueReader2))
//type MapReader2 func(func(sz int, r ValueReader2))
//
//type ValueReader2 func(func(interface{}))

//type ValueReader3 interface {
//	Next() bool
//	Type() Type
//	Size() int
//}
//
//type ValueReader2 interface {
//	Len() int
//
//	ReadString() string
//	ReadInt() int
//	ReadFloat() float64
//	ReadBoolean() bool
//	ReadBinary() []byte
//
//	ReadArray() (int, ValueReader2)
//	ReadMap() (int, MapReader)
//
//	DiscardOne()
//}
//
//type MapReader interface {
//}
