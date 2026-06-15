package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/logging"
)

const testService = "opsweaver-server"

func decodeLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatalf("expected log output, got empty buffer")
	}
	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("failed to decode log line %q: %v", line, err)
	}
	return record
}

func TestNew_WritesJSONWithServiceField(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(logging.Options{Service: testService, Writer: &buf})

	logger.Info("hello", "user", "alice")

	record := decodeLine(t, &buf)
	if _, ok := record["time"]; !ok {
		t.Errorf("missing time field, got %v", record)
	}
	if got, want := record["level"], "INFO"; got != want {
		t.Errorf("level = %v, want %v", got, want)
	}
	if got, want := record["msg"], "hello"; got != want {
		t.Errorf("msg = %v, want %v", got, want)
	}
	if got, want := record["service"], testService; got != want {
		t.Errorf("service = %v, want %v", got, want)
	}
	if got, want := record["user"], "alice"; got != want {
		t.Errorf("user = %v, want %v", got, want)
	}
}

func TestNew_MasksSensitiveTopLevelKeys(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{"token", "token", "tok-secret-123"},
		{"password", "password", "p@ssw0rd!"},
		{"authorization mixed case", "Authorization", "Bearer abc.def.ghi"},
		{"secret", "secret", "shh-do-not-tell"},
		{"api_key", "api_key", "ak-live-xyz"},
		{"apikey", "apikey", "ak-live-uvw"},
		{"dsn", "dsn", "postgres://user:pass@host/db"},
		{"master_key", "master_key", "mk-raw-value"},
		{"master_key_base64", "master_key_base64", "bWFzdGVy"},
		{"internal_service_token", "internal_service_token", "ist-hush"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.New(logging.Options{Service: testService, Writer: &buf})

			logger.Info("sensitive", tc.key, tc.value)

			raw := buf.String()
			if strings.Contains(raw, tc.value) {
				t.Fatalf("output should not contain raw secret %q, got %q", tc.value, raw)
			}

			record := decodeLine(t, &buf)
			got, ok := record[tc.key]
			if !ok {
				t.Fatalf("missing key %q in record %v", tc.key, record)
			}
			if got != "***" {
				t.Errorf("record[%q] = %v, want \"***\"", tc.key, got)
			}
		})
	}
}

func TestNew_MasksSensitiveKeysInGroup(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(logging.Options{Service: testService, Writer: &buf})

	const tokenValue = "tok-123"
	logger.WithGroup("auth").Info("login", "token", tokenValue)

	raw := buf.String()
	if strings.Contains(raw, tokenValue) {
		t.Fatalf("output should not contain raw token %q, got %q", tokenValue, raw)
	}

	record := decodeLine(t, &buf)
	auth, ok := record["auth"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth group to be an object, got %v (type %T)", record["auth"], record["auth"])
	}
	if got := auth["token"]; got != "***" {
		t.Errorf("auth.token = %v, want \"***\"", got)
	}
}

func TestNew_KeepsNonSensitiveFields(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(logging.Options{Service: testService, Writer: &buf})

	logger.Info("request done",
		"request_id", "req-42",
		"user_id", "user-7",
		"duration_ms", 125,
		"error", "connection reset",
	)

	record := decodeLine(t, &buf)
	if got, want := record["request_id"], "req-42"; got != want {
		t.Errorf("request_id = %v, want %v", got, want)
	}
	if got, want := record["user_id"], "user-7"; got != want {
		t.Errorf("user_id = %v, want %v", got, want)
	}
	if got, want := record["duration_ms"], float64(125); got != want {
		t.Errorf("duration_ms = %v, want %v", got, want)
	}
	if got, want := record["error"], "connection reset"; got != want {
		t.Errorf("error = %v, want %v", got, want)
	}
}

func TestNew_DefaultsToStderrWhenWriterNil(t *testing.T) {
	logger := logging.New(logging.Options{Service: testService})
	if logger == nil {
		t.Fatal("expected non-nil logger when Writer is nil")
	}
	t.Log("stderr default path exercised; stderr output not captured to avoid noise")
}

func TestNew_RespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(logging.Options{
		Service: testService,
		Writer:  &buf,
		Level:   slog.LevelWarn,
	})

	logger.Info("should not appear")
	if buf.Len() != 0 {
		t.Fatalf("expected no output at Info level, got %q", buf.String())
	}

	logger.Warn("warning here")
	record := decodeLine(t, &buf)
	if got, want := record["level"], "WARN"; got != want {
		t.Errorf("level = %v, want %v", got, want)
	}
	if got, want := record["msg"], "warning here"; got != want {
		t.Errorf("msg = %v, want %v", got, want)
	}
}
