package consul_service_test

import (
	consulService "s3MediaStreamer/app/services/consul_service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	service := consulService.Service{} // Create an instance of the Service type
	ip := service.GetLocalIP()

	assert.NotEmpty(t, ip)
}
