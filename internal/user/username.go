package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
)

// GetUserByUsername GET /api/users/username/:username
func GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	var user User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// On retourne uniquement les champs publics
	publicUser := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"avatar_url": user.AvatarURL,
		"bio":        user.Bio,
	}

	c.JSON(http.StatusOK, gin.H{"user": publicUser})
}
