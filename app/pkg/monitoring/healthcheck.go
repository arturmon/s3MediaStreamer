package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status"`
}

// HealthGET godoc
// @Summary Get health status of the application
// @Description Checks and returns the current health status of the application
// @Tags health-controller
// @Accept  */*
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 500 {object} HealthResponse
// @Router /health [get]
func HealthGET(c *gin.Context) {
	if config.AppHealth {
		c.JSON(http.StatusOK, HealthResponse{
			Status: "UP",
		})
	} else {
		c.JSON(http.StatusInternalServerError, HealthResponse{
			Status: "DOWN",
		})
	}
}

func PingStorage(ctx context.Context, dbOps model.DBOperations) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := dbOps.Ping(ctx)
		if err != nil {
			config.AppHealth = false
			fmt.Println("Error pinging database:", err)
		} else {
			config.AppHealth = true
		}
	}
}
