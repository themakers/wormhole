package user

import "github.com/themakers/wormhole/tests/messenger/message"

type User interface {
	SetPublicity(bool) error

	GetInfo() *struct {
		FirstName string
		LastName  string
	}
}
