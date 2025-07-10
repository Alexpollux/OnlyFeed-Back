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

func TestIsSubscriberAndPrice(t *testing.T) {
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
		name                 string
		subscriberID         string
		creatorID            string
		mockRows             *sqlmock.Rows
		expectedIsSubscriber bool
		expectedPrice        *float64
		expectedError        bool
	}{
		{
			name:         "Active subscription exists",
			subscriberID: "subscriber1",
			creatorID:    "creator1",
			mockRows: sqlmock.NewRows([]string{"id", "created_at", "subscriber_id", "creator_id", "status", "stripe_subscription_id", "price"}).
				AddRow("sub1", time.Now(), "subscriber1", "creator1", "active", "stripe_sub_123", 9.99),
			expectedIsSubscriber: true,
			expectedPrice:        func() *float64 { p := 9.99; return &p }(),
			expectedError:        false,
		},
		{
			name:                 "No subscription exists",
			subscriberID:         "subscriber1",
			creatorID:            "creator1",
			mockRows:             sqlmock.NewRows([]string{"id", "created_at", "subscriber_id", "creator_id", "status", "stripe_subscription_id", "price"}),
			expectedIsSubscriber: false,
			expectedPrice:        nil,
			expectedError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `SELECT`
			mock.ExpectQuery(query).WillReturnRows(tt.mockRows)

			isSubscriber, price, err := IsSubscriberAndPrice(tt.subscriberID, tt.creatorID)

			assert.Equal(t, tt.expectedIsSubscriber, isSubscriber)
			assert.Equal(t, tt.expectedPrice, price)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
