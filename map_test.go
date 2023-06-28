package keymutex

import (
	"sync"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	var m Map[string]

	// Zero value should be usable.
	m.Lock("foo")

	m.Go("blah", func() {
		m.Lock("blah")
		panic("should not be reached")
	})

	m.Unlock("foo")

	time.Sleep(time.Millisecond * 100)

	m.Lock("bar")
	m.Unlock("bar")

	m.Do("bar", func() {})

	m.Lock("bar")
	m.Unlock("bar")

	if l := m.Len(); l != 1 {
		t.Fatal("expected 1 mutex, got", l)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		m.Go("race", func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * 100)
		})
	}
	wg.Wait()
}

func BenchmarkMap_LockUnlock(b *testing.B) {
	var m Map[string]

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m.Lock("foo")
		m.Unlock("foo")
	}
}
