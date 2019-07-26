package remote_peer

import (
	"reflect"
	"sync"
)

type refs struct {
	rp   *remotePeer
	refs map[string]ref
	lock sync.RWMutex
}

func (rpm *refs) put(name string, rv reflect.Value, root bool) {
	rpm.lock.Lock()
	defer rpm.lock.Unlock()

	rpm.refs[name] = ref{
		name: name,
		val:  rv,
		//root: root, // FIXME
		rp:   rpm.rp,
	}
}

func (rpm *refs) del(name string) {
	rpm.lock.Lock()
	defer rpm.lock.Unlock()

	delete(rpm.refs, name)
}

func (rpm *refs) get(name string) (ref, bool) {
	rpm.lock.RLock()
	defer rpm.lock.RUnlock()

	ref, ok := rpm.refs[name]
	return ref, ok
}
