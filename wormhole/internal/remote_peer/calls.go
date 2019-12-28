package remote_peer

import (
	"context"
	"github.com/themakers/wormhole/wormhole/wire_io"
	"sync"
)

type ResultFunc func(context.Context, wire_io.ArrayReader)

type outgoingCalls struct {
	calls map[string]ResultFunc
	lock  sync.RWMutex
}

func (c *outgoingCalls) put(ctx context.Context, id string, res ResultFunc) (<-chan struct{}, func()) {
	ctx, cancel := context.WithCancel(ctx)

	c.lock.Lock()
	defer c.lock.Unlock()

	c.calls[id] = func(ctx context.Context, ar wire_io.ArrayReader) {
		res(ctx, ar)
		cancel()
	}

	return ctx.Done(), func() {
		c.lock.Lock()
		defer c.lock.Unlock()

		delete(c.calls, id)

		cancel()
	}
}

func (c *outgoingCalls) get(id string) ResultFunc {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.calls[id]
}
