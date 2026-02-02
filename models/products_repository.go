package models

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type productsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) ProductRepository {
	return &productsRepository{db: db}
}

func (r *productsRepository) GetAllProducts(ctx context.Context, opts ProductListOptions) ([]Product, int64, error) {
	var products []Product
	var total int64

	// Base query with context
	query := r.db.WithContext(ctx).Model(&Product{})

	// Join category for filtering (more efficient than Preload when filtering)
	query = query.Joins("LEFT JOIN categories ON categories.id = products.category_id")

	// Apply filters
	if opts.CategoryCode != "" {
		query = query.Where("categories.code = ?", opts.CategoryCode)
	}

	if opts.PriceLessThan != nil {
		query = query.Where("products.price < ?", opts.PriceLessThan)
	}

	// Count total BEFORE pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Apply pagination and load associations
	err := query.
		Preload("Category").
		Preload("Variants").
		Offset(opts.Offset).
		Limit(opts.Limit).
		Find(&products).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch products: %w", err)
	}

	return products, total, nil
}

func (r *productsRepository) GetProductByCode(ctx context.Context, code string) (*Product, error) {
	var product Product

	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Variants").
		Where("code = ?", code).
		First(&product).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to fetch product %s: %w", code, err)
	}

	return &product, nil
}
