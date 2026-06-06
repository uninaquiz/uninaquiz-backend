package controllers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/ports"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	createUserUseCase ports.ICreateUserPort
	loginUserUseCase  ports.ILoginUserPort
}

func NewAuthController(createUserUseCase ports.ICreateUserPort, loginUserUseCase ports.ILoginUserPort, r *gin.RouterGroup) *AuthController {
	ctrl := &AuthController{
		createUserUseCase: createUserUseCase,
		loginUserUseCase:  loginUserUseCase,
	}
	ctrl.setupRoutes(r)
	return ctrl
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var cmd commands.CreateUserCommand

	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response, err := ctrl.createUserUseCase.Run(c.Request.Context(), cmd, 7*24*time.Hour)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
		default:
			log.Printf("register error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": response.Token,
		"user":  response.User,
	})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var cmd commands.LoginCommand

	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response, err := ctrl.loginUserUseCase.Run(c.Request.Context(), cmd, 7*24*time.Hour)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		default:
			log.Printf("login error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": response.Token,
		"user":  response.User,
	})
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	// JWT é stateless; logout é gerenciado pelo cliente descartando o token
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (ctrl *AuthController) setupRoutes(r *gin.RouterGroup) {
	r.POST("/register", ctrl.Register)
	r.POST("/login", ctrl.Login)
	r.POST("/logout", ctrl.Logout)
}
