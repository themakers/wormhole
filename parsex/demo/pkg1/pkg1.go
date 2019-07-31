package pkg1

import (
	"context"
	"time"
)


type Model struct {
	ID   string
	Time time.Time
}

type Greeter1 interface {
	Hello1_1(ctx context.Context, q GreeterHelloReq) (GreeterHelloResp, error)
	Hello1_2(ctx context.Context, q GreeterHelloReq) (GreeterHelloResp, error)
}


type Greeter2 interface {
	Hello2_1(ctx context.Context, q GreeterHelloReq) (GreeterHelloResp, error)
	Hello2_2(ctx context.Context, q GreeterHelloReq) (GreeterHelloResp, error)
}

type GreeterHelloReq struct {
	Name        string
	NameChanged func(ctx context.Context, data []Model) string
}

type GreeterHelloResp struct {
	Name  string
	Reply func(ctx context.Context, data []Model) string
}

/*
const call = client.User.Events({
	Message: 'daniil',
	CallableRef: (data) => {

	},
	BalanceChanged: (data) => {

	}
})

call.closed((err) => {})

call.response((resp) => {
	resp.Reply()
})

call.cancel()
*/
