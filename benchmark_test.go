package messagebus

import (
	"runtime"
	"testing"
)

func run(b *testing.B, bus MessageBus, out chan<- int) {
	for n := 0; n < b.N; n++ {
		bus.Publish("topic", n, out)
	}
}

func runParallel(b *testing.B, bus MessageBus, out chan<- int) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish("topic", 1, out)
		}
	})
}

func runBenchmark(b *testing.B, subscribersAmount int, runInParallel bool) {
	ch := make(chan int, b.N)
	defer close(ch)

	bus := New(b.N)

	for i := 0; i < subscribersAmount; i++ {
		bus.Subscribe("topic", func(i int, out chan<- int) { out <- i })
	}

	b.ResetTimer()

	go func() {
		if runInParallel {
			runParallel(b, bus, ch)
		} else {
			run(b, bus, ch)
		}
	}()

	var i = 0
	for i < (b.N * subscribersAmount) {
		select {
		case <-ch:
			i++
		}
	}
}

func BenchmarkBus(b *testing.B) {
	runBenchmark(b, 1, false)
}

func BenchmarkBusParallel(b *testing.B) {
	runBenchmark(b, 1, true)
}

func BenchmarkBus100(b *testing.B) {
	runBenchmark(b, 100, false)
}

func BenchmarkBus100Parallel(b *testing.B) {
	runBenchmark(b, 100, true)
}

func BenchmarkBus1000(b *testing.B) {
	runBenchmark(b, 1000, false)
}

func BenchmarkBus1000Parallel(b *testing.B) {
	runBenchmark(b, 1000, true)
}

func BenchmarkBusNumCPU(b *testing.B) {
	runBenchmark(b, runtime.NumCPU()+1, false)
}

func BenchmarkBusNumCPUParallel(b *testing.B) {
	runBenchmark(b, runtime.NumCPU()+1, true)
}
