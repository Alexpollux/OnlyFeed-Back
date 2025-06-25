package subscription

import "time"

type Subscription struct {
	ID                   string `gorm:"primaryKey"`
	CreatedAt            time.Time
	SubscriberID         string
	CreatorID            string
	Status               string
	StripeSubscriptionID string
	Price                float64
}
