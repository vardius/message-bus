Vardius - message-bus
================
[![Build Status](https://travis-ci.org/vardius/message-bus.svg?branch=master)](https://travis-ci.org/vardius/message-bus)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/message-bus)](https://goreportcard.com/report/github.com/vardius/message-bus)
[![codecov](https://codecov.io/gh/vardius/message-bus/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/message-bus)
[![](https://godoc.org/github.com/vardius/message-bus?status.svg)](http://godoc.org/github.com/vardius/message-bus)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/message-bus/blob/master/LICENSE.md)

Go simple async message bus.

ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/message-bus/issues) to manage them.

HOW TO USE
==================================================

1. [GoDoc](http://godoc.org/github.com/vardius/message-bus)

## Benchmark
**CPU: 3,3 GHz Intel Core i7**

**RAM: 16 GB 2133 MHz LPDDR3**

```bash
➜  message-bus git:(master) ✗ go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
BenchmarkBus-4                   3000000               534 ns/op              56 B/op          3 allocs/op
BenchmarkBusParallel-4           5000000               313 ns/op              48 B/op          2 allocs/op
BenchmarkBus100-4                 100000             14651 ns/op              56 B/op          3 allocs/op
BenchmarkBus100Parallel-4         300000             14130 ns/op              48 B/op          2 allocs/op
BenchmarkBus1000-4                 10000            159269 ns/op              56 B/op          3 allocs/op
BenchmarkBus1000Parallel-4         10000            142578 ns/op              48 B/op          2 allocs/op
BenchmarkBusNumCPU-4             1000000              1155 ns/op              56 B/op          3 allocs/op
BenchmarkBusNumCPUParallel-4     2000000               774 ns/op              48 B/op          2 allocs/op
PASS
ok      message-bus    23.125s
```

## Basic example
```go
package main

import (
    "fmt"

    "github.com/vardius/message-bus"
)

func main() {
    queueSize := 100
    bus := messagebus.New(queueSize)

    var wg sync.WaitGroup
    wg.Add(2)

    bus.Subscribe("topic", func(v bool) {
        defer wg.Done()
        fmt.Println(v)
    })

    bus.Subscribe("topic", func(v bool) {
        defer wg.Done()
        fmt.Println(v)
    })

    // Publish block only when the buffer of one of the subscribers is full.
    // change the buffer size altering queueSize when creating new messagebus
    bus.Publish("topic", true)
    wg.Wait()
}
```

## Docker service

1. [pubsub](https://github.com/vardius/pubsub) - gRPC message-oriented middleware on top of message-bus, event ingestion and delivery system.

License
-------

This package is released under the MIT license. See the complete license in the package:

[LICENSE](LICENSE.md)
