package message

import (
	"time"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

// Conversation représente une conversation entre deux utilisateurs
type Conversation struct {
	ID            string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	User1ID       string     `json:"user1_id" gorm:"index"`
	User2ID       string     `json:"user2_id" gorm:"index"`
	User1         user.User  `json:"user1" gorm:"foreignKey:User1ID"`
	User2         user.User  `json:"user2" gorm:"foreignKey:User2ID"`
	LastMessage   *Message   `json:"last_message,omitempty" gorm:"foreignKey:ConversationID"`
	LastMessageAt *time.Time `json:"last_message_at"`
}

// Message représente un message dans une conversation
type Message struct {
	ID             string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	ConversationID string       `json:"conversation_id" gorm:"index"`
	Conversation   Conversation `json:"-" gorm:"foreignKey:ConversationID"`
	SenderID       string       `json:"sender_id" gorm:"index"`
	Sender         user.User    `json:"sender" gorm:"foreignKey:SenderID"`
	ReceiverID     string       `json:"receiver_id" gorm:"index"`
	Receiver       user.User    `json:"receiver" gorm:"foreignKey:ReceiverID"`
	Content        string       `json:"content" gorm:"type:text"`
	MessageType    MessageType  `json:"message_type" gorm:"default:'text'"`
	MediaURL       string       `json:"media_url,omitempty"`
	IsRead         bool         `json:"is_read" gorm:"default:false"`
	ReadAt         *time.Time   `json:"read_at,omitempty"`
	IsDeleted      bool         `json:"is_deleted" gorm:"default:false"`
	DeletedAt      *time.Time   `json:"deleted_at,omitempty"`
}

// MessageType définit les types de messages possibles
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
	MessageTypeFile  MessageType = "file"
)

// CreateMessageInput structure pour créer un nouveau message
type CreateMessageInput struct {
	ReceiverID  string      `json:"receiver_id" binding:"required"`
	Content     string      `json:"content"`
	MessageType MessageType `json:"message_type" binding:"required"`
}

// ConversationResponse structure pour la réponse d'une conversation
type ConversationResponse struct {
	ID            string           `json:"id"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	OtherUser     ConversationUser `json:"other_user"`
	LastMessage   *MessageResponse `json:"last_message,omitempty"`
	LastMessageAt *time.Time       `json:"last_message_at"`
	UnreadCount   int64            `json:"unread_count"`
}

// ConversationUser structure pour l'utilisateur dans une conversation
type ConversationUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	IsCreator bool   `json:"is_creator"`
}

// MessageResponse structure pour la réponse d'un message
type MessageResponse struct {
	ID             string           `json:"id"`
	CreatedAt      time.Time        `json:"created_at"`
	ConversationID string           `json:"conversation_id"`
	Sender         ConversationUser `json:"sender"`
	Content        string           `json:"content"`
	MessageType    MessageType      `json:"message_type"`
	MediaURL       string           `json:"media_url,omitempty"`
	IsRead         bool             `json:"is_read"`
	ReadAt         *time.Time       `json:"read_at,omitempty"`
	IsDeleted      bool             `json:"is_deleted"`
}

// ConversationDeletion représente une suppression de conversation côté utilisateur
type ConversationDeletion struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt      time.Time `json:"created_at"`
	UserID         string    `json:"user_id" gorm:"index"`
	ConversationID string    `json:"conversation_id" gorm:"index"`
	DeletedAt      time.Time `json:"deleted_at"`
}
