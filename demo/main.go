package main

//go:generate sh -c "go install github.com/themakers/nowire && nowire go"

import (
	"context"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/themakers/nowire/nowire"
)

func main() {
	log := nowire.NewZapLogger(true).Named("peer-1")

	lp1 := nowire.NewLocalPeer(log, nil)

	RegisterGreeterServer(lp1, func(rp nowire.RemotePeer) Greeter {
		return &greeter{
			log:  log,
			peer: NewGreeterClient(rp),
		}
	})

	go (func() {
		time.Sleep(100 * time.Millisecond)
		log := nowire.NewZapLogger(true).Named("peer-2")

		lp2 := nowire.NewLocalPeer(log, nowire.NewPeerCallbacks(
			func(rp nowire.RemotePeer) {
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

		if err := nowire.WebSocketConnect(context.TODO(), lp2, "ws://localhost:7532"); err != nil {
			log.DPanic("Error initiating connection", zap.Error(err))
		}
	})()

	http.ListenAndServe(":7532", nowire.WebSocketAcceptor(lp1))
}

type greeter struct {
	log  *zap.Logger
	peer Greeter
}

func (gr *greeter) Hello(name string, reply func(data []Model) string) string {
	gr.log.Info(name)
	return reply([]Model{{ID: "12121", Time: time.Now()}})
}
