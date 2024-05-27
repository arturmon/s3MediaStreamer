package health

import (
	"context"
)

/*
func (wrapper *HealthCheckWrapper) pingRabbitMQ(ctx context.Context, logger *logs.Logger) bool {
	ping := wrapper.rabbitmqRepository.Ping(ctx)
	if !ping {
		logger.Errorf("Error pinging Rabbitmq")
		return false
	}
	return true
}

*/

func (wrapper *HealthCheckService) pingRabbitMQ(_ context.Context) {
	ping := wrapper.rabbitmq.IsClosed()
	if ping {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, false, "rabbit")
		wrapper.logger.Errorf("Error pinging Rabbitmq")
	} else {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, true, "rabbit")
	}
}
