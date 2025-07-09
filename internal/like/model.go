package like

import (
	"time"
)

type Like struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`
	UserID    string    `json:"user_id" gorm:"index"`
	PostID    string    `json:"post_id" gorm:"index"`
}

type LikeResponse struct {
	PostID    string `json:"post_id"`
	LikeCount int64  `json:"like_count"`
	IsLiked   bool   `json:"is_liked"`
}

func (Like) TableName() string {
	return "likes"
}
