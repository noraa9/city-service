package domain

import (
	"time"

	"github.com/google/uuid"
)

// Roles in this project. We keep them as constants so:
// - validation can reference the same canonical strings
// - typos don't silently create new roles
const (
	RoleMonitor    = "monitor"
	RoleContractor = "contractor"
	RoleAdmin      = "admin"
)

// User is the core domain model for authentication + ownership.
//
// Important design choice:
// - This struct is "pure domain": no JSON tags, no SQL tags.
//   JSON belongs in DTOs, SQL mapping belongs in repository layer.
type User struct {
	ID           uuid.UUID
	FullName     string
	Email        string
	PasswordHash string
	Phone        string
	Role         string

	// Contractor-only fields are nullable in DB, so we model them as pointers.
	// nil means "not provided / not applicable".
	CompanyName       *string
	ResponsiblePerson *string
	CompanyPhone      *string

	CreatedAt time.Time
}

