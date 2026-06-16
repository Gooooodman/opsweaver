// Package servicehttp runs HTTP servers with bounded graceful shutdown.
package servicehttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrListen = errors.New("http server listen failed")

func Serve(ctx context.Context, srv *http.Server, shutdownTimeout time.Duration) error {
	serverErr := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serverErr <- err
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("%w: %w", ErrListen, err)
		}
		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown failed: %w", err)
	}
	if err := <-serverErr; err != nil {
		return fmt.Errorf("%w: %w", ErrListen, err)
	}
	return nil
}
