package dto

// DTOs ("Data Transfer Objects") are shapes for HTTP input/output.
//
// Key idea:
// - Domain structs are pure business entities.
// - DTOs have JSON/form tags + validation tags (because that's HTTP-facing).

type RegisterRequest struct {
	FullName          string  `json:"full_name" validate:"required"`
	Email             string  `json:"email" validate:"required,email"`
	Password          string  `json:"password" validate:"required,min=6"`
	Phone             string  `json:"phone"`
	Role              string  `json:"role" validate:"required,oneof=monitor contractor admin"`
	CompanyName       *string `json:"company_name"`
	ResponsiblePerson *string `json:"responsible_person"`
	CompanyPhone      *string `json:"company_phone"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID                string  `json:"id"`
	FullName          string  `json:"full_name"`
	Email             string  `json:"email"`
	Phone             string  `json:"phone"`
	Role              string  `json:"role"`
	CompanyName       *string `json:"company_name,omitempty"`
	ResponsiblePerson *string `json:"responsible_person,omitempty"`
	CompanyPhone      *string `json:"company_phone,omitempty"`
}

