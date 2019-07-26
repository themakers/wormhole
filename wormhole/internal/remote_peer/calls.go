package remote_peer

import (
	"github.com/themakers/wormhole/wormhole/internal/proto"
	"sync"
)

type outgoingCall struct {
	id  string
	ref string

	res chan proto.Result
}

type outgoingCalls struct {
	calls map[string]*outgoingCall
	lock  sync.RWMutex
}

func (c *outgoingCalls) put(id string, ref string) *outgoingCall {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.calls[id] = &outgoingCall{
		id:  id,
		ref: ref,
		res: make(chan proto.Result, 1),
	}
	return c.calls[id]
}

func (c *outgoingCalls) del(id string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.calls, id)
}

func (c *outgoingCalls) get(id string) (*outgoingCall, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	call, ok := c.calls[id]
	return call, ok
}
