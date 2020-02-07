package parsex

import "errors"

var (
	ErrNotAbsoluteFilePath = errors.New(
		"Specified file path isn't absolute",
	)
)
