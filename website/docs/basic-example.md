---
id: basic-example
title: Basic example
sidebar_label: Basic example
---

Publish block only when the buffer of one of the subscribers is full.
Change the buffer size altering queueSize when creating new message bus.

```go
package main

import (
    "fmt"
    "sync"

    "github.com/vardius/message-bus"
)

func main() {
    queueSize := 100
    bus := messagebus.New(queueSize)

    var wg sync.WaitGroup
    wg.Add(2)

    _ = bus.Subscribe("topic", func(v bool) {
        defer wg.Done()
        fmt.Println(v)
    })

    _ = bus.Subscribe("topic", func(v bool) {
        defer wg.Done()
        fmt.Println(v)
    })

    bus.Publish("topic", true)
    wg.Wait()
}
```
