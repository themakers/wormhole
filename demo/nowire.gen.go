package main

import (
	"reflect"

	"github.com/themakers/wormhole/wormhole"
)

/****************************************************************
** Greeter Client
********/

var _ Greeter = (*impl_client_Greeter)(nil)

type impl_client_Greeter struct {
	peer wormhole.RemotePeer
}

func NewGreeterClient(peer wormhole.RemotePeer) Greeter {
	return &impl_client_Greeter{peer: peer}
}

func (impl *impl_client_Greeter) Hello(name string, reply func(data []Model) string) (r0 string) {
	mtype, _ := reflect.TypeOf(impl).MethodByName("Hello")
	impl.peer.(wormhole.RemotePeerGenerated).MakeOutgoingCall("Greeter", "Hello", mtype.Type, []interface{}{name, reply}, []interface{}{&r0})
	return
}

/****************************************************************
** Greeter Server
********/

func RegisterGreeterServer(peer wormhole.LocalPeer, constructor func(caller wormhole.RemotePeer) Greeter) {
	peer.(wormhole.LocalPeerGenerated).RegisterInterface("Greeter", func(caller wormhole.RemotePeer) {
		ifc := constructor(caller)
		val := reflect.ValueOf(ifc)

		caller.(wormhole.RemotePeerGenerated).RegisterMethod("Greeter", "Hello", val.MethodByName("Hello"))
	})
}
