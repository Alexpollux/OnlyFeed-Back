package user

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
)

// GetUser GET /api/users/:id
func GetUser(c *gin.Context) {
	route := c.FullPath()

	currentUserID := c.GetString("user_id")

	id := c.Param("id")
	var user User

	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		logs.LogJSON("WARN", "User not found", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": currentUserID,
			"extra":  fmt.Sprintf("User not found : %s", id),
		})
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
	logs.LogJSON("INFO", "User fetched successfully", map[string]interface{}{
		"route":  route,
		"userID": currentUserID,
		"extra":  fmt.Sprintf("User fetched successfully : %s", id),
	})
}

// UpdateUser PATCH /api/users/:id
func UpdateUser(c *gin.Context) {
	route := c.FullPath()

	currentUserID := c.GetString("user_id")

	id := c.Param("id")
	var user User

	// Vérifie que l'utilisateur existe
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		logs.LogJSON("WARN", "User not found", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": currentUserID,
			"extra":  fmt.Sprintf("User not found : %s", id),
		})
		return
	}

	// Bind les champs envoyés
	var input map[string]interface{}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requête invalide"})
		logs.LogJSON("ERROR", "Invalid request", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": currentUserID,
			"extra":  fmt.Sprintf("Invalid request : %p", input),
		})
		return
	}

	// Update uniquement les champs fournis
	if err := database.DB.Model(&user).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour utilisateur"})
		logs.LogJSON("ERROR", "User update error", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": currentUserID,
			"extra":  fmt.Sprintf("User update error : %p", input),
		})
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
	logs.LogJSON("INFO", "User updated successfully", map[string]interface{}{
		"route":  route,
		"userID": currentUserID,
		"extra":  fmt.Sprintf("User updated successfully : %s", id),
	})
}

// DeleteUser DELETE /api/users/:id
func DeleteUser(c *gin.Context) {
	route := c.FullPath()

	currentUserID := c.GetString("user_id")

	id := c.Param("id")

	client := resty.New()
	supabaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	resp, err := client.R().
		SetHeader("apikey", supabaseServiceKey).
		SetHeader("Authorization", "Bearer "+supabaseServiceKey).
		Delete(supabaseURL + "/auth/v1/admin/users/" + id)

	if err != nil || resp.IsError() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur Supabase de suppression d'utilisateur", "details": resp.String()})
		logs.LogJSON("ERROR", "Supabase user deletion error", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": currentUserID,
			"extra":  fmt.Sprintf("Supabase user deletion error : %p", resp),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur supprimé"})
	logs.LogJSON("INFO", "User deleted successfully", map[string]interface{}{
		"route":  route,
		"userID": currentUserID,
		"extra":  fmt.Sprintf("User deleted successfully : %s", id),
	})
}

// SearchUsers GET /api/users/search
func SearchUsers(c *gin.Context) {
	route := c.FullPath()

	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paramètre de recherche 'q' requis"})
		logs.LogJSON("WARN", "Search parameter ‘q’ required", map[string]interface{}{
			"route": route,
		})
		return
	}

	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La recherche doit contenir au moins 2 caractères"})
		logs.LogJSON("WARN", "The search must contain at least 2 characters", map[string]interface{}{
			"route": route,
			"extra": fmt.Sprintf("The search is : %s", query),
		})
		return
	}

	var users []User
	// Recherche par username ou firstname/lastname
	if err := database.DB.
		Where("username ILIKE ? OR firstname ILIKE ? OR lastname ILIKE ?",
				"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(20). // Limiter à 20 résultats
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la recherche"})
		logs.LogJSON("WARN", "Search error", map[string]interface{}{
			"route": route,
			"extra": fmt.Sprintf("The search is : %s", query),
		})
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
	logs.LogJSON("INFO", "User search is successful", map[string]interface{}{
		"route": route,
		"extra": fmt.Sprintf("The search is : %s", query),
	})
}
