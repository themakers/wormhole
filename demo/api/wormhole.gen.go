package api

import (
	"context"
	"reflect"
	"time"

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
** Greeter Client (KeepAlive)
********/

var _ Greeter = (*wormholeGreeterKeepAliveClientImpl)(nil)

type wormholeGreeterKeepAliveClientImpl struct {
	peer wormhole.LocalPeer
	id   string
	to   time.Duration
}

func AcquireKeepAliveGreeter(peer wormhole.LocalPeer, id string, to time.Duration) Greeter {
	return &wormholeGreeterKeepAliveClientImpl{peer: peer, id: id, to: to}
}

func (impl *wormholeGreeterKeepAliveClientImpl) Hello(ctx context.Context, arg GreeterHelloReq) (ret GreeterHelloResp, err error) {
	waitCtx, cancel := context.WithTimeout(ctx, impl.to)
	defer cancel()
	if peer := impl.peer.(wormhole.LocalPeerGenerated).WaitFor(waitCtx, impl.id); peer != nil {
		return ret, peer.(wormhole.RemotePeerGenerated).MakeRootOutgoingCall("Greeter", "Hello", reflect.TypeOf(impl.Hello), ctx, arg, &ret)
	} else {
		return ret, wormhole.ErrTimeout
	}
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
