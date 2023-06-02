package keymutex

import "sync"

// Map is a map of mutual exclusion locks. The zero value is safe to use.
type Map[K comparable] struct {
	master sync.Mutex
	m      map[K]*sync.Mutex
}

func (m *Map[K]) init() {
	if m.m == nil {
		m.m = make(map[K]*sync.Mutex)
	}
}

func (m *Map[K]) Lock(key K) {
	m.master.Lock()
	m.init()
	lock, ok := m.m[key]
	if !ok {
		lock = &sync.Mutex{}
		m.m[key] = lock
	}
	m.master.Unlock()

	lock.Lock()
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
}

// Go is a convenience method that guards the given function with a mutex.
func (m *Map[K]) Go(key K, f func()) {
	m.Lock(key)
	go func() {
		defer m.Unlock(key)
		f()
	}()
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
