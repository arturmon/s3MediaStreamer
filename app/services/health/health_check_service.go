package health

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/services/db"
	"s3MediaStreamer/app/services/s3"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	checkHealthDBTimeoutSeconds     = 1
	checkHealthRabbitTimeoutSeconds = 2
	checkHealthS3TimeoutSeconds     = 3
)

type Repository interface {
}

// HealthCheckWrapper предоставляет обертку для периодической проверки состояния здоровья различных сервисов.
type Service struct {
	HealthMetrics *Metric
	DBRepository  db.Repository
	rabbitmq      *amqp091.Connection
	s3Repository  s3.Repository
	logger        *logs.Logger
}

// NewHealthCheckWrapper создает новую обертку для проверки здоровья.
func NewHealthCheckWrapper(metrics *Metric, dbOps db.Repository, amqpClient *amqp091.Connection, s3Handler s3.Repository, logger *logs.Logger) *Service {
	return &Service{
		HealthMetrics: metrics,
		DBRepository:  dbOps,
		rabbitmq:      amqpClient,
		s3Repository:  s3Handler,
		logger:        logger,
	}
}

// Metrics представляет компонент здоровья приложения.
type Metrics struct {
	Status bool   `json:"status"`
	Name   string `json:"name"`
}

// Metric представляет метрику здоровья приложения.
type Metric struct {
	Components []Metrics
	Mutex      sync.Mutex
}

func NewHealthMetrics() *Metric {
	return &Metric{
		Components: []Metrics{},
	}
}

// StartHealthChecks запускает периодические проверки состояния различных сервисов.
func (wrapper *Service) StartHealthChecks() {
	go wrapper.periodicPing(wrapper.pingDatabase, time.Second*checkHealthDBTimeoutSeconds)
	go wrapper.periodicPing(wrapper.pingRabbitMQ, time.Second*checkHealthRabbitTimeoutSeconds)
	go wrapper.periodicPing(wrapper.pingS3, time.Second*checkHealthS3TimeoutSeconds)
}

// periodicPing is a generic function for periodic health_handler checks.
func (wrapper *Service) periodicPing(pingFunc func(context.Context), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		pingFunc(context.Background())
	}
}

func (wrapper *Service) CheckMonitoring(ctx context.Context, resultChan chan<- bool) {
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
