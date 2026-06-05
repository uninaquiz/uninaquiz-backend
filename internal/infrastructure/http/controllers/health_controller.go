package controllers

import "github.com/gin-gonic/gin"

type HealthController struct {
}

func NewHealthController(r *gin.RouterGroup) *HealthController {
	h := &HealthController{}
	h.setupRoutes(r)
	return h
}

func (h *HealthController) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (h *HealthController) setupRoutes(r *gin.RouterGroup) {
	r.GET("/health", h.HealthCheck)
}
