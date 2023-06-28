package keymutex

import (
	"sync"
)

// Map is a map of mutual exclusion locks. The zero value is safe to use.
type Map[K comparable] struct {
	master sync.Mutex
	cond   *sync.Cond
	m      map[K]*sync.Mutex
}

func (m *Map[K]) initLock(key K) *sync.Mutex {
	if m.cond == nil {
		m.cond = sync.NewCond(&m.master)
	}
	if m.m == nil {
		m.m = make(map[K]*sync.Mutex)
	}
	lock, ok := m.m[key]
	if !ok {
		lock = &sync.Mutex{}
		m.m[key] = lock
	}
	return lock
}

func (m *Map[K]) Lock(key K) {
	m.master.Lock()
	defer m.master.Unlock()
	for {
		keyLock := m.initLock(key)
		if keyLock.TryLock() {
			return
		}
		m.cond.Wait()
	}
}

func (m *Map[K]) Unlock(key K) {
	m.master.Lock()
	defer m.master.Unlock()

	if m.m == nil {
		panic("unlock of unlocked mutex")
	}

	lock, ok := m.m[key]
	if !ok {
		panic("unlock of unlocked mutex")
	}

	lock.Unlock()

	// While this may create redundant allocation, it prevents
	// runaway memory growth.
	delete(m.m, key)

	m.cond.Broadcast()
}

// Go is a convenience method that guards the given function with a mutex.
// It locks synchronously, but runs fn in a separate goroutine.
func (m *Map[K]) Go(key K, fn func()) {
	m.Lock(key)
	go func() {
		defer m.Unlock(key)
		fn()
	}()
}

// Do is a convenience method that guards the given function with a mutex.
// It is similar to Go, but runs fn synchronously.
func (m *Map[K]) Do(key K, fn func()) {
	m.Lock(key)
	defer m.Unlock(key)
	fn()
}

// Len returns the number of mutexes currently in the map. The number
// of pending goroutines is at least this large, but may be larger
// under contention.
//
// Len is provided for debugging purposes. For example, making sure
// that there are no outstanding goroutines after a test finishes or
// that the number of mutexes is monotonically decreasing in the case
// of a wind-down.
func (m *Map[K]) Len() int {
	m.master.Lock()
	defer m.master.Unlock()

	if m.m == nil {
		return 0
	}

	return len(m.m)
}
