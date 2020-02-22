package types

import "fmt"

type ErrAmbigiousSelector struct {
	Sel string
}

func (as ErrAmbigiousSelector) Error() string {
	return fmt.Sprintf(
		"ambigious selector %s",
		as.Sel,
	)
}

type ErrSelectorUndefined struct {
	Sel string
}

func (su ErrSelectorUndefined) Error() string {
	return fmt.Sprintf(
		"selector %s is undefined",
		su.Sel,
	)
}
