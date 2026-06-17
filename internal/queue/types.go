// Package queue defines the async task contracts shared between
// opsweaver-server (producer) and opsweaver-worker (consumer).
//
// Only the task type strings and their payload shapes are owned here; the
// retry policy lives in the retry subpackage and the Redis/Asynq wiring lives
// in the client and server files. Keeping the type contracts in one place lets
// the producer and consumer evolve together without leaking Asynq details into
// business code.
package queue

import (
	"encoding/json"
	"fmt"
)

// Task type strings. These are the keys registered against the asynq ServeMux
// and the values passed to asynq.NewTask, so producer and consumer must agree
// on them exactly.
const (
	// TypeDiagnosisRun is the async task that runs a diagnosis workflow
	// (deliver-pod-diagnosis-workflow change). Payload: DiagnosisRunPayload.
	TypeDiagnosisRun = "diagnosis:run"

	// TypeMCPSyncTools is the async task that pulls the remote tool list from a
	// registered MCP server and reconciles it into the local ToolSpec registry
	// (add-declarative-capability-runtime change). Payload: MCPSyncToolsPayload.
	TypeMCPSyncTools = "mcp:sync_tools"
)

// Queue names. P0 keeps a single dedicated queue so diagnosis and MCP sync
// compete for the same worker pool; priority is left to a future change.
const (
	// QueueDefault is the only queue processed by opsweaver-worker in P0.
	QueueDefault = "default"
)

// DiagnosisRunPayload is the payload for TypeDiagnosisRun.
type DiagnosisRunPayload struct {
	// TaskID is the business-level diagnosis task identifier persisted in
	// opsweaver_server_db. The worker uses it to claim execution rights and to
	// write status back via the server's internal API.
	TaskID string `json:"task_id"`
}

// MCPSyncToolsPayload is the payload for TypeMCPSyncTools.
type MCPSyncToolsPayload struct {
	// MCPServerID is the registry row id of the MCP server whose tool list is
	// being reconciled.
	MCPServerID string `json:"mcp_server_id"`
}

// encode marshals a payload into the JSON bytes asynq.NewTask expects. It is
// unexported because callers go through the typed Client.Enqueue* methods.
func encode(payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("queue: encode payload: %w", err)
	}
	return raw, nil
}

// decode unmarshals JSON bytes into the destination. Handlers receive raw task
// bytes and use this to recover the typed payload.
func decode(raw []byte, dst any) error {
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("queue: decode payload: %w", err)
	}
	return nil
}
