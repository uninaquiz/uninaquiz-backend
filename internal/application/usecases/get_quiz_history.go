package usecases

import (
	"context"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/mappers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/repositories"
)

type GetQuizHistoryUseCase struct {
	quizRepository repositories.IQuizRepository
}

func NewGetQuizHistoryUseCase(quizRepository repositories.IQuizRepository) *GetQuizHistoryUseCase {
	return &GetQuizHistoryUseCase{
		quizRepository: quizRepository,
	}
}

func (usc *GetQuizHistoryUseCase) Run(ctx context.Context, userID string) ([]dto.QuizHistoryResponse, error) {
	quizzes, err := usc.quizRepository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.QuizHistoryResponse, 0, len(quizzes))
	for _, quiz := range quizzes {
		response = append(response, *mappers.ToQuizHistoryResponse(quiz))
	}

	return response, nil
}
