package wormparse

import (
	"errors"
	"fmt"
	"strings"
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
	Loops []Loop

	Loop struct {
		Nodes []PackageInfo
		index int
	}
)

func (l Loops) Error() string {
	res := make([]string, len(l))
	for i, loop := range l {
		res[i] = loop.Error()
	}
	return fmt.Sprintf(
		"Multiple loops:\n%s",
		strings.Join(res, "\n\n"),
	)
}

func (l Loop) Error() string {
	res := make([]string, len(l.Nodes))
	for i, node := range l.Nodes {
		res[i] = fmt.Sprintf(
			"%s :: %s :: %s",
			node.PkgName,
			node.PkgPath,
			node.PkgFullPath,
		)
	}
	return fmt.Sprintf(
		"An import loop:\n%s",
		strings.Join(res, "\n"),
	)
}
