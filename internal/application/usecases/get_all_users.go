package usecases

import (
	"context"
	"math"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/mappers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/queries"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/repositories"
)

type GetAllUsersUseCase struct {
	userRepository repositories.IUserRepository
}

func NewGetAllUsersUseCase(userRepository repositories.IUserRepository) *GetAllUsersUseCase {
	return &GetAllUsersUseCase{
		userRepository: userRepository,
	}
}

func (usc *GetAllUsersUseCase) Run(ctx context.Context, input queries.GetAllUsersQuery) (*dto.GetAllUsersResponse, error) {
	users, total, err := usc.userRepository.GetAll(ctx, input.Page, input.Limit)
	if err != nil {
		return nil, err
	}

	usersResponse := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		usersResponse = append(usersResponse, *mappers.ToUserResponse(user))
	}

	totalPages := int(math.Ceil(float64(total) / float64(input.Limit)))

	return &dto.GetAllUsersResponse{
		Data:       usersResponse,
		Total:      total,
		Page:       input.Page,
		Limit:      input.Limit,
		TotalPages: totalPages,
	}, nil
}
