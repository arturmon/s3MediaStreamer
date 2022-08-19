package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthGET(c *gin.Context) {
	if health == true {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	} else {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"status": "DOWN",
		})
	}
}
