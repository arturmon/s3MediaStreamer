package amqp

import (
	"context"
	"encoding/json"
	"s3MediaStreamer/app/internal/logs"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	SubscribeExlusive     = false
	SubscribeNoLocal      = false
	SubscribeNoWait       = false
	reconnectSleepSeconds = 5
)

func (c *Handler) ConsumeMessages(ctx context.Context, messages <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			// If the context is canceled, return and stop processing messages.
			return
		case message, ok := <-messages:
			// Check if the channel is closed (no more messages).
			if !ok {
				return
			}

			// Process the message depending on the action
			var messageBody map[string]interface{}
			err := json.Unmarshal(message.Body, &messageBody)
			if err != nil {
				c.logger.Errorf("Error unmarshalling message: %v", err)
				if !c.autoAck {
					// Reject message without resending if autoAck = false
					err = message.Reject(false)
					if err != nil {
						return
					}
				}
				continue
			}

			// Process the message
			go func(msg amqp091.Delivery) {
				c.HandleMessage(ctx, messageBody)
				// If autoAck = false, process the confirmation manually
				if !c.autoAck {
					err = msg.Ack(false)
					if err != nil {
						c.logger.Errorf("Error acknowledging message: %v", err)
					}
				}
			}(message)
		}
	}
}

// ConsumeMessagesWithPool starts consuming messages using a worker pool.
func (c *Handler) ConsumeMessagesWithPool(ctx context.Context, logger *logs.Logger, messageClient *Handler, numWorkers int, workerDone chan struct{}) error {
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
			worker.StartProcessing(ctx, notificationChannel, &wg, numWorkers, workerDone)

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

// Consume starts consuming messages from the queue.
func (c *Handler) Consume(ctx context.Context) (<-chan amqp091.Delivery, error) {
	messages, err := c.channel.Consume(
		c.queue.Name,      // queue
		"",                // consumer
		c.autoAck,         // auto-ack
		SubscribeExlusive, // exclusive
		SubscribeNoLocal,  // no-local
		SubscribeNoWait,   // no-wait
		nil,               // args
	)
	if err != nil {
		return nil, err
	}

	// Start processing the messages in a separate goroutine
	go c.ConsumeMessages(ctx, messages)

	return messages, nil
}
