package health_handler

import (
	"net/http"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/health"

	"github.com/gin-gonic/gin"
)

type MonitoringServiceInterface interface {
}

type MonitoringHandler struct {
	monitoringService health.HealthCheckService
}

func NewMonitoringHandler(monitoringService health.HealthCheckService) *MonitoringHandler {
	return &MonitoringHandler{monitoringService}
}

// LivenessGET godoc
// @Summary Get liveness status of the application
// @Description Checks and returns the liveness status of the application
// @Tags health_handler-controller
// @Accept  */*
// @Produce json
// @BasePath	/
// @Success 200 {object} model.LivenessResponse
// @Failure 502 {object} model.ErrorResponse "Internal Server Error"
// @Router /health/liveness [get]
func (h *MonitoringHandler) LivenessGET(c *gin.Context, wrapper *health.HealthCheckService) {
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

// ReadinessGET godoc
// @Summary Get readiness status of the application
// @Description Checks and returns the readiness status of the application
// @Tags health_handler-controller
// @Accept  */*
// @Produce json
// @BasePath	/
// @Success 200 {object} model.ReadinessResponse
// @Failure 502 {object} model.ErrorResponse "Internal Server Error"
// @Router /health/readiness [get]
func (h *MonitoringHandler) ReadinessGET(c *gin.Context, wrapper *health.HealthCheckService) {
	//wrapper.HealthMetrics.Mutex.Lock()
	//defer wrapper.HealthMetrics.Mutex.Unlock()
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
