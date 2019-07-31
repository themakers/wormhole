package main

//go:generate sh -c "cd ../api && go install github.com/themakers/wormhole/cmd/wormhole && wormhole go"

import (
	"context"
	"github.com/themakers/wormhole/demo/api"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole/wormhole_websocket"
	"log"
	"net/http"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	lp := wormhole.NewLocalPeer(nil)
	defer lp.Close()

	api.RegisterGreeterHandler(lp, func(rp wormhole.RemotePeer) api.Greeter {
		return &greeter{
			peer: api.AcquireGreeter(rp),
		}
	})

	s := http.Server{
		Addr:    "localhost:7532",
		Handler: wormhole_websocket.Acceptor(lp),
	}

	go (func() {
		<-ctx.Done()
		s.Shutdown(ctx)
	})()

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

type greeter struct {
	peer api.Greeter
}

func (gr *greeter) Hello(ctx context.Context, q api.GreeterHelloReq) (api.GreeterHelloResp, error) {
	log.Println("Hello()", "name", q.Message)

	n, err := q.CallableRef(ctx, "Hello, "+q.Message+"!")

	log.Println("CallableRef()", "n", n)

	return api.GreeterHelloResp{
		Message: "++++ " + n,
	}, err
}
