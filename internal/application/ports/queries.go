package ports

import (
	"context"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/queries"
)

type IGetAllUsersPort interface {
	Run(ctx context.Context, input queries.GetAllUsersQuery) (*dto.GetAllUsersResponse, error)
}

type IGetQuizHistoryPort interface {
	Run(ctx context.Context, userID string) ([]dto.QuizHistoryResponse, error)
}
