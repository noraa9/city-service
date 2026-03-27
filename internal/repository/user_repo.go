package repository

import (
	"context"
	"database/sql"
	"fmt"

	"city-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserRepo is a PostgreSQL implementation of UserRepository.
type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

// dbUser is the SQL mapping struct.
//
// It lives in repository layer because it knows about database column names.
type dbUser struct {
	ID                uuid.UUID      `db:"id"`
	FullName          string         `db:"full_name"`
	Email             string         `db:"email"`
	PasswordHash      string         `db:"password_hash"`
	Phone             sql.NullString `db:"phone"`
	Role              string         `db:"role"`
	CompanyName       sql.NullString `db:"company_name"`
	ResponsiblePerson sql.NullString `db:"responsible_person"`
	CompanyPhone      sql.NullString `db:"company_phone"`
	CreatedAt         sql.NullTime   `db:"created_at"`
}

func (r *UserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	var out dbUser

	// We return the row so callers get the generated UUID + created_at.
	q := `
		INSERT INTO users (
			full_name, email, password_hash, phone, role,
			company_name, responsible_person, company_phone
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING
			id, full_name, email, password_hash, phone, role,
			company_name, responsible_person, company_phone, created_at
	`

	if err := r.db.GetContext(ctx, &out, q,
		u.FullName,
		u.Email,
		u.PasswordHash,
		nullString(u.Phone),
		u.Role,
		nullStringPtr(u.CompanyName),
		nullStringPtr(u.ResponsiblePerson),
		nullStringPtr(u.CompanyPhone),
	); err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	return toDomainUser(out), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var out dbUser
	q := `
		SELECT
			id, full_name, email, password_hash, phone, role,
			company_name, responsible_person, company_phone, created_at
		FROM users
		WHERE email = $1
	`
	if err := r.db.GetContext(ctx, &out, q, email); err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("select user by email: %w", err)
	}
	return toDomainUser(out), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var out dbUser
	q := `
		SELECT
			id, full_name, email, password_hash, phone, role,
			company_name, responsible_person, company_phone, created_at
		FROM users
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &out, q, id); err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("select user by id: %w", err)
	}
	return toDomainUser(out), nil
}

func toDomainUser(u dbUser) domain.User {
	return domain.User{
		ID:                u.ID,
		FullName:          u.FullName,
		Email:             u.Email,
		PasswordHash:      u.PasswordHash,
		Phone:             u.Phone.String,
		Role:              u.Role,
		CompanyName:       stringPtr(u.CompanyName),
		ResponsiblePerson: stringPtr(u.ResponsiblePerson),
		CompanyPhone:      stringPtr(u.CompanyPhone),
		CreatedAt:         u.CreatedAt.Time,
	}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullStringPtr(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func stringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}
