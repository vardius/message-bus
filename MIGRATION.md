# Migrating from v1 to v2

v2 keeps the idea: publish to a topic, and every subscriber's handler gets the message on its own goroutine,
in order. What changes is the shape. A bus carries one message type, so handlers take it typed instead of
through reflection; Subscribe hands back a func that unsubscribes; Publish takes a context; and Close shuts
the whole bus down, waiting for queued messages. Every call and every behaviour change is listed below.

## 1. Update the import path

```sh
go get github.com/vardius/message-bus/v2
```

```go
import messagebus "github.com/vardius/message-bus/v2"
```

v2 needs Go 1.26 or newer.

## 2. Before and after

```go
// v1
bus := messagebus.New(100)

handler := func(e OrderPlaced) { sendReceipt(e) }
if err := bus.Subscribe("orders", handler); err != nil {
	return err
}
bus.Publish("orders", OrderPlaced{ID: 7})

bus.Unsubscribe("orders", handler)
```

```go
// v2
bus := messagebus.New[OrderPlaced](100)

unsubscribe, err := bus.Subscribe("orders", func(e OrderPlaced) { sendReceipt(e) })
if err != nil {
	return err
}
if err := bus.Publish(ctx, "orders", OrderPlaced{ID: 7}); err != nil {
	return err
}

unsubscribe()
```

A v1 handler took any number of arguments; a v2 handler takes one message. If yours took several, make
them the fields of a struct and use that as the bus's type. To carry different kinds of message, use a bus
per kind, or an interface they all implement.

## 3. Every call

| v1 | v2 |
| --- | --- |
| `New(handlerQueueSize int) MessageBus` | `New[T any](queueSize int) *Bus[T]` |
| `type MessageBus interface { … }` | `type Bus[T any] struct { … }`: declare an interface where you use it, if you need one |
| `Publish(topic string, args ...interface{})` | `Publish(ctx context.Context, topic string, msg T) error` |
| `Subscribe(topic string, fn interface{}) error` | `Subscribe(topic string, fn func(T)) (unsubscribe func(), err error)` |
| `Unsubscribe(topic string, fn interface{}) error` | `unsubscribe()`, the func Subscribe returned |
| `Close(topic string)` | `unsubscribe()` for each subscription to that topic |
| No way to wait for handlers | `Close(ctx context.Context) error` shuts the whole bus down and waits for them |

## 4. Behaviour changes

- **Handlers are checked by the compiler.** v1 accepted any value as a handler and returned an error if it
  wasn't a func; a message whose arguments didn't match the handler's panicked on the subscriber's goroutine
  and crashed the program. In v2 both are compile errors.
- **Unsubscribing uses the func Subscribe returned.** v1 matched handlers by func value, which failed in
  several ways: a method value such as `s.Handle` is a new func each time, so it was never found and
  Unsubscribe returned nil; and of a func subscribed twice, Unsubscribe removed both, one, or panicked,
  depending on where the two sat among the topic's subscribers. In v2 every Subscribe makes its own
  subscription, and its func removes exactly that one. Calling it again does nothing.
- **Unsubscribing never fails.** v1 returned an error for a topic with no subscribers; the v2 func has
  nothing to report. As in v1, messages already in the subscriber's queue are still handled.
- **A full queue no longer holds up the bus.** v1's Publish waited for room while holding the bus's lock, so
  Subscribe, Unsubscribe and Close on any topic waited too, then every other Publish behind them; a handler
  that subscribed or unsubscribed while Publish waited for room in its own queue deadlocked. v2's Publish waits without a lock,
  and gives up when its context is done, returning the context's error.
- **Publish returns an error:** the context's, wrapped, if it gave up on a full queue, or `ErrClosed` after
  Close. Messages to a topic without subscribers are dropped without an error, as before.
- **Close shuts the bus down and waits.** v1's `Close(topic)` unsubscribed a topic's handlers and returned at
  once. v2's `Close(ctx)` ends every subscription, waits until every queued message has been handled, and
  makes later Publish and Subscribe calls return `ErrClosed`. If ctx is done first, it returns ctx's error.
- **Queue sizes are checked in New.** `New(0)` panicked in v1 and now makes unbuffered queues: Publish waits
  until each handler takes the message. A negative size panics in `New`, where v1 panicked in the first
  `Subscribe`.
- **A nil handler panics in Subscribe** with "messagebus: nil handler", rather than with a nil pointer
  dereference.
