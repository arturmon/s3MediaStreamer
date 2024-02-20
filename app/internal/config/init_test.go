package config_test

import (
	"skeleton-golange-application/app/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	// Call GetConfig and check if it returns a non-nil Config instance
	config := config.GetConfig()
	assert.NotNil(t, config)
}
