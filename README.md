# 🚌 message-bus

[![Test](https://github.com/vardius/message-bus/actions/workflows/test.yml/badge.svg)](https://github.com/vardius/message-bus/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/vardius/message-bus/v2.svg)](https://pkg.go.dev/github.com/vardius/message-bus/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/message-bus/v2)](https://goreportcard.com/report/github.com/vardius/message-bus/v2)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

Publish typed messages to topics, and handle them on their own goroutines, within one Go program.

- **Typed.** A `Bus[T]` carries one message type and handlers are `func(T)`: no reflection, no `interface{}`,
  and a handler that doesn't fit is a compile error.
- **In order.** Each subscriber has its own queue and goroutine, and handles its messages one at a time,
  in the order they were queued.
- **Bounded waits.** When a queue is full, Publish waits for room without holding a lock, so the rest of the bus
  carries on, and gives up when its context is done.
- **Unsubscribe with a handle.** Subscribe returns the func that ends that subscription.
- **Graceful shutdown.** Close waits until every queued message has been handled.
- **Lean.** Publish allocates nothing while there is room, and publishers to different topics rarely share a lock.

📖 **[Documentation](https://vardius.github.io/message-bus/)** · [API reference](https://pkg.go.dev/github.com/vardius/message-bus/v2) · [Examples](examples) · [Migrating from v1](MIGRATION.md)

## Install

```sh
go get github.com/vardius/message-bus/v2
```

Requires Go 1.26 or newer.

## Publish and subscribe

```go
bus := messagebus.New[OrderPlaced](100) // each subscriber queues up to 100 messages

unsubscribe, err := bus.Subscribe("orders", func(e OrderPlaced) {
	sendReceipt(e)
})
if err != nil {
	return err
}
defer unsubscribe()

if err := bus.Publish(ctx, "orders", OrderPlaced{ID: 7}); err != nil {
	return err
}
```

Publish puts the message in the queue of every subscriber of the topic and returns; it doesn't wait for the handlers.
Each handler runs on a goroutine of its own and gets its messages in the order its queue took them.
A topic nobody subscribes to drops the message.

## Full queues

When a subscriber's queue is full, Publish waits until its handler takes a message or the subscriber unsubscribes.
If ctx is done first, Publish gives up and returns ctx's error, wrapped. While it waits it holds no lock, so the rest
of the bus carries on.

```go
ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
defer cancel()
if err := bus.Publish(ctx, "jobs", job); errors.Is(err, context.DeadlineExceeded) {
	// the handler is behind: skip this job, or retry later
}
```

`New(0)` makes unbuffered queues: Publish waits until each handler takes the message.

## Unsubscribe and close

The func Subscribe returns ends that subscription. Messages already in its queue are still handled, and it doesn't
wait for them, so a handler can unsubscribe itself.

Close shuts the whole bus down: it ends every subscription and waits until every queued message has been handled.
After Close, Publish and Subscribe return `ErrClosed`. Stop the publishers first, then close:

```go
srv.Shutdown(ctx) // no more requests, so nothing publishes
if err := bus.Close(ctx); err != nil {
	log.Printf("some messages weren't handled: %v", err)
}
```

## Benchmarks

Time and allocations per Publish, on an Apple M4 Max with Go 1.27.0, median of 5 runs. Both versions ran on the same
harness, with queues of 1024 messages and handlers that count what they get. Run v2's own suite with `go test -bench .`.

| Scenario | v1 | v2 |
| --- | ---: | ---: |
| 1 subscriber | 191 ns · 2 allocs | **48.6 ns · 0 allocs** |
| 10 subscribers | 895 ns · 2 allocs | **449 ns · 0 allocs** |
| 100 subscribers | 13.0 µs · 2 allocs | **9.81 µs · 0 allocs** |
| 100 topics, 1 subscriber each | 147 ns · 2 allocs | **112 ns · 0 allocs** |
| 1 subscriber, 16 publishers at once | 348 ns · 2 allocs | **109 ns · 0 allocs** |
| 100 topics, 16 publishers at once | 159 ns · 2 allocs | **27.7 ns · 0 allocs** |

With 16 publishers at once, the time is wall time per call across all of them, so it measures throughput.
With many subscribers, most of the time goes to the Go scheduler, waking each subscriber's goroutine.
Topics are spread over 64 locks by their hash, so publishers to different topics rarely contend.

## Upgrading from v1

```go
// v1
bus := messagebus.New(100)
handler := func(e OrderPlaced) { sendReceipt(e) }
bus.Subscribe("orders", handler)
bus.Publish("orders", OrderPlaced{ID: 7})
bus.Unsubscribe("orders", handler)

// v2
bus := messagebus.New[OrderPlaced](100)
unsubscribe, err := bus.Subscribe("orders", func(e OrderPlaced) { sendReceipt(e) })
err = bus.Publish(ctx, "orders", OrderPlaced{ID: 7})
unsubscribe()
```

[MIGRATION.md](MIGRATION.md) maps every v1 call to its v2 form and lists the behaviour changes.

## Between programs

message-bus works within one process. [pubsub](https://github.com/vardius/pubsub) is a gRPC server built on
message-bus v1 that delivers published events to subscribed clients over the network.

## Contributing

Bug reports and pull requests are welcome: see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

MIT, see [LICENSE.md](LICENSE.md).
