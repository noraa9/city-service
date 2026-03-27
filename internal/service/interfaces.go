package service

import (
	"context"
	"io"
	"time"

	"city-service/internal/domain"

	"github.com/google/uuid"
)

// Services contain business logic.
//
// Clean Architecture rule:
// - services DO NOT import net/http
// - handlers own HTTP concerns (parsing, status codes, JSON)

type AuthService interface {
	Register(ctx context.Context, in RegisterInput) (token string, user domain.User, err error)
	Login(ctx context.Context, email, password string) (token string, user domain.User, err error)
}

type CategoryService interface {
	List(ctx context.Context) ([]domain.Category, error)
}

type RequestService interface {
	Create(ctx context.Context, creator domain.User, in CreateRequestInput) (domain.Request, error)
	ListAll(ctx context.Context, f RequestFilters) ([]domain.Request, error)
	ListMine(ctx context.Context, userID uuid.UUID) ([]domain.Request, error)
	ListNew(ctx context.Context, f RequestFilters) ([]domain.Request, error)
	ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]domain.Request, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Request, error)
	Cancel(ctx context.Context, requestID uuid.UUID, userID uuid.UUID, reason, comment string) (domain.Request, error)
	Take(ctx context.Context, requestID uuid.UUID, contractorID uuid.UUID) (domain.Request, error)
}

type CompletionService interface {
	Complete(ctx context.Context, requestID uuid.UUID, contractorID uuid.UUID, in CompleteRequestInput) (domain.Request, error)
}

// FileStorage is a small abstraction over MinIO.
// It makes services independent from a concrete MinIO implementation.
type FileStorage interface {
	UploadFile(ctx context.Context, bucketName, fileName string, file io.Reader, size int64, contentType string) (string, error)
}

// Inputs are service-level shapes (not HTTP DTOs).
// They are easier to unit-test than net/http requests.

type RegisterInput struct {
	FullName          string
	Email             string
	Password          string
	Phone             string
	Role              string
	CompanyName       *string
	ResponsiblePerson *string
	CompanyPhone      *string
}

type CreateRequestInput struct {
	Title       string
	CategoryID  int
	Description string
	Urgency     string
	Deadline    *time.Time
	Location    string

	// Photo is optional for creating a request.
	// If nil, the request will be created without a photo_url.
	Photo *UploadFile
}

type CompleteRequestInput struct {
	DaysSpent int
	Comment   string

	// Photo is required when completing a request.
	Photo UploadFile
}

// UploadFile carries file data into the service layer.
//
// Important decision:
// - handlers read multipart files from net/http
// - services decide storage details (MinIO bucket, naming, URL)
// This keeps upload logic in services as required by your spec.
type UploadFile struct {
	Reader           io.Reader
	Size             int64
	ContentType      string
	OriginalFilename string
}

// RequestFilters mirrors repository filters but is owned by service layer.
// We keep it in service so handlers depend on services, not repositories.
type RequestFilters struct {
	Status     *string
	CategoryID *int
	Urgency    *string

	DateFrom *time.Time
	DateTo   *time.Time

	ContractorID *uuid.UUID
}
