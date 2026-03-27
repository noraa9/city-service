package handler

import (
	"net/http"
	"strconv"

	"city-service/internal/dto"
	"city-service/internal/middleware"
	"city-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ContractorHandler struct {
	requests    service.RequestService
	completions service.CompletionService
	validate    *validator.Validate
	maxUpload   int64
}

func NewContractorHandler(requests service.RequestService, completions service.CompletionService) *ContractorHandler {
	return &ContractorHandler{
		requests:    requests,
		completions: completions,
		validate:    validator.New(),
		maxUpload:   20 << 20, // 20MB
	}
}

// List
// @Summary      List available requests
// @Description  Contractor sees all requests with status=new
// @Tags         contractor
// @Produce      json
// @Security     BearerAuth
// @Param        category_id  query  int     false  "Filter by category ID"
// @Param        urgency      query  string  false  "Filter by urgency: low, medium, critical"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /contractor/requests [get]
func (h *ContractorHandler) List(w http.ResponseWriter, r *http.Request) {
	var f service.RequestFilters

	if v := r.URL.Query().Get("urgency"); v != "" {
		f.Urgency = &v
	}
	if v := r.URL.Query().Get("category_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			f.CategoryID = &id
		}
	}

	reqs, err := h.requests.ListNew(r.Context(), f)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]dto.RequestResponse, 0, len(reqs))
	for _, rr := range reqs {
		out = append(out, toRequestResponse(rr))
	}
	respondJSON(w, http.StatusOK, out)
}

// MyRequests
// @Summary      My active requests
// @Description  Contractor sees requests assigned to them
// @Tags         contractor
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /contractor/requests/my [get]
func (h *ContractorHandler) MyRequests(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	reqs, err := h.requests.ListByContractor(r.Context(), u.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]dto.RequestResponse, 0, len(reqs))
	for _, rr := range reqs {
		out = append(out, toRequestResponse(rr))
	}
	respondJSON(w, http.StatusOK, out)
}

// GetByID
// @Summary      Get request details
// @Description  Contractor views full request details including monitor contact
// @Tags         contractor
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Request UUID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /contractor/requests/{id} [get]
func (h *ContractorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	req, err := h.requests.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toRequestResponse(req))
}

// Take
// @Summary      Take request
// @Description  Contractor takes a request. Status changes: new → in_progress
// @Tags         contractor
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Request UUID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /contractor/requests/{id}/take [post]
func (h *ContractorHandler) Take(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	updated, err := h.requests.Take(r.Context(), id, u.ID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toRequestResponse(updated))
}

// Complete
// @Summary      Complete request
// @Description  Contractor closes request with completion data and photo. Status changes: in_progress → done
// @Tags         contractor
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id          path      string  true   "Request UUID"
// @Param        days_spent  formData  int     true   "Number of days spent"
// @Param        comment     formData  string  false  "Comment about monitor's work"
// @Param        photo       formData  file    false  "Photo of completed work"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /contractor/requests/{id}/complete [post]
func (h *ContractorHandler) Complete(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := r.ParseMultipartForm(h.maxUpload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	daysSpent, _ := strconv.Atoi(r.FormValue("days_spent"))
	form := dto.CompleteRequestDTO{
		DaysSpent: daysSpent,
		Comment:   r.FormValue("comment"),
	}
	if err := h.validate.Struct(form); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		respondError(w, http.StatusBadRequest, "photo is required")
		return
	}
	defer file.Close()

	ct := header.Header.Get("Content-Type")

	updated, err := h.completions.Complete(r.Context(), id, u.ID, service.CompleteRequestInput{
		DaysSpent: form.DaysSpent,
		Comment:   form.Comment,
		Photo: service.UploadFile{
			Reader:           file,
			Size:             header.Size,
			ContentType:      ct,
			OriginalFilename: header.Filename,
		},
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toRequestResponse(updated))
}
