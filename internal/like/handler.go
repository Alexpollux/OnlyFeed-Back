// internal/like/handler.go
package like

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
)

// ToggleLike POST/DELETE /api/posts/:id/like
func ToggleLike(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")
	postID := c.Param("id")
}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		logs.LogJSON("WARN", "Unauthenticated user", map[string]interface{}{
			"route": route,
			"postID": postID,
		})
		return
	}

	// ✅ Vérifier si le post existe (CORRECTION)
	var postCount int64
	if err := database.DB.Table("posts").Where("id = ?", postID).Count(&postCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de base de données"})
		logs.LogJSON("ERROR", "Database error", map[string]interface{}{
			"error":    err.Error(),
			"route":    route,
			"userID":   userID,
			"postID":   postID,
		})
		
		return
	}
	if postCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post non trouvé"})
		logs.LogJSON("WARN", "Post not found", map[string]interface{}{
			"route":    route,
			"userID":   userID,
			"postID":   postID,
		})
		return
	}

	// Vérifier si l'utilisateur a déjà liké ce post
	var existingLike Like
	err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingLike).Error

	if err == nil {
		// Le like existe, on le supprime (unlike)
		if err := database.DB.Delete(&existingLike).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression du like"})
			logs.LogJSON("ERROR", "Error when unliking", map[string]interface{}{
				"error":    err.Error(),
				"route":    route,
				"userID":   userID,
				"postID":   postID,
			})
			return
		}
	} else if err == gorm.ErrRecordNotFound {
		// Le like n'existe pas, on le crée
		newLike := Like{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
			UserID:    userID,
			PostID:    postID,
		}

		if err := database.DB.Create(&newLike).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'ajout du like"})
			logs.LogJSON("ERROR", "Error when liking", map[string]interface{}{
				"error":    err.Error(),
			return
		})
	} else {
		// Erreur de base de données
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de base de données"})
		logs.LogJSON("ERROR", "Database error", map[string]interface{}{
			"error":    err.Error(),
			"route":    route,
			"userID":   userID,
			"postID":   postID,
		})
		return
	}

	// Retourner le statut mis à jour
	response := getLikeStatus(postID, userID)
	c.JSON(http.StatusOK, response)
}

// GetLikeStatus GET /api/posts/:id/likes
func GetLikeStatus(c *gin.Context) {
	route := c.FullPath()
	postID := c.Param("id")
	userID := c.GetString("user_id") // Peut être vide si non connecté

	// ✅ Vérifier si le post existe (CORRECTION)
	var postCount int64
	if err := database.DB.Table("posts").Where("id = ?", postID).Count(&postCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de base de données"})
		logs.LogJSON("ERROR", "Database error", map[string]interface{}{
			"error":    err.Error(),
			"route":    route,
			"userID":   userID,
			"postID":   postID,
		})
		return
	}
	if postCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post non trouvé"})
		logs.LogJSON("WARN", "Post not found", map[string]interface{}{
			"route":    route,
			"userID":   userID,
			"postID":   postID,
		})
		return
	}

	response := getLikeStatus(postID, userID)
	c.JSON(http.StatusOK, response)
}

// ✅ NOUVELLE FONCTION - GetPostByIDWithLikes GET /api/posts/:id (version avec likes)
func GetPostByIDWithLikes(c *gin.Context) {
	route := c.FullPath()
	postID := c.Param("id")
	userID := c.GetString("user_id") // Peut être vide si non connecté

	// Récupérer le post
	var post struct {
		ID          string    `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UserID      string    `json:"user_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		MediaURL    string    `json:"media_url"`
		IsPaid      bool      `json:"is_paid"`
	}

	if err := database.DB.Table("posts").Where("id = ?", postID).First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post non trouvé"})
		logs.LogJSON("WARN", "Post not found", map[string]interface{}{
			"route":    route,
			"userID":   userID,
			"postID":   postID,
		})
		return
	}

	// Vérification des permissions pour les posts payants
	if post.IsPaid {
		if userID == "" || post.UserID != userID {
			// Ici vous pourriez implémenter une vérification d'abonnement
			// Pour l'instant, seul le créateur peut voir son propre post payant
			c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé à ce contenu premium"})
			logs.LogJSON("WARN", "Unauthorized access to premium content", map[string]interface{}{
				"route":    route,
				"userID":   userID,
				"postID":   postID,
			})
			return
		}
	}

	// Ajouter les informations de likes
	likeStatus := getLikeStatus(postID, userID)

	// Construire la réponse avec le format attendu par le frontend
	response := gin.H{
		"ID":          post.ID,
		"Title":       post.Title,
		"Description": post.Description,
		"MediaURL":    post.MediaURL,
		"IsPaid":      post.IsPaid,
		"CreatedAt":   post.CreatedAt,
		"UserID":      post.UserID,
		"like_count":  likeStatus.LikeCount,
		"is_liked":    likeStatus.IsLiked,
	}

	c.JSON(http.StatusOK, response)
}

// GetPostsWithLikes GET /api/posts (version étendue avec likes)
func GetPostsWithLikes(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")
	showPaywalled := c.Query("paywalled") == "true"

	query := database.DB.Order("created_at DESC")

	if !showPaywalled || userID == "" {
		query = query.Where("is_paid = ?", false)
	} else {
		query = query.Where("is_paid = ? OR (is_paid = ? AND user_id = ?)", false, true, userID)
	}

	var posts []struct {
		ID          string    `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UserID      string    `json:"user_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		MediaURL    string    `json:"media_url"`
		IsPaid      bool      `json:"is_paid"`
	}

	if err := query.Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des posts"})
		logs.LogJSON("ERROR", "Error during data retrieval", map[string]interface{}{
			"error":    err.Error(),
			"route":    route,
			"userID":   userID,
		})
		return

	}

	// Ajouter les informations de likes pour chaque post
	var postsWithLikes []gin.H
	for _, post := range posts {
		likeStatus := getLikeStatus(post.ID, userID)

		postWithLikes := gin.H{
			"id":          post.ID,
			"created_at":  post.CreatedAt,
			"user_id":     post.UserID,
			"title":       post.Title,
			"description": post.Description,
			"media_url":   post.MediaURL,
			"is_paid":     post.IsPaid,
			"like_count":  likeStatus.LikeCount,
			"is_liked":    likeStatus.IsLiked,
		}
		postsWithLikes = append(postsWithLikes, postWithLikes)
	}

	c.JSON(http.StatusOK, gin.H{"posts": postsWithLikes})
	logs.LogJSON("INFO", "Posts retrieved successfully", map[string]interface{}{
		"route":    route,
		"userID":   userID,
	})	
}

// Fonction utilitaire pour obtenir le statut des likes
func getLikeStatus(postID, userID string) LikeResponse {
	var likeCount int64
	database.DB.Model(&Like{}).Where("post_id = ?", postID).Count(&likeCount)

	var isLiked bool
	if userID != "" {
		var existingLike Like
		err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingLike).Error
		isLiked = (err == nil)
	}

	return LikeResponse{
		PostID:    postID,
		LikeCount: likeCount,
		IsLiked:   isLiked,
	}
}
