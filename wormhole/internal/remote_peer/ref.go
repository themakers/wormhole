package remote_peer

import (
	"context"
	"log"
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

func (r ref) call(ctx context.Context, args []reflect.Value) ([]reflect.Value, error) {

	if r.root() {
		t := r.val.Type().In(1)
		arg := reflect.New(t).Elem()

		for i := 0; i < t.NumField(); i++ {
			arg.Field(i).Set(args[i])
		}

		args = []reflect.Value{arg}
	}

	log.Println("CALL!", r.name, ctx, args)
	rets := r.val.Call(append(
		[]reflect.Value{reflect.ValueOf(ctx)},
		args...,
	))

	var err error
	errVal := rets[len(rets)-1]
	if errVal.IsValid() && errVal.CanInterface() /*&& !errVal.IsNil()*/ {
		if e, ok := errVal.Interface().(error); ok {
			err = e
		}
	}

	if r.root() {
		ret := rets[0]

		t := ret.Type()
		rets = make([]reflect.Value, t.NumField())

		for i := 0; i < t.NumField(); i++ {
			rets[i] = ret.Field(i)
		}
	} else {
		rets = rets[:len(rets)-1]
	}

	return rets, err
}

//
//func (r ref) argTypes() (types []reflect.Type) {
//	if r.root() {
//		for i := 0; i < r.val.Type().In(1).NumField(); i++ {
//			types = append(types, r.val.Type().In(1).Field(i).Type)
//		}
//	} else {
//		for i := 1; i < r.val.Type().NumIn(); i++ {
//			types = append(types, r.val.Type().In(i))
//		}
//	}
//}
//
//func (r ref) retTypes() (types []reflect.Type) {
//	if r.root() {
//		for i := 0; i < r.val.Type().Out(0).NumField(); i++ {
//			types = append(types, r.val.Type().Out(0).Field(i).Type)
//		}
//	} else {
//		for i := 1; i < r.val.Type().NumOut(); i++ {
//			types = append(types, r.val.Type().Out(i))
//		}
//	}
//
//}
