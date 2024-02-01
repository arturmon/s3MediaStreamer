package amqp

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/s3"
	"sync"

	"github.com/streadway/amqp"
)

// MessageClient represents an AMQP message client.
type MessageClient struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queue     amqp.Queue
	cfg       *config.Config
	s3Handler s3.HandlerS3
	logger    *logging.Logger
	storage   *model.DBConfig
}

// Worker represents a worker that processes AMQP messages.
type Worker struct {
	MessageClient *MessageClient
	workerDone    chan struct{}
}

// NewWorker creates a new Worker instance.
func NewWorker(messageClient *MessageClient, workerDone chan struct{}) *Worker {
	return &Worker{
		MessageClient: messageClient,
		workerDone:    workerDone,
	}
}

// NewAMQPClient creates a new instance of the MessageClient.
func NewAMQPClient(queueName string, cfg *config.Config, logger *logging.Logger) (*MessageClient, error) {
	var amqpURL string
	logger.Info("Starting AMQP Client...")
	if cfg.MessageQueue.BrokerPort != 0 {
		amqpURL = fmt.Sprintf("%s:%d", cfg.MessageQueue.Broker, cfg.MessageQueue.BrokerPort)
	} else {
		amqpURL = cfg.MessageQueue.Broker
	}

	amqpURLpriv := fmt.Sprintf("amqp://%s:%s@%s", cfg.MessageQueue.User, cfg.MessageQueue.Pass, amqpURL)

	conn, err := amqp.Dial(amqpURLpriv)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		queueName,       // queue name
		QueueDurable,    // durable
		QueueAutoDelete, // delete when unused
		QueueExclusive,  // exclusive
		QueueNoWait,     // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, err
	}

	s3Handler, err := s3.NewS3Handler(cfg, logger)
	if err != nil {
		return nil, err
	}

	storage, err := model.NewDBConfig(cfg, logger)
	if err != nil {
		return nil, err
	}

	logger.Infof("Connect AMQP Client: amqp://%s:***@%s", cfg.MessageQueue.User, amqpURL)

	return &MessageClient{
		conn:      conn,
		channel:   channel,
		queue:     queue,
		cfg:       cfg,
		s3Handler: s3Handler,
		logger:    logger,
		storage:   storage,
	}, nil
}

// Consume starts consuming messages from the queue.
func (c *MessageClient) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	messages, err := c.channel.Consume(
		c.queue.Name,      // queue
		"",                // consumer
		SubscribeAutoAck,  // auto-ack
		SubscribeExlusive, // exclusive
		SubscribeNoLocal,  // no-local
		SubscribeNoWait,   // no-wait
		nil,               // args
	)
	if err != nil {
		return nil, err
	}

	// Start processing the messages in a separate goroutine
	go c.consumeMessages(ctx, messages)

	c.logger.Infof("AMQP Consume: %s", c.queue.Name)
	return messages, nil
}

// StartProcessing starts processing messages using a worker pool.
func (w *Worker) StartProcessing(ctx context.Context, messages <-chan amqp.Delivery, wg *sync.WaitGroup, numWorkers int, workerDone chan struct{}) {
	// Start worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go w.worker(ctx, messages, wg)
	}

	// Defer wg.Done() after starting all workers
	defer func() {
		wg.Wait()
		close(workerDone)
	}()
}

// worker is the function that each worker executes.
func (w *Worker) worker(ctx context.Context, messages <-chan amqp.Delivery, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	// Worker function body remains unchanged
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-messages:
			if !ok {
				return
			}
			w.MessageClient.handleMessage(ctx, message)
		}
	}
}

// ConsumeMessagesWithPool starts consuming messages using a worker pool.
func ConsumeMessagesWithPool(ctx context.Context, logger logging.Logger, messageClient *MessageClient, numWorkers int, workerDone chan struct{}) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a WaitGroup to wait for the goroutine and workers to finish
	var wg sync.WaitGroup
	wg.Add(1)

	// Start consuming messages in a separate goroutine
	go func() {
		defer wg.Done()

		notificationChannel, err := messageClient.Consume(ctx)
		if err != nil {
			logger.Printf("Error getting objectInfo: %v\n", err)
			return
		}

		// Start processing messages using a worker pool
		worker := NewWorker(messageClient, workerDone)
		worker.StartProcessing(ctx, notificationChannel, &wg, numWorkers, workerDone)
	}()

	// Wait for the goroutine to finish
	wg.Wait()

	return nil
}
