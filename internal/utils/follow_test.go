package utils

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
)

func TestIsFollowing(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Configure GORM with mock
	dialector := postgres.New(postgres.Config{
		Conn:                 mockDB,
		DriverName:           "postgres",
		PreferSimpleProtocol: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// Assign mock DB to database.DB for testing
	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	tests := []struct {
		name           string
		followerID     string
		followingID    string
		mockRows       *sqlmock.Rows
		expectedResult bool
		expectedError  bool
	}{
		{
			name:        "User is following",
			followerID:  "user1",
			followingID: "user2",
			mockRows: sqlmock.NewRows([]string{"id", "created_at", "follower_id", "creator_id"}).
				AddRow("follow1", time.Now(), "user1", "user2"),
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "User is not following",
			followerID:     "user1",
			followingID:    "user2",
			mockRows:       sqlmock.NewRows([]string{"id", "created_at", "follower_id", "creator_id"}),
			expectedResult: false,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `SELECT`
			mock.ExpectQuery(query).WillReturnRows(tt.mockRows)

			result, err := IsFollowing(tt.followerID, tt.followingID)

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
