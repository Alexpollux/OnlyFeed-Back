package user

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
)

// GetUser GET /api/users/:id
func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user User

	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// Construction de la réponse avec condition sur isAdmin
	response := gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"firstname":  user.Firstname,
		"lastname":   user.Lastname,
		"avatar_url": user.AvatarURL,
		"bio":        user.Bio,
		"language":   user.Language,
		"theme":      user.Theme,
		"is_creator": user.IsCreator,
	}

	if user.IsAdmin {
		response["is_admin"] = true
	}
	if user.IsCreator {
		response["subscription_price"] = user.SubscriptionPrice
	}

	c.JSON(http.StatusOK, gin.H{"user": response})
}

// UpdateUser PATCH /api/users/:id
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user User

	// Vérifie que l'utilisateur existe
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// Bind les champs envoyés
	var input map[string]interface{}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requête invalide"})
		return
	}

	// Update uniquement les champs fournis
	if err := database.DB.Model(&user).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour utilisateur"})
		return
	}

	// Construction de la réponse avec condition sur isAdmin
	response := gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"firstname":  user.Firstname,
		"lastname":   user.Lastname,
		"avatar_url": user.AvatarURL,
		"bio":        user.Bio,
		"language":   user.Language,
		"theme":      user.Theme,
	}

	if user.IsAdmin {
		response["is_admin"] = true
	}
	if user.IsCreator {
		response["subscription_price"] = user.SubscriptionPrice
	}

	c.JSON(http.StatusOK, gin.H{"user": response})
}

// DeleteUser DELETE /api/users/:id
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	client := resty.New()
	supabaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	resp, err := client.R().
		SetHeader("apikey", supabaseServiceKey).
		SetHeader("Authorization", "Bearer "+supabaseServiceKey).
		Delete(supabaseURL + "/auth/v1/admin/users/" + id)

	if err != nil || resp.IsError() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur suppression Supabase", "details": resp.String()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur supprimé ✅"})
}

// 🆕 SearchUsers GET /api/users/search
func SearchUsers(c *gin.Context) {
	query := c.Query("q")
	currentUserID := c.GetString("user_id")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paramètre de recherche 'q' requis"})
		return
	}

	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La recherche doit contenir au moins 2 caractères"})
		return
	}

	var users []User
	// Recherche par username ou firstname/lastname, exclut l'utilisateur actuel
	if err := database.DB.
		Where("(username ILIKE ? OR firstname ILIKE ? OR lastname ILIKE ?) AND id != ?",
				"%"+query+"%", "%"+query+"%", "%"+query+"%", currentUserID).
		Limit(20). // Limiter à 20 résultats
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la recherche"})
		return
	}

	// Formatter la réponse
	var response []gin.H
	for _, user := range users {
		response = append(response, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"firstname":  user.Firstname,
			"lastname":   user.Lastname,
			"avatar_url": user.AvatarURL,
			"is_creator": user.IsCreator,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": response})
}
