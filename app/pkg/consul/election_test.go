package consul_test

import (
	"github.com/stretchr/testify/assert"
	"skeleton-golange-application/app/pkg/consul"

	"testing"
	"time"
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
