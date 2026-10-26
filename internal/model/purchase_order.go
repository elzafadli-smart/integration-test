package model

import (
	"time"

	"github.com/google/uuid"
)

type PurchaseOrder struct {
	ID          uuid.UUID           `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID           `gorm:"type:uuid;not null;index" json:"companyId"`
	OrderNumber string              `gorm:"type:varchar(50);not null;uniqueIndex:idx_company_po_number" json:"orderNumber"`
	VendorName  string              `gorm:"type:varchar(255);not null" json:"vendorName"`
	Status      string              `gorm:"type:varchar(30);not null;default:'DRAFT'" json:"status"`
	TotalAmount float64             `gorm:"type:numeric(15,2);not null;default:0" json:"totalAmount"`
	Notes       string              `gorm:"type:text" json:"notes"`
	Items       []PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt   time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
}

type PurchaseOrderItem struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID       uuid.UUID `gorm:"type:uuid;not null;index" json:"companyId"`
	PurchaseOrderID uuid.UUID `gorm:"type:uuid;not null;index" json:"purchaseOrderId"`
	ProductID       uuid.UUID `gorm:"type:uuid;not null" json:"productId"`
	ProductName     string    `gorm:"type:varchar(255);not null" json:"productName"`
	Quantity        int       `gorm:"type:int;not null" json:"quantity"`
	UnitPrice       float64   `gorm:"type:numeric(15,2);not null" json:"unitPrice"`
	Subtotal        float64   `gorm:"type:numeric(15,2);not null" json:"subtotal"`
}

type CreatePOItemRequest struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unitPrice"`
}

type CreatePurchaseOrderRequest struct {
	VendorName string                `json:"vendorName"`
	Notes      string                `json:"notes"`
	Items      []CreatePOItemRequest `json:"items"`
}

type UpdatePOStatusRequest struct {
	Status string `json:"status"`
}
