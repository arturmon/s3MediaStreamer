package monitoring

import (
	"context"
	"net/http"
	"skeleton-golange-application/app/pkg/client/model"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type HealthMetrics struct {
	HealthStatus bool
	// Add other metric fields here as needed
}

func NewHealthMetrics() *HealthMetrics {
	return &HealthMetrics{
		HealthStatus: false,
		// Initialize other metric fields here
	}
}

type HealthResponse struct {
	Status string `json:"status"`
}

// HealthGET godoc
// @Summary Get health status of the application
// @Description Checks and returns the current health status of the application
// @Tags health-controller
// @Accept  */*
// @Produce json
// @BasePath	/
// @Success 200 {object} HealthResponse
// @Failure 500 {object} HealthResponse
// @Router /health [get]
func HealthGET(c *gin.Context, metrics *HealthMetrics) {
	if metrics.HealthStatus {
		c.JSON(http.StatusOK, HealthResponse{
			Status: "UP",
		})
	} else {
		c.JSON(http.StatusInternalServerError, HealthResponse{
			Status: "DOWN",
		})
	}
}

func PingStorage(ctx context.Context, dbOps model.DBOperations, metrics *HealthMetrics) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := dbOps.Ping(ctx)
		if err != nil {
			metrics.HealthStatus = false
			log.Infof("Error pinging database: %v", err)
		} else {
			metrics.HealthStatus = true
		}
	}
}
