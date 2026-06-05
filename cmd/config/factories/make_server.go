package factories

import (
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/infrastructure/http/controllers"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
}

func MakeServer() *Server {
	engine := gin.Default()

	apiGroup := engine.Group("/api")
	controllers.NewHealthController(apiGroup)

	return &Server{
		Engine: engine,
	}
}
