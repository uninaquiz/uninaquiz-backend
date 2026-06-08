package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/queries"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/usecases"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetAllUsersUseCase_Run(t *testing.T) {
	errDB := errors.New("db error")
	now := time.Now()

	fakeUsers := []entities.User{
		{ID: "u1", Username: "alice", CreatedAt: now, UpdatedAt: now},
		{ID: "u2", Username: "bob", CreatedAt: now, UpdatedAt: now},
	}

	tests := []struct {
		name         string
		input        queries.GetAllUsersQuery
		mockBehavior func(repo *mocks.MockIUserRepository)
		wantErr      error
		wantTotal    int64
		wantPages    int
		wantCount    int
	}{
		{
			name:  "success — returns paginated users",
			input: queries.GetAllUsersQuery{Page: 1, Limit: 10},
			mockBehavior: func(repo *mocks.MockIUserRepository) {
				repo.EXPECT().GetAll(gomock.Any(), 1, 10).Return(fakeUsers, int64(2), nil)
			},
			wantErr:   nil,
			wantTotal: 2,
			wantPages: 1,
			wantCount: 2,
		},
		{
			name:  "success — second page with partial results",
			input: queries.GetAllUsersQuery{Page: 2, Limit: 1},
			mockBehavior: func(repo *mocks.MockIUserRepository) {
				repo.EXPECT().GetAll(gomock.Any(), 2, 1).Return(fakeUsers[1:], int64(2), nil)
			},
			wantErr:   nil,
			wantTotal: 2,
			wantPages: 2,
			wantCount: 1,
		},
		{
			name:  "empty result set",
			input: queries.GetAllUsersQuery{Page: 1, Limit: 10},
			mockBehavior: func(repo *mocks.MockIUserRepository) {
				repo.EXPECT().GetAll(gomock.Any(), 1, 10).Return([]entities.User{}, int64(0), nil)
			},
			wantErr:   nil,
			wantTotal: 0,
			wantPages: 0,
			wantCount: 0,
		},
		{
			name:  "GetAll db error → propagated",
			input: queries.GetAllUsersQuery{Page: 1, Limit: 10},
			mockBehavior: func(repo *mocks.MockIUserRepository) {
				repo.EXPECT().GetAll(gomock.Any(), 1, 10).Return(nil, int64(0), errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := mocks.NewMockIUserRepository(ctrl)
			tt.mockBehavior(repoMock)

			uc := usecases.NewGetAllUsersUseCase(repoMock)
			resp, err := uc.Run(context.Background(), tt.input)

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
			if resp.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", resp.Total, tt.wantTotal)
			}
			if resp.TotalPages != tt.wantPages {
				t.Errorf("TotalPages = %d, want %d", resp.TotalPages, tt.wantPages)
			}
			if len(resp.Data) != tt.wantCount {
				t.Errorf("len(Data) = %d, want %d", len(resp.Data), tt.wantCount)
			}
			if resp.Page != tt.input.Page {
				t.Errorf("Page = %d, want %d", resp.Page, tt.input.Page)
			}
			if resp.Limit != tt.input.Limit {
				t.Errorf("Limit = %d, want %d", resp.Limit, tt.input.Limit)
			}
		})
	}
}
