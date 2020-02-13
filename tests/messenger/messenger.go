package messenger

import (
	"context"

	"github.com/themakers/wormhole/tests/messenger/message"
	userPackage "github.com/themakers/wormhole/tests/messenger/user"
)

//go:generate sh -c "go install github.com/themakers/wormhole/cmd/wormhole && wormhole go"

type Messenger interface {
	SignUp(context.Context, MessengerSignUpReq) error
	ListUsers(context.Context) ([]userPackage.User, error)
	Text(context.Context, userPackage.User, message.Data) error
}

type MessengerSignUpReq struct {
	FirstName             string
	LastName              string
	MessageStreamCallback func(context.Context)
}

type MessengerSignUpResp struct {
}

// type Greeter interface {
// 	Hello(ctx context.Context, q GreeterHelloReq) (GreeterHelloResp, error)
// }

// type GreeterHelloReq struct {
// 	Message     string
// 	CallableRef func(ctx context.Context, data message.Message) (message.Message, error)
// }

// type GreeterHelloResp struct {
// 	Message message.Message
// 	Reply   func(ctx context.Context, msg message.Message) (message.Message, error) // TODO
// }
