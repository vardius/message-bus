package messagebus_test

import (
	"fmt"
	"sync"

	"github.com/vardius/message-bus"
)

func Example() {
	bus := messagebus.New()

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

	bus.Publish("topic", true)
	wg.Wait()
	// Unordered output
	// s1 true
	// s2 true
}
