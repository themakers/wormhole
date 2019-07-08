package main

//go:generate sh -c "go install github.com/themakers/wormhole && wormhole go"

import (
	"context"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/themakers/wormhole/wormhole"
)

func main() {
	log := wormhole.NewZapLogger(true).Named("peer-1")

	lp1 := wormhole.NewLocalPeer(log, nil)

	RegisterGreeterServer(lp1, func(rp wormhole.RemotePeer) Greeter {
		return &greeter{
			log:  log,
			peer: NewGreeterClient(rp),
		}
	})

	go (func() {
		time.Sleep(100 * time.Millisecond)
		log := wormhole.NewZapLogger(true).Named("peer-2")

		lp2 := wormhole.NewLocalPeer(log, wormhole.NewPeerCallbacks(
			func(rp wormhole.RemotePeer) {
				log.Info("Peer connected!")
				res := NewGreeterClient(rp).Hello("a", func(data []Model) string {
					log.Info("2", zap.Any("i", data))
					time.Sleep(100 * time.Millisecond)
					return "ajaja"
				})
				log.Info(res)
				os.Exit(0)
			},
			func(id string) {},
		))

		if err := wormhole.WebSocketConnect(context.TODO(), lp2, "ws://localhost:7532"); err != nil {
			log.DPanic("Error initiating connection", zap.Error(err))
		}
	})()

	http.ListenAndServe(":7532", wormhole.WebSocketAcceptor(lp1))
}

type greeter struct {
	log  *zap.Logger
	peer Greeter
}

func (gr *greeter) Hello(name string, reply func(data []Model) string) string {
	gr.log.Info(name)
	return reply([]Model{{ID: "12121", Time: time.Now()}})
}
