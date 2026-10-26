package model

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"companyId"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"productId"`
	Warehouse string    `gorm:"type:varchar(100);not null" json:"warehouse"`
	Quantity  int       `gorm:"type:int;not null;default:0" json:"quantity"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type AdjustInventoryRequest struct {
	ProductID uuid.UUID `json:"productId"`
	Warehouse string    `json:"warehouse"`
	Quantity  int       `json:"quantity"`
}
