🚌 message-bus
================
[![Build Status](https://travis-ci.org/taills/message-bus.svg?branch=master)](https://travis-ci.org/taills/message-bus)
[![Go Report Card](https://goreportcard.com/badge/github.com/taills/message-bus)](https://goreportcard.com/report/github.com/taills/message-bus)
[![codecov](https://codecov.io/gh/taills/message-bus/branch/master/graph/badge.svg)](https://codecov.io/gh/taills/message-bus)
[![](https://godoc.org/github.com/taills/message-bus?status.svg)](https://pkg.go.dev/github.com/taills/message-bus)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/taills/message-bus/blob/master/LICENSE.md)

<img align="right" height="180px" src="website/src/static/img/logo.png" alt="logo" />

Go simple async message bus.

> **Note**: This repository is forked from [github.com/vardius/message-bus](https://github.com/vardius/message-bus)

📖 ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/taills/message-bus/issues) to manage them.

## 📚 Documentation

For **documentation** (_including examples_), **visit [rafallorenz.com/message-bus](http://rafallorenz.com/message-bus)**

For **GoDoc** reference, **visit [pkg.go.dev](https://pkg.go.dev/github.com/taills/message-bus)**

🚏 HOW TO USE
==================================================

## 🚅 Benchmark

```bash
❯ go test -bench=. -cpu=8 -benchmem
goos: darwin
goarch: arm64
pkg: github.com/taills/message-bus
cpu: Apple M3 Max
BenchmarkPublish-8                   	 9355280	       128.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkSubscribe-8                 	 2483044	       532.5 ns/op	     824 B/op	       6 allocs/op
Benchmark1Subscriber1Topic-8         	10066575	       119.5 ns/op	       0 B/op	       0 allocs/op
Benchmark1Subscriber100Topics-8      	 9209472	       141.7 ns/op	       0 B/op	       0 allocs/op
Benchmark100Subscribers1Topic-8      	  180798	      5714 ns/op	       0 B/op	       0 allocs/op
Benchmark100Subscribers100Topics-8   	  158925	      6465 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/taills/message-bus	42.235s
```

👉 **[Click here](https://rafallorenz.com/message-bus/docs/benchmark)** to see all benchmark results.

## Features
- [Documentation](https://rafallorenz.com/message-bus/)

🚏 HOW TO USE
==================================================

- [Basic example](https://rafallorenz.com/message-bus/docs/basic-example)
- [Pub/Sub](https://rafallorenz.com/message-bus/docs/pubsub)

📜 [License](LICENSE.md)
-------

This package is released under the MIT license. See the complete license in the package.
