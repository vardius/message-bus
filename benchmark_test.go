package messagebus

import "testing"

func BenchmarkWorker(b *testing.B) {
	bus := New()

	bus.Subscribe("topic", func(v bool) {})
	bus.Subscribe("topic", func(v bool) {})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish("topic", true)
		}
	})
}
