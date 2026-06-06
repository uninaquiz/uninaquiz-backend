package usecases

import (
	"context"

	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/repositories"
)

type DeleteQuizHistoryUseCase struct {
	quizRepository repositories.IQuizRepository
}

func NewDeleteQuizHistoryUseCase(quizRepository repositories.IQuizRepository) *DeleteQuizHistoryUseCase {
	return &DeleteQuizHistoryUseCase{
		quizRepository: quizRepository,
	}
}

func (usc *DeleteQuizHistoryUseCase) Run(ctx context.Context, id string, userID string) error {
	quiz, err := usc.quizRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if quiz == nil {
		return domainerrors.ErrQuizNotFound
	}

	if quiz.UserID != userID {
		return domainerrors.ErrQuizForbidden
	}

	return usc.quizRepository.Delete(ctx, id)
}
