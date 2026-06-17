package redisx_test

import (
	"errors"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/config"
	"github.com/Gooooodman/opsweaver/internal/platform/redisx"
	"github.com/hibiken/asynq"
)

// TestAsynqConnOptsUsesDB0 verifies the Asynq connection option points at the
// Redis DB reserved for the queue (DB 0), per design.md decision 4. The factory
// must reject any config that deviates from the reserved assignment so the
// queue and cache never collide.
func TestAsynqConnOptsUsesDB0(t *testing.T) {
	opts, err := redisx.AsynqConnOpts(config.RedisConfig{
		Addr: "localhost:6379",
		DB:   0,
	})
	if err != nil {
		t.Fatalf("AsynqConnOpts() error = %v", err)
	}

	redisOpts, ok := opts.(asynq.RedisClientOpt)
	if !ok {
		t.Fatalf("AsynqConnOpts() = %T, want asynq.RedisClientOpt", opts)
	}
	if redisOpts.Addr != "localhost:6379" {
		t.Errorf("Addr = %q, want localhost:6379", redisOpts.Addr)
	}
	if redisOpts.DB != 0 {
		t.Errorf("DB = %d, want 0 (Asynq queue DB)", redisOpts.DB)
	}
}

// TestCacheConnOptsUsesDB1 verifies the cache connection option points at the
// Redis DB reserved for cache/state (DB 1), per design.md decision 4.
func TestCacheConnOptsUsesDB1(t *testing.T) {
	opts, err := redisx.CacheConnOpts(config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1,
	})
	if err != nil {
		t.Fatalf("CacheConnOpts() error = %v", err)
	}

	redisOpts, ok := opts.(asynq.RedisClientOpt)
	if !ok {
		t.Fatalf("CacheConnOpts() = %T, want asynq.RedisClientOpt", opts)
	}
	if redisOpts.Addr != "localhost:6379" {
		t.Errorf("Addr = %q, want localhost:6379", redisOpts.Addr)
	}
	if redisOpts.DB != 1 {
		t.Errorf("DB = %d, want 1 (cache DB)", redisOpts.DB)
	}
}

// TestAsynqConnOptsRejectsNonZeroDB ensures the queue factory fails closed
// when handed a config whose DB is not the reserved 0. The DB assignment is a
// deployment invariant, not a runtime tunable.
func TestAsynqConnOptsRejectsNonZeroDB(t *testing.T) {
	_, err := redisx.AsynqConnOpts(config.RedisConfig{
		Addr: "localhost:6379",
		DB:   2,
	})
	if err == nil {
		t.Fatal("AsynqConnOpts() error = nil, want DB assignment error")
	}
	if !errors.Is(err, redisx.ErrDBMismatch) {
		t.Errorf("AsynqConnOpts() error = %v, want ErrDBMismatch", err)
	}
}

// TestCacheConnOptsRejectsNonOneDB ensures the cache factory fails closed when
// handed a config whose DB is not the reserved 1.
func TestCacheConnOptsRejectsNonOneDB(t *testing.T) {
	_, err := redisx.CacheConnOpts(config.RedisConfig{
		Addr: "localhost:6379",
		DB:   0,
	})
	if err == nil {
		t.Fatal("CacheConnOpts() error = nil, want DB assignment error")
	}
	if !errors.Is(err, redisx.ErrDBMismatch) {
		t.Errorf("CacheConnOpts() error = %v, want ErrDBMismatch", err)
	}
}

// TestConnOptsRejectsEmptyAddr ensures neither factory silently produces a
// connection option that would dial an empty address.
func TestConnOptsRejectsEmptyAddr(t *testing.T) {
	if _, err := redisx.AsynqConnOpts(config.RedisConfig{DB: 0}); err == nil {
		t.Error("AsynqConnOpts(empty addr) error = nil, want error")
	}
	if _, err := redisx.CacheConnOpts(config.RedisConfig{DB: 1}); err == nil {
		t.Error("CacheConnOpts(empty addr) error = nil, want error")
	}
}
