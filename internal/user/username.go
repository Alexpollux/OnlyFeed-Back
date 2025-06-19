package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/utils"
)

// GetUserByUsername GET /api/users/username/:username
func GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	currentUserID := c.GetString("user_id")

	var user User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// On retourne uniquement les champs publics
	dataUser := gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"avatar_url": user.AvatarURL,
			"bio":        user.Bio,
			"is_creator": user.IsCreator,
		},
		"stats": gin.H{},
	}

	var isFollowing *bool

	if currentUserID != "" && currentUserID != user.ID {
		ok, err := utils.IsFollowing(currentUserID, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification du suivi"})
			return
		}
		isFollowing = &ok
	}

	if user.ID == currentUserID {
		dataUser["user"].(gin.H)["email"] = user.Email
		dataUser["user"].(gin.H)["firstname"] = user.Firstname
		dataUser["user"].(gin.H)["lastname"] = user.Lastname
		dataUser["user"].(gin.H)["language"] = user.Language
		dataUser["user"].(gin.H)["theme"] = user.Theme
	} else {
		dataUser["is_following"] = isFollowing
	}

	var followersCount, subscribersCount, totalPosts, paidPosts int64

	database.DB.Model(&utils.Follow{}).Where("creator_id = ?", user.ID).Count(&followersCount)
	database.DB.Table("subscriptions").Where("creator_id = ?", user.ID).Count(&subscribersCount)
	database.DB.Table("posts").Where("user_id = ?", user.ID).Count(&totalPosts)
	database.DB.Table("posts").Where("user_id = ? AND is_paid = TRUE", user.ID).Count(&paidPosts)

	stats := dataUser["stats"].(gin.H)
	stats["followers_count"] = followersCount
	stats["subscribers_count"] = subscribersCount
	stats["posts_count"] = totalPosts
	stats["paid_posts_count"] = paidPosts

	c.JSON(http.StatusOK, dataUser)
}
