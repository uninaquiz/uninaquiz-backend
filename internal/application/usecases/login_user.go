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

type LoginUserUseCase struct {
	userRepository repositories.IUserRepository
	hasher         services.IHasher
	tokenService   services.ITokenService
}

func NewLoginUserUseCase(userRepository repositories.IUserRepository, hasher services.IHasher, tokenService services.ITokenService) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepository: userRepository,
		hasher:         hasher,
		tokenService:   tokenService,
	}
}

func (usc *LoginUserUseCase) Run(ctx context.Context, input commands.LoginCommand, expirationTime time.Duration) (*dto.LoginResponse, error) {
	user, err := usc.userRepository.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	if !usc.hasher.ComparePassword(input.Password, user.Password) {
		return nil, domainerrors.ErrInvalidCredentials
	}

	token, err := usc.tokenService.GenerateToken(user.ID, user.Username, expirationTime)
	if err != nil {
		return nil, err
	}

	userResponse := mappers.ToUserResponse(*user)
	return &dto.LoginResponse{Token: token, User: *userResponse}, nil
}
