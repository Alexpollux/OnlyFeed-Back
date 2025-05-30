// internal/post/comment.go
package post

import (
	"time"
)

type Comment struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID    string    `json:"post_id" gorm:"index"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username" gorm:"-"`          // Ne pas stocker en DB, rempli dynamiquement
	Content   string    `json:"text" gorm:"column:content"` // JSON "text" -> DB "content"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Structure pour la création d'un commentaire
type CreateCommentInput struct {
	PostID string `json:"post_id" binding:"required"` // Changé de uint à string
	Text   string `json:"text" binding:"required"`
}
