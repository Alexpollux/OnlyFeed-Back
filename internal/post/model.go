package post

import (
	"time"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

type Post struct {
	ID          string `gorm:"primaryKey"` // UUID venant de auth.users
	CreatedAt   time.Time
	UserID      string
	User        user.User `gorm:"foreignKey:UserID"`
	Title       string
	Description string
	MediaURL    string
	IsPaid      bool
}
