package api

//go:generate sh -c "go install github.com/themakers/wormhole/cmd/wormhole && wormhole go"

import (
	"context"
	"time"
)

type Model struct {
	ID   string
	Time time.Time
}

type Greeter interface {
	Hello(ctx context.Context, q GreeterHelloReq) (GreeterHelloResp, error)
}

type GreeterHelloReq struct {
	Message     string
	CallableRef func(ctx context.Context, data string) (string, error)
}

type GreeterHelloResp struct {
	Message string
	// Reply func(ctx context.Context, data string) (string, error) // TODO
}
