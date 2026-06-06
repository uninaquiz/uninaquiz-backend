package factories

import (
	"fmt"
	"log"

	"github.com/EmanuelErnesto/uninaquiz-backend/cmd/config"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/adapters"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/ports"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/services"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/usecases"
	domainrepos "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/repositories"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/database/repositories"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/http/controllers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	Engine *gin.Engine
}

type Container struct {
	Config                   config.Config
	DB                       *gorm.DB
	Hasher                   services.IHasher
	TokenService             services.ITokenService
	AIService                services.IAIService
	UserRepository           domainrepos.IUserRepository
	QuizRepository           domainrepos.IQuizRepository
	CreateUserUseCase        ports.ICreateUserPort
	GetAllUsersUseCase       ports.IGetAllUsersPort
	LoginUserUseCase         ports.ILoginUserPort
	GenerateQuizUseCase      ports.IGenerateQuizPort
	SaveQuizHistoryUseCase   ports.ISaveQuizHistoryPort
	GetQuizHistoryUseCase    ports.IGetQuizHistoryPort
	DeleteQuizHistoryUseCase ports.IDeleteQuizHistoryPort
	Server                   *Server
}

func NewContainer() *Container {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	hasher := adapters.NewHashPasswordAdapter()
	tokenService := adapters.NewJwtTokenAdapter(cfg.JWTSecret)

	aiService, err := adapters.NewGeminiAIAdapter(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("failed to initialize AI service: %v", err)
	}

	userRepository := repositories.NewUserRepository(db)
	quizRepository := repositories.NewQuizRepository(db)

	createUserUseCase := usecases.NewCreateUserUseCase(userRepository, hasher, tokenService)
	getAllUsersUseCase := usecases.NewGetAllUsersUseCase(userRepository)
	loginUserUseCase := usecases.NewLoginUserUseCase(userRepository, hasher, tokenService)
	generateQuizUseCase := usecases.NewGenerateQuizUseCase(aiService, quizRepository)
	saveQuizHistoryUseCase := usecases.NewSaveQuizHistoryUseCase(quizRepository)
	getQuizHistoryUseCase := usecases.NewGetQuizHistoryUseCase(quizRepository)
	deleteQuizHistoryUseCase := usecases.NewDeleteQuizHistoryUseCase(quizRepository)

	engine := gin.Default()

	authMiddleware := middleware.AuthMiddleware(tokenService)

	apiGroup := engine.Group("/api")
	controllers.NewHealthController(apiGroup)
	controllers.NewUserController(getAllUsersUseCase, apiGroup)

	authGroup := engine.Group("/auth")
	controllers.NewAuthController(createUserUseCase, loginUserUseCase, authGroup)

	quizGroup := engine.Group("/quiz")
	controllers.NewQuizController(
		generateQuizUseCase,
		saveQuizHistoryUseCase,
		getQuizHistoryUseCase,
		deleteQuizHistoryUseCase,
		authMiddleware,
		quizGroup,
	)

	return &Container{
		Config:                   cfg,
		DB:                       db,
		Hasher:                   hasher,
		TokenService:             tokenService,
		AIService:                aiService,
		UserRepository:           userRepository,
		QuizRepository:           quizRepository,
		CreateUserUseCase:        createUserUseCase,
		GetAllUsersUseCase:       getAllUsersUseCase,
		LoginUserUseCase:         loginUserUseCase,
		GenerateQuizUseCase:      generateQuizUseCase,
		SaveQuizHistoryUseCase:   saveQuizHistoryUseCase,
		GetQuizHistoryUseCase:    getQuizHistoryUseCase,
		DeleteQuizHistoryUseCase: deleteQuizHistoryUseCase,
		Server:                   &Server{Engine: engine},
	}
}
