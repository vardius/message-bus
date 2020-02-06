🚌 message-bus
================
[![Build Status](https://travis-ci.org/vardius/message-bus.svg?branch=master)](https://travis-ci.org/vardius/message-bus)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/message-bus)](https://goreportcard.com/report/github.com/vardius/message-bus)
[![codecov](https://codecov.io/gh/vardius/message-bus/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/message-bus)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvardius%2Fmessage-bus.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvardius%2Fmessage-bus?ref=badge_shield)
[![](https://godoc.org/github.com/vardius/message-bus?status.svg)](http://godoc.org/github.com/vardius/message-bus)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/message-bus/blob/master/LICENSE.md)

<img align="right" height="180px" src="docs/img/logo.png" alt="logo" />

Go simple async message bus.

📖 ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/message-bus/issues) to manage them.

## 📚 Documentation

For **documentation** (_including examples_), **visit [rafallorenz.com/message-bus](http://rafallorenz.com/message-bus)**

For **GoDoc** reference, **visit [godoc.org/github.com/vardius/message-bus](http://godoc.org/github.com/vardius/message-bus)**

🚏 HOW TO USE
==================================================

## 🚅 Benchmark

Time complexity of a `Publish` method is considered to be [linear time `O(n)`](https://en.wikipedia.org/wiki/Time_complexity#Linear_time). Where **n** corresponds to the number of *subscribers* for a given **topic**.

**CPU: 3,3 GHz Intel Core i7**

**RAM: 16 GB 2133 MHz LPDDR3**

```bash
➜  message-bus git:(master) ✗ go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/vardius/message-bus
BenchmarkPublish-4                   	 4430224	       250 ns/op	       0 B/op	       0 allocs/op
BenchmarkSubscribe-4                 	  598240	      2037 ns/op	     735 B/op	       5 allocs/op
```

## 🏫 Basic example
```go
package main

import (
    "fmt"
	"sync"

    "github.com/vardius/message-bus"
)

func main() {
    queueSize := 100
    bus := messagebus.New(queueSize)

    var wg sync.WaitGroup
    wg.Add(2)

    _ = bus.Subscribe("topic", func(v bool) {
        defer wg.Done()
        fmt.Println(v)
    })

    _ = bus.Subscribe("topic", func(v bool) {
        defer wg.Done()
        fmt.Println(v)
    })

    // Publish block only when the buffer of one of the subscribers is full.
    // change the buffer size altering queueSize when creating new messagebus
    bus.Publish("topic", true)
    wg.Wait()
}
```

## 🐳 Docker service

[pubsub](https://github.com/vardius/pubsub) - gRPC message-oriented middleware on top of message-bus, event ingestion and delivery system.

📜 [License](LICENSE.md)
-------

This package is released under the MIT license. See the complete license in the package:

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvardius%2Fmessage-bus.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvardius%2Fmessage-bus?ref=badge_large)
