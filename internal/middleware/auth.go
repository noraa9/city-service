package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"city-service/internal/domain"
	"city-service/internal/repository"
	jwtpkg "city-service/pkg/jwt"

	"github.com/google/uuid"
)

// We use a private type for context keys to avoid collisions with other packages.
type ctxKey string

const userCtxKey ctxKey = "user"

// UserFromContext returns the authenticated user (if middleware put it there).
func UserFromContext(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(userCtxKey).(domain.User)
	return u, ok
}

type AuthMiddleware struct {
	users     repository.UserRepository
	jwtSecret string
}

func NewAuthMiddleware(users repository.UserRepository, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{users: users, jwtSecret: jwtSecret}
}

// Authenticate does:
// 1) Read "Authorization: Bearer <token>"
// 2) Parse & validate JWT
// 3) Load user from DB by user_id in claims
// 4) Put user into request context
// 5) Call next handler
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			respondError(w, http.StatusUnauthorized, "missing Authorization header")
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respondError(w, http.StatusUnauthorized, "invalid Authorization header")
			return
		}

		claims, err := jwtpkg.ParseToken(parts[1], m.jwtSecret)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token user_id")
			return
		}

		u, err := m.users.GetByID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				respondError(w, http.StatusUnauthorized, "user not found")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to load user")
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
