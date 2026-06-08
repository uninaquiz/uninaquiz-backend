package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/usecases"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetQuizHistoryUseCase_Run(t *testing.T) {
	errDB := errors.New("db error")
	now := time.Now()
	userID := "user-1"

	fakeQuizzes := []entities.Quiz{
		{ID: "q1", UserID: userID, Topic: "Math", Difficulty: "easy", Score: 4, Total: 5, CreatedAt: now},
		{ID: "q2", UserID: userID, Topic: "Physics", Difficulty: "hard", Score: 2, Total: 5, CreatedAt: now},
	}

	tests := []struct {
		name         string
		userID       string
		mockBehavior func(repo *mocks.MockIQuizRepository)
		wantErr      error
		wantCount    int
	}{
		{
			name:   "success — returns quiz history",
			userID: userID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindAllByUserID(gomock.Any(), userID).Return(fakeQuizzes, nil)
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:   "no quizzes found → returns empty slice",
			userID: userID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindAllByUserID(gomock.Any(), userID).Return([]entities.Quiz{}, nil)
			},
			wantErr:   nil,
			wantCount: 0,
		},
		{
			name:   "FindAllByUserID db error → propagated",
			userID: userID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindAllByUserID(gomock.Any(), userID).Return(nil, errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizRepoMock := mocks.NewMockIQuizRepository(ctrl)
			tt.mockBehavior(quizRepoMock)

			uc := usecases.NewGetQuizHistoryUseCase(quizRepoMock)
			resp, err := uc.Run(context.Background(), tt.userID)

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
			if len(resp) != tt.wantCount {
				t.Errorf("len(resp) = %d, want %d", len(resp), tt.wantCount)
			}
		})
	}
}
