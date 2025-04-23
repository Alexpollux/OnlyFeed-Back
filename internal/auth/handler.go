package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
	"github.com/gin-gonic/gin"
)

// Signup : Inscription
func Signup(c *gin.Context) {
	supabaseBaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")

	var input struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Username  string `json:"username"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
		AvatarURL string `json:"avatar_url"`
		Bio       string `json:"bio"`
		Language  string `json:"language"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requête invalide"})
		return
	}

	if input.Email == "" || input.Password == "" || input.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Champs requis manquants"})
		return
	}

	// Vérification que email et username n'existe pas
	if user.ExistsByEmail(input.Email) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email déjà utilisé"})
		return
	}
	if user.ExistsByUsername(input.Username) {
		c.JSON(http.StatusConflict, gin.H{"error": "Nom d'utilisateur déjà utilisé"})
		return
	}

	validLanguages := map[string]bool{
		"fr": true,
		"en": true,
	}
	if !validLanguages[input.Language] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Langue non supportée"})
		return
	}

	// Étape 1 – Appel à Supabase Auth
	authPayload := map[string]string{
		"email":    input.Email,
		"password": input.Password,
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

	// Lire la réponse AVANT de faire quoi que ce soit
	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{"error": "Erreur Auth", "details": string(respBytes)})
		return
	}

	// Étape 2 –Extraire le user.id
	// Sans la confirmation par mail
	var authResp struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	// Avec la confirmation par mail
	//var authResp struct {
	//	ID string `json:"id"`
	//}
	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur parsing user.id"})
		return
	}

	userID := authResp.User.ID // sans confirmation par mail
	//userID := authResp.ID // avec confirmation par mail
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aucun ID utilisateur renvoyé par Supabase"})
		return
	}

	// Étape 3 – Créer l’utilisateur dans ta table personnalisée
	newUser := user.User{
		ID:        userID,
		CreatedAt: time.Now(),
		Username:  input.Username,
		Firstname: input.Firstname,
		Lastname:  input.Lastname,
		AvatarURL: input.AvatarURL,
		Bio:       input.Bio,
		Email:     input.Email,
		Language:  input.Language,
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur insertion base utilisateurs"})
		return
	}

	// Réponse finale
	c.JSON(http.StatusCreated, gin.H{
		"message": "Utilisateur inscrit 🎉",
		"user":    newUser,
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
