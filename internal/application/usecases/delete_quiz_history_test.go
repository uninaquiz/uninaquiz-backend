package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/usecases"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestDeleteQuizHistoryUseCase_Run(t *testing.T) {
	var (
		errDB    = errors.New("db error")
		ownerID  = "user-owner-1"
		otherID  = "user-other-1"
		quizID   = "quiz-1"
		fakeQuiz = &entities.Quiz{ID: quizID, UserID: ownerID}
	)

	tests := []struct {
		name         string
		id           string
		userID       string
		mockBehavior func(repo *mocks.MockIQuizRepository)
		wantErr      error
	}{
		{
			name:   "success — quiz deleted",
			id:     quizID,
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(fakeQuiz, nil)
				repo.EXPECT().Delete(gomock.Any(), quizID).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "quiz not found (nil) → ErrQuizNotFound",
			id:     "nonexistent",
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "nonexistent").Return(nil, nil)
			},
			wantErr: domainerrors.ErrQuizNotFound,
		},
		{
			name:   "FindByID db error → propagated",
			id:     quizID,
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name:   "different userID → ErrQuizForbidden",
			id:     quizID,
			userID: otherID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(fakeQuiz, nil)
			},
			wantErr: domainerrors.ErrQuizForbidden,
		},
		{
			name:   "Delete db error → propagated",
			id:     quizID,
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(fakeQuiz, nil)
				repo.EXPECT().Delete(gomock.Any(), quizID).Return(errDB)
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

			uc := usecases.NewDeleteQuizHistoryUseCase(quizRepoMock)
			err := uc.Run(context.Background(), tt.id, tt.userID)

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
