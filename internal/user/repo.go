package user

import "github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"

func ExistsByEmail(email string) bool {
	var count int64
	database.DB.Model(&User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func ExistsByUsername(username string) bool {
	var count int64
	database.DB.Model(&User{}).Where("username = ?", username).Count(&count)
	return count > 0
}
