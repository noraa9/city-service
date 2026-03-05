package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"city-service/internal/domain"
	"city-service/internal/repository"
	miniopkg "city-service/pkg/minio"

	"github.com/google/uuid"
)

type RequestSvc struct {
	requests      repository.RequestRepository
	categories    repository.CategoryRepository
	cancellations repository.CancellationRepository
	storage       FileStorage
	bucket        string
}

func NewRequestService(
	requests repository.RequestRepository,
	categories repository.CategoryRepository,
	cancellations repository.CancellationRepository,
	storage FileStorage,
	bucket string,
) *RequestSvc {
	return &RequestSvc{
		requests:      requests,
		categories:    categories,
		cancellations: cancellations,
		storage:       storage,
		bucket:        bucket,
	}
}

func (s *RequestSvc) Create(ctx context.Context, creator domain.User, in CreateRequestInput) (domain.Request, error) {
	// Validate category exists (better error than a DB FK violation).
	if _, err := s.categories.GetByID(ctx, in.CategoryID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.Request{}, fmt.Errorf("category not found")
		}
		return domain.Request{}, err
	}

	// Upload optional photo to MinIO and store the resulting URL.
	var photoURL *string
	if in.Photo != nil {
		objectName := miniopkg.BuildObjectName(uuid.NewString(), in.Photo.OriginalFilename)
		url, err := s.storage.UploadFile(ctx, s.bucket, objectName, in.Photo.Reader, in.Photo.Size, in.Photo.ContentType)
		if err != nil {
			return domain.Request{}, err
		}
		photoURL = &url
	}

	// Generate request number according to spec:
	//   {N}{L}{YYMMDD}
	// - N = sequential count + 1
	// - L = U (monitor) or S (service/contractor)
	// - YYMMDD = current date
	count, err := s.requests.CountAll(ctx)
	if err != nil {
		return domain.Request{}, err
	}
	reqNumber := generateRequestNumber(count, creator.Role, time.Now())

	req := domain.Request{
		RequestNumber: reqNumber,
		Title:         in.Title,
		CategoryID:    in.CategoryID,
		Description:   in.Description,
		Urgency:       in.Urgency,
		Deadline:      in.Deadline,
		Location:      in.Location,
		PhotoURL:      photoURL,
		Status:        domain.StatusNew,
		UserID:        creator.ID,
	}

	created, err := s.requests.Create(ctx, req)
	if err != nil {
		return domain.Request{}, err
	}
	return created, nil
}

func (s *RequestSvc) ListAll(ctx context.Context, f RequestFilters) ([]domain.Request, error) {
	rf := repository.RequestFilters{
		Status:       f.Status,
		CategoryID:   f.CategoryID,
		Urgency:      f.Urgency,
		DateFrom:     f.DateFrom,
		DateTo:       f.DateTo,
		ContractorID: f.ContractorID,
	}
	return s.requests.ListAll(ctx, rf)
}

func (s *RequestSvc) ListMine(ctx context.Context, userID uuid.UUID) ([]domain.Request, error) {
	return s.requests.ListByUser(ctx, userID)
}

func (s *RequestSvc) ListNew(ctx context.Context, f RequestFilters) ([]domain.Request, error) {
	rf := repository.RequestFilters{
		CategoryID: f.CategoryID,
		Urgency:    f.Urgency,
	}
	return s.requests.ListNew(ctx, rf)
}

func (s *RequestSvc) ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]domain.Request, error) {
	return s.requests.ListByContractor(ctx, contractorID)
}

func (s *RequestSvc) GetByID(ctx context.Context, id uuid.UUID) (domain.Request, error) {
	return s.requests.GetByID(ctx, id)
}

func (s *RequestSvc) Cancel(ctx context.Context, requestID uuid.UUID, userID uuid.UUID, reason, comment string) (domain.Request, error) {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return domain.Request{}, err
	}

	// Business rules from spec:
	// 1) request must belong to current user
	// 2) only status "new" can be cancelled
	if req.UserID != userID {
		return domain.Request{}, fmt.Errorf("you can only cancel your own requests")
	}
	if req.Status != domain.StatusNew {
		return domain.Request{}, fmt.Errorf("only new requests can be cancelled")
	}

	if err := s.requests.Cancel(ctx, requestID); err != nil {
		return domain.Request{}, err
	}

	_, err = s.cancellations.Create(ctx, domain.Cancellation{
		RequestID: requestID,
		Reason:    reason,
		Comment:   comment,
	})
	if err != nil {
		return domain.Request{}, err
	}

	return s.requests.GetByID(ctx, requestID)
}

func (s *RequestSvc) Take(ctx context.Context, requestID uuid.UUID, contractorID uuid.UUID) (domain.Request, error) {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return domain.Request{}, err
	}

	// Business rules from spec:
	// - must be "new"
	// - must not have a contractor yet
	if req.Status != domain.StatusNew {
		return domain.Request{}, fmt.Errorf("only new requests can be taken")
	}
	if req.ContractorID != nil {
		return domain.Request{}, fmt.Errorf("request is already taken")
	}

	if err := s.requests.AssignContractor(ctx, requestID, contractorID, time.Now()); err != nil {
		return domain.Request{}, err
	}
	return s.requests.GetByID(ctx, requestID)
}

func generateRequestNumber(existingCount int, role string, now time.Time) string {
	letter := "S"
	if role == domain.RoleMonitor {
		letter = "U"
	}
	// YYMMDD: Go reference time "Mon Jan 2 15:04:05 -0700 MST 2006"
	// Format "060102" means: 06=year, 01=month, 02=day.
	date := now.Format("060102")
	return fmt.Sprintf("%d%s%s", existingCount+1, letter, date)
}

