package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8485"
	testPort = "8485"
)

var (
	srv        *http.Server
	httpClient *http.Client
)

func TestMain(m *testing.M) {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		// In Docker, we won't have .env file, so just continue
	}

	// Initialize database connection
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	db, close := database.NewWithHost(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
		host,
	)
	defer close()

	// Set up connection pool
	sqlDB, err := db.DB()
	if err != nil {
		panic("Failed to get database connection: " + err.Error())
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)

	// Initialize repositories
	prodRepo := models.NewProductsRepository(db)
	catRepo := models.NewCategoriesRepository(db)

	// Initialize services
	catalogService := catalog.NewCatalogService(prodRepo)

	// Initialize handlers
	catalogHandler := catalog.NewCatalogHandler(catalogService)
	categoriesHandler := categories.NewCategoriesHandler(catRepo)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", catalogHandler.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", catalogHandler.HandleGetByCode)
	mux.HandleFunc("GET /categories", categoriesHandler.HandleGet)
	mux.HandleFunc("POST /categories", categoriesHandler.HandlePost)

	// Set up the HTTP server
	srv = &http.Server{
		Addr:    fmt.Sprintf("localhost:%s", testPort),
		Handler: mux,
	}

	// Start the server in background
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("Server failed: " + err.Error())
		}
	}()

	// Wait for server to be ready
	httpClient = &http.Client{Timeout: 10 * time.Second}
	time.Sleep(500 * time.Millisecond)

	// Run tests
	code := m.Run()

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	os.Exit(code)
}

// Helper function to make GET requests
func get(t *testing.T, path string) *http.Response {
	resp, err := httpClient.Get(baseURL + path)
	require.NoError(t, err)
	return resp
}

// Helper function to make POST requests
func post(t *testing.T, path string, body interface{}) *http.Response {
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := httpClient.Post(baseURL+path, "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	return resp
}

// Helper function to decode JSON response
func decodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, v))
}

func TestGetCatalog_Pagination(t *testing.T) {
	resp := get(t, "/catalog?limit=3")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result catalog.CatalogResponse
	decodeJSON(t, resp, &result)

	assert.Equal(t, 3, len(result.Products))
	assert.Equal(t, 0, result.Offset)
	assert.Equal(t, 3, result.Limit)
	assert.GreaterOrEqual(t, result.Total, int64(3))
}

func TestGetCatalog_CategoryFilter(t *testing.T) {
	resp := get(t, "/catalog?category=shoes")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result catalog.CatalogResponse
	decodeJSON(t, resp, &result)

	assert.GreaterOrEqual(t, len(result.Products), 1)
	for _, product := range result.Products {
		assert.NotNil(t, product.Category)
		assert.Equal(t, "shoes", product.Category.Code)
	}
}

func TestGetCatalog_PriceFilter(t *testing.T) {
	resp := get(t, "/catalog?price_less_than=10")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result catalog.CatalogResponse
	decodeJSON(t, resp, &result)

	for _, product := range result.Products {
		assert.Less(t, product.Price, 10.0)
	}
}

func TestGetCatalog_CombinedFilters(t *testing.T) {
	resp := get(t, "/catalog?category=clothing&price_less_than=15")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result catalog.CatalogResponse
	decodeJSON(t, resp, &result)

	for _, product := range result.Products {
		assert.NotNil(t, product.Category)
		assert.Equal(t, "clothing", product.Category.Code)
		assert.Less(t, product.Price, 15.0)
	}
}

func TestGetCatalog_InvalidParameters(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"invalid limit", "/catalog?limit=abc"},
		{"invalid offset", "/catalog?offset=-1"},
		{"invalid price", "/catalog?price_less_than=abc"},
		{"negative price", "/catalog?price_less_than=-10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := get(t, tt.path)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestGetProductByCode_Success(t *testing.T) {
	resp := get(t, "/catalog/PROD001")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result catalog.ProductDetailDTO
	decodeJSON(t, resp, &result)

	assert.Equal(t, "PROD001", result.Code)
	assert.Greater(t, result.Price, 0.0)
	assert.NotNil(t, result.Category)
	assert.GreaterOrEqual(t, len(result.Variants), 1)

	// Verify variants have prices (either their own or inherited)
	for _, variant := range result.Variants {
		assert.Greater(t, variant.Price, 0.0)
		assert.NotEmpty(t, variant.SKU)
		assert.NotEmpty(t, variant.Name)
	}
}

func TestGetProductByCode_NotFound(t *testing.T) {
	resp := get(t, "/catalog/NONEXISTENT")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var errResp struct {
		Error string `json:"error"`
	}
	decodeJSON(t, resp, &errResp)
	assert.Contains(t, errResp.Error, "not found")
}

func TestGetCategories(t *testing.T) {
	resp := get(t, "/categories")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result categories.CategoriesResponse
	decodeJSON(t, resp, &result)

	assert.GreaterOrEqual(t, len(result.Categories), 3)

	// Verify expected categories exist
	codes := make(map[string]bool)
	for _, cat := range result.Categories {
		codes[cat.Code] = true
		assert.NotEmpty(t, cat.Name)
	}

	assert.True(t, codes["clothing"])
	assert.True(t, codes["shoes"])
	assert.True(t, codes["accessories"])
}

func TestPostCategory_Success(t *testing.T) {
	// Generate unique code to avoid conflicts
	uniqueCode := fmt.Sprintf("test-%d", time.Now().Unix())

	reqBody := categories.CreateCategoryRequest{
		Code: uniqueCode,
		Name: "Test Category",
	}

	resp := post(t, "/categories", reqBody)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result categories.CategoryDTO
	decodeJSON(t, resp, &result)

	assert.Equal(t, uniqueCode, result.Code)
	assert.Equal(t, "Test Category", result.Name)

	// Verify it appears in the list
	listResp := get(t, "/categories")
	var listResult categories.CategoriesResponse
	decodeJSON(t, listResp, &listResult)

	found := false
	for _, cat := range listResult.Categories {
		if cat.Code == uniqueCode {
			found = true
			break
		}
	}
	assert.True(t, found, "Newly created category should appear in list")
}

func TestPostCategory_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body categories.CreateCategoryRequest
	}{
		{"missing code", categories.CreateCategoryRequest{Code: "", Name: "Test"}},
		{"missing name", categories.CreateCategoryRequest{Code: "test", Name: ""}},
		{"both missing", categories.CreateCategoryRequest{Code: "", Name: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := post(t, "/categories", tt.body)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestPostCategory_DuplicateCode(t *testing.T) {
	reqBody := categories.CreateCategoryRequest{
		Code: "clothing",
		Name: "Duplicate Clothing",
	}

	resp := post(t, "/categories", reqBody)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var errResp struct {
		Error string `json:"error"`
	}
	decodeJSON(t, resp, &errResp)
	assert.Contains(t, errResp.Error, "already exists")
}

func TestEndToEnd_ProductWithCategory(t *testing.T) {
	// 1. Get initial catalog
	resp1 := get(t, "/catalog")
	var initialCatalog catalog.CatalogResponse
	decodeJSON(t, resp1, &initialCatalog)

	// 2. Get a specific product
	if len(initialCatalog.Products) > 0 {
		productCode := initialCatalog.Products[0].Code

		resp2 := get(t, "/catalog/"+productCode)
		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		var productDetail catalog.ProductDetailDTO
		decodeJSON(t, resp2, &productDetail)

		assert.Equal(t, productCode, productDetail.Code)

		// 3. If product has a category, verify it exists in categories list
		if productDetail.Category != nil {
			resp3 := get(t, "/categories")
			var categoriesList categories.CategoriesResponse
			decodeJSON(t, resp3, &categoriesList)

			found := false
			for _, cat := range categoriesList.Categories {
				if cat.Code == productDetail.Category.Code {
					found = true
					assert.Equal(t, productDetail.Category.Name, cat.Name)
					break
				}
			}
			assert.True(t, found, "Product's category should exist in categories list")
		}
	}
}
