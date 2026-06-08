//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/adapters"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/services"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/usecases"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/database/repositories"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/http/controllers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const testJWTSecret = "integration-test-jwt-secret-key"

type stubAIService struct{}

func (s *stubAIService) GenerateQuiz(_ context.Context, topic, _ string) ([]dto.QuizQuestion, error) {
	return []dto.QuizQuestion{
		{
			Text:         "Stub question for " + topic,
			Options:      []string{"Option A", "Option B", "Option C", "Option D"},
			CorrectIndex: 0,
			Explanation:  "Stub explanation",
		},
	}, nil
}

var _ services.IAIService = (*stubAIService)(nil)

type testEnv struct {
	db       *gorm.DB
	router   *gin.Engine
	tokenSvc services.ITokenService
	token    string
	userID   string
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(
		ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("uninaquiz_test"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	runMigrations(t, connStr)

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	hasher := adapters.NewHashPasswordAdapter()
	tokenSvc := adapters.NewJwtTokenAdapter(testJWTSecret)
	aiSvc := &stubAIService{}

	userRepo := repositories.NewUserRepository(db)
	quizRepo := repositories.NewQuizRepository(db)

	createUserUC := usecases.NewCreateUserUseCase(userRepo, hasher, tokenSvc)
	loginUserUC := usecases.NewLoginUserUseCase(userRepo, hasher, tokenSvc)
	getAllUsersUC := usecases.NewGetAllUsersUseCase(userRepo)
	generateQuizUC := usecases.NewGenerateQuizUseCase(aiSvc, quizRepo)
	saveHistoryUC := usecases.NewSaveQuizHistoryUseCase(quizRepo)
	getHistoryUC := usecases.NewGetQuizHistoryUseCase(quizRepo)
	getQuizUC := usecases.NewGetQuizUseCase(quizRepo)
	deleteHistoryUC := usecases.NewDeleteQuizHistoryUseCase(quizRepo)

	engine := gin.New()
	authMiddleware := middleware.AuthMiddleware(tokenSvc)

	apiGroup := engine.Group("/api")
	controllers.NewUserController(getAllUsersUC, apiGroup)

	authGroup := apiGroup.Group("/auth")
	controllers.NewAuthController(createUserUC, loginUserUC, authGroup)

	quizGroup := apiGroup.Group("/quiz")
	controllers.NewQuizController(
		generateQuizUC,
		saveHistoryUC,
		getHistoryUC,
		getQuizUC,
		deleteHistoryUC,
		authMiddleware,
		quizGroup,
	)

	return &testEnv{db: db, router: engine, tokenSvc: tokenSvc}
}

func runMigrations(t *testing.T, connStr string) {
	t.Helper()

	_, file, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(file), "..", "..", "db", "migrations")

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open sql db for migrations: %v", err)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
}

func (e *testEnv) registerAndLogin(t *testing.T, username, password string) {
	t.Helper()

	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	e.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: status=%d body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}

	e.token = resp["token"].(string)
	user := resp["user"].(map[string]interface{})
	e.userID = user["id"].(string)
}
