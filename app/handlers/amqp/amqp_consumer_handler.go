package amqp

import (
	"context"
	"encoding/json"
	"s3MediaStreamer/app/internal/logs"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

const (
	SubscribeAutoAck      = true
	SubscribeExlusive     = false
	SubscribeNoLocal      = false
	SubscribeNoWait       = false
	reconnectSleepSeconds = 5
)

func (c *Handler) ConsumeMessages(ctx context.Context, messages <-chan amqp.Delivery) {
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

			// Handle the message based on its action
			var messageBody map[string]interface{}
			err := json.Unmarshal(message.Body, &messageBody)
			if err != nil {
				// error unmarshal
				continue
			}
			go c.HandleMessage(ctx, messageBody)
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
			case <-messageClient.Conn.NotifyClose(make(chan *amqp.Error)):
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
func (c *Handler) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
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
	go c.ConsumeMessages(ctx, messages)

	return messages, nil
}
