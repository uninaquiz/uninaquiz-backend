package repositories

import (
	"context"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

type IUserRepository interface {
	Create(ctx context.Context, user entities.User) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	GetAll(ctx context.Context, page int, limit int) ([]entities.User, int64, error)
}
