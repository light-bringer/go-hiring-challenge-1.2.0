package categories

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/dto"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// Validation constants
const (
	MaxCodeLength = 50
	MaxNameLength = 255
)

// Response DTOs
type CategoriesResponse struct {
	Categories []dto.CategoryDTO `json:"categories"`
}

type CreateCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Validate checks if the request is valid
func (r CreateCategoryRequest) Validate() error {
	if r.Code == "" {
		return fmt.Errorf("code is required")
	}
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Code) > MaxCodeLength {
		return fmt.Errorf("code must not exceed %d characters", MaxCodeLength)
	}
	if len(r.Name) > MaxNameLength {
		return fmt.Errorf("name must not exceed %d characters", MaxNameLength)
	}
	return nil
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

	categoryDTOs := make([]dto.CategoryDTO, len(categories))
	for i, c := range categories {
		categoryDTOs[i] = dto.CategoryDTO{
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
	if err := req.Validate(); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
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

	response := dto.CategoryDTO{
		Code: category.Code,
		Name: category.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/categories/%s", category.Code))
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}
