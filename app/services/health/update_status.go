package health

// UpdateHealthStatus обновляет статус здоровья компонента.
func (wrapper *HealthCheckService) UpdateHealthStatus(metrics *HealthMetric, status bool, component string) {
	metrics.Mutex.Lock()
	defer metrics.Mutex.Unlock()

	for i, comp := range metrics.Components {
		if comp.Name == component {
			metrics.Components[i].Status = status
			return
		}
	}
	metrics.Components = append(metrics.Components, HeathMetrics{
		Status: status,
		Name:   component,
	})
}
