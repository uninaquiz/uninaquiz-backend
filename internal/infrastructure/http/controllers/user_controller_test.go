package controllers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/queries"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/http/controllers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/mocks"
	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func setupUserRouter(getAllUC *mocks.MockIGetAllUsersPort) *gin.Engine {
	r := gin.New()
	group := r.Group("/api/v1")
	controllers.NewUserController(getAllUC, group)
	return r
}

func TestUserController_GetUsers(t *testing.T) {
	now := time.Now()
	fakeResp := &dto.GetAllUsersResponse{
		Data: []dto.UserResponse{
			{ID: "u1", Username: "alice", CreatedAt: now, UpdatedAt: now},
			{ID: "u2", Username: "bob", CreatedAt: now, UpdatedAt: now},
		},
		Total:      2,
		Page:       1,
		Limit:      10,
		TotalPages: 1,
	}

	tests := []struct {
		name         string
		query        string
		mockBehavior func(getAllUC *mocks.MockIGetAllUsersPort)
		wantStatus   int
		wantTotal    int64
	}{
		{
			name:  "200 — default pagination (page=1, limit=10 from struct defaults)",
			query: "",
			mockBehavior: func(getAllUC *mocks.MockIGetAllUsersPort) {
				// gin ShouldBindQuery fills Page=1 and Limit=10 from binding:"default=..." tags
				getAllUC.EXPECT().Run(gomock.Any(), queries.GetAllUsersQuery{Page: 1, Limit: 10}).Return(fakeResp, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  2,
		},
		{
			name:  "500 — internal error",
			query: "",
			mockBehavior: func(getAllUC *mocks.MockIGetAllUsersPort) {
				getAllUC.EXPECT().Run(gomock.Any(), gomock.Any()).Return(nil, errors.New("db crash"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			getAllMock := mocks.NewMockIGetAllUsersPort(ctrl)
			tt.mockBehavior(getAllMock)

			router := setupUserRouter(getAllMock)

			url := "/api/v1/users"
			if tt.query != "" {
				url += "?" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d — body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
