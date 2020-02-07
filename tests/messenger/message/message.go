package message

import (
	"time"

	"github.com/themakers/wormhole/tests/messenger/user"
)

type Message struct {
	ID   string
	Data Data
	Time time.Time

	Meta struct {
		Deleted bool
	}

	From *user.User
}

type Data struct {
	Title string
	Text  string
}
