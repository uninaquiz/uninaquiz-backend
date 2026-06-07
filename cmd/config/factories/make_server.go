package factories

import (
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
	GetQuizUseCase           ports.IGetQuizPort
	DeleteQuizHistoryUseCase ports.IDeleteQuizHistoryPort
	Server                   *Server
}

func NewContainer() *Container {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
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
	getQuizUseCase := usecases.NewGetQuizUseCase(quizRepository)
	deleteQuizHistoryUseCase := usecases.NewDeleteQuizHistoryUseCase(quizRepository)

	engine := gin.Default()

	engine.Use(middleware.CORSMiddleware())

	authMiddleware := middleware.AuthMiddleware(tokenService)

	// Health check at root level
	rootGroup := engine.Group("")
	controllers.NewHealthController(rootGroup)

	apiGroup := engine.Group("/api")
	controllers.NewUserController(getAllUsersUseCase, apiGroup)

	authGroup := apiGroup.Group("/auth")
	controllers.NewAuthController(createUserUseCase, loginUserUseCase, authGroup)

	quizGroup := apiGroup.Group("/quiz")
	controllers.NewQuizController(
		generateQuizUseCase,
		saveQuizHistoryUseCase,
		getQuizHistoryUseCase,
		getQuizUseCase,
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
		GetQuizUseCase:           getQuizUseCase,
		DeleteQuizHistoryUseCase: deleteQuizHistoryUseCase,
		Server:                   &Server{Engine: engine},
	}
}
