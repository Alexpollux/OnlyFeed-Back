// internal/admin/handler.go
package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
)

// GetDashboardStats GET /api/admin/stats
func GetDashboardStats(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")

	// Paramètres optionnels
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	// Parse des dates si fournies
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format de date invalide pour start_date"})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // 30 jours par défaut
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format de date invalide pour end_date"})
			return
		}
	} else {
		endDate = time.Now()
	}

	// Statistiques générales
	var totalUsers, totalPosts, totalLikes, totalMessages int64
	var creatorsCount, premiumPosts int64

	// Total des utilisateurs
	database.DB.Table("users").Count(&totalUsers)

	// Total des créateurs
	database.DB.Table("users").Where("is_creator = true").Count(&creatorsCount)

	// Total des posts
	database.DB.Table("posts").Count(&totalPosts)

	// Total des posts premium
	database.DB.Table("posts").Where("is_paid = true").Count(&premiumPosts)

	// Total des likes
	database.DB.Table("likes").Count(&totalLikes)

	// Total des messages
	database.DB.Table("messages").Where("is_deleted = false").Count(&totalMessages)

	stats := gin.H{
		"total_users":    totalUsers,
		"total_posts":    totalPosts,
		"total_likes":    totalLikes,
		"total_messages": totalMessages,
		"creators_count": creatorsCount,
		"premium_posts":  premiumPosts,
		"date_range": gin.H{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
		},
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
	logs.LogJSON("INFO", "Admin stats retrieved successfully", map[string]interface{}{
		"route":  route,
		"userID": userID,
	})
}

// GetChartData GET /api/admin/charts/:type
func GetChartData(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")
	chartType := c.Param("type")

	// Paramètres
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			startDate = time.Now().AddDate(0, 0, -30)
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			endDate = time.Now()
		}
	} else {
		endDate = time.Now()
	}

	switch chartType {
	case "evolution":
		data := getEvolutionData(startDate, endDate)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "distribution":
		data := getDistributionData(startDate, endDate)
		c.JSON(http.StatusOK, gin.H{"data": data})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type de graphique non supporté"})
		return
	}

	logs.LogJSON("INFO", "Chart data retrieved successfully", map[string]interface{}{
		"route":     route,
		"userID":    userID,
		"chartType": chartType,
		"startDate": startDate.Format("2006-01-02"),
		"endDate":   endDate.Format("2006-01-02"),
	})
}

func getEvolutionData(startDate, endDate time.Time) []gin.H {
	var results []gin.H

	// Générer les données jour par jour
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		dayStart := d
		dayEnd := d.AddDate(0, 0, 1)

		var usersCount, postsCount, likesCount, messagesCount int64

		// Utilisateurs créés ce jour
		database.DB.Table("users").
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&usersCount)

		// Posts créés ce jour
		database.DB.Table("posts").
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&postsCount)

		// Likes créés ce jour
		database.DB.Table("likes").
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&likesCount)

		// Messages créés ce jour
		database.DB.Table("messages").
			Where("created_at >= ? AND created_at < ? AND is_deleted = false", dayStart, dayEnd).
			Count(&messagesCount)

		results = append(results, gin.H{
			"date":     d.Format("2006-01-02"),
			"users":    usersCount,
			"posts":    postsCount,
			"likes":    likesCount,
			"messages": messagesCount,
		})
	}

	return results
}

func getDistributionData(startDate, endDate time.Time) []gin.H {
	var freePosts, paidPosts, totalLikes, totalMessages int64

	// Posts gratuits dans la période
	database.DB.Table("posts").
		Where("created_at >= ? AND created_at <= ? AND is_paid = false", startDate, endDate).
		Count(&freePosts)

	// Posts payants dans la période
	database.DB.Table("posts").
		Where("created_at >= ? AND created_at <= ? AND is_paid = true", startDate, endDate).
		Count(&paidPosts)

	// Likes dans la période
	database.DB.Table("likes").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Count(&totalLikes)

	// Messages dans la période
	database.DB.Table("messages").
		Where("created_at >= ? AND created_at <= ? AND is_deleted = false", startDate, endDate).
		Count(&totalMessages)

	return []gin.H{
		{"name": "Posts gratuits", "value": freePosts, "color": "#3B82F6"},
		{"name": "Posts premium", "value": paidPosts, "color": "#F59E0B"},
		{"name": "Likes", "value": totalLikes, "color": "#EF4444"},
		{"name": "Messages", "value": totalMessages, "color": "#8B5CF6"},
	}
}

// GetTopUsers GET /api/admin/top-users
func GetTopUsers(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Top utilisateurs par nombre de posts
	var topByPosts []struct {
		UserID    string `json:"user_id"`
		Username  string `json:"username"`
		PostCount int64  `json:"post_count"`
	}

	database.DB.Table("posts").
		Select("posts.user_id, users.username, COUNT(posts.id) as post_count").
		Joins("LEFT JOIN users ON posts.user_id = users.id").
		Group("posts.user_id, users.username").
		Order("post_count DESC").
		Limit(limit).
		Scan(&topByPosts)

	// Top utilisateurs par nombre de likes reçus
	var topByLikes []struct {
		UserID     string `json:"user_id"`
		Username   string `json:"username"`
		LikesCount int64  `json:"likes_count"`
	}

	database.DB.Table("likes").
		Select("posts.user_id, users.username, COUNT(likes.id) as likes_count").
		Joins("LEFT JOIN posts ON likes.post_id = posts.id").
		Joins("LEFT JOIN users ON posts.user_id = users.id").
		Group("posts.user_id, users.username").
		Order("likes_count DESC").
		Limit(limit).
		Scan(&topByLikes)

	c.JSON(http.StatusOK, gin.H{
		"top_by_posts": topByPosts,
		"top_by_likes": topByLikes,
	})

	logs.LogJSON("INFO", "Top users retrieved successfully", map[string]interface{}{
		"route":  route,
		"userID": userID,
		"limit":  limit,
	})
}
