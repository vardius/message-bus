package message_bus

import (
	"bytes"
	"testing"
)

func BenchmarkWorker(b *testing.B) {
	bus := New()

	bus.Subscribe("topic", func(v bool) {})
	bus.Subscribe("topic", func(v bool) {})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		for pb.Next() {
			buf.Reset()
			bus.Publish("topic", true)
		}
	})
}
