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
	var u User

	if err := database.DB.First(&u, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	c.JSON(http.StatusOK, u)
}

// UpdateUser PATCH /api/users/:id
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var u User

	// Vérifie que l'utilisateur existe
	if err := database.DB.First(&u, "id = ?", id).Error; err != nil {
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
	if err := database.DB.Model(&u).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour utilisateur"})
		return
	}

	c.JSON(http.StatusOK, u)
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
