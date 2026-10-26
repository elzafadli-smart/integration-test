package repository

import (
	"context"
	"errors"

	"be-integration-example/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseOrderRepository interface {
	Create(ctx context.Context, po *model.PurchaseOrder) error
	FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.PurchaseOrder, error)
	FindAll(ctx context.Context, companyID uuid.UUID) ([]model.PurchaseOrder, error)
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status string) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

type purchaseOrderRepository struct {
	db *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &purchaseOrderRepository{db: db}
}

func (r *purchaseOrderRepository) Create(ctx context.Context, po *model.PurchaseOrder) error {
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *purchaseOrderRepository) FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("company_id = ? AND id = ?", companyID, id).
		First(&po).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &po, nil
}

func (r *purchaseOrderRepository) FindAll(ctx context.Context, companyID uuid.UUID) ([]model.PurchaseOrder, error) {
	var pos []model.PurchaseOrder
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&pos).Error
	return pos, err
}

func (r *purchaseOrderRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).
		Model(&model.PurchaseOrder{}).
		Where("company_id = ? AND id = ?", companyID, id).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *purchaseOrderRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete line items first
		if err := tx.Where("company_id = ? AND purchase_order_id = ?", companyID, id).
			Delete(&model.PurchaseOrderItem{}).Error; err != nil {
			return err
		}

		res := tx.Where("company_id = ? AND id = ?", companyID, id).
			Delete(&model.PurchaseOrder{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
