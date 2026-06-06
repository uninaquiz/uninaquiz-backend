package repositories

import (
	"context"
	"errors"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/database/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (usr *UserRepository) Create(ctx context.Context, user entities.User) (*entities.User, error) {
	model := models.UserToModel(user)
	if err := usr.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (usr *UserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	var model models.UserModel
	if err := usr.db.WithContext(ctx).Where("username = ?", username).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (usr *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := usr.db.WithContext(ctx).Model(&models.UserModel{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (usr *UserRepository) GetAll(ctx context.Context, page int, limit int) ([]entities.User, int64, error) {
	var userModels []models.UserModel
	var total int64

	if err := usr.db.WithContext(ctx).Model(&models.UserModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := usr.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&userModels).Error; err != nil {
		return nil, 0, err
	}

	users := make([]entities.User, 0, len(userModels))
	for _, m := range userModels {
		users = append(users, *m.ToDomain())
	}

	return users, total, nil
}
