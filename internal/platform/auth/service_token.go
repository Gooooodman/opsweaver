package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/Gooooodman/opsweaver/internal/platform/apperror"
	"github.com/Gooooodman/opsweaver/internal/platform/httpx"
)

const ServiceTokenHeader = "X-OpsWeaver-Service-Token"

func ServiceTokenMiddleware(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validServiceToken(expected, r.Header.Get(ServiceTokenHeader)) {
			httpx.WriteError(w, r, apperror.New(apperror.CodeUnauthenticated, "invalid service token"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validServiceToken(expected, provided string) bool {
	if expected == "" {
		return false
	}

	expectedHash := sha256.Sum256([]byte(expected))
	providedHash := sha256.Sum256([]byte(provided))
	return subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) == 1
}
