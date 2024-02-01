package amqp

import (
	"context"
	"sync"

	"github.com/streadway/amqp"
)

type Workers interface {
	StartProcessing(ctx context.Context, messages <-chan amqp.Delivery, wg *sync.WaitGroup, numWorkers int, workerDone chan struct{})
	worker(ctx context.Context, messages <-chan amqp.Delivery, wg *sync.WaitGroup)
}
