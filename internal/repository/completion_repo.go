package repository

import (
	"context"
	"fmt"
	"time"

	"city-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CompletionRepo struct {
	db *sqlx.DB
}

func NewCompletionRepo(db *sqlx.DB) *CompletionRepo {
	return &CompletionRepo{db: db}
}

type dbCompletion struct {
	ID        uuid.UUID `db:"id"`
	RequestID uuid.UUID `db:"request_id"`
	DaysSpent int       `db:"days_spent"`
	Comment   string    `db:"comment"`
	PhotoURL  string    `db:"photo_url"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *CompletionRepo) Create(ctx context.Context, c domain.Completion) (domain.Completion, error) {
	var out dbCompletion
	q := `
		INSERT INTO completions (request_id, days_spent, comment, photo_url)
		VALUES ($1,$2,$3,$4)
		RETURNING id, request_id, days_spent, comment, photo_url, created_at
	`
	if err := r.db.GetContext(ctx, &out, q, c.RequestID, c.DaysSpent, c.Comment, c.PhotoURL); err != nil {
		return domain.Completion{}, fmt.Errorf("insert completion: %w", err)
	}
	return domain.Completion{
		ID:        out.ID,
		RequestID: out.RequestID,
		DaysSpent: out.DaysSpent,
		Comment:   out.Comment,
		PhotoURL:  out.PhotoURL,
		CreatedAt: out.CreatedAt,
	}, nil
}
