package keymutex

import (
	"context"
	"sync"
	"sync/atomic"
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

// lockLoop assumes that the master lock is already held.
func (m *Map[K]) lockLoop(key K, cancel *int64) bool {
	for {
		keyLock := m.initLock(key)
		if keyLock.TryLock() {
			return true
		} else if cancel != nil && atomic.LoadInt64(cancel) == 1 {
			return false
		}
		m.cond.Wait()
	}
}

// LockCtx locks the mutex for the given key, returning true if the lock
// was acquired before the context is canceled.
func (m *Map[K]) LockCtx(ctx context.Context, key K) bool {
	m.master.Lock()
	defer m.master.Unlock()

	// We must create a child context here in-case the parent never closes,
	// which would cause a leak.
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	var c int64
	go func() {
		<-ctx.Done()
		atomic.StoreInt64(&c, 1)
		m.cond.Broadcast()
	}()

	return m.lockLoop(key, &c)
}

// TryLock attempts to lock the mutex for the given key, returning true
// if the mutex was successfully locked.
//
// It returns immediately.
func (m *Map[K]) TryLock(key K) bool {
	m.master.Lock()
	defer m.master.Unlock()

	keyLock := m.initLock(key)
	return keyLock.TryLock()
}

func (m *Map[K]) Lock(key K) {
	m.master.Lock()
	defer m.master.Unlock()
	m.lockLoop(key, nil)
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
