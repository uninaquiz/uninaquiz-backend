package ports

import (
	"context"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
)

type ICreateUserPort interface {
	Run(ctx context.Context, input commands.CreateUserCommand, expirationTime time.Duration) (*dto.CreateUserResponse, error)
}

type ILoginUserPort interface {
	Run(ctx context.Context, input commands.LoginCommand, expirationTime time.Duration) (*dto.LoginResponse, error)
}

type ISaveQuizHistoryPort interface {
	Run(ctx context.Context, input commands.SaveQuizHistoryCommand, userID string) error
}

type IDeleteQuizHistoryPort interface {
	Run(ctx context.Context, id string, userID string) error
}

type IGenerateQuizPort interface {
	Run(ctx context.Context, input commands.GenerateQuizCommand, userID string) (*dto.GenerateQuizResponse, error)
}
