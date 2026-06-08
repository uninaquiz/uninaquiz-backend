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

func TestLoginUserUseCase_Run(t *testing.T) {
	var (
		errDB    = errors.New("db error")
		errToken = errors.New("token error")
		fakeUser = &entities.User{
			ID:        "uuid-fake-1",
			Username:  "john_doe",
			Password:  "$2a$10$hashed",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	)

	tests := []struct {
		name         string
		input        commands.LoginCommand
		mockBehavior func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService)
		wantErr      error
	}{
		{
			name:  "success — valid credentials issue token",
			input: commands.LoginCommand{Username: "john_doe", Password: "secret123"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().FindByUsername(gomock.Any(), "john_doe").Return(fakeUser, nil)
				hasher.EXPECT().ComparePassword("secret123", "$2a$10$hashed").Return(true)
				token.EXPECT().GenerateToken("uuid-fake-1", "john_doe", gomock.Any()).Return("jwt.access.token", nil)
			},
			wantErr: nil,
		},
		{
			name:  "user not found (nil) → ErrInvalidCredentials",
			input: commands.LoginCommand{Username: "ghost", Password: "password1"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().FindByUsername(gomock.Any(), "ghost").Return(nil, nil)
			},
			wantErr: domainerrors.ErrInvalidCredentials,
		},
		{
			name:  "FindByUsername db error → propagated",
			input: commands.LoginCommand{Username: "john_doe", Password: "password1"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().FindByUsername(gomock.Any(), "john_doe").Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name:  "wrong password → ErrInvalidCredentials",
			input: commands.LoginCommand{Username: "john_doe", Password: "wrongpass"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().FindByUsername(gomock.Any(), "john_doe").Return(fakeUser, nil)
				hasher.EXPECT().ComparePassword("wrongpass", "$2a$10$hashed").Return(false)
			},
			wantErr: domainerrors.ErrInvalidCredentials,
		},
		{
			name:  "GenerateToken fails → propagated",
			input: commands.LoginCommand{Username: "john_doe", Password: "secret123"},
			mockBehavior: func(repo *mocks.MockIUserRepository, hasher *mocks.MockIHasher, token *mocks.MockITokenService) {
				repo.EXPECT().FindByUsername(gomock.Any(), "john_doe").Return(fakeUser, nil)
				hasher.EXPECT().ComparePassword("secret123", "$2a$10$hashed").Return(true)
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

			uc := usecases.NewLoginUserUseCase(repoMock, hasherMock, tokenMock)
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
			if resp.Token == "" {
				t.Error("Run() expected non-empty token in response")
			}
			if resp.User.Username != fakeUser.Username {
				t.Errorf("Run() user.Username = %v, want %v", resp.User.Username, fakeUser.Username)
			}
		})
	}
}
