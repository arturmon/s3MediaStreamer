package consul_service_test

import (
	"github.com/stretchr/testify/assert"
	consul_service "skeleton-golange-application/app/pkg/consul-service"
	"testing"
)

func TestGetLocalIP(t *testing.T) {
	service := consul_service.Service{} // Create an instance of the Service type
	ip := service.GetLocalIP()

	assert.NotEmpty(t, ip)
}
