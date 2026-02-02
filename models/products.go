package models

import (
	"time"
	"github.com/shopspring/decimal"
)

// Product represents a product in the catalog.
// It includes a unique code and a price.
type Product struct {
	ID         uint             `gorm:"primaryKey"`
	Code       string           `gorm:"uniqueIndex;not null"`
	Price      decimal.Decimal  `gorm:"type:decimal(10,2);not null"`
	CategoryID *uint            `gorm:"index"`
	Category   *Category        `gorm:"foreignKey:CategoryID"`
	Variants   []Variant        `gorm:"foreignKey:ProductID"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (p *Product) TableName() string {
	return "products"
}
