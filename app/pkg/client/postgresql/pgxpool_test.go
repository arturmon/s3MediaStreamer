package postgresql_test

import (
	"skeleton-golange-application/app/pkg/client/postgresql/mocks"
	"skeleton-golange-application/model"
	"testing"
	"time"

	"github.com/bojanz/currency"

	"github.com/google/uuid"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFindUserToEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	expectedUser := model.User{
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
	user := model.User{
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
	expectedAlbum := model.Album{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Artist:      "Artist name",
		Code:        "ALBUM123",
		Description: "A short description of the application",
		Sender:      "amqp",
	}
	price, err := currency.NewAmount("111.11", "USD")
	assert.NoError(t, err)
	expectedAlbum.Price = price

	mockCollectionQuery.EXPECT().CreateIssue(gomock.AssignableToTypeOf(&expectedAlbum)).Return(nil)

	err = mockCollectionQuery.CreateIssue(&expectedAlbum)

	// Verify the result
	assert.NoError(t, err)
}

func TestCreateMany(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	expectedAlbums := []model.Album{
		{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       "Title 1",
			Artist:      "Artist 1",
			Code:        "ALBUM123",
			Description: "Description 1",
			Sender:      "amqp",
		},
		{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       "Title 2",
			Artist:      "Artist 2",
			Code:        "ALBUM456",
			Description: "Description 2",
			Sender:      "amqp",
		},
	}
	price, err := currency.NewAmount("111.11", "USD")
	price2, err2 := currency.NewAmount("222.11", "EUR")
	assert.NoError(t, err)
	assert.NoError(t, err2)
	expectedAlbums[0].Price = price
	expectedAlbums[1].Price = price2

	mockCollectionQuery.EXPECT().CreateMany(gomock.AssignableToTypeOf(expectedAlbums)).Return(nil)

	err = mockCollectionQuery.CreateMany(expectedAlbums)

	// Verify the result
	assert.NoError(t, err)
}

func TestGetAllIssues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	expectedAlbums := []model.Album{
		{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       "Title 1",
			Artist:      "Artist 1",
			Code:        "ALBUM123",
			Description: "Description 1",
			Sender:      "amqp",
		},
		{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       "Title 2",
			Artist:      "Artist 2",
			Code:        "ALBUM456",
			Description: "Description 2",
			Sender:      "amqp",
		},
	}
	price, err := currency.NewAmount("111.11", "USD")
	price2, err2 := currency.NewAmount("222.11", "EUR")
	assert.NoError(t, err)
	assert.NoError(t, err2)
	expectedAlbums[0].Price = price
	expectedAlbums[1].Price = price2

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
	expectedAlbum := model.Album{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Title:       "Title",
		Artist:      "Artist",
		Code:        "ALBUM123",
		Description: "Description",
		Sender:      "amqp",
	}
	price, err := currency.NewAmount("111.11", "USD")
	assert.NoError(t, err)
	expectedAlbum.Price = price

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

func TestUpdateIssue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollectionQuery := mocks.NewMockPostgresCollectionQuery(ctrl)

	// Set up an expected call on the mockCollectionQuery
	mockCollectionQuery.EXPECT().UpdateIssue(gomock.Any()).Return(nil)

	albumToUpdate := model.Album{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Title:       "Updated Title",
		Artist:      "Updated Artist",
		Code:        "ALBUM789",
		Description: "Updated Description",
		Sender:      "amqp",
	}
	price, err := currency.NewAmount("111.11", "USD")
	assert.NoError(t, err)
	albumToUpdate.Price = price

	err = mockCollectionQuery.UpdateIssue(&albumToUpdate)

	// Verify the result
	assert.NoError(t, err)
}
