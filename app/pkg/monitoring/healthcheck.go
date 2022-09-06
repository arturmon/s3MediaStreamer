package monitoring

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"skeleton-golange-application/app/internal/config"
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

func PingStorage(ctx context.Context, client *mongo.Client, cfg *config.Config) {
	ticker := time.NewTicker(1 * time.Second)
	var err error
	for range ticker.C {
		if cfg.Storage.Type == "mongo" {
			err = client.Ping(ctx, readpref.Primary())
		}
		if err != nil {
			config.AppHealth = false
		} else {
			config.AppHealth = true
		}
	}
}
