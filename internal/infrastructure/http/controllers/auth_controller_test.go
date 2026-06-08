package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/http/controllers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/mocks"
	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuthRouter(
	createUC *mocks.MockICreateUserPort,
	loginUC *mocks.MockILoginUserPort,
) *gin.Engine {
	r := gin.New()
	group := r.Group("/auth")
	controllers.NewAuthController(createUC, loginUC, group)
	return r
}

func TestAuthController_Register(t *testing.T) {
	var (
		errInternal = errors.New("unexpected db failure")
		fakeResp    = &dto.CreateUserResponse{
			User:  dto.UserResponse{ID: "u1", Username: "alice", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			Token: "jwt.token",
		}
	)

	tests := []struct {
		name         string
		body         interface{}
		mockBehavior func(createUC *mocks.MockICreateUserPort)
		wantStatus   int
	}{
		{
			name: "201 — valid registration",
			body: commands.CreateUserCommand{Username: "alice", Password: "pass1234"},
			mockBehavior: func(createUC *mocks.MockICreateUserPort) {
				createUC.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeResp, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:         "400 — missing fields",
			body:         map[string]string{"username": "al"},
			mockBehavior: func(createUC *mocks.MockICreateUserPort) {},
			wantStatus:   http.StatusBadRequest,
		},
		{
			name: "409 — username already exists",
			body: commands.CreateUserCommand{Username: "existing", Password: "pass1234"},
			mockBehavior: func(createUC *mocks.MockICreateUserPort) {
				createUC.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, domainerrors.ErrUserAlreadyExists)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "500 — unexpected internal error",
			body: commands.CreateUserCommand{Username: "someuser", Password: "pass1234"},
			mockBehavior: func(createUC *mocks.MockICreateUserPort) {
				createUC.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errInternal)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			createMock := mocks.NewMockICreateUserPort(ctrl)
			loginMock := mocks.NewMockILoginUserPort(ctrl)
			tt.mockBehavior(createMock)

			router := setupAuthRouter(createMock, loginMock)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d — body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestAuthController_Login(t *testing.T) {
	fakeResp := &dto.LoginResponse{
		Token: "jwt.token",
		User:  dto.UserResponse{ID: "u1", Username: "alice"},
	}

	tests := []struct {
		name         string
		body         interface{}
		mockBehavior func(loginUC *mocks.MockILoginUserPort)
		wantStatus   int
	}{
		{
			name: "200 — valid login",
			body: commands.LoginCommand{Username: "alice", Password: "pass1234"},
			mockBehavior: func(loginUC *mocks.MockILoginUserPort) {
				loginUC.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeResp, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:         "400 — missing password",
			body:         map[string]string{"username": "alice"},
			mockBehavior: func(loginUC *mocks.MockILoginUserPort) {},
			wantStatus:   http.StatusBadRequest,
		},
		{
			name: "401 — invalid credentials",
			body: commands.LoginCommand{Username: "alice", Password: "wrongpass"},
			mockBehavior: func(loginUC *mocks.MockILoginUserPort) {
				loginUC.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, domainerrors.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "500 — internal server error",
			body: commands.LoginCommand{Username: "alice", Password: "pass1234"},
			mockBehavior: func(loginUC *mocks.MockILoginUserPort) {
				loginUC.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db crash"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			createMock := mocks.NewMockICreateUserPort(ctrl)
			loginMock := mocks.NewMockILoginUserPort(ctrl)
			tt.mockBehavior(loginMock)

			router := setupAuthRouter(createMock, loginMock)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d — body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestAuthController_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMock := mocks.NewMockICreateUserPort(ctrl)
	loginMock := mocks.NewMockILoginUserPort(ctrl)
	router := setupAuthRouter(createMock, loginMock)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Logout status = %d, want 200", w.Code)
	}
}
