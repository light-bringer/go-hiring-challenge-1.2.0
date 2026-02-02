package catalog

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/dto"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

// Pagination constants
const (
	DefaultOffset = 0
	DefaultLimit  = 10
	MinLimit      = 1
	MaxLimit      = 100
)

// DecimalPrice is a custom type that marshals decimal.Decimal as JSON number with precision
type DecimalPrice struct {
	decimal.Decimal
}

func (d DecimalPrice) MarshalJSON() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *DecimalPrice) UnmarshalJSON(data []byte) error {
	var err error
	d.Decimal, err = decimal.NewFromString(string(data))
	return err
}

// Response DTOs
type CatalogResponse struct {
	Products []ProductDTO `json:"products"`
	Total    int64        `json:"total"`
	Offset   int          `json:"offset"`
	Limit    int          `json:"limit"`
}

type ProductDTO struct {
	Code     string           `json:"code"`
	Price    DecimalPrice     `json:"price"`
	Category *dto.CategoryDTO `json:"category,omitempty"`
}

type ProductDetailDTO struct {
	Code     string           `json:"code"`
	Price    DecimalPrice     `json:"price"`
	Category *dto.CategoryDTO `json:"category,omitempty"`
	Variants []VariantDTO     `json:"variants"`
}

type VariantDTO struct {
	Name  string       `json:"name"`
	SKU   string       `json:"sku"`
	Price DecimalPrice `json:"price"`
}

type CatalogHandler struct {
	service *CatalogService
}

func NewCatalogHandler(service *CatalogService) *CatalogHandler {
	return &CatalogHandler{service: service}
}

// HandleGet handles GET /catalog with pagination and filtering
func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	offset, limit, err := parsePaginationParams(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse filters
	filters, err := parseFilters(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Build options
	opts := models.ProductListOptions{
		Offset:        offset,
		Limit:         limit,
		CategoryCode:  filters.CategoryCode,
		PriceLessThan: filters.PriceLessThan,
	}

	// Fetch products via service
	products, total, err := h.service.ListProducts(r.Context(), opts)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	response := CatalogResponse{
		Products: products,
		Total:    total,
		Offset:   offset,
		Limit:    limit,
	}

	api.OKResponse(w, response)
}

// HandleGetByCode handles GET /catalog/{code}
func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Product code is required")
		return
	}

	product, err := h.service.GetProductWithVariants(r.Context(), code)
	if err != nil {
		if errors.Is(err, models.ErrProductNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "Product not found")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch product")
		return
	}

	api.OKResponse(w, product)
}

// Helper: Parse pagination parameters
func parsePaginationParams(r *http.Request) (offset, limit int, err error) {
	offset = DefaultOffset
	limit = DefaultLimit

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("invalid offset parameter")
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < MinLimit {
			return 0, 0, fmt.Errorf("invalid limit parameter")
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}
	}

	return offset, limit, nil
}

// Helper: Parse filter parameters
type Filters struct {
	CategoryCode  string
	PriceLessThan *decimal.Decimal
}

func parseFilters(r *http.Request) (Filters, error) {
	filters := Filters{}

	if category := r.URL.Query().Get("category"); category != "" {
		filters.CategoryCode = category
	}

	if priceStr := r.URL.Query().Get("price_less_than"); priceStr != "" {
		price, err := decimal.NewFromString(priceStr)
		if err != nil {
			return filters, fmt.Errorf("invalid price_less_than parameter")
		}
		if price.IsNegative() {
			return filters, fmt.Errorf("price_less_than must be positive")
		}
		filters.PriceLessThan = &price
	}

	return filters, nil
}
