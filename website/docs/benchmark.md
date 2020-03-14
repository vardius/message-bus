---
id: benchmark
title: Benchmark
sidebar_label: Benchmark
---

Time complexity of a `Publish` method is considered to be [linear time `O(n)`](https://en.wikipedia.org/wiki/Time_complexity#Linear_time). Where **n** corresponds to the number of *subscribers* for a given **topic**.

**Last Test Updated:** 2020-03-14

*test environment*

- **Processor** 3.3 GHz Dual-Core Intel Core i7
- **Memory** 16 GB 2133 MHz LPDDR3
- **Go** go1.13.1 darwin/amd64
- **OS** macOs Catalina 10.15.3

### Built-in

```bash
➜  message-bus git:(master) ✗ go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/vardius/message-bus
BenchmarkPublish-4                   	 4430224	       250 ns/op	       0 B/op	       0 allocs/op
BenchmarkSubscribe-4                 	  598240	      2037 ns/op	     735 B/op	       5 allocs/op
```
