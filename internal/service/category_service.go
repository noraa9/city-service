package service

import (
	"context"
	"fmt"

	"city-service/internal/domain"
	"city-service/internal/repository"
)

type CategorySvc struct {
	categories repository.CategoryRepository
}

func NewCategoryService(categories repository.CategoryRepository) *CategorySvc {
	return &CategorySvc{categories: categories}
}

func (s *CategorySvc) List(ctx context.Context) ([]domain.Category, error) {
	cats, err := s.categories.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}
