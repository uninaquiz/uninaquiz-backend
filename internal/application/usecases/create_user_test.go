package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/usecases"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestCreateUserUseCase_Run(t *testing.T) {
	var (
		errDB    = errors.New("db error")
		errHash  = errors.New("hash error")
		errToken = errors.New("token error")
		fakeUser = &entities.User{
			ID:        "uuid-fake-1",
			Username:  "john_doe",
			Password:  "$2a$hashed",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	)

	tests := []struct {
		name         string
		input        commands.CreateUserCommand
		mockBehavior func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService)
		wantErr      error
		wantToken    bool
	}{
		{
			name:  "success — new user registered and JWT issued",
			input: commands.CreateUserCommand{Username: "john_doe", Password: "secret123"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().ExistsByUsername(gomock.Any(), "john_doe").Return(false, nil)
				hasher.EXPECT().HashPassword("secret123").Return("$2a$hashed", nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fakeUser, nil)
				token.EXPECT().GenerateToken("uuid-fake-1", "john_doe", 7*24*time.Hour).Return("jwt.token.here", nil)
			},
			wantErr:   nil,
			wantToken: true,
		},
		{
			name:  "duplicate username → ErrUserAlreadyExists",
			input: commands.CreateUserCommand{Username: "existing_user", Password: "password1"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().ExistsByUsername(gomock.Any(), "existing_user").Return(true, nil)
			},
			wantErr: domainerrors.ErrUserAlreadyExists,
		},
		{
			name:  "ExistsByUsername db error → propagated",
			input: commands.CreateUserCommand{Username: "user_a", Password: "password1"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().ExistsByUsername(gomock.Any(), "user_a").Return(false, errDB)
			},
			wantErr: errDB,
		},
		{
			name:  "HashPassword fails → propagated",
			input: commands.CreateUserCommand{Username: "user_b", Password: "password2"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().ExistsByUsername(gomock.Any(), "user_b").Return(false, nil)
				hasher.EXPECT().HashPassword("password2").Return("", errHash)
			},
			wantErr: errHash,
		},
		{
			name:  "Create repository fails → propagated",
			input: commands.CreateUserCommand{Username: "user_c", Password: "password3"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().ExistsByUsername(gomock.Any(), "user_c").Return(false, nil)
				hasher.EXPECT().HashPassword("password3").Return("$2a$hashed", nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name:  "GenerateToken fails → propagated",
			input: commands.CreateUserCommand{Username: "user_d", Password: "password4"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().ExistsByUsername(gomock.Any(), "user_d").Return(false, nil)
				hasher.EXPECT().HashPassword("password4").Return("$2a$hashed", nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fakeUser, nil)
				token.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errToken)
			},
			wantErr: errToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := mocks.NewMockIUserRepository(ctrl)
			hasherMock := mocks.NewMockIHasher(ctrl)
			tokenMock := mocks.NewMockITokenService(ctrl)

			tt.mockBehavior(repoMock, hasherMock, tokenMock)

			uc := usecases.NewCreateUserUseCase(repoMock, hasherMock, tokenMock)
			resp, err := uc.Run(context.Background(), tt.input, 7*24*time.Hour)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				}
				if resp != nil {
					t.Error("Run() expected nil response on error, got non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Run() unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("Run() expected non-nil response on success")
			}
			if tt.wantToken && resp.Token == "" {
				t.Error("Run() expected non-empty token in response")
			}
		})
	}
}
