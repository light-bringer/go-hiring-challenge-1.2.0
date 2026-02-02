package categories

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// Response DTOs
type CategoriesResponse struct {
	Categories []CategoryDTO `json:"categories"`
}

type CategoryDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CategoriesHandler struct {
	categoryRepo models.CategoryRepository
}

func NewCategoriesHandler(repo models.CategoryRepository) *CategoriesHandler {
	return &CategoriesHandler{categoryRepo: repo}
}

// HandleGet handles GET /categories
func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.GetAllCategories(r.Context())
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	categoryDTOs := make([]CategoryDTO, len(categories))
	for i, c := range categories {
		categoryDTOs[i] = CategoryDTO{
			Code: c.Code,
			Name: c.Name,
		}
	}

	response := CategoriesResponse{Categories: categoryDTOs}
	api.OKResponse(w, response)
}

// HandlePost handles POST /categories
func (h *CategoriesHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Code and name are required")
		return
	}

	category := &models.Category{
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.categoryRepo.CreateCategory(r.Context(), category); err != nil {
		if errors.Is(err, models.ErrDuplicateCode) {
			api.ErrorResponse(w, http.StatusConflict, "Category with this code already exists")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to create category")
		return
	}

	dto := CategoryDTO{
		Code: category.Code,
		Name: category.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto)
}
