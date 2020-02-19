package defparser

import (
	"errors"
	"fmt"
)

var (
	ErrEnvVarNotSetGOROOT = errors.New(
		"GOROOT env variable isn't set",
	)

	ErrEnvVarNotSetGOPATH = errors.New(
		"GOPATH env variable isn't set",
	)
	ErrNotAbsoluteFilePath = errors.New(
		"Specified file path isn't absolute",
	)
)

type (
	PackagingError error
)

type (
	Loop string
)

func (l Loop) Error() string {
	return fmt.Sprintf(
		"Package in a loop: %s",
		l,
	)
}
