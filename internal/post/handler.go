package post

import (
	"fmt"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/storage"
)

// CreatePost gère la création d'un nouveau post avec média
func CreatePost(c *gin.Context) {
	// Récupération de l'ID utilisateur depuis le contexte (ajouté par le middleware d'authentification)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	// Récupération des données du formulaire
	title := c.PostForm("title")
	description := c.PostForm("description")
	isPaidStr := c.PostForm("is_paid")
	isPaid := isPaidStr == "true"

	var user user.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// Si l'utilisateur n'est pas créateur, forcer isPaid à false
	if !user.IsCreator {
		isPaid = false
	}

	// Vérification des champs obligatoires
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le titre est obligatoire"})
		return
	}

	// Upload du média
	file, header, err := c.Request.FormFile("media")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Aucun média fourni", "details": err.Error()})
		return
	}
	defer file.Close()

	// Validation du type de fichier
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true, ".heic": true,
		".mp4": true, ".mov": true, ".avi": true,
	}

	if !validExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Extension de fichier invalide"})
		return
	}

	// Génération d'un nom de fichier unique
	postID := uuid.New().String()
	filename := fmt.Sprintf("post_%s%s", postID, ext)
	contentType := header.Header.Get("Content-Type")

	// Upload du fichier vers S3
	url, err := storage.UploadToS3(file, filename, contentType, "posts")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'upload", "details": err.Error()})
		return
	}

	// Création du post en base de données
	newPost := Post{
		ID:          postID,
		CreatedAt:   time.Now(),
		UserID:      userID.(string),
		Title:       title,
		Description: description,
		MediaURL:    url,
		IsPaid:      isPaid,
	}

	if err := database.DB.Create(&newPost).Error; err != nil {
		// Si l'insertion en BDD échoue, on tente de supprimer le fichier déjà uploadé
		urlParts := strings.Split(url, ".amazonaws.com/")
		if len(urlParts) > 1 {
			_ = storage.DeleteFromS3(urlParts[1]) // On ignore l'erreur ici
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post créé avec succès",
		"post":    newPost,
	})
}

// GetUserPosts récupère tous les posts d'un utilisateur
func GetUserPosts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	var posts []Post
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// GetAllPosts récupère tous les posts, avec filtrage optionnel pour les posts payants
func GetAllPosts(c *gin.Context) {
	// Récupérer l'ID utilisateur si disponible (utilisateur connecté)
	userID, userLoggedIn := c.Get("user_id")

	// Paramètre de requête pour filtrer par contenu payant/gratuit
	showPaywalled := c.Query("paywalled") == "true"

	query := database.DB.Order("created_at DESC")

	// Filtrer les posts selon les règles d'accès
	if !showPaywalled || !userLoggedIn {
		// Par défaut ou utilisateur non connecté: montrer uniquement les posts gratuits
		query = query.Where("is_paid = ?", false)
	} else if userLoggedIn {
		// Utilisateur connecté qui veut voir du contenu payant:
		// Montrer les posts gratuits et ses propres posts payants
		// Note: pour un système d'abonnement complet, vous devriez ajouter une vérification ici
		query = query.Where("is_paid = ? OR (is_paid = ? AND user_id = ?)", false, true, userID)
	}

	var posts []Post
	if err := query.Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// GetPostByID récupère un post spécifique par son ID
func GetPostByID(c *gin.Context) {
	postID := c.Param("id")

	var post Post
	if err := database.DB.First(&post, "id = ?", postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post non trouvé"})
		return
	}

	// Vérification si l'utilisateur a accès au post s'il est payant
	if post.IsPaid {
		userID, exists := c.Get("user_id")
		if !exists || post.UserID != userID.(string) {
			// Ici, vous pourriez implémenter une vérification d'abonnement
			// Pour l'instant, seul le créateur peut voir son propre post payant
			c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé à ce contenu premium"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"post": post,
	})
}

// DeletePost supprime un post
func DeletePost(c *gin.Context) {
	postID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	// Vérifier que le post existe et appartient à l'utilisateur
	var post Post
	if err := database.DB.First(&post, "id = ? AND user_id = ?", postID, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post non trouvé ou vous n'êtes pas autorisé à le supprimer"})
		return
	}

	// Extraire la clé du média de l'URL pour le supprimer de S3
	if post.MediaURL != "" {
		urlParts := strings.Split(post.MediaURL, ".amazonaws.com/")
		if len(urlParts) > 1 {
			mediaKey := urlParts[1]
			if err := storage.DeleteFromS3(mediaKey); err != nil {
				// On continue même en cas d'erreur pour supprimer l'entrée en BDD
				fmt.Printf("Erreur lors de la suppression du média sur S3: %v\n", err)
			}
		}
	}

	// Supprimer l'entrée en base de données
	if err := database.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression du post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post supprimé avec succès",
	})
}

// GetCommentsByPostID récupère tous les commentaires pour un post spécifique
func GetCommentsByPostID(c *gin.Context) {
	postID := c.Param("id")

	// Vérifier que le post existe
	var post Post
	if err := database.DB.First(&post, "id = ?", postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post non trouvé"})
		return
	}

	// Vérifier l'accès si le post est payant
	if post.IsPaid {
		userID, exists := c.Get("user_id")
		if !exists || post.UserID != userID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé à ce contenu premium"})
			return
		}
	}

	var comments []Comment
	if err := database.DB.Where("post_id = ?", postID).Order("created_at desc").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des commentaires"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
	})
}

// CreateComment ajoute un nouveau commentaire
func CreateComment(c *gin.Context) {
	// Récupération de l'ID utilisateur depuis le contexte
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	// Récupération des données du commentaire
	var input struct {
		PostID string `json:"post_id" binding:"required"`
		Text   string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Vérifier que le post existe
	var post Post
	if err := database.DB.First(&post, "id = ?", input.PostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post non trouvé"})
		return
	}

	// Vérifier l'accès si le post est payant
	if post.IsPaid {
		if post.UserID != userID.(string) {
			// Pour l'instant, seul le créateur peut commenter sur son propre post payant
			// Ici, tu pourrais implémenter une vérification d'abonnement
			c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé à ce contenu premium"})
			return
		}
	}

	// Récupérer le nom d'utilisateur
	var username string
	row := database.DB.Table("users").Select("username").Where("id = ?", userID).Row()
	if err := row.Scan(&username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de l'utilisateur"})
		return
	}

	// Création du commentaire
	comment := Comment{
		PostID:    input.PostID,
		UserID:    userID.(string),
		Username:  username,   // Sera ignoré par GORM grâce à gorm:"-"
		Content:   input.Text, // Utilise Content au lieu de Text
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du commentaire"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Commentaire ajouté avec succès",
		"comment": comment,
	})
}

// DeleteComment supprime un commentaire
func DeleteComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	// Vérifier que le commentaire existe et appartient à l'utilisateur
	var comment Comment
	if err := database.DB.First(&comment, "id = ? AND user_id = ?", commentID, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Commentaire non trouvé ou vous n'êtes pas autorisé à le supprimer"})
		return
	}

	// Supprimer le commentaire
	if err := database.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression du commentaire"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Commentaire supprimé avec succès",
	})
}
