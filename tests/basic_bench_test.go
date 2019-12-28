package tests

import (
	"context"
	"fmt"
	"github.com/themakers/wormhole/tests/api"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole_websocket"
	"net"
	"net/http"
	"testing"
	"time"
)

func BenchmarkBasic(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := StartServer(ctx)

	client := CreateClient(ctx, port)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if resp, err := client.Hello(ctx, api.GreeterHelloReq{
			Message: "+",
			CallableRef: func(ctx context.Context, data string) (res string, err error) {
				return data + "+", nil
			},
		}); err != nil {
			panic(err)
		} else {
			if resp.Message == "" {

			}
		}
	}
}

func CreateClient(ctx context.Context, port int) api.Greeter {
	lp := wormhole.NewLocalPeer(nil)
	defer lp.Close()

	api.RegisterGreeterHandler(lp, func(rp wormhole.RemotePeer) api.Greeter {
		return &greeter{
			client: true,
			peer:   api.AcquireGreeter(rp),
		}
	})

	addr := fmt.Sprintf("ws://localhost:%d", port)

	go wormhole_websocket.StayConnected(ctx, lp, addr)

	return api.AcquireKeepAliveGreeter(lp, addr, 2*time.Second)
}

func StartServer(ctx context.Context) int {

	lp := wormhole.NewLocalPeer(nil)
	defer lp.Close()

	api.RegisterGreeterHandler(lp, func(rp wormhole.RemotePeer) api.Greeter {
		return &greeter{
			server: true,
			peer:   api.AcquireGreeter(rp),
		}
	})

	s := http.Server{
		Addr:    "localhost:0",
		Handler: wormhole_websocket.Acceptor(lp),
	}

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	go (func() {
		<-ctx.Done()
		//s.Shutdown(ctx)
	})()

	go (func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	})()

	return lis.Addr().(*net.TCPAddr).Port
}

type greeter struct {
	client bool
	server bool
	peer   api.Greeter
}

func (gr *greeter) Hello(ctx context.Context, q api.GreeterHelloReq) (api.GreeterHelloResp, error) {
	n, err := q.CallableRef(ctx, q.Message+"+")

	if gr.server {
		if resp, err := gr.peer.Hello(ctx, api.GreeterHelloReq{
			Message: "+",
			CallableRef: func(ctx context.Context, data string) (res string, err error) {
				return data + "+", nil
			},
		}); err != nil {
			panic(err)
		} else {
			if resp.Message == "" {

			}
		}
	}

	return api.GreeterHelloResp{
		Message: n + "+",
	}, err
}
