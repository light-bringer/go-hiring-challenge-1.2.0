package models

import "time"

type Category struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"uniqueIndex;not null"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *Category) TableName() string {
	return "categories"
}
