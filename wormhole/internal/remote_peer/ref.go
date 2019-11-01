package remote_peer

import (
	"context"
	"reflect"
	"strings"
)

type ref struct {
	name string
	val  reflect.Value

	rp *remotePeer
}

func (r ref) root() bool {
	return strings.Contains(r.name, ".")
}

func (r ref) returnsRef() bool {
	for _, o := range methodRetTypes(r.val.Type(), r.root()) {
		if o.Kind() == reflect.Func {
			return true
		}
	}
	return false
}

func (r ref) call(ctx context.Context, args []reflect.Value) ([][]reflect.Value, error) {

	if r.root() {
		t := r.val.Type().In(1)
		arg := reflect.New(t).Elem()

		for i := 0; i < t.NumField(); i++ {
			arg.Field(i).Set(args[i])
		}

		args = []reflect.Value{arg}
	}

	retsRaw := r.val.Call(append(
		[]reflect.Value{reflect.ValueOf(ctx)},
		args...,
	))

	var err error
	errVal := retsRaw[len(retsRaw)-1]
	if errVal.IsValid() && errVal.CanInterface() /*&& !errVal.IsNil()*/ {
		if e, ok := errVal.Interface().(error); ok {
			err = e
		}
	}

	var rets [][]reflect.Value

	if r.root() {
		ret := retsRaw[0]

		t := ret.Type()
		rets = make([][]reflect.Value, t.NumField())

		for i := 0; i < t.NumField(); i++ {
			rets[i] = []reflect.Value{reflect.ValueOf(t.Field(i).Name), ret.Field(i)}
		}
	} else {
		retsRaw = retsRaw[:len(retsRaw)-1]

		for _, v := range retsRaw {
			rets = append(rets, []reflect.Value{reflect.ValueOf(""), v})
		}
	}

	return rets, err
}
