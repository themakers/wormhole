package message

import (
	"errors"
	"strings"

)

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
