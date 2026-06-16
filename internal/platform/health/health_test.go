package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Gooooodman/opsweaver/internal/platform/health"
)

type readyResponse struct {
	Status string `json:"status"`
	Checks []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"checks"`
}

func readBody(t *testing.T, body io.ReadCloser) string {
	t.Helper()
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

func TestLiveHandler_AlwaysReturnsOK(t *testing.T) {
	c := health.New(health.Options{})

	srv := httptest.NewServer(c.LiveHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content-type: got %q, want application/json prefix", ct)
	}
	body := readBody(t, resp.Body)
	if !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("body: got %q, want contain status ok", body)
	}
}

func TestReadyHandler_AllDependenciesUp(t *testing.T) {
	c := health.New(health.Options{})
	c.Register("postgres", func(ctx context.Context) error { return nil })
	c.Register("redis", func(ctx context.Context) error { return nil })

	srv := httptest.NewServer(c.ReadyHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}

	var parsed readyResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()

	if parsed.Status != "ok" {
		t.Fatalf("overall status: got %q, want ok", parsed.Status)
	}
	if len(parsed.Checks) != 2 {
		t.Fatalf("checks count: got %d, want 2", len(parsed.Checks))
	}

	seen := map[string]string{}
	for _, ck := range parsed.Checks {
		seen[ck.Name] = ck.Status
	}
	if seen["postgres"] != "ok" {
		t.Fatalf("postgres status: got %q, want ok", seen["postgres"])
	}
	if seen["redis"] != "ok" {
		t.Fatalf("redis status: got %q, want ok", seen["redis"])
	}
}

func TestReadyHandler_OneDependencyDown(t *testing.T) {
	c := health.New(health.Options{})
	c.Register("postgres", func(ctx context.Context) error {
		return errors.New("connection refused: postgres://user:pw@host/db")
	})
	c.Register("redis", func(ctx context.Context) error { return nil })

	srv := httptest.NewServer(c.ReadyHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d, want 503", resp.StatusCode)
	}

	body := readBody(t, resp.Body)

	if !strings.Contains(body, `"status":"degraded"`) {
		t.Fatalf("body: missing overall degraded status: %s", body)
	}
	if !strings.Contains(body, "postgres") {
		t.Fatalf("body: must contain the failing checker name: %s", body)
	}
	if strings.Contains(body, "pw") {
		t.Fatalf("body leaks credential fragment 'pw': %s", body)
	}
	if strings.Contains(body, "postgres://") {
		t.Fatalf("body leaks raw DSN substring 'postgres://': %s", body)
	}
}

func TestReadyHandler_Timeout(t *testing.T) {
	c := health.New(health.Options{Timeout: 50 * time.Millisecond})
	c.Register("slow", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	srv := httptest.NewServer(c.ReadyHandler())
	defer srv.Close()

	start := time.Now()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	elapsed := time.Since(start)
	resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d, want 503", resp.StatusCode)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("elapsed too long: got %v, want < 500ms (probe should respect timeout)", elapsed)
	}
}

func TestReadyHandler_TimeoutWhenProbeIgnoresContext(t *testing.T) {
	c := health.New(health.Options{Timeout: 50 * time.Millisecond})
	c.Register("stuck", func(ctx context.Context) error {
		select {}
	})

	srv := httptest.NewServer(c.ReadyHandler())
	defer srv.Close()

	client := http.Client{Timeout: time.Second}
	start := time.Now()
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	elapsed := time.Since(start)
	body := readBody(t, resp.Body)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d, want 503", resp.StatusCode)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("elapsed too long: got %v, want < 500ms (handler must enforce timeout)", elapsed)
	}
	if !strings.Contains(body, `"name":"stuck"`) || !strings.Contains(body, `"status":"down"`) {
		t.Fatalf("body: got %s, want stuck dependency marked down", body)
	}
}

func TestReadyHandler_ReportsDependencyStatus(t *testing.T) {
	seen := map[string]bool{}
	c := health.New(health.Options{
		RecordDependency: func(name string, up bool) {
			seen[name] = up
		},
	})
	c.Register("postgres", func(ctx context.Context) error { return nil })
	c.Register("redis", func(ctx context.Context) error { return errors.New("down") })

	srv := httptest.NewServer(c.ReadyHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	resp.Body.Close()

	if seen["postgres"] != true {
		t.Fatalf("postgres metric status: got %v, want true", seen["postgres"])
	}
	if seen["redis"] != false {
		t.Fatalf("redis metric status: got %v, want false", seen["redis"])
	}
}

func TestReadyHandler_Parallel(t *testing.T) {
	c := health.New(health.Options{Timeout: time.Second})
	c.Register("a", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	c.Register("b", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	srv := httptest.NewServer(c.ReadyHandler())
	defer srv.Close()

	start := time.Now()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	elapsed := time.Since(start)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	if elapsed >= 180*time.Millisecond {
		t.Fatalf("elapsed: got %v, want < 180ms (probes must run in parallel)", elapsed)
	}
}

func TestRegister_Replaces(t *testing.T) {
	c := health.New(health.Options{})
	c.Register("dep", func(ctx context.Context) error { return nil })
	c.Register("dep", func(ctx context.Context) error { return errors.New("boom") })

	srv := httptest.NewServer(c.ReadyHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	body := readBody(t, resp.Body)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d, want 503 (second registration must win)", resp.StatusCode)
	}
	if !strings.Contains(body, `"status":"degraded"`) {
		t.Fatalf("body: want degraded after replacement, got %s", body)
	}
}
