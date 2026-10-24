package messagebus_test

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"

	messagebus "github.com/vardius/message-bus/v2"
)

// Handlers do nothing, so what is measured is the bus's own cost: queueing, waking the handlers, and waiting
// for room when they fall behind. The queues hold 1024 messages.

func BenchmarkPublish(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			bus := subscribed(b, n, 1024)
			ctx := context.Background()
			b.ReportAllocs()
			msg := 0
			for b.Loop() {
				msg++
				_ = bus.Publish(ctx, "t", msg)
			}
		})
	}
}

// onTopics returns a bus with one subscriber on each of n topics, and the topics' names.
func onTopics(b *testing.B, n int) (*messagebus.Bus[int], []string) {
	bus := subscribed(b, 0, 1024)
	names := make([]string, n)
	for i := range names {
		names[i] = "topic " + strconv.Itoa(i)
		if _, err := bus.Subscribe(names[i], noop); err != nil {
			b.Fatal(err)
		}
	}
	return bus, names
}

func BenchmarkPublishTopics(b *testing.B) {
	bus, names := onTopics(b, 100)
	ctx := context.Background()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_ = bus.Publish(ctx, names[i%len(names)], i)
		i++
	}
}

// With no buffer, Publish hands each message straight to the handler, waiting whenever it is busy.
func BenchmarkPublishUnbuffered(b *testing.B) {
	bus := messagebus.New[int](0)
	if _, err := bus.Subscribe("t", noop); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = bus.Close(context.Background()) })
	ctx := context.Background()
	b.ReportAllocs()
	msg := 0
	for b.Loop() {
		msg++
		_ = bus.Publish(ctx, "t", msg)
	}
}

func BenchmarkPublishParallel(b *testing.B) {
	bus := subscribed(b, 1, 1024)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		msg := 0
		for pb.Next() {
			msg++
			_ = bus.Publish(ctx, "t", msg)
		}
	})
}

func BenchmarkPublishParallelTopics(b *testing.B) {
	bus, names := onTopics(b, 100)
	ctx := context.Background()
	var seed atomic.Int64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := int(seed.Add(13)) // each goroutine starts on a different topic
		for pb.Next() {
			_ = bus.Publish(ctx, names[i%len(names)], i)
			i++
		}
	})
}

func BenchmarkSubscribe(b *testing.B) {
	bus := subscribed(b, 0, 1024)
	b.ReportAllocs()
	for b.Loop() {
		unsubscribe, _ := bus.Subscribe("t", noop)
		unsubscribe()
	}
}
