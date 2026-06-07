package usecases

import (
	"context"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/mappers"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/repositories"
)

type GetQuizUseCase struct {
	quizRepository repositories.IQuizRepository
}

func NewGetQuizUseCase(quizRepository repositories.IQuizRepository) *GetQuizUseCase {
	return &GetQuizUseCase{
		quizRepository: quizRepository,
	}
}

func (usc *GetQuizUseCase) Run(ctx context.Context, quizID, userID string) (*dto.GetQuizResponse, error) {
	quiz, err := usc.quizRepository.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if quiz == nil {
		return nil, domainerrors.ErrQuizNotFound
	}

	if quiz.UserID != userID {
		return nil, domainerrors.ErrQuizForbidden
	}

	return mappers.ToGetQuizResponse(*quiz), nil
}
