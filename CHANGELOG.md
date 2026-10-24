# Changelog

## v2.0.0 (2026-10-24)

A generic, reflection-free bus. See [MIGRATION.md](MIGRATION.md) for the before and after of every call.

### Added

- `Bus[T]`: a bus carries messages of one type, and handlers are `func(T)`, checked by the compiler.
- `Subscribe` returns a func that ends that subscription, and only that one.
- `Close(ctx)` shuts the bus down: it ends every subscription and waits until every queued message has been
  handled, or ctx is done.
- `ErrClosed`, which `Publish` and `Subscribe` return after `Close`.
- Runnable examples in `examples/`, each with a test: order events, backpressure, and a graceful shutdown.

### Changed

- Module path is `github.com/vardius/message-bus/v2`; Go 1.26 or newer is required.
- `Publish` takes a context and returns an error. When a queue is full it waits for room without holding a
  lock, so Subscribe, unsubscribing, Close and publishing to other topics carry on, and it gives up when the
  context is done.
- `New(0)` makes unbuffered queues instead of panicking; a negative size panics in `New`.
- Faster, and `Publish` allocates nothing while there is room. Publishing to one subscriber went from 191 ns and 2 allocations to
  48.6 ns and none; with 16 goroutines publishing to 100 topics, from 159 ns to 27.7 ns. Allocation counts are
  guarded by tests.

### Removed

- The `MessageBus` interface, in favour of the concrete `*Bus[T]`.
- `Unsubscribe(topic, fn)`, which matched handlers by func value, and `Close(topic)`.

### Fixed

- A handler that subscribed or unsubscribed while `Publish` waited for room in its own queue deadlocked.
- A full queue stalled `Subscribe`, `Unsubscribe`, `Close` and then every `Publish`, on every topic.
- A handler that called `runtime.Goexit` ended its subscription's goroutine, so its queue was never emptied
  again. The subscription now carries on with the next message.
- `Unsubscribe` of a handler subscribed twice removed both, removed one, or panicked, depending on where the
  two sat among the topic's subscribers. It never found a method value, and returned nil.
