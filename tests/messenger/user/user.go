package user

import "time"

type User interface {
	SetPublicity(bool, time.Time) error

	// GetInfo() *struct {
	// 	FirstName string
	// 	LastName  string
	// }
}
