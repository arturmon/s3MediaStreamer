package health

import (
	"context"
)

func (wrapper *Service) pingRabbitMQ(_ context.Context) {
	ping := wrapper.rabbitmq.IsClosed()
	if ping {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, false, "rabbit")
		wrapper.logger.Errorf("Error pinging Rabbitmq")
	} else {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, true, "rabbit")
	}
}
