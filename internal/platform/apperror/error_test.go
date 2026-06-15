package apperror_test

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/apperror"
)

func TestNew_SetsFields(t *testing.T) {
	e := apperror.New(apperror.CodeInvalidArgument, "missing field")
	if e == nil {
		t.Fatal("expected non-nil error")
	}
	if e.Code != apperror.CodeInvalidArgument {
		t.Errorf("Code = %q, want %q", e.Code, apperror.CodeInvalidArgument)
	}
	if e.Message != "missing field" {
		t.Errorf("Message = %q, want %q", e.Message, "missing field")
	}
	if e.Err != nil {
		t.Errorf("Err = %v, want nil", e.Err)
	}
}

func TestWrap_PreservesCause(t *testing.T) {
	e := apperror.Wrap(io.EOF, apperror.CodeInternal, "read failed")
	if e.Code != apperror.CodeInternal {
		t.Errorf("Code = %q, want %q", e.Code, apperror.CodeInternal)
	}
	if e.Message != "read failed" {
		t.Errorf("Message = %q, want %q", e.Message, "read failed")
	}
	if got := e.Unwrap(); got != io.EOF {
		t.Errorf("Unwrap = %v, want io.EOF", got)
	}
	if !errors.Is(e, io.EOF) {
		t.Error("errors.Is should return true for wrapped io.EOF")
	}
}

func TestError_String(t *testing.T) {
	e := apperror.New(apperror.CodeInvalidArgument, "missing field")
	if got := e.Error(); got != "invalid_argument: missing field" {
		t.Errorf("Error() = %q, want %q", got, "invalid_argument: missing field")
	}

	wrapped := apperror.Wrap(errors.New("db down"), apperror.CodeInternal, "internal failure")
	got := wrapped.Error()
	if got != "internal: internal failure" {
		t.Errorf("Error() = %q, want %q", got, "internal: internal failure")
	}
	for _, leak := range []string{"db down"} {
		if containsSubstr(got, leak) {
			t.Errorf("Error() = %q must not leak cause %q", got, leak)
		}
	}
}

func TestFrom_RecognizesAppError(t *testing.T) {
	orig := apperror.Wrap(io.EOF, apperror.CodeInternal, "boom")
	got, ok := apperror.From(orig)
	if !ok {
		t.Fatal("From(*apperror.Error) ok = false, want true")
	}
	if got != orig {
		t.Errorf("From returned %v, want %v", got, orig)
	}

	if _, ok := apperror.From(io.EOF); ok {
		t.Error("From(io.EOF) ok = true, want false")
	}

	if _, ok := apperror.From(nil); ok {
		t.Error("From(nil) ok = true, want false")
	}
}

func TestHTTPStatus_KnownCodes(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{apperror.CodeInvalidArgument, http.StatusBadRequest},
		{apperror.CodeUnauthenticated, http.StatusUnauthorized},
		{apperror.CodePermissionDenied, http.StatusForbidden},
		{apperror.CodeNotFound, http.StatusNotFound},
		{apperror.CodeConflict, http.StatusConflict},
		{apperror.CodeInternal, http.StatusInternalServerError},
		{apperror.CodeUnavailable, http.StatusServiceUnavailable},
		{"foo", http.StatusInternalServerError},
		{"", http.StatusInternalServerError},
	}
	for _, c := range cases {
		if got := apperror.HTTPStatus(c.code); got != c.want {
			t.Errorf("HTTPStatus(%q) = %d, want %d", c.code, got, c.want)
		}
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
