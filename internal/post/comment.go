// internal/post/comment.go
package post

import (
	"time"
)

type Comment struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID    string    `json:"post_id" gorm:"index"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"text" gorm:"column:content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
