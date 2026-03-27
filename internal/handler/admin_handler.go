package handler

import (
	"net/http"
	"strconv"
	"time"

	"city-service/internal/service"

	"github.com/google/uuid"
)

type AdminHandler struct {
	requests   service.RequestService
	categories service.CategoryService
}

func NewAdminHandler(requests service.RequestService, categories service.CategoryService) *AdminHandler {
	return &AdminHandler{requests: requests, categories: categories}
}

// ListAll
// @Summary      List all requests (admin)
// @Description  Admin gets all requests with full filters
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        status        query  string  false  "Filter by status"
// @Param        category_id   query  int     false  "Filter by category"
// @Param        urgency       query  string  false  "Filter by urgency"
// @Param        contractor_id query  string  false  "Filter by contractor UUID"
// @Param        date_from     query  string  false  "Filter from date YYYY-MM-DD"
// @Param        date_to       query  string  false  "Filter to date YYYY-MM-DD"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Router       /admin/requests [get]
func (h *AdminHandler) ListAll(w http.ResponseWriter, r *http.Request) {
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
	if v := r.URL.Query().Get("contractor_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.ContractorID = &id
		}
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		if d, err := time.Parse("2006-01-02", v); err == nil {
			f.DateFrom = &d
		}
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		if d, err := time.Parse("2006-01-02", v); err == nil {
			// include full day
			end := d.Add(24*time.Hour - time.Nanosecond)
			f.DateTo = &end
		}
	}

	reqs, err := h.requests.ListAll(r.Context(), f)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]any, 0, len(reqs))
	for _, rr := range reqs {
		out = append(out, toRequestResponse(rr))
	}
	respondJSON(w, http.StatusOK, out)
}

// Stats
// @Summary      Get statistics (admin)
// @Description  Returns counts by status, category and urgency
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Router       /admin/stats [get]
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	// For a beginner-friendly first version, we compute stats in memory using the request list.
	// Later, you can optimize this with dedicated SQL aggregate queries.

	reqs, err := h.requests.ListAll(r.Context(), service.RequestFilters{})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	byStatus := map[string]int{}
	byUrgency := map[string]int{}
	byCategory := map[string]int{} // name -> count

	for _, rr := range reqs {
		byStatus[rr.Status]++
		byUrgency[rr.Urgency]++
		if rr.Category != nil {
			byCategory[rr.Category.Name]++
		}
	}

	// Convert category map into a stable array for JSON.
	catRows := make([]map[string]any, 0, len(byCategory))
	for name, count := range byCategory {
		catRows = append(catRows, map[string]any{"category": name, "count": count})
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"total":       len(reqs),
		"by_status":   byStatus,
		"by_category": catRows,
		"by_urgency":  byUrgency,
	})
}
