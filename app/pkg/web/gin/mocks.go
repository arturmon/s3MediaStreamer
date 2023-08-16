package gin

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
)

type MockDBOperations struct {
	mockAlbums []config.Album
}

func (m *MockDBOperations) Connect() error {
	// Implement mock behavior for Connect
	return nil
}

func (m *MockDBOperations) Ping(ctx context.Context) error {
	// Implement mock behavior for Ping
	return nil
}

func (m *MockDBOperations) GetIssuesByCode(code string) (*config.Album, error) {
	// Implement mock behavior for GetIssuesByCode

	// Let's assume that we have a mockAlbums variable which holds the list of mock albums.
	// We will iterate through the list and find the album with the given code.
	for i := range m.mockAlbums {
		if m.mockAlbums[i].Code == code {
			// If we find the album with the given code, we return a pointer to it as the result.
			return &m.mockAlbums[i], nil
		}
	}

	// If the album with the given code was not found, we can return an error.
	// You can customize the error message based on your requirements.
	return nil, fmt.Errorf("album with code %s not found", code)
}

func (m *MockDBOperations) DeleteOne(code string) error {
	// Implement mock behavior for DeleteOne

	// Let's assume that we have a mockAlbums variable which holds the list of mock albums.
	// We will iterate through the list and find the album with the given code.
	for i := 0; i < len(m.mockAlbums); i++ {
		if m.mockAlbums[i].Code == code {
			// If we find the album with the given code, we will "delete" it from the list.
			// In this example, "deleting" means removing the album from the list.
			m.mockAlbums = append(m.mockAlbums[:i], m.mockAlbums[i+1:]...)
			return nil // Return nil to indicate successful "deletion."
		}
	}

	// If the album with the given code was not found, we can return an error.
	// You can customize the error message based on your requirements.
	return fmt.Errorf("album with code %s not found", code)
}

func (m *MockDBOperations) DeleteAll() error {
	// Implement mock behavior for DeleteAll

	// Clear the list of mock albums to simulate deleting all records.
	m.mockAlbums = []config.Album{}

	return nil // Return nil to indicate successful "deletion."
}
