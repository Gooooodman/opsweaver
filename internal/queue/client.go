package queue

import (
	"context"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
)

// ErrClosed is returned by Enqueue* methods after Close has been called.
var ErrClosed = errors.New("queue: client closed")

// Client is the producer side of the async queue. It wraps *asynq.Client so
// callers enqueue typed payloads (DiagnosisRun, MCPSyncTools) without touching
// asynq.Task directly. The underlying connection is owned by this Client and
// must be released with Close.
type Client struct {
	client *asynq.Client
}

// NewClient returns a Client backed by the given Redis connection option. The
// option should be produced by redisx.AsynqConnOpts so it points at DB 0.
func NewClient(connOpt asynq.RedisConnOpt) (*Client, error) {
	if connOpt == nil {
		return nil, errors.New("queue: redis connection option is required")
	}
	return &Client{client: asynq.NewClient(connOpt)}, nil
}

// EnqueueDiagnosisRun enqueues a TypeDiagnosisRun task with the given payload.
// The returned TaskInfo carries the asynq-assigned id and state.
func (c *Client) EnqueueDiagnosisRun(ctx context.Context, p DiagnosisRunPayload) (*asynq.TaskInfo, error) {
	return c.enqueue(ctx, TypeDiagnosisRun, p)
}

// EnqueueMCPSyncTools enqueues a TypeMCPSyncTools task with the given payload.
func (c *Client) EnqueueMCPSyncTools(ctx context.Context, p MCPSyncToolsPayload) (*asynq.TaskInfo, error) {
	return c.enqueue(ctx, TypeMCPSyncTools, p)
}

func (c *Client) enqueue(ctx context.Context, typeName string, payload any) (*asynq.TaskInfo, error) {
	if c.client == nil {
		return nil, ErrClosed
	}
	raw, err := encode(payload)
	if err != nil {
		return nil, err
	}
	// MaxRetry is applied here so every enqueued task honours the controlled
	// retry budget from the retry subpackage without each caller repeating it.
	info, err := c.client.EnqueueContext(ctx, asynq.NewTask(typeName, raw), asynq.MaxRetry(maxRetries))
	if err != nil {
		return nil, fmt.Errorf("queue: enqueue %s: %w", typeName, err)
	}
	return info, nil
}

// Close releases the underlying asynq client connection. After Close returns,
// further Enqueue* calls return ErrClosed.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("queue: close client: %w", err)
	}
	c.client = nil
	return nil
}
