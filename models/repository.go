package models

import (
	"context"

	"github.com/shopspring/decimal"
)

// ProductRepository defines product data access
type ProductRepository interface {
	GetAllProducts(ctx context.Context, opts ProductListOptions) ([]Product, int64, error)
	GetProductByCode(ctx context.Context, code string) (*Product, error)
}

// CategoryRepository defines category data access
type CategoryRepository interface {
	GetAllCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, category *Category) error
	GetCategoryByCode(ctx context.Context, code string) (*Category, error)
}

// ProductListOptions encapsulates query parameters
type ProductListOptions struct {
	Offset        int
	Limit         int
	CategoryCode  string
	PriceLessThan *decimal.Decimal
}
