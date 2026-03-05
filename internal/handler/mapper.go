package handler

import (
	"time"

	"city-service/internal/domain"
	"city-service/internal/dto"
)

// Handlers translate domain models into DTOs for JSON responses.

func toUserResponse(u domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:                u.ID.String(),
		FullName:          u.FullName,
		Email:             u.Email,
		Phone:             u.Phone,
		Role:              u.Role,
		CompanyName:       u.CompanyName,
		ResponsiblePerson: u.ResponsiblePerson,
		CompanyPhone:      u.CompanyPhone,
	}
}

func toCategoryResponse(c domain.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:   c.ID,
		Name: c.Name,
		Slug: c.Slug,
	}
}

func toRequestResponse(r domain.Request) dto.RequestResponse {
	var deadline *string
	if r.Deadline != nil {
		s := r.Deadline.Format("2006-01-02")
		deadline = &s
	}

	var takenAt *string
	if r.TakenAt != nil {
		s := r.TakenAt.Format(time.RFC3339)
		takenAt = &s
	}

	var contractor *dto.UserResponse
	if r.Contractor != nil {
		c := toUserResponse(*r.Contractor)
		contractor = &c
	}

	// Category is optional in DB, but UI expects it; we return an empty object if nil.
	cat := dto.CategoryResponse{}
	if r.Category != nil {
		cat = toCategoryResponse(*r.Category)
	}

	monitor := dto.UserResponse{}
	if r.User != nil {
		monitor = toUserResponse(*r.User)
	}

	return dto.RequestResponse{
		ID:            r.ID.String(),
		RequestNumber: r.RequestNumber,
		Title:         r.Title,
		Category:      cat,
		Description:   r.Description,
		Urgency:       r.Urgency,
		Deadline:      deadline,
		Location:      r.Location,
		PhotoURL:      r.PhotoURL,
		Status:        r.Status,
		Monitor:       monitor,
		Contractor:    contractor,
		TakenAt:       takenAt,
		CreatedAt:     r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     r.UpdatedAt.Format(time.RFC3339),
	}
}

