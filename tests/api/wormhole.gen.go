package api

import (
	"context"
	"fmt"

	__wormhole "github.com/themakers/wormhole/wormhole"
	wormhole "github.com/themakers/wormhole/wormhole"
	__reflect_hack "github.com/themakers/wormhole/wormhole/reflect_hack"
	__wire_io "github.com/themakers/wormhole/wormhole/wire_io"
)

/****************************************************************
** Greeter Client
********/

var _ Greeter = (*wormholeGreeterClientImpl)(nil)

type wormholeGreeterClientImpl struct {
	peer __wormhole.RemotePeer
}

func AcquireGreeter(peer __wormhole.RemotePeer) Greeter {
	return &wormholeGreeterClientImpl{peer: peer}
}

func (__impl *wormholeGreeterClientImpl) Hello(ctx context.Context, q GreeterHelloReq) (__2 GreeterHelloResp, __3 error) {
	__peer := __impl.peer.(__wormhole.RemotePeerGenerated)

	__doneCtx, __done := context.WithCancel(ctx)
	defer __done()

	__peer.Call("Greeter.Hello", ctx, 1, func(__rr __wormhole.RegisterUnnamedRefFunc, __w __wire_io.ValueWriter) {

		__reflect_hack.WriteAny(__peer, __rr, __w, q)

	}, func(ctx context.Context, __ar __wire_io.ArrayReader) {
		defer __done()
		__sz, __r, __err := __ar()
		if __err != nil {
			panic(__err)
		}
		if __sz != 2 {
			panic(fmt.Sprintf("return values count mismatch: %d != %d", 2, __sz))
		}

		__reflect_hack.ReadAny(__peer, __r, &__2)
		__reflect_hack.ReadAny(__peer, __r, &__3)

	})

	<-__doneCtx.Done()

	return
}

/****************************************************************
** Greeter Handler
********/

func RegisterGreeterHandler(localPeer wormhole.LocalPeer, constructor func(wormhole.RemotePeer) Greeter) {
	localPeer.(__wormhole.LocalPeerGenerated).RegisterInterface("Greeter", func(peer __wormhole.RemotePeer) {
		__ifc := constructor(peer)
		__peer := peer.(__wormhole.RemotePeerGenerated)

		__peer.RegisterServiceRef("Greeter.Hello", func(ctx context.Context, __ar __wire_io.ArrayReader, __wf func(int, func(__wormhole.RegisterUnnamedRefFunc, __wire_io.ValueWriter))) {
			var (
				q GreeterHelloReq
			)

			__sz, __r, __err := __ar()
			if __err != nil {
				panic(__err)
			}
			if __sz != 1 {
				panic(fmt.Sprintf("arguments count mismatch: %d != %d", 1, __sz))
			}

			__reflect_hack.ReadAny(__peer, __r, &q)

			__2, __3 := __ifc.Hello(ctx, q)

			__wf(2, func(__rr __wormhole.RegisterUnnamedRefFunc, __w __wire_io.ValueWriter) {
				__reflect_hack.WriteAny(__peer, __rr, __w, __2)
				__reflect_hack.WriteAny(__peer, __rr, __w, __3)
			})
		})
	})
}
