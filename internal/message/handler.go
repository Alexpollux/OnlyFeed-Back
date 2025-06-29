package message

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/storage"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

// GetConversations récupère toutes les conversations de l'utilisateur connecté
func GetConversations(c *gin.Context) {
	userID := c.GetString("user_id")

	// Récupérer les conversations où l'utilisateur n'a pas créé de suppression
	var conversations []Conversation
	if err := database.DB.
		Where("(user1_id = ? OR user2_id = ?) AND id NOT IN (?)",
			userID, userID,
			database.DB.Table("conversation_deletions").
				Select("conversation_id").
				Where("user_id = ?", userID)).
		Preload("User1").
		Preload("User2").
		Order("last_message_at DESC NULLS LAST, created_at DESC").
		Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des conversations"})
		return
	}

	var response []ConversationResponse
	for _, conv := range conversations {
		// Déterminer l'autre utilisateur
		var otherUser user.User
		if conv.User1ID == userID {
			otherUser = conv.User2
		} else {
			otherUser = conv.User1
		}

		// Récupérer la date de suppression pour cet utilisateur (si elle existe)
		var deletionTime *time.Time
		var deletion ConversationDeletion
		if err := database.DB.
			Where("user_id = ? AND conversation_id = ?", userID, conv.ID).
			First(&deletion).Error; err == nil {
			deletionTime = &deletion.DeletedAt
		}

		// Compter les messages non lus postérieurs à la suppression
		var unreadCount int64
		unreadQuery := database.DB.Model(&Message{}).
			Where("conversation_id = ? AND receiver_id = ? AND is_read = false AND is_deleted = false", conv.ID, userID)

		if deletionTime != nil {
			unreadQuery = unreadQuery.Where("created_at > ?", *deletionTime)
		}

		unreadQuery.Count(&unreadCount)

		// Récupérer le dernier message postérieur à la suppression
		var lastMessage *MessageResponse
		if conv.LastMessageAt != nil {
			var msg Message
			lastMsgQuery := database.DB.
				Where("conversation_id = ?", conv.ID).
				Preload("Sender").
				Order("created_at DESC")

			if deletionTime != nil {
				lastMsgQuery = lastMsgQuery.Where("created_at > ?", *deletionTime)
			}

			if err := lastMsgQuery.First(&msg).Error; err == nil {
				lastMessage = &MessageResponse{
					ID:             msg.ID,
					CreatedAt:      msg.CreatedAt,
					ConversationID: msg.ConversationID,
					Sender: ConversationUser{
						ID:        msg.Sender.ID,
						Username:  msg.Sender.Username,
						AvatarURL: msg.Sender.AvatarURL,
						IsCreator: msg.Sender.IsCreator,
					},
					Content:     msg.Content,
					MessageType: msg.MessageType,
					MediaURL:    msg.MediaURL,
					IsRead:      msg.IsRead,
					ReadAt:      msg.ReadAt,
					IsDeleted:   msg.IsDeleted,
				}
			}
		}

		// Ne pas inclure les conversations sans messages après suppression
		if deletionTime != nil && lastMessage == nil {
			continue
		}

		convResponse := ConversationResponse{
			ID:        conv.ID,
			CreatedAt: conv.CreatedAt,
			UpdatedAt: conv.UpdatedAt,
			OtherUser: ConversationUser{
				ID:        otherUser.ID,
				Username:  otherUser.Username,
				AvatarURL: otherUser.AvatarURL,
				IsCreator: otherUser.IsCreator,
			},
			LastMessage:   lastMessage,
			LastMessageAt: conv.LastMessageAt,
			UnreadCount:   unreadCount,
		}

		response = append(response, convResponse)
	}

	c.JSON(http.StatusOK, gin.H{"conversations": response})
}

// GetConversationMessages récupère les messages d'une conversation
func GetConversationMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	// Vérifier que l'utilisateur fait partie de la conversation
	var conversation Conversation
	if err := database.DB.
		Where("id = ? AND (user1_id = ? OR user2_id = ?)", conversationID, userID, userID).
		First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Conversation non trouvée"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de la conversation"})
		}
		return
	}

	// Récupérer la date de suppression pour cet utilisateur (si elle existe)
	var deletionTime *time.Time
	var deletion ConversationDeletion
	if err := database.DB.
		Where("user_id = ? AND conversation_id = ?", userID, conversationID).
		First(&deletion).Error; err == nil {
		deletionTime = &deletion.DeletedAt
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	// Récupérer les messages postérieurs à la suppression
	var messages []Message
	msgQuery := database.DB.
		Where("conversation_id = ? AND is_deleted = false", conversationID)

	if deletionTime != nil {
		msgQuery = msgQuery.Where("created_at > ?", *deletionTime)
	}

	if err := msgQuery.
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des messages"})
		return
	}

	// Marquer les messages comme lus (seulement ceux postérieurs à la suppression)
	go func() {
		markAsReadQuery := database.DB.Model(&Message{}).
			Where("conversation_id = ? AND receiver_id = ? AND is_read = false", conversationID, userID)

		if deletionTime != nil {
			markAsReadQuery = markAsReadQuery.Where("created_at > ?", *deletionTime)
		}

		markAsReadQuery.Updates(map[string]interface{}{
			"is_read": true,
			"read_at": time.Now(),
		})
	}()

	// Convertir en response format
	var response []MessageResponse
	for _, msg := range messages {
		msgResponse := MessageResponse{
			ID:             msg.ID,
			CreatedAt:      msg.CreatedAt,
			ConversationID: msg.ConversationID,
			Sender: ConversationUser{
				ID:        msg.Sender.ID,
				Username:  msg.Sender.Username,
				AvatarURL: msg.Sender.AvatarURL,
				IsCreator: msg.Sender.IsCreator,
			},
			Content:     msg.Content,
			MessageType: msg.MessageType,
			MediaURL:    msg.MediaURL,
			IsRead:      msg.IsRead,
			ReadAt:      msg.ReadAt,
			IsDeleted:   msg.IsDeleted,
		}
		response = append(response, msgResponse)
	}

	c.JSON(http.StatusOK, gin.H{"messages": response})
}

// SendMessage envoie un nouveau message
func SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")

	// Vérifier si c'est un message avec média ou texte
	var input CreateMessageInput
	var mediaURL string

	// Tentative de parsing JSON pour message texte
	if err := c.ShouldBindJSON(&input); err != nil {
		// Si erreur JSON, alors c'est probablement un form-data avec média
		receiverID := c.PostForm("receiver_id")
		content := c.PostForm("content")
		messageTypeStr := c.PostForm("message_type")

		if receiverID == "" || messageTypeStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "receiver_id et message_type sont requis"})
			return
		}

		input = CreateMessageInput{
			ReceiverID:  receiverID,
			Content:     content,
			MessageType: MessageType(messageTypeStr),
		}

		// Traitement du fichier média si présent
		if input.MessageType != MessageTypeText {
			file, header, err := c.Request.FormFile("media")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Fichier média requis pour ce type de message"})
				return
			}
			defer file.Close()

			// Validation du type de fichier
			ext := strings.ToLower(filepath.Ext(header.Filename))
			validExtensions := getValidExtensions(input.MessageType)

			if !validExtensions[ext] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Extension de fichier invalide"})
				return
			}

			// Upload du fichier
			messageID := uuid.New().String()
			filename := fmt.Sprintf("message_%s%s", messageID, ext)
			contentType := header.Header.Get("Content-Type")

			url, err := storage.UploadToS3(file, filename, contentType, "messages")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'upload du fichier"})
				return
			}
			mediaURL = url
		}
	}

	// Vérifier que l'utilisateur destinataire existe
	var receiver user.User
	if err := database.DB.First(&receiver, "id = ?", input.ReceiverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur destinataire non trouvé"})
		return
	}

	// Vérifier qu'on n'envoie pas un message à soi-même
	if userID == input.ReceiverID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Impossible d'envoyer un message à soi-même"})
		return
	}

	// Trouver ou créer la conversation
	conversation, err := findOrCreateConversation(userID, input.ReceiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de la conversation"})
		return
	}

	// Créer le message
	message := Message{
		ConversationID: conversation.ID,
		SenderID:       userID,
		ReceiverID:     input.ReceiverID,
		Content:        input.Content,
		MessageType:    input.MessageType,
		MediaURL:       mediaURL,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := database.DB.Create(&message).Error; err != nil {
		// Si création échoue et qu'on a uploadé un fichier, le supprimer
		if mediaURL != "" {
			urlParts := strings.Split(mediaURL, ".amazonaws.com/")
			if len(urlParts) > 1 {
				_ = storage.DeleteFromS3(urlParts[1])
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'envoi du message"})
		return
	}

	// Si l'utilisateur destinataire avait supprimé la conversation,
	// supprimer l'enregistrement de suppression pour lui permettre de voir les nouveaux messages
	database.DB.Where("user_id = ? AND conversation_id = ?", input.ReceiverID, conversation.ID).
		Delete(&ConversationDeletion{})

	// Faire de même pour l'expéditeur au cas où
	database.DB.Where("user_id = ? AND conversation_id = ?", userID, conversation.ID).
		Delete(&ConversationDeletion{})

	// Mettre à jour la conversation avec le dernier message
	now := time.Now()
	database.DB.Model(&conversation).Updates(map[string]interface{}{
		"last_message_at": now,
		"updated_at":      now,
	})

	// Récupérer le message avec les relations pour la réponse
	database.DB.Preload("Sender").First(&message, message.ID)

	response := MessageResponse{
		ID:             message.ID,
		CreatedAt:      message.CreatedAt,
		ConversationID: message.ConversationID,
		Sender: ConversationUser{
			ID:        message.Sender.ID,
			Username:  message.Sender.Username,
			AvatarURL: message.Sender.AvatarURL,
			IsCreator: message.Sender.IsCreator,
		},
		Content:     message.Content,
		MessageType: message.MessageType,
		MediaURL:    message.MediaURL,
		IsRead:      message.IsRead,
		ReadAt:      message.ReadAt,
		IsDeleted:   message.IsDeleted,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         response,
		"conversation_id": conversation.ID,
	})
}

// MarkMessageAsRead marque un message comme lu
func MarkMessageAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("id")

	var message Message
	if err := database.DB.
		Where("id = ? AND receiver_id = ?", messageID, userID).
		First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message non trouvé"})
		return
	}

	if !message.IsRead {
		now := time.Now()
		if err := database.DB.Model(&message).Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour du message"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marqué comme lu"})
}

// DeleteMessage supprime un message (soft delete)
func DeleteMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("id")

	var message Message
	if err := database.DB.
		Where("id = ? AND sender_id = ?", messageID, userID).
		First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message non trouvé"})
		return
	}

	now := time.Now()
	if err := database.DB.Model(&message).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression du message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message supprimé"})
}

// DeleteConversation supprime une conversation côté utilisateur (soft delete)
func DeleteConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	// Vérifier que l'utilisateur fait partie de la conversation
	var conversation Conversation
	if err := database.DB.
		Where("id = ? AND (user1_id = ? OR user2_id = ?)", conversationID, userID, userID).
		First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Conversation non trouvée"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de la conversation"})
		}
		return
	}

	// Vérifier si l'utilisateur a déjà supprimé cette conversation
	var existingDeletion ConversationDeletion
	err := database.DB.
		Where("user_id = ? AND conversation_id = ?", userID, conversationID).
		First(&existingDeletion).Error

	if err == nil {
		// Déjà supprimée
		c.JSON(http.StatusOK, gin.H{"message": "Conversation déjà supprimée"})
		return
	}

	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification"})
		return
	}

	// Créer l'enregistrement de suppression
	deletion := ConversationDeletion{
		UserID:         userID,
		ConversationID: conversationID,
		CreatedAt:      time.Now(),
		DeletedAt:      time.Now(),
	}

	if err := database.DB.Create(&deletion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression de la conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation supprimée avec succès"})
}

// Fonctions utilitaires

func findOrCreateConversation(user1ID, user2ID string) (*Conversation, error) {
	var conversation Conversation

	// Chercher une conversation existante
	err := database.DB.
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
			user1ID, user2ID, user2ID, user1ID).
		First(&conversation).Error

	if err == nil {
		return &conversation, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Créer une nouvelle conversation
	conversation = Conversation{
		User1ID:   user1ID,
		User2ID:   user2ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&conversation).Error; err != nil {
		return nil, err
	}

	return &conversation, nil
}

func getValidExtensions(messageType MessageType) map[string]bool {
	switch messageType {
	case MessageTypeImage:
		return map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	case MessageTypeVideo:
		return map[string]bool{".mp4": true, ".mov": true, ".avi": true, ".mkv": true}
	case MessageTypeAudio:
		return map[string]bool{".mp3": true, ".wav": true, ".aac": true, ".m4a": true}
	case MessageTypeFile:
		return map[string]bool{".pdf": true, ".doc": true, ".docx": true, ".txt": true, ".zip": true}
	default:
		return map[string]bool{}
	}
}
