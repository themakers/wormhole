package main

//go:generate sh -c "cd ../../tests/api && go install github.com/themakers/wormhole/cmd/wormhole && wormhole go"

import (
	"context"
	"github.com/pkg/profile"
	"github.com/themakers/wormhole/tests/api"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole_websocket"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	pp := profile.Start(profile.MemProfile, profile.ProfilePath("./client.pprof"), profile.NoShutdownHook)
	defer pp.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	lp := wormhole.NewLocalPeer(wormhole.NewPeerCallbacks(func(peer wormhole.RemotePeer) {

	}, func(peer wormhole.RemotePeer) {

	}))
	defer lp.Close()

	go wormhole_websocket.StayConnected(ctx, lp, wormhole.NewPeerCallbacks(func(peer wormhole.RemotePeer) {

		go func() {
			greeter := api.AcquireGreeter(peer)

			//ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)

			if resp, err := greeter.Hello(ctx, api.GreeterHelloReq{
				Message: "Jun",
				CallableRef: func(ctx context.Context, data string) (res string, err error) {
					log.Println("CallableRef()", data)
					time.Sleep(1 * time.Second)
					return "¡¡¡¡ " + data + " !!!!", nil
				},
			}); err != nil {
				panic(err)
			} else {
				log.Println("resp:", resp)
			}

			//os.Exit(0)
		}()

	}, func(peer wormhole.RemotePeer) {

	}), "ws://localhost:7532")

	go func() {
		defer cancel()
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		<-c
	}()

	<-ctx.Done()
}
