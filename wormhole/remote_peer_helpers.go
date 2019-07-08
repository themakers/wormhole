package wormhole

import (
	"reflect"
	"sync"
)

type rpMethods struct {
	methods map[string]reflect.Value
	lock    sync.RWMutex
}

func (rpm *rpMethods) put(name string, mval reflect.Value) {
	rpm.lock.Lock()
	defer rpm.lock.Unlock()

	rpm.methods[name] = mval
}

func (rpm *rpMethods) del(name string) {
	rpm.lock.Lock()
	defer rpm.lock.Unlock()

	delete(rpm.methods, name)
}

func (rpm *rpMethods) get(name string) (reflect.Value, bool) {
	rpm.lock.RLock()
	defer rpm.lock.RUnlock()

	mval, ok := rpm.methods[name]
	return mval, ok
}

type rpCalls struct {
	calls map[string]*Call
	lock  sync.RWMutex
}

func (c *rpCalls) put(name string, call *Call) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.calls[name] = call
}

func (c *rpCalls) del(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.calls, name)
}

func (c *rpCalls) get(name string) (*Call, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	call, ok := c.calls[name]
	return call, ok
}

// case reflect.String:
// case reflect.Array, reflect.Slice:
// case reflect.Bool:
// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
// 	reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// case reflect.Float32, reflect.Float64:

// case []byte:
// case bool:
// case int:
// case float64:
