package model

import (
	"time"

	"github.com/google/uuid"
)

type SalesOrder struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"companyId"`
	CustomerID  uuid.UUID `gorm:"type:uuid;not null" json:"customerId"`
	OrderNumber string    `gorm:"type:varchar(100);not null" json:"orderNumber"`
	TotalAmount float64   `gorm:"type:numeric(15,2);not null;default:0" json:"totalAmount"`
	Status      string    `gorm:"type:varchar(50);not null;default:'DRAFT'" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type CreateSalesOrderRequest struct {
	CustomerID  uuid.UUID `json:"customerId"`
	OrderNumber string    `json:"orderNumber"`
	TotalAmount float64   `json:"totalAmount"`
}

type UpdateSalesOrderStatusRequest struct {
	Status string `json:"status"`
}
