package amqp

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// Define a mock type that implements the MessageClient interface
type MockMessageClient struct {
	mock.Mock
}

func (m *MockMessageClient) publishMessage(types string, data interface{}) error {
	args := m.Called(types, data)
	return args.Error(0)
}

func TestPublishMessage(t *testing.T) {
	// Create an instance of the MockMessageClient
	mockClient := new(MockMessageClient)

	// Configure the expected behavior of the mock
	mockClient.On("publishMessage", "messageType", mock.Anything).Return(nil)

	// Call the method you want to test
	err := mockClient.publishMessage("messageType", "messageData")

	// Assert that the expected behavior was invoked
	assert.NoError(t, err)

	// Assert that the method was called with the expected arguments
	mockClient.AssertCalled(t, "publishMessage", "messageType", "messageData")
}
