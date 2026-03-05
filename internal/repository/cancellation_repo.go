package repository

import (
	"context"
	"fmt"
	"time"

	"city-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CancellationRepo struct {
	db *sqlx.DB
}

func NewCancellationRepo(db *sqlx.DB) *CancellationRepo {
	return &CancellationRepo{db: db}
}

type dbCancellation struct {
	ID        uuid.UUID `db:"id"`
	RequestID uuid.UUID `db:"request_id"`
	Reason    string    `db:"reason"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *CancellationRepo) Create(ctx context.Context, c domain.Cancellation) (domain.Cancellation, error) {
	var out dbCancellation
	q := `
		INSERT INTO cancellations (request_id, reason, comment)
		VALUES ($1,$2,$3)
		RETURNING id, request_id, reason, comment, created_at
	`
	if err := r.db.GetContext(ctx, &out, q, c.RequestID, c.Reason, c.Comment); err != nil {
		return domain.Cancellation{}, fmt.Errorf("insert cancellation: %w", err)
	}
	return domain.Cancellation{
		ID:        out.ID,
		RequestID: out.RequestID,
		Reason:    out.Reason,
		Comment:   out.Comment,
		CreatedAt: out.CreatedAt,
	}, nil
}

