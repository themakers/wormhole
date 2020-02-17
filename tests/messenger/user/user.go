package user

import (
	"text/scanner"
	"time"
)

type TestType struct {
	Data    int
	Scanner scanner.Position `json:"pos"`
	Subdoc  struct {
		a uint
		b string
		e error
		d time.Duration
		r rune
	}
}

type User interface {
	SetPublicity(bool, time.Time, scanner.Scanner) error

	// GetInfo() *struct {
	// 	FirstName string
	// 	LastName  string
	// }
}
