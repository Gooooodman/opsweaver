package httpx_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/apperror"
	"github.com/Gooooodman/opsweaver/internal/platform/httpx"
)

func TestTraceID_ContextRoundTrip(t *testing.T) {
	ctx := httpx.WithTraceID(context.Background(), "abc123")
	if got := httpx.TraceIDFrom(ctx); got != "abc123" {
		t.Errorf("TraceIDFrom = %q, want %q", got, "abc123")
	}

	if got := httpx.TraceIDFrom(context.Background()); got != "" {
		t.Errorf("TraceIDFrom(empty ctx) = %q, want \"\"", got)
	}
}

func TestNewTraceID_Unique(t *testing.T) {
	hexRe := regexp.MustCompile(`^[0-9a-f]{32}$`)
	seen := make(map[string]struct{}, 256)
	for i := 0; i < 256; i++ {
		id := httpx.NewTraceID()
		if !hexRe.MatchString(id) {
			t.Fatalf("NewTraceID() = %q, want 32 lowercase hex chars", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate trace id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestWriteError_AppError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(httpx.WithTraceID(req.Context(), "abc"))

	err := apperror.New(apperror.CodeNotFound, "user not found")
	httpx.WriteError(rec, req, err)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := rec.Header().Get("X-Trace-ID"); got != "abc" {
		t.Errorf("X-Trace-ID = %q, want abc", got)
	}

	var body httpx.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != apperror.CodeNotFound {
		t.Errorf("body.Code = %q, want %q", body.Code, apperror.CodeNotFound)
	}
	if body.Message != "user not found" {
		t.Errorf("body.Message = %q, want %q", body.Message, "user not found")
	}
	if body.TraceID != "abc" {
		t.Errorf("body.TraceID = %q, want abc", body.TraceID)
	}
}

func TestWriteError_GenericError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	httpx.WriteError(rec, req, errors.New("db down"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	bodyBytes := rec.Body.Bytes()
	if string(bodyBytes) == "" {
		t.Fatal("empty body")
	}
	if containsSubstr(string(bodyBytes), "db down") {
		t.Errorf("body leaks underlying error: %s", bodyBytes)
	}

	var body httpx.ErrorResponse
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != apperror.CodeInternal {
		t.Errorf("body.Code = %q, want %q", body.Code, apperror.CodeInternal)
	}
	if body.Message != "internal server error" {
		t.Errorf("body.Message = %q, want %q", body.Message, "internal server error")
	}

	hexRe := regexp.MustCompile(`^[0-9a-f]{32}$`)
	if !hexRe.MatchString(body.TraceID) {
		t.Errorf("body.TraceID = %q, want 32 hex chars", body.TraceID)
	}
	if got := rec.Header().Get("X-Trace-ID"); got != body.TraceID {
		t.Errorf("X-Trace-ID header = %q, body.TraceID = %q; must match", got, body.TraceID)
	}
}

func TestWriteJSON_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(httpx.WithTraceID(req.Context(), "trace-xyz"))

	payload := map[string]any{"hello": "world", "n": float64(42)}
	httpx.WriteJSON(rec, req, http.StatusOK, payload)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := rec.Header().Get("X-Trace-ID"); got != "trace-xyz" {
		t.Errorf("X-Trace-ID = %q, want trace-xyz", got)
	}

	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out["hello"] != "world" {
		t.Errorf("body[hello] = %v, want world", out["hello"])
	}
	if out["n"] != float64(42) {
		t.Errorf("body[n] = %v, want 42", out["n"])
	}
}

func TestWriteError_PreservesTraceIDFromContext(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(httpx.WithTraceID(req.Context(), "preset-trace"))

	httpx.WriteError(rec, req, apperror.New(apperror.CodeConflict, "conflict"))

	var body httpx.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.TraceID != "preset-trace" {
		t.Errorf("body.TraceID = %q, want preset-trace", body.TraceID)
	}
	if got := rec.Header().Get("X-Trace-ID"); got != "preset-trace" {
		t.Errorf("X-Trace-ID = %q, want preset-trace", got)
	}
}

func containsSubstr(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
