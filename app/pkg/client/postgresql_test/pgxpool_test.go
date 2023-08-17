package postgresql_test

import (
	"skeleton-golange-application/app/pkg/client/postgresql/mocks"
	"testing"
	"time"

	"github.com/google/uuid"

	"skeleton-golange-application/app/internal/config"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFindUserToEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	expectedUser := config.User{
		Email: "john@example.com",
	}
	mockCollectionQuery.EXPECT().FindUserToEmail("john@example.com").Return(expectedUser, nil)

	user, err := mockCollectionQuery.FindUserToEmail("john@example.com")

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up expectations on the mockCollectionQuery
	mockCollectionQuery.EXPECT().CreateUser(gomock.Any()).Return(nil)

	// Convert the string UUID to uuid.UUID type
	userID, err := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		t.Fatal(err)
	}

	// Test your PgClient method using the mockCollectionQuery
	user := config.User{
		ID:       userID,
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: []byte("password123"),
	}
	err = mockCollectionQuery.CreateUser(user)
	assert.NoError(t, err)
}

func TestDeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	emailToDelete := "john@example.com"
	mockCollectionQuery.EXPECT().DeleteUser(emailToDelete).Return(nil)

	err := mockCollectionQuery.DeleteUser(emailToDelete)

	// Verify the result
	assert.NoError(t, err)
}

func TestCreateIssue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	expectedAlbum := config.Album{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Artist:      "Artist name",
		Price:       111.111,
		Code:        "ALBUM123",
		Description: "A short description of the application",
		Completed:   false,
	}
	mockCollectionQuery.EXPECT().CreateIssue(gomock.AssignableToTypeOf(&expectedAlbum)).Return(nil)

	err := mockCollectionQuery.CreateIssue(&expectedAlbum)

	// Verify the result
	assert.NoError(t, err)
}

func TestCreateMany(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	expectedAlbums := []config.Album{
		{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       "Title 1",
			Artist:      "Artist 1",
			Price:       111.111,
			Code:        "ALBUM123",
			Description: "Description 1",
			Completed:   false,
		},
		{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       "Title 2",
			Artist:      "Artist 2",
			Price:       222.222,
			Code:        "ALBUM456",
			Description: "Description 2",
			Completed:   true,
		},
	}
	mockCollectionQuery.EXPECT().CreateMany(gomock.AssignableToTypeOf(expectedAlbums)).Return(nil)

	err := mockCollectionQuery.CreateMany(expectedAlbums)

	// Verify the result
	assert.NoError(t, err)
}

func TestGetAllIssues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	expectedAlbums := []config.Album{
		{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       "Title 1",
			Artist:      "Artist 1",
			Price:       111.111,
			Code:        "ALBUM123",
			Description: "Description 1",
			Completed:   false,
		},
		{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       "Title 2",
			Artist:      "Artist 2",
			Price:       222.222,
			Code:        "ALBUM456",
			Description: "Description 2",
			Completed:   true,
		},
	}
	mockCollectionQuery.EXPECT().GetAllIssues().Return(expectedAlbums, nil)

	albums, err := mockCollectionQuery.GetAllIssues()

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, expectedAlbums, albums)
}

func TestGetIssuesByCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	expectedAlbum := config.Album{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Title:       "Title",
		Artist:      "Artist",
		Price:       123.45,
		Code:        "ALBUM123",
		Description: "Description",
		Completed:   true,
	}
	mockCollectionQuery.EXPECT().GetIssuesByCode("ALBUM123").Return(expectedAlbum, nil)

	album, err := mockCollectionQuery.GetIssuesByCode("ALBUM123")

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, expectedAlbum, album)
}

func TestDeleteOne(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	mockCollectionQuery.EXPECT().DeleteOne("ALBUM123").Return(nil)

	err := mockCollectionQuery.DeleteOne("ALBUM123")

	// Verify the result
	assert.NoError(t, err)
}

func TestDeleteAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	mockCollectionQuery.EXPECT().DeleteAll().Return(nil)

	err := mockCollectionQuery.DeleteAll()

	// Verify the result
	assert.NoError(t, err)
}

func TestMarkCompleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	mockCollectionQuery.EXPECT().MarkCompleted("ALBUM123").Return(nil)

	err := mockCollectionQuery.MarkCompleted("ALBUM123")

	// Verify the result
	assert.NoError(t, err)
}

func TestUpdateIssue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	mockCollectionQuery.EXPECT().UpdateIssue(gomock.Any()).Return(nil)

	albumToUpdate := config.Album{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Title:       "Updated Title",
		Artist:      "Updated Artist",
		Price:       99.99,
		Code:        "ALBUM789",
		Description: "Updated Description",
		Completed:   false,
	}

	err := mockCollectionQuery.UpdateIssue(&albumToUpdate)

	// Verify the result
	assert.NoError(t, err)
}
