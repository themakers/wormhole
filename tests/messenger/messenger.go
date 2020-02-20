package messenger

import (
	"github.com/themakers/wormhole/tests/messenger/message"
	"github.com/themakers/wormhole/tests/messenger/user"
)

// type T0 interface {
// 	A()
// 	B() struct {
// 		a int
// 		b uint
// 	}
// }

type T0 struct {
	a int
	b uint
}

type T1 struct {
	d int
	s struct {
		u user.User
		b int
	}
	i interface {
		A(bool, T0) int
		B()
	}
	dup struct {
		a int
		b uint
	}
}

type I interface {
	A(T0) T1
	B(T1) T0
	С(*chan struct{ T0 }) [5]struct {
		user.User
		OLOLO user.TROLOLO
	}
}

func F(m message.Message) *struct {
	a string
	I
} {
	return nil
}

//go:generate sh -c "go install github.com/themakers/wormhole/cmd/wormhole && wormhole go"

// var _ interface{} = wormparse.Parse

// type Messenger interface {
// 	SignUp(context.Context, MessengerSignUpReq) error
// 	ListUsers(context.Context) ([]userPackage.User, error)
// 	Text(context.Context, userPackage.User, message.Data) error
// }

// type MessengerSignUpReq struct {
// 	FirstName             string
// 	LastName              string
// 	MessageStreamCallback func(context.Context)
// 	c                     chan MessengerSignUpReq
// }

// type MessengerSignUpResp struct {
// }

// func TestFunc(a int, b Messenger) interface {
// 	A() map[string]string
// 	B()
// }

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
