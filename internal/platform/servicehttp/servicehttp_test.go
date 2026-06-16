package servicehttp_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Gooooodman/opsweaver/internal/platform/servicehttp"
)

func TestServeReturnsListenError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen reserved port: %v", err)
	}
	addr := ln.Addr().String()
	defer ln.Close()

	srv := &http.Server{
		Addr:              addr,
		Handler:           http.NewServeMux(),
		ReadHeaderTimeout: time.Second,
	}

	err = servicehttp.Serve(context.Background(), srv, time.Second)
	if err == nil {
		t.Fatal("Serve() error = nil, want listen error")
	}
	if !errors.Is(err, servicehttp.ErrListen) {
		t.Fatalf("Serve() error = %v, want ErrListen", err)
	}
}

func TestServeShutsDownOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           http.NewServeMux(),
		ReadHeaderTimeout: time.Second,
	}

	done := make(chan error, 1)
	go func() {
		done <- servicehttp.Serve(ctx, srv, time.Second)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Serve() did not return after context cancel")
	}
}
