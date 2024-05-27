package health

import (
	"context"
)

func (wrapper *HealthCheckService) pingS3(ctx context.Context) {
	err := wrapper.s3Repository.Ping(ctx)
	if err != nil {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, false, "s3")
		wrapper.logger.Errorf("Error pinging S3: %v", err)
	} else {
		wrapper.UpdateHealthStatus(wrapper.HealthMetrics, true, "s3")
	}
}
