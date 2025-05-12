package user

import "time"

type User struct {
	ID        string `gorm:"primaryKey"` // UUID venant de auth.users
	CreatedAt time.Time
	Username  string
	Firstname string
	Lastname  string
	AvatarURL string
	Bio       string
	Email     string
	Language  string
	Theme     string
	IsAdmin   bool
}
