package domain

import (
	"time"

	"github.com/google/uuid"
)

// Completion stores "closing" info once a contractor finishes a request.
type Completion struct {
	ID        uuid.UUID
	RequestID uuid.UUID

	DaysSpent int
	Comment   string
	PhotoURL  string

	CreatedAt time.Time
}

