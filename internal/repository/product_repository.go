package repository

import (
	"context"
	"errors"

	"be-integration-example/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.Product, error)
	FindAll(ctx context.Context, companyID uuid.UUID) ([]model.Product, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindAll(ctx context.Context, companyID uuid.UUID) ([]model.Product, error) {
	var products []model.Product
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&products).Error
	return products, err
}

func (r *productRepository) Update(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		Delete(&model.Product{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
