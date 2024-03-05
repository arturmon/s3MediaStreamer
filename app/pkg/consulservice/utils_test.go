package consulservice_test

import (
	consul_service "s3MediaStreamer/app/pkg/consulservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	service := consul_service.Service{} // Create an instance of the Service type
	ip := service.GetLocalIP()

	assert.NotEmpty(t, ip)
}
