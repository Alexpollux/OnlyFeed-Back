package follow

import (
	"fmt"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
	"net/http"

	"github.com/google/uuid"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
	"github.com/gin-gonic/gin"
)

// FollowUser POST /api/follow/:id
func FollowUser(c *gin.Context) {
	route := c.FullPath()

	followerID := c.GetString("user_id")
	followingID := c.Param("id")

	if followerID == followingID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Impossible de se suivre soi-même"})
		logs.LogJSON("WARN", "Impossible to follow yourself", map[string]interface{}{
			"route":  route,
			"userID": followerID,
		})
		return
	}

	var existing Follow
	if err := database.DB.
		Where("follower_id = ? AND creator_id = ?", followerID, followingID).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Déjà suivi"})
		logs.LogJSON("WARN", "Already followed", map[string]interface{}{
			"route":  route,
			"userID": followerID,
			"extra":  fmt.Sprintf("followingID : %s", followingID),
		})
		return
	}

	newFollow := Follow{
		ID:         uuid.New().String(),
		FollowerID: followerID,
		CreatorID:  followingID,
	}

	if err := database.DB.Create(&newFollow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur ajout du follow", "details": err.Error()})
		logs.LogJSON("ERROR", "Error adding follow", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": followerID,
			"extra":  fmt.Sprintf("followingID : %s", followingID),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Utilisateur suivi"})
	logs.LogJSON("INFO", "Followed user", map[string]interface{}{
		"route":  route,
		"userID": followerID,
		"extra":  fmt.Sprintf("followingID : %s", followingID),
	})
}

// UnfollowUser DELETE /api/follow/:id
func UnfollowUser(c *gin.Context) {
	route := c.FullPath()

	followerID := c.GetString("user_id")
	followingID := c.Param("id")

	// Supprime le follow
	if err := database.DB.
		Where("follower_id = ? AND creator_id = ?", followerID, followingID).
		Delete(&Follow{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur unfollow"})
		logs.LogJSON("ERROR", "Error unfollow", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": followerID,
			"extra":  fmt.Sprintf("followingID : %s", followingID),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur unfollow"})
	logs.LogJSON("INFO", "User unfollow", map[string]interface{}{
		"route":  route,
		"userID": followerID,
		"extra":  fmt.Sprintf("followingID : %s", followingID),
	})
}

// GetFollowing GET /api/following
func GetFollowing(c *gin.Context) {
	route := c.FullPath()

	followerID := c.GetString("user_id")

	var follows []Follow
	if err := database.DB.
		Where("follower_id = ?", followerID).
		Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération creator"})
		logs.LogJSON("ERROR", "Error recovery creator", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": followerID,
		})
		return
	}

	var usersFollowed []user.User
	var ids []string
	for _, f := range follows {
		ids = append(ids, f.CreatorID)
	}

	if err := database.DB.Where("id IN ?", ids).Find(&usersFollowed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération des utilisateurs suivis"})
		logs.LogJSON("ERROR", "Error retrieving followed users", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": followerID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"following": usersFollowed})
	logs.LogJSON("INFO", "Recovering the list of users followed", map[string]interface{}{
		"route":  route,
		"userID": followerID,
	})
}

// GetFollowers GET /api/followers/:id
func GetFollowers(c *gin.Context) {
	route := c.FullPath()

	userID := c.Param("id")

	var follows []Follow
	if err := database.DB.
		Where("creator_id = ?", userID).
		Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération followers"})
		logs.LogJSON("ERROR", "Error recovery followers", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": userID,
		})
		return
	}

	var followers []user.User
	var ids []string
	for _, f := range follows {
		ids = append(ids, f.FollowerID)
	}

	if err := database.DB.Where("id IN ?", ids).Find(&followers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération des utilisateurs followers"})
		logs.LogJSON("ERROR", "Error recovering followers", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": userID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"followers": followers})
	logs.LogJSON("INFO", "Recovering the list of user followers", map[string]interface{}{
		"route":  route,
		"userID": userID,
	})
}
