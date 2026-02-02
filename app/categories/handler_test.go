package categories

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// Mock repository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetCategoryByCode(ctx context.Context, code string) (*models.Category, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func TestHandleGet_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	categories := []models.Category{
		{Code: "clothing", Name: "Clothing"},
		{Code: "shoes", Name: "Shoes"},
	}

	mockRepo.On("GetAllCategories", mock.Anything).Return(categories, nil)

	req := httptest.NewRequest("GET", "/categories", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGet(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response CategoriesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response.Categories))
}

func TestHandlePost_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	mockRepo.On("CreateCategory", mock.Anything, mock.AnythingOfType("*models.Category")).Return(nil)

	reqBody := CreateCategoryRequest{Code: "electronics", Name: "Electronics"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handler.HandlePost(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestHandlePost_MissingFields(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	reqBody := CreateCategoryRequest{Code: "", Name: "Electronics"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handler.HandlePost(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlePost_DuplicateCode(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	mockRepo.On("CreateCategory", mock.Anything, mock.AnythingOfType("*models.Category")).
		Return(models.ErrDuplicateCode)

	reqBody := CreateCategoryRequest{Code: "clothing", Name: "Clothing"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handler.HandlePost(recorder, req)

	assert.Equal(t, http.StatusConflict, recorder.Code)
}
