package wormhole

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/rs/xid"
	"go.uber.org/zap"
)

/****************************************************************
** Interfaces
********/

type RemotePeer interface {
	Context() context.Context
	run() error
	close()
}

type RemotePeerGenerated interface {
	RegisterMethod(ifc, method string, mval reflect.Value)
	MakeOutgoingCall(ifc, method string, mtype reflect.Type, ins []interface{}, outs []interface{})
}

/****************************************************************
** IMPL
********/

func newRemotePeer(log *zap.Logger, dc DataChannel) RemotePeer {
	rp := &remotePeer{
		log: log,
		dc:  dc,
	}

	rp.methods.methods = map[string]reflect.Value{}
	rp.calls.calls = map[string]*Call{}

	return rp
}

var (
	_ RemotePeer          = new(remotePeer)
	_ RemotePeerGenerated = new(remotePeer)
)

type remotePeer struct {
	log     *zap.Logger
	dc      DataChannel
	methods rpMethods
	calls   rpCalls
}

/****************************************************************
** Interface with generated code
********/

func (rp *remotePeer) RegisterMethod(ifc, method string, mval reflect.Value) {
	methodName := fmtMethod(ifc, method)
	rp.log.Info(">> Registering method", zap.String("name", methodName), zap.Stringer("mval", mval))
	rp.methods.put(methodName, mval)
}

func (rp *remotePeer) MakeOutgoingCall(ifc, method string, mtype reflect.Type, ins []interface{}, outs []interface{}) {
	inVs := []reflect.Value{}
	for _, in := range ins {
		inVs = append(inVs, reflect.ValueOf(in))
	}

	outVs := rp.makeOutgoingCall(fmtMethod(ifc, method), mtype, inVs)

	for i, outV := range outVs {
		reflect.ValueOf(outs[i]).Elem().Set(outV)
	}
}

/****************************************************************
** Public interface
********/

func (rp *remotePeer) Context() context.Context {
	return rp.dc.Context()
}

/****************************************************************
** Package level interface
********/

func (rp *remotePeer) close() {
	rp.dc.Close()
}

func (rp *remotePeer) run() error {
	rp.log.Debug("Running remote peer handler")
	for {
		msg, err := rp.dc.ReadMessage()
		if err != nil {
			rp.log.DPanic("Error reading message from channel", zap.Error(err))
			return err
		}

		go (func() error {
			switch msg := msg.(type) {
			case *Call:
				if err := rp.handleIncomingCall(msg); err != nil {
					return err
				}
			case *Result:
				if err := rp.handleIncomingResult(msg); err != nil {
					return err
				}
			}
			return nil
		})()
	}
}

/****************************************************************
** Private interface
********/

func (rp *remotePeer) makeOutgoingCall(methodName string, mtype reflect.Type, ins []reflect.Value) (outs []reflect.Value) {
	callID := xid.New().String()

	call := &Call{
		ID:      callID,
		Method:  methodName,
		results: make(chan []json.RawMessage, 1),
	}

	for _, inV := range ins {
		inT := inV.Type()
		switch inT.Kind() {
		case reflect.String,
			reflect.Array, reflect.Slice,
			reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Struct:
			call.Vars = append(call.Vars, marshal(inV.Interface()))
		case reflect.Func:
			methodName := xid.New().String()
			rp.methods.put(methodName, inV)
			defer rp.methods.del(methodName)
			call.Vars = append(call.Vars, marshal(methodName))
		}
	}

	rp.calls.put(callID, call)
	defer rp.calls.del(callID)

	if err := rp.dc.WriteMessage(call); err != nil {
		panic(err)
	}

	select {
	case results := <-call.results:
		for i, res := range results {
			v := reflect.New(mtype.Out(i))
			unmarshal(res, v.Elem())
			outs = append(outs, v.Elem())
		}
	}

	return
}

func (rp *remotePeer) sendResults(callID string, results []reflect.Value) error {
	res := &Result{
		Call: callID,
	}

	for _, r := range results {
		res.Vars = append(res.Vars, marshal(r.Interface()))
	}

	return rp.dc.WriteMessage(res)
}

func (rp *remotePeer) handleIncomingCall(call *Call) error {
	rp.log.Debug("Got INCOMING Call", zap.Any("call", call))
	if method, ok := rp.methods.get(call.Method); ok {
		args := []reflect.Value{}
		for i, in := range call.Vars {
			inT := method.Type().In(i)
			switch inT.Kind() {
			case reflect.Func:
				args = append(args, reflect.MakeFunc(inT, func(ins []reflect.Value) (outs []reflect.Value) {
					var name string
					unmarshal(in, &name)
					return rp.makeOutgoingCall(name, inT, ins)
				}))
			default:
				inV := reflect.New(inT)

				unmarshal(in, inV.Elem())

				args = append(args, inV.Elem())
			}
		}

		results := method.Call(args)

		return rp.sendResults(call.ID, results)
	} else {
		err := fmt.Errorf("Method not found: %s", call.Method)
		rp.log.DPanic("Error handling call", zap.Error(err))
		return err
	}
}

func (rp *remotePeer) handleIncomingResult(result *Result) error {
	rp.log.Debug("Got INCOMING Results", zap.Any("result", result))
	if call, ok := rp.calls.get(result.Call); ok {
		call.results <- result.Vars
	} else {
		err := fmt.Errorf("Pending call not found: %s", result.Call)
		rp.log.DPanic("Error handling results", zap.Error(err))
		return err
	}
	return nil
}

func fmtMethod(ifc, method string) string {
	return fmt.Sprintf("%s.%s", ifc, method)
}
