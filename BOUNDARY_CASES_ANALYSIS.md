# Boundary Cases Analysis

## Current Test Coverage

### ✅ Covered Boundary Cases

#### Pagination
- ✅ Normal pagination (offset=5, limit=20)
- ✅ Invalid limit (non-numeric: "abc")

#### Product Retrieval
- ✅ Product not found (404)
- ✅ Variant price inheritance (nil price)
- ✅ Variant with custom price

#### Category Creation
- ✅ Empty code field
- ✅ Empty name field
- ✅ Duplicate code (409 Conflict)

#### Filtering
- ✅ Category filter
- ✅ Price filter
- ✅ Combined filters
- ✅ Invalid price format
- ✅ Negative price

---

## ❌ Missing Boundary Cases

### HIGH Priority

#### 1. Pagination Boundaries
- ❌ **limit = 0** (should error: below minimum)
- ❌ **limit = 1** (minimum valid value)
- ❌ **limit = 100** (maximum valid value)
- ❌ **limit = 101** (should cap at 100)
- ❌ **limit = 999** (should cap at 100)
- ❌ **offset = -1** (negative offset should error)
- ❌ **offset > total products** (should return empty array)
- ❌ **Missing offset/limit** (should use defaults)

#### 2. Field Length Validation
- ❌ **Category code exactly 50 chars** (max valid)
- ❌ **Category code 51 chars** (should error)
- ❌ **Category name exactly 255 chars** (max valid)
- ❌ **Category name 256 chars** (should error)

#### 3. Price Boundaries
- ❌ **price_less_than = 0** (edge case)
- ❌ **price_less_than = 0.01** (very small)
- ❌ **price_less_than = 999999.99** (very large)
- ❌ **Product with price = 0** (free item)

#### 4. Empty Results
- ❌ **GET /catalog with no matching filters** (empty array)
- ❌ **GET /categories when none exist** (empty array)
- ❌ **Product with 0 variants**

#### 5. Special Characters
- ❌ **Category code with special chars** (@, #, spaces)
- ❌ **Category name with unicode** (émojis, 中文)
- ❌ **SQL injection attempt in filters**

### MEDIUM Priority

#### 6. Decimal Precision
- ❌ **Price with many decimals** (19.99999999)
- ❌ **Price rounding** (19.995 → 20.00 or 19.99?)
- ❌ **Very large price** (999999999.99)

#### 7. Empty String vs Null
- ❌ **Product with null category** (no category assigned)
- ❌ **Variant with price = 0 vs null** (different meanings?)

#### 8. Concurrent Operations
- ❌ **Two users creating same category code simultaneously**
- ❌ **Reading while writing**

### LOW Priority

#### 9. Malformed Requests
- ❌ **Invalid JSON body**
- ❌ **Missing Content-Type header**
- ❌ **Extra unknown fields in JSON**

#### 10. Database Connection
- ❌ **Database connection lost during query**
- ❌ **Query timeout**

---

## Recommended Additional Tests

### Immediate (Add to PR)

```go
// catalog/handler_test.go

func TestHandleGet_LimitBoundaries(t *testing.T) {
    tests := []struct {
        name           string
        limit          string
        expectedLimit  int
        expectedStatus int
    }{
        {"limit_zero", "0", 0, http.StatusBadRequest},
        {"limit_one", "1", 1, http.StatusOK},
        {"limit_max", "100", 100, http.StatusOK},
        {"limit_above_max", "101", 100, http.StatusOK}, // Capped
        {"limit_very_large", "999", 100, http.StatusOK}, // Capped
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

func TestHandleGet_OffsetBoundaries(t *testing.T) {
    // Test offset=-1 (should error)
    // Test offset=0 (valid)
    // Test offset > total (empty results)
}

func TestHandleGet_EmptyResults(t *testing.T) {
    // Test with filter that matches nothing
    // Should return empty array, not error
}

func TestHandleGetByCode_ProductWithNoVariants(t *testing.T) {
    // Product exists but has empty variants array
}

func TestHandleGetByCode_EmptyCode(t *testing.T) {
    // Request with empty code parameter
    // Should return 400 Bad Request
}
```

```go
// categories/handler_test.go

func TestHandlePost_FieldLengthBoundaries(t *testing.T) {
    tests := []struct {
        name          string
        code          string
        categoryName  string
        expectedStatus int
    }{
        {"code_max_length", strings.Repeat("a", 50), "Valid", http.StatusCreated},
        {"code_too_long", strings.Repeat("a", 51), "Valid", http.StatusBadRequest},
        {"name_max_length", "valid", strings.Repeat("a", 255), http.StatusCreated},
        {"name_too_long", "valid", strings.Repeat("a", 256), http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

func TestHandlePost_SpecialCharacters(t *testing.T) {
    // Test unicode, emojis, special chars
}

func TestHandlePost_InvalidJSON(t *testing.T) {
    // Test malformed JSON body
}

func TestHandleGet_EmptyCategories(t *testing.T) {
    // Repository returns empty array
    // Should return 200 with empty array, not 404
}
```

### Integration Tests

```go
// test/integration/api_test.go

func TestGetCatalog_PriceBoundaries(t *testing.T) {
    // price_less_than=0
    // price_less_than=0.01
    // price_less_than=999999.99
}

func TestGetCatalog_OffsetExceedsTotal(t *testing.T) {
    // Get total count
    // Request with offset > total
    // Should return empty array
}

func TestDecimalPrecision(t *testing.T) {
    // Create product with price 19.99999999
    // Verify precision is maintained
}
```

---

## Risk Assessment

| Boundary Case | Risk Level | Impact if Missed |
|--------------|------------|------------------|
| Limit = 0 | HIGH | Could cause DB to return all records |
| Offset < 0 | HIGH | Could cause DB errors or wrong results |
| Field length > DB limit | HIGH | DB constraint violation error |
| SQL injection | CRITICAL | Security vulnerability |
| Offset > total | MEDIUM | Poor UX, but not breaking |
| Price = 0 | MEDIUM | Business logic issue |
| Empty results | LOW | UX issue only |
| Unicode handling | LOW | Display issues |

---

## Test Coverage Metrics

### Current Coverage
- **Unit Tests**: 77-78% statement coverage
- **Boundary Cases**: ~40% coverage
- **Edge Cases**: ~30% coverage

### Target Coverage
- **Unit Tests**: 80%+ statement coverage
- **Boundary Cases**: 80%+ coverage
- **Edge Cases**: 60%+ coverage

---

## Action Items

### For Current PR
1. Add limit boundary tests (0, 1, 100, 101)
2. Add negative offset test
3. Add field length validation tests
4. Add empty result tests

### For Follow-up PR
1. Add special character handling tests
2. Add decimal precision tests
3. Add concurrent operation tests
4. Add malformed request tests

### For Production
1. Add monitoring for boundary case errors
2. Add alerts for unexpected edge cases
3. Add logging for validation failures
