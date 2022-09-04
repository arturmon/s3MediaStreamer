package monitoring

import (
	"net/http"
	"skeleton-golange-application/app/internal/config"

	"github.com/gin-gonic/gin"
)

func HealthGET(c *gin.Context) {
	if config.AppHealth == true {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	} else {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"status": "DOWN",
		})
	}
}
