package follow

import (
	"time"
)

type Follow struct {
	ID         string `gorm:"primaryKey"`
	CreatedAt  time.Time
	FollowerID string `gorm:"type:uuid"`
	CreatorID  string `gorm:"type:uuid"`
}
