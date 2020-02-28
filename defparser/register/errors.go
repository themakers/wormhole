package register

import "fmt"

type ErrUnableRegistrate struct {
	DefinitionName string
	PackagePath    string
}

func (e ErrUnableRegistrate) Error() string {
	return fmt.Sprintf(
		"unable to registrate definition \"%s\" in package \"%s\"",
		e.DefinitionName,
		e.PackagePath,
	)
}
