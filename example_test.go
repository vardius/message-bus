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

	_ = bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		fmt.Println("s1", v)
	})

	_ = bus.Subscribe("topic", func(v bool) {
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

func Example_second() {
	queueSize := 2
	subscribersAmount := 3

	ch := make(chan int, queueSize)
	defer close(ch)

	bus := messagebus.New(queueSize)

	for i := 0; i < subscribersAmount; i++ {
		_ = bus.Subscribe("topic", func(i int, out chan<- int) { out <- i })
	}

	go func() {
		for n := 0; n < queueSize; n++ {
			bus.Publish("topic", n, ch)
		}
	}()

	var sum = 0
	for sum < (subscribersAmount * queueSize) {
		select {
		case <-ch:
			sum++
		}
	}

	fmt.Println(sum)
	// Output:
	// 6
}
