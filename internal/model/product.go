package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"companyId"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	SKU       string    `gorm:"type:varchar(100);not null" json:"sku"`
	Barcode   string    `gorm:"type:varchar(100);default:''" json:"barcode"`
	Price     float64   `gorm:"type:numeric(15,2);not null;default:0" json:"price"`
	Stock     int       `gorm:"type:int;not null;default:0" json:"stock"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type CreateProductRequest struct {
	Name    string  `json:"name"`
	SKU     string  `json:"sku"`
	Barcode string  `json:"barcode"`
	Price   float64 `json:"price"`
	Stock   int     `json:"stock"`
}

type UpdateProductRequest struct {
	Name    *string  `json:"name"`
	Barcode *string  `json:"barcode"`
	Price   *float64 `json:"price"`
	Stock   *int     `json:"stock"`
}
