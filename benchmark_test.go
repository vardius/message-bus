package messagebus

import (
	"runtime"
	"testing"
)

func addSubscribers(bus MessageBus, max int) {
	for i := 0; i < max; i++ {
		bus.Subscribe("topic", func(v bool) {})
	}
}

func run(b *testing.B, bus MessageBus) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bus.Publish("topic", true)
	}
}

func runParallel(b *testing.B, bus MessageBus) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish("topic", true)
		}
	})
}

func BenchmarkWorkerNumCPU(b *testing.B) {
	bus := New(runtime.NumCPU())
	addSubscribers(bus, runtime.NumCPU())

	run(b, bus)
}

func BenchmarkWorkerNumCPUParallel(b *testing.B) {
	bus := New(runtime.NumCPU())
	addSubscribers(bus, runtime.NumCPU())

	runParallel(b, bus)
}

func BenchmarkWorker(b *testing.B) {
	bus := New(100)
	addSubscribers(bus, 100)

	run(b, bus)
}

func BenchmarkWorkerParallel(b *testing.B) {
	bus := New(100)
	addSubscribers(bus, 100)

	runParallel(b, bus)
}

func BenchmarkWorker100(b *testing.B) {
	bus := New(100)
	addSubscribers(bus, 100)

	run(b, bus)
}

func BenchmarkWorker100Parallel(b *testing.B) {
	bus := New(100)
	addSubscribers(bus, 100)

	runParallel(b, bus)
}
