package gin

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
)

type MockDBOperations struct {
	mockStorage []config.User
	mockAlbums  []config.Album
}

func (m *MockDBOperations) Connect() error {
	// Implement mock behavior for Connect
	return nil
}

func (m *MockDBOperations) Close(ctx context.Context) error {
	// Implement mock behavior for Close
	return nil
}

func (m *MockDBOperations) Ping(ctx context.Context) error {
	// Implement mock behavior for Ping
	return nil
}

func (m *MockDBOperations) FindUserToEmail(email string) (config.User, error) {
	// Implement mock behavior for FindUserToEmail
	// You can return a mock user based on the provided email
	return config.User{}, nil
}

func (m *MockDBOperations) CreateUser(user config.User) error {
	// Implement mock behavior for CreateUser
	return nil
}

func (m *MockDBOperations) DeleteUser(email string) error {
	// Implement mock behavior for DeleteUser
	return nil
}

func (m *MockDBOperations) CreateIssue(task *config.Album) error {
	// Implement mock behavior for CreateIssue
	return nil
}

func (m *MockDBOperations) CreateMany(list []*config.Album) error {
	// Implement mock behavior for CreateMany
	return nil
}

func (m *MockDBOperations) GetAllIssues() ([]config.Album, error) {
	// Implement mock behavior for GetAllIssues
	// You can return a list of mock albums
	return m.mockAlbums, nil
}

func (m *MockDBOperations) GetIssuesByCode(code string) (*config.Album, error) {
	// Implement mock behavior for GetIssuesByCode

	// Let's assume that we have a mockAlbums variable which holds the list of mock albums.
	// We will iterate through the list and find the album with the given code.
	for _, album := range m.mockAlbums {
		if album.Code == code {
			// If we find the album with the given code, we return a pointer to it as the result.
			return &album, nil
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
	for i, album := range m.mockAlbums {
		if album.Code == code {
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

func (m *MockDBOperations) MarkCompleted(code string) error {
	// Implement mock behavior for MarkCompleted
	return nil
}
