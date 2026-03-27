package domain

import (
	"time"

	"github.com/google/uuid"
)

// Cancellation reasons match the DB CHECK constraint.
const (
	CancelNotRelevant = "not_relevant"
	CancelWrongData   = "wrong_data"
	CancelMistake     = "mistake"
	CancelOther       = "other"
)

// Cancellation records *why* a request was cancelled.
type Cancellation struct {
	ID        uuid.UUID
	RequestID uuid.UUID
	Reason    string
	Comment   string
	CreatedAt time.Time
}
