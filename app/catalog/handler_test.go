package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// Mock repository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetAllProducts(ctx context.Context, opts models.ProductListOptions) ([]models.Product, int64, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) GetProductByCode(ctx context.Context, code string) (*models.Product, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func TestHandleGet_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	category := &models.Category{Code: "clothing", Name: "Clothing"}
	products := []models.Product{
		{Code: "PROD001", Price: decimal.NewFromFloat(10.99), Category: category},
	}

	mockRepo.On("GetAllProducts", mock.Anything, mock.Anything).Return(products, int64(1), nil)

	req := httptest.NewRequest("GET", "/catalog", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGet(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response CatalogResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, 1, len(response.Products))
	assert.Equal(t, "PROD001", response.Products[0].Code)
}

func TestHandleGet_Pagination(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	mockRepo.On("GetAllProducts", mock.Anything, mock.MatchedBy(func(opts models.ProductListOptions) bool {
		return opts.Offset == 5 && opts.Limit == 20
	})).Return([]models.Product{}, int64(100), nil)

	req := httptest.NewRequest("GET", "/catalog?offset=5&limit=20", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGet(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response CatalogResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, 5, response.Offset)
	assert.Equal(t, 20, response.Limit)
}

func TestHandleGet_InvalidLimit(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	req := httptest.NewRequest("GET", "/catalog?limit=abc", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGet(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleGetByCode_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	category := &models.Category{Code: "clothing", Name: "Clothing"}
	price1 := decimal.NewFromFloat(11.99)
	product := &models.Product{
		Code:     "PROD001",
		Price:    decimal.NewFromFloat(10.99),
		Category: category,
		Variants: []models.Variant{
			{Name: "Variant A", SKU: "SKU001A", Price: &price1},
			{Name: "Variant B", SKU: "SKU001B", Price: nil}, // Inherits price
		},
	}

	mockRepo.On("GetProductByCode", mock.Anything, "PROD001").Return(product, nil)

	req := httptest.NewRequest("GET", "/catalog/PROD001", nil)
	req.SetPathValue("code", "PROD001")
	recorder := httptest.NewRecorder()

	handler.HandleGetByCode(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response ProductDetailDTO
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, "PROD001", response.Code)
	assert.Equal(t, 2, len(response.Variants))
	assert.Equal(t, "11.99", response.Variants[0].Price.String())
	assert.Equal(t, "10.99", response.Variants[1].Price.String()) // Inherited
}

func TestHandleGetByCode_NotFound(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	mockRepo.On("GetProductByCode", mock.Anything, "INVALID").Return(nil, models.ErrProductNotFound)

	req := httptest.NewRequest("GET", "/catalog/INVALID", nil)
	req.SetPathValue("code", "INVALID")
	recorder := httptest.NewRecorder()

	handler.HandleGetByCode(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

// Boundary case tests

func TestHandleGet_LimitBoundaries(t *testing.T) {
	tests := []struct {
		name           string
		queryString    string
		expectedLimit  int
		expectedStatus int
	}{
		{"limit_zero", "limit=0", 0, http.StatusBadRequest},
		{"limit_one", "limit=1", 1, http.StatusOK},
		{"limit_max", "limit=100", 100, http.StatusOK},
		{"limit_above_max", "limit=101", 100, http.StatusOK}, // Capped at max
		{"limit_very_large", "limit=999", 100, http.StatusOK}, // Capped at max
		{"limit_negative", "limit=-5", 0, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			service := NewCatalogService(mockRepo)
			handler := NewCatalogHandler(service)

			if tt.expectedStatus == http.StatusOK {
				mockRepo.On("GetAllProducts", mock.Anything, mock.MatchedBy(func(opts models.ProductListOptions) bool {
					return opts.Limit == tt.expectedLimit
				})).Return([]models.Product{}, int64(0), nil)
			}

			req := httptest.NewRequest("GET", "/catalog?"+tt.queryString, nil)
			recorder := httptest.NewRecorder()

			handler.HandleGet(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK {
				var response CatalogResponse
				json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedLimit, response.Limit)
			}
		})
	}
}

func TestHandleGet_OffsetBoundaries(t *testing.T) {
	tests := []struct {
		name           string
		queryString    string
		expectedOffset int
		expectedStatus int
	}{
		{"offset_zero", "offset=0", 0, http.StatusOK},
		{"offset_positive", "offset=50", 50, http.StatusOK},
		{"offset_negative", "offset=-1", 0, http.StatusBadRequest},
		{"offset_large", "offset=1000", 1000, http.StatusOK}, // Returns empty if exceeds total
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			service := NewCatalogService(mockRepo)
			handler := NewCatalogHandler(service)

			if tt.expectedStatus == http.StatusOK {
				mockRepo.On("GetAllProducts", mock.Anything, mock.MatchedBy(func(opts models.ProductListOptions) bool {
					return opts.Offset == tt.expectedOffset
				})).Return([]models.Product{}, int64(0), nil)
			}

			req := httptest.NewRequest("GET", "/catalog?"+tt.queryString, nil)
			recorder := httptest.NewRecorder()

			handler.HandleGet(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestHandleGet_EmptyResults(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	// Return empty array when no products match filter
	mockRepo.On("GetAllProducts", mock.Anything, mock.Anything).Return([]models.Product{}, int64(0), nil)

	req := httptest.NewRequest("GET", "/catalog?category=nonexistent", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGet(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response CatalogResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, int64(0), response.Total)
	assert.Equal(t, 0, len(response.Products))
}

func TestHandleGet_DefaultPagination(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	// Test that defaults are applied when no pagination params provided
	mockRepo.On("GetAllProducts", mock.Anything, mock.MatchedBy(func(opts models.ProductListOptions) bool {
		return opts.Offset == 0 && opts.Limit == 10 // Should use defaults
	})).Return([]models.Product{}, int64(0), nil)

	req := httptest.NewRequest("GET", "/catalog", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGet(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response CatalogResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, 0, response.Offset) // Default
	assert.Equal(t, 10, response.Limit) // Default
}

func TestHandleGetByCode_EmptyCode(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	req := httptest.NewRequest("GET", "/catalog/", nil)
	req.SetPathValue("code", "")
	recorder := httptest.NewRecorder()

	handler.HandleGetByCode(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleGetByCode_ProductWithNoVariants(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewCatalogService(mockRepo)
	handler := NewCatalogHandler(service)

	category := &models.Category{Code: "clothing", Name: "Clothing"}
	product := &models.Product{
		Code:     "PROD999",
		Price:    decimal.NewFromFloat(10.99),
		Category: category,
		Variants: []models.Variant{}, // No variants
	}

	mockRepo.On("GetProductByCode", mock.Anything, "PROD999").Return(product, nil)

	req := httptest.NewRequest("GET", "/catalog/PROD999", nil)
	req.SetPathValue("code", "PROD999")
	recorder := httptest.NewRecorder()

	handler.HandleGetByCode(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response ProductDetailDTO
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, "PROD999", response.Code)
	assert.Equal(t, 0, len(response.Variants))
}

func TestHandleGet_PriceBoundaries(t *testing.T) {
	tests := []struct {
		name           string
		queryString    string
		expectedStatus int
	}{
		{"price_zero", "price_less_than=0", http.StatusBadRequest}, // Negative/zero not allowed
		{"price_very_small", "price_less_than=0.01", http.StatusOK},
		{"price_large", "price_less_than=999999.99", http.StatusOK},
		{"price_invalid", "price_less_than=abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			service := NewCatalogService(mockRepo)
			handler := NewCatalogHandler(service)

			if tt.expectedStatus == http.StatusOK {
				mockRepo.On("GetAllProducts", mock.Anything, mock.Anything).Return([]models.Product{}, int64(0), nil)
			}

			req := httptest.NewRequest("GET", "/catalog?"+tt.queryString, nil)
			recorder := httptest.NewRecorder()

			handler.HandleGet(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}
