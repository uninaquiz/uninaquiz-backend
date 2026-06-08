package repositories

import (
	"context"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

type IQuizRepository interface {
	Create(ctx context.Context, quiz entities.Quiz) (*entities.Quiz, error)
	FindAllByUserID(ctx context.Context, userID string) ([]entities.Quiz, error)
	FindByID(ctx context.Context, id string) (*entities.Quiz, error)
	FindByUserTopicAndDifficulty(ctx context.Context, userID, topic, difficulty string) (*entities.Quiz, error)
	UpdateScore(ctx context.Context, id string, score int) error
	Delete(ctx context.Context, id string) error
}
