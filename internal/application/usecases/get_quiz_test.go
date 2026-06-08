package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/usecases"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetQuizUseCase_Run(t *testing.T) {
	errDB := errors.New("db error")
	ownerID := "user-owner-1"
	otherID := "user-other-1"
	quizID := "quiz-1"
	now := time.Now()

	fakeQuiz := &entities.Quiz{
		ID:         quizID,
		UserID:     ownerID,
		Topic:      "Mathematics",
		Difficulty: "easy",
		Score:      4,
		Total:      5,
		CreatedAt:  now,
		Questions: []entities.QuizQuestion{
			{ID: "qq1", QuizID: quizID, Position: 0, Text: "What is 2+2?", Options: []string{"3", "4", "5"}, CorrectIndex: 1, Explanation: "Basic addition"},
		},
	}

	tests := []struct {
		name         string
		quizID       string
		userID       string
		mockBehavior func(repo *mocks.MockIQuizRepository)
		wantErr      error
		checkResp    func(t *testing.T, resp interface{})
	}{
		{
			name:   "success — owner retrieves quiz with questions",
			quizID: quizID,
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(fakeQuiz, nil)
			},
			wantErr: nil,
		},
		{
			name:   "quiz not found (nil) → ErrQuizNotFound",
			quizID: "nonexistent",
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "nonexistent").Return(nil, nil)
			},
			wantErr: domainerrors.ErrQuizNotFound,
		},
		{
			name:   "FindByID db error → propagated",
			quizID: quizID,
			userID: ownerID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name:   "different userID → ErrQuizForbidden",
			quizID: quizID,
			userID: otherID,
			mockBehavior: func(repo *mocks.MockIQuizRepository) {
				repo.EXPECT().FindByID(gomock.Any(), quizID).Return(fakeQuiz, nil)
			},
			wantErr: domainerrors.ErrQuizForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizRepoMock := mocks.NewMockIQuizRepository(ctrl)
			tt.mockBehavior(quizRepoMock)

			uc := usecases.NewGetQuizUseCase(quizRepoMock)
			resp, err := uc.Run(context.Background(), tt.quizID, tt.userID)

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
			if resp.ID != fakeQuiz.ID {
				t.Errorf("ID = %v, want %v", resp.ID, fakeQuiz.ID)
			}
			if resp.Topic != fakeQuiz.Topic {
				t.Errorf("Topic = %v, want %v", resp.Topic, fakeQuiz.Topic)
			}
		})
	}
}
