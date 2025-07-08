package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
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
	route := c.FullPath()

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
		missingFields := ""
		if email != "" {
			missingFields = "Email"
		}
		if password != "" {
			if missingFields != "" {
				missingFields += " & "
			}
			missingFields += "Password"
		}
		if username != "" {
			if missingFields != "" {
				missingFields += " & "
			}
			missingFields += "Username"
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "Champs requis manquants"})
		logs.LogJSON("WARN", "Missing required fields", map[string]interface{}{
			"route": route,
			"extra": fmt.Sprintf("Missing required fields : %s", missingFields),
		})
		return
	}

	// Vérification unique email/username
	if user.ExistsByEmail(email) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email déjà utilisé"})
		logs.LogJSON("WARN", "Email already used", map[string]interface{}{
			"route": route,
			"extra": fmt.Sprintf("Email already used : %s", email),
		})
		return
	}
	if user.ExistsByUsername(username) {
		c.JSON(http.StatusConflict, gin.H{"error": "Nom d'utilisateur déjà utilisé"})
		logs.LogJSON("WARN", "Username already used", map[string]interface{}{
			"route": route,
			"extra": fmt.Sprintf("Username already used : %s", username),
		})
		return
	}

	// Vérification de la langue
	validLanguages := map[string]bool{"fr": true, "en": true}
	if language != "" && !validLanguages[language] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Langue non supportée"})
		logs.LogJSON("WARN", "Language not supported", map[string]interface{}{
			"route":    route,
			"email":    email,
			"username": username,
			"extra":    fmt.Sprintf("Language not supported : %s", language),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de connexion à Supabase"})
		logs.LogJSON("ERROR", "Supabase Auth Error", map[string]interface{}{
			"email": email,
			"error": err.Error(),
			"route": route,
		})
		return
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{"error": "Erreur Auth", "details": string(respBytes)})
		logs.LogJSON("ERROR", "Auth Error", map[string]interface{}{
			"email": email,
			"route": route,
			"extra": fmt.Sprintf("Response : %p", resp),
		})
		return
	}

	var authResp struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur parsing user.id"})
		logs.LogJSON("ERROR", "Error parsing user.id", map[string]interface{}{
			"email": email,
			"error": err.Error(),
			"route": route,
			"extra": fmt.Sprintf("Response : %p", respBytes),
		})
		return
	}

	userID := authResp.User.ID
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aucun ID utilisateur renvoyé"})
		logs.LogJSON("ERROR", "No user ID returned", map[string]interface{}{
			"email": email,
			"route": route,
		})
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
			logs.LogJSON("WARN", "Invalid file extension", map[string]interface{}{
				"route":  route,
				"userID": userID,
				"extra":  fmt.Sprintf("Invalid file extension : %s", ext),
			})
			return
		}

		filename := fmt.Sprintf("user_%s%s", userID, ext)
		contentType := header.Header.Get("Content-Type")

		url, err := storage.UploadToS3(file, filename, contentType, "avatars")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'upload S3", "details": err.Error()})
			logs.LogJSON("ERROR", "S3 upload error", map[string]interface{}{
				"route":  route,
				"userID": userID,
			})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de l'utilisateur"})
		logs.LogJSON("ERROR", "Error creating user", map[string]interface{}{
			"route":  route,
			"userID": userID,
			"extra":  fmt.Sprintf("User Data : %v", newUser),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Utilisateur inscrit 🎉",
		"user":    newUser,
	})
	logs.LogJSON("INFO", "User created successfully", map[string]interface{}{
		"route":  route,
		"userID": userID,
	})
}

func Login(c *gin.Context) {
	route := c.FullPath()

	supabaseBaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil || input.Email == "" || input.Password == "" {
		missingFields := ""
		if input.Email != "" {
			missingFields = "Email"
		}
		if input.Password != "" {
			if missingFields != "" {
				missingFields += " & "
			}
			missingFields += "Password"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Champs email et password requis"})
		logs.LogJSON("WARN", "Missing required fields", map[string]interface{}{
			"route": route,
			"extra": fmt.Sprintf("Missing required fields : %s", missingFields),
		})
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
		logs.LogJSON("ERROR", "Supabase Auth Error", map[string]interface{}{
			"email": input.Email,
			"error": err.Error(),
			"route": route,
		})
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error":   "Erreur d'authentification",
			"details": string(bodyBytes),
		})
		logs.LogJSON("ERROR", "Auth Error", map[string]interface{}{
			"email": input.Email,
			"route": route,
			"extra": fmt.Sprintf("Response : %p", resp),
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
		logs.LogJSON("ERROR", "Supabase response parsing error", map[string]interface{}{
			"email": input.Email,
			"error": err.Error(),
			"route": route,
			"extra": fmt.Sprintf("Response : %p", bodyBytes),
		})
		return
	}

	// Parser manuellement l'access_token pour extraire user_id (champ "sub")
	tokenClaims := utils.ParseJWTClaims(supabaseResp.AccessToken)
	userID, ok := tokenClaims["sub"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID utilisateur manquant dans le token"})
		logs.LogJSON("ERROR", "User ID missing from token", map[string]interface{}{
			"email": input.Email,
			"route": route,
			"extra": fmt.Sprintf("Token : %p", tokenClaims),
		})
		return
	}

	// Récupération du user depuis ta base
	var u user.User
	if err := database.DB.First(&u, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur introuvable en base"})
		logs.LogJSON("WARN", "User not found", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": userID,
			"extra":  fmt.Sprintf("User not found : %s", userID),
		})
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
		"is_creator": u.IsCreator,
	}

	if u.IsAdmin {
		respUser["is_admin"] = true
	}
	if u.IsCreator {
		respUser["subscription_price"] = u.SubscriptionPrice
	}

	// Réponse personnalisée
	c.JSON(http.StatusOK, gin.H{
		"access_token":  supabaseResp.AccessToken,
		"refresh_token": supabaseResp.RefreshToken,
		"user":          respUser,
	})
	logs.LogJSON("INFO", "User successfully logged in", map[string]interface{}{
		"route":  route,
		"userID": userID,
	})
}

func Logout(c *gin.Context) {
	route := c.FullPath()

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de connexion à Supabase"})
		logs.LogJSON("ERROR", "Supabase Auth Error", map[string]interface{}{
			"error": err.Error(),
			"route": route,
		})
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
