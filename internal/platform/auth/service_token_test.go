package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/apperror"
	"github.com/Gooooodman/opsweaver/internal/platform/httpx"
)

const testServiceTokenHeader = "X-OpsWeaver-Service-Token"

func TestServiceTokenMiddleware_AllowsValidServiceToken(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/internal", nil)
	req.Header.Set(testServiceTokenHeader, "expected-token")
	rec := httptest.NewRecorder()

	ServiceTokenMiddleware("expected-token", next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestServiceTokenMiddleware_RejectsMissingOrInvalidServiceToken(t *testing.T) {
	cases := []struct {
		name  string
		token string
	}{
		{name: "missing token", token: ""},
		{name: "wrong token", token: "wrong-token"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/internal", nil)
			if c.token != "" {
				req.Header.Set(testServiceTokenHeader, c.token)
			}
			rec := httptest.NewRecorder()

			ServiceTokenMiddleware("expected-token", next).ServeHTTP(rec, req)

			if called {
				t.Fatal("next handler was called")
			}
			assertUnauthorizedServiceTokenResponse(t, rec)
		})
	}
}

func TestServiceTokenMiddleware_RejectsWhenExpectedTokenIsEmpty(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/internal", nil)
	req.Header.Set(testServiceTokenHeader, "provided-token")
	rec := httptest.NewRecorder()

	ServiceTokenMiddleware("", next).ServeHTTP(rec, req)

	if called {
		t.Fatal("next handler was called")
	}
	assertUnauthorizedServiceTokenResponse(t, rec)
}

func assertUnauthorizedServiceTokenResponse(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var body httpx.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != apperror.CodeUnauthenticated {
		t.Errorf("body.Code = %q, want %q", body.Code, apperror.CodeUnauthenticated)
	}
	if body.Message != "invalid service token" {
		t.Errorf("body.Message = %q, want %q", body.Message, "invalid service token")
	}
	for _, leaked := range []string{"expected-token", "wrong-token", "provided-token"} {
		if strings.Contains(rec.Body.String(), leaked) {
			t.Errorf("body leaks token %q: %s", leaked, rec.Body.String())
		}
	}
}
