// Package messagebus delivers messages to the subscribers of a topic, within one program.
//
// A Bus carries messages of one type. Publish puts a message in the queue of every subscriber of a topic and
// returns; each subscriber's handler runs on a goroutine of its own and takes its messages one at a time.
//
//	bus := messagebus.New[OrderPlaced](100)
//	unsubscribe, err := bus.Subscribe("orders", func(e OrderPlaced) {
//		fmt.Println("send the receipt for", e.ID)
//	})
//	...
//	err = bus.Publish(ctx, "orders", OrderPlaced{ID: 7})
//
// The type parameter is the message's type, so handlers take it typed: no reflection, no type assertions.
// To send several kinds of message, use one Bus per type, or an interface they all implement.
//
// # Order
//
// Each subscriber gets messages in the order its queue took them, so messages published from one goroutine
// arrive in the order they were published. Handlers don't wait for each other, but a Publish that finds a queue
// full waits for room before it moves on to the subscribers after it. Messages from goroutines publishing at
// the same time can interleave differently for different subscribers.
//
// # Full queues
//
// Every subscriber has a queue of the size given to New. Publish doesn't wait for handlers, only for room:
// when a queue is full, it waits until the handler takes a message, the subscriber unsubscribes, or the
// context is done. While it waits it holds no lock, so other topics, Subscribe, unsubscribe and Close carry on.
// A handler that publishes to its own topic and finds its own queue full waits for itself: give that Publish
// a context with a deadline, or the bus a bigger queue.
//
// # Unsubscribing and closing
//
// Subscribe returns a func that ends the subscription. Messages already in its queue are still handled;
// the func doesn't wait for that, so a handler can call it to unsubscribe itself.
//
// Close ends every subscription and waits until the handlers have handled every queued message and returned.
// After Close, Publish and Subscribe return ErrClosed. A message Publish put in a queue is always handled,
// unless the program exits first. Publish calls that run at the same time as unsubscribe or Close may or
// may not reach that subscriber: stop publishing first, then Close.
//
// A handler that panics crashes the program, as a panic on any goroutine does: recover in the handler
// to carry on. A handler that calls runtime.Goexit ends that call only: the subscription goes on with its next
// message, on a new goroutine.
//
// # Cost
//
// Publish allocates nothing, but for two things: when it waits for room, a cancellable ctx makes its Done channel
// the first time it is asked for it, and when it gives up, it wraps ctx's error. It copies the message into each queue, so pass a pointer for a large message;
// the handlers then share what it points to. Topics are spread over 64 locks by their hash, so publishers
// to different topics rarely contend. Subscribe allocates the subscriber, its queue and its goroutine.
package messagebus
