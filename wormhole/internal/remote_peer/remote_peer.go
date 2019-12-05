package remote_peer

import (
	"context"
	"fmt"
	"github.com/themakers/wormhole/wormhole/internal/base"
	"github.com/themakers/wormhole/wormhole/internal/proto"
	"reflect"

	"github.com/rs/xid"
)

func fmtMethod(ifc, method string) string {
	return fmt.Sprintf("%s.%s", ifc, method)
}

/****************************************************************
** InterfacesMap
********/

type RemotePeer interface {
	Context() context.Context
	ReceiverWorker() error

	SetState(state interface{})
	GetState() (state interface{})

	Close()
}

type RemotePeerGenerated interface {
	RegisterRootRef(ifc, method string, mval reflect.Value)
	MakeRootOutgoingCall(ifc, method string, mtype reflect.Type, ctx context.Context, arg, ret interface{}) error
}

/****************************************************************
** IMPL
********/

func NewRemotePeer(dc base.DataChannel) RemotePeer {
	rp := &remotePeer{
		dc: dc,
	}

	rp.refs.rp = rp
	rp.refs.refs = map[string]ref{}

	rp.outgoingCalls.calls = map[string]*outgoingCall{}

	return rp
}

var (
	_ RemotePeer          = new(remotePeer)
	_ RemotePeerGenerated = new(remotePeer)
)

type remotePeer struct {
	dc base.DataChannel

	refs          refs
	outgoingCalls outgoingCalls

	state interface{}
}

/****************************************************************
** Code for interfacing with generated code
********/

func (rp *remotePeer) RegisterRootRef(ifc, method string, ref reflect.Value) {
	methodName := fmtMethod(ifc, method)

	rp.refs.put(methodName, ref, true)
}

func (rp *remotePeer) MakeRootOutgoingCall(ifc, method string, mt reflect.Type, ctx context.Context, arg, ret interface{}) error {
	var inVals [][]reflect.Value

	for i := 0; i < reflect.ValueOf(arg).NumField(); i++ {
		inVals = append(inVals, []reflect.Value{
			reflect.ValueOf(reflect.TypeOf(arg).Field(i).Name),
			reflect.ValueOf(arg).Field(i),
		})
	}

	outVals, remoteErr, err := rp.makeOutgoingCall(fmtMethod(ifc, method), mt, true, ctx, inVals)
	if err != nil {
		return err
	}

	retVal := reflect.ValueOf(ret).Elem()
	for i := 0; i < retVal.NumField(); i++ {
		retVal.Field(i).Set(outVals[i])
	}

	if remoteErr != "" {
		return error(&RemoteError{Text: remoteErr})
	}

	return nil
}

/****************************************************************
** Public interface
********/

func (rp *remotePeer) Context() context.Context {
	return rp.dc.Context()
}

func (rp *remotePeer) SetState(state interface{}) {
	rp.state = state
}

func (rp *remotePeer) GetState() interface{} {
	return rp.state
}

/****************************************************************
** Package level interface
********/

func (rp *remotePeer) Close() {
	rp.dc.Close()
}

func (rp *remotePeer) ReceiverWorker() error {
	var (
		ctx, cancel = context.WithCancel(rp.dc.Context())
		msgsCh      = make(chan interface{}, 128)
		errorsCh    = make(chan error)
	)
	defer cancel()

	go (func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg, err := rp.dc.ReadMessage()
			if err != nil {
				errorsCh <- err
				return
			}
			msgsCh <- msg
		}
	})()

	for {
		select {
		case err := <-errorsCh:
			if err != nil {
				return err
			} else {
				// TODO Warn
			}
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-msgsCh:
			go rp.handleProtocolMessage(msg, errorsCh)
		}
	}
}
func (rp *remotePeer) handleProtocolMessage(msg interface{}, errorsCh chan error) {
	switch msg := msg.(type) {
	case *proto.CallMsg:
		if err := rp.handleIncomingCall(msg); err != nil {
			errorsCh <- err
		}
	case *proto.ResultMsg:
		if err := rp.handleIncomingResult(msg); err != nil {
			errorsCh <- err
		}
	default:
		panic("shit happened")
	}
}

/****************************************************************
** Outgoing outgoingCall
********/

func (rp *remotePeer) makeOutgoingCall(methodName string, mtype reflect.Type, root bool, ctx context.Context, ins [][]reflect.Value) (outs []reflect.Value, remoteErr string, err error) {

	protoCall := &proto.CallMsg{}

	for _, inV := range ins {
		inT := inV[1].Type()
		switch inT.Kind() {
		case reflect.String,
			reflect.Array, reflect.Slice,
			reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Struct:
			protoCall.Vars = append(protoCall.Vars, inV)
		case reflect.Func:
			methodName := xid.New().String()
			rp.refs.put(methodName, inV[1], false)
			defer rp.refs.del(methodName)
			protoCall.Vars = append(protoCall.Vars, []reflect.Value{inV[0], reflect.ValueOf(methodName)})
		}
	}

	callID := xid.New().String()
	outgoingCall := rp.outgoingCalls.put(callID, methodName)
	defer rp.outgoingCalls.del(callID) // FIXME Replace with context

	// TODO Send 'cancelled' message with call id when context cancelled

	protoCall.Ref = methodName
	protoCall.ID = callID

	err = rp.dc.WriteMessage(protoCall)
	if err != nil {
		return
	}

	select {
	case result := <-outgoingCall.res:
		//for i, _ /*res*/ := range result.Vals {
		//	v := reflect.New(mtype.Out(i))
		//	//unmarshal(res, v.Elem())
		//	//res.
		//	outs = append(outs, v.Elem())
		//}

		//> Convert results
		outs = rp.valsRemote2Local(methodRetTypes(mtype, root), result.Vals)
		remoteErr = result.Error
	}

	return
}

func (rp *remotePeer) handleIncomingResult(result *proto.ResultMsg) error {

	if call, ok := rp.outgoingCalls.get(result.Call); ok {
		call.res <- result.Result
	} else {
		err := fmt.Errorf("pending outgoingCall not found: %s", result.Call)
		return err
	}
	return nil
}

/****************************************************************
** Incoming call
********/

func (rp *remotePeer) handleIncomingCall(call *proto.CallMsg) error {

	if ref, ok := rp.refs.get(call.Ref); ok {

		args := rp.valsRemote2Local(methodArgTypes(ref.val.Type(), ref.root()), call.Vars)

		// TODO handle incoming context options
		ctx := context.TODO()

		resultVals, err := ref.call(ctx, args)

		// FIXME !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// FIXME Register returned refs until context cancelled

		result := proto.Result{
			Vals: resultVals,
		}

		if err != nil {
			result.Error = err.Error()
		}

		return rp.dc.WriteMessage(&proto.ResultMsg{
			Call:   call.ID,
			Result: result,
		})

	} else {
		err := fmt.Errorf("ref not found: %s", call.Ref)
		return err
	}
}

/****************************************************************
** Magic
********/

func (rp *remotePeer) valsRemote2Local(types []reflect.Type, values [][]reflect.Value) (vars []reflect.Value) {
	for i, in := range values {
		inT := types[i]

		switch inT.Kind() {
		case reflect.Func:
			vars = append(vars, rp.makeFuncThatMakesOutgoingCall(in[1].String(), inT))
		default:
			inV := reflect.New(inT).Elem()
			inV.Set(in[1].Convert(inT))
			vars = append(vars, inV)
		}
	}
	return
}

func (rp *remotePeer) makeFuncThatMakesOutgoingCall(ref string, t reflect.Type) reflect.Value {

	populateDefaultOuts := func(t reflect.Type, err error) (vals []reflect.Value) {
		vals = make([]reflect.Value, t.NumOut())
		for i := 0; i < t.NumOut()-1; i++ {
			vals[i] = reflect.New(t.Out(i)).Elem()
		}
		vals[len(vals)-1] = reflect.ValueOf(err)
		return
	}

	return reflect.MakeFunc(t, func(ins []reflect.Value) []reflect.Value {
		//> First argument is always a Context
		ctx := ins[0].Interface().(context.Context)

		var insProto [][]reflect.Value
		for _, v := range ins[1:] {
			insProto = append(insProto, []reflect.Value{reflect.ValueOf(""), v})
		}

		outs, remoteErr, err := rp.makeOutgoingCall(ref, t, false, ctx, insProto)

		//> Last result is always an error
		if err != nil {
			return populateDefaultOuts(t, err)
		}
		if remoteErr != "" {
			return append(outs, reflect.ValueOf(error(&RemoteError{Text: remoteErr})))
		} else {
			var err error
			return append(outs, reflect.ValueOf(&err).Elem())
		}
	})
}

func methodArgTypes(t reflect.Type, root bool) (types []reflect.Type) {
	if root {
		t := t.In(1)
		for i := 0; i < t.NumField(); i++ {
			types = append(types, t.Field(i).Type)
		}
	} else {
		for i := 1; i < t.NumIn(); i++ {
			types = append(types, t.In(i))
		}
	}
	return
}

func methodRetTypes(t reflect.Type, root bool) (types []reflect.Type) {
	if root {
		t := t.Out(0)
		for i := 0; i < t.NumField(); i++ {
			types = append(types, t.Field(i).Type)
		}
	} else {
		for i := 0; i < t.NumOut()-1; i++ {
			types = append(types, t.Out(i))
		}
	}

	return
}

type RemoteError struct {
	Text string
}

func (e *RemoteError) Error() string {
	return e.Text
}
