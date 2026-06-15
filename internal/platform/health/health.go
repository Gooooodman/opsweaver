// Package health provides liveness and readiness HTTP handlers backed by a
// set of named dependency probes. Probes run in parallel under a per-check
// timeout and only their names and ok/down status are exposed to clients —
// raw probe errors are intentionally hidden to avoid leaking DSN or secret
// fragments.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

const defaultTimeout = 2 * time.Second

type CheckFunc func(ctx context.Context) error

// Options configures a Checker. All fields are optional.
type Options struct {
	Timeout time.Duration
}

// Checker registers named dependency probes and serves liveness and readiness
// HTTP endpoints. Register is not safe for concurrent use and is expected to
// be called only at process startup. The handlers returned by LiveHandler and
// ReadyHandler are safe for concurrent requests.
type Checker struct {
	timeout time.Duration
	names   []string
	checks  map[string]CheckFunc
}

func New(opts Options) *Checker {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Checker{
		timeout: timeout,
		checks:  make(map[string]CheckFunc),
	}
}

func (c *Checker) Register(name string, fn CheckFunc) {
	if _, exists := c.checks[name]; !exists {
		c.names = append(c.names, name)
	}
	c.checks[name] = fn
}

func (c *Checker) LiveHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

type checkResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type readyPayload struct {
	Status string        `json:"status"`
	Checks []checkResult `json:"checks"`
}

func (c *Checker) ReadyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), c.timeout)
		defer cancel()

		results := make([]checkResult, len(c.names))
		var wg sync.WaitGroup
		for i, name := range c.names {
			wg.Add(1)
			go func(idx int, n string, fn CheckFunc) {
				defer wg.Done()
				status := "ok"
				if err := fn(ctx); err != nil {
					status = "down"
				}
				results[idx] = checkResult{Name: n, Status: status}
			}(i, name, c.checks[name])
		}
		wg.Wait()

		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})

		overall := "ok"
		code := http.StatusOK
		for _, r := range results {
			if r.Status != "ok" {
				overall = "degraded"
				code = http.StatusServiceUnavailable
				break
			}
		}

		writeJSON(w, code, readyPayload{Status: overall, Checks: results})
	})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
