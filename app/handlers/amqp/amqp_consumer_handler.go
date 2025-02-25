package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

const (
	SubscribeExlusive     = false
	SubscribeNoLocal      = false
	SubscribeNoWait       = false
	reconnectSleepSeconds = 5
)

func (c *Handler) ConsumeMessages(ctx context.Context, queueName string, messages <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			// If the context is canceled, return and stop processing messages.
			return
		case message, ok := <-messages:
			if !ok {
				return
			}

			// Process each message
			c.processMessage(ctx, queueName, message)
		}
	}
}

// processMessage handles processing for a single message.
func (c *Handler) processMessage(ctx context.Context, queueName string, message amqp091.Delivery) {
	var messageBody map[string]interface{}
	err := json.Unmarshal(message.Body, &messageBody)
	if err != nil {
		c.logger.Errorf("Error unmarshalling message: %v", err)
		c.rejectMessageIfNeeded(message)
		return
	}

	// Process and acknowledge the message
	go c.handleAndAcknowledge(ctx, queueName, message, messageBody)
}

// handleAndAcknowledge processes the message and acknowledges it.
func (c *Handler) handleAndAcknowledge(ctx context.Context, queueName string, message amqp091.Delivery, messageBody map[string]interface{}) {
	// Handle the message
	c.HandleMessage(ctx, queueName, messageBody)

	// Acknowledge the message if autoAck is false
	if !c.autoAck {
		if err := message.Ack(false); err != nil {
			c.logger.Errorf("Error acknowledging message: %v", err)
		}
	}
}

// rejectMessageIfNeeded rejects the message if autoAck is false.
func (c *Handler) rejectMessageIfNeeded(message amqp091.Delivery) {
	if !c.autoAck {
		if err := message.Reject(false); err != nil {
			c.logger.Errorf("Error rejecting message: %v", err)
		}
	}
}

/*
// consumeMessagesWithPool starts consuming messages using a worker pool.
func (c *Handler) consumeMessagesWithPool(ctx context.Context, queueName string, logger *logs.Logger, messageClient *Handler, numWorkers int, workerDone chan struct{}) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a WaitGroup to wait for the goroutine and workers to finish
	var wg sync.WaitGroup
	wg.Add(1)

	// Start consuming messages in a separate goroutine
	go func() {
		defer wg.Done()

		for {
			if messageClient == nil {
				return
			}

			notificationChannel, err := messageClient.Consume(ctx)
			if err != nil {
				logger.Errorf("Error getting objectInfo: %v\n", err)
				time.Sleep(reconnectSleepSeconds * time.Second)
				continue
			}

			// Start processing messages using a worker pool
			worker := NewWorker(messageClient, workerDone)
			worker.StartProcessing(ctx, queueName, notificationChannel, &wg, numWorkers)

			// Block here until the connection is closed, then attempt to reconnect
			select {
			case <-ctx.Done():
				return
			case <-messageClient.Conn.NotifyClose(make(chan *amqp091.Error)):
				logger.Warn("RabbitMQ connection closed, attempting to reconnect...")
				time.Sleep(reconnectSleepSeconds * time.Second)
			}
		}
	}()

	// Wait for the goroutine to finish
	wg.Wait()

	return nil
}

*/

// consumeMessagesFromQueue consumes messages from a specific queue.
func (c *Handler) consumeMessagesFromQueue(
	ctx context.Context,
	queueName string,
	queue *amqp091.Queue,
	numWorkers int,
	workerDone chan struct{},
) error {
	messages, err := c.channels[queueName].Consume(
		queue.Name,        // queue
		"",                // consumer
		c.autoAck,         // auto-ack
		SubscribeExlusive, // exclusive
		SubscribeNoLocal,  // no-local
		SubscribeNoWait,   // no-wait
		nil,               // args
	)
	if err != nil {
		return err
	}

	worker := NewWorker(c, workerDone)
	var wg sync.WaitGroup
	worker.StartProcessing(ctx, queueName, messages, &wg, numWorkers)
	wg.Wait()
	return nil
}

// Consume sets up the consumer for a given queue.
func (c *Handler) Consume(ctx context.Context) (<-chan amqp091.Delivery, error) {
	// Ensure the channel exists
	if c.channels == nil {
		return nil, fmt.Errorf("no channels available")
	}

	// Assume we're consuming from a queue (for simplicity, use the first queue here).
	// Adjust as needed to select a specific queue from the channels map.
	for queueName, channel := range c.channels {
		// Consume messages from the queue
		messages, err := channel.Consume(
			queueName,         // queue
			"",                // consumer
			c.autoAck,         // auto-ack
			SubscribeExlusive, // exclusive
			SubscribeNoLocal,  // no-local
			SubscribeNoWait,   // no-wait
			nil,               // args
		)
		if err != nil {
			return nil, fmt.Errorf("error consuming messages from queue %s: %w", queueName, err)
		}
		go c.ConsumeMessages(ctx, queueName, messages)
		// return messages, nil
	}

	// In case no channels are found
	return nil, fmt.Errorf("no valid channels available")
}
