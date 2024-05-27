package health

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/services/db"
	"s3MediaStreamer/app/services/s3"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

const (
	checkHealthDBTimeoutSeconds     = 1
	checkHealthRabbitTimeoutSeconds = 2
	checkHealthS3TimeoutSeconds     = 3
)

type HealthRepository interface {
}

// HealthCheckWrapper предоставляет обертку для периодической проверки состояния здоровья различных сервисов.
type HealthCheckService struct {
	HealthMetrics  *HealthMetric
	DBRepository   db.DBRepository
	rabbitmq       *amqp.Connection
	s3Repository   s3.S3Repository
	s3DBRepository s3.S3DBRepository
	logger         *logs.Logger
}

// NewHealthCheckWrapper создает новую обертку для проверки здоровья.
func NewHealthCheckWrapper(metrics *HealthMetric, dbOps db.DBRepository, amqpClient *amqp.Connection, s3Handler s3.S3Repository, logger *logs.Logger) *HealthCheckService {
	return &HealthCheckService{
		HealthMetrics: metrics,
		DBRepository:  dbOps,
		rabbitmq:      amqpClient,
		s3Repository:  s3Handler,
		logger:        logger,
	}
}

// HeathMetrics представляет компонент здоровья приложения.
type HeathMetrics struct {
	Status bool   `json:"status"`
	Name   string `json:"name"`
}

// HealthMetric представляет метрику здоровья приложения.
type HealthMetric struct {
	Components []HeathMetrics
	Mutex      sync.Mutex
}

func NewHealthMetrics() *HealthMetric {
	return &HealthMetric{
		Components: []HeathMetrics{},
	}
}

// StartHealthChecks запускает периодические проверки состояния различных сервисов.
func (wrapper *HealthCheckService) StartHealthChecks() {
	go wrapper.periodicPing(wrapper.pingDatabase, time.Second*checkHealthDBTimeoutSeconds)
	go wrapper.periodicPing(wrapper.pingRabbitMQ, time.Second*checkHealthRabbitTimeoutSeconds)
	go wrapper.periodicPing(wrapper.pingS3, time.Second*checkHealthS3TimeoutSeconds)
}

// periodicPing is a generic function for periodic health_handler checks.
func (wrapper *HealthCheckService) periodicPing(pingFunc func(context.Context), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		pingFunc(context.Background())
	}
}

func (wrapper *HealthCheckService) CheckMonitoring(ctx context.Context, resultChan chan<- bool) {
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
