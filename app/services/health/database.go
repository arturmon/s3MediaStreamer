package health

import (
	"context"
)

func (wrapper *Service) pingDatabase(ctx context.Context) {
	err := wrapper.DBRepository.Ping(ctx)
	if err != nil {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, false, "db")
		wrapper.logger.Errorf("Error pinging database: %v", err)
	} else {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, true, "db")
	}
}
