package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"

	messagebus "github.com/vardius/message-bus/v2"
)

func TestSignup(t *testing.T) {
	tests := []struct {
		name string
		form string
		want int
	}{
		{"with a name", "name=ana", http.StatusAccepted},
		{"without a name", "", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := messagebus.New[SignedUp](1)
			defer bus.Close(t.Context())
			if _, err := bus.Subscribe("signups", func(SignedUp) {}); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(tt.form))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			signup(bus)(w, r)

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestSignupAfterClose(t *testing.T) {
	bus := messagebus.New[SignedUp](1)
	if err := bus.Close(t.Context()); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	signup(bus)(w, httptest.NewRequest(http.MethodPost, "/signup?name=ana", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestShutdown(t *testing.T) {
	bus := messagebus.New[SignedUp](100)
	var mu sync.Mutex
	var welcomed []string
	if _, err := bus.Subscribe("signups", func(e SignedUp) {
		mu.Lock()
		defer mu.Unlock()
		welcomed = append(welcomed, e.Name)
	}); err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: signup(bus)}
	go srv.Serve(ln)

	for _, name := range []string{"ana", "ben", "cy"} {
		resp, err := http.PostForm("http://"+ln.Addr().String()+"/signup", url.Values{"name": {name}})
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	if err := shutdown(t.Context(), srv, bus); err != nil {
		t.Fatalf("shutdown() error = %v", err)
	}
	// Every event published before shutdown was handled by the time it returned.
	if want := []string{"ana", "ben", "cy"}; !slices.Equal(welcomed, want) {
		t.Errorf("welcomed %v, want %v", welcomed, want)
	}
}
