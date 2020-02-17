package user

type User interface {
	SetPublicity(bool) error

	GetInfo() *struct {
		FirstName string
		LastName  string
	}
}
