package service

import (
	"context"
	"fmt"

	"city-service/internal/domain"
	"city-service/internal/repository"
	miniopkg "city-service/pkg/minio"

	"github.com/google/uuid"
)

type CompletionSvc struct {
	requests    repository.RequestRepository
	completions repository.CompletionRepository
	storage     FileStorage
	bucket      string
}

func NewCompletionService(
	requests repository.RequestRepository,
	completions repository.CompletionRepository,
	storage FileStorage,
	bucket string,
) *CompletionSvc {
	return &CompletionSvc{
		requests:    requests,
		completions: completions,
		storage:     storage,
		bucket:      bucket,
	}
}

func (s *CompletionSvc) Complete(ctx context.Context, requestID uuid.UUID, contractorID uuid.UUID, in CompleteRequestInput) (domain.Request, error) {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return domain.Request{}, err
	}

	// Business rules from spec:
	// 1) request belongs to current contractor
	// 2) status is "in_progress"
	if req.ContractorID == nil || *req.ContractorID != contractorID {
		return domain.Request{}, fmt.Errorf("you can only complete requests assigned to you")
	}
	if req.Status != domain.StatusInProgress {
		return domain.Request{}, fmt.Errorf("only in_progress requests can be completed")
	}

	// Upload completion photo to MinIO (required).
	objectName := miniopkg.BuildObjectName(uuid.NewString(), in.Photo.OriginalFilename)
	url, err := s.storage.UploadFile(ctx, s.bucket, objectName, in.Photo.Reader, in.Photo.Size, in.Photo.ContentType)
	if err != nil {
		return domain.Request{}, err
	}

	_, err = s.completions.Create(ctx, domain.Completion{
		RequestID: requestID,
		DaysSpent: in.DaysSpent,
		Comment:   in.Comment,
		PhotoURL:  url,
	})
	if err != nil {
		return domain.Request{}, err
	}

	if err := s.requests.MarkDone(ctx, requestID); err != nil {
		return domain.Request{}, err
	}

	return s.requests.GetByID(ctx, requestID)
}
