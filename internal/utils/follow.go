package utils

import (
	"errors"
	"gorm.io/gorm"
	"time"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
)

type Follow struct {
	ID         string `gorm:"primaryKey"`
	CreatedAt  time.Time
	FollowerID string
	CreatorID  string
}

func IsFollowing(followerID, followingID string) (bool, error) {
	var follow Follow
	err := database.DB.
		Where("follower_id = ? AND creator_id = ?", followerID, followingID).
		First(&follow).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // L'utilisateur ne suit pas
		}
		return false, err // Une erreur s'est produite
	}

	return true, nil // L'utilisateur suit
}
