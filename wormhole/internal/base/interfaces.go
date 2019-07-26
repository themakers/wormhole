package base

import "context"

type DataChannel interface {
	Context() context.Context
	ReadMessage() (interface{}, error)
	WriteMessage(interface{}) error
	Close() error
}
