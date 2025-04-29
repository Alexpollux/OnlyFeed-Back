package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/storage"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

func Signup(c *gin.Context) {
	supabaseBaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")

	// Lecture du form-data
	email := c.PostForm("email")
	password := c.PostForm("password")
	username := c.PostForm("username")
	firstname := c.PostForm("firstname")
	lastname := c.PostForm("lastname")
	bio := c.PostForm("bio")
	language := c.PostForm("language")

	if email == "" || password == "" || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Champs requis manquants"})
		return
	}

	// Vérification unique email/username
	if user.ExistsByEmail(email) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email déjà utilisé"})
		return
	}
	if user.ExistsByUsername(username) {
		c.JSON(http.StatusConflict, gin.H{"error": "Nom d'utilisateur déjà utilisé"})
		return
	}

	// Vérification de la langue
	validLanguages := map[string]bool{"fr": true, "en": true}
	if language != "" && !validLanguages[language] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Langue non supportée"})
		return
	}

	// Étape 1 – Créer l'utilisateur dans Supabase Auth
	authPayload := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonBody, _ := json.Marshal(authPayload)
	req, _ := http.NewRequest("POST", supabaseBaseURL+"/auth/v1/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("apikey", supabaseAnonKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur Supabase Auth"})
		return
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{"error": "Erreur Auth", "details": string(respBytes)})
		return
	}

	var authResp struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur parsing user.id"})
		return
	}

	userID := authResp.User.ID
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aucun ID utilisateur renvoyé"})
		return
	}

	// Étape 2 – Upload avatar si présent
	var avatarURL string

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

		url, err := storage.UploadToS3(file, filename, contentType, "avatars")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur upload S3", "details": err.Error()})
			return
		}
		avatarURL = url
	}

	// Étape 3 – Enregistrement final en BDD
	newUser := user.User{
		ID:        userID,
		CreatedAt: time.Now(),
		Username:  username,
		Firstname: firstname,
		Lastname:  lastname,
		AvatarURL: avatarURL,
		Bio:       bio,
		Email:     email,
		Language:  language,
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur insertion base utilisateurs"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Utilisateur inscrit 🎉",
		"user":       newUser,
		"avatar_url": avatarURL,
	})
}

func Login(c *gin.Context) {
	supabaseBaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL") // idem ici
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requête invalide"})
		return
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(
		"POST",
		supabaseBaseURL+"/auth/v1/token?grant_type=password",
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur connexion Supabase"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}
