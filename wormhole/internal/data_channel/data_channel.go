package data_channel

import (
	"context"

	"github.com/themakers/wormhole/wormhole/wire_io"
)

type DataChannel interface {
	Context() context.Context
	Addr() string
	Close() error

	MessageWriter(int, func(w wire_io.ValueWriter) error) error
	MessageReader() (int, wire_io.ValueReader, func(), error)
}
