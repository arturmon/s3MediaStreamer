package consulelection_test

import (
	consul_election "s3MediaStreamer/app/pkg/consulelection"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
