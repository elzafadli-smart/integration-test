package repository

import (
	"context"
	"errors"

	"be-integration-example/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *model.Supplier) error
	FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.Supplier, error)
	FindAll(ctx context.Context, companyID uuid.UUID) ([]model.Supplier, error)
	Update(ctx context.Context, supplier *model.Supplier) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *model.Supplier) error {
	return r.db.WithContext(ctx).Create(supplier).Error
}

func (r *supplierRepository) FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.Supplier, error) {
	var supplier model.Supplier
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		First(&supplier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &supplier, nil
}

func (r *supplierRepository) FindAll(ctx context.Context, companyID uuid.UUID) ([]model.Supplier, error) {
	var suppliers []model.Supplier
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&suppliers).Error
	return suppliers, err
}

func (r *supplierRepository) Update(ctx context.Context, supplier *model.Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}

func (r *supplierRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		Delete(&model.Supplier{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
