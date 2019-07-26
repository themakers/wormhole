package main

//go:generate sh -c "go install github.com/themakers/wormhole/cmd/wormhole && wormhole go"

import (
	"context"
	"github.com/themakers/wormhole/wormhole/wormhole_websocket"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/themakers/wormhole/wormhole"
)

func main() {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	go (func() {
		log := log.Named("peer-2")
		time.Sleep(100 * time.Millisecond)

		lp2 := wormhole.NewLocalPeer(log, wormhole.NewPeerCallbacks(
			func(rp wormhole.RemotePeer) {
				log.Info("Peer connected!")
				res, err := AcquireGreeter(rp).Hello(context.TODO(), GreeterHelloReq{
					Name: "Jun",
					NameChanged: func(ctx context.Context, data string) (string, error) {
						log.Info("AJAJA2", zap.String("name", data))
						return "+" + data + "+", nil
					},
				})
				log.Info("DONE", zap.Any("res", res), zap.Error(err))
				// os.Exit(0)
			},
			func(id string) {},
		))

		if err := wormhole_websocket.Connect(context.TODO(), lp2, "ws://localhost:7532"); err != nil {
			log.Panic("Error initiating connection", zap.Error(err))
		}
	})()

	{
		log := log.Named("peer-1")
		lp1 := wormhole.NewLocalPeer(log, nil)

		RegisterGreeterHandler(lp1, func(rp wormhole.RemotePeer) Greeter {
			return &greeter{
				log:  log,
				peer: AcquireGreeter(rp),
			}
		})

		if err := http.ListenAndServe("localhost:7532", wormhole_websocket.Acceptor(lp1)); err != nil {
			log.Panic("Error listening", zap.Error(err))
		}
	}
}

type greeter struct {
	log  *zap.Logger
	peer Greeter
}

func (gr *greeter) Hello(ctx context.Context, q GreeterHelloReq) (GreeterHelloResp, error) {
	gr.log.Info("AJAJA", zap.String("name", q.Name))
	n, err := q.NameChanged(ctx, "Hello, " + q.Name + "!")

	return GreeterHelloResp{
		Name: "!!!!!" + n,
	}, err
}

func (gr *greeter) Hello12(ctx context.Context, q GreeterHelloReq) (GreeterHelloResp, error) {
	gr.log.Info(q.Name)
	q.NameChanged(ctx, time.Now().String())
	return GreeterHelloResp{}, nil
}
