package main

import (
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
				NewGreeterClient(rp).Hello("a", func(data []Model) {
					log.Info("2", zap.Any("i", data))
					time.Sleep(100 * time.Millisecond)
					os.Exit(0)
				})
			},
			func(id string) {},
		))

		if err := nowire.WebSocketConnect(lp2, "ws://localhost:7532"); err != nil {
			log.DPanic("Error initiating connection", zap.Error(err))
		}
	})()

	http.ListenAndServe(":7532", nowire.WebSocketAcceptor(lp1))
}

type greeter struct {
	log  *zap.Logger
	peer Greeter
}

func (gr *greeter) Hello(name string, reply func(data []Model)) {
	gr.log.Info("1", zap.String("i", name))
	reply([]Model{{ID: "12121", Time: time.Now()}})
}
