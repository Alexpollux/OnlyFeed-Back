package stripe

import (
	"github.com/stripe/stripe-go/v78"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78/subscription"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
)

type Subscription struct {
	ID                   string `gorm:"primaryKey"`
	CreatedAt            time.Time
	SubscriberID         string
	CreatorID            string
	Status               string
	StripeSubscriptionID string
	Price                float64
}

func Unsubscribe(c *gin.Context) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	subscriberID := c.GetString("user_id")
	creatorID := c.Param("creator_id")

	var existing Subscription
	if err := database.DB.
		Where("subscriber_id = ? AND creator_id = ? AND status = ?", subscriberID, creatorID, "active").
		First(&existing).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Abonnement introuvable"})
		return
	}

	if existing.StripeSubscriptionID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID d’abonnement Stripe manquant"})
		return
	}

	// Récupération du StripeAccountID du créateur
	var creator struct {
		StripeAccountID string
	}
	if err := database.DB.
		Table("users").
		Select("stripe_account_id").
		Where("id = ?", creatorID).
		Scan(&creator).Error; err != nil || creator.StripeAccountID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de récupérer le compte Stripe du créateur"})
		return
	}

	// Injecter StripeAccount dans les paramètres
	baseParams := &stripe.SubscriptionCancelParams{}
	baseParams.StripeAccount = &creator.StripeAccountID

	_, err := subscription.Cancel(existing.StripeSubscriptionID, baseParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Exrreur lors de l’annulation Stripe"})
		return
	}

	existing.Status = "cancelled"
	if err := database.DB.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour locale"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Abonnement annulé"})
}
