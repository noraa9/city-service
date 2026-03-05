package repository

import (
	"context"
	"database/sql"
	"fmt"

	"city-service/internal/domain"

	"github.com/jmoiron/sqlx"
)

type CategoryRepo struct {
	db *sqlx.DB
}

func NewCategoryRepo(db *sqlx.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

type dbCategory struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Slug string `db:"slug"`
}

func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	var rows []dbCategory
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, name, slug FROM categories ORDER BY id`); err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	out := make([]domain.Category, 0, len(rows))
	for _, c := range rows {
		out = append(out, domain.Category{ID: c.ID, Name: c.Name, Slug: c.Slug})
	}
	return out, nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int) (domain.Category, error) {
	var row dbCategory
	if err := r.db.GetContext(ctx, &row, `SELECT id, name, slug FROM categories WHERE id = $1`, id); err != nil {
		if err == sql.ErrNoRows {
			return domain.Category{}, ErrNotFound
		}
		return domain.Category{}, fmt.Errorf("get category: %w", err)
	}
	return domain.Category{ID: row.ID, Name: row.Name, Slug: row.Slug}, nil
}

