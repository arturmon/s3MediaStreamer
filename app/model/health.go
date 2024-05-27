package model

type HealthResponse struct {
	Status string `json:"status"`
}

// LivenessResponse represents the response for the liveness probe.
type LivenessResponse struct {
	Status string `json:"status"`
}

// ReadinessResponse represents the response for the readiness probe.
type ReadinessResponse struct {
	Status string `json:"status"`
}
