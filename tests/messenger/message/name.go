package message

import (
	"errors"
	"strings"

	"github.com/themakers/wormhole/parsex"
)

var _ = parsex.PWD()

type Name string

func NewName(name string) (Name, error) {
	if len(name) < 2 {
		return "", errors.New("name is too short")
	}
	return Name(name), nil
}

func (n Name) String() string {
	return strings.ToUpper(string(n[0])) + string(n[1:])
}
