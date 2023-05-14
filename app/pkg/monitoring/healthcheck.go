package monitoring

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"time"
)

func HealthGET(c *gin.Context) {
	if config.AppHealth {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "DOWN",
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
