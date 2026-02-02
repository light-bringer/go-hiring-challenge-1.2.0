package categories

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/dto"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestHandleGetByCode_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	category := &models.Category{Code: "shoes", Name: "Shoes"}
	mockRepo.On("GetCategoryByCode", mock.Anything, "shoes").Return(category, nil)

	req := httptest.NewRequest("GET", "/categories/shoes", nil)
	req.SetPathValue("code", "shoes")
	recorder := httptest.NewRecorder()

	handler.HandleGetByCode(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response dto.CategoryDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "shoes", response.Code)
	assert.Equal(t, "Shoes", response.Name)
}

func TestHandleGetByCode_NotFound(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	mockRepo.On("GetCategoryByCode", mock.Anything, "missing").Return(nil, models.ErrCategoryNotFound)

	req := httptest.NewRequest("GET", "/categories/missing", nil)
	req.SetPathValue("code", "missing")
	recorder := httptest.NewRecorder()

	handler.HandleGetByCode(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandleGetByCode_EmptyCode(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	req := httptest.NewRequest("GET", "/categories/", nil)
	req.SetPathValue("code", "")
	recorder := httptest.NewRecorder()

	handler.HandleGetByCode(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
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

// Boundary case tests

func TestHandlePost_FieldLengthBoundaries(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		categoryName   string
		expectedStatus int
	}{
		{"code_max_length", string(make([]byte, 50)), "Valid Name", http.StatusCreated},
		{"code_too_long", string(make([]byte, 51)), "Valid Name", http.StatusBadRequest},
		{"name_max_length", "validcode", string(make([]byte, 255)), http.StatusCreated},
		{"name_too_long", "validcode", string(make([]byte, 256)), http.StatusBadRequest},
		{"both_max", string(make([]byte, 50)), string(make([]byte, 255)), http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			handler := NewCategoriesHandler(mockRepo)

			if tt.expectedStatus == http.StatusCreated {
				mockRepo.On("CreateCategory", mock.Anything, mock.AnythingOfType("*models.Category")).Return(nil)
			}

			// Fill arrays with valid characters
			code := tt.code
			if len(code) > 0 {
				for i := range []byte(code) {
					code = code[:i] + "a" + code[i+1:]
				}
			}

			name := tt.categoryName
			if len(name) > 0 {
				for i := range []byte(name) {
					name = name[:i] + "a" + name[i+1:]
				}
			}

			reqBody := CreateCategoryRequest{Code: code, Name: name}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
			recorder := httptest.NewRecorder()

			handler.HandlePost(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestHandlePost_BothFieldsEmpty(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	reqBody := CreateCategoryRequest{Code: "", Name: ""}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handler.HandlePost(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlePost_InvalidJSON(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	// Malformed JSON
	req := httptest.NewRequest("POST", "/categories", bytes.NewReader([]byte("{invalid json")))
	recorder := httptest.NewRecorder()

	handler.HandlePost(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlePost_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		categoryName   string
		expectedStatus int
	}{
		{"code_with_spaces", "hello world", "Name", http.StatusCreated},
		{"code_with_special_chars", "hello@#$%", "Name", http.StatusCreated},
		{"name_with_unicode", "code", "Café ☕", http.StatusCreated},
		{"name_with_emoji", "code", "Emoji 😀", http.StatusCreated},
		{"name_with_chinese", "code", "中文名称", http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			handler := NewCategoriesHandler(mockRepo)

			mockRepo.On("CreateCategory", mock.Anything, mock.AnythingOfType("*models.Category")).Return(nil)

			reqBody := CreateCategoryRequest{Code: tt.code, Name: tt.categoryName}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
			recorder := httptest.NewRecorder()

			handler.HandlePost(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestHandleGet_EmptyCategories(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoriesHandler(mockRepo)

	// Return empty array
	mockRepo.On("GetAllCategories", mock.Anything).Return([]models.Category{}, nil)

	req := httptest.NewRequest("GET", "/categories", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGet(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response CategoriesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(response.Categories))
}

func TestHandlePost_Whitespace(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		categoryName   string
		expectedStatus int
	}{
		{"code_only_spaces", "   ", "Name", http.StatusBadRequest}, // Empty after trim
		{"name_only_spaces", "code", "   ", http.StatusBadRequest}, // Empty after trim
		{"code_with_leading_trailing", "  code  ", "Name", http.StatusCreated},
		{"name_with_leading_trailing", "code", "  Name  ", http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			handler := NewCategoriesHandler(mockRepo)

			if tt.expectedStatus == http.StatusCreated {
				mockRepo.On("CreateCategory", mock.Anything, mock.AnythingOfType("*models.Category")).Return(nil)
			}

			reqBody := CreateCategoryRequest{Code: tt.code, Name: tt.categoryName}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
			recorder := httptest.NewRecorder()

			handler.HandlePost(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}
