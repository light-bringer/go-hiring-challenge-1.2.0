package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrDuplicateCode    = errors.New("category code already exists")
)

type categoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) CategoryRepository {
	return &categoriesRepository{db: db}
}

func (r *categoriesRepository) GetAllCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	return categories, nil
}

func (r *categoriesRepository) CreateCategory(ctx context.Context, category *Category) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		// Check for unique constraint violation (PostgreSQL error code 23505)
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
		   strings.Contains(err.Error(), "duplicate key") ||
		   strings.Contains(err.Error(), "23505") {
			return ErrDuplicateCode
		}
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *categoriesRepository) GetCategoryByCode(ctx context.Context, code string) (*Category, error) {
	var category Category
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}
	return &category, nil
}
