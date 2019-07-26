package remote_peer

import (
	"context"
	"fmt"
	"github.com/themakers/wormhole/wormhole/internal/base"
	"github.com/themakers/wormhole/wormhole/internal/proto"
	"log"
	"reflect"

	"github.com/rs/xid"
	"go.uber.org/zap"
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
	Close()
}

type RemotePeerGenerated interface {
	RegisterRootRef(ifc, method string, mval reflect.Value)
	MakeRootOutgoingCall(ifc, method string, mtype reflect.Type, ctx context.Context, arg, ret interface{}) error
}

/****************************************************************
** IMPL
********/

func NewRemotePeer(log *zap.Logger, dc base.DataChannel) RemotePeer {
	rp := &remotePeer{
		log: log,
		dc:  dc,
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
	log *zap.Logger

	dc base.DataChannel

	refs          refs
	outgoingCalls outgoingCalls
}

/****************************************************************
** Code for interfacing with generated code
********/

func (rp *remotePeer) RegisterRootRef(ifc, method string, ref reflect.Value) {
	methodName := fmtMethod(ifc, method)

	rp.log.Info(">> Registering val", zap.String("name", methodName), zap.Stringer("ref", ref))

	rp.refs.put(methodName, ref, true)
}

func (rp *remotePeer) MakeRootOutgoingCall(ifc, method string, mt reflect.Type, ctx context.Context, arg, ret interface{}) error {
	var inVals []reflect.Value

	for i := 0; i < reflect.ValueOf(arg).NumField(); i++ {
		inVals = append(inVals, reflect.ValueOf(arg).Field(i))
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

/****************************************************************
** Package level interface
********/

func (rp *remotePeer) Close() {
	rp.dc.Close()
}

func (rp *remotePeer) ReceiverWorker() error {
	rp.log.Debug("Running remote peer handler")

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
				rp.log.Error("Error reading message from channel", zap.Error(err))
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
				// TODO
				// rp.log.Warn("",)
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
		rp.log.Panic("shit happened", zap.Stringer("msg", reflect.TypeOf(msg)))
	}
}

/****************************************************************
** Outgoing outgoingCall
********/

func (rp *remotePeer) makeOutgoingCall(methodName string, mtype reflect.Type, root bool, ctx context.Context, ins []reflect.Value) (outs []reflect.Value, remoteErr string, err error) {

	protoCall := &proto.CallMsg{}

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
			protoCall.Vars = append(protoCall.Vars, inV)
		case reflect.Func:
			methodName := xid.New().String()
			rp.refs.put(methodName, inV, false)
			defer rp.refs.del(methodName)
			protoCall.Vars = append(protoCall.Vars, reflect.ValueOf(methodName))
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
		log.Printf("RESSSSSS! %#v %#v", result.Vals, result.Error)

		//> Convert results
		outs = rp.valsRemote2Local(methodRetTypes(mtype, root), result.Vals)
		remoteErr = result.Error
	}

	return
}

func (rp *remotePeer) handleIncomingResult(result *proto.ResultMsg) error {
	rp.log.Debug("Got INCOMING res", zap.Any("result", result))

	if call, ok := rp.outgoingCalls.get(result.Call); ok {
		call.res <- result.Result
	} else {
		err := fmt.Errorf("Pending outgoingCall not found: %s", result.Call)

		rp.log.DPanic("Error handling results", zap.Error(err))

		return err
	}
	return nil
}

/****************************************************************
** Incoming call
********/

func (rp *remotePeer) handleIncomingCall(call *proto.CallMsg) error {
	rp.log.Debug("Got INCOMING call", zap.Any("call", call))

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
		rp.log.DPanic("Error handling outgoingCall", zap.Error(err))
		return err
	}
}

/****************************************************************
** Magic
********/

func (rp *remotePeer) valsRemote2Local(types []reflect.Type, values []reflect.Value) (vars []reflect.Value) {
	for i, in := range values {
		inT := types[i]

		switch inT.Kind() {
		case reflect.Func:
			vars = append(vars, rp.makeFuncThatMakesOutgoingCall(in.String(), inT))
		default:
			inV := reflect.New(inT).Elem()
			inV.Set(in.Convert(inT))
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
		outs, remoteErr, err := rp.makeOutgoingCall(ref, t, false, ctx, ins[1:])

		log.Printf("1234567890 %#v - %#v - %#v", outs, remoteErr, err)

		//> Last result is always an error
		if err != nil {
			return populateDefaultOuts(t, err)
		}
		if remoteErr != "" {
			return append(outs, reflect.ValueOf(error(&RemoteError{Text: remoteErr})))
		} else {
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
	log.Println(root, t, types)
	return
}

type RemoteError struct {
	Text string
}

func (e *RemoteError) Error() string {
	return e.Text
}

/*


* l register ref

* l remote outgoingCall (root?)
* l args - runtime->proto (root? +register refs)

* l send outgoingCall
* r recv outgoingCall

* r args - proto->runtime (+create funcs)
* r runtime outgoingCall (root?)
* r rets - runtime->proto (root?/ +register refs)

* r send resp
* l recv resp

* l rets - proto->runtime (+create funcs)

* l repeat


 */
