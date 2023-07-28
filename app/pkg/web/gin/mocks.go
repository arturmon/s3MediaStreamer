package gin

import (
	"context"
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

func (m *MockDBOperations) CreateIssue(task config.Album) error {
	// Implement mock behavior for CreateIssue
	return nil
}

func (m *MockDBOperations) CreateMany(list []config.Album) error {
	// Implement mock behavior for CreateMany
	return nil
}

func (m *MockDBOperations) GetAllIssues() ([]config.Album, error) {
	// Implement mock behavior for GetAllIssues
	// You can return a list of mock albums
	return m.mockAlbums, nil
}

func (m *MockDBOperations) GetIssuesByCode(code string) (config.Album, error) {
	// Implement mock behavior for GetIssuesByCode
	// You can return a mock album based on the provided code
	return config.Album{}, nil
}

func (m *MockDBOperations) DeleteOne(code string) error {
	// Implement mock behavior for DeleteOne
	return nil
}

func (m *MockDBOperations) DeleteAll() error {
	// Implement mock behavior for DeleteAll
	return nil
}

func (m *MockDBOperations) MarkCompleted(code string) error {
	// Implement mock behavior for MarkCompleted
	return nil
}
