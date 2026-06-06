package controllers

import (
	"errors"
	"net/http"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/ports"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

type QuizController struct {
	generateQuizUseCase      ports.IGenerateQuizPort
	saveQuizHistoryUseCase   ports.ISaveQuizHistoryPort
	getQuizHistoryUseCase    ports.IGetQuizHistoryPort
	deleteQuizHistoryUseCase ports.IDeleteQuizHistoryPort
}

func NewQuizController(
	generateQuizUseCase ports.IGenerateQuizPort,
	saveQuizHistoryUseCase ports.ISaveQuizHistoryPort,
	getQuizHistoryUseCase ports.IGetQuizHistoryPort,
	deleteQuizHistoryUseCase ports.IDeleteQuizHistoryPort,
	authMiddleware gin.HandlerFunc,
	r *gin.RouterGroup,
) *QuizController {
	ctrl := &QuizController{
		generateQuizUseCase:      generateQuizUseCase,
		saveQuizHistoryUseCase:   saveQuizHistoryUseCase,
		getQuizHistoryUseCase:    getQuizHistoryUseCase,
		deleteQuizHistoryUseCase: deleteQuizHistoryUseCase,
	}
	ctrl.setupRoutes(authMiddleware, r)
	return ctrl
}

func (ctrl *QuizController) Generate(c *gin.Context) {
	var cmd commands.GenerateQuizCommand
	userID, _ := c.Get("user_id")

	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response, err := ctrl.generateQuizUseCase.Run(c.Request.Context(), cmd, userID.(string))
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidTopic):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		case errors.Is(err, domainerrors.ErrQuizAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *QuizController) GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	response, err := ctrl.getQuizHistoryUseCase.Run(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *QuizController) SaveHistory(c *gin.Context) {
	var cmd commands.SaveQuizHistoryCommand
	userID, _ := c.Get("user_id")

	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := ctrl.saveQuizHistoryUseCase.Run(c.Request.Context(), cmd, userID.(string)); err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrQuizAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (ctrl *QuizController) DeleteHistory(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	if err := ctrl.deleteQuizHistoryUseCase.Run(c.Request.Context(), id, userID.(string)); err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrQuizNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		case errors.Is(err, domainerrors.ErrQuizForbidden):
			c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (ctrl *QuizController) setupRoutes(auth gin.HandlerFunc, r *gin.RouterGroup) {
	r.POST("/generate", auth, ctrl.Generate)
	r.GET("/history", auth, ctrl.GetHistory)
	r.POST("/history", auth, ctrl.SaveHistory)
	r.DELETE("/history/:id", auth, ctrl.DeleteHistory)
}
