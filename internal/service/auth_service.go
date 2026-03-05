package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"city-service/internal/domain"
	"city-service/internal/repository"
	jwtpkg "city-service/pkg/jwt"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type AuthSvc struct {
	users     repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

// NewAuthService wires dependencies for authentication logic.
func NewAuthService(users repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) *AuthSvc {
	return &AuthSvc{
		users:     users,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *AuthSvc) Register(ctx context.Context, in RegisterInput) (string, domain.User, error) {
	// Hash password BEFORE saving. We never store plain-text passwords.
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	u := domain.User{
		FullName:          in.FullName,
		Email:             in.Email,
		PasswordHash:      string(hashBytes),
		Phone:             in.Phone,
		Role:              in.Role,
		CompanyName:       in.CompanyName,
		ResponsiblePerson: in.ResponsiblePerson,
		CompanyPhone:      in.CompanyPhone,
	}

	created, err := s.users.Create(ctx, u)
	if err != nil {
		// Friendly message for unique email constraint.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", domain.User{}, fmt.Errorf("email already registered")
		}
		return "", domain.User{}, err
	}

	// Create JWT token so the client can log in immediately after registration.
	token, err := jwtpkg.GenerateToken(created.ID.String(), created.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return "", domain.User{}, err
	}

	return token, created, nil
}

func (s *AuthSvc) Login(ctx context.Context, email, password string) (string, domain.User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", domain.User{}, fmt.Errorf("invalid email or password")
		}
		return "", domain.User{}, err
	}

	// Compare bcrypt hash with the provided password.
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", domain.User{}, fmt.Errorf("invalid email or password")
	}

	token, err := jwtpkg.GenerateToken(u.ID.String(), u.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return "", domain.User{}, err
	}

	return token, u, nil
}

