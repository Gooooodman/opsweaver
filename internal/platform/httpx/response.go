package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/Gooooodman/opsweaver/internal/platform/apperror"
)

type ContextKey string

const TraceIDContextKey ContextKey = "trace_id"

const traceIDHeader = "X-Trace-ID"

const contentTypeJSON = "application/json"

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDContextKey, traceID)
}

func TraceIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(TraceIDContextKey).(string)
	return v
}

func NewTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(b[:])
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	traceID := TraceIDFrom(r.Context())
	if traceID == "" {
		traceID = NewTraceID()
	}

	code := apperror.CodeInternal
	message := "internal server error"
	if appErr, ok := apperror.From(err); ok {
		if appErr.Code != "" {
			code = appErr.Code
		}
		if appErr.Message != "" {
			message = appErr.Message
		}
	}

	status := apperror.HTTPStatus(code)
	body := ErrorResponse{Code: code, Message: message, TraceID: traceID}

	w.Header().Set(traceIDHeader, traceID)
	payload, mErr := json.Marshal(body)
	if mErr != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
		return
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	traceID := TraceIDFrom(r.Context())
	if traceID == "" {
		traceID = NewTraceID()
	}
	w.Header().Set(traceIDHeader, traceID)

	data, err := json.Marshal(payload)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
		return
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
