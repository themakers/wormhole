package base

import "context"

type DataChannel interface {
	Context() context.Context
	Addr() string
	ReadMessage() (interface{}, error)
	WriteMessage(interface{}) error
	Close() error
}
