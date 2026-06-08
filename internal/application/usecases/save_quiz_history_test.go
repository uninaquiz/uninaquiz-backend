package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/usecases"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestSaveQuizHistoryUseCase_Run(t *testing.T) {
	var (
		errDB    = errors.New("db error")
		ownerID  = "user-owner-1"
		otherID  = "user-other-1"
		quizID   = "quiz-1"
		fakeQuiz = &entities.Quiz{ID: quizID, UserID: ownerID, Score: 0, Total: 5}
	)

	tests := []struct {
		name         string
		input        commands.SaveQuizHistoryCommand
		userID       string
		mockBehavior func(repo *mocks.MockIQuizRepository)
		wantErr      error
	}{
		{
			name:   "success — score persisted",
			input:  commands.SaveQuizHistoryCommand{ID: quizID, Score: 4},
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(fakeQuiz, nil)
				repo.EXPECT().UpdateScore(gomock.Any(), quizID, 4).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "quiz not found (nil) → ErrQuizNotFound",
			input:  commands.SaveQuizHistoryCommand{ID: "nonexistent", Score: 3},
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "nonexistent").Return(nil, nil)
			},
			wantErr: domainerrors.ErrQuizNotFound,
		},
		{
			name:   "FindByID db error → propagated",
			input:  commands.SaveQuizHistoryCommand{ID: quizID, Score: 2},
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name:   "different userID → ErrQuizForbidden",
			input:  commands.SaveQuizHistoryCommand{ID: quizID, Score: 5},
			userID: otherID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(fakeQuiz, nil)
			},
			wantErr: domainerrors.ErrQuizForbidden,
		},
		{
			name:   "UpdateScore db error → propagated",
			input:  commands.SaveQuizHistoryCommand{ID: quizID, Score: 3},
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(fakeQuiz, nil)
				repo.EXPECT().UpdateScore(gomock.Any(), quizID, 3).Return(errDB)
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

			uc := usecases.NewSaveQuizHistoryUseCase(quizRepoMock)
			err := uc.Run(context.Background(), tt.input, tt.userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Run() unexpected error: %v", err)
			}
		})
	}
}
