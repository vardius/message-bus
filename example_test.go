package messagebus_test

import (
	"fmt"
	"sync"

	messagebus "github.com/vardius/message-bus"
)

func Example() {
	queueSize := 100
	bus := messagebus.New(queueSize)

	var wg sync.WaitGroup
	wg.Add(2)

	bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		fmt.Println("s1", v)
	})

	bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		fmt.Println("s2", v)
	})

	// Publish block only when the buffer of one of the subscribers is full.
	// change the buffer size altering queueSize when creating new messagebus
	bus.Publish("topic", true)
	wg.Wait()

	// Unordered output:
	// s1 true
	// s2 true
}
