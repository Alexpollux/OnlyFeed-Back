package post

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

// GetPostsByUsername GET /api/users/username/:username/posts
func GetPostsByUsername(c *gin.Context) {
	username := c.Param("username")

	var u user.User
	if err := database.DB.Where("username = ?", username).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur introuvable"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de récupération de l'utilisateur"})
		}
		return
	}

	requesterID := c.GetString("user_id")

	var isSubscribed bool
	if requesterID != "" {
		var count int64
		database.DB.Table("subscriptions").Where("subscriber_id = ? AND creator_id = ?", requesterID, u.ID).Count(&count)
		isSubscribed = count > 0
	}

	var posts []Post
	query := database.DB.Preload("User").Where("user_id = ?", u.ID)
	if !isSubscribed {
		query = query.Where("is_paid = false")
	}

	if err := query.Order("created_at DESC").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de récupération des posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}
