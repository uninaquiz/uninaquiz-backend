package usecases

import (
	"context"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/mappers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/services"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/repositories"
)

type CreateUserUseCase struct {
	userRepository repositories.IUserRepository
	hasher         services.IHasher
	tokenService   services.ITokenService
}

func NewCreateUserUseCase(userRepository repositories.IUserRepository, hasher services.IHasher, tokenService services.ITokenService) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepository: userRepository,
		hasher:         hasher,
		tokenService:   tokenService,
	}
}

func (usc *CreateUserUseCase) Run(ctx context.Context, input commands.CreateUserCommand, expirationTime time.Duration) (*dto.CreateUserResponse, error) {
	exists, err := usc.userRepository.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainerrors.ErrUserAlreadyExists
	}

	userDomain := mappers.ToUserEntity(input)
	passwordHash, err := usc.hasher.HashPassword(userDomain.Password)
	if err != nil {
		return nil, err
	}
	userDomain.Password = passwordHash

	userCreated, err := usc.userRepository.Create(ctx, *userDomain)
	if err != nil {
		return nil, err
	}

	token, err := usc.tokenService.GenerateToken(userCreated.ID, userCreated.Username, expirationTime)
	if err != nil {
		return nil, err
	}

	response := mappers.ToUserResponse(*userCreated)
	return &dto.CreateUserResponse{User: *response, Token: token}, nil
}
