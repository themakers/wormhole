package user

import "context"

type User struct {
	I string
	c context.Context
}

// type (
// 	TestType struct {
// 		Data    int
// 		Scanner scanner.Position `json:"pos"`
// 		Subdoc  *struct {
// 			a uint
// 			b string
// 			e error
// 			d time.Duration
// 			r rune
// 		}

// 		INTER interface {
// 			OLOLO(CAP) (struct {
// 				AAA int
// 				BBB uint
// 			}, error)
// 		}
// 	}

// 	CAP struct {
// 		arr     [2]int
// 		sl      []byte
// 		closure func(i int) func(TestType) error
// 		m       map[string]interface{}
// 	}

// 	User interface {
// 		SetPublicity(bool, time.Time, scanner.Scanner) error
// 	}
// )

// func (t *TestType) JustAnotherPerlHacker(a int, b uint) error {
// 	return nil
