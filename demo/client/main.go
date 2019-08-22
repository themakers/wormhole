package main

//go:generate sh -c "cd ../api && go install github.com/themakers/wormhole/cmd/wormhole && wormhole go"

import (
	"context"
	"github.com/themakers/wormhole/tests/api"
	"github.com/themakers/wormhole/wormhole/wormhole_websocket"
	"log"
	"time"

	"github.com/themakers/wormhole/wormhole"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	lp := wormhole.NewLocalPeer(nil)
	defer lp.Close()

	go wormhole_websocket.StayConnected(ctx, lp, "ws://localhost:7532")

	greeter := api.AcquireKeepAliveGreeter(lp, "ws://localhost:7532", 2*time.Second)

	ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)

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
}
