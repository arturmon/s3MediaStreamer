package amqp

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/rabbitmq"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

// Interface defines methods for working with RabbitMQ.
type Interface interface {
	HandleMessage(ctx context.Context, messageBody map[string]interface{})
}

// Handler represents a repository for working with RabbitMQ.
type Handler struct {
	amqpService rabbitmq.Service
	Conn        *amqp091.Connection
	channels    map[string]*amqp091.Channel
	queues      map[string]*amqp091.Queue
	logger      *logs.Logger
	autoAck     bool
}

// NewAMQPHandler creates a new RabbitMQRepository instance with support for multiple queues.
func NewAMQPHandler(amqpService rabbitmq.Service, conn *amqp091.Connection, queueConfigs []model.QueueConfig, cfg *model.Config, logger *logs.Logger) (*Handler, error) {
	handler := &Handler{
		amqpService: amqpService,
		Conn:        conn,
		channels:    make(map[string]*amqp091.Channel),
		queues:      make(map[string]*amqp091.Queue),
		logger:      logger,
		autoAck:     cfg.Bus.SubscribeAutoAck,
	}

	// Create channels and queues for all specified queue configs
	for _, queueConfig := range queueConfigs {
		rabbitChannel, err := newRabbitMQChanel(conn)
		if err != nil {
			return nil, fmt.Errorf("error creating channel for queue '%s': %w", queueConfig.Name, err)
		}

		// Check the queue using checkQueue function
		errCheckQueue := checkQueue(&queueConfig, rabbitChannel, logger)
		if rabbitChannel.IsClosed() {
			rabbitChannel, err = newRabbitMQChanel(conn)
			if err != nil {
				return nil, fmt.Errorf("error creating new channel for queue '%s' after checking: %w", queueConfig.Name, errCheckQueue)
			}
		}
		if errCheckQueue != nil {
			// Handle specific AMQP error codes
			switch errCheckQueue.Code {
			case amqp091.NotFound:
				// Queue does not exist, so recreate the channel and continue with queue creation
				logger.Infof("Queue '%s' not found, error: %v, create new queue '%s'", queueConfig.Name, errCheckQueue, queueConfig.Name)
			case amqp091.AccessRefused:
				// Queue exists with different parameters, log the information and skip queue creation
				logger.Infof("Queue '%s' exists with different parameters: %v", queueConfig.Name, errCheckQueue)
				// Skip creating the queue since it exists with different parameters
				return nil, fmt.Errorf("queue '%s' already exists with different parameters: %w", queueConfig.Name, errCheckQueue)
			default:
				// For other AMQP errors, log the error and return it
				return nil, fmt.Errorf("failed to check queue '%s': %w", queueConfig.Name, errCheckQueue)
			}
		}

		// Add the created channel to the handler
		handler.channels[queueConfig.Name] = rabbitChannel

		// Declare or create the queue
		rabbitQueue, errQueue := newRabbitMQQueue(&queueConfig, rabbitChannel)
		if errQueue != nil {
			switch errQueue.Code {
			case amqp091.PreconditionFailed:
				logger.Fatalf("The queue '%s' already exists but with different parameters: %v", queueConfig.Name, errQueue)
			case amqp091.ResourceLocked:
				logger.Fatalf("The queue '%s' is blocked: %v", queueConfig.Name, errQueue)
			default:
				logger.Errorf("Error declaring or creating queue '%s': %v", queueConfig.Name, errQueue)
				return nil, errQueue
			}
		}

		handler.queues[queueConfig.Name] = rabbitQueue

		// Log information about the connected channel and queue
		logger.Infof("Connected to channel for queue '%s'", queueConfig.Name)
		logger.Infof("Queue '%s' details: Durable=%v, AutoDelete=%v, Exclusive=%v, NoWait=%v",
			queueConfig.Name,
			queueConfig.Durable,
			queueConfig.AutoDelete,
			queueConfig.Exclusive,
			queueConfig.NoWait,
		)
	}
	return handler, nil
}

func (c *Handler) Ping(_ context.Context) bool {
	return !c.Conn.IsClosed()
}

// StartAMQPConsumers starts consumers for all configured queues.
func (c *Handler) StartAMQPConsumers(ctx context.Context) {
	numWorkers := 5
	var wg sync.WaitGroup
	workerDone := make(chan struct{})

	// Start processing messages for each queue
	for queueName, queue := range c.queues {
		wg.Add(1) // Increment the worker counter

		go func(queueName string, queue *amqp091.Queue) {
			defer wg.Done() // Decrement the counter when the worker is finished

			if err := c.consumeMessagesFromQueue(ctx, queueName, queue, numWorkers, workerDone); err != nil {
				c.logger.Fatalf("Error consuming from queue %s: %v", queueName, err)
			}
		}(queueName, queue)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()         // Wait for all workers to complete
		close(workerDone) // Close the channel after completion
	}()
}

func (c *Handler) HandleMessage(ctx context.Context, queueName string, messageBody map[string]interface{}) {
	c.amqpService.HandleMessage(ctx, queueName, messageBody)
}
