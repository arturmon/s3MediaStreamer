package consul

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEventLeader(t *testing.T) {
	notify := Notify{T: "Test"}
	eventLeader := true

	notify.EventLeader(eventLeader)
}

func TestInitializeLeaderElection(t *testing.T) {
	// You might need to set up a Consul client for testing.
	config := &LeaderElectionConfig{
		CheckTimeout: time.Second * 5,
		// Add other necessary fields.
	}

	election := InitializeLeaderElection(config)

	assert.NotNil(t, election)
}

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()

	assert.NotEmpty(t, ip)
}
