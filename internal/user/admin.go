package user

import (
	"errors"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"gorm.io/gorm"
)

// IsAdmin vérifie si un utilisateur est admin à partir de son ID
func IsAdmin(userID string) (bool, error) {
	var isAdmin bool
	if err := database.DB.Model(&User{}).Select("is_admin").Where("id = ?", userID).Scan(&isAdmin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // utilisateur introuvable, donc pas admin
		}
		return false, err // erreur autre
	}
	return isAdmin, nil
}
