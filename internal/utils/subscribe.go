package utils

import (
	"errors"
	"gorm.io/gorm"
	"time"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
)

type Subscription struct {
	ID                   string `gorm:"primaryKey"`
	CreatedAt            time.Time
	SubscriberID         string
	CreatorID            string
	Status               string
	StripeSubscriptionID string
	Price                float64
}

func IsSubscriberAndPrice(subscriberID, creatorID string) (bool, *float64, error) {
	var subscription Subscription
	err := database.DB.
		Where("subscriber_id = ? AND creator_id = ?", subscriberID, creatorID).
		First(&subscription).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil // L'utilisateur ne suit pas
		}
		return false, nil, err // Une erreur s'est produite
	}

	return subscription.Status == "active", &subscription.Price, nil // L'utilisateur suit
}
