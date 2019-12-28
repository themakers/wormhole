package reflect_hack

import (
	"context"
	"errors"
	"fmt"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole/internal/remote_peer"
	"github.com/themakers/wormhole/wormhole/wire_io"
	"log"
	"reflect"
)

var (
	typeInt     = reflect.TypeOf(int(1))
	typeFloat64 = reflect.TypeOf(float64(1))
	typeError   = func() reflect.Type {
		var err error
		return reflect.TypeOf(&err).Elem()
	}()
)

func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func makeRefFunc(p wormhole.RemotePeerGenerated, frv reflect.Value) remote_peer.RefFunc {
	frt := frv.Type()

	return func(ctx context.Context, ar wire_io.ArrayReader, w func(int, func(remote_peer.RegisterUnnamedRefFunc, wire_io.ValueWriter))) {
		numIn := frt.NumIn()

		sz, r, err := ar()
		if err != nil {
			panic(err)
		}

		if sz != numIn-1 {
			panic(fmt.Sprintf("actual/received arguments count mismatch: %d != %d", numIn-1, sz))
		}

		var ins = make([]reflect.Value, numIn)
		ins[0] = reflect.ValueOf(ctx)

		for n := 1; n < numIn; n++ {
			inT := frt.In(n)
			inV := reflect.New(inT).Elem()
			ReadAny(p, r, inV)
			ins[n] = inV
		}

		outs := frv.Call(ins)

		w(len(outs), func(rr remote_peer.RegisterUnnamedRefFunc, w wire_io.ValueWriter) {
			for _, out := range outs {
				WriteAny(p, rr, w, out)
			}
		})
	}
}

func WriteAny(p wormhole.RemotePeerGenerated, rr remote_peer.RegisterUnnamedRefFunc, w wire_io.ValueWriter, v interface{}) {
	if v == nil {
		perror(w.WriteNil())
	} else if rv, ok := v.(reflect.Value); ok {
		if rv.IsValid() {
			writeAny(p, rr, w, rv.Type(), rv)
		} else {
			// TODO Warn?
			perror(w.WriteNil())
		}
	} else {
		writeAny(p, rr, w, reflect.TypeOf(v), reflect.ValueOf(v))
	}
}

func writeAny(p wormhole.RemotePeerGenerated, registerRef remote_peer.RegisterUnnamedRefFunc, w wire_io.ValueWriter, rt reflect.Type, rv reflect.Value) {

	if !rv.IsValid() {
		// TODO Warn?
		perror(w.WriteNil())
		return
	}

	if rt == typeError {
		if rv.IsNil() {
			perror(w.WriteNil())
			return
		}
		v := rv.MethodByName("Error").Call([]reflect.Value{})
		perror(w.WriteString(v[0].String()))
		return
	}

	switch rt.Kind() {
	case reflect.Invalid:
		panic("invalid kind")
	case reflect.Ptr:
		writeAny(p, registerRef, w, rt.Elem(), rv.Elem())
	case reflect.Func:
		perror(w.WriteString(registerRef(makeRefFunc(p, rv))))

	case reflect.Struct:
		perror(w.WriteArray(rv.NumField(), func(w wire_io.ValueWriter) error {
			for i := 0; i < rv.NumField(); i++ {
				sf := rt.Field(i)
				fv := rv.Field(i)

				if fv.IsValid() {
					writeAny(p, registerRef, w, sf.Type, fv)
				} else {
					// TODO Warn?
					perror(w.WriteNil())
				}
			}

			return nil
		}))

	case reflect.Map:
		keys := rv.MapKeys()
		perror(w.WriteMap(len(keys), func(writeEntry wire_io.MapEntryWriter) error {
			for _, key := range keys {
				fv := rv.MapIndex(key)

				perror(writeEntry(key.String(), func(w wire_io.ValueWriter) error {
					if fv.IsValid() {
						writeAny(p, registerRef, w, fv.Type(), fv)
					} else {
						// TODO Warn?
						perror(w.WriteNil())
					}

					return nil
				}))
			}

			return nil
		}))

	case reflect.Slice, reflect.Array:
		l := rv.Len()
		perror(w.WriteArray(l, func(w wire_io.ValueWriter) error {
			for i := 0; i < l; i++ {
				rv := rv.Index(i)
				if rv.IsValid() {
					writeAny(p, registerRef, w, rv.Type(), rv)
				} else {
					// TODO Warn?
					perror(w.WriteNil())
				}
			}

			return nil
		}))

	case reflect.String:
		perror(w.WriteString(rv.String()))
	case reflect.Bool:
		perror(w.WriteBoolean(rv.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		perror(w.WriteInt(int(rv.Convert(typeInt).Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		perror(w.WriteInt(int(rv.Convert(typeInt).Int())))
	case reflect.Float32, reflect.Float64:
		perror(w.WriteFloat(rv.Convert(typeFloat64).Float()))

	case reflect.Complex64, reflect.Complex128:
		panic("interface is not supported")
	case reflect.Interface:
		panic("interface is not supported")
	case reflect.Chan:
		panic("channel is not supported")
	case reflect.Uintptr:
		panic("uintptr is not supported")
	case reflect.UnsafePointer:
		panic("unsafe pointer is not supported")
	default:
		panic(fmt.Sprintf("unsupported type %v", rt))
	}
}

func makeFuncThatMakesOutgoingCall(p wormhole.RemotePeerGenerated, ft reflect.Type, ref string) reflect.Value {
	return reflect.MakeFunc(ft, func(args []reflect.Value) (results []reflect.Value) {

		results = make([]reflect.Value, ft.NumOut())
		for i := 0; i < ft.NumOut(); i++ {
			results[i] = reflect.New(ft.Out(i)).Elem()
		}

		ctx := args[0].Interface().(context.Context)

		doneCtx, done := context.WithCancel(ctx)
		defer done()

		p.Call(ref, ctx, ft.NumIn()-1, func(registerRefFunc remote_peer.RegisterUnnamedRefFunc, vw wire_io.ValueWriter) {

			for _, arg := range args[1:] {
				WriteAny(p, registerRefFunc, vw, arg)
			}

		}, func(ctx context.Context, ar wire_io.ArrayReader) {
			defer done()

			sz, vr, err := ar()
			if err != nil {
				panic(err)
			}
			if sz != ft.NumOut() {
				panic(fmt.Sprintf("return values count mismatch: %d != %d", ft.NumOut(), sz))
			}

			for i := 0; i < ft.NumOut(); i++ {
				ReadAny(p, vr, results[i])
			}

		})

		<-doneCtx.Done()

		log.Println("OOPS", ref, len(results))

		return
	})
}

func ReadAny(p wormhole.RemotePeerGenerated, r wire_io.ValueReader, v interface{}) {
	if v == nil {
		// ???
		_, err := r()
		perror(err)
	} else if rv, ok := v.(reflect.Value); ok {
		if rv.IsValid() {
			readAny(p, r, rv.Type(), rv)
		} else {
			// ???
			_, err := r()
			perror(err)
		}
	} else {
		readAny(p, r, reflect.TypeOf(v), reflect.ValueOf(v))
	}
}

func readAny(p wormhole.RemotePeerGenerated, r wire_io.ValueReader, rt reflect.Type, rv reflect.Value) {

	if !rv.IsValid() {
		// TODO Warn?
		_, err := r()
		perror(err)
		return
	}

	if rt == typeError {
		v, err := r()
		perror(err)

		if v != nil {
			rv.Set(reflect.ValueOf(errors.New(v.(string))))
		}

		return
	}

	switch rt.Kind() {
	case reflect.Invalid:
		panic("invalid kind")
	case reflect.Ptr:
		readAny(p, r, rt.Elem(), rv.Elem())
	case reflect.Func:
		// TODO
		//w.WriteString(registerRef(makeRefFunc(p, rv)))
		val, err := r()
		perror(err)

		rv.Set(makeFuncThatMakesOutgoingCall(p, rt, val.(string)))

	case reflect.Struct:
		v, err := r()
		perror(err)

		if v, ok := v.(wire_io.ArrayReader); ok {
			sz, v, err := v()
			perror(err)

			n := rv.NumField()
			if n != sz {
				panic("TODO")
			}

			for i := 0; i < n; i++ {
				sf := rt.Field(i)
				fv := rv.Field(i)

				readAny(p, v, sf.Type, fv)
			}
		} else {
			panic("TODO")
		}

	case reflect.Map:
		v, err := r()
		perror(err)

		if v, ok := v.(wire_io.MapReader); ok {
			sz, v, err := v()
			perror(err)

			var mrv reflect.Value
			if rv.IsNil() {
				mrv = reflect.MakeMapWithSize(rt, sz)
				rv.Set(mrv)
			} else {
				mrv = rv
			}

			for i := 0; i < sz; i++ {
				kt := rt.Key()
				kv := reflect.New(kt).Elem()

				readAny(p, v, kt, kv)

				vt := rt.Elem()
				vv := reflect.New(vt).Elem()

				readAny(p, v, vt, vv)

				mrv.SetMapIndex(kv, vv)
			}
		} else {
			panic("TODO")
		}

	case reflect.Slice, reflect.Array:
		v, err := r()
		perror(err)

		if v, ok := v.(wire_io.ArrayReader); ok {
			sz, v, err := v()
			perror(err)

			srv := reflect.MakeSlice(rt, sz, sz)

			for i := 0; i < sz; i++ {
				et := rt.Elem()
				ev := reflect.New(et).Elem()

				readAny(p, v, et, ev)

				srv.Index(i).Set(ev)
			}

			rv.Set(srv)
		} else {
			panic("TODO")
		}

	case reflect.String:
		v, err := r()
		perror(err)
		rv.Set(reflect.ValueOf(v).Convert(rt))
	case reflect.Bool:
		v, err := r()
		perror(err)
		rv.Set(reflect.ValueOf(v).Convert(rt))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := r()
		perror(err)
		rv.Set(reflect.ValueOf(v).Convert(rt))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := r()
		perror(err)
		rv.Set(reflect.ValueOf(v).Convert(rt))
	case reflect.Float32, reflect.Float64:
		v, err := r()
		perror(err)
		rv.Set(reflect.ValueOf(v).Convert(rt))

	case reflect.Complex64, reflect.Complex128:
		panic("interface is not supported")
	case reflect.Interface:
		panic("interface is not supported")
	case reflect.Chan:
		panic("channel is not supported")
	case reflect.Uintptr:
		panic("uintptr is not supported")
	case reflect.UnsafePointer:
		panic("unsafe pointer is not supported")
	default:
		panic(fmt.Sprintf("unsupported type %v", rt))
	}

}
