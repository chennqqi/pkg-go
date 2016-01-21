package pkgmemoize // import "go.pedge.io/pkg/memoize"

import "sync"

// EMemoFunc is a memoizeable function with errors.
type EMemoFunc func(int) (interface{}, error)

// MemoFunc is a memoizeable function without errors.
type MemoFunc func(int) interface{}

// NewEMemoFunc returns a new EMemoFunc.
//
// values and errors returned from f cannot be mutated, they will be persisted.
func NewEMemoFunc(f func(int, func(int) (interface{}, error)) (interface{}, error)) EMemoFunc {
	return newEMemoFunc(f).Do
}

// NewMemoFunc returns a new MemoFunc.
//
// values returned from f cannot be mutated, they will be persisted.
func NewMemoFunc(f func(int, func(int) interface{}) interface{}) MemoFunc {
	return newMemoFunc(f).Do
}

type valueError struct {
	value interface{}
	err   error
}

type eMemoFunc struct {
	f     func(int, func(int) (interface{}, error)) (interface{}, error)
	cache map[int]*valueError
	lock  *sync.RWMutex
}

func newEMemoFunc(f func(int, func(int) (interface{}, error)) (interface{}, error)) *eMemoFunc {
	return &eMemoFunc{f, make(map[int]*valueError), &sync.RWMutex{}}
}

func (e *eMemoFunc) Do(i int) (interface{}, error) {
	e.lock.RLock()
	cached, ok := e.cache[i]
	e.lock.RUnlock()
	if ok {
		return cached.value, cached.err
	}
	value, err := e.f(i, e.Do)
	cached = &valueError{value: value, err: err}
	e.lock.Lock()
	e.cache[i] = cached
	e.lock.Unlock()
	return value, err
}

type memoFunc struct {
	f     func(int, func(int) interface{}) interface{}
	cache map[int]interface{}
	lock  *sync.RWMutex
}

func newMemoFunc(f func(int, func(int) interface{}) interface{}) *memoFunc {
	return &memoFunc{f, make(map[int]interface{}), &sync.RWMutex{}}
}

func (m *memoFunc) Do(i int) interface{} {
	m.lock.RLock()
	value, ok := m.cache[i]
	m.lock.RUnlock()
	if ok {
		return value
	}
	value = m.f(i, m.Do)
	m.lock.Lock()
	m.cache[i] = value
	m.lock.Unlock()
	return value
}
