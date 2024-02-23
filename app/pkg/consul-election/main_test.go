package consul_election_test

import (
	"github.com/stretchr/testify/assert"
	consul_election "skeleton-golange-application/app/pkg/consul-election"
	"testing"
	"time"
)

func TestInitializeLeaderElection(t *testing.T) {
	// You might need to set up a Consul client for testing.
	config := &consul_election.LeaderElectionConfig{
		CheckTimeout: time.Second * 5,
		// Add other necessary fields.
	}

	election := consul_election.InitializeLeaderElection(config)

	assert.NotNil(t, election)
}
