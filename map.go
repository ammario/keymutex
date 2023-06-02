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
