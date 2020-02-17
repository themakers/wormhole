package user

import (
	"text/scanner"
	"time"
)

type TestType struct {
	Data int
}

type User interface {
	SetPublicity(bool, time.Time, scanner.Scanner) error

	// GetInfo() *struct {
	// 	FirstName string
	// 	LastName  string
	// }
}
