package wormhole_msgp

import (
	"errors"
	"fmt"
	"github.com/themakers/wormhole/wormhole/wire_io"
	"github.com/tinylib/msgp/msgp"
	"io"
	"sync"
)

var getWriter = func() func(w io.Writer) (*msgp.Writer, func()) {
	pool := sync.Pool{
		New: func() interface{} {
			return msgp.NewWriterSize(nil, 2*1024)
		},
	}
	return func(w io.Writer) (*msgp.Writer, func()) {
		wr := pool.Get().(*msgp.Writer)
		wr.Reset(w)
		return wr, func() {
			pool.Put(wr)
		}
	}
}()

var getReader = func() func(r io.Reader) (*msgp.Reader, func()) {
	pool := sync.Pool{
		New: func() interface{} {
			return msgp.NewReaderSize(nil, 2*1024)
		},
	}
	return func(r io.Reader) (*msgp.Reader, func()) {
		rr := pool.Get().(*msgp.Reader)
		rr.Reset(r)
		return rr, func() {
			pool.Put(rr)
		}
	}
}()

////////////////////////////////////////////////////////////////
//// Handler
////

var Handler wire_io.Handler = new(handler)

type handler struct{}

func (handler) NewReader(r io.Reader) (int, wire_io.ValueReader, func(), error) {
	rr, done := getReader(r)
	sz, vr, err := newArrayReader(rr)()
	if err != nil {
		done()
		return 0, nil, nil, err
	} else {
		return sz, vr, done, err
	}
}

func (handler) NewWriter(sz int, w io.Writer, wf func(wire_io.ValueWriter) error) error {
	mw, done := getWriter(w)
	defer done()

	//mw := msgp.NewWriter(w)

	if err := mw.WriteArrayHeader(uint32(sz)); err != nil {
		return err
	}

	vw := newValueWriter(mw)

	if err := wf(vw); err != nil {
		return err
	}

	return mw.Flush()
}

////////////////////////////////////////////////////////////////
//// Reader
////

func newArrayReader(r *msgp.Reader) wire_io.ArrayReader {
	return func() (int, wire_io.ValueReader, error) {
		sz, err := r.ReadArrayHeader()
		if err != nil {
			return 0, nil, err
		}

		return int(sz), newValueReader(r), nil
	}
}

func newMapReader(r *msgp.Reader) wire_io.MapReader {
	return func() (int, wire_io.ValueReader, error) {
		sz, err := r.ReadMapHeader()
		if err != nil {
			return 0, nil, err
		}

		return int(sz), newValueReader(r), nil
	}
}

func newValueReader(r *msgp.Reader) wire_io.ValueReader {
	return func() (v interface{}, err error) {
		t, err := r.NextType()
		if err != nil {
			return nil, err
		}

		//defer func() {
		//	log.Println("READ:", v, err)
		//}()

		switch t {
		case msgp.MapType:
			return newMapReader(r), nil
		case msgp.ArrayType:
			return newArrayReader(r), nil
		case msgp.StrType:
			return r.ReadString()
		case msgp.BinType:
			return r.ReadBytes(nil)
		case msgp.Float64Type:
			return r.ReadFloat64()
		case msgp.Float32Type:
			return r.ReadFloat32()
		case msgp.BoolType:
			return r.ReadBool()
		case msgp.IntType:
			return r.ReadInt()
		case msgp.UintType:
			return r.ReadUint()
		case msgp.NilType:
			return nil, r.ReadNil()
		default:
			return nil, errors.New(fmt.Sprintf("unexpected type: %s", t.String()))
		}
	}
}

////////////////////////////////////////////////////////////////
//// Writer
////

var _ wire_io.ValueWriter = new(valueWriter)

type valueWriter struct {
	w *msgp.Writer
}

func newValueWriter(w *msgp.Writer) wire_io.ValueWriter {
	return &valueWriter{w: w}
}

func (vw *valueWriter) WriteNil() error {
	//log.Println("WriteNil")
	return vw.w.WriteNil()
}

func (vw *valueWriter) WriteString(v string) error {
	//log.Println("WriteString", v)
	return vw.w.WriteString(v)
}

func (vw *valueWriter) WriteInt(v int) error {
	//log.Println("WriteInt", v)
	return vw.w.WriteInt(v)
}

func (vw *valueWriter) WriteFloat(v float64) error {
	//log.Println("WriteFloat", v)
	return vw.w.WriteFloat64(v)
}

func (vw *valueWriter) WriteBoolean(v bool) error {
	//log.Println("WriteBoolean", v)
	return vw.w.WriteBool(v)
}

func (vw *valueWriter) WriteBinary(v []byte) error {
	//log.Println("WriteBinary", v)
	return vw.w.WriteBytes(v)
}

func (vw *valueWriter) WriteArray(sz int, wfn func(wire_io.ValueWriter) error) error {
	//log.Println("WriteArray", sz)
	if err := vw.w.WriteArrayHeader(uint32(sz)); err != nil {
		return err
	}

	return wfn(newValueWriter(vw.w))
}

func (vw *valueWriter) WriteMap(sz int, wfn func(wire_io.MapEntryWriter) error) error {
	//log.Println("WriteMap")
	if err := vw.w.WriteMapHeader(uint32(sz)); err != nil {
		return err
	}

	w := newValueWriter(vw.w)

	return wfn(func(key string, val func(wire_io.ValueWriter) error) error {
		if err := w.WriteString(key); err != nil {
			return err
		}

		if err := val(w); err != nil {
			return err
		}

		return nil
	})
}
