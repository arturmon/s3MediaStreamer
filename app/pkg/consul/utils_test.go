package consul_test

import (
	"github.com/stretchr/testify/assert"
	"skeleton-golange-application/app/pkg/consul"
	"testing"
)

func TestGetLocalIP(t *testing.T) {
	ip := consul.GetLocalIP()

	assert.NotEmpty(t, ip)
}
