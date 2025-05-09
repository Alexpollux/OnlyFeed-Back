package user

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/storage"
)

func GetMe(c *gin.Context) {
	userID := c.GetString("user_id")

	var user User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
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
		"created_at": user.CreatedAt,
	}

	if user.IsAdmin {
		response["is_admin"] = true
	}

	c.JSON(http.StatusOK, gin.H{"user": response})
}

func UpdateMe(c *gin.Context) {
	userID := c.GetString("user_id")

	var user User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// Parse des champs texte
	username := c.PostForm("username")
	firstname := c.PostForm("firstname")
	lastname := c.PostForm("lastname")
	bio := c.PostForm("bio")
	language := c.PostForm("language")

	if username != "" {
		user.Username = username
	}
	if firstname != "" {
		user.Firstname = firstname
	}
	if lastname != "" {
		user.Lastname = lastname
	}
	if bio != "" {
		user.Bio = bio
	}
	if language != "" {
		user.Language = language
	}

	// Vérification et remplacement de la photo
	file, header, err := c.Request.FormFile("profile_picture")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		validExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".heic": true}
		if !validExtensions[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Extension fichier invalide"})
			return
		}

		filename := fmt.Sprintf("user_%s%s", userID, ext)
		contentType := header.Header.Get("Content-Type")

		// Supprimer ancienne image si existante
		if user.AvatarURL != "" {
			oldKey := filepath.Base(user.AvatarURL)
			if err := storage.DeleteFromS3("avatars/" + oldKey); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur suppression ancienne image", "details": err.Error()})
				return
			}
		}

		// Uploader nouvelle image
		url, err := storage.UploadToS3(file, filename, contentType, "avatars")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur upload S3", "details": err.Error()})
			return
		}
		user.AvatarURL = url
	}

	// Sauvegarde finale
	if err := database.DB.Save(&user).Error; err != nil {
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
		"created_at": user.CreatedAt,
	}

	if user.IsAdmin {
		response["is_admin"] = true
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profil mis à jour", "user": response})
}
