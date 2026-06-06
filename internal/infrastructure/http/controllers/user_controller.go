package controllers

import (
	"log"
	"net/http"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/ports"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/queries"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	getAllUsersUseCase ports.IGetAllUsersPort
}

func NewUserController(getAllUsersUseCase ports.IGetAllUsersPort, r *gin.RouterGroup) *UserController {
	usc := &UserController{
		getAllUsersUseCase: getAllUsersUseCase,
	}
	usc.setupRoutes(r)
	return usc
}

func (uc *UserController) GetUsers(c *gin.Context) {
	var query queries.GetAllUsersQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response, err := uc.getAllUsersUseCase.Run(c.Request.Context(), query)
	if err != nil {
		log.Printf("get users error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (uc *UserController) setupRoutes(r *gin.RouterGroup) {
	r.GET("/users", uc.GetUsers)
}
