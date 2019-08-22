package messagebus

import (
	"strconv"
	"testing"
)

func benchmark(b *testing.B, subscribersCount, topicsCount int) {
	ch := make(chan int, b.N)
	defer close(ch)

	handlerQueueSize := b.N
	bus := New(handlerQueueSize)

	for i := 0; i < topicsCount; i++ {
		for j := 0; j < subscribersCount; j++ {
			bus.Subscribe(strconv.Itoa(i), func(value int, out chan<- int) {
				out <- value
			})
		}
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish(strconv.Itoa(topicsCount-1), 1, ch)
		}
	})

	var i = 0
	for i < b.N*subscribersCount {
		select {
		case <-ch:
			i++
		}
	}
}

func BenchmarkOneSubscriberPerOneTopic(b *testing.B) {
	benchmark(b, 1, 1)
}

func BenchmarkOneSubscriberPerHundredTopics(b *testing.B) {
	benchmark(b, 1, 100)
}

func BenchmarkHundredSubscribersPerOneTopic(b *testing.B) {
	benchmark(b, 100, 1)
}

func BenchmarkHundredSubscribersPerHundredTopics(b *testing.B) {
	benchmark(b, 100, 100)
}
