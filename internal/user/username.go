package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/utils"
)

// GetUserByUsername GET /api/users/username/:username
func GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	currentUserID := c.GetString("user_id")

	var user User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		logs.LogJSON("WARN", "User not found", map[string]interface{}{
			"error":    err.Error(),
			"route":    "/api/users/username/:username",
			"username": username,
			"userID":   currentUserID,
		})
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
	var isSubscriber *bool
	var subscriptionPrice *float64

	if currentUserID != "" && currentUserID != user.ID {
		okFollow, err := utils.IsFollowing(currentUserID, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification du suivi"})
			logs.LogJSON("ERROR", "Error during follow-up verification", map[string]interface{}{
				"error":    err.Error(),
				"route":    "/api/users/username/:username",
				"username": username,
				"userID":   currentUserID,
			})
			return
		}
		isFollowing = &okFollow

		okSubscribe, okPrice, err := utils.IsSubscriberAndPrice(currentUserID, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification de l'abonnement"})
			logs.LogJSON("ERROR", "Subscription verification error", map[string]interface{}{
				"error":    err.Error(),
				"route":    "/api/users/username/:username",
				"username": username,
				"userID":   currentUserID,
			})
			return
		}
		isSubscriber = &okSubscribe
		subscriptionPrice = okPrice
	}

	if user.ID == currentUserID {
		dataUser["user"].(gin.H)["email"] = user.Email
		dataUser["user"].(gin.H)["firstname"] = user.Firstname
		dataUser["user"].(gin.H)["lastname"] = user.Lastname
		dataUser["user"].(gin.H)["language"] = user.Language
		dataUser["user"].(gin.H)["theme"] = user.Theme
		if user.IsAdmin {
			dataUser["user"].(gin.H)["is_admin"] = true
		}
	} else {
		dataUser["is_following"] = isFollowing
	}
	if user.IsCreator {
		dataUser["is_subscriber"] = isSubscriber
		if isSubscriber != nil && *isSubscriber {
			dataUser["subscription_price"] = subscriptionPrice
		} else {
			dataUser["subscription_price"] = user.SubscriptionPrice
		}
	}

	var followersCount, subscribersCount, totalPosts, paidPosts int64

	database.DB.Model(&utils.Follow{}).Where("creator_id = ?", user.ID).Count(&followersCount)
	database.DB.Model(&utils.Subscription{}).Where("creator_id = ?", user.ID).Count(&subscribersCount)
	database.DB.Table("posts").Where("user_id = ?", user.ID).Count(&totalPosts)
	database.DB.Table("posts").Where("user_id = ? AND is_paid = TRUE", user.ID).Count(&paidPosts)

	stats := dataUser["stats"].(gin.H)
	stats["followers_count"] = followersCount
	stats["subscribers_count"] = subscribersCount
	stats["posts_count"] = totalPosts
	stats["paid_posts_count"] = paidPosts

	if user.ID == currentUserID {
		var followupsCount, subscriptionsCount int64

		database.DB.Model(&utils.Follow{}).Where("follower_id = ?", user.ID).Count(&followupsCount)
		database.DB.Model(&utils.Subscription{}).Where("subscriber_id = ?", user.ID).Count(&subscriptionsCount)

		stats["followups_count"] = followupsCount
		stats["subscriptions_count"] = subscriptionsCount
	}

	c.JSON(http.StatusOK, dataUser)
	logs.LogJSON("INFO", "User fetched successfully", map[string]interface{}{
		"route":    "/api/users/username/:username",
		"username": username,
		"userID":   currentUserID,
	})
}
