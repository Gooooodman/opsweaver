// Package retry implements the controlled-retry policy from the
// async-task-queue spec: temporary errors retry with exponential backoff up to
// a fixed budget (3 attempts), and permanent errors are marked as non-retriable
// so the task is archived immediately.
//
// The policy is expressed as plain functions and an asynq.RetryDelayFunc so it
// can be wired into an asynq.Config without the queue package owning retry
// internals.
package retry

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/hibiken/asynq"
)

// maxAttempts is the absolute retry budget. It must match the MaxRetry set at
// enqueue time in the parent queue package; keeping it here lets tests assert
// the budget in one place.
const maxAttempts = 3

// MaxAttempts returns the configured retry budget (maximum retries per task).
func MaxAttempts() int { return maxAttempts }

// permanentError wraps an underlying error and signals to Asynq that the task
// must not be retried. Asynq archives any error that satisfies
// errors.Is(err, asynq.SkipRetry), so embedding SkipRetry here opts the wrapped
// error into that path.
type permanentError struct {
	cause error
}

func (e *permanentError) Error() string {
	return fmt.Sprintf("queue: permanent error: %v", e.cause)
}

func (e *permanentError) Unwrap() []error {
	return []error{asynq.SkipRetry, e.cause}
}

// Permanent marks err as a permanent failure (invalid argument, policy denial,
// etc.). Handlers return retry.Permanent(err) to make Asynq archive the task
// instead of re-enqueuing it. A nil err is reported as-is to surface a bug.
func Permanent(err error) error {
	if err == nil {
		return errors.New("queue: Permanent called with nil error")
	}
	return &permanentError{cause: err}
}

// IsPermanent reports whether err is a permanent error produced by Permanent.
func IsPermanent(err error) bool {
	var pe *permanentError
	return errors.As(err, &pe)
}

// Cause extracts the underlying error from a Permanent wrapper, returning err
// unchanged if it is not wrapped.
func Cause(err error) error {
	var pe *permanentError
	if errors.As(err, &pe) {
		return pe.cause
	}
	return err
}

// IsFailure is the asynq IsFailure predicate. A nil error is not a failure;
// everything else (temporary or permanent) is, because the task either retries
// or is archived. The predicate exists so future non-fatal skips can be added
// in one place.
func IsFailure(err error) bool {
	return err != nil
}

// RetryDelay is the asynq.RetryDelayFunc for the queue. It uses exponential
// backoff with jitter: base * 2^(n-1) plus up to 25% jitter, capped so the P0
// ceiling stays well under a minute per attempt. n is the retry count (1-based
// on first retry) provided by Asynq.
func RetryDelay(n int, _ error, _ *asynq.Task) time.Duration {
	if n < 1 {
		n = 1
	}
	const (
		base   = 2 * time.Second
		capDur = 30 * time.Second
	)
	// 2^(n-1) * base, capped to avoid runaway delays at the top of the budget.
	backoff := time.Duration(math.Pow(2, float64(n-1))) * base
	if backoff > capDur {
		backoff = capDur
	}
	// Jitter of up to 25% of the backoff spreads retries across workers and
	// prevents thundering-herd reattempts against a recovering dependency.
	jitter := time.Duration(rand.Int64N(int64(backoff) / 4))
	d := backoff + jitter
	if d <= 0 {
		d = base
	}
	return d
}

// ApplyPolicy wires the retry policy into an asynq.Config. It is the single
// integration point the queue.Server uses so retry details stay in this package.
func ApplyPolicy(cfg *asynq.Config) {
	cfg.RetryDelayFunc = RetryDelay
	cfg.IsFailure = IsFailure
}
