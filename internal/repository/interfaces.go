package repository

import (
	"context"
	"time"

	"city-service/internal/domain"

	"github.com/google/uuid"
)

// Repositories are the ONLY layer allowed to talk to PostgreSQL.
//
// Clean Architecture rule:
// - services depend on these interfaces (not on sqlx)
// - concrete implementations live in files like user_repo.go, request_repo.go

type UserRepository interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

type CategoryRepository interface {
	List(ctx context.Context) ([]domain.Category, error)
	GetByID(ctx context.Context, id int) (domain.Category, error)
}

// RequestFilters are optional query parameters used by list endpoints.
type RequestFilters struct {
	Status     *string
	CategoryID *int
	Urgency    *string

	// Date filters are used in admin stats/list. For monitor/contractor they can be unused.
	DateFrom *time.Time
	DateTo   *time.Time

	ContractorID *uuid.UUID
}

type RequestRepository interface {
	CountAll(ctx context.Context) (int, error)

	Create(ctx context.Context, r domain.Request) (domain.Request, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Request, error)

	ListAll(ctx context.Context, f RequestFilters) ([]domain.Request, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Request, error)

	// Contractor-specific listing helpers.
	ListNew(ctx context.Context, f RequestFilters) ([]domain.Request, error)
	ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]domain.Request, error)

	// State transitions.
	Cancel(ctx context.Context, requestID uuid.UUID) error
	AssignContractor(ctx context.Context, requestID uuid.UUID, contractorID uuid.UUID, takenAt time.Time) error
	MarkDone(ctx context.Context, requestID uuid.UUID) error
}

type CancellationRepository interface {
	Create(ctx context.Context, c domain.Cancellation) (domain.Cancellation, error)
}

type CompletionRepository interface {
	Create(ctx context.Context, c domain.Completion) (domain.Completion, error)
}
