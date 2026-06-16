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
	"time"
)

const defaultTimeout = 2 * time.Second

type CheckFunc func(ctx context.Context) error

// Options configures a Checker. All fields are optional.
type Options struct {
	Timeout          time.Duration
	RecordDependency func(name string, up bool)
}

// Checker registers named dependency probes and serves liveness and readiness
// HTTP endpoints. Register is not safe for concurrent use and is expected to
// be called only at process startup. The handlers returned by LiveHandler and
// ReadyHandler are safe for concurrent requests.
type Checker struct {
	timeout time.Duration
	names   []string
	checks  map[string]CheckFunc
	record  func(name string, up bool)
}

func New(opts Options) *Checker {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Checker{
		timeout: timeout,
		checks:  make(map[string]CheckFunc),
		record:  opts.RecordDependency,
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
		resultCh := make(chan checkResult, len(c.names))
		for i, name := range c.names {
			results[i] = checkResult{Name: name, Status: "down"}
			go func(n string, fn CheckFunc) {
				status := "ok"
				if err := fn(ctx); err != nil {
					status = "down"
				}
				resultCh <- checkResult{Name: n, Status: status}
			}(name, c.checks[name])
		}

		byName := make(map[string]int, len(results))
		for i, result := range results {
			byName[result.Name] = i
		}
		for range c.names {
			select {
			case result := <-resultCh:
				results[byName[result.Name]] = result
			case <-ctx.Done():
				goto ready
			}
		}

	ready:
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
		for _, result := range results {
			if c.record != nil {
				c.record(result.Name, result.Status == "ok")
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
