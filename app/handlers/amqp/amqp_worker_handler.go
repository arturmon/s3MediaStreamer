package amqp

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/streadway/amqp"
)

// Worker represents a worker that processes AMQP messages.
type Worker struct {
	MessageClient *AmqpHandler
	workerDone    chan struct{}
}

// NewWorker creates a new Worker instance.
func NewWorker(messageClient *AmqpHandler, workerDone chan struct{}) *Worker {
	return &Worker{
		MessageClient: messageClient,
		workerDone:    workerDone,
	}
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
			var messageBody map[string]interface{}
			err := json.Unmarshal(message.Body, &messageBody)
			if err != nil {
				// error unmarshal
				continue
			}

			w.MessageClient.HandleMessage(ctx, messageBody)
		}
	}
}
