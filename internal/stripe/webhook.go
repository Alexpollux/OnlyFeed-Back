package stripe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/subscription"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

func HandleStripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lecture du corps échouée"})
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	sigHeader := c.GetHeader("Stripe-Signature")

	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signature Stripe invalide"})
		return
	}

	switch event.Type {

	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err == nil {
			handleCheckoutSessionCompleted(session)
		}

	default:
		fmt.Printf("⚠️  Événement non géré : %s\n", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func handleCheckoutSessionCompleted(session stripe.CheckoutSession) {
	fmt.Println("✅ Abonnement confirmé via Stripe !")

	creatorID := session.Metadata["creator_id"]
	subscriberID := session.Metadata["subscriber_id"]

	if creatorID == "" || subscriberID == "" {
		fmt.Println("❌ Metadata manquante")
		return
	}

	var creator user.User
	if err := database.DB.First(&creator, "id = ?", creatorID).Error; err != nil {
		fmt.Println("❌ Erreur lors de la récupération des infos du créateur")
		return
	}

	// Récupérer l'ID d'abonnement Stripe
	subscriptionID := session.Subscription.ID
	if subscriptionID == "" {
		fmt.Println("❌ ID d’abonnement Stripe manquant dans la session")
		return
	}

	// Vérifie si déjà abonné
	var existing subscription.Subscription
	err := database.DB.Where("subscriber_id = ? AND creator_id = ?", subscriberID, creatorID).First(&existing).Error
	if err == nil {
		if existing.Status == "active" {
			fmt.Println("ℹ️ Déjà abonné (actif) → aucune action")
			return
		}

		// Réactiver abonnement annulé
		existing.Status = "active"
		existing.StripeSubscriptionID = subscriptionID
		existing.Price = creator.SubscriptionPrice
		if err := database.DB.Save(&existing).Error; err != nil {
			fmt.Println("❌ Erreur lors de la réactivation :", err)
			return
		}

		fmt.Println("✅ Abonnement réactivé")
		return
	}

	// Crée l’abonnement
	sub := subscription.Subscription{
		CreatedAt:            time.Now(),
		SubscriberID:         subscriberID,
		CreatorID:            creatorID,
		Status:               "active",
		StripeSubscriptionID: subscriptionID,
		Price:                creator.SubscriptionPrice,
	}
	if err := database.DB.Create(&sub).Error; err != nil {
		fmt.Println("❌ Erreur lors de la création de l'abonnement :", err)
		return
	}

	fmt.Printf("✅ Abonnement créé : %s → %s\n", subscriberID, creatorID)
}
