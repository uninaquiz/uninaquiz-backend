package mappers

import (
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
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
		Username: cmd.Username,
		Password: cmd.Password,
	}
}
