package report

import (
	"time"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/post"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

// ReportType définit les types d'entités qui peuvent être signalées
type ReportType string

const (
	ReportTypePost    ReportType = "post"
	ReportTypeUser    ReportType = "user"
	ReportTypeComment ReportType = "comment"
)

// ReportReason définit les raisons de signalement
type ReportReason string

const (
	ReasonInappropriateContent ReportReason = "inappropriate_content"
	ReasonSpam                 ReportReason = "spam"
	ReasonHateSpeech           ReportReason = "hate_speech"
	ReasonImpersonation        ReportReason = "impersonation"
	ReasonCopyright            ReportReason = "copyright"
	ReasonOther                ReportReason = "other"
)

// ReportStatus définit les statuts d'un signalement
type ReportStatus string

const (
	StatusPending  ReportStatus = "pending"
	StatusReviewed ReportStatus = "reviewed"
	StatusResolved ReportStatus = "resolved"
	StatusRejected ReportStatus = "rejected"
)

// Report représente un signalement
type Report struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Utilisateur qui fait le signalement
	ReporterID string    `json:"reporter_id" gorm:"index"`
	Reporter   user.User `json:"reporter" gorm:"foreignKey:ReporterID"`

	// Entité signalée
	TargetType ReportType `json:"target_type" gorm:"index"`
	TargetID   string     `json:"target_id" gorm:"index"`

	// Contenu du signalement
	Reason      ReportReason `json:"reason"`
	Description string       `json:"description" gorm:"type:text"`

	// Statut et traitement
	Status     ReportStatus `json:"status" gorm:"default:'pending';index"`
	AdminID    *string      `json:"admin_id,omitempty"`
	Admin      *user.User   `json:"admin,omitempty" gorm:"foreignKey:AdminID"`
	AdminNote  string       `json:"admin_note" gorm:"type:text"`
	ResolvedAt *time.Time   `json:"resolved_at,omitempty"`
}

// CreateReportInput structure pour créer un signalement
type CreateReportInput struct {
	TargetType  ReportType   `json:"target_type" binding:"required"`
	TargetID    string       `json:"target_id" binding:"required"`
	Reason      ReportReason `json:"reason" binding:"required"`
	Description string       `json:"description"`
}

// UpdateReportInput structure pour mettre à jour un signalement (admin)
type UpdateReportInput struct {
	Status    ReportStatus `json:"status" binding:"required"`
	AdminNote string       `json:"admin_note"`
}

// ReportWithTarget structure pour la réponse avec les détails de la cible
type ReportWithTarget struct {
	Report
	TargetPost    *post.Post    `json:"target_post,omitempty"`
	TargetUser    *user.User    `json:"target_user,omitempty"`
	TargetComment *post.Comment `json:"target_comment,omitempty"`
}

// Validation des raisons de signalement
func (r ReportReason) IsValid() bool {
	switch r {
	case ReasonInappropriateContent, ReasonSpam, ReasonHateSpeech,
		ReasonImpersonation, ReasonCopyright, ReasonOther:
		return true
	default:
		return false
	}
}

// Validation des types de signalement
func (t ReportType) IsValid() bool {
	switch t {
	case ReportTypePost, ReportTypeUser, ReportTypeComment:
		return true
	default:
		return false
	}
}
