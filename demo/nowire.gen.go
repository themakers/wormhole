package main

import (
	"reflect"

	"github.com/themakers/nowire/nowire"
)

/****************************************************************
** Greeter Client
********/

var _ Greeter = (*impl_client_Greeter)(nil)

type impl_client_Greeter struct {
	peer nowire.RemotePeer
}

func NewGreeterClient(peer nowire.RemotePeer) Greeter {
	return &impl_client_Greeter{peer: peer}
}

func (impl *impl_client_Greeter) Hello(name string, reply func(data []Model) string) (r0 string) {
	mtype, _ := reflect.TypeOf(impl).MethodByName("Hello")
	impl.peer.(nowire.RemotePeerGenerated).MakeOutgoingCall("Greeter", "Hello", mtype.Type, []interface{}{name, reply}, []interface{}{&r0})
	return
}

/****************************************************************
** Greeter Server
********/

func RegisterGreeterServer(peer nowire.LocalPeer, constructor func(caller nowire.RemotePeer) Greeter) {
	peer.(nowire.LocalPeerGenerated).RegisterInterface("Greeter", func(caller nowire.RemotePeer) {
		ifc := constructor(caller)
		val := reflect.ValueOf(ifc)

		caller.(nowire.RemotePeerGenerated).RegisterMethod("Greeter", "Hello", val.MethodByName("Hello"))
	})
}
