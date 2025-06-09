package follow

import (
	"github.com/google/uuid"
	"net/http"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	//"github.com/ArthurDelaporte/OnlyFeed-Back/internal/subscription"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
	"github.com/gin-gonic/gin"
)

// FollowUser POST /api/follow/:id
func FollowUser(c *gin.Context) {
	followerID := c.GetString("user_id")
	followingID := c.Param("id")

	if followerID == followingID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Impossible de se suivre soi-même"})
		return
	}

	var existing Follow
	if err := database.DB.
		Where("follower_id = ? AND creator_id = ?", followerID, followingID).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Déjà suivi"})
		return
	}

	newFollow := Follow{
		ID:         uuid.New().String(),
		FollowerID: followerID,
		CreatorID:  followingID,
	}

	if err := database.DB.Create(&newFollow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur ajout du follow", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Utilisateur suivi ✅"})
}

// UnfollowUser DELETE /api/follow/:id
func UnfollowUser(c *gin.Context) {
	followerID := c.GetString("user_id")
	followingID := c.Param("id")

	//// Vérifie s’il y a une subscription active
	//var sub subscription.Subscription
	//if err := database.DB.
	//	Where("subscriber_id = ? AND creator_id = ? AND status = ?", followerID, followingID, "active").
	//	First(&sub).Error; err == nil {
	//	c.JSON(http.StatusForbidden, gin.H{"error": "Impossible de unfollow un créateur abonné actif"})
	//	return
	//}

	// Supprime le follow
	if err := database.DB.
		Where("follower_id = ? AND creator_id = ?", followerID, followingID).
		Delete(&Follow{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur unfollow"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur unfollow ✅"})
}

// GetFollowing GET /api/following
func GetFollowing(c *gin.Context) {
	followerID := c.GetString("user_id")

	var follows []Follow
	if err := database.DB.
		Where("follower_id = ?", followerID).
		Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération creator"})
		return
	}

	var usersFollowed []user.User
	var ids []string
	for _, f := range follows {
		ids = append(ids, f.CreatorID)
	}

	if err := database.DB.Where("id IN ?", ids).Find(&usersFollowed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération users suivis"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"following": usersFollowed})
}

// GetFollowers GET /api/followers/:id
func GetFollowers(c *gin.Context) {
	userID := c.Param("id")

	var follows []Follow
	if err := database.DB.
		Where("creator_id = ?", userID).
		Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération followers"})
		return
	}

	var followers []user.User
	var ids []string
	for _, f := range follows {
		ids = append(ids, f.FollowerID)
	}

	if err := database.DB.Where("id IN ?", ids).Find(&followers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération users followers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"followers": followers})
}
