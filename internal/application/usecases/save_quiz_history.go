package usecases

import (
	"context"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/mappers"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/repositories"
)

type SaveQuizHistoryUseCase struct {
	quizRepository repositories.IQuizRepository
}

func NewSaveQuizHistoryUseCase(quizRepository repositories.IQuizRepository) *SaveQuizHistoryUseCase {
	return &SaveQuizHistoryUseCase{
		quizRepository: quizRepository,
	}
}

func (usc *SaveQuizHistoryUseCase) Run(ctx context.Context, input commands.SaveQuizHistoryCommand, userID string) error {
	exists, err := usc.quizRepository.ExistsByUserTopicAndDifficulty(ctx, userID, input.Topic, input.Difficulty)
	if err != nil {
		return err
	}
	if exists {
		return domainerrors.ErrQuizAlreadyExists
	}

	quiz := mappers.ToQuizEntity(input, userID)
	_, err = usc.quizRepository.Create(ctx, *quiz)
	return err
}
