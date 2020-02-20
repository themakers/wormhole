package message

import (
	"time"

	"github.com/themakers/wormhole/tests/messenger/user"
)

type Data struct {
	Title string
	Text  string
}

type Message struct {
	ID   string
	Data Data
	Time time.Time

	Meta struct {
		Deleted bool
	}

	From *user.User
}
