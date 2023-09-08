package gin

import (
	"github.com/bamzi/jobrunner"
	"github.com/gin-gonic/gin"
	"net/http"
)

// JobStatus godoc
// @Summary All Job status
// @Description Check if the application server is running jobs
// @Tags health-controller
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "OK"
// @Failure 404 {object} map[string]string "Not Found"
// @Router /job/status [get]
func JobStatus(c *gin.Context) {
	c.JSON(http.StatusOK, jobrunner.StatusJson())
}
