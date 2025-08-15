package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)


type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}


type HealthHandler struct {
	startTime time.Time
	version   string
}


func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   "1.0.0",
	}
}


func (h *HealthHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(h.startTime)
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Version:   h.version,
		Services: map[string]string{
			"gateway": "healthy",
		},
	}

	log.Debug().Msg("Health check requested")
	c.JSON(http.StatusOK, response)
}


