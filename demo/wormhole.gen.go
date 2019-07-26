package main

import (
	"context"
	"reflect"

	"github.com/themakers/wormhole/wormhole"
)

/****************************************************************
** Greeter Client
********/

var _ Greeter = (*wormholeGreeterClientImpl)(nil)

type wormholeGreeterClientImpl struct {
	peer wormhole.RemotePeer
}

func AcquireGreeter(peer wormhole.RemotePeer) Greeter {
	return &wormholeGreeterClientImpl{peer: peer}
}

func (impl *wormholeGreeterClientImpl) Hello(ctx context.Context, arg GreeterHelloReq) (ret GreeterHelloResp, err error) {
	return ret, impl.peer.(wormhole.RemotePeerGenerated).MakeRootOutgoingCall("Greeter", "Hello", reflect.TypeOf(impl.Hello), ctx, arg, &ret)
}

/****************************************************************
** Greeter Handler
********/

func RegisterGreeterHandler(peer wormhole.LocalPeer, constructor func(caller wormhole.RemotePeer) Greeter) {
	peer.(wormhole.LocalPeerGenerated).RegisterInterface("Greeter", func(caller wormhole.RemotePeer) {
		ifc := constructor(caller)
		val := reflect.ValueOf(ifc)

		caller.(wormhole.RemotePeerGenerated).RegisterRootRef("Greeter", "Hello", val.MethodByName("Hello"))
	})
}
