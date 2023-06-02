package keymutex

import (
	"testing"
	"time"
)

func TestMmap(t *testing.T) {
	var m Map[string]

	// Zero value should be usable.
	m.Lock("foo")
	go func() {
		m.Lock("foo")
		panic("should not be reached")
	}()

	time.Sleep(time.Millisecond * 100)

	m.Lock("bar")
	m.Unlock("bar")

	m.Lock("bar")
	m.Unlock("bar")

	if len(m.m) != 1 {
		t.Fatal("expected 1 mutex")
	}
}

func BenchmarkMap_LockUnlock(b *testing.B) {
	var m Map[string]

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m.Lock("foo")
		m.Unlock("foo")
	}
}
