package usecases

import (
	"context"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
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
	quiz, err := usc.quizRepository.FindByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if quiz == nil {
		return domainerrors.ErrQuizNotFound
	}
	if quiz.UserID != userID {
		return domainerrors.ErrQuizForbidden
	}

	return usc.quizRepository.UpdateScore(ctx, input.ID, input.Score)
}
