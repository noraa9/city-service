package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims is what we store inside the JWT.
//
// We embed jwt.RegisteredClaims to get standard fields like:
// - ExpiresAt
// - IssuedAt
// - Subject, etc.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwtlib.RegisteredClaims
}

// GenerateToken creates a signed JWT string for a user.
//
// Design decision:
// - We keep token creation in a small package so services/middleware don't duplicate it.
func GenerateToken(userID, role, secret string, expiry time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ParseToken validates a token and returns its claims.
//
// What we validate:
// - signature uses our secret
// - signing method is HMAC (HS256)
// - token is not expired (we check ExpiresAt explicitly to keep the logic easy to learn)
func ParseToken(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}

	parsed, err := jwtlib.ParseWithClaims(tokenString, claims, func(t *jwtlib.Token) (any, error) {
		// Reject tokens that use a different signing algorithm.
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	},
		// Extra safety: only accept HS256.
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Minimal expiry validation (the library can do more, but this is clear for beginners).
	if claims.ExpiresAt == nil || time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

