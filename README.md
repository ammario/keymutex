# keymutex

[![Go Reference](https://pkg.go.dev/badge/github.com/ammario/keymutex.svg)](https://pkg.go.dev/github.com/ammario/keymutex)

keymutex offers a mutex map for serializing access to logical resources.

There are already many packages that do something similar. `keymutex` differs
in that it offers a generic interface and doesn't shard. Thus, it has these
drawbacks:

* its memory usage scales linearly with the number of in-progress locks
* every mutex operation is amplified 3x (map mutex lock and unlock)
* it does not limit concurrency


If you're interested in a sharded mutex, check out [neverlee/keymutex](https://github.com/neverlee/keymutex).


### Install
```
go get github.com/ammario/keymutex@main
```

### Usage

The basics:
```go
var km keymutex.Map[string]

km.Lock("foo")
go func(){
	defer km.Unlock("foo")
	// do something
}()
```

More ergonomic:
```go
km.Go("foo", func(){
	// do something
})
```

## Performance

While `keymutex` performs ~3x worse than sharded mutexes, it's still fast enough
for most use cases.

For example:

```go
func BenchmarkMap_LockUnlock(b *testing.B) {
	var m Map[string]

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m.Lock("foo")
		m.Unlock("foo")
	}
}
```

Gives:
````
goos: darwin
goarch: amd64
pkg: github.com/ammario/keymutex
cpu: VirtualApple @ 2.50GHz
BenchmarkMap_LockUnlock-10      19941822                51.35 ns/op            8 B/op          1 allocs/op
PASS
ok      github.com/ammario/keymutex     1.349s
```