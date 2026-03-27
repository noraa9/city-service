package handler

import (
	"net/http"

	"city-service/internal/service"
)

type CategoryHandler struct {
	categories service.CategoryService
}

func NewCategoryHandler(categories service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

// List categories
// @Summary      List categories
// @Description  Get all request categories
// @Tags         categories
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.categories.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]any, 0, len(cats))
	for _, c := range cats {
		out = append(out, toCategoryResponse(c))
	}

	respondJSON(w, http.StatusOK, out)
}
