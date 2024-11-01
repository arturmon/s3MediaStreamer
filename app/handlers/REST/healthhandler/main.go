package healthhandler

import (
	"net/http"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/health"

	"github.com/gin-gonic/gin"
)

type MonitoringServiceInterface interface {
}

type Handler struct {
	monitoringService health.Service
}

func NewMonitoringHandler(monitoringService health.Service) *Handler {
	return &Handler{monitoringService}
}

func (h *Handler) LivenessGET(c *gin.Context, wrapper *health.Service) {
	// Use pingDatabase for liveness probe
	err := wrapper.DBRepository.Ping(c)
	if err != nil {
		c.JSON(http.StatusOK, model.LivenessResponse{
			Status: "DOWN",
		})
		return
	}

	c.JSON(http.StatusOK, model.LivenessResponse{
		Status: "UP",
	})
}

func (h *Handler) ReadinessGET(c *gin.Context, wrapper *health.Service) {
	// wrapper.HealthMetrics.Mutex.Lock()
	// defer wrapper.HealthMetrics.Mutex.Unlock()
	status := "UP"

	// Create a slice to store HealthCheckComponent for each component
	components := wrapper.HealthMetrics.Components
	for _, component := range wrapper.HealthMetrics.Components {
		if !component.Status {
			status = "DOWN"
		}
	}

	if status == "UP" {
		c.JSON(http.StatusOK, components)
	} else {
		c.JSON(http.StatusBadGateway, components)
	}
}
