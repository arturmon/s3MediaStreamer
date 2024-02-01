package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfig(t *testing.T) {
	// Call GetConfig and check if it returns a non-nil Config instance
	config := GetConfig()
	assert.NotNil(t, config)
}
