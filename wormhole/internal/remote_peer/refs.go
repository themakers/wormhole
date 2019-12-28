package remote_peer

import (
	"context"
	"github.com/themakers/wormhole/wormhole/wire_io"
	"sync"
)

type RefFunc func(ctx context.Context, ar wire_io.ArrayReader, w func(int, func(RegisterUnnamedRefFunc, wire_io.ValueWriter)))

type refs struct {
	refs map[string]RefFunc
	lock sync.RWMutex
}

func (rpm *refs) put(name string, ref RefFunc) func() {
	rpm.lock.Lock()
	defer rpm.lock.Unlock()

	rpm.refs[name] = ref

	return func() {
		delete(rpm.refs, name)
	}
}

func (rpm *refs) remove(fns ...func()) {
	rpm.lock.Lock()
	defer rpm.lock.Unlock()

	for _, fn := range fns {
		fn()
	}
}

func (rpm *refs) get(name string) RefFunc {
	rpm.lock.RLock()
	defer rpm.lock.RUnlock()

	return rpm.refs[name]
}
