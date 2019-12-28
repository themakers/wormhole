package remote_peer

import (
	"context"
	"fmt"
	"github.com/themakers/wormhole/wormhole/internal/data_channel"
	"github.com/themakers/wormhole/wormhole/wire_io"

	"github.com/rs/xid"
)

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
	Call(ref string, ctx context.Context, narg int, write func(RegisterUnnamedRefFunc, wire_io.ValueWriter), res ResultFunc)
	RegisterServiceRef(ref string, call RefFunc)
}

type RegisterUnnamedRefFunc func(call RefFunc) string

/****************************************************************
** IMPL
********/

func NewRemotePeer(dc data_channel.DataChannel) RemotePeer {
	rp := &remotePeer{
		dc: dc,
	}

	rp.refs.refs = map[string]RefFunc{}

	rp.outgoingCalls.calls = map[string]ResultFunc{}

	return rp
}

var (
	_ RemotePeer          = new(remotePeer)
	_ RemotePeerGenerated = new(remotePeer)
)

type remotePeer struct {
	dc data_channel.DataChannel

	refs          refs
	outgoingCalls outgoingCalls

	rmRootRefs []func()

	state interface{}
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
	rp.refs.remove(rp.rmRootRefs...)
	rp.dc.Close()
}

func (rp *remotePeer) ReceiverWorker() error {
	var (
		ctx, cancel = context.WithCancel(rp.dc.Context())
		errorsCh    = make(chan error)
	)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		sz, mr, err := rp.dc.MessageReader()
		if err != nil {
			return err
		}

		go func(sz int, mr wire_io.ValueReader) {
			_, err := mr() //> Version
			if err != nil {
				panic(err)
			}

			mType, err := mr()
			if err != nil {
				panic(err)
			}

			switch mType.(int) { //> Message
			case 1:
				if sz != 5 {
					panic(fmt.Sprintf("unrecognized message size: %d", sz))
				}

				call, err := mr()
				if err != nil {
					panic(err)
				}
				ref, err := mr()
				if err != nil {
					panic(err)
				}
				ar, err := mr()
				if err != nil {
					panic(err)
				}

				if err := rp.handleIncomingCall(ctx, call.(string), ref.(string), ar.(wire_io.ArrayReader)); err != nil {
					errorsCh <- err
				}
			case 2:
				if sz != 4 {
					panic(fmt.Sprintf("unrecognized message size: %d", sz))
				}

				call, err := mr()
				if err != nil {
					panic(err)
				}

				ar, err := mr()
				if err != nil {
					panic(err)
				}

				if err := rp.handleIncomingResult(ctx, call.(string), ar.(wire_io.ArrayReader)); err != nil {
					errorsCh <- err
				}
			default:
				panic("unhandled protocol message")
			}
		}(sz, mr)
	}
}

func (rp *remotePeer) handleIncomingResult(ctx context.Context, call string, ar wire_io.ArrayReader) error {

	if res := rp.outgoingCalls.get(call); res != nil {
		res(ctx, ar)
	} else {
		err := fmt.Errorf("pending outgoing call not found, but result was received: %s", call)
		return err
	}

	return nil
}

func (rp *remotePeer) handleIncomingCall(ctx context.Context, call, ref string, ar wire_io.ArrayReader) error {

	if callRef := rp.refs.get(ref); callRef != nil {

		callRef(ctx, ar, func(n int, wf func(RegisterUnnamedRefFunc, wire_io.ValueWriter)) {
			if err := rp.dc.MessageWriter(4, func(w wire_io.ValueWriter) error {
				if err := w.WriteInt(1); err != nil { //> Version
					panic(err)
				}
				if err := w.WriteInt(2); err != nil { //> Type: Result
					panic(err)
				}
				if err := w.WriteString(call); err != nil { //> Call
					panic(err)
				}
				if err := w.WriteArray(n, func(w wire_io.ValueWriter) error {
					wf(func(call RefFunc) string {
						// TODO
						panic("unimplemented")
					}, w)

					return nil
				}); err != nil {
					panic(err)
				}

				return nil
			}); err != nil {
				panic(err)
			}
		})

	} else {
		return fmt.Errorf("ref not found: %s", ref)
	}

	return nil
}

/****************************************************************
** Outgoing Call
********/

func (rp *remotePeer) Call(ref string, ctx context.Context, nIns int, write func(RegisterUnnamedRefFunc, wire_io.ValueWriter), res ResultFunc) {
	callID := xid.New().String()

	resCh, rmCall := rp.outgoingCalls.put(ctx, callID, res)
	defer rmCall()

	var rmRefs []func()

	registerRef := func(call RefFunc) string {
		ref := xid.New().String()
		rmRefs = append(rmRefs, rp.refs.put(ref, call))
		return ref
	}

	defer rp.refs.remove(rmRefs...)

	if err := rp.dc.MessageWriter(5, func(mw wire_io.ValueWriter) error {
		if err := mw.WriteInt(1); err != nil { //> Version
			panic(err)
		}
		if err := mw.WriteInt(1); err != nil { //> Call
			panic(err)
		}
		if err := mw.WriteString(callID); err != nil { //> Call ID
			panic(err)
		}
		if err := mw.WriteString(ref); err != nil { //> Ref
			panic(err)
		}
		if err := mw.WriteArray(nIns, func(w wire_io.ValueWriter) error {
			write(registerRef, w)
			return nil
		}); err != nil { //> Values
			panic(err)
		}

		return nil
	}); err != nil {
		panic(err)
	}

	<-resCh
}

func (rp *remotePeer) RegisterServiceRef(ref string, call RefFunc) {
	rp.rmRootRefs = append(rp.rmRootRefs, rp.refs.put(ref, call))
}
