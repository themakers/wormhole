package wormparse

import (
	"errors"
	"fmt"
)

var (
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
