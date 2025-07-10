package user

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
)

func TestIsAdmin(t *testing.T) {
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
		userID         string
		mockRows       *sqlmock.Rows
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "User is admin",
			userID:         "admin-user-id",
			mockRows:       sqlmock.NewRows([]string{"is_admin"}).AddRow(true),
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "User is not admin",
			userID:         "regular-user-id",
			mockRows:       sqlmock.NewRows([]string{"is_admin"}).AddRow(false),
			expectedResult: false,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `SELECT`
			mock.ExpectQuery(query).WillReturnRows(tt.mockRows)

			result, err := IsAdmin(tt.userID)

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
