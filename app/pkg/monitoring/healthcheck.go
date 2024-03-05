package monitoring

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/pkg/amqp"
	"s3MediaStreamer/app/pkg/client/model"
	"s3MediaStreamer/app/pkg/logging"
	"s3MediaStreamer/app/pkg/s3"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// LivenessResponse represents the response for the liveness probe.
type LivenessResponse struct {
	Status string `json:"status"`
}

// ReadinessResponse represents the response for the readiness probe.
type ReadinessResponse struct {
	Status string `json:"status"`
}

type HeathMetrics struct {
	Status bool   `json:"status"`
	Name   string `json:"name"`
}

type HealthMetric struct {
	Components []HeathMetrics `json:"components"`
	mutex      sync.Mutex
}

func NewHealthMetrics() *HealthMetric {
	return &HealthMetric{
		Components: []HeathMetrics{},
	}
}

type HealthResponse struct {
	Status string `json:"status"`
}

// LivenessGET godoc
// @Summary Get liveness status of the application
// @Description Checks and returns the liveness status of the application
// @Tags health-controller
// @Accept  */*
// @Produce json
// @BasePath	/
// @Success 200 {object} LivenessResponse
// @Failure 502 {object} model.ErrorResponse "Internal Server Error"
// @Router /health/liveness [get]
func LivenessGET(c *gin.Context, wrapper *HealthCheckWrapper) {
	// Use pingDatabase for liveness probe
	err := wrapper.DBOps.Ping(context.Background())
	if err != nil {
		c.JSON(http.StatusOK, LivenessResponse{
			Status: "DOWN",
		})
		return
	}

	c.JSON(http.StatusOK, LivenessResponse{
		Status: "UP",
	})
}

// ReadinessGET godoc
// @Summary Get readiness status of the application
// @Description Checks and returns the readiness status of the application
// @Tags health-controller
// @Accept  */*
// @Produce json
// @BasePath	/
// @Success 200 {object} ReadinessResponse
// @Failure 502 {object} model.ErrorResponse "Internal Server Error"
// @Router /health/readiness [get]
func ReadinessGET(c *gin.Context, wrapper *HealthCheckWrapper) {
	wrapper.HealthMetrics.mutex.Lock()
	defer wrapper.HealthMetrics.mutex.Unlock()
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

func UpdateHealthStatus(metrics *HealthMetric, status bool, component string) {
	metrics.mutex.Lock()
	defer metrics.mutex.Unlock()

	for i, comp := range metrics.Components {
		if comp.Name == component {
			metrics.Components[i].Status = status
			return
		}
	}
	log.Debugf("Updating health status. Component: %s, Status: %t", component, status)
	metrics.Components = append(metrics.Components, HeathMetrics{
		Status: status,
		Name:   component,
	})
}

// HealthCheckWrapper provides a wrapper for periodic health checks.
type HealthCheckWrapper struct {
	HealthMetrics *HealthMetric
	DBOps         model.DBOperations
	AMQPClient    *amqp.MessageClient
	S3Handler     s3.HandlerS3
	Logger        *logging.Logger
}

// NewHealthCheckWrapper initializes a new HealthCheckWrapper.
func NewHealthCheckWrapper(
	metrics *HealthMetric,
	dbOps model.DBOperations,
	amqpClient *amqp.MessageClient,
	s3Handler s3.HandlerS3,
	logger *logging.Logger) *HealthCheckWrapper {
	return &HealthCheckWrapper{
		HealthMetrics: metrics,
		DBOps:         dbOps,
		AMQPClient:    amqpClient,
		S3Handler:     s3Handler,
		Logger:        logger,
	}
}

// StartHealthChecks starts the periodic health checks for different services.
func (wrapper *HealthCheckWrapper) StartHealthChecks() {
	go wrapper.periodicPing(wrapper.pingDatabase, time.Second*checkHealthDBTimeoutSeconds)
	go wrapper.periodicPing(wrapper.pingRabbitMQ, time.Second*checkHealthRabbitTimeoutSeconds)
	go wrapper.periodicPing(wrapper.pingS3, time.Second*checkHealthS3TimeoutSeconds)
	// Add more services with different timers as needed
}

// periodicPing is a generic function for periodic health checks.
func (wrapper *HealthCheckWrapper) periodicPing(pingFunc func(context.Context), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		pingFunc(context.Background())
	}
}

// pingDatabase checks the health of the database.
func (wrapper *HealthCheckWrapper) pingDatabase(ctx context.Context) {
	err := wrapper.DBOps.Ping(ctx)
	if err != nil {
		UpdateHealthStatus(wrapper.HealthMetrics, false, "db")
		wrapper.Logger.Errorf("Error pinging database: %v", err)
	} else {
		UpdateHealthStatus(wrapper.HealthMetrics, true, "db")
	}
}

// pingRabbitMQ checks the health of the RabbitMQ connection.
func (wrapper *HealthCheckWrapper) pingRabbitMQ(_ context.Context) {
	if wrapper.AMQPClient.Conn.IsClosed() {
		UpdateHealthStatus(wrapper.HealthMetrics, false, "rabbit")
		wrapper.Logger.Errorf("Error pinging Rabbitmq")
	} else {
		UpdateHealthStatus(wrapper.HealthMetrics, true, "rabbit")
	}
}

// pingS3 checks the health of the S3 connection.
func (wrapper *HealthCheckWrapper) pingS3(ctx context.Context) {
	err := wrapper.S3Handler.Ping(ctx)
	if err != nil {
		UpdateHealthStatus(wrapper.HealthMetrics, false, "s3")
		wrapper.Logger.Errorf("Error pinging S3: %v", err)
	} else {
		UpdateHealthStatus(wrapper.HealthMetrics, true, "s3")
	}
}

func (wrapper *HealthCheckWrapper) CheckMonitoring(ctx context.Context, resultChan chan<- bool) {
	go func() {
		defer func() {
			// Optionally, you can close the channel here if needed
			// close(resultChan)
		}()

		ticker := time.NewTicker(time.Second) // Adjust the interval as needed
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return // exit goroutine when context is canceled
			case <-ticker.C:
				// Read the components without locks
				components := wrapper.HealthMetrics.Components
				allComponentsHealthy := true

				for _, component := range components {
					if !component.Status {
						allComponentsHealthy = false
						break
					}
				}

				// Send the result through the provided channel
				resultChan <- allComponentsHealthy
			}
		}
	}()
}
