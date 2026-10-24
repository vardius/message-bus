package messagebus_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	messagebus "github.com/vardius/message-bus/v2"
)

// recorder is a handler that keeps every message it is given.
type recorder[T comparable] struct {
	mu  sync.Mutex
	got []T
}

func (r *recorder[T]) handle(msg T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.got = append(r.got, msg)
}

func (r *recorder[T]) messages() []T {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.got)
}

func (r *recorder[T]) want(t *testing.T, name string, want ...T) {
	t.Helper()
	if got := r.messages(); !slices.Equal(got, want) {
		t.Errorf("%s got %v, want %v", name, got, want)
	}
}

// blocked returns a handler that records its messages, and waits for release before handling the first one.
func blocked[T comparable](r *recorder[T], release <-chan struct{}) func(T) {
	first := true
	return func(msg T) {
		if first {
			first = false
			<-release
		}
		r.handle(msg)
	}
}

func subscribe[T any](t *testing.T, bus *messagebus.Bus[T], topic string, fn func(T)) func() {
	t.Helper()
	unsubscribe, err := bus.Subscribe(topic, fn)
	if err != nil {
		t.Fatalf("Subscribe(%q) error = %v", topic, err)
	}
	return unsubscribe
}

func publish[T any](t *testing.T, bus *messagebus.Bus[T], topic string, msgs ...T) {
	t.Helper()
	for _, msg := range msgs {
		if err := bus.Publish(t.Context(), topic, msg); err != nil {
			t.Fatalf("Publish(%q, %v) error = %v", topic, msg, err)
		}
	}
}

// publishInBackground starts a Publish that is expected to wait, and returns where its error will arrive.
func publishInBackground[T any](t *testing.T, bus *messagebus.Bus[T], topic string, msg T) <-chan error {
	t.Helper()
	errc := make(chan error, 1)
	go func() { errc <- bus.Publish(t.Context(), topic, msg) }()
	synctest.Wait()
	select {
	case err := <-errc:
		t.Fatalf("Publish(%q, %v) = %v, want it to wait for room", topic, msg, err)
	default:
	}
	return errc
}

func closeBus[T any](t *testing.T, bus *messagebus.Bus[T]) {
	t.Helper()
	if err := bus.Close(t.Context()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestPublish(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[string](4)
		var a, b, other recorder[string]
		subscribe(t, bus, "orders", a.handle)
		subscribe(t, bus, "orders", b.handle)
		subscribe(t, bus, "payments", other.handle)

		publish(t, bus, "orders", "placed", "paid")
		publish(t, bus, "nobody listens", "dropped")
		closeBus(t, bus)

		a.want(t, "first subscriber", "placed", "paid")
		b.want(t, "second subscriber", "placed", "paid")
		other.want(t, "other topic")
	})
}

func TestOrderPerSubscriber(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// A queue of 2 makes Publish wait for room most of the time.
		bus := messagebus.New[int](2)
		subs := make([]recorder[int], 3)
		for i := range subs {
			subscribe(t, bus, "seq", subs[i].handle)
		}

		want := make([]int, 1000)
		for i := range want {
			want[i] = i
		}
		publish(t, bus, "seq", want...)
		closeBus(t, bus)

		for i := range subs {
			subs[i].want(t, fmt.Sprintf("subscriber %d", i), want...)
		}
	})
}

func TestSubscribeSameHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](4)
		// v1 matched handlers by func value. Depending on where the copies sat, its Unsubscribe removed one,
		// removed both, or panicked; and a method value like r.handle never matched at all.
		var r, other recorder[int]
		first := subscribe(t, bus, "t", r.handle)
		subscribe(t, bus, "t", other.handle)
		second := subscribe(t, bus, "t", r.handle)
		third := subscribe(t, bus, "t", r.handle)

		second()
		publish(t, bus, "t", 1)
		synctest.Wait()
		r.want(t, "after one unsubscribe", 1, 1)

		first()
		third()
		second() // again: does nothing
		publish(t, bus, "t", 2)
		closeBus(t, bus)
		r.want(t, "after all unsubscribed", 1, 1)
		other.want(t, "other subscriber", 1, 2)
	})
}

func TestUnsubscribeHandlesQueued(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](4)
		release := make(chan struct{})
		var r recorder[int]
		unsubscribe := subscribe(t, bus, "t", blocked(&r, release))

		publish(t, bus, "t", 1, 2, 3)
		synctest.Wait()
		unsubscribe()
		publish(t, bus, "t", 4)

		close(release)
		closeBus(t, bus)
		r.want(t, "subscriber", 1, 2, 3)
	})
}

func TestUnsubscribeFromHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](4)
		var r recorder[int]
		var unsubscribe func()
		ready := make(chan struct{})
		unsubscribe = subscribe(t, bus, "once", func(n int) {
			<-ready
			r.handle(n)
			unsubscribe()
		})
		close(ready)

		publish(t, bus, "once", 1)
		synctest.Wait()
		publish(t, bus, "once", 2)
		closeBus(t, bus)
		r.want(t, "subscriber", 1)
	})
}

func TestFullQueue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](1)
		release := make(chan struct{})
		var r recorder[int]
		subscribe(t, bus, "t", blocked(&r, release))

		publish(t, bus, "t", 1) // the handler takes it and waits
		synctest.Wait()
		publish(t, bus, "t", 2) // fills the queue
		errc := publishInBackground(t, bus, "t", 3)

		close(release)
		if err := <-errc; err != nil {
			t.Errorf("Publish() error = %v", err)
		}
		closeBus(t, bus)
		r.want(t, "subscriber", 1, 2, 3)
	})
}

func TestUnbuffered(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](0)
		release := make(chan struct{})
		var r recorder[int]
		subscribe(t, bus, "t", blocked(&r, release))
		synctest.Wait()

		publish(t, bus, "t", 1) // the handler is waiting for it
		errc := publishInBackground(t, bus, "t", 2)

		close(release)
		if err := <-errc; err != nil {
			t.Errorf("Publish() error = %v", err)
		}
		closeBus(t, bus)
		r.want(t, "subscriber", 1, 2)
	})
}

// fullQueue returns a bus whose subscriber to "full" has a full queue of one message and a handler that
// waits for release. Its messages go to r.
func fullQueue(t *testing.T, r *recorder[string], release <-chan struct{}) *messagebus.Bus[string] {
	t.Helper()
	bus := messagebus.New[string](1)
	subscribe(t, bus, "full", blocked(r, release))
	publish(t, bus, "full", "taken")
	synctest.Wait()
	publish(t, bus, "full", "queued")
	return bus
}

func TestPublishContextDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		var before, full, after recorder[string]
		bus := messagebus.New[string](1)
		subscribe(t, bus, "t", before.handle)
		subscribe(t, bus, "t", blocked(&full, release))
		subscribe(t, bus, "t", after.handle)
		publish(t, bus, "t", "taken")
		synctest.Wait()
		publish(t, bus, "t", "queued")

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		start := time.Now()
		err := bus.Publish(ctx, "t", "late")

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Publish() error = %v, want %v", err, context.DeadlineExceeded)
		}
		if want := `messagebus: publish to "t": context deadline exceeded`; err.Error() != want {
			t.Errorf("Publish() error = %q, want %q", err, want)
		}
		if waited := time.Since(start); waited != time.Second {
			t.Errorf("Publish() waited %v, want 1s", waited)
		}

		close(release)
		closeBus(t, bus)
		before.want(t, "subscriber before the full one", "taken", "queued", "late")
		full.want(t, "full subscriber", "taken", "queued")
		after.want(t, "subscriber after the full one", "taken", "queued")
	})
}

func TestUnsubscribeWakesPublisher(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		var full, next recorder[string]
		bus := messagebus.New[string](1)
		unsubscribe := subscribe(t, bus, "t", blocked(&full, release))
		subscribe(t, bus, "t", next.handle)
		publish(t, bus, "t", "taken")
		synctest.Wait()
		publish(t, bus, "t", "queued")
		errc := publishInBackground(t, bus, "t", "late")

		unsubscribe()
		if err := <-errc; err != nil {
			t.Errorf("Publish() error = %v", err)
		}

		close(release)
		closeBus(t, bus)
		full.want(t, "unsubscribed", "taken", "queued")
		next.want(t, "next subscriber", "taken", "queued", "late")
	})
}

// A publisher waiting on one queue mustn't hold up the rest of the bus. In v1 it held the bus's lock while
// it waited, so each call below hung.
func TestWaitingPublisherBlocksNothing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		var full, later, other recorder[string]
		bus := fullQueue(t, &full, release)
		unsubscribeLater := subscribe(t, bus, "full", later.handle)
		errc := publishInBackground(t, bus, "full", "late")

		subscribe(t, bus, "other", other.handle)
		publish(t, bus, "other", "elsewhere")
		subscribe(t, bus, "full", func(string) {})()
		// Unsubscribing a subscriber the publisher hasn't reached yet doesn't wait for it either.
		unsubscribeLater()

		close(release)
		if err := <-errc; err != nil {
			t.Errorf("Publish() error = %v", err)
		}
		closeBus(t, bus)
		full.want(t, "full subscriber", "taken", "queued", "late")
		later.want(t, "unsubscribed while the publisher waited")
		other.want(t, "other topic", "elsewhere")
	})
}

// In v1 a handler that subscribed while Publish waited for room in its own queue deadlocked: Publish held
// the read lock the handler's Subscribe needed, and only the handler could make room.
func TestHandlerSubscribesWhilePublishWaits(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](1)
		release := make(chan struct{})
		var r, inner recorder[int]
		subscribe(t, bus, "t", func(n int) {
			r.handle(n)
			if n != 1 {
				return
			}
			<-release
			unsubscribe, err := bus.Subscribe("t", inner.handle)
			if err != nil {
				t.Errorf("Subscribe() from a handler error = %v", err)
				return
			}
			unsubscribe()
		})
		publish(t, bus, "t", 1)
		synctest.Wait()
		publish(t, bus, "t", 2)
		errc := publishInBackground(t, bus, "t", 3)

		close(release)
		if err := <-errc; err != nil {
			t.Errorf("Publish() error = %v", err)
		}
		closeBus(t, bus)
		r.want(t, "subscriber", 1, 2, 3)
		inner.want(t, "subscribed from the handler")
	})
}

func TestClose(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](4)
		release := make(chan struct{})
		var r, unsubscribed recorder[int]
		subscribe(t, bus, "t", blocked(&r, release))
		unsubscribe := subscribe(t, bus, "u", blocked(&unsubscribed, release))
		publish(t, bus, "t", 1, 2, 3)
		publish(t, bus, "u", 1, 2)
		synctest.Wait()
		unsubscribe()

		closed := make(chan error)
		go func() { closed <- bus.Close(t.Context()) }()
		synctest.Wait()
		select {
		case err := <-closed:
			t.Fatalf("Close() = %v before the queued messages were handled", err)
		default:
		}

		close(release)
		if err := <-closed; err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		// Close waits for subscriptions that were ended earlier too.
		r.want(t, "subscriber", 1, 2, 3)
		unsubscribed.want(t, "unsubscribed before Close", 1, 2)
	})
}

func TestAfterClose(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](4)
		var r recorder[int]
		unsubscribe := subscribe(t, bus, "t", r.handle)
		closeBus(t, bus)

		if err := bus.Publish(t.Context(), "t", 1); !errors.Is(err, messagebus.ErrClosed) {
			t.Errorf("Publish() error = %v, want %v", err, messagebus.ErrClosed)
		}
		if unsubscribe, err := bus.Subscribe("t", r.handle); unsubscribe != nil || !errors.Is(err, messagebus.ErrClosed) {
			t.Errorf("Subscribe() = %p, %v, want nil, %v", unsubscribe, err, messagebus.ErrClosed)
		}
		unsubscribe() // does nothing
		closeBus(t, bus)
		r.want(t, "subscriber")
	})
}

func TestCloseWakesPublisher(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		var r recorder[string]
		bus := fullQueue(t, &r, release)
		errc := publishInBackground(t, bus, "full", "late")

		closed := make(chan error)
		go func() { closed <- bus.Close(t.Context()) }()
		if err := <-errc; !errors.Is(err, messagebus.ErrClosed) {
			t.Errorf("Publish() error = %v, want %v", err, messagebus.ErrClosed)
		}

		close(release)
		if err := <-closed; err != nil {
			t.Errorf("Close() error = %v", err)
		}
		r.want(t, "subscriber", "taken", "queued")
	})
}

func TestCloseContextDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		var r recorder[string]
		bus := fullQueue(t, &r, release)

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		start := time.Now()
		err := bus.Close(ctx)

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Close() error = %v, want %v", err, context.DeadlineExceeded)
		}
		if want := "messagebus: close: context deadline exceeded"; err.Error() != want {
			t.Errorf("Close() error = %q, want %q", err, want)
		}
		if waited := time.Since(start); waited != time.Second {
			t.Errorf("Close() waited %v, want 1s", waited)
		}

		// The handlers carry on, and a later Close waits for them.
		close(release)
		closeBus(t, bus)
		r.want(t, "subscriber", "taken", "queued")
	})
}

func TestCloseDrainedContextDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](4)
		var r recorder[int]
		subscribe(t, bus, "t", r.handle)
		publish(t, bus, "t", 1)
		closeBus(t, bus)

		// Nothing is left to wait for, so a done ctx doesn't make Close fail.
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		for range 100 {
			if err := bus.Close(ctx); err != nil {
				t.Fatalf("Close() of a drained bus error = %v, want nil", err)
			}
		}
		r.want(t, "subscriber", 1)
	})
}

func TestHandlerGoexit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](4)
		var r recorder[int]
		subscribe(t, bus, "t", func(msg int) {
			r.handle(msg)
			if msg == 1 {
				runtime.Goexit()
			}
		})
		publish(t, bus, "t", 1, 2, 3)
		closeBus(t, bus)

		r.want(t, "subscriber", 1, 2, 3)
	})
}

func TestCloseFromHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](4)
		var got error
		subscribe(t, bus, "t", func(int) {
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			got = bus.Close(ctx) // waits for this handler, so only ctx ends it
		})
		publish(t, bus, "t", 1)
		closeBus(t, bus)

		if !errors.Is(got, context.DeadlineExceeded) {
			t.Errorf("Close() from a handler error = %v, want %v", got, context.DeadlineExceeded)
		}
	})
}

func TestConcurrentUse(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const publishers, messages, churn = 8, 200, 100
		topics := []string{"a", "b", "c", "d"}
		bus := messagebus.New[int](2)
		steady := make([]recorder[int], len(topics))
		for i, topic := range topics {
			subscribe(t, bus, topic, steady[i].handle)
		}

		var wg sync.WaitGroup
		for _, topic := range topics {
			wg.Go(func() {
				for range churn {
					unsubscribe, err := bus.Subscribe(topic, func(int) {})
					if err != nil {
						t.Errorf("Subscribe() error = %v", err)
						return
					}
					unsubscribe()
				}
			})
		}
		for p := range publishers {
			wg.Go(func() {
				for i := range messages {
					if err := bus.Publish(t.Context(), topics[(p+i)%len(topics)], i); err != nil {
						t.Errorf("Publish() error = %v", err)
					}
				}
			})
		}
		wg.Wait()
		closeBus(t, bus)

		// Subscribed before anything was published, so each gets every message on its topic.
		for i := range steady {
			if got, want := len(steady[i].messages()), publishers*messages/len(topics); got != want {
				t.Errorf("subscriber to %q got %d messages, want %d", topics[i], got, want)
			}
		}
	})
}

func TestCloseWhilePublishing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const publishers, before = 8, 100
		bus := messagebus.New[int](1)
		started := make(chan struct{})
		handled := 0
		subscribe(t, bus, "t", func(int) {
			if handled++; handled == before {
				close(started)
			}
		})

		var wg sync.WaitGroup
		for range publishers {
			wg.Go(func() {
				for {
					err := bus.Publish(t.Context(), "t", 1)
					if errors.Is(err, messagebus.ErrClosed) {
						return
					}
					if err != nil {
						t.Errorf("Publish() error = %v", err)
						return
					}
				}
			})
		}
		<-started
		closeBus(t, bus)
		wg.Wait()
	})
}

func TestNewNegativeQueuePanics(t *testing.T) {
	defer func() {
		if got, want := recover(), "messagebus: negative queue size"; got != want {
			t.Errorf("New(-1) panicked with %v, want %q", got, want)
		}
	}()
	messagebus.New[int](-1)
}

func TestSubscribeNilHandlerPanics(t *testing.T) {
	bus := messagebus.New[int](1)
	defer func() {
		if got, want := recover(), "messagebus: nil handler"; got != want {
			t.Errorf("Subscribe(nil) panicked with %v, want %q", got, want)
		}
	}()
	_, _ = bus.Subscribe("t", nil)
}
