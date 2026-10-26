package model

import (
	"time"

	"github.com/google/uuid"
)

type Supplier struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID     uuid.UUID `gorm:"type:uuid;not null;index" json:"companyId"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	ContactPerson string    `gorm:"type:varchar(255)" json:"contactPerson"`
	Phone         string    `gorm:"type:varchar(50)" json:"phone"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type CreateSupplierRequest struct {
	Name          string `json:"name"`
	ContactPerson string `json:"contactPerson"`
	Phone         string `json:"phone"`
}

type UpdateSupplierRequest struct {
	Name          *string `json:"name"`
	ContactPerson *string `json:"contactPerson"`
	Phone         *string `json:"phone"`
}
