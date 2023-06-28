package keymutex

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMap_Leak(t *testing.T) {
	var m Map[string]

	m.Go("blah", func() {
		m.Lock("blah")
		panic("should never be reached")
	})
}

func TestMap_LockCtx(t *testing.T) {
	t.Parallel()

	var m Map[string]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ok := m.LockCtx(ctx, "foo")
	require.True(t, ok)

	m.Unlock("foo")

	m.Lock("foo")

	// Should fail in a second.

	ctx, cancel = context.WithTimeout(ctx, time.Second)
	defer cancel()

	startWaiting := time.Now()
	ok = m.LockCtx(ctx, "foo")
	require.False(t, ok)

	require.WithinDuration(t, startWaiting.Add(time.Second), time.Now(), time.Millisecond*50)
}

func TestMap(t *testing.T) {
	t.Parallel()

	var m Map[string]
	// Zero value should be usable.
	m.Lock("foo")

	m.Unlock("foo")

	time.Sleep(time.Millisecond * 100)

	m.Lock("bar")
	m.Unlock("bar")

	m.Do("bar", func() {})

	m.Lock("bar")
	m.Unlock("bar")

	if l := m.Len(); l != 0 {
		t.Fatal("expected 0 mutex, got", l)
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
