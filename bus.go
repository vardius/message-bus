package messagebus

import (
	"context"
	"errors"
	"fmt"
	"hash/maphash"
	"slices"
	"sync"
)

// ErrClosed is what Publish and Subscribe return once Close has been called.
var ErrClosed = errors.New("messagebus: bus closed")

// Bus delivers messages of type T to the subscribers of a topic. Create one with New; its methods are safe
// to call from any goroutine, handlers included.
type Bus[T any] struct {
	queueSize int
	seed      maphash.Seed
	shards    [shardCount]shard[T]

	closeOnce sync.Once
	handlers  sync.WaitGroup // one per subscriber goroutine
	drained   chan struct{}  // closed once Close has been called and every handler has returned
}

// Topics are spread over shards, each with its own lock, so publishers to different topics rarely contend.
const shardCount = 64

type shard[T any] struct {
	mu     sync.RWMutex
	topics map[string][]*subscriber[T] // never changed in place: writers store a new slice
	closed bool
	_      [88]byte // pads the shard to 128 bytes, so neighbouring shards' locks don't share a cache line
}

type subscriber[T any] struct {
	topic string
	fn    func(T)
	queue chan T
	// done is closed when the subscriber is removed, which wakes publishers waiting on its full queue.
	done    chan struct{}
	removed bool           // guarded by the shard's lock
	senders sync.WaitGroup // publishers waiting for room without the lock: the queue is closed only after them
}

// New returns a Bus that gives each subscriber a queue of queueSize messages.
// A queueSize of 0 means unbuffered: Publish waits until each subscriber takes the message.
// New panics if queueSize is negative.
func New[T any](queueSize int) *Bus[T] {
	if queueSize < 0 {
		panic("messagebus: negative queue size")
	}

	b := &Bus[T]{queueSize: queueSize, seed: maphash.MakeSeed(), drained: make(chan struct{})}
	for i := range b.shards {
		b.shards[i].topics = make(map[string][]*subscriber[T])
	}
	return b
}

func (b *Bus[T]) shard(topic string) *shard[T] {
	return &b.shards[maphash.String(b.seed, topic)%shardCount]
}

// Publish puts msg in the queue of every subscriber of topic, in the order they subscribed, and returns.
// It doesn't wait for handlers, only for room: when a queue is full, Publish waits until it has room,
// the subscriber unsubscribes, or ctx is done. ctx is only consulted then.
//
// If ctx is done first, Publish returns its error, wrapped: subscribers before the full one have the message,
// the rest don't get it. After Close, or when Close is called while it waits and before every queue has the
// message, Publish returns ErrClosed; the queues that took it still have it handled.
// A topic without subscribers drops msg.
func (b *Bus[T]) Publish(ctx context.Context, topic string, msg T) error {
	sh := b.shard(topic)
	sh.mu.RLock()
	if sh.closed {
		sh.mu.RUnlock()
		return ErrClosed
	}

	subs := sh.topics[topic]
	for i, s := range subs {
		// Only possible after waiting below, which let go of the lock.
		if s.removed {
			continue
		}
		select {
		case s.queue <- msg:
			continue
		default:
		}

		// The queue is full. Wait for room without the lock, so subscribing, unsubscribing and closing never
		// wait on a slow handler; senders keeps the queue open until this message is in or given up.
		s.senders.Add(1)
		sh.mu.RUnlock()
		queued, err := s.send(ctx, msg)
		if err != nil {
			return fmt.Errorf("messagebus: publish to %q: %w", topic, err)
		}
		sh.mu.RLock()
		// Closed while waiting: ErrClosed, unless the message is in every queue already, and so will be handled.
		if sh.closed && (!queued || i < len(subs)-1) {
			sh.mu.RUnlock()
			return ErrClosed
		}
	}
	sh.mu.RUnlock()

	return nil
}

// send waits for room in the queue, and reports whether msg went in: it doesn't when s is removed first.
func (s *subscriber[T]) send(ctx context.Context, msg T) (queued bool, err error) {
	defer s.senders.Done()

	select {
	case s.queue <- msg:
		return true, nil
	case <-s.done:
		return false, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// Subscribe calls fn with every message published to topic from now on, one at a time, in the order its queue
// took them, on a goroutine of its own. Each call makes a new subscription, even for the same fn.
//
// unsubscribe stops the subscription: messages already in its queue are still handled, then its goroutine ends.
// It doesn't wait for that, so a handler can unsubscribe itself, and calling it again does nothing.
// After Close, Subscribe returns ErrClosed. It panics if fn is nil.
func (b *Bus[T]) Subscribe(topic string, fn func(T)) (unsubscribe func(), err error) {
	if fn == nil {
		panic("messagebus: nil handler")
	}

	s := &subscriber[T]{topic: topic, fn: fn, queue: make(chan T, b.queueSize), done: make(chan struct{})}
	sh := b.shard(topic)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if sh.closed {
		return nil, ErrClosed
	}

	sh.topics[topic] = append(slices.Clip(sh.topics[topic]), s)
	b.handlers.Add(1)
	go s.run(&b.handlers)

	return func() { b.unsubscribe(sh, s) }, nil
}

func (s *subscriber[T]) run(handlers *sync.WaitGroup) {
	finished := false
	defer func() {
		if !finished {
			// fn called runtime.Goexit, which ends its goroutine but not the subscription: carry on, on a new one.
			handlers.Add(1)
			go s.run(handlers)
		}
		handlers.Done()
	}()

	for msg := range s.queue {
		s.fn(msg)
	}
	finished = true
}

func (b *Bus[T]) unsubscribe(sh *shard[T], s *subscriber[T]) {
	sh.mu.Lock()
	if s.removed {
		sh.mu.Unlock()
		return
	}
	s.remove()
	if subs := sh.topics[s.topic]; len(subs) == 1 {
		delete(sh.topics, s.topic)
	} else {
		i := slices.Index(subs, s)
		sh.topics[s.topic] = slices.Concat(subs[:i], subs[i+1:])
	}
	sh.mu.Unlock()

	s.closeQueue()
}

// remove marks s removed, so no publisher starts sending to it, and wakes those waiting on its full queue.
// The shard's lock must be held.
func (s *subscriber[T]) remove() {
	s.removed = true
	close(s.done)
}

// closeQueue closes the queue of a removed subscriber once no publisher can send to it, which ends its
// goroutine after the queued messages are handled. It waits only for publishers that remove has woken.
func (s *subscriber[T]) closeQueue() {
	s.senders.Wait()
	close(s.queue)
}

// Close stops the bus and waits for every handler to finish: Publish and Subscribe return ErrClosed from now on,
// publishers waiting on a full queue give up, and messages already queued are handled before Close returns.
//
// If ctx is done first, Close returns its error, wrapped, and the handlers carry on with the queued messages.
// Every call waits in the same way, so it's safe to call Close more than once. A handler that calls Close
// waits for itself: that Close returns only when its ctx is done.
func (b *Bus[T]) Close(ctx context.Context) error {
	b.closeOnce.Do(b.shutdown)

	select {
	case <-b.drained:
		return nil
	case <-ctx.Done():
		select {
		case <-b.drained: // both were ready: select picked at random
			return nil
		default:
			return fmt.Errorf("messagebus: close: %w", ctx.Err())
		}
	}
}

func (b *Bus[T]) shutdown() {
	var removed []*subscriber[T]
	for i := range b.shards {
		sh := &b.shards[i]
		sh.mu.Lock()
		sh.closed = true
		for _, subs := range sh.topics {
			for _, s := range subs {
				s.remove()
				removed = append(removed, s)
			}
		}
		clear(sh.topics)
		sh.mu.Unlock()
	}

	for _, s := range removed {
		s.closeQueue()
	}
	go func() {
		b.handlers.Wait()
		close(b.drained)
	}()
}
