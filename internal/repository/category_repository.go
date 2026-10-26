package repository

import (
	"context"
	"errors"

	"be-integration-example/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.Category, error)
	FindAll(ctx context.Context, companyID uuid.UUID) ([]model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) FindByID(ctx context.Context, companyID, id uuid.UUID) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindAll(ctx context.Context, companyID uuid.UUID) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		Delete(&model.Category{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
