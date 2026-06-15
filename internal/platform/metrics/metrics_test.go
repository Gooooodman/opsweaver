package metrics_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNew_RegistersAllCollectors(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, err := metrics.New(metrics.Options{Namespace: "opsweaver", Service: "opsweaver-server"}, reg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil Metrics")
	}
	if m.HTTPRequestsTotal == nil {
		t.Error("HTTPRequestsTotal is nil")
	}
	if m.HTTPRequestDuration == nil {
		t.Error("HTTPRequestDuration is nil")
	}
	if m.DependencyUp == nil {
		t.Error("DependencyUp is nil")
	}
	if m.AsynqProcessedTotal == nil {
		t.Error("AsynqProcessedTotal is nil")
	}
	if m.AsynqProcessDuration == nil {
		t.Error("AsynqProcessDuration is nil")
	}

	m.HTTPRequestsTotal.WithLabelValues("GET", "/v1/ping", "200").Inc()
	m.HTTPRequestDuration.WithLabelValues("GET", "/v1/ping").Observe(0.1)
	m.DependencyUp.WithLabelValues("postgres").Set(1)
	m.AsynqProcessedTotal.WithLabelValues("default", "send_email", "success").Inc()
	m.AsynqProcessDuration.WithLabelValues("default", "send_email").Observe(0.5)

	body := renderHandler(t, m.Handler())
	expected := []string{
		"opsweaver_http_requests_total",
		"opsweaver_http_request_duration_seconds",
		"opsweaver_dependency_up",
		"opsweaver_asynq_processed_total",
		"opsweaver_asynq_process_duration_seconds",
	}
	for _, name := range expected {
		if !strings.Contains(body, name) {
			t.Errorf("rendered metrics missing %q\nbody:\n%s", name, body)
		}
	}
}

func TestNew_DuplicateRegistrationReturnsError(t *testing.T) {
	reg := prometheus.NewRegistry()
	if _, err := metrics.New(metrics.Options{Namespace: "opsweaver", Service: "opsweaver-server"}, reg); err != nil {
		t.Fatalf("first New returned error: %v", err)
	}
	_, err := metrics.New(metrics.Options{Namespace: "opsweaver", Service: "opsweaver-server"}, reg)
	if err == nil {
		t.Fatal("expected error on duplicate registration, got nil")
	}
	var alreadyRegistered prometheus.AlreadyRegisteredError
	if !errors.As(err, &alreadyRegistered) {
		t.Errorf("expected error to wrap prometheus.AlreadyRegisteredError, got %T: %v", err, err)
	}
}

func TestNew_RequiresNamespaceAndService(t *testing.T) {
	cases := []struct {
		name   string
		opts   metrics.Options
		expect string
	}{
		{
			name:   "missing namespace",
			opts:   metrics.Options{Namespace: "", Service: "opsweaver-server"},
			expect: "namespace",
		},
		{
			name:   "missing service",
			opts:   metrics.Options{Namespace: "opsweaver", Service: ""},
			expect: "service",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := metrics.New(tc.opts, prometheus.NewRegistry())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.expect) {
				t.Errorf("expected error mentioning %q, got %q", tc.expect, err.Error())
			}
		})
	}
}

func TestHandler_ReturnsPrometheusContentType(t *testing.T) {
	m, err := metrics.New(metrics.Options{Namespace: "opsweaver", Service: "opsweaver-server"}, prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type starting with text/plain, got %q", ct)
	}
}

func TestHandler_ExposesConstantServiceLabel(t *testing.T) {
	m, err := metrics.New(metrics.Options{Namespace: "opsweaver", Service: "opsweaver-server"}, prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	m.HTTPRequestsTotal.WithLabelValues("GET", "/v1/ping", "200").Inc()

	body := renderHandler(t, m.Handler())
	if !strings.Contains(body, `service="opsweaver-server"`) {
		t.Errorf("rendered metrics missing constant label service=\"opsweaver-server\"\nbody:\n%s", body)
	}
}

func TestNew_SeparateRegistriesAreIndependent(t *testing.T) {
	regA := prometheus.NewRegistry()
	regB := prometheus.NewRegistry()

	mA, err := metrics.New(metrics.Options{Namespace: "opsweaver", Service: "opsweaver-server"}, regA)
	if err != nil {
		t.Fatalf("New A returned error: %v", err)
	}
	mB, err := metrics.New(metrics.Options{Namespace: "opsweaver", Service: "opsweaver-worker"}, regB)
	if err != nil {
		t.Fatalf("New B returned error: %v", err)
	}
	if mA == nil || mB == nil {
		t.Fatal("expected both Metrics instances non-nil")
	}

	mA.HTTPRequestsTotal.WithLabelValues("GET", "/a", "200").Inc()
	mB.HTTPRequestsTotal.WithLabelValues("GET", "/b", "200").Inc()

	bodyA := renderHandler(t, mA.Handler())
	bodyB := renderHandler(t, mB.Handler())
	if !strings.Contains(bodyA, `service="opsweaver-server"`) {
		t.Errorf("registry A missing service=opsweaver-server label\nbody:\n%s", bodyA)
	}
	if !strings.Contains(bodyB, `service="opsweaver-worker"`) {
		t.Errorf("registry B missing service=opsweaver-worker label\nbody:\n%s", bodyB)
	}
	if strings.Contains(bodyA, `service="opsweaver-worker"`) {
		t.Errorf("registry A unexpectedly contains worker label\nbody:\n%s", bodyA)
	}
	if strings.Contains(bodyB, `service="opsweaver-server"`) {
		t.Errorf("registry B unexpectedly contains server label\nbody:\n%s", bodyB)
	}
}

func renderHandler(t *testing.T, h http.Handler) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	return string(buf)
}
