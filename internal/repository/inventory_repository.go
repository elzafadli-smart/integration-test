package repository

import (
	"context"
	"errors"

	"be-integration-example/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InventoryRepository interface {
	Adjust(ctx context.Context, inventory *model.Inventory) error
	FindByProduct(ctx context.Context, companyID, productID uuid.UUID) ([]model.Inventory, error)
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) Adjust(ctx context.Context, inventory *model.Inventory) error {
	var existing model.Inventory
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND product_id = ? AND warehouse = ?", inventory.CompanyID, inventory.ProductID, inventory.Warehouse).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.WithContext(ctx).Create(inventory).Error
		}
		return err
	}
	existing.Quantity += inventory.Quantity
	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *inventoryRepository) FindByProduct(ctx context.Context, companyID, productID uuid.UUID) ([]model.Inventory, error) {
	var items []model.Inventory
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND product_id = ?", companyID, productID).
		Find(&items).Error
	return items, err
}

func (r *inventoryRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		Delete(&model.Inventory{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
