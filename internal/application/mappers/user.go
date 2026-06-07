package mappers

import (
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	"github.com/google/uuid"
)

func ToUserResponse(user entities.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToUserEntity(cmd commands.CreateUserCommand) *entities.User {
	return &entities.User{
		ID:       uuid.New().String(),
		Username: cmd.Username,
		Password: cmd.Password,
	}
}
