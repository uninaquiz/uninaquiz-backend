package services

import (
	"context"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
)

type IAIService interface {
	GenerateQuiz(ctx context.Context, topic string, difficulty string) ([]dto.QuizQuestion, error)
}
