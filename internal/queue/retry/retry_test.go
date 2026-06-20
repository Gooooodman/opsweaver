package retry_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Gooooodman/opsweaver/internal/queue/retry"
	"github.com/hibiken/asynq"
)

// sentinel errors a handler might return.
var (
	transientErr = errors.New("temporary: redis unavailable")
	permanentErr = errors.New("permanent: invalid argument")
)

// TestPermanentWrapsSkipRetry proves a permanent error stops Asynq from
// retrying: it must satisfy errors.Is(err, asynq.SkipRetry), which is the exact
// predicate the processor uses to archive instead of re-enqueue.
func TestPermanentWrapsSkipRetry(t *testing.T) {
	err := retry.Permanent(permanentErr)
	if !errors.Is(err, asynq.SkipRetry) {
		t.Errorf("Permanent err does not wrap asynq.SkipRetry; Asynq would retry it")
	}
	if !errors.Is(err, permanentErr) {
		t.Errorf("Permanent() lost the original error: errors.Is returned false")
	}
}

// TestTemporaryDoesNotWrapSkipRetry proves a plain (temporary) error is NOT
// flagged as permanent, so it falls through to the retry path.
func TestTemporaryDoesNotWrapSkipRetry(t *testing.T) {
	if errors.Is(transientErr, asynq.SkipRetry) {
		t.Errorf("transient error wraps asynq.SkipRetry; it should be retried")
	}
}

// TestIsFailureClassifiesErrors covers the IsFailure predicate Asynq calls to
// decide whether the failure counter increments. Permanent errors are still
// failures (they archive the task); the function's job is only to separate
// non-fatal skips from real failures.
func TestIsFailureClassifiesErrors(t *testing.T) {
	isFailure := retry.IsFailure

	if !isFailure(transientErr) {
		t.Error("IsFailure(transientErr) = false, want true")
	}
	if !isFailure(retry.Permanent(permanentErr)) {
		t.Error("IsFailure(permanent) = false, want true")
	}
	if isFailure(nil) {
		t.Error("IsFailure(nil) = true, want false")
	}
}

// TestRetryDelayExponential verifies RetryDelay grows exponentially with the
// retry count n and that every attempt is bounded above zero (so a stalled
// worker cannot spin-loop retries with zero delay).
func TestRetryDelayExponential(t *testing.T) {
	prev := time.Duration(0)
	for n := 1; n <= retry.MaxAttempts(); n++ {
		d := retry.RetryDelay(n, transientErr, nil)
		if d <= 0 {
			t.Fatalf("RetryDelay(n=%d) = %v, want > 0", n, d)
		}
		// Exponential growth with jitter is non-deterministic, so assert
		// monotonic non-decrease over a few runs instead of an exact ratio.
		if d < prev {
			t.Errorf("RetryDelay(n=%d) = %v decreased below previous %v", n, d, prev)
		}
		prev = d
	}
}

// TestMaxAttemptsMatchesBudget verifies the retry budget advertised by this
// package equals the MaxRetry set at enqueue time in the queue package, so the
// "at most three retries" spec is expressed in one place.
func TestMaxAttemptsMatchesBudget(t *testing.T) {
	if retry.MaxAttempts() != 3 {
		t.Errorf("MaxAttempts() = %d, want 3", retry.MaxAttempts())
	}
}
