package queue_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/queue"
	"github.com/hibiken/asynq"
)

// TestTaskTypeConstants guards the producer/consumer contract: the type strings
// must not drift, because the server registers handlers and the client enqueues
// against them independently.
func TestTaskTypeConstants(t *testing.T) {
	if queue.TypeDiagnosisRun != "diagnosis:run" {
		t.Errorf("TypeDiagnosisRun = %q, want %q", queue.TypeDiagnosisRun, "diagnosis:run")
	}
	if queue.TypeMCPSyncTools != "mcp:sync_tools" {
		t.Errorf("TypeMCPSyncTools = %q, want %q", queue.TypeMCPSyncTools, "mcp:sync_tools")
	}
	if queue.QueueDefault != "default" {
		t.Errorf("QueueDefault = %q, want %q", queue.QueueDefault, "default")
	}
}

// TestPayloadRoundTrip verifies that the payload types JSON-encode and decode
// symmetrically, since the client marshals and the server's ServeMux adapter
// unmarshals them at opposite ends of the wire.
func TestPayloadRoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		payload any
		target  func() any
	}{
		{
			name:    "diagnosis run",
			payload: queue.DiagnosisRunPayload{TaskID: "diag-42"},
			target:  func() any { return &queue.DiagnosisRunPayload{} },
		},
		{
			name:    "mcp sync tools",
			payload: queue.MCPSyncToolsPayload{MCPServerID: "mcp-7"},
			target:  func() any { return &queue.MCPSyncToolsPayload{} },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			dst := tc.target()
			if err := json.Unmarshal(raw, dst); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			// Round-trip equality is what the producer/consumer rely on; assert
			// via re-marshalling both sides to a canonical JSON form.
			want, _ := json.Marshal(tc.payload)
			got, _ := json.Marshal(dst)
			if string(got) != string(want) {
				t.Errorf("round-trip mismatch: got %s, want %s", got, want)
			}
		})
	}
}

// TestServerHandlerDispatch verifies the typed handler registration decodes the
// payload and routes to the correct handler by task type. It drives the
// ServeMux directly with a constructed task, avoiding any Redis dependency.
func TestServerHandlerDispatch(t *testing.T) {
	srv, err := queue.NewServer(asynq.RedisClientOpt{Addr: "unused"}, queue.DefaultServerConfig())
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	var gotDiag queue.DiagnosisRunPayload
	var gotMCP queue.MCPSyncToolsPayload

	srv.HandleDiagnosisRun(func(ctx context.Context, p queue.DiagnosisRunPayload) error {
		gotDiag = p
		return nil
	})
	srv.HandleMCPSyncTools(func(ctx context.Context, p queue.MCPSyncToolsPayload) error {
		gotMCP = p
		return nil
	})

	wantDiag := queue.DiagnosisRunPayload{TaskID: "diag-99"}
	wantMCP := queue.MCPSyncToolsPayload{MCPServerID: "mcp-3"}

	diagRaw, _ := json.Marshal(wantDiag)
	mcpRaw, _ := json.Marshal(wantMCP)

	if err := srv.DispatchForTest(context.Background(), queue.TypeDiagnosisRun, diagRaw); err != nil {
		t.Fatalf("dispatch diagnosis: %v", err)
	}
	if gotDiag != wantDiag {
		t.Errorf("diagnosis handler got %+v, want %+v", gotDiag, wantDiag)
	}

	if err := srv.DispatchForTest(context.Background(), queue.TypeMCPSyncTools, mcpRaw); err != nil {
		t.Fatalf("dispatch mcp: %v", err)
	}
	if gotMCP != wantMCP {
		t.Errorf("mcp handler got %+v, want %+v", gotMCP, wantMCP)
	}
}

// TestServerHandlerPropagatesError ensures a handler error is surfaced rather
// than swallowed by the ServeMux adapter.
func TestServerHandlerPropagatesError(t *testing.T) {
	srv, err := queue.NewServer(asynq.RedisClientOpt{Addr: "unused"}, queue.DefaultServerConfig())
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	wantErr := errors.New("boom")
	srv.HandleDiagnosisRun(func(ctx context.Context, p queue.DiagnosisRunPayload) error {
		return wantErr
	})

	raw, _ := json.Marshal(queue.DiagnosisRunPayload{TaskID: "diag-1"})
	if err := srv.DispatchForTest(context.Background(), queue.TypeDiagnosisRun, raw); !errors.Is(err, wantErr) {
		t.Fatalf("dispatch error = %v, want %v", err, wantErr)
	}
}
