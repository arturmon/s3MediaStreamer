package consul_test

import (
	"skeleton-golange-application/app/pkg/consul"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitializeLeaderElection(t *testing.T) {
	// You might need to set up a Consul client for testing.
	config := &consul.LeaderElectionConfig{
		CheckTimeout: time.Second * 5,
		// Add other necessary fields.
	}

	election := consul.InitializeLeaderElection(config)

	assert.NotNil(t, election)
}

func TestGetLocalIP(t *testing.T) {
	ip := consul.GetLocalIP()

	assert.NotEmpty(t, ip)
}
