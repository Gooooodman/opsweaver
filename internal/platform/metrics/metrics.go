// Package metrics provides per-service Prometheus collectors for HTTP
// requests, dependency status, and Asynq task processing. Each Metrics
// instance owns a caller-supplied registry; re-registering the same
// collectors on the same registry returns a wrapped AlreadyRegisteredError.
package metrics

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Options configures a Metrics instance. Both fields are required.
type Options struct {
	Namespace string
	Service   string
}

// Metrics bundles the Prometheus collectors exposed by a single service.
type Metrics struct {
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	DependencyUp         *prometheus.GaugeVec
	AsynqProcessedTotal  *prometheus.CounterVec
	AsynqProcessDuration *prometheus.HistogramVec

	registry prometheus.Registerer
	gatherer prometheus.Gatherer
}

// New constructs a Metrics, registers every collector on reg, and attaches
// a constant service=<opts.Service> label to each.
func New(opts Options, reg *prometheus.Registry) (*Metrics, error) {
	if opts.Namespace == "" {
		return nil, errors.New("metrics: namespace is required")
	}
	if opts.Service == "" {
		return nil, errors.New("metrics: service is required")
	}
	if reg == nil {
		return nil, errors.New("metrics: registry is required")
	}

	constLabels := prometheus.Labels{"service": opts.Service}

	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   opts.Namespace,
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests processed, partitioned by method, route, and status code.",
			ConstLabels: constLabels,
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   opts.Namespace,
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request latency in seconds, partitioned by method and route.",
			ConstLabels: constLabels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	dependencyUp := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   opts.Namespace,
			Name:        "dependency_up",
			Help:        "Current readiness status of a named dependency (1 = up, 0 = down).",
			ConstLabels: constLabels,
		},
		[]string{"dependency"},
	)

	asynqProcessedTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   opts.Namespace,
			Name:        "asynq_processed_total",
			Help:        "Total number of Asynq tasks processed, partitioned by queue, task type, and outcome.",
			ConstLabels: constLabels,
		},
		[]string{"queue", "task_type", "status"},
	)

	asynqProcessDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   opts.Namespace,
			Name:        "asynq_process_duration_seconds",
			Help:        "Asynq task processing latency in seconds, partitioned by queue and task type.",
			ConstLabels: constLabels,
			Buckets:     prometheus.ExponentialBuckets(0.01, 2, 12),
		},
		[]string{"queue", "task_type"},
	)

	collectors := []struct {
		name string
		c    prometheus.Collector
	}{
		{"http_requests_total", httpRequestsTotal},
		{"http_request_duration_seconds", httpRequestDuration},
		{"dependency_up", dependencyUp},
		{"asynq_processed_total", asynqProcessedTotal},
		{"asynq_process_duration_seconds", asynqProcessDuration},
	}

	var registered []prometheus.Collector
	for _, item := range collectors {
		if err := reg.Register(item.c); err != nil {
			for _, prev := range registered {
				reg.Unregister(prev)
			}
			return nil, fmt.Errorf("metrics: register %s: %w", item.name, err)
		}
		registered = append(registered, item.c)
	}

	return &Metrics{
		HTTPRequestsTotal:    httpRequestsTotal,
		HTTPRequestDuration:  httpRequestDuration,
		DependencyUp:         dependencyUp,
		AsynqProcessedTotal:  asynqProcessedTotal,
		AsynqProcessDuration: asynqProcessDuration,
		registry:             reg,
		gatherer:             reg,
	}, nil
}

// Handler returns the Prometheus text-format exposition handler bound to
// this Metrics instance's gatherer.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.gatherer, promhttp.HandlerOpts{})
}
