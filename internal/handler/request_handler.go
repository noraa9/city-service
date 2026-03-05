package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"city-service/internal/dto"
	"city-service/internal/middleware"
	"city-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type RequestHandler struct {
	requests  service.RequestService
	validate  *validator.Validate
	maxUpload int64
}

func NewRequestHandler(requests service.RequestService) *RequestHandler {
	return &RequestHandler{
		requests:  requests,
		validate:  validator.New(),
		maxUpload: 20 << 20, // 20MB (simple safety limit for photos)
	}
}

// List
// @Summary      List all requests
// @Description  Get all requests with optional filters. Available to monitors.
// @Tags         monitor
// @Produce      json
// @Security     BearerAuth
// @Param        status       query  string  false  "Filter by status: new, in_progress, done, cancelled"
// @Param        category_id  query  int     false  "Filter by category ID"
// @Param        urgency      query  string  false  "Filter by urgency: low, medium, critical"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /requests [get]
func (h *RequestHandler) List(w http.ResponseWriter, r *http.Request) {
	var f service.RequestFilters

	if v := r.URL.Query().Get("status"); v != "" {
		f.Status = &v
	}
	if v := r.URL.Query().Get("urgency"); v != "" {
		f.Urgency = &v
	}
	if v := r.URL.Query().Get("category_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			f.CategoryID = &id
		}
	}

	// Simple "date=today" filter from spec.
	if r.URL.Query().Get("date") == "today" {
		now := time.Now()
		from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		to := from.Add(24*time.Hour - time.Nanosecond)
		f.DateFrom = &from
		f.DateTo = &to
	}

	reqs, err := h.requests.ListAll(r.Context(), f)
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
// @Summary      My requests
// @Description  Get current monitor's own requests
// @Tags         monitor
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /requests/my [get]
func (h *RequestHandler) MyRequests(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	reqs, err := h.requests.ListMine(r.Context(), u.ID)
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
// @Summary      Get request by ID
// @Description  Get full details of a request
// @Tags         monitor
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Request UUID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /requests/{id} [get]
func (h *RequestHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
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

// Create
// @Summary      Create request
// @Description  Monitor creates a new city request with optional photo
// @Tags         monitor
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        title        formData  string  true   "Request title"
// @Param        category_id  formData  int     true   "Category ID (1-5)"
// @Param        description  formData  string  true   "Problem description"
// @Param        urgency      formData  string  true   "Urgency: low | medium | critical"
// @Param        deadline     formData  string  false  "Deadline in format YYYY-MM-DD"
// @Param        location     formData  string  true   "Address / location text"
// @Param        photo        formData  file    false  "Photo of the problem"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /requests [post]
func (h *RequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse multipart form. maxUpload protects memory use.
	if err := r.ParseMultipartForm(h.maxUpload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	form := dto.CreateRequestDTO{
		Title:       r.FormValue("title"),
		CategoryID:  categoryID,
		Description: r.FormValue("description"),
		Urgency:     r.FormValue("urgency"),
		Deadline:    r.FormValue("deadline"),
		Location:    r.FormValue("location"),
	}

	if err := h.validate.Struct(form); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var deadline *time.Time
	if form.Deadline != "" {
		d, err := time.Parse("2006-01-02", form.Deadline)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid deadline format (use YYYY-MM-DD)")
			return
		}
		deadline = &d
	}

	var upload *service.UploadFile
	file, header, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		ct := header.Header.Get("Content-Type")
		upload = &service.UploadFile{
			Reader:           file,
			Size:             header.Size,
			ContentType:      ct,
			OriginalFilename: header.Filename,
		}
	}

	created, err := h.requests.Create(r.Context(), u, service.CreateRequestInput{
		Title:       form.Title,
		CategoryID:  form.CategoryID,
		Description: form.Description,
		Urgency:     form.Urgency,
		Deadline:    deadline,
		Location:    form.Location,
		Photo:       upload,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toRequestResponse(created))
}

// Cancel
// @Summary      Cancel request
// @Description  Monitor cancels own request with a reason
// @Tags         monitor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string              true  "Request UUID"
// @Param        body  body  dto.CancelRequestDTO  true  "Cancel reason"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /requests/{id}/cancel [post]
func (h *RequestHandler) Cancel(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CancelRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.requests.Cancel(r.Context(), id, u.ID, req.Reason, req.Comment)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toRequestResponse(updated))
}

