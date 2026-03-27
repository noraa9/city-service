package dto

// CreateRequestDTO is parsed from multipart/form-data.
//
// Important:
// - Tags use `form:"..."` because Chi/net/http will parse a form.
// - The photo file itself is not in this struct; it's read from r.FormFile("photo").
type CreateRequestDTO struct {
	Title       string `form:"title" validate:"required"`
	CategoryID  int    `form:"category_id" validate:"required"`
	Description string `form:"description" validate:"required"`
	Urgency     string `form:"urgency" validate:"required,oneof=low medium critical"`
	Deadline    string `form:"deadline"` // "2025-12-31"
	Location    string `form:"location" validate:"required"`
}

type CancelRequestDTO struct {
	Reason  string `json:"reason" validate:"required,oneof=not_relevant wrong_data mistake other"`
	Comment string `json:"comment"`
}

// CompleteRequestDTO is parsed from multipart/form-data.
// Photo file is required and comes from r.FormFile("photo").
type CompleteRequestDTO struct {
	DaysSpent int    `form:"days_spent" validate:"required,min=1"`
	Comment   string `form:"comment"`
}

// CategoryResponse is embedded into RequestResponse.
type CategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type RequestResponse struct {
	ID            string           `json:"id"`
	RequestNumber string           `json:"request_number"`
	Title         string           `json:"title"`
	Category      CategoryResponse `json:"category"`
	Description   string           `json:"description"`
	Urgency       string           `json:"urgency"`
	Deadline      *string          `json:"deadline,omitempty"`
	Location      string           `json:"location"`
	PhotoURL      *string          `json:"photo_url,omitempty"`
	Status        string           `json:"status"`
	Monitor       UserResponse     `json:"monitor"`
	Contractor    *UserResponse    `json:"contractor,omitempty"`
	TakenAt       *string          `json:"taken_at,omitempty"`
	CreatedAt     string           `json:"created_at"`
	UpdatedAt     string           `json:"updated_at"`
}
