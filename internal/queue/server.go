package queue

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gooooodman/opsweaver/internal/queue/retry"
	"github.com/hibiken/asynq"
)

// Controlled-retry budget shared by the producer (Enqueue MaxRetry) and the
// consumer (retry subpackage IsFailure / RetryDelayFunc). Per async-task-queue
// spec: at most three retries with exponential backoff, permanent errors
// excluded from the budget.
const maxRetries = 3

// DefaultServerConfig returns the asynq server config that opsweaver-worker
// uses. It encodes the controlled-retry policy (exponential backoff, permanent
// errors skipped) at the queue level via retry.ApplyPolicy.
func DefaultServerConfig() asynq.Config {
	cfg := asynq.Config{
		Queues: map[string]int{
			QueueDefault: 1,
		},
	}
	retry.ApplyPolicy(&cfg)
	return cfg
}

// DiagnosisHandler handles TypeDiagnosisRun tasks.
type DiagnosisHandler func(ctx context.Context, p DiagnosisRunPayload) error

// MCPHandler handles TypeMCPSyncTools tasks.
type MCPHandler func(ctx context.Context, p MCPSyncToolsPayload) error

// Server is the consumer side of the async queue. It wraps *asynq.Server and an
// asynq.ServeMux, exposing typed handler registration so handlers receive
// decoded payloads instead of raw task bytes.
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewServer returns a Server backed by the given Redis connection option and
// the provided asynq Config. The option should be produced by
// redisx.AsynqConnOpts so it points at DB 0.
func NewServer(connOpt asynq.RedisConnOpt, cfg asynq.Config) (*Server, error) {
	if connOpt == nil {
		return nil, errors.New("queue: redis connection option is required")
	}
	return &Server{
		server: asynq.NewServer(connOpt, cfg),
		mux:    asynq.NewServeMux(),
	}, nil
}

// HandleDiagnosisRun registers a handler for TypeDiagnosisRun tasks.
func (s *Server) HandleDiagnosisRun(h DiagnosisHandler) {
	s.mux.HandleFunc(TypeDiagnosisRun, func(ctx context.Context, t *asynq.Task) error {
		var p DiagnosisRunPayload
		if err := decode(t.Payload(), &p); err != nil {
			return err
		}
		return h(ctx, p)
	})
}

// HandleMCPSyncTools registers a handler for TypeMCPSyncTools tasks.
func (s *Server) HandleMCPSyncTools(h MCPHandler) {
	s.mux.HandleFunc(TypeMCPSyncTools, func(ctx context.Context, t *asynq.Task) error {
		var p MCPSyncToolsPayload
		if err := decode(t.Payload(), &p); err != nil {
			return err
		}
		return h(ctx, p)
	})
}

// Start begins processing registered task types and returns after the worker
// goroutines start. Callers must keep the process alive and invoke Shutdown
// during graceful termination.
func (s *Server) Start() error {
	if err := s.server.Start(s.mux); err != nil {
		return fmt.Errorf("queue: start server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server, draining in-flight tasks.
func (s *Server) Shutdown() {
	s.server.Shutdown()
}

// DispatchForTest routes a synthesised task with the given type and payload
// through the registered handlers, without starting the asynq server or
// touching Redis. It exists to keep the producer/consumer payload contract
// unit-testable in isolation.
func (s *Server) DispatchForTest(ctx context.Context, typeName string, payload []byte) error {
	return s.mux.ProcessTask(ctx, asynq.NewTask(typeName, payload))
}
