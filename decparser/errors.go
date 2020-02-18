package decparser

import "errors"

var (
	ErrEnvVarNotSetGOROOT = errors.New(
		"GOROOT env variable isn't set",
	)

	ErrEnvVarNotSetGOPATH = errors.New(
		"GOPATH env variable isn't set",
	)
)
