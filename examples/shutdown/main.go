// Command shutdown serves HTTP and publishes an event for every signup. On Ctrl-C it stops taking requests,
// then waits for the events already published to be handled.
//
//	go run ./examples/shutdown
//	curl -d name=ana localhost:8080/signup
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	messagebus "github.com/vardius/message-bus/v2"
)

type SignedUp struct{ Name string }

// signup publishes a SignedUp event and answers 202 Accepted: the welcome email goes out later.
func signup(bus *messagebus.Bus[SignedUp]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		// If the queue is full and the client gives up, so does Publish.
		if err := bus.Publish(r.Context(), "signups", SignedUp{Name: name}); err != nil {
			http.Error(w, "try again later", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// shutdown stops the server first, so nothing publishes any more, then closes the bus,
// which waits for the handlers to finish the queued events. Both share ctx's deadline.
func shutdown(ctx context.Context, srv *http.Server, bus *messagebus.Bus[SignedUp]) error {
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("stop the server: %w", err)
	}
	if err := bus.Close(ctx); err != nil {
		return fmt.Errorf("handle queued events: %w", err)
	}
	return nil
}

func main() {
	bus := messagebus.New[SignedUp](100)
	if _, err := bus.Subscribe("signups", func(e SignedUp) {
		log.Printf("welcome email sent to %s", e.Name)
	}); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("POST /signup", signup(bus))
	srv := &http.Server{Addr: "localhost:8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()

	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := shutdown(timeout, srv, bus); err != nil {
		log.Fatal(err)
	}
}
