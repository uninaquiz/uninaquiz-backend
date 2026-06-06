package mappers

import (
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

func ToQuizHistoryResponse(quiz entities.Quiz) *dto.QuizHistoryResponse {
	return &dto.QuizHistoryResponse{
		ID:         quiz.ID,
		Topic:      quiz.Topic,
		Difficulty: string(quiz.Difficulty),
		Score:      quiz.Score,
		Total:      quiz.Total,
		CreatedAt:  quiz.CreatedAt.UnixMilli(),
	}
}

func ToQuizEntity(cmd commands.SaveQuizHistoryCommand, userID string) *entities.Quiz {
	return &entities.Quiz{
		ID:         cmd.ID,
		UserID:     userID,
		Topic:      cmd.Topic,
		Difficulty: entities.TQuizDifficulty(cmd.Difficulty),
		Score:      cmd.Score,
		Total:      cmd.Total,
		CreatedAt:  time.UnixMilli(cmd.CreatedAt),
	}
}
