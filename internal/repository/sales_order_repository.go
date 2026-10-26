package repository

import (
	"context"
	"errors"

	"be-integration-example/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesOrderRepository interface {
	Create(ctx context.Context, order *model.SalesOrder) error
	FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.SalesOrder, error)
	FindAll(ctx context.Context, companyID uuid.UUID) ([]model.SalesOrder, error)
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status string) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

type salesOrderRepository struct {
	db *gorm.DB
}

func NewSalesOrderRepository(db *gorm.DB) SalesOrderRepository {
	return &salesOrderRepository{db: db}
}

func (r *salesOrderRepository) Create(ctx context.Context, order *model.SalesOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *salesOrderRepository) FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.SalesOrder, error) {
	var order model.SalesOrder
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *salesOrderRepository) FindAll(ctx context.Context, companyID uuid.UUID) ([]model.SalesOrder, error) {
	var orders []model.SalesOrder
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *salesOrderRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).
		Model(&model.SalesOrder{}).
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

func (r *salesOrderRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		Delete(&model.SalesOrder{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
