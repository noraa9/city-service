package middleware

import (
	"net/http"
)

// RequireRole is a middleware factory.
//
// Usage:
//   r.Use(authMiddleware.Authenticate)
//   r.Use(middleware.RequireRole("monitor"))
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			if _, ok := allowed[u.Role]; !ok {
				respondError(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

