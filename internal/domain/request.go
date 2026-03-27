package domain

import (
	"time"

	"github.com/google/uuid"
)

// Request statuses. Keep them in one place for consistency across API + DB checks.
const (
	StatusNew        = "new"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
	StatusCancelled  = "cancelled"
)

// Urgency levels for requests.
const (
	UrgencyLow      = "low"
	UrgencyMedium   = "medium"
	UrgencyCritical = "critical"
)

// Request is the main business entity.
//
// It contains optional "joined" objects (Category/User/Contractor) that repositories
// may populate when they perform JOIN queries.
type Request struct {
	ID            uuid.UUID
	RequestNumber string
	Title         string

	CategoryID int
	Category   *Category // optional: filled by JOIN

	Description string
	Urgency     string
	Deadline    *time.Time
	Location    string
	PhotoURL    *string

	Status string

	UserID uuid.UUID
	User   *User // optional: filled by JOIN

	ContractorID *uuid.UUID
	Contractor   *User // optional: filled by JOIN

	TakenAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
