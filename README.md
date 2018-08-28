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
goos: darwin
goarch: amd64
pkg: github.com/vardius/message-bus
BenchmarkWorker100-4                           3         352169727 ns/op
BenchmarkWorkerNumCPU-4                        5         301247338 ns/op
BenchmarkWorker-4                             10         182287824 ns/op
BenchmarkWorker100Parallel-4                  10         117857721 ns/op
BenchmarkWorkerNumCPUParallel-4               10         113668659 ns/op
BenchmarkWorkerParallel-4                     10         105661820 ns/op
PASS
ok      github.com/vardius/message-bus  15.889s
```

## Basic example
```go
package main

import (
    "fmt"
    "runtime"

    "github.com/vardius/message-bus"
)

func main() {
    bus := messagebus.New(runtime.NumCPU())

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

    bus.Publish("topic", true)
    wg.Wait()
}
```

License
-------

This package is released under the MIT license. See the complete license in the package:

[LICENSE](LICENSE.md)
