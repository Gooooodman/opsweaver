// Package redisx wires Redis connection options for OpsWeaver services.
//
// Redis is shared between two logical tenants in P0 (design.md decision 4):
// Asynq uses DB 0 for the task queue, and the cache/state layer uses DB 1.
// This package is the single place that converts a config.RedisConfig into an
// asynq.RedisConnOpt while enforcing that reserved DB assignment, so the queue
// and cache can never accidentally collide on the same logical database.
package redisx

import (
	"errors"

	"github.com/Gooooodman/opsweaver/internal/platform/config"
	"github.com/hibiken/asynq"
)

// Reserved logical DB numbers. These mirror design.md decision 4 and the
// validation rules in config.validate(); they are constants rather than config
// because the assignment is a deployment invariant.
const (
	asynqQueueDB = 0
	cacheDB      = 1
)

// ErrDBMismatch is returned when a RedisConfig's DB does not match the reserved
// assignment for the tenant it is being wired into. Wrap-checked so callers can
// distinguish a misconfigured DB from other failures.
var ErrDBMismatch = errors.New("redisx: redis DB does not match reserved assignment")

// AsynqConnOpts builds the asynq Redis connection option for the task queue.
// The supplied config must use DB 0; any other value fails closed to protect
// the queue/cache isolation invariant.
func AsynqConnOpts(cfg config.RedisConfig) (asynq.RedisConnOpt, error) {
	if cfg.Addr == "" {
		return nil, errors.New("redisx: asynq addr is required")
	}
	if cfg.DB != asynqQueueDB {
		return nil, ErrDBMismatch
	}
	return asynq.RedisClientOpt{Addr: cfg.Addr, DB: cfg.DB}, nil
}

// CacheConnOpts builds the asynq Redis connection option for the cache/state
// layer. The supplied config must use DB 1; any other value fails closed.
func CacheConnOpts(cfg config.RedisConfig) (asynq.RedisConnOpt, error) {
	if cfg.Addr == "" {
		return nil, errors.New("redisx: cache addr is required")
	}
	if cfg.DB != cacheDB {
		return nil, ErrDBMismatch
	}
	return asynq.RedisClientOpt{Addr: cfg.Addr, DB: cfg.DB}, nil
}
