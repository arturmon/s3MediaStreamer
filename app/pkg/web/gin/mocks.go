package gin

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/model"
)

type MockDBOperations struct {
	mockTracks []model.Track
}

func (m *MockDBOperations) Connect() error {
	// Implement mock behavior for Connect
	return nil
}

func (m *MockDBOperations) Ping(_ context.Context) error {
	// Implement mock behavior for Ping
	return nil
}

func (m *MockDBOperations) GetIssuesByCode(code string) (*model.Track, error) {
	// Implement mock behavior for GetIssuesByCode

	// Let's assume that we have a mockTracks variable which holds the list of mock tracks.
	// We will iterate through the list and find the track with the given code.
	for i := range m.mockTracks {
		if m.mockTracks[i].Code == code {
			// If we find the track with the given code, we return a pointer to it as the result.
			return &m.mockTracks[i], nil
		}
	}

	// If the track with the given code was not found, we can return an error.
	// You can customize the error message based on your requirements.
	return nil, fmt.Errorf("track with code %s not found", code)
}

func (m *MockDBOperations) DeleteOne(code string) error {
	// Implement mock behavior for DeleteOne

	// Let's assume that we have a mockTracks variable which holds the list of mock tracks.
	// We will iterate through the list and find the track with the given code.
	for i := 0; i < len(m.mockTracks); i++ {
		if m.mockTracks[i].Code == code {
			// If we find the track with the given code, we will "delete" it from the list.
			// In this example, "deleting" means removing the track from the list.
			m.mockTracks = append(m.mockTracks[:i], m.mockTracks[i+1:]...)
			return nil // Return nil to indicate successful "deletion."
		}
	}

	// If the track with the given code was not found, we can return an error.
	// You can customize the error message based on your requirements.
	return fmt.Errorf("track with code %s not found", code)
}

func (m *MockDBOperations) DeleteAll() error {
	// Implement mock behavior for DeleteAll

	// Clear the list of mock tracks to simulate deleting all records.
	m.mockTracks = []model.Track{}

	return nil // Return nil to indicate successful "deletion."
}
