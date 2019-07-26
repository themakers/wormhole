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
	//mt, _ := reflect.TypeOf(impl).MethodByName("Hello")
	err = impl.peer.(wormhole.RemotePeerGenerated).MakeRootOutgoingCall("Greeter", "Hello", reflect.TypeOf(impl.Hello), ctx, arg, &ret)
	return
}

func (impl *wormholeGreeterClientImpl) Hello12(ctx context.Context, arg GreeterHelloReq) (ret GreeterHelloResp, err error) {
	//mt, _ := reflect.TypeOf(impl).MethodByName("Hello12")
	err = impl.peer.(wormhole.RemotePeerGenerated).MakeRootOutgoingCall("Greeter", "Hello12", reflect.TypeOf(impl.Hello12), ctx, arg, &ret)
	return
}

/****************************************************************
** Greeter Handler
********/

func RegisterGreeterHandler(peer wormhole.LocalPeer, constructor func(caller wormhole.RemotePeer) Greeter) {
	peer.(wormhole.LocalPeerGenerated).RegisterInterface("Greeter", func(caller wormhole.RemotePeer) {
		ifc := constructor(caller)
		val := reflect.ValueOf(ifc)

		caller.(wormhole.RemotePeerGenerated).RegisterRootRef("Greeter", "Hello", val.MethodByName("Hello"))
		caller.(wormhole.RemotePeerGenerated).RegisterRootRef("Greeter", "Hello12", val.MethodByName("Hello12"))
	})
}

/****************************************************************
** Greeter2 Client
********/

var _ Greeter2 = (*wormholeGreeter2ClientImpl)(nil)

type wormholeGreeter2ClientImpl struct {
	peer wormhole.RemotePeer
}

func AcquireGreeter2(peer wormhole.RemotePeer) Greeter2 {
	return &wormholeGreeter2ClientImpl{peer: peer}
}

func (impl *wormholeGreeter2ClientImpl) Hello21(ctx context.Context, arg GreeterHelloReq) (ret GreeterHelloResp, err error) {
	//mt, _ := reflect.TypeOf(impl).MethodByName("Hello21")
	err = impl.peer.(wormhole.RemotePeerGenerated).MakeRootOutgoingCall("Greeter2", "Hello21", reflect.TypeOf(impl.Hello21), ctx, arg, &ret)
	return
}

func (impl *wormholeGreeter2ClientImpl) Hello22(ctx context.Context, arg GreeterHelloReq) (ret GreeterHelloResp, err error) {
	//mt, _ := reflect.TypeOf(impl).MethodByName("Hello22")
	err = impl.peer.(wormhole.RemotePeerGenerated).MakeRootOutgoingCall("Greeter2", "Hello22", reflect.TypeOf(impl.Hello22), ctx, arg, &ret)
	return
}

/****************************************************************
** Greeter2 Handler
********/

func RegisterGreeter2Handler(peer wormhole.LocalPeer, constructor func(caller wormhole.RemotePeer) Greeter2) {
	peer.(wormhole.LocalPeerGenerated).RegisterInterface("Greeter2", func(caller wormhole.RemotePeer) {
		ifc := constructor(caller)
		val := reflect.ValueOf(ifc)

		caller.(wormhole.RemotePeerGenerated).RegisterRootRef("Greeter2", "Hello21", val.MethodByName("Hello21"))
		caller.(wormhole.RemotePeerGenerated).RegisterRootRef("Greeter2", "Hello22", val.MethodByName("Hello22"))
	})
}
