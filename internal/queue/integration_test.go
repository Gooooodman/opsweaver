package queue_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Gooooodman/opsweaver/internal/platform/config"
	"github.com/Gooooodman/opsweaver/internal/platform/redisx"
	"github.com/Gooooodman/opsweaver/internal/queue"
)

func TestEnqueueAndConsume(t *testing.T) {
	addr := os.Getenv("OPSWEAVER_INTEGRATION_REDIS_ADDR")
	if addr == "" {
		t.Skip("OPSWEAVER_INTEGRATION_REDIS_ADDR is not set")
	}

	connOpt, err := redisx.AsynqConnOpts(config.RedisConfig{Addr: addr, DB: 0})
	if err != nil {
		t.Fatalf("AsynqConnOpts() error = %v", err)
	}

	server, err := queue.NewServer(connOpt, queue.DefaultServerConfig())
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	consumed := make(chan queue.DiagnosisRunPayload, 1)
	server.HandleDiagnosisRun(func(_ context.Context, payload queue.DiagnosisRunPayload) error {
		consumed <- payload
		return nil
	})
	if err := server.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(server.Shutdown)

	client, err := queue.NewClient(connOpt)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	want := queue.DiagnosisRunPayload{
		TaskID: fmt.Sprintf("integration-%d", time.Now().UnixNano()),
	}
	if _, err := client.EnqueueDiagnosisRun(context.Background(), want); err != nil {
		t.Fatalf("EnqueueDiagnosisRun() error = %v", err)
	}

	select {
	case got := <-consumed:
		if got != want {
			t.Fatalf("consumed payload = %+v, want %+v", got, want)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for queued task")
	}
}
