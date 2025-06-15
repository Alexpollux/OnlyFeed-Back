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
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/utils"
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
	theme := c.PostForm("theme")
	if theme != "dark" {
		theme = "light"
	}

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
		Theme:     theme,
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur insertion base utilisateurs"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Utilisateur inscrit 🎉",
		"user":    newUser,
	})
}

func Login(c *gin.Context) {
	supabaseBaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil || input.Email == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Champs email et password requis"})
		return
	}

	// Préparation de la requête
	payload := map[string]string{
		"email":    input.Email,
		"password": input.Password,
	}
	jsonBody, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		"POST",
		supabaseBaseURL+"/auth/v1/token?grant_type=password",
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("apikey", supabaseAnonKey)
	req.Header.Set("Content-Type", "application/json")

	// Exécution
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de connexion à Supabase"})
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error":   "Erreur d'authentification",
			"details": string(bodyBytes),
		})
		return
	}

	// Parsing de la réponse Supabase
	var supabaseResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.Unmarshal(bodyBytes, &supabaseResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de parsing de la réponse Supabase"})
		return
	}

	// Parser manuellement l'access_token pour extraire user_id (champ "sub")
	tokenClaims := utils.ParseJWTClaims(supabaseResp.AccessToken)
	userID, ok := tokenClaims["sub"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID utilisateur manquant dans le token"})
		return
	}

	// Récupération du user depuis ta base
	var u user.User
	if err := database.DB.First(&u, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur introuvable en base"})
		return
	}

	// Construction de la réponse avec condition sur isAdmin
	respUser := gin.H{
		"id":         u.ID,
		"email":      u.Email,
		"username":   u.Username,
		"firstname":  u.Firstname,
		"lastname":   u.Lastname,
		"avatar_url": u.AvatarURL,
		"bio":        u.Bio,
		"language":   u.Language,
		"theme":      u.Theme,
	}

	if u.IsAdmin {
		respUser["is_admin"] = true
	}

	// Réponse personnalisée
	c.JSON(http.StatusOK, gin.H{
		"access_token":  supabaseResp.AccessToken,
		"refresh_token": supabaseResp.RefreshToken,
		"user":          respUser,
	})
}

func Logout(c *gin.Context) {
	supabaseBaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")

	authHeader := c.GetHeader("Authorization")
	refreshToken := c.GetHeader("X-Refresh-Token")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token requis"})
		return
	}
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token requis"})
		return
	}

	payload := map[string]string{"refresh_token": refreshToken}
	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", supabaseBaseURL+"/auth/v1/logout", bytes.NewBuffer(jsonBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création requête Supabase"})
		return
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("apikey", supabaseAnonKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur requête Supabase"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": "Erreur logout Supabase", "details": string(body)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Déconnexion réussie"})
}
