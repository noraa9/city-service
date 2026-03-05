package handler

import (
	"encoding/json"
	"net/http"

	"city-service/internal/dto"
	"city-service/internal/service"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	auth     service.AuthService
	validate *validator.Validate
}

func NewAuthHandler(auth service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth:     auth,
		validate: validator.New(),
	}
}

// Register
// @Summary      Register new user
// @Description  Register as monitor, contractor or admin
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.RegisterRequest  true  "Registration data"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, user, err := h.auth.Register(r.Context(), service.RegisterInput{
		FullName:          req.FullName,
		Email:             req.Email,
		Password:          req.Password,
		Phone:             req.Phone,
		Role:              req.Role,
		CompanyName:       req.CompanyName,
		ResponsiblePerson: req.ResponsiblePerson,
		CompanyPhone:      req.CompanyPhone,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, dto.AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	})
}

// Login
// @Summary      Login
// @Description  Login with email and password, returns JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.LoginRequest  true  "Login credentials"
// @Success      200   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, user, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, dto.AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	})
}

