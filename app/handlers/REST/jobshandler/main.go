package jobshandler

import (
	"net/http"

	"github.com/bamzi/jobrunner"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type JobServiceInterface interface {
}

type Handler struct {
}

func NewJobHandler() *Handler {
	return &Handler{}
}

// JobStatus godoc
// @Summary All Job status
// @Description Check if the application server is running jobs
// @Tags health-controller
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "OK"
// @Failure 404 {object} map[string]string "Not Found"
// @Router /job/status [get]
func (h *Handler) JobStatus(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "JobStatus")
	defer span.End()
	c.JSON(http.StatusOK, jobrunner.StatusJson())
}
