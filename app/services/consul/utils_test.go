package consul_test

import (
	"s3MediaStreamer/app/services/consul"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	service := consul.Service{} // Create an instance of the Service type
	ip := service.GetLocalIP()

	assert.NotEmpty(t, ip)
}
