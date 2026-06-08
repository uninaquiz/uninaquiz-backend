package repositories

import (
	"context"
	"errors"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/database/models"
	"gorm.io/gorm"
)

type QuizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) *QuizRepository {
	return &QuizRepository{db: db}
}

func (r *QuizRepository) Create(ctx context.Context, quiz entities.Quiz) (*entities.Quiz, error) {
	model := models.QuizToModel(quiz)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *QuizRepository) FindAllByUserID(ctx context.Context, userID string) ([]entities.Quiz, error) {
	var quizModels []models.QuizModel
	if err := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&quizModels).Error; err != nil {
		return nil, err
	}

	quizzes := make([]entities.Quiz, 0, len(quizModels))
	for _, m := range quizModels {
		quizzes = append(quizzes, *m.ToDomain())
	}
	return quizzes, nil
}

func (r *QuizRepository) FindByID(ctx context.Context, id string) (*entities.Quiz, error) {
	var model models.QuizModel
	if err := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Where("id = ?", id).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *QuizRepository) FindByUserTopicAndDifficulty(ctx context.Context, userID, topic, difficulty string) (*entities.Quiz, error) {
	var model models.QuizModel
	err := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Where("user_id = ? AND LOWER(topic) = LOWER(?) AND difficulty = ?", userID, topic, difficulty).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *QuizRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.QuizModel{}).Error
}

func (r *QuizRepository) UpdateScore(ctx context.Context, id string, score int) error {
	return r.db.WithContext(ctx).
		Model(&models.QuizModel{}).
		Where("id = ?", id).
		Update("score", score).Error
}
