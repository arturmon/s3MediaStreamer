package health

import (
	"context"
)

/*
func (wrapper *HealthCheckWrapper) pingDatabase(ctx context.Context, logger *logs.Logger) bool {
	err := wrapper.dbRepository.Ping(ctx)
	if err != nil {
		logger.Errorf("Error pinging database: %v", err)
		return false
	}
	return true
}

*/

func (wrapper *HealthCheckService) pingDatabase(ctx context.Context) {
	err := wrapper.DBRepository.Ping(ctx)
	if err != nil {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, false, "db")
		wrapper.logger.Errorf("Error pinging database: %v", err)
	} else {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, true, "db")
	}
}
