package register

import "fmt"

type ErrUnableRegistrate struct {
	DefName     string
	PackagePath string
}

func (e ErrUnableRegistrate) Error() string {
	return fmt.Sprintf(
		"unable to registrate definition \"%s\" in package \"%s\"",
		e.DefName,
		e.PackagePath,
	)
}
